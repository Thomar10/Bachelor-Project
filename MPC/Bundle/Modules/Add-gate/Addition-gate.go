package Add_gate

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	secretsharing "MPC/Secret-Sharing"
	"fmt"
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
	case primebundle.PrimeBundle:
		if match.Type == "Share" {
			receivedShares[match.From] = match.Shares
			//receivedShares = append(receivedShares, match.Shares...)
			//fmt.Println(receivedShares)
		} else if match.Type == "Result" {
			if len(receivedResults) != partySize {
				receivedResults = append(receivedResults, match.Result)
			}
		} else {
			panic("Given type is unknown")
		}
	}
}

var secretSharing secretsharing.Secret_Sharing
var partySize int


var receivedShares =  make(map[int][]int)
var receivedResults []int

func Add(input1, input2 int, sSharing secretsharing.Secret_Sharing) int {

	return input1 + input2

}