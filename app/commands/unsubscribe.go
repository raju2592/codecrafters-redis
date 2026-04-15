package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func UnsubscribeHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	channel := resp.GetStringValue(input[1])

	delete(conn.subscribedChannels, channel)
	if len(conn.subscribedChannels) == 0 {
		conn.mode = NormalMode
	}

	subscribers.Update(channel, func (conns []*ConnMeta, exists bool) ([]*ConnMeta, bool) {
		newConns := make([]*ConnMeta, 0, len(conns))
		for _, c := range conns {
			if c != conn {
				newConns = append(newConns, c)
			}
		}
		return newConns, len(newConns) > 0
	})

	return resp.SerializeArray([]resp.RespValue{
		{ Ttype: resp.RespBulkString, Value: []byte("unsubscribe") },
		{ Ttype: resp.RespBulkString, Value: []byte(channel) },
		{ Ttype: resp.RespInt, Value: int64(len(conn.subscribedChannels)) },
	})
}
