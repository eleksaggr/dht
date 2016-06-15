package dht

import (
	"errors"
	"io"
	"log"
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
	log.Printf("Registring with Leader...\n")
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

	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (follower *Follower) Unregister() (err error) {
	message := Message{
		Action: Message_UNREGISTER.Enum(),
		Id:     follower.id[:],
	}

	data, err := proto.Marshal(&message)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", follower.leaderHost)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Write(data)

	return nil
}

// Handle handles all message types a Follower must respond to.
func (follower *Follower) Handle(m *Message, w io.Writer) (err error) {
	select {
	case <-follower.signal:
		follower.Unregister()
	default:
	}

	switch m.GetAction() {
	case Message_GET:
		log.Printf("[EVENT]Follower handling GET.\n")
		err = follower.OnGet(m, w)
	case Message_SET:
		log.Printf("[EVENT]Follower handling SET.\n")
		err = follower.OnSet(m, w)
	case Message_DELETE:
		log.Printf("[EVENT]Follower handling DELETE.\n")
		err = follower.OnDelete(m, w)
	case Message_GET_ALL:
		log.Printf("[EVENT]Follower handling GET_ALL.\n")
		err = follower.OnGetAll(m, w)
	case Message_NOOP:
		log.Printf("[EVENT]Follower handling NOOP.\n")
		err = follower.OnNoop(m, w)
	case Message_EXIT:
		log.Printf("[EVENT]Follower handling EXIT.\n")
		err = follower.OnExit(m, w)
	default:
		log.Printf("[EVENT]Follower handling fell through.\n")
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

	log.Printf("Table after Set: %v\n", follower.table)
	return nil
}

// OnDelete is called when a message with type Message_DELETE is received.
func (follower *Follower) OnDelete(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	delete(follower.table, m.GetKey())
	follower.mutex.Unlock()

	log.Printf("Table after Delete: %v\n", follower.table)
	return nil
}

func (follower *Follower) OnGetAll(m *Message, w io.Writer) (err error) {
	follower.mutex.Lock()
	keys := make([]string, len(follower.table))
	values := make([]string, len(follower.table))
	i := 0
	for k, v := range follower.table {
		keys[i] = k
		values[i] = v
		i++
	}
	follower.mutex.Unlock()

	message := Message{
		Action: Message_GET_ALL.Enum(),
		Keys:   keys,
		Values: values,
	}

	data, err := proto.Marshal(&message)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

func (follower *Follower) OnExit(m *Message, w io.Writer) (err error) {
	follower.Stop()
	return nil
}

// OnNoop is called when a message with type Message_NOOP is received.
func (follower *Follower) OnNoop(m *Message, w io.Writer) error {
	return nil
}
