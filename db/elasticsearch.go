package db

import (
	"bytes"
	"encoding/json"
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

    createProductIndex()
}

func createProductIndex() {
    mapping := map[string]interface{}{
        "settings": map[string]interface{}{
            "analysis": map[string]interface{}{
                "analyzer": map[string]interface{}{
                    "custom_analyzer": map[string]interface{}{
                        "type":      "custom",
                        "tokenizer": "standard",
                        "filter":    []string{"lowercase", "asciifolding"},
                    },
                },
            },
        },
        "mappings": map[string]interface{}{
            "properties": map[string]interface{}{
                "name": map[string]interface{}{
                    "type":     "text",
                    "analyzer": "custom_analyzer",
                },
                "sku": map[string]interface{}{
                    "type": "keyword",
                },
                "category": map[string]interface{}{
                    "type": "keyword",
                },
                "price": map[string]interface{}{
                    "type": "double",
                },
            },
        },
    }

    body, _ := json.Marshal(mapping)
    res, err := ElasticsearchClient.Indices.Create(
        "products",
        ElasticsearchClient.Indices.Create.WithBody(bytes.NewReader(body)),
    )
    if err != nil {
        log.Fatalf("Failed to create Elasticsearch index: %v", err)
    }
    defer res.Body.Close()
    log.Println("✅ Elasticsearch product index created")
}
