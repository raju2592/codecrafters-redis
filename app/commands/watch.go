package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func WatchHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	key := string(input[1].Value.([]byte))

	if conn.mode == MultiMode {
		return resp.RespValue{
			Ttype: resp.RespSimpleError,
			Value: "ERR WATCH inside MULTI is not allowed",
		}
	}

	Watch(key, conn)
	conn.watchedKeys[key] = true
	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
