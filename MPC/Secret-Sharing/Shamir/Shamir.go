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
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type Shamir struct {

}

type Receiver struct {

}


func (r Receiver) Receive(bundle bundle.Bundle) {
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "Share"{
			wiresMutex.Lock()
			wires[match.Gate] = match.Shares[0]
			wiresMutex.Unlock()
		} else if match.Type == "MultShare" {
			gateMutex.Lock()
			multMap := gateMult[match.Gate]
			if multMap == nil {
				multMap = make(map[int]finite.Number)
			}
			multMap[match.From] = match.Shares[0]
			gateMult[match.Gate] = multMap
			gateMutex.Unlock()

		} else if match.Type == "Result" {
			fmt.Println("Got a result", match)
			resultMutex.Lock()
			receivedResults = resultGate[match.Gate]
			if receivedResults == nil {
				receivedResults = make(map[int]finite.Number)
			}
			receivedResults[match.From] = match.Result
			resultGate[match.Gate] = receivedResults
			resultMutex.Unlock()
		}else if match.Type == "EDShare" {
			eMultMutex.Lock()
			eMultMap := eMult[match.Gate]
			if eMultMap == nil {
				eMultMap = make(map[int]finite.Number)
			}
			eMultMap[match.From] = match.Shares[0]
			eMult[match.Gate] = eMultMap
			eMultMutex.Unlock()
			dMultMutex.Lock()
			dMultMap := dMult[match.Gate]
			if dMultMap == nil {
				dMultMap = make(map[int]finite.Number)
			}
			dMultMap[match.From] = match.Shares[1]
			dMult[match.Gate] = dMultMap
			dMultMutex.Unlock()
		} else if match.Type == "EDResult" {
			eOpenMutex.Lock()
			eOpenMap[match.Gate] = match.Shares[0]
			eOpenMutex.Unlock()

			dOpenMutex.Lock()
			dOpenMap[match.Gate] = match.Shares[1]
			dOpenMutex.Unlock()
		}
	}
}

var function string
var wires = make(map[int]finite.Number)
var gateMult = make(map[int]map[int]finite.Number)
var eMult = make(map[int]map[int]finite.Number)
var dMult = make(map[int]map[int]finite.Number)
var eOpenMap = make(map[int]finite.Number)
var dOpenMap = make(map[int]finite.Number)
var eMultMutex = &sync.Mutex{}
var dMultMutex = &sync.Mutex{}
var eOpenMutex = &sync.Mutex{}
var dOpenMutex = &sync.Mutex{}
var wiresMutex = &sync.Mutex{}
var gateMutex = &sync.Mutex{}
var resultMutex = &sync.Mutex{}
var resultGate = make(map[int]map[int]finite.Number)
var receivedResults = make(map[int]finite.Number)
var corrupts = 0
var tripleCounter = 1
var field finite.Finite
var x = make(map[int]finite.Number)
var y = make(map[int]finite.Number)
var z = make(map[int]finite.Number)
var EDReconstructionCounter = 0



func (s Shamir) ResetSecretSharing() {
	function = ""
	wires = make(map[int]finite.Number)
	gateMult = make(map[int]map[int]finite.Number)
	eMult = make(map[int]map[int]finite.Number)
	dMult = make(map[int]map[int]finite.Number)
	eOpenMap = make(map[int]finite.Number)
	dOpenMap = make(map[int]finite.Number)
	resultGate = make(map[int]map[int]finite.Number)
	receivedResults = make(map[int]finite.Number)
	corrupts = 0
	tripleCounter = 1
	x = make(map[int]finite.Number)
	y = make(map[int]finite.Number)
	z = make(map[int]finite.Number)
	EDReconstructionCounter = 0
}



func (s Shamir) SetFunction(f string) {
	function = f
}

func (s Shamir) ComputeShares(parties int, secret finite.Number) []finite.Number {
	return field.ComputeShares(parties, secret)
}


func (s Shamir) ComputeResult(results []finite.Number) finite.Number {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
}


func (s Shamir) SetField(f finite.Finite) {
	field = f
}

//Returns a new triple (x, y, z) with xy=z
func getTriple() []finite.Number {
	result := []finite.Number{x[tripleCounter], y[tripleCounter], z[tripleCounter]}
	tripleCounter++
	return result
}

//Sets all the triples created in Preperation
func (s Shamir) SetTriple(xMap, yMap, zMap map[int]finite.Number) {
	x = xMap
	y = yMap
	z = zMap
}

//Register a receiver
func (s Shamir) RegisterReceiver() {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)
}


func (s Shamir) TheOneRing(circuit Circuit.Circuit, secret finite.Number, preprocessed bool) finite.Number {
	corrupts = (network.GetParties() - 1) / 2
	partySize := network.GetParties()

	doesIHaveAnInput := false
	iAm := network.GetPartyNumber()
	for _, gate := range circuit.Gates {
		if gate.Input_one == iAm || gate.Input_two == iAm {
			if gate.Operation == "Output" {
				continue
			}
			doesIHaveAnInput = true
			break
		}
	}
	var result finite.Number
	switch field.(type) {
		case Binary.Binary:
			for i, sec := range secret.Binary {
				binarySec := make([]int, 8)
				binarySec[7] = sec
				share := s.ComputeShares(partySize, finite.Number{Binary: binarySec})
				distributeShares(share, partySize, network.GetPartyNumber() * len(secret.Binary) + i - len(secret.Binary) + 1)
			}
		case Prime.Prime:
			if doesIHaveAnInput {
				shares := s.ComputeShares(partySize, secret)
				distributeShares(shares, partySize, network.GetPartyNumber())
			}
	}

	outputGates := outputSize(circuit)
	for {
		for i, gate := range circuit.Gates {
			wiresMutex.Lock()
			input1, found1 := wires[gate.Input_one]
			input2, found2 := wires[gate.Input_two]
			wiresMutex.Unlock()
			//Found1 and found2 if for multiplication and addition gates
			//Found1 and input2 = 0 is for multiply-with-constant and output gates
			if found1 && found2 || found1 && gate.Input_two == 0 {
				var output finite.Number
				switch gate.Operation {
				case "Addition":
					output = field.Add(input1, input2)
					wiresMutex.Lock()
					wires[gate.GateNumber] = output
					wiresMutex.Unlock()
				case "Multiplication":
					if preprocessed  {
						//Turn false for concurrent multiplication
						if true {
							output = processedMultReturn(input1, input2, gate, partySize)
							wiresMutex.Lock()
							wires[gate.GateNumber] = output
							wiresMutex.Unlock()
						}else {
							go processedMult(input1, input2, gate, partySize)
						}
					}else {
						output = nonProcessedMult(input1, input2, gate, partySize)
						wiresMutex.Lock()
						wires[gate.GateNumber] = output
						wiresMutex.Unlock()
					}

				case "Output":
					distributeResult([]finite.Number{input1}, gate.Output, gate.GateNumber)

				case "Multiply-Constant":
					constantString := gate.Input_constant
					constant := field.GetConstant(constantString)
					output = field.Mul(input1, constant)
					wiresMutex.Lock()
					wires[gate.GateNumber] = output
					wiresMutex.Unlock()
				}

				//Remove gate from circuits.gates, so we do not iterate the same gate again
				circuit.Gates = removeGate(circuit, i)
				//Restart for-loop
				break
			}
		}
		var done = false
		if len(circuit.Gates) != 0 {
			continue
		}
		switch field.(type) {
			case Prime.Prime:
				for {
					resultMutex.Lock()
					resultLen := len(resultGate)
					resultMutex.Unlock()
					if resultLen > 0 {
						break
					}
					if outputGates == 0 {
						//No outputs for this party - return 0
						result.Prime = big.NewInt(0)
						return result
					}
				}
				resultMutex.Lock()
				keys := reflect.ValueOf(resultGate).MapKeys()
				key := keys[0]
				if len(resultGate[(key.Interface()).(int)]) >= corrupts + 1 {
					result = Reconstruct(resultGate[(key.Interface()).(int)])
					done = true
				}
				resultMutex.Unlock()
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
								resultMutex.Lock()
								resultMapLen := len(resultGate[k])
								resultMutex.Unlock()
								if resultMapLen >= corrupts + 1  {
									resultMutex.Lock()
									resultBit := Reconstruct(resultGate[k]).Binary[7]
									trueResult[i] = resultBit
									resultMutex.Unlock()
									break
								}
							}
						}
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
	}
	return result
}


func nonProcessedMult(input1, input2 finite.Number, gate Circuit.Gate, partySize int) finite.Number {
	interMult := field.Mul(input1, input2)
	multShares := field.ComputeShares(partySize, interMult)
	distributeMultShares(multShares, partySize, gate.GateNumber)
	for {
		gateMutex.Lock()
		multMaap := gateMult[gate.GateNumber]
		multMapLen := len(multMaap)
		if multMapLen >= network.GetParties() {//2 * corrupts + 1  {
			result := Reconstruct(multMaap)
			gateMutex.Unlock()
			return result
		}
		gateMutex.Unlock()
	}


}
func processedMultReturn(input1, input2 finite.Number, gate Circuit.Gate, partySize int) finite.Number{
	triple := getTriple()
	xt := field.Mul(triple[0], finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)}) //-x
	yt := field.Mul(triple[1], finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)}) //-y
	e := field.Add(input1, xt) //input1 - x
	d := field.Add(input2, yt) //input2 - y
	reconstructED(e, d, partySize, gate)
	var eOpen, dOpen finite.Number
	//Wait for the open values of e and d to be present in the map
	for {
		eOpenMutex.Lock()
		eOpenValue, foundE := eOpenMap[gate.GateNumber]
		eOpenMutex.Unlock()
		dOpenMutex.Lock()
		dOpenValue, foundD := dOpenMap[gate.GateNumber]
		dOpenMutex.Unlock()
		if foundE && foundD {
			eOpen = eOpenValue
			dOpen = dOpenValue
			break
		}
	}
	//Calculate ab
	eb := field.Mul(eOpen, input2)
	da := field.Mul(dOpen, input1)
	edInv := field.Mul(field.Mul(eOpen, dOpen), finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)})
	daed := field.Add(da, edInv)
	ebdaed := field.Add(eb, daed)
	ab := field.Add(triple[2], ebdaed)
	return ab
}

func processedMult(input1, input2 finite.Number, gate Circuit.Gate, partySize int) {
	triple := getTriple()
	xt := field.Mul(triple[0], finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)}) //-x
	yt := field.Mul(triple[1], finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)}) //-y
	e := field.Add(input1, xt)//input1 - triple[0]
	d := field.Add(input2, yt)//input2 - triple[1]
	reconstructED(e, d, partySize, gate)
	var eOpen, dOpen finite.Number
	for {
		eOpenMutex.Lock()
		eOpenValue, foundE := eOpenMap[gate.GateNumber]
		eOpenMutex.Unlock()
		dOpenMutex.Lock()
		dOpenValue, foundD := dOpenMap[gate.GateNumber]
		dOpenMutex.Unlock()
		if foundE && foundD {
			eOpen = eOpenValue
			dOpen = dOpenValue
			break
		}
	}
	//Calculate ab
	eb := field.Mul(eOpen, input2)
	da := field.Mul(dOpen, input1)
	edInv := field.Mul(field.Mul(eOpen, dOpen), finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)})
	daed := field.Add(da, edInv)
	ebdaed := field.Add(eb, daed)
	ab := field.Add(triple[2], ebdaed)
	wiresMutex.Lock()
	wires[gate.GateNumber] = ab
	wiresMutex.Unlock()
}

//Distributes the shares e and d. If its the parties turn to reconstruct
//the party will also wait for enough shares to reconstruct and distribute
//the open value of e and d
func reconstructED(e, d finite.Number, partySize int, gate Circuit.Gate) {
	//Distribute the e and d share for reconstruction
	distributeED([]finite.Number{e, d}, partySize, gate.GateNumber, false)
	//Should I reconstruct e and d
	if (gate.GateNumber % partySize) + 1 == network.GetPartyNumber()  {
		EDReconstructionCounter++
		//Reconstruct e
		for {
			eMultMutex.Lock()
			eMultLength :=  len(eMult[gate.GateNumber])
			eMultMutex.Unlock()
			if eMultLength >= corrupts + 1 {
				break
			}
		}
		eMultMutex.Lock()
		eMultGate := eMult[gate.GateNumber]
		eOpen := Reconstruct(eMultGate)
		eMultMutex.Unlock()

		//Reconstruct d
		dMultMutex.Lock()
		dMultLength :=  len(dMult[gate.GateNumber])
		dMultMutex.Unlock()
		for {
			if dMultLength >= corrupts + 1 {
				break
			}
		}
		dMultMutex.Lock()
		dMultGate := dMult[gate.GateNumber]
		dOpen := Reconstruct(dMultGate)
		dMultMutex.Unlock()
		distributeED([]finite.Number{eOpen, dOpen}, partySize, gate.GateNumber, true)
	}
}

//Returns the number of output gates for the party
func outputSize(circuit Circuit.Circuit) int {
	result := 0
	for _, gate := range circuit.Gates {
		if gate.Operation == "Output" && gate.Output == network.GetPartyNumber() {
			result++
		}
	}
	return result
}

//Distributes shares e and d. E needs to be places on the first index (0) and d on second index (1)
//If forAll is true distribute the open value of e and d
//If forAll is false distribute e and d shares to be open to the correct party to reconstruct
func distributeED(shares []finite.Number, partySize int, gate int, forAll bool) {
	if forAll {
		for party := 1; party <= partySize; party++ {
			shareBundle := numberbundle.NumberBundle{
				ID:     uuid.Must(uuid.NewRandom()).String(),
				Type:   "EDResult",
				Shares: shares,
				From:   network.GetPartyNumber(),
				Gate: 	gate,
			}
			if party == network.GetPartyNumber() {
				eOpenMutex.Lock()
				eOpenMap[gate] = shares[0]
				eOpenMutex.Unlock()
				dOpenMutex.Lock()
				dOpenMap[gate] = shares[1]
				dOpenMutex.Unlock()
			} else {
				network.Send(shareBundle, party)
			}
		}
	}else {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "EDShare",
			Shares: shares,
			From:   network.GetPartyNumber(),
			Gate:   gate,
		}

		if network.GetPartyNumber() == (gate % partySize) + 1 {
			eMultMutex.Lock()
			eMultMap := eMult[gate]
			if eMultMap == nil {
				eMultMap = make(map[int]finite.Number)
			}
			eMultMap[(gate % partySize) + 1 ] = shares[0]
			eMult[gate] = eMultMap
			eMultMutex.Unlock()
			dMultMutex.Lock()
			dMultMap := dMult[gate]
			if dMultMap == nil {
				dMultMap = make(map[int]finite.Number)
			}
			dMultMap[(gate % partySize) + 1 ] = shares[1]
			dMult[gate] = dMultMap
			dMultMutex.Unlock()
		} else {
			network.Send(shareBundle, (gate % partySize) + 1 )
		}
	}
}

//Distributes multiplication shares for the non processed protocol
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
			multMap := gateMult[gate]
			if multMap == nil {
				multMap = make(map[int]finite.Number)
			}
			multMap[party] = shares[party - 1]
			gateMult[gate] = multMap
			gateMutex.Unlock()
		}else {
			network.Send(shareBundle, party)
		}
	}
}

//Distributes the shares for the protocol
func distributeShares(shares []finite.Number, partySize int, gate int) {
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
		}else {
			network.Send(shareBundle, party)
		}
	}
}

//Distributes the result share for reconstruction
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
}

//Removes a gate from the gates in the circuit
func removeGate(circuit Circuit.Circuit, i int) []Circuit.Gate {
	b := make([]Circuit.Gate, len(circuit.Gates))
	copy(b, circuit.Gates)
	// Remove the element at index i from a.
	b[i] = b[len(b)-1] // Copy last element to index i.
	b = b[:len(b)-1]   // Truncate slice.
	return b
}

func (s Shamir) ComputeFunction(shares map[int][]finite.Number, party int) []finite.Number {
	panic("implement me")
}