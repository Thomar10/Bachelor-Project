package Simple_Sharing

import (
	finite "MPC/Finite-fields"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
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
		keys := reflect.ValueOf(shares).MapKeys()
		var keysArray []int
		for _, k := range keys {
			keysArray = append(keysArray, (k.Interface()).(int))
		}
		sort.Ints(keysArray)
		size := keysArray[0]
		//Vi kan kun gange 2 - parties er ikke altid 1 og 2, men vælger laveste som 1 og næste som 2
		party1 := keysArray[0]
		party2 := keysArray[1]
		i := party - 1
		for j := 0; j < len(shares[size]); j++ {
			fmt.Println("J is:", j)
			//Sidste party
			if i == len(shares[size]) {
				result[0] += (shares[party1][0] * shares[party2][j]) % field.GetSize()
			}else {
				result[0] += (shares[party1][i] * shares[party2][j]) % field.GetSize()
			}

		}
		if i - 1 < 0 {
			result[0] += (shares[party1][len(shares[size]) +  (i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
		} else if i - 2 < 0 {
			result[0] += (shares[party1][(i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
		} else {
			result[0] += (shares[party1][(i - 1)] * shares[party2][(i - 2)]) % field.GetSize()
		}
		//if result[0] < 0 {
		//	result[0] = field.GetSize() + result[0] % field.GetSize()
		//}else {
		//	result[0] = result[0] % field.GetSize()
		//}
		fmt.Println("result is ", result[0])
		//result[0] = 0
		////TODO remove hardcoding
		////TODO Virker ikke for hvor result[0] er negativ - hmm
		//if party == 1 {
		//	result[0] = shares[1][0] * shares[2][0] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		//} else if party == 2 {
		//	result[0] = shares[1][1] * shares[2][1] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		//} else if party == 3 {
		//	result[0] = shares[1][0] * shares[2][0] + shares[1][0] * shares[2][1] + shares[1][1] * shares[2][0]
		//}

		result[0] = result[0] % field.GetSize()
		fmt.Println("result is ", result[0])
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