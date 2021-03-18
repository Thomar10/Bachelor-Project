package Secret_Sharing

import (
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"math/big"
)

type Secret_Sharing interface {
	SetField(field finite.Finite)
	ComputeFunction(map[int][]*big.Int, int) []*big.Int
	SetFunction(f string)
	TheOneRing(circuit Circuit.Circuit, secret int) int
	ComputeShares(parties int, secret *big.Int) []*big.Int
	ComputeResult([]*big.Int) int
}