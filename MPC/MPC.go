package main

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	simplesharing "MPC/Secret-Sharing/Simple-Sharing"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strconv"
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
		case primebundle.PrimeBundle:
			if match.Type == "Prime" {
				myPartyNumber = network.GetPartyNumber()
				createFieldAndShares(match.Prime)
				distributeShares()
			} else if match.Type == "Share" {
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

var finiteField finite.Finite
var bundleType bundle.Bundle
var secretSharing secretsharing.Secret_Sharing
var partySize int
var secret int
var shares []int
var receivedShares =  make(map[int][]int)
var receivedResults []int
var myPartyNumber int

func main() {
	if os.Args[1] == "-p" {
		finiteField = Prime.Prime{}
		finiteField.InitSeed()
		bundleType = primebundle.PrimeBundle{}
	}else {
		finiteField = Binary.Binary{}
		//TODO Add binary bundle
	}
	partySize, _ = strconv.Atoi(os.Args[2])

	secret, _ = strconv.Atoi(os.Args[3])

	//TODO fjern hardcoding
	secretSharing = simplesharing.Simple_Sharing{}

	receiver := Receiver{}
	network.RegisterReceiver(receiver)
	isFirst := network.Init(partySize)
	if isFirst {
		finiteSize := finiteField.GenerateField()
		switch bundleType.(type) {
		case primebundle.PrimeBundle:
			bundleType = primebundle.PrimeBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Prime",
				Prime: finiteSize,
			}
		default:
			fmt.Println(":(")
		}
		for {
			if network.IsReady() {
				myPartyNumber = network.GetPartyNumber()
				fmt.Println(myPartyNumber)
				for i := 1; i <= partySize; i++ {
					if myPartyNumber != i {
						network.Send(bundleType, i)
					}
				}
				break
			}
		}
		createFieldAndShares(finiteSize)
		distributeShares()
	}

	for{
		if partySize == len(receivedShares) {
			//Udregn function
			//TODO fjern hardcoding
			funcResult := secretSharing.ComputeFunction(receivedShares)
			distributeResult(funcResult)
			break
		}
	}
	for {
		if partySize == len(receivedResults) {
			result := secretSharing.ComputeResult(receivedResults)// / (partySize - 1)
			fmt.Println("Got the following result: ", result)
			fmt.Println("My peer list looks as following: ", network.Peers())
			break
		}
	}
}

func createFieldAndShares(fieldSize int) {
	finiteField.SetSize(fieldSize)
	secretSharing.SetField(finiteField)
	shares = secretSharing.ComputeShares(partySize, secret)
	fmt.Println("My shares are:", shares)
}

func distributeShares() {
	for party := 1; party <= partySize; party++ {
		shareCopy := make([]int, len(shares))
		copy(shareCopy, shares)
		shareSlice := shareCopy[:party - 1]
		shareSlice2 := shareCopy[party:]
		shareSlice = append(shareSlice, shareSlice2...)
		 shareBundle := primebundle.PrimeBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: shareSlice,
			From: myPartyNumber,
		}
		fmt.Println("Sending shares: ", shareSlice, "To party", party)
		if myPartyNumber == party {
			receivedShares[myPartyNumber] = shareSlice
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []int) {
	counter := 0
	for party := 1; party <= partySize; party++ {
		if myPartyNumber != party {
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