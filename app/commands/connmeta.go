package commands

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type ConnMode int

const (
	NormalMode = iota
	SubscribedMode
	MultiMode
)

type ConnMeta struct {
	mu sync.Mutex
	net.Conn
	subscribedChannels map[string]bool
	mode ConnMode
	commandQueue [][]resp.RespValue
	watchedKeys map[string]bool
	dirty atomic.Bool
}
