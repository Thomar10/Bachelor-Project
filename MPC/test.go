package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
)

/*
Test fil til at teste go kode uden at køre hele programmet xd
*/



func main() {

	/*openGates := make([][]int, 3)
	for i, _ := range openGates {
		openGates[i] = []int{1+i, 2+i, 3+i}
	}
	fmt.Println(openGates[2:])*/

	createdCircuit := createCircuit(4)
	file, _ := json.MarshalIndent(createdCircuit, "", " ")

	_ = ioutil.WriteFile("test.json", file, 0644)







	//Hyper test
	alpha := []int{1, 2, 3}
	beta := []int{4, 5, 6}
	matrix := make([][]int, 3)
	for i, _ := range matrix {
		matrix[i] = make([]int, 3)
		for j, _ := range matrix {
			matrix[i][j] = 1
		}
	}
	//fmt.Println(matrix)
	for i, _ := range matrix {
		for j, _ := range matrix {
			for k := 0; k < 3; k++ {
				if k == j {
					continue
				}else {
					matrix[i][j] = ((beta[i] - alpha[k]) / (alpha[j] - alpha[k]) * matrix[i][j]) % 17
				}
			}
			if matrix[i][j] < 0 {
				matrix[i][j] = 17 + matrix[i][j]
			}
		}
	}
	//fmt.Println(matrix)
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
