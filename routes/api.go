package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"query-service/cache"
	"query-service/db"
	"query-service/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// getProductsByCategory retrieves products by category ID, with pagination
func getProductsByCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query MongoDB with pagination
	var products []models.Product
	cursor, err := db.ProductCollection.Find(ctx, bson.M{"category.id": categoryID}, options.Find().SetSkip(int64((page-1)*size)).SetLimit(int64(size)))
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

	// Return the products
	c.JSON(http.StatusOK, gin.H{"data": products, "page": page, "size": size})
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

// getCustomerOrders retrieves a customer's order history, with pagination
func getCustomerOrders(c *gin.Context) {
	customerID := c.Param("customerId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query MongoDB with pagination
	var orders []models.Order
	cursor, err := db.OrderCollection.Find(ctx, bson.M{"customerId": customerID}, options.Find().SetSkip(int64((page-1)*size)).SetLimit(int64(size)))
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

	// Return the orders
	c.JSON(http.StatusOK, gin.H{"data": orders, "page": page, "size": size})
}

// searchProducts searches for products using Elasticsearch
func searchProducts(c *gin.Context) {
	query := c.Query("q")
	categoryID := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// Build Elasticsearch query with pagination and category filter
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"name", "description", "category.name"},
					}},
				},
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"category.id": categoryID}},
				},
			},
		},
		"from": (page - 1) * size,
		"size": size,
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
	r.GET("/products/search", searchProducts)
}
