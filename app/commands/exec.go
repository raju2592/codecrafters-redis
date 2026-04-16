package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func ExecHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	if conn.mode != MultiMode {
		return resp.RespValue{Ttype: resp.RespSimpleError, Value: "ERR EXEC without MULTI"}
	}

	results := make([]resp.RespValue, len(conn.commandQueue))

	for i, command := range conn.commandQueue {
		commandName := strings.ToUpper(resp.GetStringValue(command[0]))
		handler := handlers[commandName]
		results[i] = handler(command, conn)
	}

	conn.mode = NormalMode
	return resp.RespValue{Ttype: resp.RespArray, Value: results}
}
