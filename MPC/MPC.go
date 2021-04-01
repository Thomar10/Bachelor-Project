package main

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
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
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
		case numberbundle.NumberBundle:
			if match.Type == "Prime" {
				myPartyNumber = network.GetPartyNumber()
				createField(finite.Number{Prime: match.Prime.Prime})
				sizeSet = true
			} else {
				panic("Given type is unknown: "+ match.Type)
			}
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

func main() {
	if os.Args[1] == "-p" {
		finiteField = Prime.Prime{}
		sec, _ := strconv.Atoi(os.Args[4])
		secret = finite.Number{Prime: big.NewInt(int64(sec))}
	}else {
		finiteField = Binary.Binary{}
		sec :=  os.Args[4]
		secByte := make([]int, len(sec))
		for i, r := range sec {
			secByte[i], _ = strconv.Atoi(string(r))
		}
		secret = finite.Number{Binary: secByte}
	}
	finiteField.InitSeed()
	bundleType = numberbundle.NumberBundle{}
	if os.Args[2] == "-s" {
		secretSharing = Simple_Sharing.Simple_Sharing{}
		loadCircuit("SimpleCircuit.json")
	}else if os.Args[2] == "-sss" {
		secretSharing = Shamir.Shamir{}
		loadCircuit("Circuit.json")
		//loadCircuit("2BitAdder.json")
		fmt.Println("Circuit", circuit)
	} else {
		panic("No secret sharing given")
	}
	partySize, _ = strconv.Atoi(os.Args[3])



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

	result := secretSharing.TheOneRing(circuit, secret)
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