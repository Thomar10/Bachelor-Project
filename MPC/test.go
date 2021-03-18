package main

import (
	"math"
)

/*
Test fil til at teste go kode uden at køre hele programmet xd
*/

/*
func main() {

	fmt.Println(calcT(3))


	secretSharing := Shamir.Shamir{}
	finiteField := Prime.Prime{}
	//finiteField.SetSize(3780287809)
	finiteField.SetSize(137)
	secretSharing.SetField(finiteField)
	shares := secretSharing.ComputeShares(8, 5)

	fmt.Println(shares)
	test := make(map[int]int)
	for i := 1; i < len(shares); i++ {
		test[i] = shares[i - 1]
	}
	fmt.Println(test)
	fmt.Println("Reconstructed original share", Shamir.Reconstruct(test))
	//shares := make(map[int][]int)
	//shares[3] = []int{6}
	//shares[4] = []int{6}
	//shares[5] = []int{8}
	//secretSharing.ComputeFunction(shares, 1)

	//fmt.Println(permutationsInts)
	//fmt.Println(math.Pow(2,3))
	//fmt.Println(findInverse(-2, 11))

}
*/

func calcT(parties int) int {
	return (parties - 1) / 2
}

func findInverse(a int, prime int) int {
	if a < 0 {
		a = prime + a
	}
	return int(math.Pow(float64(a), float64(prime - 2))) % prime
}

//Itoh–Tsujii inversion algorithm
func findItohInverse(a int, prime int) int {
	var r = (prime - 1) / (prime - 1)
	var ar1 = int(math.Pow(float64(a), float64(r - 1))) % prime
	var ar = ar1 * a
	var arp1 = (1 / ar) % prime
	var a1 = arp1 * ar1
	return a1
}
