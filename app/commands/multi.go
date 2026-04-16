package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func MultiHandler(input []resp.RespValue, conn *ConnMeta) []byte {
	return resp.SerializeSimpleString("OK")
}
