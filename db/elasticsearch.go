package db

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

var ElasticsearchClient *elasticsearch.Client

func InitElasticsearch() {
    cfg := elasticsearch.Config{
        Addresses: []string{
            "http://localhost:9200", // Elasticsearch server address
        },
    }

    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        log.Fatalf("Error creating Elasticsearch client: %v", err)
    }

    // Test the connection
    res, err := client.Info()
    if err != nil {
        log.Fatalf("Error connecting to Elasticsearch: %v", err)
    }
    defer res.Body.Close()

    log.Println("✅ Connected to Elasticsearch!")
    ElasticsearchClient = client
}
