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
var waitTime int
var wasParty1 = false


func resetTheWholeShit() {
	sizeSetMutex.Lock()
	sizeSet = false
	sizeSetMutex.Unlock()
	fmt.Println("Resetting done list")
	doneMutex.Lock()
	doneList = []int{}
	doneMutex.Unlock()
	fmt.Println("Resetting network")
	//network.ResetNetwork()
	Preparation.ResetPrep()
}

func testInitNetwork(circuitToLoad, hostAddress string) {
	loadCircuit(circuitToLoad + ".json")
	if circuit.SecretSharing == "Shamir" {
		secretSharing = Shamir.Shamir{}
	} else {
		secretSharing = Simple_Sharing.Simple_Sharing{}
	}
	secretSharing.ResetSecretSharing()
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
	isFirst := network.InitToHost(partySize, hostAddress)

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
}

func main() {
	if os.Args[1] == "test" {
		var avgTime time.Duration
		minTime := 9999 * time.Second
		maxTime := time.Duration(0)
		testInitNetwork("YaoBits20","192.168.1.100:62123")
		for i:= 0; i < 100; i++ {
			fmt.Println("Im on iteration", i + 1)
			secretToTest := finite.Number{Prime: big.NewInt(5)}
			result, timee := MPCTest(secretToTest)
			waitTime = network.GetPartyNumber()
			fmt.Println("Result", result)
			fmt.Println("Took", timee)
			avgTime += timee
			if timee < minTime {
				minTime = timee
			}
			if timee > maxTime {
				maxTime = timee
			}
			if network.GetPartyNumber() == 1 {
				wasParty1 = true
				if result.Binary[0] != 1 {
					panic("Got wrong result for the test")
				}
			}
			resetTheWholeShit()
			if !wasParty1 {
				time.Sleep(5 * time.Second)
			}

		}
		fmt.Println("It took on average", avgTime / 100)
		fmt.Println("The lowest runtime", minTime)
		fmt.Println("The highest runtime", maxTime)

	} else {
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


}


func MPCTest(secret finite.Number) (finite.Number, time.Duration) {
	secretSharing.ResetSecretSharing()
	if preprocessing {
		fmt.Println("Preprocessing!")
		corrupts := (partySize - 1) / 2
		Preparation.Prepare(circuit, finiteField, corrupts, secretSharing)
		fmt.Println("Done preprocessing")
	}
	for {
		if network.GetPartyNumber() == 1 {
			secret = finite.Number{Prime: big.NewInt(5), Binary: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
			break
		}
		if network.GetPartyNumber() == 2 {
			secret = finite.Number{Prime: big.NewInt(5), Binary: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
			break
		}else {
			secret = finite.Number{Prime: big.NewInt(5), Binary: []int{}}
			break
		}
	}

	fmt.Println("Im calling the one ring with secret", secret)
	startTime := time.Now()
	result := secretSharing.TheOneRing(circuit, secret, preprocessing)
	endTime := time.Since(startTime)
	fmt.Println("got a result")
	distributeDone()
	for {
		doneMutex.Lock()
		//fmt.Println("doneList", doneList)
		if len(doneList) == partySize {
			doneMutex.Unlock()
			break
		}
		doneMutex.Unlock()
	}
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
		bundlee := numberbundle.NumberBundle{
			ID:    uuid.Must(uuid.NewRandom()).String(),
			Type:  "Done",
			From:  me,
		}
		if party == network.GetPartyNumber() {
			doneMutex.Lock()
			doneList = append(doneList, me)
			doneMutex.Unlock()
		}else {
			network.Send(bundlee, party)
		}
	}

}