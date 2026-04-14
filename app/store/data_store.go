package store

import "sync"

var data sync.Map

func Set(key string, value []byte) {
	data.Store(key, value)
}

func Get(key string) ([]byte, bool) {
	v, ok := data.Load(key)
	if ok {
		return v.([]byte), ok
	} else {
		return nil, ok
	}
}
