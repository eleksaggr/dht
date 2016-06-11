package dht

import "net"

type Leader struct {
	Follower
}

func (leader *Leader) Handle(conn net.Conn) (err error) {
	err = leader.Follower.Handle(conn)
	if err != nil {
		return err
	}

	return nil
}
