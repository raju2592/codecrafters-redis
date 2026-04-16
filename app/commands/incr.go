package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func IncrHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	commandKey := string(input[1].Value.([]byte))

	v, err := Incr(commandKey)

	if err != nil {
		return resp.SerializeInteger(int64(v))
	}

	return resp.SerializeError("ERR value is not an integer or out of range")
}
