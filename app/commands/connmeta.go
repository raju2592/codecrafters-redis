package commands

import (
	"net"
	"sync"
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
}
