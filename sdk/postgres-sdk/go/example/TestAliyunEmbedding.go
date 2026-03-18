package main

import (
	"fmt"
	"log"

	"github.com/example/postgres-sdk-go/postgres"
)

func main() {
	config := postgres.NewConfig("localhost", 5432, "postgres", "postgres", "postgres")
	client := postgres.NewClient(config, true, "", "")

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Testing Aliyun Embedding...")
	embedding, err := client.GetEmbedding("Hello, world!")
	if err != nil {
		log.Fatalf("Failed to get embedding: %v", err)
	}
	fmt.Printf("Embedding dimension: %d\n", len(embedding))
	fmt.Printf("First 5 values: %v\n", embedding[:5])

	fmt.Println("\nTesting text search...")
	collectionName := "test_aliyun_embedding"

	if err := client.CreateCollection(collectionName, 1024, "IVFFLAT"); err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	defer client.DropCollection(collectionName)

	texts := []string{
		"Hello, world!",
		"这是一个测试",
		"Machine learning is fun",
		"PostgreSQL with pgvector",
	}
	metadata := []map[string]interface{}{
		{"source": "test", "id": 1},
		{"source": "test", "id": 2},
		{"source": "test", "id": 3},
		{"source": "test", "id": 4},
	}

	if err := client.InsertTextWithEmbedding(collectionName, texts, metadata); err != nil {
		log.Fatalf("Failed to insert texts: %v", err)
	}

	results, err := client.SearchByText(collectionName, "Hello", 3, "")
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Println("\nSearch results:")
	for i, result := range results {
		fmt.Printf("%d. ID: %d, Score: %.4f, Metadata: %v\n", i+1, result.ID, result.Score, result.Metadata)
	}

	fmt.Println("\nAll tests passed!")
}
