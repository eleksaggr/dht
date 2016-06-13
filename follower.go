package dht

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/golang/protobuf/proto"
)

// Follower is an implementation dht.Role that is dependent upon a Leader.
type Follower struct {
	*Node
}

// NewFollower initializes a new Follower.
func NewFollower(node *Node) (follower *Follower, err error) {
	if node == nil {
		return nil, errors.New("Node may not be nil.")
	}

	return &Follower{
		Node: node,
	}, nil
}

// Register registers a Follower with a Leader under the address leaderHost.
func (follower *Follower) Register(leaderHost string) (err error) {
	fmt.Printf("Connecting...\n")
	conn, err := net.Dial("tcp", leaderHost)
	if err != nil {
		return err
	}

	message := &Message{
		Action: Message_REGISTER.Enum(),

		Id:   follower.id[:],
		Host: proto.String(follower.host),
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	fmt.Printf("Sending registration...\n")
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

// Handle handles all message types a Follower must respond to.
func (follower *Follower) Handle(m *Message, w io.Writer) (err error) {
	fmt.Printf("Follower is handling message...\n")
	switch m.GetAction() {
	case Message_GET:
		err = follower.OnGet(m, w)
	case Message_SET:
		err = follower.OnSet(m, w)
	case Message_DELETE:
		err = follower.OnDelete(m, w)
	case Message_NOOP:
		err = follower.OnNoop(m, w)
	default:
		err = errors.New("Unrecognized action in message.")
	}
	return err
}

// OnGet is called when a message with type Message_GET is received.
func (follower *Follower) OnGet(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	value, ok := follower.table[m.GetKey()]
	follower.mutex.Unlock()

	message := Message{
		Action:  Message_GET.Enum(),
		Value:   &value,
		Success: &ok,
	}

	data, err := proto.Marshal(&message)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

// OnSet is called when a message with type Message_SET is received.
func (follower *Follower) OnSet(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	follower.table[m.GetKey()] = m.GetValue()
	follower.mutex.Unlock()

	return nil
}

// OnDelete is called when a message with type Message_DELETE is received.
func (follower *Follower) OnDelete(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	delete(follower.table, m.GetKey())
	follower.mutex.Unlock()

	return nil
}

// OnNoop is called when a message with type Message_NOOP is received.
func (follower *Follower) OnNoop(m *Message, w io.Writer) error {
	return nil
}
