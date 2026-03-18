package main

import (
	"fmt"
	"log"

	"github.com/example/postgres-sdk-go/postgres"
)

func main() {
	client, err := postgres.NewClientFromURL("postgresql://carrier:GNerfiSP4dpZjwcJ@121.43.142.153:5432/carrier")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("\n=== List collections ===")
	collections, err := client.ListCollections()
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}
	fmt.Printf("Collections: %v\n", collections)

	collectionName := "test_vectors_go"
	dimension := 128

	fmt.Printf("\n=== Create collection %s with dimension %d ===\n", collectionName, dimension)
	err = client.CreateCollection(collectionName, dimension, "IVFFLAT")
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	fmt.Println("\n=== Insert vectors ===")
	vectors := make([][]float64, 3)
	for i := 0; i < 3; i++ {
		vec := make([]float64, dimension)
		for j := 0; j < dimension; j++ {
			vec[j] = float64(i+1) * 0.1 * float64(j)
		}
		vectors[i] = vec
	}

	metadata := []map[string]interface{}{
		{"text": "Hello world", "category": "test"},
		{"text": "Go SDK", "category": "sdk"},
		{"text": "PostgreSQL vector", "category": "database"},
	}

	err = client.InsertVectors(collectionName, vectors, metadata)
	if err != nil {
		log.Fatalf("Failed to insert vectors: %v", err)
	}

	fmt.Println("\n=== Search vectors ===")
	queryVector := make([]float64, dimension)
	for j := 0; j < dimension; j++ {
		queryVector[j] = 0.15 * float64(j)
	}

	results, err := client.SearchVectors(collectionName, queryVector, 3, "")
	if err != nil {
		log.Fatalf("Failed to search vectors: %v", err)
	}

	for _, result := range results {
		fmt.Printf("ID: %d, Score: %.4f, Metadata: %v\n", result.ID, result.Score, result.Metadata)
	}

	fmt.Println("\n=== Get table list ===")
	tables, err := client.GetTableList("")
	if err != nil {
		log.Fatalf("Failed to get table list: %v", err)
	}
	fmt.Printf("Tables: %v\n", tables)

	if len(tables) > 0 {
		fmt.Printf("\n=== Get fields for table %s ===\n", tables[0])
		fields, err := client.GetTableFields("public", tables[0])
		if err != nil {
			log.Fatalf("Failed to get table fields: %v", err)
		}
		fmt.Printf("Fields: %v\n", fields)
	}

	fmt.Println("\n=== Drop collection ===")
	err = client.DropCollection(collectionName)
	if err != nil {
		log.Fatalf("Failed to drop collection: %v", err)
	}
}
