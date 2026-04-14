package resp

func GetStringValue(v RespValue) string {
	if v.Ttype == RespSimpleString {
		return v.Value.(string)
	} else if v.Ttype == RespBulkString {
		return string(v.Value.([]byte))
	}
	return ""
}
