package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"query-service/cache"
	"query-service/db"
	"query-service/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// getProductByID retrieves a product by its ID, with Redis caching
func getProductByID(c *gin.Context) {
	id := c.Param("productId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the product from Redis cache
	cacheKey := "product:" + id
	cachedProduct, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached product
		var product models.Product
		json.Unmarshal([]byte(cachedProduct), &product)
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": product})
		return
	}

	// If cache miss, query MongoDB
	var product models.Product
	err = db.ProductCollection.FindOne(ctx, bson.M{"productId": id}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Cache the product in Redis with a 10-minute expiration
	productJSON, _ := json.Marshal(product)
	cache.RedisClient.Set(ctx, cacheKey, productJSON, 10*time.Minute)

	// Return the product from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": product})
}

// getProductsByCategory retrieves products by category ID, with Redis caching
func getProductsByCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the products from Redis cache
	cacheKey := "products:category:" + categoryID
	cachedProducts, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached products
		var products []models.Product
		json.Unmarshal([]byte(cachedProducts), &products)
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": products})
		return
	}

	// If cache miss, query MongoDB
	var products []models.Product
	cursor, err := db.ProductCollection.Find(ctx, bson.M{"category.id": categoryID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product models.Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	// Cache the products in Redis with a 10-minute expiration
	productsJSON, _ := json.Marshal(products)
	cache.RedisClient.Set(ctx, cacheKey, productsJSON, 10*time.Minute)

	// Return the products from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": products})
}

// getInventory retrieves the current inventory for a product, with Redis caching
func getInventory(c *gin.Context) {
	productID := c.Param("productId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the inventory from Redis cache
	cacheKey := "inventory:" + productID
	cachedInventory, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached inventory
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": cachedInventory})
		return
	}

	// If cache miss, query MongoDB
	var product models.Product
	err = db.ProductCollection.FindOne(ctx, bson.M{"productId": productID}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Cache the inventory in Redis with a 10-minute expiration
	cache.RedisClient.Set(ctx, cacheKey, product.CurrentInventory, 10*time.Minute)

	// Return the inventory from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": product.CurrentInventory})
}

// getOrderByID retrieves an order by its ID, with Redis caching
func getOrderByID(c *gin.Context) {
	id := c.Param("orderId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the order from Redis cache
	cacheKey := "order:" + id
	cachedOrder, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached order
		var order models.Order
		json.Unmarshal([]byte(cachedOrder), &order)
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": order})
		return
	}

	// If cache miss, query MongoDB
	var order models.Order
	err = db.OrderCollection.FindOne(ctx, bson.M{"orderId": id}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Cache the order in Redis with a 10-minute expiration
	orderJSON, _ := json.Marshal(order)
	cache.RedisClient.Set(ctx, cacheKey, orderJSON, 10*time.Minute)

	// Return the order from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": order})
}

// getCustomerByID retrieves a customer by its ID, with Redis caching
func getCustomerByID(c *gin.Context) {
	id := c.Param("customerId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the customer from Redis cache
	cacheKey := "customer:" + id
	cachedCustomer, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached customer
		var customer models.Customer
		json.Unmarshal([]byte(cachedCustomer), &customer)
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": customer})
		return
	}

	// If cache miss, query MongoDB
	var customer models.Customer
	err = db.CustomerCollection.FindOne(ctx, bson.M{"customerId": id}).Decode(&customer)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Cache the customer in Redis with a 10-minute expiration
	customerJSON, _ := json.Marshal(customer)
	cache.RedisClient.Set(ctx, cacheKey, customerJSON, 10*time.Minute)

	// Return the customer from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": customer})
}

// getCustomerOrders retrieves a customer's order history, with Redis caching
func getCustomerOrders(c *gin.Context) {
	customerID := c.Param("customerId")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to retrieve the customer's orders from Redis cache
	cacheKey := "customer:" + customerID + ":orders"
	cachedOrders, err := cache.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached orders
		var orders []models.Order
		json.Unmarshal([]byte(cachedOrders), &orders)
		c.JSON(http.StatusOK, gin.H{"source": "cache", "data": orders})
		return
	}

	// If cache miss, query MongoDB
	var orders []models.Order
	cursor, err := db.OrderCollection.Find(ctx, bson.M{"customerId": customerID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var order models.Order
		cursor.Decode(&order)
		orders = append(orders, order)
	}

	// Cache the customer's orders in Redis with a 10-minute expiration
	ordersJSON, _ := json.Marshal(orders)
	cache.RedisClient.Set(ctx, cacheKey, ordersJSON, 10*time.Minute)

	// Return the orders from MongoDB
	c.JSON(http.StatusOK, gin.H{"source": "database", "data": orders})
}

// searchProducts searches for products using Elasticsearch
func searchProducts(c *gin.Context) {
	query := c.Query("q") // Get the search query from the request

	// Build the Elasticsearch query
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description", "category.name"},
			},
		},
	}

	// Serialize the query to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode search query"})
		return
	}

	// Perform the search request
	res, err := db.ElasticsearchClient.Search(
		db.ElasticsearchClient.Search.WithContext(context.Background()),
		db.ElasticsearchClient.Search.WithIndex("products"),
		db.ElasticsearchClient.Search.WithBody(&buf),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search query"})
		return
	}
	defer res.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse search response"})
		return
	}

	// Return the search results
	c.JSON(http.StatusOK, result)
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/products/:productId", getProductByID)
	r.GET("/products/category/:categoryId", getProductsByCategory)
	r.GET("/inventory/:productId", getInventory)
	r.GET("/orders/:orderId", getOrderByID)
	r.GET("/customers/:customerId", getCustomerByID)
	r.GET("/customers/:customerId/orders", getCustomerOrders)
	r.GET("/products/search", searchProducts) // Add search endpoint
}
