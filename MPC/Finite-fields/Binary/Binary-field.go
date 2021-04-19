package Binary

import (
	finite "MPC/Finite-fields"
	"math/rand"
	"sync"
	"time"
)

type Binary struct {

}

var convMutex = &sync.Mutex{}


func (p Binary) FilledUp(numbers []finite.Number) bool {
	for _, number := range numbers {
		if number.Binary[0] == -1 {
			return false
		}
	}
	return true
}

func (b Binary) GetConstant(constant int) finite.Number {
	constantByte := ConvertXToByte(constant)
	return finite.Number{Binary: constantByte}
}

var field finite.Number

func (b Binary) InitSeed() {
	field = finite.Number{Binary: make([]int, 8)}
	rand.Seed(time.Now().UnixNano())
}

func (b Binary) SetSize(f finite.Number) {
	field = f
}

func (b Binary) ComputeShares(parties int, secret finite.Number) []finite.Number {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
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

func intToBinaryArray(number, arraySize int) []int {
	convMutex.Lock()
	defer convMutex.Unlock()
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

func ConvertXToByte(x int) []int {
	result := reverse(intToBinaryArray(x, 8))
	return result
}

func calculatePolynomial(polynomial [][]int, x int) []int {
	//fmt.Println("polynomial", polynomial)
	//fmt.Println("x", x)
	var result = make([]int, 8)
	var xByte = ConvertXToByte(x)
	for i := 0; i < len(polynomial); i++ {
		result = bitAdd(bitMult(polynomial[i], bitExponent(xByte, i)), result)

	}
	//fmt.Println(result)
	return result
}

func BitExponent(byte []int, x int) []int {
	result := []int{0, 0, 0, 0, 0, 0, 0, 1}
	for i := 1; i <= x; i++ {
		result = bitMult(result, byte)
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

func CreateRandomByte() []int {
	result := make([]int, 8)
	for  i := 1; i < len(result); i++ {
		result[i] = rand.Intn(2)
	}
	return result
}


func (b Binary) GenerateField() finite.Number {
	return finite.Number{Binary: make([]int, 8)}
}

func (b Binary) GetSize() finite.Number {
	return field
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

func bitAdd(b1 []int, b2 []int) []int {

	bitRes := make([]int, len(b1))

	for i := 0; i < len(b1); i++ {
		bitRes[i] = b1[i] ^ b2[i]
	}
	return bitRes
}

func (b Binary) Add(n1, n2 finite.Number) finite.Number {
	n1.Binary = bitAdd(n1.Binary, n2.Binary)
	return n1
}


func (b Binary) Mul(n1, n2 finite.Number) finite.Number {
	n1.Binary = bitMult(n1.Binary, n2.Binary)
	return n1
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
func findInverseBit(a []int) []int {
	result := a
	for i := 1; i < 255; i++ {
		result = bitMult(bitExponent(a, 2), result)
	}
	return result
}

func (b Binary) FindInverse(a, p finite.Number) finite.Number {
	result := findInverseBit(a.Binary)
	return finite.Number{Binary: result}
}