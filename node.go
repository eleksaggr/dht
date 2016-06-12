package dht

import (
	"log"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

// Node represents a machine in the cluster.
type Node struct {
	id *uuid.UUID

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
func NewNode(host string) (node *Node, err error) {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	node = &Node{
		id: id,

		host:     host,
		Listener: listener,

		table: make(map[string]string),
		mutex: &sync.Mutex{},

		stop: make(chan bool, 1),
	}

	return node, nil
}

func (node *Node) Run() {
	for {
		select {
		case <-node.stop:
			break
		default:
			conn, err := node.Accept()
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			// Wrap this in a closure, so defer will close after whichever Handle function is called.
			go func(conn net.Conn) {
				defer node.Close()

				// Read message from the connection.
				buffer := make([]byte, 4096)
				_, err := conn.Read(buffer)
				if err != nil {
					log.Printf("%v\n", err)
					return
				}

				// Unmarshal protobuf message.
				var message Message
				if err := proto.Unmarshal(buffer, &message); err != nil {
					log.Printf("%v\n", err)
					return
				}

				// Pass to appropiate handler.
				node.role.Handle(message)
			}(conn)
		}
	}
}

// Role returns the role of the Node.
func (node *Node) Role() Role {
	return node.role
}
