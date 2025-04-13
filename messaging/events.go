package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"query-service/cache"
	db "query-service/db"
	"query-service/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// RegisterEventHandlers registers all event handlers with the consumer
func RegisterEventHandlers(consumer *Consumer) {
	consumer.RegisterHandler("ProductCreated", handleProductCreated)
	consumer.RegisterHandler("ProductUpdated", handleProductUpdated)
	consumer.RegisterHandler("InventoryChanged", handleInventoryChanged)
	consumer.RegisterHandler("OrderCreated", handleOrderCreated)
	consumer.RegisterHandler("OrderStatusChanged", handleOrderStatusChanged)

	log.Println("✅ Event handlers registered")
}

// handleProductCreated processes ProductCreated events
func handleProductCreated(ctx context.Context, data interface{}) error {
	product := models.Product{}
	if err := mapToStruct(data, &product); err != nil {
		return fmt.Errorf("invalid product data: %w", err)
	}

	// Transaction context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: Add to MongoDB
	_, err := db.ProductCollection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to insert product into MongoDB: %w", err)
	}

	// Step 2: Add to Elasticsearch
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product for Elasticsearch: %w", err)
	}

	res, err := db.ElasticsearchClient.Index(
		"products",
		bytes.NewReader(productJSON),
		db.ElasticsearchClient.Index.WithContext(ctx),
		db.ElasticsearchClient.Index.WithDocumentID(product.ProductID),
	)
	if err != nil {
		return fmt.Errorf("failed to index product in Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	// Step 3: Invalidate Redis cache
	if err := cache.RedisClient.Del(ctx, "products:category:"+product.Category.ID).Err(); err != nil {
		log.Printf("⚠️ Warning: Failed to invalidate Redis cache: %v", err)
		// Continue despite cache invalidation failure
	}

	log.Printf("✅ Product created: %s - %s", product.ProductID, product.Name)
	return nil
}

// handleProductUpdated processes ProductUpdated events
func handleProductUpdated(ctx context.Context, data interface{}) error {
	product := models.Product{}
	if err := mapToStruct(data, &product); err != nil {
		return fmt.Errorf("invalid product data: %w", err)
	}

	// Transaction context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: Update MongoDB
	_, err := db.ProductCollection.UpdateOne(
		ctx,
		bson.M{"productId": product.ProductID},
		bson.M{"$set": product},
	)
	if err != nil {
		return fmt.Errorf("failed to update product in MongoDB: %w", err)
	}

	// Step 2: Update Elasticsearch
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product for Elasticsearch: %w", err)
	}

	res, err := db.ElasticsearchClient.Index(
		"products",
		bytes.NewReader(productJSON),
		db.ElasticsearchClient.Index.WithContext(ctx),
		db.ElasticsearchClient.Index.WithDocumentID(product.ProductID),
	)
	if err != nil {
		return fmt.Errorf("failed to update product in Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	// Step 3: Invalidate Redis caches
	pipe := cache.RedisClient.Pipeline()
	pipe.Del(ctx, "product:"+product.ProductID)
	pipe.Del(ctx, "products:category:"+product.Category.ID)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("⚠️ Warning: Failed to invalidate Redis caches: %v", err)
		// Continue despite cache invalidation failure
	}

	log.Printf("✅ Product updated: %s - %s", product.ProductID, product.Name)
	return nil
}

// handleInventoryChanged processes InventoryChanged events
func handleInventoryChanged(ctx context.Context, data interface{}) error {
	inventoryChange := struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
	}{}
	if err := mapToStruct(data, &inventoryChange); err != nil {
		return fmt.Errorf("invalid inventory data: %w", err)
	}

	// Transaction context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: Update MongoDB
	result, err := db.ProductCollection.UpdateOne(
		ctx,
		bson.M{"productId": inventoryChange.ProductID},
		bson.M{"$set": bson.M{"currentInventory": inventoryChange.Quantity}},
	)
	if err != nil {
		return fmt.Errorf("failed to update inventory in MongoDB: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product not found: %s", inventoryChange.ProductID)
	}

	// Step 2: Update Redis cache
	err = cache.RedisClient.Set(
		ctx,
		"inventory:"+inventoryChange.ProductID,
		inventoryChange.Quantity,
		10*time.Minute,
	).Err()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to update Redis cache: %v", err)
		// Continue despite cache update failure
	}

	log.Printf("✅ Inventory updated for product %s: %d units",
		inventoryChange.ProductID, inventoryChange.Quantity)
	return nil
}

// handleOrderCreated processes OrderCreated events
func handleOrderCreated(ctx context.Context, data interface{}) error {
	order := models.Order{}
	if err := mapToStruct(data, &order); err != nil {
		return fmt.Errorf("invalid order data: %w", err)
	}

	// Transaction context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: Add to MongoDB
	_, err := db.OrderCollection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to insert order into MongoDB: %w", err)
	}

	// Step 2: Update customer order history
	orderHistoryEntry := models.OrderHistoryEntry{
		OrderID:     order.OrderID,
		OrderNumber: order.OrderNumber,
		Date:        order.Created,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
	}

	_, err = db.CustomerCollection.UpdateOne(
		ctx,
		bson.M{"customerId": order.CustomerID},
		bson.M{"$push": bson.M{"orderHistory": orderHistoryEntry}},
	)
	if err != nil {
		return fmt.Errorf("failed to update customer order history: %w", err)
	}

	// Step 3: Invalidate Redis customer orders cache
	cacheKey := "customer:" + order.CustomerID + ":orders"
	if err := cache.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("⚠️ Warning: Failed to invalidate Redis cache: %v", err)
		// Continue despite cache invalidation failure
	}

	log.Printf("✅ Order created: %s for customer %s", order.OrderID, order.CustomerID)
	return nil
}

// handleOrderStatusChanged processes OrderStatusChanged events
func handleOrderStatusChanged(ctx context.Context, data interface{}) error {
	statusChange := struct {
		OrderID string `json:"orderId"`
		Status  string `json:"status"`
	}{}
	if err := mapToStruct(data, &statusChange); err != nil {
		return fmt.Errorf("invalid order status data: %w", err)
	}

	// Transaction context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Step 1: Get order to retrieve customer ID
	var order models.Order
	err := db.OrderCollection.FindOne(ctx, bson.M{"orderId": statusChange.OrderID}).Decode(&order)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Step 2: Update order status in MongoDB
	_, err = db.OrderCollection.UpdateOne(
		ctx,
		bson.M{"orderId": statusChange.OrderID},
		bson.M{"$set": bson.M{
			"status":  statusChange.Status,
			"updated": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update order status in MongoDB: %w", err)
	}

	// Step 3: Update order status in customer order history
	_, err = db.CustomerCollection.UpdateOne(
		ctx,
		bson.M{
			"customerId":          order.CustomerID,
			"orderHistory.orderId": statusChange.OrderID,
		},
		bson.M{"$set": bson.M{
			"orderHistory.$.status": statusChange.Status,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update order status in customer history: %w", err)
	}

	// Step 4: Invalidate Redis caches
	pipe := cache.RedisClient.Pipeline()
	pipe.Del(ctx, "order:"+statusChange.OrderID)
	pipe.Del(ctx, "customer:"+order.CustomerID+":orders")
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("⚠️ Warning: Failed to invalidate Redis caches: %v", err)
		// Continue despite cache invalidation failure
	}

	log.Printf("✅ Order status updated: %s -> %s", statusChange.OrderID, statusChange.Status)
	return nil
}

func mapToStruct(data interface{}, target interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}
