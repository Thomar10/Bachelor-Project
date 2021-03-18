package Finite_fields

import "math/big"

type Finite interface {
	GenerateField() *big.Int
	SetSize(f *big.Int)
	GetSize() *big.Int
	InitSeed()
}