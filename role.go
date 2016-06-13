package dht

import "io"

type Role interface {
	Register(leaderHost string) error

	Handle(m *Message, w io.Writer) error

	OnGet(m *Message, w io.Writer) error
	OnSet(m *Message, w io.Writer) error
	OnDelete(m *Message, w io.Writer) error
	OnNoop(m *Message, w io.Writer) error
}
