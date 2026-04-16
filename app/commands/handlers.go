package commands

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type CommandHandler func(input []resp.RespValue, conn *ConnMeta) resp.RespValue

var handlers map[string]CommandHandler

func init() {
	handlers = map[string]CommandHandler{
		"PING":        PingHandler,
		"ECHO":        EchoHandler,
		"SET":         SetHandler,
		"GET":         GetHandler,
		"SUBSCRIBE":   SubscribeHandler,
		"PUBLISH":     PublishHandler,
		"UNSUBSCRIBE": UnsubscribeHandler,
		"INCR":        IncrHandler,
		"MULTI":       MultiHandler,
		"EXEC":        ExecHandler,
		"DISCARD": DiscardHandler,
	}
}
