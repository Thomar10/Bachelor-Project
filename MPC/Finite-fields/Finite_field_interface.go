package Finite_fields

import (
	"math/big"
)

type Finite interface {
	GenerateField() Number
	SetSize(f Number)
	GetSize() Number
	InitSeed()
	Add(n1, n2 Number) Number
	Mul(n1, n2 Number) Number
	ComputeShares(parties int, secret Number) []Number
	FindInverse(a, prime Number) Number
	GetConstant(constantString int) Number
	FilledUp([]Number) bool
}
type Number struct {
	Prime *big.Int
	Binary []int
}
