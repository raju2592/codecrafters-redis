package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func SubscribeHandler(input []resp.RespValue, conn net.Conn) []byte {
	channel := resp.GetStringValue(input[1])

	return resp.SerializeArray([]resp.RespValue{
		{ Ttype: resp.RespBulkString, Value: []byte("subscribe") },
		{ Ttype: resp.RespBulkString, Value: []byte(channel) },
		{ Ttype: resp.RespInt, Value: int64(1) },
	})
}
