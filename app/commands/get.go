package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func GetHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	commandKey := string(input[1].Value.([]byte))
	value, ok := Get(commandKey)
	if !ok {
		return nullBulkString
	}
	return resp.SerializeBulkString(value)
}
