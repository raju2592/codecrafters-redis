package store

import (
	"hash/fnv"
	"math/rand"
	"sync"
	"time"
)

const N_SHARD = 64
const N_PASSIVE_WORKER = 3
const ACTIVE_EXPIRATION_PERIOD = 500
const N_RANDOM_SAMPLE_SIZE = 20

type keyShard struct {
	mu sync.Mutex
	keyWithExpMap map[string]int	
	keyWithExp []string
}

type valueWithMeta struct {
	value []byte
	expiresAt time.Time
}

var data sync.Map
var shards []keyShard = make([]keyShard, N_SHARD)

func handleExpiredKeys() {
	for {
		key := <- expiredKeys
		deleteIfExpired(key)
	}
}

func init() {
	for i := range shards {
		shards[i].keyWithExpMap = make(map[string]int)
	}

	// spawn passing key expiration workers
	for i := 0; i < N_PASSIVE_WORKER; i++ {
		go handleExpiredKeys()
	}

	// spawn active key expiration goroutine
	go doRandomExpiration()
}

func doRandomExpiration() {
	ticker := time.NewTicker(ACTIVE_EXPIRATION_PERIOD * time.Millisecond)

	for range ticker.C {
		keys := []string{}
		taken := 0
		for i := 0; taken < N_RANDOM_SAMPLE_SIZE && i < N_RANDOM_SAMPLE_SIZE * 2; i++ {
			shardIndex := rand.Intn(N_SHARD)
			shard := &shards[shardIndex]

			shard.mu.Lock()

			if len(shard.keyWithExp) == 0 {
				shard.mu.Unlock()
				continue
			}

			keyIndex := rand.Intn(len(shard.keyWithExp))
			keys = append(keys, shard.keyWithExp[keyIndex])
			taken++

			shard.mu.Unlock()
		}

		for _, key := range keys {
			deleteIfExpired(key)
		}
	}

}

var expiredKeys chan string = make(chan string, 1000)

func getShardIndex(key string) int {
    h := fnv.New32a()
    h.Write([]byte(key))
    return int(h.Sum32() % N_SHARD)
}

func getShardLock(key string) int {
	i := getShardIndex(key)
	shards[i].mu.Lock()
	return i
}

func releaseShardLock(i int) {
	shards[i].mu.Unlock()
}

// this assumes shard lock is already locked
func addKeyWithExp(key string, shardIndex int) {
	shard := &shards[shardIndex]
	_, ok := shard.keyWithExpMap[key]
	if ok {
		return
	}

	shard.keyWithExp = append(shard.keyWithExp, key)
	shard.keyWithExpMap[key] = len(shard.keyWithExp) - 1
}

func removeKeyWithExp(key string, shardIndex int) {
	shard := &shards[shardIndex]
	i, ok := shard.keyWithExpMap[key]
	if !ok {
		return
	}

	n := len(shard.keyWithExp)
	swappedKey := shard.keyWithExp[n - 1]

	shard.keyWithExp[i] = shard.keyWithExp[n - 1]
	shard.keyWithExp = shard.keyWithExp[:n - 1]

	if n != 1 {
		shard.keyWithExpMap[swappedKey] = i
	}
	delete(shard.keyWithExpMap, key)
}

func hasExpired(m time.Time) bool{
	return !m.IsZero() && m.Before(time.Now())
}

func deleteIfExpired(key string) {
	i := getShardLock(key)
	defer releaseShardLock(i)
	v, ok := data.Load(key)
	if !ok {
		return
	}
	vm := v.(valueWithMeta)
	if hasExpired(vm.expiresAt) {
		data.Delete(key)
		removeKeyWithExp(key, i)
	}
}

type SetOptions struct {
	Ttl int
}

func Set(key string, value []byte, opt SetOptions) {
	i := getShardLock(key)
	defer releaseShardLock(i)

	vm := valueWithMeta{
		value: value,
	}

	if opt.Ttl > 0 {
		vm.expiresAt = time.Now().Add(time.Duration(opt.Ttl) * time.Millisecond)
	}

	data.Store(key, vm)

	if opt.Ttl > 0 {
		addKeyWithExp(key, i)
	} else {
		removeKeyWithExp(key, i)
	}
}


func queueExpiration(key string) {
	select {
	case expiredKeys <- key:
	default:
	}
}

func Get(key string) ([]byte, bool) {
	v, ok := data.Load(key)
	if ok {
		vm := v.(valueWithMeta)
		if hasExpired(vm.expiresAt) {
			queueExpiration(key)
			return nil, false
		}
		return vm.value, ok
	} else {
		return nil, false
	}
}
