package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PingHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	if (conn.mode == SubscribedMode) {
		return resp.SerializeArray([]resp.RespValue{
			{ Ttype: resp.RespBulkString, Value: []byte("pong") },
			{ Ttype: resp.RespBulkString, Value: []byte("") },
		})
	}
	return []byte("+PONG\r\n")
}
