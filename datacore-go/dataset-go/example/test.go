package main

import (
	"fmt"
	"log"

	"github.com/example/dataset-go/scheduler"
)

func main() {
	sched := scheduler.NewScheduler()

	fmt.Println("Testing SQL Scheduler Core...")
	fmt.Println("=")

	err := sched.LoadFromFile("/Users/a1-6/Documents/code/trae/autoFlow/services/mountcore/scheduler/sql_scheduler.json")
	if err != nil {
		log.Fatalf("Failed to load: %v", err)
	}

	fmt.Printf("Loaded %d SQL statements\n", sched.Count())

	fmt.Println("\nTesting GetSQL...")
	testUUID := "550e8400-e29b-41d4-a716-446655440001"
	sqlStr, exists := sched.GetSQL(testUUID)
	if exists {
		fmt.Printf("SQL for uuid %s:\n%s\n", testUUID, sqlStr)
	} else {
		fmt.Printf("SQL not found for uuid: %s\n", testUUID)
	}

	fmt.Println("\nTesting GetSQLItem...")
	item, exists := sched.GetSQLItem(testUUID)
	if exists {
		fmt.Printf("Item for uuid %s:\n", testUUID)
		fmt.Printf("  Name: %s\n", item.Name)
		fmt.Printf("  Description: %s\n", item.Description)
		fmt.Printf("  SQL: %s\n", item.SQL)
	}

	fmt.Println("\nTesting ListAllUUIDs...")
	uuids := sched.ListAllUUIDs()
	fmt.Printf("All UUIDs (%d):\n", len(uuids))
	for i, uuid := range uuids {
		fmt.Printf("  %d. %s\n", i+1, uuid)
	}

	fmt.Println("\nAll tests passed!")
}
