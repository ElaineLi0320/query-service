package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
    ProductCollection  *mongo.Collection
    OrderCollection    *mongo.Collection
    CustomerCollection *mongo.Collection
)

func InitMongoDB() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("MongoDB connection error:", err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("MongoDB ping failed:", err)
    }

    db := client.Database("just_browsing")
    ProductCollection = db.Collection("products")
    OrderCollection = db.Collection("orders")
    CustomerCollection = db.Collection("customers")

    log.Println("✅ Connected to MongoDB!")
}
