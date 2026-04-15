package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func SubscribeHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	channel := resp.GetStringValue(input[1])

	conn.subscribedChannels[channel] = true
	conn.mode = SubscribedMode

	subscribers.Update(channel, func (conns []*ConnMeta, exists bool) ([]*ConnMeta, bool) {
		newConns := make([]*ConnMeta, len(conns)+1)
		copy(newConns, conns)
		newConns[len(conns)] = conn
		return newConns, true
	})

	return resp.SerializeArray([]resp.RespValue{
		{ Ttype: resp.RespBulkString, Value: []byte("subscribe") },
		{ Ttype: resp.RespBulkString, Value: []byte(channel) },
		{ Ttype: resp.RespInt, Value: int64(len(conn.subscribedChannels)) },
	})
}
