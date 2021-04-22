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
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
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
			//Placer bare mutex for at -race ikke skriger om det
			sizeSetMutex.Lock()
			sizeSet = true
			sizeSetMutex.Unlock()
		}
		if match.Type == "Done" {
			doneMutex.Lock()
			doneList = append(doneList, match.From)
			doneMutex.Unlock()
		}
	}
}

var finiteField finite.Finite
var bundleType bundle.Bundle
var secretSharing secretsharing.Secret_Sharing
var partySize int
var secret finite.Number
var sizeSet bool
var doneList []int
var myPartyNumber int
var circuit Circuit.Circuit
var preprocessing = false
var doneMutex = &sync.Mutex{}
var sizeSetMutex = &sync.Mutex{}

func main() {

	circuitToLoad := os.Args[1]
	loadCircuit(circuitToLoad + ".json")
	var sec string
	if len(os.Args) > 2 {
		sec = os.Args[2]
	} else {
		sec = "-1"
	}
	if circuit.SecretSharing == "Shamir" {
		secretSharing = Shamir.Shamir{}
	} else {
		secretSharing = Simple_Sharing.Simple_Sharing{}
	}
	if circuit.Field == "Prime" {
		finiteField = Prime.Prime{}
		s, _ := strconv.Atoi(sec)
		secret = finite.Number{Prime: big.NewInt(int64(s))}
	} else {
		finiteField = Binary.Binary{}
		secByte := make([]int, len(sec))
		for i, r := range sec {
			secByte[i], _ = strconv.Atoi(string(r))
		}

		if sec == "-1" {
			secByte = make([]int, 0)
		}

		secret = finite.Number{Binary: secByte}
	}

	partySize = circuit.PartySize
	preprocessing = circuit.Preprocessing
	finiteField.InitSeed()
	bundleType = numberbundle.NumberBundle{}

	receiver := Receiver{}
	network.RegisterReceiver(receiver)
	Preparation.RegisterReceiver()
	secretSharing.RegisterReceiver()

	isFirst := network.Init(partySize)
	if isFirst {
		finiteSize := finiteField.GenerateField()
		switch bundleType.(type) {
		case numberbundle.NumberBundle:
			bundleType = numberbundle.NumberBundle{
				ID:    uuid.Must(uuid.NewRandom()).String(),
				Type:  "Prime",
				Prime: finiteSize,
			}
		default:
			fmt.Println(":(")
		}
		for {
			if network.IsReady() {
				myPartyNumber = network.GetPartyNumber()
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
		sizeSetMutex.Lock()
		sizeSetValue := sizeSet
		sizeSetMutex.Unlock()
		if sizeSetValue {
			break
		}
	}

	if preprocessing {
		fmt.Println("Preprocessing!")
		corrupts := (partySize - 1) / 2
		Preparation.Prepare(circuit, finiteField, corrupts, secretSharing)
	}

	fmt.Println("Done preprocessing")
	fmt.Println("I am party", network.GetPartyNumber())

	startTime := time.Now()
	result := secretSharing.TheOneRing(circuit, secret, preprocessing)
	endTime := time.Since(startTime)

	distributeDone()
	for {
		doneMutex.Lock()
		if len(doneList) == partySize {
			fmt.Println("Donelist", doneList)
			break
		}
		doneMutex.Unlock()
	}
	switch finiteField.(type) {
	case Prime.Prime:
		fmt.Println("Final result:", result.Prime)
	case Binary.Binary:
		fmt.Println("Final result:", result.Binary)
	}
	fmt.Println("The protocol took", endTime)
}


func MPCTest(circuitToLoad string, secret finite.Number, address, hostAddress string) (finite.Number, time.Duration) {
	loadCircuit(circuitToLoad)
	if circuit.SecretSharing == "Shamir" {
		secretSharing = Shamir.Shamir{}
	} else {
		secretSharing = Simple_Sharing.Simple_Sharing{}
	}
	if circuit.Field == "Prime" {
		finiteField = Prime.Prime{}
	}else {
		finiteField = Binary.Binary{}
	}
	partySize = circuit.PartySize
	preprocessing = circuit.Preprocessing
	finiteField.InitSeed()
	bundleType = numberbundle.NumberBundle{}

	receiver := Receiver{}
	network.RegisterReceiver(receiver)
	Preparation.RegisterReceiver()
	secretSharing.RegisterReceiver()

	network.InitWithHostAddress(partySize, address, hostAddress)

	startTime := time.Now()
	result := secretSharing.TheOneRing(circuit, secret, preprocessing)
	endTime := time.Since(startTime)

	return result, endTime
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
func distributeDone() {
	me := network.GetPartyNumber()
	for party := 1; party <= partySize; party++ {
		bundle := numberbundle.NumberBundle{
			ID:    uuid.Must(uuid.NewRandom()).String(),
			Type:  "Done",
			From:  me,
		}
		if party == network.GetPartyNumber() {
			doneMutex.Lock()
			doneList = append(doneList, me)
			doneMutex.Unlock()
		}else {
			network.Send(bundle, party)
		}
	}

}