package Simple_Sharing

import (
	finite "MPC/Finite-fields"
	"math/rand"
)

var field finite.Finite
var function string

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

func (s Simple_Sharing) SetFunction(f string) {
	function = f
}

func (s Simple_Sharing) ComputeFunction(shares map[int][]int, party int) []int {
	resultSize := len(shares[1])
	result := make([]int, resultSize)
	if function == "add" {
		for i := 0; i < resultSize; i++ {
			for _, share := range shares {
				//TODO Skift add om til field.Add - samt ændre hardcoding generelt
				result[i] += share[i]
			}
			result[i] = result[i] % field.GetSize()
		}
	} else if function == "multiply" {
		//TODO remove hardcoding
		if party == 1 {
			result[0] = shares[1][0] * shares[2][0] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		} else if party == 2 {
			result[0] = shares[1][1] * shares[2][1] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		} else if party == 3 {
			result[0] = shares[1][0] * shares[2][0] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		}

		result[0] = result[0] % field.GetSize()
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