package blockchain2

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"github.com/azd1997/golang-MimbleWimble-try/secp256k1zkp"
	"github.com/azd1997/golang-MimbleWimble-try/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"io"
)

type Input struct {
	Features OutputFeatures
	Commit   secp256k1zkp.Commitment
}

//实现Message接口的Bytes方法，对input序列化成字节数组
func (input *Input) Bytes() []byte {
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint8(input.Features)); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(input.Commit); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

func (input *Input) Read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &input.Features); err != nil {
		return err
	}

	commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
	if _, err := io.ReadFull(r, commitment); err != nil {
		return err
	}

	input.Commit = commitment

	return nil
}

// Hash returns a hash of the serialised input.
func (input *Input) Hash() []byte {
	hashed := blake2b.Sum256(input.Bytes())
	return hashed[:]
}


//序列化为字节数组
func (input *Input) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(input)
	utils.Handle(err)

	return encoded.Bytes()
}

//反序列化
func DeserializeInput(data []byte) Input {

	var input Input

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&input)
	utils.Handle(err)

	return input
}

//获取input的哈希值
func (input *Input) GetHash() []byte {
	var hash [32]byte
	inputCopy := *input

	hash = sha256.Sum256(inputCopy.Serialize())
	return hash[:]
}