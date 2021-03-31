package Multiplication

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	Finite_fields "MPC/Finite-fields"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	"fmt"
	"github.com/google/uuid"
	"math/big"
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
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

var shares []Finite_fields.Number
var receivedShares =  make(map[int][]Finite_fields.Number)
var receivedResults []Finite_fields.Number

func Multiply(secret Finite_fields.Number, sSharing secretsharing.Secret_Sharing, pSize int) Finite_fields.Number {
	partySize = pSize
	secretSharing = sSharing

	receiver := Receiver{}

	network.RegisterReceiver(receiver)

	r := secret.Prime.Cmp(big.NewInt(-1))
	if r != 0 {
		shares = secretSharing.ComputeShares(partySize, secret)
		fmt.Println("My shares are:", shares)

		distributeShares()
	}

	for{
		if len(receivedShares) == 2 {
			//Udregn function
			//TODO fjern hardcoding
			funcResult := secretSharing.ComputeFunction(receivedShares, network.GetPartyNumber())
			return funcResult[0]
		}
	}
}

func distributeShares() {
	for party := 1; party <= partySize; party++ {
		shareCopy := make([]Finite_fields.Number, len(shares))
		copy(shareCopy, shares)
		shareSlice := shareCopy[:party - 1]
		shareSlice2 := shareCopy[party:]
		shareSlice = append(shareSlice, shareSlice2...)
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: shareSlice,
			From:   network.GetPartyNumber(),
		}
		//fmt.Println("Sending shares: ", shareSlice, "To party", party)
		if network.GetPartyNumber() == party {
			receivedShares[network.GetPartyNumber()] = shareSlice
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []Finite_fields.Number) {
	counter := 0
	for party := 1; party <= partySize; party++ {
		if network.GetPartyNumber() != party {
			shareBundle := numberbundle.NumberBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Result",
				Result: result[counter],
			}
			counter++
			network.Send(shareBundle, party)
		} else {
			receivedResults = append(receivedResults, result...)
		}
	}
}

