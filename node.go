package dht

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

type Type uint

const (
	LEADER   Type = iota
	FOLLOWER Type = iota
)

// Node represents a machine in the cluster.
type Node struct {
	id uuid.UUID

	// Network
	host string
	net.Listener

	// Role
	role Role

	// Table
	table map[string]string
	mutex *sync.Mutex

	stop chan bool
}

// NewNode creates a new Node. The node listens for incoming connections on the address host.
func NewNode(host string, roleType Type) (node *Node, err error) {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	node = &Node{
		id: *id,

		host:     host,
		Listener: listener,

		table: make(map[string]string),
		mutex: &sync.Mutex{},

		stop: make(chan bool, 1),
	}

	var role Role
	switch roleType {
	case LEADER:
		role, err = NewLeader(node)
		if err != nil {
			return nil, err
		}
	case FOLLOWER:
		role, err = NewFollower(node)
		if err != nil {
			return nil, err
		}
	}
	node.role = role

	return node, nil
}

func (node *Node) Assign(role Role) {
	node.role = role
}

func (node *Node) Register(leaderHost string) (err error) {
	return node.role.Register(leaderHost)
}

func (node *Node) Run() {
loop:
	for {
		select {
		case <-node.stop:
			break loop
		default:
			fmt.Printf("Waiting for connection...\n")
			conn, err := node.Accept()
			if err != nil {
				log.Printf("Accept: %v\n", err)
				return
			}
			fmt.Printf("Accepted connection.\n")
			// Wrap this in a closure, so defer will close after whichever Handle function is called.

			go func(conn net.Conn) {
				defer conn.Close()

				// Read message from the connection.
				fmt.Printf("Reading from connection.\n")
				buffer := make([]byte, 4096)
				n, err := conn.Read(buffer)
				if err != nil {

					log.Printf("Read: %v\n", err)
					return
				}
				// Remove padding.
				data := make([]byte, n)
				copy(data, buffer)

				// Unmarshal protobuf message.
				fmt.Printf("Unmarshaling message...\n")
				var message Message
				if err = proto.Unmarshal(data, &message); err != nil {
					log.Printf("Data: %v\n%v\n", buffer, err)
					return
				}

				// Pass to appropiate handler.
				fmt.Printf("Passing message...\n")
				if err = node.role.Handle(&message, conn); err != nil {
					log.Printf("Pass: %v\n", err)
					return
				}
			}(conn)
		}
	}
	close(node.stop)
	node.Close()
}

// Role returns the role of the Node.
func (node *Node) Role() Role {
	return node.role
}
