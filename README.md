# Query Service

The Query Service is a microservice designed to handle read operations for an e-commerce platform. It integrates with MongoDB, Redis, Elasticsearch, and Kafka to provide efficient querying, caching, and event-driven data synchronization.

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-repo/query-service.git
cd query-service
```

### 2. Start Services with Docker Compose

```bash
docker compose up -d
```

This will start MongoDB, Redis, Elasticsearch, Kafka, and Zookeeper.

### 3. Create Kafka Topics

```bash
docker exec -it kafka kafka-topics --create --topic query-service-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
docker exec -it kafka kafka-topics --create --topic query-service-events-dlq --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

### 4. Run the API

```bash
go run main.go
```
