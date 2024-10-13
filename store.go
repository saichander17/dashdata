package main

import (
        "sync"
)

type Store struct {
        data map[string]string
        mu   sync.RWMutex
}

func NewStore() *Store {
        return &Store{
                data: make(map[string]string),
        }
}

func (s *Store) Set(key, value string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
        s.mu.RLock()
        defer s.mu.RUnlock()
        value, exists := s.data[key]
        return value, exists
}

func (s *Store) Delete(key string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        delete(s.data, key)
}
