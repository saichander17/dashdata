package store

import "github.com/saichander17/dashdata/internal/wal"

// Store interface defines the methods that any store implementation must provide
type Store interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Delete(key string)
	GetAll() map[string]string
	SetWAL(wal *wal.WAL)
}

// Ensure that Store implements wal.StoreOperations
var _ wal.StoreOperations = (Store)(nil)
