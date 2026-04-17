package commands

type ValueType int

const (
    TypeString ValueType = iota
    TypeList
    TypeSet
    TypeZSet
    TypeHash
    TypeStream
		TypeNone
)

func (t ValueType) String() string {
    switch t {
    case TypeString: return "string"
    case TypeList:   return "list"
    case TypeSet:    return "set"
    case TypeZSet:   return "zset"
    case TypeHash:   return "hash"
    case TypeStream: return "stream"
    default:         return "none"
    }
}
