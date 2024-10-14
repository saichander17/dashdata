package store

import "sync"
import "github.com/saichander17/dashdata/internal/wal"

type SimpleStore struct {
	data map[string]string
	mu   sync.RWMutex
	wal  *wal.WAL
}

func NewSimpleStore() *SimpleStore {
	return &SimpleStore{
		data: make(map[string]string),
	}
}

func (s *SimpleStore) SetWAL(wal *wal.WAL) {
    s.wal = wal
}

func (s *SimpleStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
    if s.wal != nil {
        s.wal.Log("SET", key, value)
    }
	s.data[key] = value
}

func (s *SimpleStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

func (s *SimpleStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
    if s.wal != nil {
        s.wal.Log("DELETE", key, "")
    }
	delete(s.data, key)
}

func (s *SimpleStore) GetAll() map[string]string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    result := make(map[string]string)
    for k, v := range s.data {
        result[k] = v
    }
    return result
}
