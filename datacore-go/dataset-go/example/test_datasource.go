package main

import (
	"fmt"
	"log"

	"github.com/example/dataset-go/database"
)

func main() {
	fmt.Println("Testing dataset-go with datasource module...")

	db, err := database.NewDatabase("")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	fmt.Println("Successfully created database with datasource module!")

	fmt.Println("\nTesting Query method - getting PostgreSQL version...")
	results, err := db.Query("SELECT version()")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Query results: %v\n", results)

	fmt.Println("\nTest passed! dataset-go successfully uses datasource module for querying!")
}
