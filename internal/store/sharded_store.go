package store

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"github.com/saichander17/dashdata/internal/wal"
)

const shardCount = 1024

type value struct {
	data    atomic.Value
	writeMu sync.Mutex
}

type ShardedStore struct {
	shards [shardCount]map[string]*value
	locks  [shardCount]sync.RWMutex
	wal  *wal.WAL
}

func (s *ShardedStore) SetWAL(wal *wal.WAL) {
    s.wal = wal
}

func NewShardedStore() *ShardedStore {
	s := &ShardedStore{}
	for i := range s.shards {
		s.shards[i] = make(map[string]*value)
	}
	return s
}

func (s *ShardedStore) shardIndex(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32() % shardCount
}

func (s *ShardedStore) Set(key, val string) {
	index := s.shardIndex(key)
	s.locks[index].RLock()
	v, exists := s.shards[index][key]
	s.locks[index].RUnlock()

	if !exists {
		s.locks[index].Lock()
		v, exists = s.shards[index][key]
		if !exists {
			v = &value{}
			s.shards[index][key] = v
		}
		s.locks[index].Unlock()
	}

	v.writeMu.Lock()
    if s.wal != nil {
        s.wal.Log("SET", key, val)
    }
	v.data.Store(val)
	v.writeMu.Unlock()
}

func (s *ShardedStore) Get(key string) (string, bool) {
	index := s.shardIndex(key)
	s.locks[index].RLock()
	v, exists := s.shards[index][key]
	s.locks[index].RUnlock()

	if !exists {
		return "", false
	}

	return v.data.Load().(string), true
}

func (s *ShardedStore) Delete(key string) {
	index := s.shardIndex(key)
	s.locks[index].Lock()
    if s.wal != nil {
        s.wal.Log("DELETE", key, "")
    }
	delete(s.shards[index], key)
	s.locks[index].Unlock()
}

func (s *ShardedStore) GetAll() map[string]string {
    result := make(map[string]string)
    for i := range s.shards {
        s.locks[i].RLock()
        for k, v := range s.shards[i] {
            result[k] = v.data.Load().(string)
        }
        s.locks[i].RUnlock()
    }
    return result
}
