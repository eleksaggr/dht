package dht

import "net"

type Role interface {
	Handle(conn net.Conn) error
}
