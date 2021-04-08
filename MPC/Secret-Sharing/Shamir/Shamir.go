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
	"reflect"
	"sort"
	"sync"
)

type Shamir struct {

}

type Receiver struct {

}


func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "Share"{
			wiresMutex.Lock()
			wires[match.Gate] = match.Shares[0]
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
		} else if match.Type == "Result" {
			resultMutex.Lock()
			receivedResults = resultGate[match.Gate]
			if receivedResults == nil {
				receivedResults = make(map[int]finite.Number)
			}
			receivedResults[match.From] = match.Result
			resultGate[match.Gate] = receivedResults
			resultMutex.Unlock()
			//receivedResults[match.From] = match.Result
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
var resultMutex = &sync.Mutex{}
var resultGate = make(map[int]map[int]finite.Number)
var receivedResults = make(map[int]finite.Number)

func (s Shamir) SetFunction(f string) {
	function = f
}

func (s Shamir) ComputeShares(parties int, secret finite.Number) []finite.Number {
	return field.ComputeShares(parties, secret)
}


func (s Shamir) ComputeResult(results []finite.Number) finite.Number {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
	//return Reconstruct(shares)
}

var field finite.Finite

func (s Shamir) SetField(f finite.Finite) {
	field = f
}

func (s Shamir) TheOneRing(circuit Circuit.Circuit, secret finite.Number) finite.Number {

	partySize := network.GetParties()

	receiver := Receiver{}

	network.RegisterReceiver(receiver)


	var result finite.Number
	switch field.(type) {
		case Binary.Binary:
			fmt.Println("Im party", network.GetPartyNumber())
			//Udregn shares
			for i, sec := range secret.Binary {
				binarySec := make([]int, 8)
				binarySec[7] = sec
				share := s.ComputeShares(partySize, finite.Number{Binary: binarySec})
				distributeShares(share, partySize, network.GetPartyNumber() * len(secret.Binary) + i - len(secret.Binary) + 1)
			}
		case Prime.Prime:
			shares := s.ComputeShares(partySize, secret)
			distributeShares(shares, partySize, network.GetPartyNumber())
	}
	//shares := s.ComputeShares(partySize, secret)



	for {
		for i, gate := range circuit.Gates {
			wiresMutex.Lock()
			input1, found1 := wires[gate.Input_one]
			input2, found2 := wires[gate.Input_two]
			wiresMutex.Unlock()
			if found1 && found2 || found1 && gate.Input_two == 0 {
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
					output = Reconstruct(multMaaaaap)
				case "Output":
					distributeResult([]finite.Number{input1}, partySize, gate.GateNumber)
					/*for {
						resultMutex.Lock()
						resultLen := len(resultGate[gate.GateNumber])
						resultMutex.Unlock()
						if resultLen == partySize {
							break
						}
					}*/
				case "Multiply-Constant":
					constantString := gate.Input_constant
					constant := field.GetConstant(constantString)
					output = field.Mul(input1, constant)
				}
				wiresMutex.Lock()
				wires[gate.GateNumber] = output
				wiresMutex.Unlock()
				circuit.Gates = removeGate(circuit, gate, i)
				if len(circuit.Gates) == 0 {
					//distributeResult([]finite.Number{output}, partySize)
				}
				//Restart for-loop
				break
				//Remove gate from circuits.gates
			}
		}
		var done = false
		if len(circuit.Gates) != 0 {
			continue//distributeResult([]finite.Number{output}, partySize)
		}
		switch field.(type) {
			case Prime.Prime:
				for {
					if len(resultGate) > 0 {
						break
					}
				}
				keys := reflect.ValueOf(resultGate).MapKeys()
				key := keys[0]
				if len(resultGate[(key.Interface()).(int)]) == partySize {
					result = Reconstruct(resultGate[(key.Interface()).(int)])
					done = true
				}

			case Binary.Binary:
				if len(resultGate) == len(secret.Binary) {
					keys := reflect.ValueOf(resultGate).MapKeys()
					var keysArray []int
					for _, k := range keys {
						keysArray = append(keysArray, (k.Interface()).(int))
					}
					sort.Ints(keysArray)
					trueResult := make([]int, len(secret.Binary))
					for i, v := range keysArray {
						hej := Reconstruct(resultGate[v]).Binary[7]
						fmt.Println("Result for gate", v, "is", hej)
						trueResult[i] = hej
					}
					result = finite.Number{Binary: trueResult}
					done = true
				}
		}
		if done {
			break
		}
		//result = Reconstruct(receivedResults)
		//fmt.Println(result)
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

func distributeShares(shares []finite.Number, partySize int, gate int) {
	fmt.Println("Sending shares", shares)
	fmt.Println("For wire", gate)
	for party := 1; party <= partySize; party++ {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: []finite.Number{shares[party - 1]},
			From:   network.GetPartyNumber(),
			Gate: 	gate,
		}

		if network.GetPartyNumber() == party {
			wiresMutex.Lock()
			wires[gate] = shares[party - 1]
			wiresMutex.Unlock()
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []finite.Number, partySize int, gate int) {
	for party := 1; party <= partySize; party++ {
		if network.GetPartyNumber() != party {
			shareBundle := numberbundle.NumberBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Result",
				Result: result[0],
				From: network.GetPartyNumber(),
				Gate: gate,
			}

			network.Send(shareBundle, party)
		} else {
			resultMutex.Lock()
			receivedResults = resultGate[gate]
			if receivedResults == nil {
				receivedResults = make(map[int]finite.Number)
			}
			receivedResults[network.GetPartyNumber()] = result[0]
			resultGate[gate] = receivedResults
			resultMutex.Unlock()
			//receivedResults[network.GetPartyNumber()] = result[0]
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

func (s Shamir) ComputeFunction(shares map[int][]finite.Number, party int) []finite.Number {
	//Reconstruct(shares)
	if function == "add" {

	}
	return nil
}