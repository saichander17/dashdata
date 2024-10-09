package main

import (
	"hash/fnv"
	"sync"
)

const lockCount = 1024 // Number of locks, can be adjusted based on expected concurrency

type Store struct {
	data  map[string]string
	locks [lockCount]sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) lockIndex(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32() % lockCount
}

func (s *Store) Set(key, value string) {
	lockIndex := s.lockIndex(key)
	s.locks[lockIndex].Lock()
	defer s.locks[lockIndex].Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	lockIndex := s.lockIndex(key)
	s.locks[lockIndex].RLock()
	defer s.locks[lockIndex].RUnlock()
	value, exists := s.data[key]
	return value, exists
}

func (s *Store) Delete(key string) {
	lockIndex := s.lockIndex(key)
	s.locks[lockIndex].Lock()
	defer s.locks[lockIndex].Unlock()
	delete(s.data, key)
}
