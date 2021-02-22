package main

import (
	bundle "MPC/Bundle"
	Prime_bundle "MPC/Bundle/Prime-bundle"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	Secret_Sharing "MPC/Secret-Sharing"
	Simple_Sharing "MPC/Secret-Sharing/Simple-Sharing"
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
		case Prime_bundle.PrimeBundle:
			if match.Type == "Prime" {
				createFieldAndShares(match.Prime)
				distributeShares()
			} else if match.Type == "Share" {
				receivedShares = append(receivedShares, match.Shares...)
				fmt.Println(receivedShares)
			} else if match.Type == "Result" {
				receivedResults = append(receivedResults, match.Result)

			} else {
				panic("Given type is unknown")
			}
	}
}

var finite_field finite.Finite
var bundle_type bundle.Bundle
var secret_sharing Secret_Sharing.Secret_Sharing
var party_size int
var secret int
var shares []int
var receivedShares []int
var receivedResults []int

func main() {
	if os.Args[1] == "-p" {
		finite_field = Prime.Prime{}
		finite_field.InitSeed()
		bundle_type = Prime_bundle.PrimeBundle{}
	}else {
		finite_field = Binary.Binary{}
		//TODO Add binary bundle
	}
	party_size, _ = strconv.Atoi(os.Args[2])

	secret, _ = strconv.Atoi(os.Args[3])

	//TODO fjern hardcoding
	secret_sharing = Simple_Sharing.Simple_Sharing{}

	receiver := Receiver{}
	network.RegisterReceiver(receiver)
	isFirst := network.Init(party_size)
	if isFirst {
		finiteSize := finite_field.GenerateField()
		switch bundle_type.(type) {
		case Prime_bundle.PrimeBundle:
			bundle_type = Prime_bundle.PrimeBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Prime",
				Prime: finiteSize,
			}
		default:
			fmt.Println(":(")
		}
		for {
			if network.IsReady() {
				for i := 0; i < party_size - 1; i++ {
					network.Send(bundle_type, i)
				}
				break
			}
		}
		createFieldAndShares(finiteSize)
		distributeShares()
	}

	for{
		if party_size * (party_size - 1) == len(receivedShares) {
			//Udregn function
			//TODO fjern hardcoding
			funcResult := secret_sharing.ComputeFunction(receivedShares)
			distributeResult(funcResult)
			break
		}
	}
	for {
		if party_size == len(receivedResults) {
			result := secret_sharing.ComputeResult(receivedResults) / (party_size - 1)
			fmt.Println("Got the following result: ", result)
			break
		}
	}
}

func createFieldAndShares(field_size int) {
	finite_field.SetSize(field_size)
	secret_sharing.SetField(finite_field)
	shares = secret_sharing.ComputeShares(party_size, secret)
	fmt.Println("My shares are:", shares)
}

func distributeShares() {
	for party := 1; party < party_size; party++ {
		shareCopy := make([]int, len(shares))
		copy(shareCopy, shares)
		share_slice := shareCopy[:party]
		share_slice2 := shareCopy[party + 1:]
		 share_bundle := Prime_bundle.PrimeBundle{
			ID: uuid.Must(uuid.NewRandom()).String(),
			Type: "Share",
			Shares: append(share_slice, share_slice2...),
			//append(shares[:party], shares[party + 1:]...),
		}

		network.Send(share_bundle, party - 1)
	}
	receivedShares = append(receivedShares, shares[1:]...)
}

func distributeResult(result int) {
	for party := 1; party < party_size; party++ {
		share_bundle := Prime_bundle.PrimeBundle{
			ID: uuid.Must(uuid.NewRandom()).String(),
			Type: "Result",
			Result: result,
		}
		network.Send(share_bundle, party - 1)
	}
	receivedResults = append(receivedResults, result)
}