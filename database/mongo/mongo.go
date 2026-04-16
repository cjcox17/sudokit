package mongo

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI      string
	Database string
}

var MongoClient *mongo.Client
var Database *mongo.Database

func InitMongo(cfg MongoConfig) {
	var err error
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		slog.Info("Connecting to MongoDB", "attempt", i+1)

		MongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
		if err == nil {
			err = MongoClient.Ping(ctx, nil)
		}

		if err == nil {
			slog.Info("MongoDB connected successfully")
			Database = MongoClient.Database(cfg.Database)
			return
		}

		slog.Warn("MongoDB connection failed", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}

	slog.Error("Could not connect to MongoDB after multiple attempts", "error", err)
	os.Exit(1)
}
