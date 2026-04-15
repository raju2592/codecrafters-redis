package commands

import (
	"hash/fnv"

	"github.com/codecrafters-io/redis-starter-go/app/shardmap"
)

var subscribers = shardmap.New[string, []*ConnMeta](func(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
})
