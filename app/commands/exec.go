package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func ExecHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	if conn.mode != MultiMode {
		return resp.RespValue{Ttype: resp.RespSimpleError, Value: "ERR EXEC without MULTI"}
	}


	for key, _ := range conn.watchedKeys {
		// Check if watched keys expired
		DeleteIfExpired(key)

		// Unwatch key
		Unwatch(key, conn)
	}

	res := resp.RespValue{
		Ttype: resp.RespNull,
	}


	if !conn.dirty.Load() {
		results := make([]resp.RespValue, len(conn.commandQueue))

		for i, command := range conn.commandQueue {
			commandName := strings.ToUpper(resp.GetStringValue(command[0]))
			handler := handlers[commandName]
			results[i] = handler(command, conn)
		}

		res = resp.RespValue{Ttype: resp.RespArray, Value: results}
	}

	conn.commandQueue = nil
	conn.mode = NormalMode
	return res
}
