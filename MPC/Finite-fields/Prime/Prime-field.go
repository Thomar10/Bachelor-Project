package Prime

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

type Prime struct {

}

func (p Prime) InitSeed() {
	//TODO find bedre plads senere eventuelt til rand seed
	rand.Seed(time.Now().UnixNano())
}

func (p Prime) SetSize(f *big.Int) {
	prime = f
}
func (p Prime) GetSize() *big.Int {
	return prime
}

var prime *big.Int

func (p Prime) GenerateField() *big.Int {
	bigPrime, err := crand.Prime(crand.Reader, 32) //32 to because it should be big enough
	if err != nil {
		fmt.Println("Unable to compute prime", err.Error())
		return big.NewInt(0)
	}
	return bigPrime
}
