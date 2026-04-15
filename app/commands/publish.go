package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PublishHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	channel := resp.GetStringValue(input[1])
	message := resp.GetStringValue(input[2])

	subs, ok := subscribers.Get(channel)

	data := resp.SerializeArray([]resp.RespValue{
		{ Ttype: resp.RespBulkString, Value: []byte("message") },
		{ Ttype: resp.RespBulkString, Value: []byte(channel) },
		{ Ttype: resp.RespBulkString, Value: []byte(message) },
	})

	for _, sub := range subs {
		_, err := sub.Write(data)
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
		}
	}

	var sc int64 = 0;
	if ok {
		sc = int64(len(subs))
	}

	return resp.SerializeInteger(sc)
}