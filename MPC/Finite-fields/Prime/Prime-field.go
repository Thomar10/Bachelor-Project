package Prime

import (
	finite "MPC/Finite-fields"
	crand "crypto/rand"
	"math/big"
	"reflect"
)

type Prime struct {

}

var primeNumber finite.Number


func (p Prime) CheckPolynomialIsConsistent(resultGate map[int]map[int]finite.Number, corrupts int, reconstructFunction func(map[int]finite.Number, int) []finite.Number) (bool, [][]finite.Number) {
	keys := reflect.ValueOf(resultGate).MapKeys()
	if len(keys) <= 0 {
		return true, [][]finite.Number{}
	}
	key := keys[0].Interface().(int)
	resultPolynomial := reconstructFunction(resultGate[key], corrupts)
	for i, v := range resultGate[key] {
		polyShare := p.CalcPoly(resultPolynomial, i)
		if !p.CompareEqNumbers(v, polyShare) {
			return false, [][]finite.Number{resultPolynomial}
		}
	}
	return true, [][]finite.Number{resultPolynomial}
}

func (p Prime) HaveEnoughForReconstruction(outputs, parties int, resultGate map[int]map[int]finite.Number) bool {
	if outputs > 0 {
		keys := reflect.ValueOf(resultGate).MapKeys()
		key := keys[0]
		if len(resultGate[(key.Interface()).(int)]) >= parties {
			return true
		}
		return false
	}
	return true
}

func (p Prime) ComputeFieldResult(outputSize int, polynomial [][]finite.Number) finite.Number {
	var result finite.Number
	if outputSize == 0 {
		//No outputs for this party - return 0
		result.Prime = big.NewInt(0)
		return result
	}else {
		//There is only one polynomial for primes (as there is only one output gate per party
		return p.CalcPoly(polynomial[0], 0)
	}
}

//Checks if a list is filled up with correct values
func (p Prime) FilledUp(numbers []finite.Number) bool {
	for _, number := range numbers {
		//r >= 0 if 0 is bigger or equal to number.Prime
		r := number.Prime.Cmp(big.NewInt(0))
		if r < 0 {
			return false
		}
	}
	return true
}

//Converts a constant from a gate into finite number
func (p Prime) GetConstant(constant int) finite.Number {
	return finite.Number{Prime: big.NewInt(int64(constant))}
}

//Computes Shamir secret shares
func (p Prime) ComputeShares(parties int, secret finite.Number, t int) []finite.Number {
	//Polynomial: 3 + 4x + 2x^2
	//Representation of that poly: [3, 4, 2]
	var polynomial = make([]*big.Int, t + 1)

	polynomial[0] = secret.Prime
	for i := 1; i < t + 1; i++ {
		polynomial[i], _ = crand.Int(crand.Reader, primeNumber.Prime)
	}

	var shares = make([]*big.Int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}

	result := make([]finite.Number, len(shares))
	for i := 0; i < len(result); i++ {
		result[i] = finite.Number{Prime: shares[i]}
	}
	return result
}

func (p Prime) InitSeed() {
}

//Sets the size of the prime field (the prime number)
func (p Prime) SetSize(f finite.Number) {
	primeNumber = f
}

//Returns the size of the prime field (the prime number)
func (p Prime) GetSize() finite.Number {
	return primeNumber
}

//Generate a random prime for the finite field
func (p Prime) GenerateField() finite.Number {
	//Could be higher number bits, would just give a higher prime number
	bigPrime, err := crand.Prime(crand.Reader, 32)
	if err != nil {
		panic("Unable to compute prime")
	}
	return finite.Number{Prime: bigPrime}
}

//Adds two numbers in a finite field
func (p Prime) Add(n1, n2 finite.Number) finite.Number {
	x := new(big.Int).Add(n1.Prime, n2.Prime)
	x.Mod(x, primeNumber.Prime)
	return finite.Number{Prime: x}
}

//Multiply two numbers in a finite field
func (p Prime) Mul(n1, n2 finite.Number) finite.Number {
	x := new(big.Int).Mul(n1.Prime, n2.Prime)
	x.Mod(x, primeNumber.Prime)
	return finite.Number{Prime: x}
}

//Compare if two numbers are equal
func (p Prime) CompareEqNumbers(share, polyShare finite.Number) bool {
	r := share.Prime.Cmp(polyShare.Prime)
	return r == 0
}

//Converter function to help call calculatePolynomial with finite number
func (p Prime) CalcPoly(poly []finite.Number, x int) finite.Number {
	polyBig := make([]*big.Int, len(poly))
	for i, _ := range poly {
		polyBig[i] = poly[i].Prime
	}
	result := calculatePolynomial(polyBig, x)
	return finite.Number{Prime: result}
}

//Calculates the polynomial f(x) for a given x and polynomial
func calculatePolynomial(polynomial []*big.Int, x int) *big.Int {
	var result = big.NewInt(0)
	for i := 0; i < len(polynomial); i++ {
		iterres := new(big.Int).Exp(big.NewInt(int64(x)), big.NewInt(int64(i)), nil)
		iterres.Mul(iterres, polynomial[i])
		result.Add(result, iterres)
	}
	return result.Mod(result, primeNumber.Prime)
}

//Finds the inverse of the number a
func (p Prime) FindInverse(a, prime finite.Number) finite.Number{
	r := a.Prime.Cmp(big.NewInt(0))
	if r < 0 {
		//a = prime + a
		a.Prime.Add(prime.Prime, a.Prime)
	}
	result := big.NewInt(1)
	result.Exp(a.Prime, new(big.Int).Sub(prime.Prime, big.NewInt(2)), prime.Prime)
	return finite.Number{Prime: result}
}