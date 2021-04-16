package Shamir

import (
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
)


/*func ReconstructPolynomial(shares map[int]*big.Int) int {
	keys := reflect.ValueOf(shares).MapKeys()
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}
	sort.Ints(keysArray)
	//deltas := make([][]int, len(keysArray))
	secret := big.NewInt(0)
	for _, key := range keysArray {
		//secret += shares[key] * computeDelta(key, keysArray)[0]
		iterRes := new(big.Int).Mul(shares[key], computeDeltaPolynomial(key, keysArray)[0])
		secret.Add(secret, iterRes)
		//deltas[i] = computeDelta(key, keysArray)
	}

	return int(secret.Mod(secret, field.GetSize()).Int64())//secret % field.GetSize()

}*/

func Reconstruct(shares map[int]finite.Number) finite.Number {
	keys := reflect.ValueOf(shares).MapKeys()
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}
	sort.Ints(keysArray)
	//deltas := make([][]int, len(keysArray))
	var secret = finite.Number{Prime: big.NewInt(0), Binary: []int{0, 0, 0, 0, 0, 0, 0, 0}}
	for _, key := range keysArray {
		//secret += shares[key] * computeDelta(key, keysArray)[0]
		delta := computeDelta(key, keysArray)
		share := shares[key]
		var interRes = field.Mul(share, delta)
		secret = field.Add(interRes, secret)

		//iterRes := new(big.Int).Mul(shares[key], computeDelta(key, keysArray))
		//secret.Add(secret, iterRes)

		//deltas[i] = computeDelta(key, keysArray)
	}
	//int(secret.Mod(secret, field.GetSize()).Int64())//secret % field.GetSize()
	return secret

}

/*func computeDeltaPolynomial(key int, keys []int) []*big.Int {
	var talker = big.NewInt(1)
	for _, j := range keys {
		if j == key {
			continue
		}
		//talker *= key - j
		talker.Mul(talker, new(big.Int).Sub(big.NewInt(int64(key)), big.NewInt(int64(j))))
	}
	//talker = talker % field.GetSize()
	talker.Mod(talker, field.GetSize())
	var inverseTalker = field.FindInverse(talker, field.GetSize())
	keyIndex := 0
	for i := 0; i < len(keys); i++ {
		if keys[i] == key {
			keyIndex = i
		}
	}
	keysWithoutkey := removeElementI(keys, keyIndex)
	permutations := computePermutations(keysWithoutkey)
	var polynomial = make([]*big.Int, len(keys))
	//keys = 3, 4, 5, 6 -> delta3
	//(x-4)(x-5)(x-6)
	// -120 + 74x -15x^2+x^3
	//[-120, 74, -15, 1]
	for i := 1; i < len(polynomial) - 1; i++ {
		polynomial[i] = multipleAllWithSize(len(keysWithoutkey) - i, permutations)
	}
	polynomial[0] = big.NewInt(1)
	polynomial[len(polynomial) - 1] = big.NewInt(1)
	for _, number := range keysWithoutkey {
		//polynomial[0] = polynomial[0] * -number
		polynomial[0].Mul(polynomial[0], big.NewInt(int64(-number)))
	}

	for i, number := range polynomial {
		//number = number % field.GetSize()
		number.Mod(number, field.GetSize())
		r := number.Cmp(big.NewInt(0))
		if r < 0 {
			//number = field.GetSize() + number
			number.Add(field.GetSize(), number)
		}
		// number * inverseTalker % prime
		x := big.NewInt(1).Mul(number, inverseTalker)
		polynomial[i] = x.Mod(x, field.GetSize())

	}

	return polynomial
}
*/


func computeDelta(key int, keys []int) finite.Number {
	var talker = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}} //rip
	for _, j := range keys {
		if j == key {
			continue
		}
		//talker *= key - j
		//talker.Mul(talker, new(big.Int).Sub(big.NewInt(int64(key)), big.NewInt(int64(j))))
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
	//talker = talker % field.GetSize()
	//talker.Mod(talker, field.GetSize())
	var inverseTalker = field.FindInverse(talker, field.GetSize())
	keyIndex := 0
	for i := 0; i < len(keys); i++ {
		if keys[i] == key {
			keyIndex = i
		}
	}
	keysWithoutkey := removeElementI(keys, keyIndex)

	var polynomial = finite.Number{Prime: big.NewInt(1), Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}}
	for _, number := range keysWithoutkey {
		//polynomial[0] = polynomial[0] * -number
		//polynomial.Mul(polynomial, big.NewInt(int64(-number)))
		var numberNumber = field.Add(
			finite.Number{
				Prime: big.NewInt(int64(-number)),
				Binary: Binary.ConvertXToByte(0)},
			finite.Number{
				Prime: field.GetSize().Prime,
				Binary: Binary.ConvertXToByte(number)})
		//numberNumber.Prime = big.NewInt(int64(-number))

		polynomial = field.Mul(polynomial, numberNumber)
	}
	//polynomial.Mod(polynomial, field.GetSize())
	/*r := polynomial.Cmp(big.NewInt(0))
	if r < 0 {
		//number = field.GetSize() + number
		polynomial.Add(field.GetSize(), polynomial)
	}*/
	//polynomial.Mul(polynomial, inverseTalker)
	polynomial = field.Mul(polynomial, inverseTalker)
	//polynomial.Mod(polynomial, field.GetSize())
	return polynomial
}
/*
func multipleAllWithSize(k int, permutations [][]int) *big.Int {
	result := big.NewInt(0)
	for _, perm := range permutations {
		if len(perm) == k {
			subresult := big.NewInt(1)
			for _, number := range perm {
				//subresult = (subresult * -number) % field.GetSize()
				subresult.Mul(subresult, big.NewInt(int64(-number)))
				subresult.Mod(subresult, field.GetSize())


			}
			//result += subresult % field.GetSize()
			result.Add(result, subresult.Mod(subresult, field.GetSize()))
		}
	}

	return result
}*/


func computePermutations(keys []int) [][]int {
	var binarySize = int(math.Pow(2, float64(len(keys))) - 1)
	fmt.Print("Binary size", binarySize)
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
		permutationsBinary[i - 1] = intToBinaryArray(i, need)
	}
	for i := 0; i < len(permutationsBinary); i++ {
		for j := 0; j < len(permutationsBinary[0]); j++ {
			if permutationsBinary[i][j] == 1 {
				permutationsInts[i][j] = keys[j]
			}
		}
	}

	for i := 0; i < len(permutationsInts); i++ {
		for j := len(keys) - 1; j >= 0; j-- {
			if permutationsInts[i][j] == 0 {
				permutationsInts[i] = removeElementI(permutationsInts[i], j)
			}
		}
	}



	//https://stackoverflow.com/questions/42629541/go-lang-sort-a-2d-array
	for k := 0; k < len(keys); k++ {
		sort.SliceStable(permutationsInts, func(i, j int) bool {
			return len(permutationsInts[i]) < len(permutationsInts[j])
		})
	}
	return permutationsInts
}

func removeElementI(a []int, i int) []int {
	c := make([]int, len(a))
	copy(c, a)
	// Remove the element at index i from a.
	c[i] = c[len(c)-1] // Copy last element to index i.
	c = c[:len(c)-1]   // Truncate slice.
	return c
}

//https://stackoverflow.com/questions/8151435/integer-to-binary-array/8151674
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



//Itoh–Tsujii inversion algorithm
/*
func findItohInverse(a int, prime int) int {
	var r = (prime - 1) / (prime - 1)
	var ar1 = int(math.Pow(float64(a), float64(r - 1))) % prime
	var ar = ar1 * a
	var arp1 = (1 / ar) % prime
	var a1 = arp1 * ar1
	return a1
}
 */
