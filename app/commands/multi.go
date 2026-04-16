package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func MultiHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	conn.mode = MultiMode
	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
