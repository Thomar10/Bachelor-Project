package Shamir

import (
	finite "MPC/Finite-fields"
	"math"
	"math/rand"
	"time"
)

type Shamir struct {

}
var function string

func (s Shamir) SetFunction(f string) {
	function = f
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

	var shares = make([]int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}

	return shares
}


func ComputeResult(shares map[int]int, parties int) int {
	return Reconstruct(shares)

}

func (s Shamir) ComputeResult(ints []int) int {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
	//return Reconstruct(shares)
}

var field finite.Finite

func (s Shamir) SetField(f finite.Finite) {
	rand.Seed(time.Now().UnixNano())
	field = f
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
