package blockchain2

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/azd1997/golang-MimbleWimble-try/secp256k1zkp"
	"github.com/sirupsen/logrus"
	"github.com/yoss22/bulletproofs"
	"io"
)

//一笔交易的输出部分，定义一笔钱的新归属。
//对于输出而言，其承诺Commit为一盲值r， （佩德森承诺：x = rG+vH）
//RangeProof
type Output struct {
	// 指示该笔输出是coinbase还是普通交易输出
	Features OutputFeatures
	// 同态承诺，以其代表了交易的数值（交易双方知道而验证者无法得知）
	Commit *bulletproofs.Point
	// A proof that the commitment is in the right range
	RangeProof bulletproofs.BulletProof
}

func (o *Output) BytesWithoutProof() []byte {
	buff := new(bytes.Buffer)

	// Write features
	if err := binary.Write(buff, binary.BigEndian, uint8(o.Features)); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(o.Commit.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Bytes implements p2p Message interface
func (o *Output) Bytes() []byte {
	buff := new(bytes.Buffer)

	if _, err := buff.Write(o.BytesWithoutProof()); err != nil {
		logrus.Fatal(err)
	}

	proof := o.RangeProof.Bytes()

	if err := binary.Write(buff, binary.BigEndian, uint64(len(proof))); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(proof); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Read implements p2p Message interface
func (o *Output) Read(r io.Reader) error {
	// Read features
	if err := binary.Read(r, binary.BigEndian, (*uint8)(&o.Features)); err != nil {
		return err
	}

	// Read commitment
	o.Commit = new(bulletproofs.Point)
	if err := o.Commit.Read(r); err != nil {
		return err
	}

	// Read range proof
	var proofLen uint64 // tha max is MaxProofSize (5134), but in message field it is uint64
	if err := binary.Read(r, binary.BigEndian, &proofLen); err != nil {
		return err
	}

	if proofLen > uint64(secp256k1zkp.MaxProofSize) {
		return fmt.Errorf("invalid range proof length: %d", proofLen)
	}

	proof := new(bulletproofs.BulletProof)
	err := proof.Read(io.LimitReader(r, int64(proofLen)))
	if err != nil {
		return errors.New("failed to deserialize range proof")
	}
	o.RangeProof = *proof

	return nil
}

// Validate returns nil if output successfully passed consensus rules
func (o *Output) Validate() error {
	return nil
}

// String implements String() interface
func (o Output) String() string {
	return fmt.Sprintf("%#v", o)
}

// Hash returns a hash of the serialised output.
func (o *Output) Hash() []byte {
	hashed := sha256.Sum256(o.BytesWithoutProof())
	return hashed[:]
}