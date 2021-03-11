package main

import (
	bundle "MPC/Bundle"
	"MPC/Bundle/Modules/Add"
	"MPC/Bundle/Modules/Multiplication"
	primebundle "MPC/Bundle/Prime-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	simplesharing "MPC/Secret-Sharing/Simple-Sharing"
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
		loadCircuit()
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
	//Udregn circuit
	for _, gate := range circuit.Gates {
		if myPartyNumber == gate.Input_one || myPartyNumber == gate.Input_two {
			if gate.Operation == "Multiplication" {
				multiplyResult := Multiplication.Multiply(secret, secretSharing, partySize)
				addResult := Add.Add(multiplyResult, secretSharing, partySize)
			} else if gate.Operation == "Addition" {
				addResult := Add.Add(secret, secretSharing, partySize)
			}
		} else {
			if partySize < gate.Input_one || myPartyNumber < gate.Input_two {
				if gate.Operation == "Multiplication" {
					multiplyResult := Multiplication.Multiply(gate.IntermediateRes, secretSharing, partySize)
					addResult := Add.Add(multiplyResult, secretSharing, partySize)
					gate.IntermediateRes = addResult
				} else if gate.Operation == "Addition" {
					addResult := Add.Add(gate.IntermediateRes, secretSharing, partySize)
				}
			}
			if gate.Operation == "Multiplication" {
				multiplyResult := Multiplication.Multiply(-1, secretSharing, partySize)
				addResult := Add.Add(multiplyResult, secretSharing, partySize)
			} else if gate.Operation == "Addition" {
				addResult := Add.Add(secret, secretSharing, partySize)
			}
		}
	}

	//multiplyResult := Multiplication.Multiply(secret, secretSharing, partySize)
	//addResult := Add.Add(multiplyResult, secretSharing, partySize)
	//fmt.Println("Done adding, got result:", addResult)
}

func createField(fieldSize int) {
	finiteField.SetSize(fieldSize)
	secretSharing.SetField(finiteField)
}

func loadCircuit() {
	jsonFile, err := os.Open("Circuit.json")
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