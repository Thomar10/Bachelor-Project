package Simple_Sharing

import (
	"fmt"
	"math/rand"
	crand "crypto/rand"
)
var prime int

func ComputeShares(parties, secret int) []int {
	var shares []int
	//Create the n - 1 random shares
	for s := 1; s < parties; s++ {
		shares = append(shares, randomNumberInZ(prime - 1))
	}
	//Create the nth share
	for share := range shares {
		secret -= share
	}
	shares = append(shares, secret % prime)

	return shares
}

func randomNumberInZ(prime int) int {
	return rand.Intn(prime)
}

func SetPrime(p int) {
	prime = p
}

func GeneratePrime() int {
	bigPrime, err := crand.Prime(crand.Reader, 32) //32 to avoid errors when converting to int
	if err != nil {
		fmt.Println("Unable to compute prime", err.Error())
		return 0
	}
	return int(bigPrime.Int64())
}

