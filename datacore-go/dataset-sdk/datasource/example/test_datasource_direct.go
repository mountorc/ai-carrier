package main

import (
	"fmt"
	"github.com/example/datasource"
)

func main() {
	fmt.Println("=== Testing Direct Datasource ===")

	manager := datasource.GetManager()
	configJSON := `{"uuid":"post4bc7-9a41-4332-93a1-a60c4d8a7e19","type":"postgres","config":{"host":"121.43.142.153","port":5432,"database":"carrier","username":"carrier","password":"GNerfiSP4dpZjwcJ"}}`

	err := manager.AddDataSource(configJSON)
	if err != nil {
		fmt.Printf("Error adding datasource: %v\n", err)
		return
	}

	ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
	if err != nil {
		fmt.Printf("Error getting datasource: %v\n", err)
		return
	}

	result, err := ds.GetAutoSet("workflow1", "workflow")
	if err != nil {
		fmt.Printf("Error getting autoset: %v\n", err)
		return
	}

	fmt.Printf("\nResult keys: %#v\n", result)
	if autoset, ok := result["autoset"]; ok {
		fmt.Printf("\nAutoset type: %T\n", autoset)
		fmt.Printf("Autoset: %#v\n", autoset)
	}
}
