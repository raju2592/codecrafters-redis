package commands

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/netio"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var okResponse []byte = []byte("+OK\r\n")
var nullBulkString []byte = []byte("$-1\r\n")

type CommandHandler func(input []resp.RespValue, conn *ConnMeta) []byte

var handlers = map[string]CommandHandler{
	"PING": PingHandler,
	"ECHO": EchoHandler,
	"SET":  SetHandler,
	"GET":  GetHandler,
	"SUBSCRIBE": SubscribeHandler,
	"PUBLISH": PublishHandler,
	"UNSUBSCRIBE": UnsubscribeHandler,
	"INCR": IncrHandler,
	"MULTI": MultiHandler,
	"EXEC": ExecHandler,
}

var subscribeModeAllowedCommands = []string{
	"SUBSCRIBE",
	"UNSUBSCRIBE",
	"PSUBSCRIBE",
	"PUNSUBSCRIBE",
	"PING",
	"QUIT",
}

func IsCommandAllowed(conn *ConnMeta, commandName string) bool {
	if conn.mode == SubscribedMode {
		return slices.Contains(subscribeModeAllowedCommands, commandName)
	}

	return true
}

// TODO: On connection close, remove the connection from all subscribed channels
// to prevent stale entries in the subscriber lists.
func HandleConn(conn net.Conn) {
	cr := netio.NewConnectionReader(conn, 1024)
	
	connMeta := &ConnMeta{
    Conn:               conn,
		mode: NormalMode,
    subscribedChannels: make(map[string]bool),
	}

	for {
		v, err := resp.ParseValue(cr.ReadByte)
		if err != nil {
			fmt.Println("Error parsing command", err.Error())
			return
		}

		input := v.Value.([]resp.RespValue)

		commandName := strings.ToUpper(resp.GetStringValue(input[0]))

		var reply []byte

		if !IsCommandAllowed(connMeta, commandName) {
			reply = []byte(fmt.Sprintf("-ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context\r\n", strings.ToLower(commandName)))
		} else if handler, ok := handlers[commandName]; ok {
			if connMeta.mode == MultiMode && commandName != "EXEC" {
				connMeta.commandQueue = append(connMeta.commandQueue, input)
				reply = resp.SerializeSimpleString("QUEUED")
			} else {
				reply = handler(input, connMeta)
			}
		}


		_, err = conn.Write(reply)
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			conn.Close()
			return
		}
	}
}
