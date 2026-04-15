package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func PublishHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	channel := resp.GetStringValue(input[1])

	subs, ok := subscribers.Get(channel)

	var sc int64 = 0;
	if ok {
		sc = int64(len(subs))
	}

	return resp.SerializeInteger(sc)
}