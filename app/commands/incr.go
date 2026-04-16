package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func IncrHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	commandKey := string(input[1].Value.([]byte))

	v, err := Incr(commandKey)

	if err == nil {
		return resp.RespValue{Ttype: resp.RespInt, Value: int64(v)}
	}

	return resp.RespValue{Ttype: resp.RespSimpleError, Value: "ERR value is not an integer or out of range"}
}
