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

func (s Simple_Sharing) ComputeFunction(shares map[int][]int) []int {
	resultSize := len(shares[1])
	result := make([]int, resultSize)
	for i := 0; i < resultSize; i++ {
		for _, share := range shares {
			//TODO Skift add om til field.Add - samt ændre hardcoding generelt
			result[i] += share[i]
		}
		result[i] = result[i] % field.GetSize()
	}
	return result
}

func (s Simple_Sharing) ComputeResult(results []int) int {
	result := 0
	for _, r := range results {
		result += r
	}
	return result % field.GetSize()
}