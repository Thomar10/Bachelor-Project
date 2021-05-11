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
	ComputeShares(parties int, secret Number, corrupts int) []Number
	CreateRandomNumber() Number
	ConvertLastShare(Number) Number
	FindInverse(a, prime Number) Number
	GetConstant(constantString int) Number
	FilledUp([]Number) bool
	CalcPoly(poly []Number, x int) Number
	CompareEqNumbers(share, polyShare Number) bool
	HaveEnoughForReconstruction(outputs, corrupt int, resultGate map[int]map[int] Number) bool
	ComputeFieldResult(outputSize int, polynomials [][]Number) Number
	CheckPolynomialIsConsistent(resultGate map[int]map[int]Number, corrupts int, reconstructFunction func(map[int]Number, int) []Number) (bool, [][]Number)
	ConstructFieldSecret(secret Number,doesIHaveAnInput bool, partySize int, corrupts int, partyNumber int) ([][]Number, []int)
}
type Number struct {
	Prime *big.Int
	Binary []int
}
