package Prime_bundle

import "math/big"

type PrimeBundle struct {
	ID string
	Type string
	Prime *big.Int
	Shares []*big.Int
	Result *big.Int
	From int
	Gate int
}
