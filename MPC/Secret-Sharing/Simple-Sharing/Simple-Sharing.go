package Simple_Sharing

import (
	finite "MPC/Finite-fields"
	"math/rand"
)

var field finite.Finite

type Simple_Sharing struct {

}

func (s Simple_Sharing) SetField(f finite.Finite) {
	field = f
}


func (s Simple_Sharing) ComputeShares(parties, secret int) []int {
	return field.ComputeShares(parties, secret)
/*	var shares []int
	//Create the n - 1 random shares
	for s := 1; s < parties; s++ {
		shares = append(shares, randomNumberInZ(prime - 1))
	}
	//Create the nth share
	for share := range shares {
		secret -= share
	}

	shares = append(shares, secret % prime)

	return shares*/
}

func randomNumberInZ(prime int) int {
	return rand.Intn(prime)
}

func (s Simple_Sharing) ComputeFunction(shares []int) int {
	result := 0
	for _, share := range shares {
		result += share
	}
	return result % field.GetSize()
}

func (s Simple_Sharing) ComputeResult(results []int) int {
	result := 0
	for _, r := range results {
		result += r
	}
	return result % field.GetSize()
}