package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/example/dataset-go/database"
	"github.com/example/dataset-go/scheduler"
	"github.com/example/dataset-go/server"
)

func main() {
	var filePath string
	var addr string
	var dsn string

	flag.StringVar(&filePath, "file", "", "Path to SQL config JSON file")
	flag.StringVar(&addr, "addr", ":8084", "Server address to listen on")
	flag.StringVar(&dsn, "dsn", "", "Database DSN (default: carrier database)")
	flag.Parse()

	if filePath == "" {
		log.Fatal("Please provide a SQL config file using -file flag")
	}

	fmt.Println("Initializing SQL Scheduler Core...")

	sched := scheduler.NewScheduler()
	if err := sched.LoadFromFile(filePath); err != nil {
		log.Fatalf("Failed to load SQL config: %v", err)
	}

	fmt.Printf("Loaded %d SQL statements\n", sched.Count())

	db, err := database.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	srv := server.NewServer(sched, db)

	if err := srv.Start(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
