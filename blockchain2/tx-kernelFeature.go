package blockchain2

// KernelFeatures is options for a kernel's structure or use
type KernelFeatures uint8

const (
	// No flags
	DefaultKernel KernelFeatures = 0
	// Kernel matching a coinbase output
	CoinbaseKernel KernelFeatures = 1 << 0
)

func (f KernelFeatures) String() string {
	switch f {
	case DefaultKernel:
		return ""
	case CoinbaseKernel:
		return "Coinbase"
	}
	return ""
}
