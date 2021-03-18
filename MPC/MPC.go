package main

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	"MPC/Secret-Sharing/Shamir"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
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
var circuit Circuit.Circuit

func main() {
	if os.Args[1] == "-p" {
		finiteField = Prime.Prime{}
		finiteField.InitSeed()
		bundleType = primebundle.PrimeBundle{}
		loadCircuit("Circuit.json")
	}else {
		finiteField = Binary.Binary{}
		//TODO Add binary bundle
	}
	partySize, _ = strconv.Atoi(os.Args[2])

	secret, _ = strconv.Atoi(os.Args[3])

	//TODO fjern hardcoding
	//secretSharing = Shamir.Shamir{}
	secretSharing = Shamir.Shamir{}

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

	result := secretSharing.TheOneRing(circuit, secret)

	fmt.Println("Final result:", result)

}

func createField(fieldSize int) {
	finiteField.SetSize(fieldSize)
	secretSharing.SetField(finiteField)
}

func loadCircuit(file string) {
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	//var circuit Circuit.Circuit
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &circuit)

}