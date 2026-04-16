package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func MultiHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	conn.mode = MultiMode
	return resp.SerializeSimpleString("OK")
}
