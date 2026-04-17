package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func TypeHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	commandKey := string(input[1].Value.([]byte))
	value := GetType(commandKey)

	return resp.RespValue{Ttype: resp.RespSimpleString, Value: value}
}
