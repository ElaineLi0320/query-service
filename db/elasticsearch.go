package db

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

var ElasticsearchClient *elasticsearch.Client

func InitElasticsearch() {
    cfg := elasticsearch.Config{
        Addresses: []string{"http://localhost:9200"},
    }
    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        log.Fatalf("Failed to create Elasticsearch client: %v", err)
    }

    ElasticsearchClient = client
    log.Println("✅ Elasticsearch initialized")
}
