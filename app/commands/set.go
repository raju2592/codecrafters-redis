package commands

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func SetHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	commandKey := string(input[1].Value.([]byte))
	commandValue := input[2].Value.([]byte)

	var ttl int

	for i := 3; i < len(input); i += 2 {
		optionKey := strings.ToUpper(resp.GetStringValue(input[i]))
		switch optionKey {
		case "EX":
			ttl, _ = strconv.Atoi(resp.GetStringValue(input[i + 1]))
			ttl *= 1000
		case "PX":
			ttl, _ = strconv.Atoi(resp.GetStringValue(input[i + 1]))
		}
	}

	Set(commandKey, commandValue, SetOptions{
		Ttl: ttl,
	})
	return okResponse
}
