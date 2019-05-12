package blockchain2

import "bytes"

//排好序的inputList
type InputList []Input

func (m InputList) Len() int {
	return len(m)
}

// Less is used to order inputs by their hash.
func (m InputList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m InputList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}