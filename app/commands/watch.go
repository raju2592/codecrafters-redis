package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func WatchHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	n := len(input)
	keys := make([]string, n - 1)

	for i, v := range input[:n - 1] {
		keys[i] = string(v.Value.([]byte))
	}

	if conn.mode == MultiMode {
		return resp.RespValue{
			Ttype: resp.RespSimpleError,
			Value: "ERR WATCH inside MULTI is not allowed",
		}
	}

	for _, key := range keys {
		Watch(key, conn)
	}

	conn.watchedKeys = append(conn.watchedKeys, keys...)
	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
