package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PingHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	return []byte("+PONG\r\n")
}
