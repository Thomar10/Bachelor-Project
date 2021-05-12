package Shamir

import (
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"math"
	"math/big"
	"reflect"
	"sort"
)


//Checks if a share is on a given polynomial
func ShareIsOnPolynomial(share finite.Number, poly []finite.Number, fromParty int) bool {
	polyShare := field.CalcPoly(poly, fromParty)
	return field.CompareEqNumbers(share, polyShare)
}

//Returns the keys sorted from the shares map
func getKeysFromMap(shares map[int]finite.Number) []int {
	keys := reflect.ValueOf(shares).MapKeys()
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}
	sort.Ints(keysArray)
	return keysArray
}


//Reconstructs a polynomial of degree describing shares given
func ReconstructPolynomial(shares map[int]finite.Number, degree int) []finite.Number {
	keysArray := getKeysFromMap(shares)
	originalF := make([]finite.Number, degree + 1)
	for i, _ := range originalF {
		originalF[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
	}
	for i := 0; i <= degree; i++ {
		deltai := computeFullDelta(keysArray[i], keysArray[:degree + 1])
		share := shares[keysArray[i]]
		for j, num := range deltai {
			deltai[j] = field.Mul(share, num)
			originalF[j] = field.Add(deltai[j], originalF[j])
		}
	}
	return originalF
}

//Reconstructs the secret (f(0)) on for a given set of shares.
func Reconstruct(shares map[int]finite.Number) finite.Number {
	keysArray := getKeysFromMap(shares)
	var secret = finite.Number{Prime: big.NewInt(0), Binary: []int{0, 0, 0, 0, 0, 0, 0, 0}}
	for _, key := range keysArray {
		delta := computeDelta(key, keysArray)
		share := shares[key]
		var interRes = field.Mul(share, delta)
		secret = field.Add(interRes, secret)
	}
	return secret
}

//Calculates the full delta for a given key
func computeFullDelta(key int, keys []int) []finite.Number {
	var talker = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	//Calculate the denominator (talker)
	for _, j := range keys {
		if j == key {
			continue
		}
		keyNumberBinary := field.Add(
			finite.Number{
				Prime: new(big.Int).Neg(big.NewInt(int64(j))),
				Binary: Binary.ConvertXToByte(0)},
			finite.Number{
				Prime: field.GetSize().Prime,
				Binary: Binary.ConvertXToByte(j)})

		var keyNumber = finite.Number{Prime: keyNumberBinary.Prime, Binary: keyNumberBinary.Binary}

		var jNumber = finite.Number{Prime: big.NewInt(int64(key)), Binary: Binary.ConvertXToByte(key)}

		interRes := field.Add(keyNumber, jNumber)
		talker = field.Mul(talker, interRes)
	}
	var inverseTalker = field.FindInverse(talker, field.GetSize())
	//Remove the key from the list to calculate the numerator
	keysWithoutkey := removeElement(keys, key)

	permutations := computePermutations(keysWithoutkey)
	var polynomial = make([]finite.Number, len(keys))
	//keys = 3, 4, 5, 6 -> delta3
	//(x-4)(x-5)(x-6)
	// -120 + 74x -15x^2+x^3
	//[-120, 74, -15, 1]
	//Calculate all the indexes except first and last index in polynomial
	for i := 1; i < len(polynomial) - 1; i++ {
		polynomial[i] = multipleAllWithSize(len(keysWithoutkey) - i, permutations)
	}
	//We need to calculate the first index and last index in the polynomial
	polynomial[0] = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	//Last index will always just be 1 (x^degree)
	polynomial[len(polynomial) - 1] = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	//First index is just all the elements in C negated and multiplied together
	for _, number := range keysWithoutkey {
		negNumber := field.Add(
			finite.Number{
				Prime: big.NewInt(int64(-number)),
				Binary: Binary.ConvertXToByte(0)},
			finite.Number{
				Prime: field.GetSize().Prime,
				Binary: Binary.ConvertXToByte(number)})
		polynomial[0] = field.Mul(polynomial[0], negNumber)
	}

	//Multiply all the indexes in the polynomial with the inverse numerator
	for i, number := range polynomial {
		polynomial[i] = field.Mul(number, inverseTalker)
	}
	return polynomial
}


//Compute delta for d(0)
func computeDelta(key int, keys []int) finite.Number {
	//Calculate the denominator (talker)
	var talker = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	for _, j := range keys {
		if j == key {
			continue
		}
		keyNumberBinary := field.Add(
			finite.Number{
				Prime: new(big.Int).Neg(big.NewInt(int64(j))),
				Binary: Binary.ConvertXToByte(0)},
			finite.Number{
				Prime: field.GetSize().Prime,
				Binary: Binary.ConvertXToByte(j)})

		var keyNumber = finite.Number{Prime: keyNumberBinary.Prime, Binary: keyNumberBinary.Binary}
		var jNumber = finite.Number{Prime: big.NewInt(int64(key)), Binary: Binary.ConvertXToByte(key)}
		interRes := field.Add(keyNumber, jNumber)
		talker = field.Mul(talker, interRes)
	}
	var inverseTalker = field.FindInverse(talker, field.GetSize())
	//Remove the key from the list to calculate the numerator
	keysWithoutkey := removeElement(keys, key)
	//Calculate the numerator (simply -j multiplied together with each other for all keys in keysWithoutKey)
	var polynomial = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	for _, number := range keysWithoutkey {
		var numberNumber = field.Add(
			finite.Number{
				Prime: big.NewInt(int64(-number)),
				Binary: Binary.ConvertXToByte(0)},
			finite.Number{
				Prime: field.GetSize().Prime,
				Binary: Binary.ConvertXToByte(number)})
		polynomial = field.Mul(polynomial, numberNumber)
	}
	polynomial = field.Mul(polynomial, inverseTalker)
	return polynomial
}

//Multiplies all the permutation sets of a given size with each other
//Ex [[2, 3], [3, 4], [1]] -> (-2*-3)+(-3*-4)
func multipleAllWithSize(k int, permutations [][]int) finite.Number {
	result := finite.Number{Prime: big.NewInt(0), Binary: []int{0, 0, 0, 0, 0, 0, 0, 0}}
	for _, perm := range permutations {
		if len(perm) == k {
			subresult := finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
			for _, number := range perm {
				negNumber := field.Add(
					finite.Number{
						Prime: new(big.Int).Neg(big.NewInt(int64(number))),
						Binary: Binary.ConvertXToByte(0)},
					finite.Number{
						Prime: field.GetSize().Prime,
						Binary: Binary.ConvertXToByte(number)})
				subresult = field.Mul(subresult, negNumber)
			}
			result = field.Add(result, subresult)
		}
	}
	return result
}

//Calculate all the subset permutations for a given array except the list itself
//Sorted from smallest to biggest lengths
//Ex: array 3, 4, 5 returns
//[[3], [4], [5], [3, 4], [3, 5], [4, 5]]
func computePermutations(keys []int) [][]int {
	var binarySize = int(math.Pow(2, float64(len(keys))) - 1)
	//https://stackoverflow.com/questions/7150035/calculating-bits-required-to-store-decimal-number#:~:text=12%20Answers&text=Well%2C%20you%20just%20have%20to,%3D%201024%20(10%20bits).
	//Calculate the number of bits needed
	need := int(math.Ceil(math.Log10(float64(binarySize)) / math.Log10(2)))
	permutationsBinary := make([][]int, binarySize - 1)
	permutationsInts := make([][]int, binarySize - 1)
	//make [][]int does not make the inner []int, so lets create that
	for i := 0; i < binarySize - 1; i++ {
		permutationsInts[i] = make([]int, len(keys))
	}
	//Calculate the binary representation for the ints so we can get our subset permutations
	for i := 1; i < binarySize; i++ {
		permutationsBinary[i - 1] = intToBinaryArray(i, need)
	}
	//For all the 1's given the index take that element from keys index into the array
	//Ex keys [3, 4, 5] and perm [0, 1, 0] -> [0, 4, 0]
	for i := 0; i < len(permutationsBinary); i++ {
		for j := 0; j < len(permutationsBinary[0]); j++ {
			if permutationsBinary[i][j] == 1 {
				permutationsInts[i][j] = keys[j]
			}
		}
	}
	//Remove the zeroes in the list
	for i := 0; i < len(permutationsInts); i++ {
		for j := len(keys) - 1; j >= 0; j-- {
			if permutationsInts[i][j] == 0 {
				permutationsInts[i] = removeElementI(permutationsInts[i], j)
			}
		}
	}

	//https://stackoverflow.com/questions/42629541/go-lang-sort-a-2d-array
	//Sort the array from smallest to biggest lengths
	for k := 0; k < len(keys); k++ {
		sort.SliceStable(permutationsInts, func(i, j int) bool {
			return len(permutationsInts[i]) < len(permutationsInts[j])
		})
	}
	return permutationsInts
}

//Removes an element from the list
func removeElement(a []int, elem int) []int {
	index := 0
	for i := 0; i < len(a); i++ {
		if a[i] == elem {
			index = i
		}
	}
	return removeElementI(a, index)
}

//Removes and element on index i from the list
func removeElementI(a []int, i int) []int {
	c := make([]int, len(a))
	copy(c, a)
	// Remove the element at index i from a.
	c[i] = c[len(c)-1] // Copy last element to index i.
	c = c[:len(c)-1]   // Truncate slice.
	return c
}

//https://stackoverflow.com/questions/8151435/integer-to-binary-array/8151674
//Converts an int into a reversed binary representation.
//That it is reversed does not matter here
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
