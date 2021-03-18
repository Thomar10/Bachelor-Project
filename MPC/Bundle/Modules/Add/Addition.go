package Add

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	"fmt"
	"github.com/google/uuid"
	"math/big"
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

var shares []*big.Int
var receivedShares =  make(map[int][]*big.Int)
var receivedResults []*big.Int

func Add(secret *big.Int, sSharing secretsharing.Secret_Sharing, pSize int) int {
	partySize = pSize
	secretSharing = sSharing

	receiver := Receiver{}

	network.RegisterReceiver(receiver)

	shares = secretSharing.ComputeShares(partySize, secret)
	fmt.Println("My shares are:", shares)

	distributeShares()

	for{
		if partySize == len(receivedShares) {
			//Udregn function
			//TODO fjern hardcoding
			funcResult := secretSharing.ComputeFunction(receivedShares, network.GetPartyNumber())
			distributeResult(funcResult)
			break
		}
	}

	for {
		if partySize == len(receivedResults) {
			result := secretSharing.ComputeResult(receivedResults) // / (partySize - 1)
			fmt.Println("Got the following result: ", result)
			fmt.Println("My peer list looks as follows: ", network.Peers())
			return result
		}
	}

}

func distributeShares() {
	for party := 1; party <= partySize; party++ {
		shareCopy := make([]*big.Int, len(shares))
		copy(shareCopy, shares)
		shareSlice := shareCopy[:party - 1]
		shareSlice2 := shareCopy[party:]
		shareSlice = append(shareSlice, shareSlice2...)
		shareBundle := primebundle.PrimeBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: shareSlice,
			From:   network.GetPartyNumber(),
		}
		fmt.Println("Sending shares: ", shareSlice, "To party", party)
		if network.GetPartyNumber() == party {
			receivedShares[network.GetPartyNumber()] = shareSlice
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []*big.Int) {
	counter := 0
	for party := 1; party <= partySize; party++ {
		if network.GetPartyNumber() != party {
			shareBundle := primebundle.PrimeBundle{
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
