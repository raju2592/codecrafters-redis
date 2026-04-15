package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func EchoHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	commandValue := input[1].Value.([]byte)
	return resp.SerializeBulkString(commandValue)
}
