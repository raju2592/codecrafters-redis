package commands

import (
	"cmp"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"slices"
	"strconv"
	"strings"
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
	value interface{}
	expiresAt time.Time
	ttype ValueType
}

var data sync.Map
var shards []keyShard = make([]keyShard, N_SHARD)
var watchers sync.Map

func handleExpiredKeys() {
	for {
		key := <- expiredKeys
		DeleteIfExpired(key)
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
			DeleteIfExpired(key)
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

func DeleteIfExpired(key string) {
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
		dirtyWatchers(key)
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
	dirtyWatchers(key)
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
		return vm.value.([]byte), ok
	} else {
		return nil, false
	}
}


func GetType(key string) string {
	v, ok := data.Load(key)
	if ok {
		vm := v.(valueWithMeta)
		if hasExpired(vm.expiresAt) {
			queueExpiration(key)
			return TypeNone.String()
		}
		return vm.ttype.String()
	} else {
		return TypeNone.String()
	}
}

func Incr(key string) (int, error) {
	i := getShardLock(key)
	defer releaseShardLock(i)

	v, ok := data.Load(key)

	intVal := 1
	var exp time.Time


	if ok {
		vm := v.(valueWithMeta)
		if hasExpired(vm.expiresAt) {
			queueExpiration(key)
		} else {
			v := vm.value.([]byte)
			exp = vm.expiresAt

			vInt, err := strconv.Atoi(string(v))

			if err != nil {
				return 0, err
			}

			intVal = vInt + 1
		}
	}

	data.Store(key, valueWithMeta{
		value: []byte(strconv.Itoa(intVal)),
		expiresAt: exp,
	})

	dirtyWatchers(key)

	return intVal, nil
}

// this assumes the key's shard lock is locked
func dirtyWatchers(key string) {
	v, ok := watchers.Load(key)
	if !ok {
		return
	}
	for _, conn := range v.([]*ConnMeta) {
		conn.dirty.Store(true)
	}
}

func Watch(key string, conn *ConnMeta) {
	i := getShardLock(key)
	defer releaseShardLock(i)

	var wl []*ConnMeta
	if v, ok := watchers.Load(key); ok {
		wl = v.([]*ConnMeta)
	}
	wl = append(wl, conn)
	watchers.Store(key, wl)
}

func Unwatch(key string, conn *ConnMeta) {
	i := getShardLock(key)
	defer releaseShardLock(i)

	v, ok := watchers.Load(key)
	if !ok {
		return
	}
	wl := v.([]*ConnMeta)

	del := false
	for i, w := range wl {
		if w == conn {
			del = true
			wl[i] = wl[len(wl) - 1]
		}
	}

	if !del {
		return
	}

	if len(wl) == 1 {
		watchers.Delete(key)
	} else {
		wl = wl[:len(wl) - 1]
		watchers.Store(key, wl)
	}
}

func LockKeys(keys []string) func() {
	indices := make([]int, len(keys))
	for i, key := range keys {
		pos := getShardIndex(key)
		indices[i] = pos
	}

	slices.Sort(indices)
	indices = slices.Compact(indices)

	for _, pos := range indices {
		shards[pos].mu.Lock()
	}

	return func ()  {
		for _, pos := range indices {
			shards[pos].mu.Unlock()
		}
	}
}

type StreamEntryId struct {
	ms int64;
	seq int64;
}

func (id StreamEntryId) Compare(b StreamEntryId) int {
	if id.ms != b.ms {
			return cmp.Compare(id.ms, b.ms)
	}
	return cmp.Compare(id.seq, b.seq)
}

func (id StreamEntryId) String() string {
	ms := strconv.FormatInt(id.ms, 10)
	seq := strconv.FormatInt(id.seq, 10)
	return ms + "-" + seq
}

type StreamEntry struct {
	id StreamEntryId
	kv map[string][]byte
}

type XaddOptions struct {
	entryId string;
	kv map[string][]byte
}

func loadValue(key string) (valueWithMeta, bool) {
	v, ok := data.Load(key)

	if ok {
		vm := v.(valueWithMeta)
		return vm, true
	}

	return valueWithMeta{}, false
}


// Fully auto-generate both parts, pass -1 for timestamp
func generateEntryId(entries []StreamEntry, timestamp int64) (StreamEntryId, bool) {
	n := len(entries)

	if n == 0 {
		if timestamp == -1 {
			timestamp = time.Now().UnixMilli();
		}

		seq := 0
		if timestamp == 0 {
			seq = 1
		}

		return StreamEntryId{ms: timestamp, seq: int64(seq)}, true 
	}

	lastEntryId := entries[n - 1].id
	
	if timestamp == -1 {
		curTs := time.Now().UnixMilli()

		newMs := max(curTs, lastEntryId.ms)
		newSeq := int64(0)
		if lastEntryId.ms == newMs {
			newSeq = lastEntryId.seq + 1
		}
		return StreamEntryId{ms: newMs, seq: newSeq}, true
	} else {
		if timestamp < lastEntryId.ms {
			return StreamEntryId{}, false
		}

		if timestamp == lastEntryId.ms {
			return StreamEntryId{ ms: timestamp, seq: lastEntryId.seq + 1}, true
		}

		return StreamEntryId{ms: timestamp, seq: 0}, true
	}
}

// this currently does not handle 
// existing key of diffrent type
func Xadd(key string, opt XaddOptions) (string, error) {
	i := getShardLock(key)
	defer releaseShardLock(i)

	vm, ok := loadValue(key)

	if !ok {
		vm = valueWithMeta{
			ttype: TypeStream,
			value: make([]StreamEntry, 0),
		}
	}

	entries := vm.value.([]StreamEntry)

	var id StreamEntryId

	idLesserError :=  errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")

	if (opt.entryId == "*") {
		id, _ = generateEntryId(entries, -1)
	} else {
		idParts := strings.Split(opt.entryId, "-")
		invalidIdError := errors.New("ERR Invalid stream ID specified as stream command argument")
		if len(idParts) != 2 {
			return "", invalidIdError
		}
		idMs, err := strconv.ParseInt(idParts[0], 10, 64)
		if err != nil {
			return "", invalidIdError
		}

		if idParts[1] == "*" {
			id, ok = generateEntryId(entries, idMs)
			if !ok {
				return "", idLesserError
			}
		} else {
			idSeq, err := strconv.ParseInt(idParts[1], 10, 64)
			if err != nil {
				return "", invalidIdError
			}
			id = StreamEntryId{
				ms: idMs, seq: idSeq,
			}
		}
	}


	if (id.Compare(StreamEntryId{seq: 0, ms: 0 }) == 0) {
		return "", errors.New("ERR The ID specified in XADD must be greater than 0-0")
	}

	if (len(entries) == 0) {
		entries = append(entries, StreamEntry{
			id: id,
			kv: opt.kv,
		})

	} else {
		lastId := entries[len(entries) - 1].id
		fmt.Println("last id " + lastId.String() + " " + id.String())
		fmt.Printf("%d", lastId.Compare(id) )
		if !(lastId.Compare(id) < 0) {
			return "", idLesserError
		}

		entries = append(entries, StreamEntry{
			id: id,
			kv: opt.kv,
		})

	}

	vm.value = entries
	data.Store(key, vm)

	dirtyWatchers(key)
	return id.String(), nil
}
