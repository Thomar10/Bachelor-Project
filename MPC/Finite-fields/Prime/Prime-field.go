package Prime

import (
	Finite_fields "MPC/Finite-fields"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

type Prime struct {

}
var prime Finite_fields.Number

func (p Prime) ComputeShares(parties int, secret Finite_fields.Number) []Finite_fields.Number {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
	//3 + 4x + 2x^2
	//[3, 4, 2]
	var polynomial = make([]*big.Int, t + 1)

	polynomial[0] = secret.Prime
	for i := 1; i < t + 1; i++ {
		//TODO Måske gøre så vi kan få error ud og tjekke på (fuck go)
		polynomial[i], _ = crand.Int(crand.Reader, prime.Prime)
	}

	var shares = make([]*big.Int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}
	result := make([]Finite_fields.Number, len(shares))
	for i := 0; i < len(result); i++ {
		result[i] = Finite_fields.Number{Prime: shares[i]}
	}
	return result
}

func (p Prime) InitSeed() {
	//TODO find bedre plads senere eventuelt til rand seed
	rand.Seed(time.Now().UnixNano())
}

func (p Prime) SetSize(f Finite_fields.Number) {
	prime = f
}
func (p Prime) GetSize() Finite_fields.Number {
	return prime
}


func (p Prime) GenerateField() Finite_fields.Number {
	bigPrime, err := crand.Prime(crand.Reader, 32) //32 to because it should be big enough
	if err != nil {
		fmt.Println("Unable to compute prime", err.Error())
		return Finite_fields.Number{Prime: big.NewInt(0)}
	}
	return Finite_fields.Number{Prime: bigPrime}
}

func (p Prime) Add(n1, n2 Finite_fields.Number) Finite_fields.Number {
	n1.Prime.Add(n1.Prime, n2.Prime)
	n1.Prime.Mod(n1.Prime, prime.Prime)
	return n1
}

func (p Prime) Mul(n1, n2 Finite_fields.Number) Finite_fields.Number {
	n1.Prime.Mul(n1.Prime, n2.Prime)
	n1.Prime.Mod(n1.Prime, prime.Prime)
	return n1
}

func calculatePolynomial(polynomial []*big.Int, x int) *big.Int {
	var result = big.NewInt(0)

	for i := 0; i < len(polynomial); i++ {
		//result += polynomial[i] * int(math.Pow(float64(x), float64(i)))
		iterres := new(big.Int).Exp(big.NewInt(int64(x)), big.NewInt(int64(i)), nil)
		iterres.Mul(iterres, polynomial[i])
		result.Add(result, iterres)
	}

	return result.Mod(result, prime.Prime)//result % field.GetSize()
}

func (p Prime) FindInverse(a, prime Finite_fields.Number) Finite_fields.Number{
	r := a.Prime.Cmp(big.NewInt(0))
	if r < 0 {
		//a = prime + a
		a.Prime.Add(prime.Prime, a.Prime)
	}
	result := big.NewInt(1)
	result.Exp(a.Prime, new(big.Int).Sub(prime.Prime, big.NewInt(2)), prime.Prime)
	return Finite_fields.Number{Prime: result}
	//return int(math.Pow(float64(a), float64(prime - 2))) % prime
}