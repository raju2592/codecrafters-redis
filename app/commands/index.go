package commands

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/netio"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var subscribeModeAllowedCommands = []string{
	"SUBSCRIBE",
	"UNSUBSCRIBE",
	"PSUBSCRIBE",
	"PUNSUBSCRIBE",
	"PING",
	"QUIT",
}

var multiModePassthroughCommands = []string{
	"EXEC",
	"DISCARD",
	"WATCH",
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

		var result resp.RespValue

		if !IsCommandAllowed(connMeta, commandName) {
			result = resp.RespValue{Ttype: resp.RespSimpleError, Value: fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", strings.ToLower(commandName))}
		} else if handler, ok := handlers[commandName]; ok {
			if connMeta.mode == MultiMode && !slices.Contains(multiModePassthroughCommands, commandName) {
				connMeta.commandQueue = append(connMeta.commandQueue, input)
				result = resp.RespValue{Ttype: resp.RespSimpleString, Value: "QUEUED"}
			} else {
				result = handler(input, connMeta)
			}
		}

		_, err = conn.Write(resp.SerializeRespValue(result))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			conn.Close()
			return
		}
	}
}
