package blockchain2

import "io"

// Message defines methods for WriteMessage/ReadMessage functions
type Message interface {
	// Read reads from reader and fit self struct
	Read(r io.Reader) error

	// Bytes returns binary data of body message
	Bytes() []byte

	// Type says whats the message type should use in header
	Type() uint8
}
