package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PublishHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	channel := resp.GetStringValue(input[1])
	message := resp.GetStringValue(input[2])

	subs, ok := subscribers.Get(channel)

	data := resp.SerializeRespValue(resp.RespValue{Ttype: resp.RespArray, Value: []resp.RespValue{
		{ Ttype: resp.RespBulkString, Value: []byte("message") },
		{ Ttype: resp.RespBulkString, Value: []byte(channel) },
		{ Ttype: resp.RespBulkString, Value: []byte(message) },
	}})

	for _, sub := range subs {
		sub.mu.Lock()
		_, err := sub.Write(data)
		sub.mu.Unlock()
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
		}
	}

	var sc int64 = 0;
	if ok {
		sc = int64(len(subs))
	}

	return resp.RespValue{Ttype: resp.RespInt, Value: sc}
}
