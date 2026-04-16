package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func ExecHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	if conn.mode != MultiMode {
		return resp.SerializeError("ERR EXEC without MULTI")
	}
	conn.mode = NormalMode
	return resp.SerializeArray([]resp.RespValue{})
}
