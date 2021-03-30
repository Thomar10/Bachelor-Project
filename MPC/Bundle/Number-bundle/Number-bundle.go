package Number_bundle

import (
	Finite_fields "MPC/Finite-fields"
	"math/big"
)

type NumberBundle struct {
	ID string
	Type string
	Prime *big.Int
	Shares []Finite_fields.Number
	Result Finite_fields.Number
	From int
	Gate int
}
