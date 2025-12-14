package main

import (
	"fmt"
	"log"
	"os"

	"glog/domain"
	"glog/tui"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("GLOG_DB_PATH")
	if dbPath == "" {
		dbPath = "glog.db"
	}

	// Open the document store
	store, err := domain.NewDocumentStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open document store: %v", err)
	}
	defer store.Close()

	// Run the browser TUI
	if err := tui.RunBrowser(store); err != nil {
		fmt.Printf("Error running browser: %v\n", err)
		os.Exit(1)
	}
}
