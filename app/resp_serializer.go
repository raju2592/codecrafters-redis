package main

import "strconv"

func SerializeBulkString(data []byte) []byte {
	len := len(data)

	end := []byte("\r\n")

	ret := append([]byte("$" + strconv.Itoa(len)), end...)
	ret = append(ret, data...)
	ret = append(ret, end...)

	return ret
}
