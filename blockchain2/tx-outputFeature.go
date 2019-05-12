package blockchain2


// OutputFeatures is options for block validation
type OutputFeatures uint8

const (
	// No flags
	DefaultOutput OutputFeatures = 0
	// Output is a coinbase output, must not be spent until maturity
	CoinbaseOutput OutputFeatures = 1 << 0
)

func (f OutputFeatures) String() string {
	switch f {
	case DefaultOutput:
		return ""
	case CoinbaseOutput:
		return "Coinbase"
	}
	return ""
}