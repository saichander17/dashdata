package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/saichander17/dashdata/internal/server"
	"github.com/saichander17/dashdata/internal/store"
	"github.com/saichander17/dashdata/internal/persistence"
	"github.com/saichander17/dashdata/internal/wal"
)

type Config struct {
    PersistenceFile string
    PersistenceInterval time.Duration
    WALFile string
}

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

    persistenceConfig := Config{
        PersistenceFile: "dashdata.gob",
        PersistenceInterval: 5 * time.Minute,
        WALFile: "dashdata.wal",
    }
    persister, wal := setupPersistence(persistenceConfig, dataStore)
    defer wal.Close()
    defer persister.SaveToDisk()

	srv := server.NewServer(dataStore, fmt.Sprintf("%d", *port), 1000, 10000, 5*time.Second)
	log.Printf("Server starting on port %d", *port)
	err := srv.Start()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setupPersistence(config Config, store store.Store) (*persistence.Persister, *wal.WAL) {
    persister := persistence.NewPersister(store, config.PersistenceFile, config.PersistenceInterval)
    wal, err := wal.NewWAL(config.WALFile)
    if err != nil {
        log.Fatalf("Failed to create WAL: %v", err)
    }

    store.SetWAL(wal)

    // Load data from disk
    if err := persister.LoadFromDisk(); err != nil {
        log.Printf("Failed to load data from disk: %v", err)
    }

    // Start periodic persistence
    persister.Start()

    return persister, wal
}
