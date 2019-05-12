package blockchain

type OutputFeature uint8

// BlockID identify block by Hash or/and Height (if not nill)
type BlockID struct {
	// Block hash, if nil - use the height
	Hash Hash
	// Block height, if nil - use the hash
	Height *uint64
}