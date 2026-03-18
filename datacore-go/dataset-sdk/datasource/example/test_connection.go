package main

import (
	"fmt"
	"log"

	"github.com/example/datasource"
)

func main() {
	fmt.Println("Testing datasource PostgreSQL connection...")

	ds, err := datasource.NewDataSource(datasource.DefaultCarrierDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer ds.Close()

	fmt.Println("Successfully connected to database!")

	fmt.Println("\nTesting Query method - getting PostgreSQL version...")
	results, err := ds.Query("SELECT version()")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Query results: %v\n", results)

	fmt.Println("\nConnection test passed!")
}
