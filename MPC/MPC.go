package main

import (
	bundle "MPC/Bundle"
	"MPC/Bundle/Modules"
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
				createField(match.Prime)
				sizeSet = true
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
var sizeSet bool
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
		createField(finiteSize)
		sizeSet = true
	}

	for {
		if sizeSet {
			break
		}
	}

	result := Modules.Add(secret, secretSharing, partySize)
	fmt.Println("Done adding, got result:", result)
}

func createField(fieldSize int) {
	finiteField.SetSize(fieldSize)
	secretSharing.SetField(finiteField)
}