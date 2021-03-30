package main

import (
	"fmt"
	"math"
)

/*
Test fil til at teste go kode uden at køre hele programmet xd
*/



func main() {


	fmt.Println(bitMult(convertXToByte(1), convertXToByte(1)))
	a := []int{0, 1, 0, 1, 0, 0, 1, 1}
	//b := []int{1, 1, 0, 0, 1, 0, 1, 0}
	fmt.Println("Inverse of a is:", findInverseBit(a))
	/*
	a1 := []int{0, 0, 1, 1}
	a2 := []int{0, 0, 1, 0}


	fmt.Println(sliceSubtraction(a1, a2))

	 */

	/*	fmt.Println(calcT(3))


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
	fmt.Println("Reconstructed original share", Shamir.Reconstruct(test))*/
	//shares := make(map[int][]int)
	//shares[3] = []int{6}
	//shares[4] = []int{6}
	//shares[5] = []int{8}
	//secretSharing.ComputeFunction(shares, 1)

	//fmt.Println(permutationsInts)
	//fmt.Println(math.Pow(2,3))
	//fmt.Println(findInverse(-2, 11))

}

func bitAdd(b1 []int, b2 []int) []int {

	bitRes := make([]int, len(b1))

	for i := 0; i < len(b1); i++ {
		bitRes[i] = b1[i] ^ b2[i]
	}

	return bitRes
}

func intToBinaryArray(number, arraySize int) []int {
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

func reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func convertXToByte(x int) []int {
	return reverse(intToBinaryArray(x, 8))
}

func bitMult(b1, b2 []int) []int {
	irreducible := makeIrreducible()
	interRes := make([]int, len(b1) * 2 - 1)

	for i := 0; i < len(b1); i++ {
		for j := 0; j < len(b2); j++ {
			interRes[i + j] = interRes[i + j] ^ (b1[i] & b2[j])
		}
	}

	for i := 0; i < 7; i++ {
		if interRes[i] == 1 {
			sliceSubtraction(interRes, irreducible)
		}

		irreducible = bitShiftRight(irreducible)
	}

	interRes = interRes[7:]

	return interRes
}

func makeIrreducible() []int{
	irreducible := make([]int, 15)
	irreducible[8] = 1
	irreducible[7] = 1
	irreducible[5] = 1
	irreducible[4] = 1
	irreducible[0] = 1
	return irreducible
}

func bitShiftRight(array []int) []int {
	length := len(array)
	result := make([]int, length)
	result = append(result[:1], array[: length - 1]...)
	return result
}

func sliceSubtraction(a1 []int, a2 []int) []int {
	for i := 0; i < len(a1); i++ {
		a1[i] = a1[i] ^ a2[i]
	}

	return a1
}


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
func findItohInverse(a []int) int {
	//var prime = 2
	//var m = 8
	var r = bitMult(convertXToByte(255), convertXToByte(1)) //r = (p^m - 1)/ p - 1 -> 255 * 1
	fmt.Println(bitAdd(r, convertXToByte(1)))
/*	var ar1 = int(math.Pow(float64(a), float64(r - 1))) % prime
	var ar = ar1 * a
	var arp1 = (1 / ar) % prime
	var a1 = arp1 * ar1*/
	return 2
}

func findInverseBit(a []int) []int {
	result := a
	for i := 1; i < 255; i++ {
		result = bitMult(bitExponent(a, 2), result)
	}
	return result
}

func bitExponent(byte []int, x int) []int {
	result := []int{0, 0, 0, 0, 0, 0, 0, 1}
	for i := 1; i <= x; i++ {
		result = bitMult(result, byte)
	}
	return result
}
