package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var okResponse []byte = []byte("+OK\r\n")
var nullBulkString []byte = []byte("$-1\r\n")

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

		var commandName string
		if command.ttype == RespSimpleString {
			commandName = command.value.(string)
		} else {
			commandName = string(command.value.([]byte))
		}

		var reply []byte

		switch strings.ToUpper(commandName) {
		case "PING":
			reply = []byte("+PONG\r\n")
		case "ECHO":
			commandValue := input[1].value.([]byte)
			reply = SerializeBulkString(commandValue)
		case "SET":
			commandKey := string(input[1].value.([]byte))
			commandValue := input[2].value.([]byte)
			store.Set(commandKey, commandValue)
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

		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
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
