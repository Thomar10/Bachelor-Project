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
var resultMapMutex = &sync.Mutex{}
var resultGate = make(map[int]map[int]finite.Number)
var receivedResults = make(map[int]finite.Number)
var corrupts = 0

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
	corrupts = (network.GetParties() - 1) / 2
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


	outputGates := outputSize(circuit)
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
					distributeMultShares(multShares, partySize, gate.GateNumber)
					gateMutex.Lock()
					multMaap := gateMult[gate.GateNumber]
					gateMutex.Unlock()
					for {
						if len(multMaap) == 2 * corrupts + 1  {
							break
						}else if len(multMaap) > 2 * corrupts + 1  {
							for k, _ := range multMaap {
								delete(multMaap, k)
								break
							}
						}

					}
					output = Reconstruct(multMaap)

				case "Output":
					distributeResult([]finite.Number{input1}, gate.Output, gate.GateNumber)
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
					if outputGates == 0 {
						//No outputs for this party - return 0
						result.Prime = big.NewInt(0)
						return result
					}
				}
				keys := reflect.ValueOf(resultGate).MapKeys()
				key := keys[0]
				if len(resultGate[(key.Interface()).(int)]) == corrupts + 1 {
					result = Reconstruct(resultGate[(key.Interface()).(int)])
					done = true
				} else if len(resultGate[(key.Interface()).(int)]) > corrupts + 1 {
					resultMapMutex.Lock()
					//Fjern indtil vi er på corrupts + 1
					resultMap := resultGate[(key.Interface()).(int)]
					for {
						if len(resultGate[(key.Interface()).(int)]) == corrupts + 1 {
							break
						} else {
							//Iterating over map gives keys in random order
							for k, _ := range resultMap {
								delete(resultMap, k)
								break
							}
						}
					}
					result = Reconstruct(resultMap)
					done = true
					resultMapMutex.Unlock()
				}

			case Binary.Binary:
				if outputGates > 0 {
					trueResult := make([]int, outputGates)
					if len(resultGate) == outputGates {
						keys := reflect.ValueOf(resultGate).MapKeys()
						var keysArray []int
						for _, k := range keys {
							keysArray = append(keysArray, (k.Interface()).(int))
						}
						sort.Ints(keysArray)
						for i, k := range keysArray {
							for {
								if len(resultGate[k]) == corrupts + 1  {
									resultMapMutex.Lock()
									resultBit := Reconstruct(resultGate[k]).Binary[7]
									trueResult[i] = resultBit
									resultMapMutex.Unlock()
									break
								}else if len(resultGate[k]) > corrupts + 1 {
									for j, _ := range resultGate[k] {
										delete(resultGate[k], j)
										break
									}
								}
							}
						}

						/*for i, v := range keysArray {
							resultBit := Reconstruct(resultGate[v]).Binary[7]
							fmt.Println(resultGate[v])
							fmt.Println("Result for gate", v, "is", resultBit)
							trueResult[i] = resultBit
						}*/
						//resultMapMutex.Unlock()
						result = finite.Number{Binary: trueResult}
						done = true
					}
				} else {
					result = finite.Number{Binary: []int{0}}
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

func outputSize(circuit Circuit.Circuit) int {
	result := 0
	for _, gate := range circuit.Gates {
		if gate.Operation == "Output" && gate.Output == network.GetPartyNumber() {
			result++
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

func distributeShares(shares []finite.Number, partySize int, gate int) {
	fmt.Println("Sending shares", shares)
	fmt.Println("For wire", gate)

	for party := 1; party <= partySize; party++ {
		fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
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

func distributeResult(result []finite.Number, party int, gate int) {
	if network.GetPartyNumber() != party {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Result",
			Result: result[0],
			From:   network.GetPartyNumber(),
			Gate:   gate,
		}
		network.Send(shareBundle, party)
	}else {
		resultMutex.Lock()
		receivedResults = resultGate[gate]
		if receivedResults == nil {
			receivedResults = make(map[int]finite.Number)
		}
		receivedResults[network.GetPartyNumber()] = result[0]
		resultGate[gate] = receivedResults
		resultMutex.Unlock()
	}
/*	for party := 1; party <= partySize; party++ {
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
	}*/
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