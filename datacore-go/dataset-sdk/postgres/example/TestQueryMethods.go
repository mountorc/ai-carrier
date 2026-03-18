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

	fmt.Println("Testing Query method...")
	results, err := client.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' LIMIT 5")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("Query results (map format):\n")
	for i, row := range results {
		fmt.Printf("  %d: %v\n", i+1, row)
	}

	fmt.Println("\nTesting QueryJSON method...")
	jsonResult, err := client.QueryJSON("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' LIMIT 5")
	if err != nil {
		log.Fatalf("QueryJSON failed: %v", err)
	}
	fmt.Printf("QueryJSON results (JSON format):\n")
	fmt.Println(jsonResult)

	fmt.Println("\nAll tests passed!")
}
