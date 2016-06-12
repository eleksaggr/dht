package dht

type Role interface {
	Handle(m *Message) error
}
