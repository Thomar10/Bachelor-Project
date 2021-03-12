package main

import (
	"fmt"
	"math"
)

/*
Test fil til at teste go kode uden at køre hele programmet xd
 */
func main() {
	var keys = []int{4, 5, 6, 7}
	var binarySize = int(math.Pow(2, float64(len(keys))) - 1)
	//https://stackoverflow.com/questions/7150035/calculating-bits-required-to-store-decimal-number#:~:text=12%20Answers&text=Well%2C%20you%20just%20have%20to,%3D%201024%20(10%20bits).
	//Udregn antal bits man skal have
	need := int(math.Ceil(math.Log10(float64(binarySize)) / math.Log10(2)))
	permutationsBinary := make([][]int, binarySize - 1)
	permutationsInts := make([][]int, binarySize - 1)
	//make [][]int laver ikke det inderste array, så de skal også laves :/
	for i := 0; i < binarySize - 1; i++ {
		permutationsInts[i] = make([]int, len(keys))
	}
	for i := 1; i < binarySize; i++ {
		permutationsBinary[i - 1] = IntToBinaryArray(i, need)
	}
	for i := 0; i < len(permutationsBinary); i++ {
		for j := 0; j < len(permutationsBinary[0]); j++ {
			if permutationsBinary[i][j] == 1 {
				permutationsInts[i][j] = keys[j]
			}
		}
	}
	//for i, perm := range permutationsInts {
	//	permutationsInts[i] =
	//}

	fmt.Println(permutationsInts)
	//fmt.Println(math.Pow(2,3))
	//fmt.Println(findInverse(-2, 11))

}


//https://stackoverflow.com/questions/8151435/integer-to-binary-array/8151674
func IntToBinaryArray(number, arraySize int) []int {
	result := make([]int, arraySize)
	for i := 0; i < arraySize; i++ {
		if number&(1<<uint8(i)) != 0 {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}

	return result
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
