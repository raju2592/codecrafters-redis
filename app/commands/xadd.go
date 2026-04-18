package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func XaddHandler(input []resp.RespValue, conn *ConnMeta) resp.RespValue {
	key := string(input[1].Value.([]byte))
	entryId := string(input[2].Value.([]byte))

	kv := make(map[string][]byte)

	n := len(input)

	for i := 3; i < n; i += 2 {
		k := string(input[i].Value.([]byte))
		v := input[i + 1].Value.([]byte)
		kv[k] = v
	} 

	err := Xadd(key, XaddOptions{
		entryId: entryId,
		kv: kv,
	})
	
	if err != nil {
		return resp.RespValue{
			Ttype: resp.RespSimpleError,
			Value: err.Error(),
		}
	}
	return resp.RespValue{Ttype: resp.RespBulkString, Value: []byte(entryId)}
}
