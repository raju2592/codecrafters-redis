package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PingHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	if (conn.mode == SubscribedMode) {
		return resp.RespValue{Ttype: resp.RespArray, Value: []resp.RespValue{
			{ Ttype: resp.RespBulkString, Value: []byte("pong") },
			{ Ttype: resp.RespBulkString, Value: []byte("") },
		}}
	}
	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "PONG"}
}
