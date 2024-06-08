package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ConnectToMongoDB establishes a connection to MongoDB.
func ConnectToMongoDB(uri string, ctx context.Context) (*mongo.Client, context.CancelFunc, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return client, cancel, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return client, cancel, err
	}

	log.Println("connected to mongodb")
	return client, cancel, nil
}
