package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func UnwatchHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	for key, _ := range conn.watchedKeys {
		Unwatch(key, conn)
	}

	return resp.RespValue{Ttype: resp.RespSimpleString, Value: "OK"}
}
