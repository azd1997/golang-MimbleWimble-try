package blockchain2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yoss22/bulletproofs"
	"io"
	"sort"
)

type BlockList []Block

// 区块结构体
type Block struct {
	Header BlockHeader

	Inputs  InputList
	Outputs OutputList
	Kernels TxKernelList
}

// Bytes implements p2p Message interface
//对区块中数据序列化为字节数组
func (b *Block) Bytes() []byte {
	buff := new(bytes.Buffer)
	if _, err := buff.Write(b.Header.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	// 值得注意的是，写入了inputs/outputs/kernels的数目
	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Inputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Outputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Kernels))); err != nil {
		logrus.Fatal(err)
	}

	// mimblewimble协议要求Inputs/Outputs/Kernel列表中按大小排好序，猜测是为了打乱输入输出，隐藏输入输出关系
	sort.Sort(b.Inputs)
	sort.Sort(b.Outputs)
	sort.Sort(b.Kernels)

	// 写入inputs
	for _, input := range b.Inputs {
		if _, err := buff.Write(input.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	// 写入outputs
	for _, output := range b.Outputs {
		if _, err := buff.Write(output.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	// 写入kernels
	for _, txKernel := range b.Kernels {
		if _, err := buff.Write(txKernel.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Type implements p2p Message interface
// 返回消息类型（区块）
func (b *Block) Type() uint8 {
	return MsgTypeBlock
}

// Read implements p2p Message interface
// 读取序列化的区块中数据，将之分离
func (b *Block) Read(r io.Reader) error {
	// 调用blockHeader的读方法，先把区块头数据读出来
	if err := b.Header.Read(r); err != nil {
		return err
	}

	// 紧接着读到的是 inputs/outputs/kernels的数目
	var inputs, outputs, kernels uint64		//binary.BigEndian大端字节序实现 参考：https://blog.csdn.net/qq_33724710/article/details/51056542
	if err := binary.Read(r, binary.BigEndian, &inputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &outputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &kernels); err != nil {
		return err
	}

	// 检查inputs/outputs/kernels的数目是否异常
	if inputs > 1000000 {
		return errors.New("transaction contains too many inputs")
	}
	if outputs > 1000000 {
		return errors.New("transaction contains too many outputs")
	}
	if kernels > 1000000 {
		return errors.New("transaction contains too many kernels")
	}

	// 根据前边得到的长度读取inputs
	b.Inputs = make([]Input, inputs)	//根据长度初始化数组
	for i := uint64(0); i < inputs; i++ {
		if err := b.Inputs[i].Read(r); err != nil {
			return err
		}
	}

	// outputs
	b.Outputs = make([]Output, outputs)
	for i := uint64(0); i < outputs; i++ {
		if err := b.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	// kernels
	b.Kernels = make([]TxKernel, kernels)
	for i := uint64(0); i < kernels; i++ {
		if err := b.Kernels[i].Read(r); err != nil {
			return err
		}
	}

	return nil
}

// String implements String() interface
//将区块内容按格式转换为字符串
func (b Block) String() string {
	return fmt.Sprintf("%#v", b)	// %#v	值的Go语法表示
}

// 计算区块哈希，只算header哈希
func (b *Block) Hash() Hash {
	return b.Header.Hash()
}

// Validate returns nil if block successfully passed BLOCK-SCOPE consensus rules
func (b *Block) Validate() error {
	logrus.Info("block scope validate")
	/*
		TODO: implement it:

		verify_weight()
		verify_sorted()
		verify_coinbase()
		verify_kernels()

	*/

	// 验证区块头内容及其中的工作量证明
	if err := b.Header.Validate(); err != nil {
		return err
	}

	// Check that consensus rule MaxBlockCoinbaseOutputs & MaxBlockCoinbaseKernels
	if len(b.Outputs) == 0 || len(b.Kernels) == 0 {
		return errors.New("invalid nocoinbase block")
	}

	// Check sorted inputs, outputs, kernels
	if err := b.verifySorted(); err != nil {
		return err
	}

	if err := b.verifyCoinbase(); err != nil {
		return err
	}

	// Verify all output values are within the correct range.
	if err := b.verifyRangeProofs(); err != nil {
		return err
	}

	if err := b.verifyKernels(); err != nil {
		return err
	}

	return nil
}

func (b *Block) verifyCoinbase() error {
	coinbase := 0

	for _, output := range b.Outputs {
		if output.Features&CoinbaseOutput == CoinbaseOutput {
			coinbase++	//一旦output中为coinbase输出，则计数器加一

			//如果coinbase输出数多于最大值（1），则报错
			if coinbase > MaxBlockCoinbaseOutputs {
				return errors.New("invalid block with few coinbase outputs")
			}

			// 对找到的这笔coinbase输出进行验证
			if err := output.Validate(); err != nil {
				return err
			}
		}
	}

	// Check the roots
	// TODO: do that

	return nil
}

func (b *Block) verifyKernels() error {
	coinbase := 0

	//找到coinbase核并验证
	for _, kernel := range b.Kernels {
		if kernel.Features&CoinbaseKernel == CoinbaseKernel {
			coinbase++

			if coinbase > MaxBlockCoinbaseKernels {
				return errors.New("invalid block with few coinbase kernels")
			}

			// Validate kernel
			if err := kernel.Validate(); err != nil {
				return err
			}
		}
	}

	// TODO: Verify that the kernel sums are correct.

	// Check the roots
	// TODO: do that

	return nil
}

// 检查inputs, outputs, kernels是否已排好序
func (b *Block) verifySorted() error {
	if !sort.IsSorted(b.Inputs) {
		return errors.New("block inputs are not sorted")
	}

	if !sort.IsSorted(b.Outputs) {
		return errors.New("block outputs are not sorted")
	}

	if !sort.IsSorted(b.Kernels) {
		return errors.New("block kernels are not sorted")
	}

	return nil
}

// 检查所有输出是否都符合数值范围要求.
func (b *Block) verifyRangeProofs() error {
	// TODO(yoss22): Batch verify these.
	prover := bulletproofs.NewProver(64)
	for _, output := range b.Outputs {
		if !prover.Verify(output.Commit, output.RangeProof) {
			return fmt.Errorf("proof verification failed for %v %v",
				output.Commit, output.RangeProof)
		}
	}
	return nil
}

