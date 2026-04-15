package commands

import "net"

type ConnMode int

const (
	NormalMode = iota
	SubscribedMode
)

type ConnMeta struct {
	net.Conn
	subscribedChannels map[string]bool
	mode ConnMode
}
