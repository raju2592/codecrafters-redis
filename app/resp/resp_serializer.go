package resp

import "strconv"

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

var end = []byte("\r\n")

func SerializeBulkString(data []byte) []byte {
	len := len(data)

	ret := append([]byte("$" + strconv.Itoa(len)), end...)
	ret = append(ret, data...)
	ret = append(ret, end...)

	return ret
}

func encodeInt[T integer](i T) []byte {
	return append([]byte(strconv.FormatInt(int64(i), 10)), end...)
}

func SerializeInteger(i int64) []byte {
	return append([]byte(":"), encodeInt(i)...)
}

func SerializeArray(arr []RespValue) []byte{
	arrLen := len(arr)
	data := append([]byte("*"), encodeInt(arrLen)...)
	for _, e := range arr {
		switch e.Ttype {
		case RespBulkString:
			value := e.Value.([]byte)
			data = append(data, SerializeBulkString(value)...)
		case RespInt:
			value := e.Value.(int64)
			data = append(data, SerializeInteger(value)...)
		}
	}
	return data
}

func SerializeError(msg string) []byte {
	return []byte("-" + msg + "\r\n")
}
