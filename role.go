package dht

import "io"

// Role defines basic operations a role must have to be considered applicable to a node.
type Role interface {
	// Register registers with a Leader under the address leaderHost.
	Register(leaderHost string) error

	Unregister() error

	// Handle handles all incoming messages.
	Handle(m *Message, w io.Writer) error

	// OnGet is called when an incoming message is of type Message_GET.
	OnGet(m *Message, w io.Writer) error
	// OnSet is called when an incoming message is of type Message_SET.
	OnSet(m *Message, w io.Writer) error
	// OnDelete is called when an incoming message is of type Message_DELETE.
	OnDelete(m *Message, w io.Writer) error
	// OnNoop is called when an incoming message is of type Message_NOOP.
	OnNoop(m *Message, w io.Writer) error
	// OnGetAll is called when an incoming message is of type Message_GET_ALL.
	OnGetAll(m *Message, w io.Writer) error
	// OnExit is called when an incoming message is of type Message_EXIT.
	OnExit(m *Message, w io.Writer) error
}
