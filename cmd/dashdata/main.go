package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/saichander17/dashdata/internal/server"
	"github.com/saichander17/dashdata/internal/store"
)

func main() {
	storeType := flag.String("store", "simple", "Type of store to use: 'simple' or 'sharded'")
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	var dataStore store.Store

	switch *storeType {
	case "simple":
		dataStore = store.NewSimpleStore()
		fmt.Println("Using SimpleStore")
	case "sharded":
		dataStore = store.NewShardedStore()
		fmt.Println("Using ShardedStore")
	default:
		fmt.Println("Invalid store type. Using SimpleStore as default.")
		dataStore = store.NewSimpleStore()
	}

	srv := server.NewServer(dataStore, fmt.Sprintf("%d", *port), 100, 1000, 5*time.Second)
	log.Printf("Server starting on port %d", *port)
	err := srv.Start()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
