package dht

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

// Type is the type of a node.
type Type uint

const (
	// LEADER is the type for the role dht.Leader
	LEADER Type = iota
	// FOLLOWER is the type for the role dht.Follower
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

// Register registers the node with a Leader under the address leaderHost.
func (node *Node) Register(leaderHost string) (err error) {
	return node.role.Register(leaderHost)
}

// Run starts a loop, in which messages are handled until Stop is called.
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

// Stop stops the Run-loop of this node.
func (node *Node) Stop() {
	node.stop <- true
}

// Role returns the role of the Node.
func (node *Node) Role() Role {
	return node.role
}
