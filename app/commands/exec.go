package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func ExecHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	if conn.mode != MultiMode {
		return resp.RespValue{Ttype: resp.RespSimpleError, Value: "ERR EXEC without MULTI"}
	}


	for _, key := range conn.watchedKeys {
		// Check if watched keys expired
		DeleteIfExpired(key)
	}

	res := resp.RespValue{
		Ttype: resp.RespNullArray,
	}

	unlockKeys := LockKeys(conn.watchedKeys)

	if !conn.dirty.Load() {
		results := make([]resp.RespValue, len(conn.commandQueue))

		for i, command := range conn.commandQueue {
			commandName := strings.ToUpper(resp.GetStringValue(command[0]))
			handler := handlers[commandName]
			results[i] = handler(command, conn)
		}

		res = resp.RespValue{Ttype: resp.RespArray, Value: results}
	}


	unlockKeys();

	conn.watchedKeys = nil
	conn.dirty.Store(false)

	conn.commandQueue = nil
	conn.mode = NormalMode
	return res
}
