package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ProductCollection *mongo.Collection
var OrderCollection *mongo.Collection
var CustomerCollection *mongo.Collection

func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	db := client.Database("query_service")
	ProductCollection = db.Collection("products")
	OrderCollection = db.Collection("orders")
	CustomerCollection = db.Collection("customers")

	log.Println("✅ MongoDB initialized")
}

func CreateIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for products
	productIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "productId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "sku", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := ProductCollection.Indexes().CreateMany(ctx, productIndexes)
	if err != nil {
		log.Fatalf("Failed to create product indexes: %v", err)
	}

	// Create indexes for orders
	orderIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "orderId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "orderNumber", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = OrderCollection.Indexes().CreateMany(ctx, orderIndexes)
	if err != nil {
		log.Fatalf("Failed to create order indexes: %v", err)
	}

	// Create indexes for customers
	customerIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "customerId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = CustomerCollection.Indexes().CreateMany(ctx, customerIndexes)
	if err != nil {
		log.Fatalf("Failed to create customer indexes: %v", err)
	}

	log.Println("✅ MongoDB indexes created")
}
