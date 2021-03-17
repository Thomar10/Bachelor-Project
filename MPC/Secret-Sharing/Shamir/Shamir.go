package Shamir

import (
	finite "MPC/Finite-fields"
	"fmt"
	"math"
	"math/rand"
)

type Shamir struct {

}
var function string
func (s Shamir) SetFunction(f string) {
	function = f
}
func ComputeResultt(shares map[int]int, parties int) int {
	testerdester := make(map[int]int)
	if len(shares) > parties {
		i := 0
		for k, v := range shares {
			testerdester[k] = v
			i++
			if i == 2 {
				break
			}
		}
		return Reconstruct(testerdester)
	} else {
		return Reconstruct(shares)
	}

}

func (s Shamir) ComputeResult(ints []int) int {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
	//return Reconstruct(shares)
}

var field finite.Finite

func (s Shamir) SetField(f finite.Finite) {
	field = f
}

func (s Shamir) ComputeShares(parties, secret int) []int {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
	//3 + 4x + 2x^2
	//[3, 4, 2]
	var polynomial = make([]int, t + 1)
	polynomial[0] = secret

	for i := 1; i < t + 1; i++ {
		polynomial[i] = rand.Intn(field.GetSize())
	}
	fmt.Println("Poly after loop", polynomial)
	var shares = make([]int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}
	fmt.Println("The polynomial!:", polynomial)
	fmt.Println("The shares for the poly:", shares)
	return shares
}

func calculatePolynomial(polynomial []int, x int) int {
	var result = 0

	for i := 0; i < len(polynomial); i++ {
		result += polynomial[i] * int(math.Pow(float64(x), float64(i)))
	}

	return result % field.GetSize()
}

func (s Shamir) ComputeFunction(shares map[int][]int, party int) []int {
	//Reconstruct(shares)
	if function == "add" {

	}
	return nil
}
