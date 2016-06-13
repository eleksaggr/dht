package dht

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

const (
	KEEPALIVE_TIMEOUT = time.Second * 8
	LOOP_TIMEOUT      = time.Second * 2
)

type Leader struct {
	*Follower

	cluster Cluster
}

func NewLeader(node *Node) (leader *Leader, err error) {
	follower, err := NewFollower(node)
	if err != nil {
		return nil, err
	}

	leader = &Leader{
		Follower: follower,
		cluster:  make(Cluster, 0),
	}
	leader.Follower.Assign(leader)

	// Add self to the cluster.
	leader.cluster.Add(leader.id, leader.host)
	fmt.Printf("Cluster: %v\n", len(leader.cluster))
	// Start timeout check
	go leader.checkTimeout()

	return leader, nil
}

func (leader *Leader) Register(leaderHost string) (err error) {
	return errors.New("Cannot register a leader.")
}

func (leader *Leader) Handle(m *Message, w io.Writer) (err error) {
	fmt.Printf("Leader is handling message...\n")
	err = leader.Follower.Handle(m, w)

	switch m.GetAction() {
	case Message_REGISTER:
		fmt.Printf("Register event called.\n")
		err = leader.OnRegister(m, w)
	case Message_UNREGISTER:
		err = leader.OnUnregister(m, w)
	default:
		err = errors.New("Unrecognized action in message.")
	}
	return err
}

func (leader *Leader) OnRegister(m *Message, w io.Writer) (err error) {
	fmt.Printf("Adding node to cluster...\n")
	var id [16]byte
	copy(id[:], m.GetId())
	leader.cluster.Add(uuid.UUID(id), m.GetHost())

	fmt.Printf("Cluster has %v elements.\n", len(leader.cluster))
	return nil
}

func (leader *Leader) OnUnregister(m *Message, w io.Writer) (err error) {
	return nil
}

func (leader *Leader) checkTimeout() {
	for {
		select {
		case <-leader.stop:
			break
		default:
			for _, member := range leader.cluster {
				if member.ID == leader.id {
					continue
				}
				if time.Since(member.LastAlive) > KEEPALIVE_TIMEOUT {
					go func(member *clusterMember) {
						// Connect to the member to see if he's still alive.
						conn, err := net.Dial("tcp", member.Host)
						if err != nil {
							// Remove member from cluster, since he's dead.
							leader.cluster.Remove(member)
						} else {
							message := Message{
								Action: Message_NOOP.Enum(),
							}
							data, err := proto.Marshal(&message)
							if err != nil {
								log.Printf("%v\n", err)
								return
							}
							if _, err := conn.Write(data); err != nil {
								log.Printf("%v\n", err)
								return
							}

							member.LastAlive = time.Now()
						}
						defer conn.Close()
					}(member)
				}
			}
		}
		time.Sleep(LOOP_TIMEOUT)
	}
}
