package store

import "sync"

type SimpleStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewSimpleStore() *SimpleStore {
	return &SimpleStore{
		data: make(map[string]string),
	}
}

func (s *SimpleStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	delete(s.data, key)
}
