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

func EncodeInt[T integer](i T) []byte {
	return append([]byte(strconv.FormatInt(int64(i), 10)), end...)
}

func SerializeInteger(i int64) []byte {
	return append([]byte(":"), EncodeInt(i)...)
}

func SerializeArray(arr []RespValue) []byte {
	arrLen := len(arr)
	data := append([]byte("*"), EncodeInt(arrLen)...)
	for _, e := range arr {
		data = append(data, SerializeRespValue(e)...)
	}
	return data
}

func SerializeSimpleError(msg string) []byte {
	return []byte("-" + msg + "\r\n")
}

func SerializeSimpleString(msg string) []byte {
	return []byte("+" + msg + "\r\n")
}


func SerializeRespValue(v RespValue) []byte {
	switch v.Ttype {
	case RespBulkString:
		return SerializeBulkString(v.Value.([]byte))
	case RespInt:
		return SerializeInteger(v.Value.(int64))
	case RespSimpleString:
		return SerializeSimpleString(v.Value.(string))
	case RespArray:
		return SerializeArray(v.Value.([]RespValue))
	case RespSimpleError:
		return SerializeSimpleError(v.Value.(string))
	case RespNull:
		return []byte("$-1\r\n")
	}
	return nil
}
