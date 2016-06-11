package dht

import (
	"errors"
	"net"
)

type Follower struct {
}

func (f *Follower) Handle(conn net.Conn) (err error) {
	if conn == nil {
		return errors.New("Connection may not be nil.")
	}

	// Handle here.

	return nil
}
