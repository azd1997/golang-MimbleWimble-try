package blockchain2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/azd1997/golang-MimbleWimble-try/secp256k1zkp"
	"github.com/sirupsen/logrus"
	"github.com/yoss22/bulletproofs"
	"golang.org/x/crypto/blake2b"
	"io"
)

//每笔交易的证明信息，包括：
//1.佩德森承诺
//2.交易的签名


// SwitchCommitHash the switch commitment hash
type SwitchCommitHash []byte // size = const SwitchCommitHashSize

// A proof that a transaction sums to zero. Includes both the transaction's
// Pedersen commitment and the signature, that guarantees that the commitments
// amount to zero.
// The signature signs the Fee and the LockHeight, which are retained for
// signature validation.
type TxKernel struct {
	// Options for a kernel's structure or use
	Features KernelFeatures
	// Fee originally included in the transaction this proof is for.
	Fee uint64
	// This kernel is not valid earlier than lockHeight blocks
	// The max lockHeight of all *inputs* to this transaction
	LockHeight uint64
	// Remainder of the sum of all transaction commitments. If the transaction
	// is well formed, amounts components should sum to zero and the excess
	// is hence a valid public key.
	Excess bulletproofs.Point
	// The signature proving the excess is a valid public key, which signs
	// the transaction fee.
	ExcessSig [64]byte
}

// Hash returns a hash of the serialised kernel.
func (k *TxKernel) Hash() []byte {
	hashed := blake2b.Sum256(k.Bytes())
	return hashed[:]
}

// Read implements p2p Message interface
func (k *TxKernel) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write features, fee & lock
	if err := binary.Write(buff, binary.BigEndian, uint8(k.Features)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, k.Fee); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, k.LockHeight); err != nil {
		logrus.Fatal(err)
	}

	// Write Excess
	if _, err := buff.Write(k.Excess.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	// Write ExcessSig
	if _, err := buff.Write(k.ExcessSig[:]); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Read implements p2p Message interface
func (k *TxKernel) Read(r io.Reader) error {
	// Read features, fee & lock
	if err := binary.Read(r, binary.BigEndian, (*uint8)(&k.Features)); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &k.Fee); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &k.LockHeight); err != nil {
		return err
	}

	// Read Excess
	if err := k.Excess.Read(r); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, k.ExcessSig[:]); err != nil {
		return err
	}

	return nil
}

var ErrInvalidSignature = errors.New("signature isn't valid")

// Validate returns nil if kernel successfully passed consensus rules.
func (k *TxKernel) Validate() error {
	// The spender signs the fee and lock height using the private key for P. If
	// the signature verifies then we know that there is no residue on G (i.e.
	// that no value is created) and that the spender is in possession of the
	// inputs.
	msg := secp256k1zkp.ComputeMessage(k.Fee, k.LockHeight)
	signature := secp256k1zkp.DecodeSignature(k.ExcessSig)

	// Excess is a Pedersen commitment to the value zero: P = γ*H + 0*G
	P := k.Excess

	if !secp256k1zkp.VerifySignature(P, msg, signature) {
		return ErrInvalidSignature
	}

	return nil
}

// String implements String() interface
func (k TxKernel) String() string {
	return fmt.Sprintf("%#v", k)
}
