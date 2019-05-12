package blockchain2

// Transaction an grin transaction
type Transaction struct {
	// The "k2" kernel offset.
	KernelOffset [32]byte
	// Set of inputs spent by the transaction
	Inputs InputList
	// Set of outputs the transaction produces
	Outputs OutputList
	// The kernels for this transaction
	Kernels TxKernelList
}