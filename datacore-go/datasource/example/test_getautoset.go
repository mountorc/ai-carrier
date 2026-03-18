package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/example/datasource"
)

func main() {
	manager := datasource.GetManager()

	configJSON := `{
		"uuid": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
		"type": "postgres",
		"config": {
			"host": "121.43.142.153",
			"port": 5432,
			"database": "carrier",
			"username": "carrier",
			"password": "GNerfiSP4dpZjwcJ"
		}
	}`

	err := manager.AddDataSource(configJSON)
	if err != nil {
		log.Fatalf("Failed to add datasource: %v", err)
	}
	fmt.Println("Datasource added successfully")

	ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
	if err != nil {
		log.Fatalf("Failed to get datasource: %v", err)
	}

	result, err := ds.GetAutoSet("workflow1", "workflow")
	if err != nil {
		log.Fatalf("Failed to get autoset: %v", err)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}

	fmt.Println("Result:")
	fmt.Println(string(jsonResult))
}
