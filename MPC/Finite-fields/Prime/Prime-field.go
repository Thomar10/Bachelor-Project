package Prime

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"time"
)

type Prime struct {

}

func (p Prime) InitSeed() {
	//TODO find bedre plads senere eventuelt til rand seed
	rand.Seed(time.Now().UnixNano())
}

func (p Prime) SetSize(f int) {
	prime = f
}
func (p Prime) GetSize() int {
	return prime
}

var prime int

func (p Prime) GenerateField() int {
	bigPrime, err := crand.Prime(crand.Reader, 32) //32 to avoid errors when converting to int
	if err != nil {
		fmt.Println("Unable to compute prime", err.Error())
		return 0
	}
	return int(bigPrime.Int64())
}


func (p Prime) ComputeShares(parties, secret int) []int {
	var shares []int
	lastShare := secret
	//Create the n - 1 random shares
	for s := 1; s < parties; s++ {
		shares = append(shares, randomNumberInZ(prime))
	}
	//Create the nth share
	for _, share := range shares {
		lastShare -= share
	}
	//Remove negative number
	if lastShare < 0 {
		//fmt.Println("prime + lastShare: ", prime + lastShare)
		lastShare = prime + lastShare % prime
	}
	shares = append(shares, lastShare % prime)

	return shares
}

func randomNumberInZ(prime int) int {
	return rand.Intn(prime)
}



//TODO Complete add multiply etc.
func Add(a, b int) int {
	return 0
}
