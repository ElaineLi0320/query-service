package messaging

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
    reader      *kafka.Reader
    handlers    map[string]EventHandler
    retryConfig RetryConfig
    dlqWriter   *kafka.Writer
    wg          sync.WaitGroup
    stopChan    chan struct{}
}

type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    BackoffFactor  float64
}

type EventHandler func(context.Context, interface{}) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, retryConfig RetryConfig) *Consumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"localhost:9092"},
        Topic:   topic,
        GroupID: groupID,
        // Set maximum wait time for new messages
        MaxWait: 30 * time.Second,
    })

    // Setup DLQ writer
    dlqWriter := &kafka.Writer{
        Addr:         kafka.TCP(brokers...),
        RequiredAcks: kafka.RequireAll,
    }

    return &Consumer{
        reader:      reader,
        handlers:    make(map[string]EventHandler),
        retryConfig: retryConfig,
        dlqWriter:   dlqWriter,
        stopChan:    make(chan struct{}),
    }
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
    c.handlers[eventType] = handler
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start() {
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        for {
            select {
            case <-c.stopChan:
                log.Println("📢 Consumer shutting down...")
                return
            default:
                c.processMessage()
            }
        }
    }()
    log.Println("✅ Kafka consumer started")
}

// processMessage reads and processes a single message from Kafka
func (c *Consumer) processMessage() {
    ctx := context.Background()
    msg, err := c.reader.ReadMessage(ctx)
    if err != nil {
        log.Printf("⚠️ Failed to read Kafka message: %v", err)
        time.Sleep(1 * time.Second) // Wait before retrying
        return
    }

    log.Printf("📩 Received message: partition=%d offset=%d key=%s", 
        msg.Partition, msg.Offset, string(msg.Key))

    var event Event
    if err := json.Unmarshal(msg.Value, &event); err != nil {
        log.Printf("⚠️ Failed to parse Kafka message: %v", err)
        c.sendToDLQ(msg, "parse_error", err.Error())
        return
    }

    handler, exists := c.handlers[event.Type]
    if !exists {
        log.Printf("⚠️ No handler registered for event type: %s", event.Type)
        c.sendToDLQ(msg, "no_handler", "No handler registered for this event type")
        return
    }

    if err := c.processWithRetry(ctx, handler, event.Data); err != nil {
        log.Printf("❌ Failed to process event after retries: %v", err)
        c.sendToDLQ(msg, "processing_error", err.Error())
        return
    }

    log.Printf("✅ Successfully processed event of type: %s", event.Type)
}

// processWithRetry attempts to process an event with exponential backoff retry
func (c *Consumer) processWithRetry(ctx context.Context, handler EventHandler, data interface{}) error {
    var lastErr error
    backoff := c.retryConfig.InitialBackoff

    // Try to process the event up to MaxRetries times
    for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
        // If this is a retry, log and wait
        if attempt > 0 {
            log.Printf("🔄 Retry attempt %d/%d after error: %v", 
                attempt, c.retryConfig.MaxRetries, lastErr)
            time.Sleep(backoff)
            backoff = time.Duration(float64(backoff) * c.retryConfig.BackoffFactor)
            if backoff > c.retryConfig.MaxBackoff {
                backoff = c.retryConfig.MaxBackoff
            }
        }

        // Process the event
        err := handler(ctx, data)
        if err == nil {
            return nil // Success
        }

        lastErr = err
    }

    return lastErr
}

// sendToDLQ sends a failed message to the Dead Letter Queue
func (c *Consumer) sendToDLQ(msg kafka.Message, errorType, errorDetail string) {
    msg.Topic = c.reader.Config().Topic + "-dlq"
    // Add error information to message headers
    msg.Headers = append(msg.Headers,
        kafka.Header{Key: "error_type", Value: []byte(errorType)},
        kafka.Header{Key: "error_detail", Value: []byte(errorDetail)},
        kafka.Header{Key: "original_topic", Value: []byte(c.reader.Config().Topic)},
        kafka.Header{Key: "failed_at", Value: []byte(time.Now().Format(time.RFC3339))},
    )

    // Write to DLQ
    err := c.dlqWriter.WriteMessages(context.Background(), msg)
    if err != nil {
        log.Printf("⚠️ Failed to send message to DLQ: %v", err)
    } else {
        log.Printf("📝 Message sent to DLQ: %s", errorType)
    }
}

// Stop gracefully shuts down the consumer
func (c *Consumer) Stop() {
    close(c.stopChan)
    c.wg.Wait()
    
    if err := c.reader.Close(); err != nil {
        log.Printf("⚠️ Error closing Kafka reader: %v", err)
    }
    
    if err := c.dlqWriter.Close(); err != nil {
        log.Printf("⚠️ Error closing DLQ writer: %v", err)
    }
    
    log.Println("✅ Kafka consumer stopped")
}
