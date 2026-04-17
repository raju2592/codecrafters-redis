package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func UnwatchHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	for _, key := range conn.watchedKeys {
		Unwatch(key, conn)
	}

	conn.watchedKeys = nil
	conn.dirty.Store(false)

	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
