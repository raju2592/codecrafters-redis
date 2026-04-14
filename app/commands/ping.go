package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PingHandler(input []resp.RespValue, conn net.Conn) []byte {
	return []byte("+PONG\r\n")
}
