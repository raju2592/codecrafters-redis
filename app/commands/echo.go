package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func EchoHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	commandValue := input[1].Value.([]byte)
	return resp.RespValue{Ttype: resp.RespBulkString, Value: commandValue}
}
