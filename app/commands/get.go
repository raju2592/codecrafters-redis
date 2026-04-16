package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func GetHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	commandKey := string(input[1].Value.([]byte))
	value, ok := Get(commandKey)
	if !ok {
		return resp.RespValue{Ttype: resp.RespNull}
	}
	return resp.RespValue{Ttype: resp.RespBulkString, Value: value}
}
