package Prime

import (
	finite "MPC/Finite-fields"
	crand "crypto/rand"
	"math/big"
)

type Prime struct {

}

var primeNumber finite.Number

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
func (p Prime) ComputeShares(parties int, secret finite.Number) []finite.Number {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
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