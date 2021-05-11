package Binary

import (
	finite "MPC/Finite-fields"
	"math/rand"
	"reflect"
	"sort"
	"time"
)

type Binary struct {
}

var field finite.Number

func (b Binary) ConstructFieldSecret(secret finite.Number, doesIHaveAnInput bool, partySize, corrupts, partyNumber int) ([][]finite.Number, []int) {
	resultSecrets := make([][]finite.Number, len(secret.Binary))
	inputGates := make([]int, len(secret.Binary))
	for i, sec := range secret.Binary {
		binarySec := make([]int, 8)
		binarySec[7] = sec
		resultSecrets[i] = b.ComputeShares(partySize, finite.Number{Binary: binarySec}, corrupts)
		inputGates[i] = partyNumber * len(secret.Binary) + i - len(secret.Binary) + 1
	}
	return resultSecrets, inputGates
}

func (b Binary) CreateRandomNumber() finite.Number {
	return finite.Number{Binary: CreateRandomByte()}
}

func (b Binary) CheckPolynomialIsConsistent(resultGate map[int]map[int]finite.Number, corrupts int, reconstructFunction func(map[int]finite.Number, int) []finite.Number) (bool, [][]finite.Number) {
	keys := reflect.ValueOf(resultGate).MapKeys()
	if len(keys) <= 0 {
		return true, [][]finite.Number{}
	}
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}
	sort.Ints(keysArray)
	resultPolynomials := make([][]finite.Number, len(keysArray))
	for j, k := range keysArray {
		resultPolynomial := reconstructFunction(resultGate[k], corrupts)
		resultPolynomials[j] = resultPolynomial
		for i, v := range resultGate[k] {
			polyShare := b.CalcPoly(resultPolynomial, i)
			if !b.CompareEqNumbers(v, polyShare) {
				return false, resultPolynomials
			}
		}
	}
	return true, resultPolynomials
}

func (b Binary) HaveEnoughForReconstruction(outputs, parties int, resultGate map[int]map[int]finite.Number) bool {
	if outputs > 0 {
		keys := reflect.ValueOf(resultGate).MapKeys()
		var keysArray []int
		for _, k := range keys {
			keysArray = append(keysArray, (k.Interface()).(int))
		}
		sort.Ints(keysArray)
		for _, k := range keysArray {
			if len(resultGate[k]) < parties  {
				return false
			}
		}
	}
	return true
}

func (b Binary) ComputeFieldResult(outputSize int, polynomials [][]finite.Number) finite.Number {
	trueResult := make([]int, outputSize)
	if outputSize > 0 {
		for i, poly := range polynomials {
			resultBit := b.CalcPoly(poly, 0).Binary[7]
			trueResult[i] = resultBit
		}
		return finite.Number{Binary: trueResult}
	} else {
		return finite.Number{Binary: []int{0}}

	}
}

//Checks if a list is filled up with correct values
func (b Binary) FilledUp(numbers []finite.Number) bool {
	for _, number := range numbers {
		if number.Binary[0] == -1 {
			return false
		}
	}
	return true
}

//Converts a constant from a gate into finite number
func (b Binary) GetConstant(constant int) finite.Number {
	constantByte := ConvertXToByte(constant)
	return finite.Number{Binary: constantByte}
}

//Initialize the random seed to be used
func (b Binary) InitSeed() {
	field = finite.Number{Binary: make([]int, 8)}
	rand.Seed(time.Now().UnixNano())
}

//Compare if two numbers are equal
func (b Binary) CompareEqNumbers(share, polyShare finite.Number) bool {
	for i, s := range share.Binary {
		if s != polyShare.Binary[i] {
			return false
		}
	}
	return true
}

//Converter function to help call calculatePolynomial with finite number
func (b Binary) CalcPoly(poly []finite.Number, x int) finite.Number {
	polyBinary := make([][]int, len(poly))
	for i, _ := range poly {
		polyBinary[i] = poly[i].Binary
	}
	result := calculatePolynomial(polyBinary, x)
	return finite.Number{Binary: result}
}

//Sets the size of the field
func (b Binary) SetSize(f finite.Number) {
	field = f
}

//Computes Shamir Secret Shares
func (b Binary) ComputeShares(parties int, secret finite.Number, t int) []finite.Number {
	//[0,0,..,1, 0] + [0,0,..,1, 0]x + [0,0,..,1, 0]x^2 (x -> [0,0,..,1, 0])
	//[[0,0,..,1, 0], [0,0,..,1, 0], [0,0,..,1, 0]] -> shares er i binary
	var polynomial = make([][]int, t + 1)
	polynomial[0] = secret.Binary
	for i := 1; i < len(polynomial); i++ {
		polynomial[i] = CreateRandomByte()
	}

	shares := make([][]int, parties)
	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}
	result := make([]finite.Number, len(shares))
	for i := 1; i <= len(result); i++ {
		result[i - 1] = finite.Number{Binary: shares[i - 1]}
	}
	return result

}

//Converts a number into its binary representation
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

//Reverses an array
func reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

//Converts a number into byte array representation
func ConvertXToByte(x int) []int {
	result := reverse(intToBinaryArray(x, 8))
	return result
}

//Calculates the polynomial f(x) for a given x and polynomial
func calculatePolynomial(polynomial [][]int, x int) []int {
	var result = make([]int, 8)
	var xByte = ConvertXToByte(x)
	for i := 0; i < len(polynomial); i++ {
		result = bitAdd(bitMult(polynomial[i], bitExponent(xByte, i)), result)

	}
	return result
}

//Calculate the bit exponent [0,...,0]^x
func bitExponent(byte []int, x int) []int {
	result := []int{0, 0, 0, 0, 0, 0, 0, 1}
	for i := 1; i <= x; i++ {
		result = bitMult(result, byte)
	}
	return result
}

//Returns a random byte array
func CreateRandomByte() []int {
	result := make([]int, 8)
	for  i := 1; i < len(result); i++ {
		result[i] = rand.Intn(2)
	}
	return result
}

//Generate the finite field
func (b Binary) GenerateField() finite.Number {
	return finite.Number{Binary: make([]int, 8)}
}

//Sets the size of the finite field
func (b Binary) GetSize() finite.Number {
	return field
}

//Multiply two byte arrays in a finite field
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

//Adds two byte arrays in a finite field
func bitAdd(b1 []int, b2 []int) []int {

	bitRes := make([]int, len(b1))

	for i := 0; i < len(b1); i++ {
		bitRes[i] = b1[i] ^ b2[i]
	}
	return bitRes
}

//Adds two numbers in a finite field
func (b Binary) Add(n1, n2 finite.Number) finite.Number {
	n1.Binary = bitAdd(n1.Binary, n2.Binary)
	return n1
}

//Multiply two numbers in a finite field
func (b Binary) Mul(n1, n2 finite.Number) finite.Number {
	n1.Binary = bitMult(n1.Binary, n2.Binary)
	return n1
}

//Creates a irreducible polynomial
func makeIrreducible() []int{
	irreducible := make([]int, 15)
	irreducible[8] = 1
	irreducible[7] = 1
	irreducible[5] = 1
	irreducible[4] = 1
	irreducible[0] = 1
	return irreducible
}

//Shifts a bit to the right
func bitShiftRight(array []int) []int {
	length := len(array)
	result := make([]int, length)
	result = append(result[:1], array[: length - 1]...)
	return result
}

//TODO fjern eventuelt eftersom det blot er addition
//Subtract two byte arrays
func sliceSubtraction(a1 []int, a2 []int) []int {
	for i := 0; i < len(a1); i++ {
		a1[i] = a1[i] ^ a2[i]
	}

	return a1
}

//Finds the inverse of a byte
func findInverseBit(a []int) []int {
	result := a
	for i := 1; i < 255; i++ {
		result = bitMult(bitExponent(a, 2), result)
	}
	return result
}

//Finds the inverse of a number
func (b Binary) FindInverse(a, p finite.Number) finite.Number {
	result := findInverseBit(a.Binary)
	return finite.Number{Binary: result}
}