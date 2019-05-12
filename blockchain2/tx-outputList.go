package blockchain2

import "bytes"

type OutputList []Output


func (m OutputList) Len() int {
	return len(m)
}

// Less is used to order outputs by their hash.
func (m OutputList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m OutputList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}