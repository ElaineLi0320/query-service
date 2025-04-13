package main

import (
	"bytes"
	"context"
	"log"
	"query-service/db"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createElasticsearchIndex() {
	// Define the Elasticsearch mappings
	mappings := `{
		"mappings": {
			"properties": {
				"productId": { "type": "keyword" },
				"sku": { "type": "keyword" },
				"name": { 
					"type": "text", 
					"fields": { 
						"keyword": { "type": "keyword" } 
					} 
				},
				"description": { "type": "text" },
				"price": { "type": "float" },
				"category": {
					"properties": {
						"id": { "type": "keyword" },
						"name": { 
							"type": "text", 
							"fields": { 
								"keyword": { "type": "keyword" } 
							} 
						}
					}
				},
				"currentInventory": { "type": "integer" },
				"attributes": {
					"properties": {
						"name": { "type": "keyword" },
						"value": { "type": "keyword" }
					}
				},
				"created": { "type": "date" },
				"updated": { "type": "date" }
			}
		}
	}`

	// Create the index
	req := esapi.IndicesCreateRequest{
		Index: "products",
		Body:  bytes.NewReader([]byte(mappings)),
	}
	res, err := req.Do(context.Background(), db.ElasticsearchClient)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch index: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response from Elasticsearch: %s", res.String())
	}

	log.Println("✅ Elasticsearch index 'products' created successfully!")
}

func main() {
	// Initialize Elasticsearch
	db.InitElasticsearch()

	// Create Elasticsearch index
	createElasticsearchIndex()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("just_browsing")

	_, err = db.Collection("products").InsertOne(ctx, bson.M{
		"productId": "p1001",
		"sku": "SKU-1001",
		"name": "Gin T-Shirt",
		"description": "High quality developer tee",
		"price": 29.99,
		"category": bson.M{
			"id": "c100",
			"name": "Apparel",
			"parentCategory": bson.M{
				"id": "c000",
				"name": "Root",
			},
		},
		"currentInventory": 100,
		"images": []string{"https://example.com/image1.png"},
		"attributes": []bson.M{
			{"name": "color", "value": "black"},
			{"name": "size", "value": "L"},
		},
		"created": time.Now(),
		"updated": time.Now(),
	})
	if err != nil {
		log.Println("⚠️ product already exists or insert error:", err)
	}

	_, err = db.Collection("orders").InsertOne(ctx, bson.M{
		"orderId": "o2001",
		"orderNumber": "ORD-20240401-001",
		"customerId": "c3001",
		"customerEmail": "alice@example.com",
		"customerName": "Alice Zhang",
		"status": "Shipped",
		"totalAmount": 59.98,
		"items": []bson.M{
			{
				"productId": "p1001",
				"productName": "Gin T-Shirt",
				"sku": "SKU-1001",
				"quantity": 2,
				"unitPrice": 29.99,
				"totalPrice": 59.98,
			},
		},
		"shippingAddress": bson.M{
			"addressLine1": "123 Developer St.",
			"addressLine2": "Unit 100",
			"city": "Vancouver",
			"state": "BC",
			"postalCode": "V6B 1A1",
			"country": "Canada",
		},
		"created": time.Now(),
		"updated": time.Now(),
	})
	if err != nil {
		log.Println("⚠️ order insert error:", err)
	}

	_, err = db.Collection("customers").InsertOne(ctx, bson.M{
		"customerId": "c3001",
		"email": "alice@example.com",
		"firstName": "Alice",
		"lastName": "Zhang",
		"phone": "+1-604-555-1234",
		"addresses": []bson.M{
			{
				"addressType": "home",
				"isDefault": true,
				"addressLine1": "123 Developer St.",
				"addressLine2": "Unit 100",
				"city": "Vancouver",
				"state": "BC",
				"postalCode": "V6B 1A1",
				"country": "Canada",
			},
		},
		"orderHistory": []bson.M{
			{
				"orderId": "o2001",
				"orderNumber": "ORD-20240401-001",
				"date": time.Now(),
				"totalAmount": 59.98,
				"status": "Shipped",
			},
		},
		"created": time.Now(),
		"updated": time.Now(),
	})
	if err != nil {
		log.Println("⚠️ customer insert error:", err)
	}

	log.Println("✅ Initial product, order and customer inserted!")
}
