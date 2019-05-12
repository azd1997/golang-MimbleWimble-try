package blockchain2

import "bytes"

// TxKernelList sortable list of kernels
type TxKernelList []TxKernel

func (m TxKernelList) Len() int {
	return len(m)
}

// Less is used to order kernels by their hash.
func (m TxKernelList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m TxKernelList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}