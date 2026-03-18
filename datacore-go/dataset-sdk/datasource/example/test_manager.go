package main

import (
	"fmt"
	"log"

	"github.com/example/datasource"
)

func main() {
	fmt.Println("Testing datasource manager with JSON config...")

	configJSON := `{"uuid":"uuid_datasource_post4bc7-9a41-4332-93a1-a60c4d8a7e19","type":"postgres","config":{"host":"http://121.43.142.153","port":5432,"charset":"utf-8","database":"carrier","password":"GNerfiSP4dpZjwcJ","username":"carrier","driver_class":"org.postgresql.Driver"}}`

	manager := datasource.GetManager()

	fmt.Println("\n1. Adding datasource...")
	err := manager.AddDataSource(configJSON)
	if err != nil {
		log.Fatalf("Failed to add datasource: %v", err)
	}
	fmt.Println("Datasource added successfully!")

	uuidDatasource := "uuid_datasource_post4bc7-9a41-4332-93a1-a60c4d8a7e19"

	fmt.Println("\n2. Testing QueryWithUUID...")
	results, err := manager.QueryWithUUID(uuidDatasource, "SELECT version()")
	if err != nil {
		log.Fatalf("QueryWithUUID failed: %v", err)
	}
	fmt.Printf("Query results: %v\n", results)

	fmt.Println("\n3. Getting datasource and using directly...")
	ds, err := manager.GetDataSource(uuidDatasource)
	if err != nil {
		log.Fatalf("Failed to get datasource: %v", err)
	}

	fmt.Println("Testing direct Query...")
	results2, err := ds.Query("SELECT current_database()")
	if err != nil {
		log.Fatalf("Direct query failed: %v", err)
	}
	fmt.Printf("Direct query results: %v\n", results2)

	fmt.Println("\n4. Removing datasource...")
	manager.RemoveDataSource(uuidDatasource)
	fmt.Println("Datasource removed!")

	fmt.Println("\nAll tests passed! Datasource manager works correctly!")
}
