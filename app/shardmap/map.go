package shardmap

import "sync"

const defaultShardCount = 64

type shard[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

type ShardedMap[K comparable, V any] struct {
	shards []shard[K, V]
	count  int
	hashFn func(K) uint32
}

func New[K comparable, V any](hashFn func(K) uint32) *ShardedMap[K, V] {
	sm := &ShardedMap[K, V]{
		shards: make([]shard[K, V], defaultShardCount),
		count:  defaultShardCount,
		hashFn: hashFn,
	}
	for i := range sm.shards {
		sm.shards[i].data = make(map[K]V)
	}
	return sm
}

func (sm *ShardedMap[K, V]) getShard(key K) *shard[K, V] {
	return &sm.shards[int(sm.hashFn(key)%uint32(sm.count))]
}

func (sm *ShardedMap[K, V]) Set(key K, value V) {
	s := sm.getShard(key)
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (sm *ShardedMap[K, V]) Get(key K) (V, bool) {
	s := sm.getShard(key)
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	return v, ok
}

func (sm *ShardedMap[K, V]) Update(key K, fn func(old V, exists bool) V) {
	s := sm.getShard(key)
	s.mu.Lock()
	old, exists := s.data[key]
	s.data[key] = fn(old, exists)
	s.mu.Unlock()
}

func (sm *ShardedMap[K, V]) Delete(key K) {
	s := sm.getShard(key)
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}
