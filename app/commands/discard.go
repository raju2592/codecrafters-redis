package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func DiscardHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	if conn.mode != MultiMode {
		return resp.RespValue{Ttype: resp.RespSimpleError, Value: "ERR DISCARD without MULTI"}
	}

	conn.commandQueue = nil
	conn.mode = NormalMode
	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
