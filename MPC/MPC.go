package main

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	"MPC/Preparation"
	secretsharing "MPC/Secret-Sharing"
	"MPC/Secret-Sharing/Shamir"
	Simple_Sharing "MPC/Secret-Sharing/Simple-Sharing"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
		case numberbundle.NumberBundle:
			if match.Type == "Prime" {
				myPartyNumber = network.GetPartyNumber()
				createField(finite.Number{Prime: match.Prime.Prime})
				sizeSet = true
			} /*else {
				panic("Given type is unknown: "+ match.Type)
			}*/
	}
}

var finiteField finite.Finite
var bundleType bundle.Bundle
var secretSharing secretsharing.Secret_Sharing
var partySize int
var secret finite.Number
var sizeSet bool
var myPartyNumber int
var circuit Circuit.Circuit
var preprocessing = true

func main() {

	circuitToLoad := os.Args[1]
	loadCircuit(circuitToLoad + ".json")
	var sec string
	if len(os.Args) > 2 {
		sec = os.Args[2]
	}else {
		sec = "-1"
	}
	if circuit.SecretSharing == "Shamir" {
		secretSharing = Shamir.Shamir{}
	}else {
		secretSharing = Simple_Sharing.Simple_Sharing{}
	}
	if circuit.Field == "Prime" {
		finiteField = Prime.Prime{}
		s, _ := strconv.Atoi(sec)
		secret = finite.Number{Prime: big.NewInt(int64(s))}
	}else {
		finiteField = Binary.Binary{}
		secByte := make([]int, len(sec))
		for i, r := range sec {
			secByte[i], _ = strconv.Atoi(string(r))
		}
		secret = finite.Number{Binary: secByte}
	}
	partySize = circuit.PartySize
	finiteField.InitSeed()
	bundleType = numberbundle.NumberBundle{}


	receiver := Receiver{}
	network.RegisterReceiver(receiver)
	isFirst := network.Init(partySize)
	if isFirst {
		finiteSize := finiteField.GenerateField()
		switch bundleType.(type) {
		case numberbundle.NumberBundle:
			bundleType = numberbundle.NumberBundle{
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
	if preprocessing {
		corrupts := (partySize - 1) / 2
		Preparation.Prepare(circuit, finiteField, corrupts, secretSharing)
		time.Sleep(1)
	}

	result := secretSharing.TheOneRing(circuit, secret, preprocessing)
	time.Sleep(10)
	switch finiteField.(type) {
		case Prime.Prime:
			fmt.Println("Final result:", result.Prime)
		case Binary.Binary:
			fmt.Println("Final result:", result.Binary)
	}


}

func createField(fieldSize finite.Number) {
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