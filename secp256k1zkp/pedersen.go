package secp256k1zkp

import (
	"encoding/hex"
	"io"
)

type Commitment []byte

// Bytes implements p2p Message interface
func (c *Commitment) Bytes() []byte {
	return *c
}

// Read implements p2p Message interface
func (c *Commitment) Read(r io.Reader) error {
	_, err := io.ReadFull(r, *c)

	return err
}

// String implements String() interface
func (c Commitment) String() string {
	return hex.EncodeToString(c.Bytes())
}
