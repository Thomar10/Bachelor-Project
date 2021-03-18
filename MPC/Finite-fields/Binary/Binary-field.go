package Binary

import "math/big"

type Binary struct {

}

func (b Binary) InitSeed() {
	panic("implement me")
}

func (b Binary) SetSize(f *big.Int) {
	panic("implement me")
}

func (b Binary) ComputeShares(parties, secret int) []*big.Int {
	panic("implement me")
}

//TODO implement
func (b Binary) GenerateField() *big.Int {
	panic("imeplement ne")
}

func (b Binary) GetSize() *big.Int {
	panic("implement me")
}