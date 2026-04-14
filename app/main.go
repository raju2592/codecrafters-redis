package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var okResponse []byte = []byte("+OK\r\n")
var nullBulkString []byte = []byte("$-1\r\n")

func getStringValue(v RespValue) string {
	if v.ttype == RespSimpleString {
		return v.value.(string)
	} else if v.ttype == RespBulkString {
		return string(v.value.([]byte))
	}
	return ""
}

func handleConn(conn net.Conn) {
	cr := NewConnectionReader(conn, 1024)

	for {
		// buf := make([]byte, 1024)
		// _, err := conn.Read(buf)



		// fmt.Println("PING recvd")

		v, err := ParseValue(cr.ReadByte)
		if err != nil {
			fmt.Println("Error parsing command", err.Error())
			return
		}

		input := v.value.([]RespValue)

		command := input[0]

		commandName := getStringValue(command)

		var reply []byte

		// TODO: this does not handle any invalid commands
		switch strings.ToUpper(commandName) {
		case "PING":
			reply = []byte("+PONG\r\n")
		case "ECHO":
			commandValue := input[1].value.([]byte)
			reply = SerializeBulkString(commandValue)
		case "SET":
			commandKey := string(input[1].value.([]byte))
			commandValue := input[2].value.([]byte)

			var ttl int

			for i := 3; i < len(input); i += 2 {
				optionKey := strings.ToUpper(getStringValue(input[i]))
				switch optionKey {
				case "EX":
					ttl, _ = strconv.Atoi(getStringValue(input[i + 1]))
					ttl *= 1000
				case "PX":
					ttl, _ = strconv.Atoi(getStringValue(input[i + 1]))
				}
			}

			store.Set(commandKey, commandValue, store.SetOptions{
				Ttl: ttl,
			})
			reply = okResponse
		case "GET":
			commandKey := string(input[1].value.([]byte))
			value, ok := store.Get(commandKey)
			if !ok {
				reply = nullBulkString
			} else {
				reply = SerializeBulkString(value)
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for true {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}
