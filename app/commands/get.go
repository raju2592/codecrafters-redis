package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func GetHandler(input []resp.RespValue, conn net.Conn) []byte {
	commandKey := string(input[1].Value.([]byte))
	value, ok := store.Get(commandKey)
	if !ok {
		return nullBulkString
	}
	return resp.SerializeBulkString(value)
}
