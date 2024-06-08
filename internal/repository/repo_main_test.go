package repository

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/zde37/Jusgo/internal/database"
)

var testRepo *Repository

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	// Set up MongoDB connection
	ctx := context.Background()
	client, cancel, err := database.ConnectToMongoDB(os.Getenv("DB_SOURCE"), ctx)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}

	database, collection := "Renew", "Test"
	if os.Getenv("DATABASE") != "" || os.Getenv("COLLECTION") != "" {
		database, collection = os.Getenv("DATABASE"), os.Getenv("COLLECTION")
	}

	col := client.Database(database).Collection(collection)
	testRepo = NewRepository(col)

	defer cancel()
	defer client.Disconnect(ctx)

	os.Exit(m.Run())
}
