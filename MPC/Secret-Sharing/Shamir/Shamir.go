package Shamir

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	_ "crypto/rand"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

type Shamir struct {

}

type Receiver struct {

}


func (r Receiver) Receive(bundle bundle.Bundle) {
	fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "Share"{
			wiresMutex.Lock()
			wires[match.From] = match.Shares[0]
			wiresMutex.Unlock()
		} else if match.Type == "MultShare" {
			gateMutex.Lock()
			multMap = gateMult[match.Gate]
			if multMap == nil {
				multMap = make(map[int]finite.Number)
			}
			multMap[match.From] = match.Shares[0]
			gateMult[match.Gate] = multMap
			gateMutex.Unlock()
			fmt.Println("Gate map after getting a share", gateMult)
			fmt.Println("The ")
		} else if match.Type == "Result" {
			receivedResults[match.From] = match.Result
		} else {
			panic("Given type is unknown")
		}
	}
}

var function string
var wires = make(map[int]finite.Number)
var multMap = make(map[int]finite.Number)
var gateMult = make(map[int]map[int]finite.Number)
var wiresMutex = &sync.Mutex{}
var gateMutex = &sync.Mutex{}
var receivedResults = make(map[int]finite.Number)

func (s Shamir) SetFunction(f string) {
	function = f
}

func (s Shamir) ComputeShares(parties int, secret finite.Number) []finite.Number {
	return field.ComputeShares(parties, secret)
}


func (s Shamir) ComputeResult(ints []*big.Int) int {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
	//return Reconstruct(shares)
}

var field finite.Finite

func (s Shamir) SetField(f finite.Finite) {
	rand.Seed(time.Now().UnixNano())
	field = f
}

func (s Shamir) TheOneRing(circuit Circuit.Circuit, sec int) finite.Number {
	var secret finite.Number
	partySize := network.GetParties()

	receiver := Receiver{}

	network.RegisterReceiver(receiver)
	switch field.(type) {
	case Prime.Prime:
		secret = finite.Number{Prime: big.NewInt(int64(sec))}
	case Binary.Binary:
		secretBinary := make([]int, 8)
		secretBinary[7] = sec
		secret = finite.Number{Binary: secretBinary}
	}

	var result finite.Number

	shares := s.ComputeShares(partySize, secret)
	fmt.Println("Shares to the problem", shares)
	distributeShares(shares, partySize)

	for {
		for i, gate := range circuit.Gates {
			wiresMutex.Lock()
			input1, found1 := wires[gate.Input_one]
			input2, found2 := wires[gate.Input_two]
			wiresMutex.Unlock()
			if found1 && found2 {
				fmt.Println("Gate ready")
				fmt.Println(gate)
				//do operation
				var output finite.Number
				switch gate.Operation {
				case "Addition":
					fmt.Println(input1)
					fmt.Println(input2)
					output = field.Add(input1, input2)
					//output = new(big.Int).Add(input1, input2)
					//output.Mod(output, field.GetSize())
				case "Multiplication":
					interMult := field.Mul(input1, input2)
					//interMult.Mod(interMult, field.GetSize())
					multShares := s.ComputeShares(partySize, interMult)
					fmt.Println("multshares", multShares)
					distributeMultShares(multShares, partySize, gate.GateNumber)
					for {
						gateMutex.Lock()
						multLen := len(gateMult[gate.GateNumber])
						gateMutex.Unlock()
						if multLen == partySize {
							break
						}
					}
					gateMutex.Lock()
					multMaaaaap := gateMult[gate.GateNumber]
					gateMutex.Unlock()
					fmt.Println("Gatemap", gateMult)
					fmt.Println("Mult map to reconstruct", multMaaaaap)
					output = Reconstruct(multMaaaaap)
				}
				wiresMutex.Lock()
				wires[gate.GateNumber] = output
				wiresMutex.Unlock()
				circuit.Gates = removeGate(circuit, gate, i)
				if len(circuit.Gates) == 0 {
					distributeResult([]finite.Number{output}, partySize)
				}
				break
				//Remove gate from circuits.gates
			}
		}
		if len(receivedResults) == partySize {
			result = Reconstruct(receivedResults)
			break
		}
	}

	return result
}


func distributeMultShares(shares []finite.Number, partySize int, gate int) {
	for party := 1; party <= partySize; party++ {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "MultShare",
			Shares: []finite.Number{shares[party - 1]},
			From:   network.GetPartyNumber(),
			Gate: gate,
		}

		if network.GetPartyNumber() == party {
			gateMutex.Lock()
			multMap = gateMult[gate]
			if multMap == nil {
				multMap = make(map[int]finite.Number)
			}
			multMap[party] = shares[party - 1]
			gateMult[gate] = multMap
			gateMutex.Unlock()
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeShares(shares []finite.Number, partySize int) {
	fmt.Println(shares)
	for party := 1; party <= partySize; party++ {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: []finite.Number{shares[party - 1]},
			From:   network.GetPartyNumber(),
		}

		if network.GetPartyNumber() == party {
			wiresMutex.Lock()
			wires[party] = shares[party - 1]
			wiresMutex.Unlock()
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []finite.Number, partySize int) {
	for party := 1; party <= partySize; party++ {
		if network.GetPartyNumber() != party {
			shareBundle := numberbundle.NumberBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Result",
				Result: result[0],
				From: network.GetPartyNumber(),
			}

			network.Send(shareBundle, party)
		} else {
			receivedResults[network.GetPartyNumber()] = result[0]
		}
	}
}

func removeGate(circuit Circuit.Circuit, gate Circuit.Gate, i int) []Circuit.Gate {
	b := make([]Circuit.Gate, len(circuit.Gates))
	copy(b, circuit.Gates)
	// Remove the element at index i from a.
	b[i] = b[len(b)-1] // Copy last element to index i.
	b = b[:len(b)-1]   // Truncate slice.
	return b
}

func (s Shamir) ComputeFunction(shares map[int][]*big.Int, party int) []*big.Int {
	//Reconstruct(shares)
	if function == "add" {

	}
	return nil
}