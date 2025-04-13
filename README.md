# Query Service

The Query Service is a microservice designed to handle read operations for an e-commerce platform. It integrates with MongoDB, Redis, Elasticsearch, and Kafka to provide efficient querying, caching, and event-driven data synchronization.

---

## Features

- **MongoDB**: Stores product, order, and customer data.
- **Redis**: Provides caching for frequently accessed data.
- **Elasticsearch**: Enables full-text search for products.
- **Kafka**: Consumes events to maintain data synchronization.
- **Gin Framework**: Provides RESTful APIs for querying data.

---

## Prerequisites

Ensure the following dependencies are installed on your system:

- [Go](https://golang.org/) (version 1.23 or higher)
- [MongoDB](https://www.mongodb.com/)
- [Redis](https://redis.io/)
- [Elasticsearch](https://www.elastic.co/)
- [Kafka](https://kafka.apache.org/)

---

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your-repo/query-service.git
   cd query-service
   ```

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

3. Start the required services:

   - **MongoDB**: Ensure MongoDB is running on `localhost:27017`.
   - **Redis**: Ensure Redis is running on `localhost:6379`.
   - **Elasticsearch**: Ensure Elasticsearch is running on `http://localhost:9200`.
   - **Kafka**: Ensure Kafka is running on `localhost:9092`.

4. Create Kafka topics:
   ```bash
   kafka-topics --create --topic query-service-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
   kafka-topics --create --topic query-service-events-dlq --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
   ```

---

## Usage

1. Initialize the database and Elasticsearch index:

   ```bash
   go run init/init_data.go
   ```

2. Start the Query Service:

   ```bash
   go run main.go
   ```

3. Test the service:
   - Health check:
     ```bash
     curl http://localhost:8081/health
     ```
   - Query a product by ID:
     ```bash
     curl http://localhost:8081/api/queries/products/p1001
     ```
   - Search for products:
     ```bash
     curl "http://localhost:8081/api/queries/products/search?q=Gin"
     ```

---

## Event Handling

The Query Service consumes events from Kafka to maintain data synchronization:

- **ProductCreated**: Adds a new product to MongoDB and Elasticsearch, invalidates Redis caches.
- **ProductUpdated**: Updates product details in MongoDB and Elasticsearch, invalidates Redis caches.
- **InventoryChanged**: Updates product inventory in MongoDB and Redis.
- **OrderCreated**: Adds a new order to MongoDB and updates the customer's order history.
- **OrderStatusChanged**: Updates the order status in MongoDB.

### Testing Kafka Events

1. Produce a test event:

   ```bash
   kafka-console-producer --broker-list localhost:9092 --topic query-service-events
   ```

2. Example events:

   - **ProductCreated**:
     ```json
     {
       "type": "ProductCreated",
       "data": {
         "productId": "p2001",
         "sku": "SKU-2001",
         "name": "Go Coffee Mug",
         "description": "Elegant coffee mug for gophers",
         "price": 19.99,
         "category": { "id": "c102", "name": "Accessories" },
         "currentInventory": 50
       }
     }
     ```
   - **InventoryChanged**:
     ```json
     {
       "type": "InventoryChanged",
       "data": { "productId": "p2001", "quantity": 45 }
     }
     ```

3. Check logs to verify event processing.

---

## Project Structure

```
query-service/
├── cache/                 # Redis initialization and utilities
│   └── redis.go
├── db/                    # MongoDB and Elasticsearch initialization
│   ├── mongo.go
│   └── elasticsearch.go
├── messaging/             # Kafka consumer and event handling
│   ├── kafka.go
│   ├── consumer.go
│   └── events.go
├── models/                # Data models for MongoDB
│   ├── products.go
│   ├── orders.go
│   └── customers.go
├── routes/                # API routes
│   └── api.go
├── init/                  # Initialization scripts
│   └── init_data.go
├── main.go                # Main entry point
├── go.mod                 # Go module dependencies
└── README.md              # Project documentation
```

---

## API Endpoints

### Products

- **Get Product by ID**: `GET /api/queries/products/:productId`
- **Get Products by Category**: `GET /api/queries/products/category/:categoryId`
- **Search Products**: `GET /api/queries/products/search?q=<query>`

### Inventory

- **Get Inventory**: `GET /api/queries/inventory/:productId`

### Orders

- **Get Order by ID**: `GET /api/queries/orders/:orderId`

### Customers

- **Get Customer by ID**: `GET /api/queries/customers/:customerId`
- **Get Customer Orders**: `GET /api/queries/customers/:customerId/orders`

---

## Configuration

- **MongoDB**: Update the connection URI in `db/mongo.go` if needed.
- **Redis**: Update the Redis configuration in `cache/redis.go`.
- **Elasticsearch**: Update the Elasticsearch address in `db/elasticsearch.go`.
- **Kafka**: Update the Kafka broker list in `messaging/consumer.go`.

---

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
