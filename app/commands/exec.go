package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func ExecHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	if conn.mode != MultiMode {
		return resp.SerializeError("ERR EXEC without MULTI")
	}


	n := len(conn.commandQueue)
	res := append([]byte("*"), resp.EncodeInt(n)...)

	for _, command := range conn.commandQueue {
		commandName := strings.ToUpper(resp.GetStringValue(command[0]))
		handler := handlers[commandName]

		output := handler(command, conn)
		res = append(res, output...)
	}

	conn.mode = NormalMode
	return res
}
