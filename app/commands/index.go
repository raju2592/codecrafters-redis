package commands

import (
	"fmt"
	"net"
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
}

func HandleConn(conn net.Conn) {
	cr := netio.NewConnectionReader(conn, 1024)
	
	connMeta := ConnMeta{
    Conn:               conn,
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

		if handler, ok := handlers[commandName]; ok {
			reply = handler(input, &connMeta)
		}

		_, err = conn.Write(reply)
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			conn.Close()
			return
		}
	}
}
