package Prime

import (
	crand "crypto/rand"
	"fmt"
)

type Prime struct {

}
func (p Prime) GenerateField() int {
	bigPrime, err := crand.Prime(crand.Reader, 32) //32 to avoid errors when converting to int
	if err != nil {
		fmt.Println("Unable to compute prime", err.Error())
		return 0
	}
	return int(bigPrime.Int64())
}

//TODO Complete add multiply etc.
func Add(a, b int) int {
	return 0
}
