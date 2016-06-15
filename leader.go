package dht

import (
	"errors"
	"hash/crc32"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nu7hatch/gouuid"
)

const (
	// KeepAliveTimeout is the duration waited until a timeout check is performed on a node.
	KeepAliveTimeout = time.Second * 8
	// LoopTimeout is the duration waited until all nodes are checked for timeout again.
	LoopTimeout = time.Second * 2
)

// Leader is an implementation of dht.Role that manages other nodes.
type Leader struct {
	*Follower

	cluster      Cluster
	clusterMutex *sync.Mutex
}

// NewLeader initializes a new leader. It also starts the checkTimeout method.
func NewLeader(node *Node) (leader *Leader, err error) {
	follower, err := NewFollower(node)
	if err != nil {
		return nil, err
	}

	leader = &Leader{
		Follower:     follower,
		cluster:      make(Cluster, 0),
		clusterMutex: &sync.Mutex{},
	}
	leader.Follower.role = leader

	// Add self to the cluster.
	leader.cluster.Add(leader.id, leader.host)

	// Start timeout check
	go leader.checkTimeout()

	return leader, nil
}

// Register returns an error when called.
func (leader *Leader) Register(leaderHost string) (err error) {
	return errors.New("Cannot register a leader.")
}

func (leader *Leader) Unregister() (err error) {
	return errors.New("Cannot unregister a leader.")
}

// Handle handles the different types of messages a Leader must process.
func (leader *Leader) Handle(m *Message, w io.Writer) (err error) {
	err = leader.Follower.Handle(m, w)
	if err == nil {
		return nil
	}

	switch m.GetAction() {
	case Message_REGISTER:
		log.Printf("[EVENT]Leader handling REGISTER.\n")
		err = leader.OnRegister(m, w)
	case Message_UNREGISTER:
		log.Printf("[EVENT]Leader handling UNREGISTER.\n")
		err = leader.OnUnregister(m, w)
	default:
		log.Printf("[EVENT]Leader handling fell through.\nMessage: %v\n", m)
		err = errors.New("Unrecognized action in message.")
	}
	return err
}

// OnRegister is called when a node tries to register with the Leader.
func (leader *Leader) OnRegister(m *Message, w io.Writer) (err error) {
	var id [16]byte
	copy(id[:], m.GetId())
	leader.clusterMutex.Lock()
	leader.cluster.Add(uuid.UUID(id), m.GetHost())
	leader.clusterMutex.Unlock()

	return nil
}

// OnUnregister is called when a node tries to unregister itself with the Leader.
func (leader *Leader) OnUnregister(m *Message, w io.Writer) (err error) {
	var id [16]byte
	copy(id[:], m.GetId())
	keys, values, err := leader.GetAll(id)
	if err != nil {
		return err
	}

	var node *clusterMember
	for _, member := range leader.cluster {
		if member.ID == id {
			node = member
		}
	}
	leader.clusterMutex.Lock()
	leader.cluster.Remove(node)
	leader.clusterMutex.Unlock()

	for i := 0; i < len(keys); i++ {
		leader.Set(keys[i], values[i])
	}

	message := Message{
		Action: Message_EXIT.Enum(),
	}

	data, err := proto.Marshal(&message)
	if err != nil {
		return err
	}
	conn, err := net.Dial("tcp", node.Host)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Write(data)

	return nil
}

// Set advices the appropiate nodes to set the key-value pair.
func (leader *Leader) Set(key, value string) (err error) {
	rendezvous := NewRendezvous(crc32.NewIEEE())
	for _, member := range leader.cluster {
		rendezvous.Add(member.ID)
	}

	nodeIDs := rendezvous.TopN(key, 2)
	nodes := make([]*clusterMember, len(nodeIDs))
	for i, id := range nodeIDs {
		for _, member := range leader.cluster {
			if member.ID == id {
				nodes[i] = member
			}
		}
	}

	message := Message{
		Action: Message_SET.Enum(),
		Key:    &key,
		Value:  &value,
	}

	for _, node := range nodes {
		conn, err := net.Dial("tcp", node.Host)
		if err != nil {
			return err
		}
		data, err := proto.Marshal(&message)
		if err != nil {
			return err
		}
		conn.Write(data)

		conn.Close()
	}
	return nil
}

// Get contacts the appropiate Node to get the value for the key.
func (leader *Leader) Get(key string) (value string, err error) {
	rendezvous := NewRendezvous(crc32.NewIEEE())
	for _, member := range leader.cluster {
		rendezvous.Add(member.ID)
	}

	nodeID := rendezvous.TopN(key, 1)[0]

	var node *clusterMember
	for _, member := range leader.cluster {
		if member.ID == nodeID {
			node = member
		}
	}

	message := Message{
		Action: Message_GET.Enum(),
		Key:    &key,
	}

	conn, err := net.Dial("tcp", node.Host)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := proto.Marshal(&message)
	if err != nil {
		return "", err
	}
	conn.Write(data)

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}

	if err := proto.Unmarshal(buffer[0:n], &message); err != nil {
		return "", err
	}

	if !message.GetSuccess() {
		return "", errors.New("Key not found.")
	}

	return message.GetValue(), nil
}

func (leader *Leader) GetAll(id uuid.UUID) (keys []string, values []string, err error) {
	var node *clusterMember
	for _, member := range leader.cluster {
		if member.ID == id {
			node = member
			break
		}
	}

	message := Message{
		Action: Message_GET_ALL.Enum(),
	}

	data, err := proto.Marshal(&message)
	if err != nil {
		return nil, nil, err
	}

	conn, err := net.Dial("tcp", node.Host)
	if err != nil {
		return nil, nil, err
	}
	conn.Write(data)

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, nil, err
	}

	if err := proto.Unmarshal(buffer[0:n], &message); err != nil {
		return nil, nil, err
	}

	keys = message.GetKeys()
	values = message.GetValues()
	conn.Close()

	return keys, values, nil
}

// Delete advices the appropiate nodes to delete the key.
func (leader *Leader) Delete(key string) (err error) {
	rendezvous := NewRendezvous(crc32.NewIEEE())
	for _, member := range leader.cluster {
		rendezvous.Add(member.ID)
	}

	nodeIDs := rendezvous.TopN(key, 2)
	nodes := make([]*clusterMember, len(nodeIDs))
	for i, id := range nodeIDs {
		for _, member := range leader.cluster {
			if member.ID == id {
				nodes[i] = member
			}
		}
	}

	message := Message{
		Action: Message_DELETE.Enum(),
		Key:    &key,
	}

	for _, node := range nodes {
		conn, err := net.Dial("tcp", node.Host)
		if err != nil {
			return err
		}
		data, err := proto.Marshal(&message)
		if err != nil {
			return err
		}
		conn.Write(data)
		conn.Close()
	}
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
				if time.Since(member.LastAlive) > KeepAliveTimeout {
					go func(member *clusterMember) {
						// Connect to the member to see if he's still alive.
						conn, err := net.Dial("tcp", member.Host)
						if err != nil {
							// Remove member from cluster, since he's dead.
							leader.clusterMutex.Lock()
							leader.cluster.Remove(member)
							leader.clusterMutex.Unlock()
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
		time.Sleep(LoopTimeout)
	}
}
