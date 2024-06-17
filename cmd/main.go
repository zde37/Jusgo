package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zde37/Jusgo/internal/controller"
	"github.com/zde37/Jusgo/internal/database"
	"github.com/zde37/Jusgo/internal/repository"
	"github.com/zde37/Jusgo/internal/service"
)

func main() {
	// Set up MongoDB connection
	ctx := context.Background()
	client, cancel, err := database.ConnectToMongoDB(os.Getenv("DB_SOURCE"), ctx)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	collection := client.Database(os.Getenv("DATABASE")).Collection(os.Getenv("COLLECTION"))
	r := repository.NewRepository(collection)
	s := service.NewService(r.Repo)
	h := controller.NewHandler(s.Srvc)

	defer cancel()
	defer client.Disconnect(ctx)

	// setup server
	srv := &http.Server{
		Addr:         os.Getenv("SERVER_ADDRESS"),
		Handler:      h.Hndl.Mux(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	quit := make(chan os.Signal, 1) // channel to listen for OS interrupt signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go cronJob()

	go func() {
		log.Printf("server started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not serve on %s: %v\n", srv.Addr, err)
		}
	}()

	<-quit // block until we receive an interrupt signal

	context, cancel := context.WithTimeout(ctx, 15*time.Second) // create a deadline to wait for shutdown
	defer cancel()

	if err := srv.Shutdown(context); err != nil { // shutdown the server gracefully
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}

// cronJob sends a request to the health route every 13 minute. To prevent the server from sleeping on render(default: 15 minutes)
func cronJob() {
	for range time.Tick(13 * time.Minute) {
		_, err := http.Get(os.Getenv("HEALTH"))
		if err != nil {
			log.Println("server is not healthy")
			return
		}
		log.Println("server is healthy")
	}
}
