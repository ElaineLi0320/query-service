package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"query-service/cache"
	"query-service/db"
	"query-service/messaging"
	"query-service/routes"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize connections
	db.InitMongo()
	cache.InitRedis()
	db.InitElasticsearch()

	// Configure and start Kafka consumer
	consumer := messaging.NewConsumer(
		[]string{"localhost:9092"},
		"query-service-events",
		"query-service-group",
		messaging.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 500 * time.Millisecond,
			MaxBackoff:     10 * time.Second,
			BackoffFactor:  2.0,
		},
	)

	// Register event handlers
	messaging.RegisterEventHandlers(consumer)

	// Start consumer in the background
	consumer.Start()

	// Create a new Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// API routes group
	api := r.Group("/api/queries")
	routes.RegisterRoutes(api)

	// Start the server in a goroutine
	go func() {
		if err := r.Run(":8081"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop the consumer
	consumer.Stop()

	// Create a deadline for shutdown
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Server gracefully stopped")
}
