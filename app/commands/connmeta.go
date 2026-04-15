package commands

import "net"

type ConnMeta struct {
	net.Conn
	subscribedChannels map[string]bool
}
