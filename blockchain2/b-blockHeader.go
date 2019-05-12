package blockchain2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/azd1997/golang-MimbleWimble-try/secp256k1zkp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"io"
	"time"
)

// 区块头定义
type BlockHeader struct {

	// 区块高度
	Height uint64
	// 前一个区块的哈希
	Previous Hash
	// 前个区块头的MMR的根哈希		MMR（Merkle Mountain Range） 参考 Flyclient 协议
	PreviousRoot Hash
	// 时间戳
	Timestamp time.Time
	// UTXO集中所有承诺的Merkle根
	UTXORoot Hash
	// UTXO集中所有RangeProof的Merkle根
	RangeProofRoot Hash
	// UTXO集中所有交易核的Merkle根
	KernelRoot Hash
	// 挖矿随机数
	Nonce uint64
	// 自创世区块起所有交易核偏移量总和
	TotalKernelOffset Hash
	// Total accumulated sum of kernel commitments since genesis block.
	// Should always equal the UTXO commitment sum minus supply.
	TotalKernelSum secp256k1zkp.Commitment
	// Total size of the output MMR after applying this block
	OutputMmrSize uint64
	// Total size of the kernel MMR after applying this block
	KernelMmrSize uint64
	// 工作量证明
	POW Proof
	// TODO: Remove or calculate this correctly.
	// 挖矿困难度
	Difficulty Difficulty

}

// 区块头的哈希是其工作量证明的哈希
func (b *BlockHeader) Hash() Hash {
	hash := blake2b.Sum256(b.POW.ProofBytes())

	return hash[:]
}

// 序列化除了POW以外的内容，用在Hash()方法中
func (b *BlockHeader) bytesWithoutPOW() []byte {
	buff := new(bytes.Buffer)

	// 写入 height of block
	if err := binary.Write(buff, binary.BigEndian, b.Height); err != nil {
		logrus.Fatal(err)
	}

	// 写入 timestamp
	if err := binary.Write(buff, binary.BigEndian, b.Timestamp.Unix()); err != nil {
		logrus.Fatal(err)
	}

	// 写入 prev blockhash
	if len(b.Previous) != BlockHashSize {
		logrus.Fatal(errors.New("invalid previous block hash len"))
	}

	if _, err := buff.Write(b.Previous); err != nil {
		logrus.Fatal(err)
	}

	if len(b.PreviousRoot) != BlockHashSize {
		logrus.Fatal(errors.New("invalid previous root hash len"))
	}

	if _, err := buff.Write(b.PreviousRoot); err != nil {
		logrus.Fatal(err)
	}

	// 写入 UTXORoot, RangeProofRoot, KernelRoot
	if len(b.UTXORoot) != BlockHashSize ||
		len(b.RangeProofRoot) != BlockHashSize ||
		len(b.KernelRoot) != BlockHashSize {
		logrus.Fatal(errors.New("invalid UTXORoot/RangeProofRoot/KernelRoot len"))
	}

	if _, err := buff.Write(b.UTXORoot); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(b.RangeProofRoot); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(b.KernelRoot); err != nil {
		logrus.Fatal(err)
	}

	//写入

	if _, err := buff.Write(b.TotalKernelOffset); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.OutputMmrSize); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.KernelMmrSize); err != nil {
		logrus.Fatal(err)
	}



	// 写入 nonce
	if err := binary.Write(buff, binary.BigEndian, b.Nonce); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

//序列化区块头内POW数据
func (b *BlockHeader) bytesPOW() []byte {
	return b.POW.Bytes()
}

// 结合前二者，返回完整的区块头序列化字节
func (b *BlockHeader) Bytes() []byte {
	var buff bytes.Buffer
	buff.Write(b.bytesWithoutPOW())
	buff.Write(b.bytesPOW())

	return buff.Bytes()
}

// 读取区块头数据
func (b *BlockHeader) Read(r io.Reader) error {

	// Read height of block
	if err := binary.Read(r, binary.BigEndian, &b.Height); err != nil {
		return err
	}

	// Read timestamp
	var ts int64
	if err := binary.Read(r, binary.BigEndian, &ts); err != nil {
		return err
	}

	// FIXME: Check timestamp is in correct range.
	b.Timestamp = time.Unix(ts, 0).UTC()

	// Read prev blockhash
	b.Previous = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.Previous); err != nil {
		return err
	}

	b.PreviousRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.PreviousRoot); err != nil {
		return err
	}

	// Read UTXORoot, RangeProofRoot, KernelRoot
	b.UTXORoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.UTXORoot); err != nil {
		return err
	}

	b.RangeProofRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.RangeProofRoot); err != nil {
		return err
	}

	b.KernelRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.KernelRoot); err != nil {
		return err
	}

	b.TotalKernelOffset = make([]byte, secp256k1zkp.SecretKeySize)
	if _, err := io.ReadFull(r, b.TotalKernelOffset); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.OutputMmrSize); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.KernelMmrSize); err != nil {
		return err
	}


	if err := binary.Read(r, binary.BigEndian, &b.Nonce); err != nil {
		return err
	}

	if err := b.POW.Read(r); err != nil {
		return err
	}

	return nil
}

// Validate returns nil if header successfully passed consensus rules
func (b *BlockHeader) Validate() error {
	logrus.Info("block header validate")


	// refuse blocks more than 12 blocks intervals in future (as in bitcoin)
	// 拒绝超过12个区块间隔的区块（将验证时时间戳与区块时间戳作差，若差值超过12个区块的时间则拒绝检查。每个区块的出块时间被设定为60s）
	if b.Timestamp.Sub(time.Now().UTC()) > time.Second*12*BlockTimeSec {
		return fmt.Errorf("invalid block time (%s)", b.Timestamp)
	}

	// TODO: Check difficulty.

	// Check POW
	isPrimaryPow := b.POW.EdgeBits != SecondPowEdgeBits

	// Either the size shift must be a valid primary POW (greater than the
	// minimum size shift) or equal to the secondary POW size shift.
	if b.POW.EdgeBits < DefaultMinEdgeBits && isPrimaryPow {
		return fmt.Errorf("cuckoo size too small: %d", b.POW.EdgeBits)
	}

	if err := b.POW.Validate(b, b.POW.EdgeBits); err != nil {
		return err
	}

	return nil
}


// String implements String() interface
func (b BlockHeader) String() string {
	return fmt.Sprintf("%#v", b)
}

//TODO:已经删除原本的version
