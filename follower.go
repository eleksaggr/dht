package dht

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/golang/protobuf/proto"
)

type Follower struct {
	*Node
}

func NewFollower(node *Node) (follower *Follower, err error) {
	if node == nil {
		return nil, errors.New("Node may not be nil.")
	}

	return &Follower{
		Node: node,
	}, nil
}

func (follower *Follower) Assign(role Role) (err error) {
	follower.Node.Assign(role)
	return nil
}

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
	fmt.Printf("%v\n", data)
	if err != nil {
		return err
	}

	fmt.Printf("Sending registration...\n")
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (follower *Follower) Handle(m *Message, w io.Writer) (err error) {
	fmt.Printf("Follower is handling message...\n")
	switch m.GetAction() {
	case Message_GET:
		err = follower.OnGet(m, w)
	case Message_SET:
		err = follower.OnSet(m, w)
	case Message_DELETE:
		err = follower.OnDelete(m, w)
	default:
		err = errors.New("Unrecognized action in message.")
	}
	return err
}

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

func (follower *Follower) OnSet(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	follower.table[m.GetKey()] = m.GetValue()
	follower.mutex.Unlock()

	return nil
}

func (follower *Follower) OnDelete(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	delete(follower.table, m.GetKey())
	follower.mutex.Unlock()

	return nil
}
