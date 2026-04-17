package resp

import (
	"errors"
	"fmt"
	"strconv"
)

type RespType int

const (
	RespInt RespType = iota
	RespSimpleString
	RespArray
	RespBulkString
	RespSimpleError
	RespNull
	RespNullArray
)

type RespValue struct {
	Ttype RespType
	Value interface{}
}

func parseUntilLineBreak(nextByte func()(byte, error)) ([]byte, error) {
	ret := []byte{}
	
	for true {
		bt, err := nextByte()

		if err != nil {
			return ret, err
		}

		if (bt == byte('\r')) {
			nxtBt, err := nextByte()

			if err != nil {
				return ret, err
			}

			if (nxtBt != byte('\n')) {
				return ret, errors.New(fmt.Sprintf("Unexpected character: %c", nxtBt))
			}

			return ret, nil
		}

		ret = append(ret, bt)
	}

	return ret, nil
}

func ParseInteger(nextByte func()(byte, error)) (int64, error) {
	data, err := parseUntilLineBreak(nextByte)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(string(data), 10, 64)

	if err != nil {
		return 0, err
	}

	return v, nil
}

func ParseSimpleString(nextByte func()(byte, error)) (string, error) {
	data, err := parseUntilLineBreak(nextByte)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func ParseBulkString(nextByte func()(byte, error)) ([]byte, error) {
	len, err := ParseInteger(nextByte)
	if err != nil {
		return nil, err
	}

	data := []byte{}
	for i := int64(0); i < len; i++ {
		b, err := nextByte()
		if err != nil {
			return nil, err
		}
		data = append(data, b)
	}

	b, err := nextByte()

	if err != nil {
		return nil, err
	}

	if b != byte('\r') {
		return nil, errors.New(fmt.Sprint("Invalid character: %c", b))
	}

	b, err = nextByte()

	if err != nil {
		return nil, err
	}

	if b != byte('\n') {
		return nil, errors.New(fmt.Sprint("Invalid character: %c", b))
	}

	return data, nil
}

func ParseArray(nextByte func()(byte, error)) ([]RespValue, error) {
	arr := []RespValue{}
	len, err := ParseInteger(nextByte)
	if err != nil {
		return nil, err
	}

	for i := int64(0); i < len; i++ {
		val, err := ParseValue(nextByte)
		if err != nil {
			return nil, err
		}

		arr = append(arr, val)
	}

	return arr, nil
}

func ParseValue(nextByte func()(byte, error)) (RespValue, error) {
	b, err := nextByte()
	if err != nil {
		return RespValue{}, err
	}

	switch b {
	case byte(':'):
		v, err := ParseInteger(nextByte)
		if err != nil {
			return RespValue{}, err
		}
		return RespValue{
			Ttype: RespInt,
			Value: v,
		}, nil
	case byte('+'):
		v, err := ParseSimpleString(nextByte)
		if err != nil {
			return RespValue{}, err
		}
		return RespValue{
			Ttype: RespSimpleString,
			Value: v,
		}, nil
	case byte('*'):
		v, err := ParseArray(nextByte)
		if err != nil {
			return RespValue{}, err
		}
		return RespValue{
			Ttype: RespArray,
			Value: v,
		}, nil
	case byte('$'):
		v, err := ParseBulkString(nextByte)
		if err != nil {
			return RespValue{}, err
		}
		return RespValue{
			Ttype: RespBulkString,
			Value: v,
		}, nil
	}

	return RespValue{}, errors.New(fmt.Sprintf("Invalid character: %c", b))
}
