package Finite_fields

type Finite interface {
	GenerateField() int
	ComputeShares(parties, secret int) []int
	SetSize(f int)
	GetSize() int
	InitSeed()
}