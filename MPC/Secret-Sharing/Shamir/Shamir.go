package Shamir

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	network "MPC/Network"
	"fmt"
	"github.com/google/uuid"
	"math"
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
	case primebundle.PrimeBundle:
		if match.Type == "Share"{
			wiresMutex.Lock()
			wires[match.From] = match.Shares[0]
			wiresMutex.Unlock()
		} else if match.Type == "Result" {
			receivedResults[match.From] = match.Result
		} else {
			panic("Given type is unknown")
		}
	}
}

var function string
var wires = make(map[int]int)
var wiresMutex = &sync.Mutex{}
var receivedResults = make(map[int]int)

func (s Shamir) SetFunction(f string) {
	function = f
}

func (s Shamir) ComputeShares(parties, secret int) []int {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
	//3 + 4x + 2x^2
	//[3, 4, 2]
	var polynomial = make([]int, t + 1)
	polynomial[0] = secret
	for i := 1; i < t + 1; i++ {
		polynomial[i] = rand.Intn(field.GetSize())
	}

	var shares = make([]int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}

	return shares
}


func computeResultt(shares map[int]int, parties int) int {
	return Reconstruct(shares)

}

func (s Shamir) ComputeResult(ints []int) int {
	panic("implement meeeeeeeeeeeeeeeeeeeeee!")
	//return Reconstruct(shares)
}

var field finite.Finite

func (s Shamir) SetField(f finite.Finite) {
	rand.Seed(time.Now().UnixNano())
	field = f
}

func (s Shamir) TheOneRing(circuit Circuit.Circuit, secret int) int {
	partySize := network.GetParties()

	receiver := Receiver{}

	network.RegisterReceiver(receiver)

	result := 0

	shares := s.ComputeShares(partySize, secret)
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
				var output int
				switch gate.Operation {
				case "Addition":
					output = (input1 + input2) % field.GetSize()
				case "Multiplication":
					output = 3
				}
				wiresMutex.Lock()
				wires[gate.GateNumber] = output
				wiresMutex.Unlock()
				circuit.Gates = removeGate(circuit, gate, i)
				if len(circuit.Gates) == 0 {
					distributeResult([]int{output}, partySize)
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

func calculatePolynomial(polynomial []int, x int) int {
	var result = 0

	for i := 0; i < len(polynomial); i++ {
		result += polynomial[i] * int(math.Pow(float64(x), float64(i)))
	}

	return result % field.GetSize()
}

func (s Shamir) ComputeFunction(shares map[int][]int, party int) []int {
	//Reconstruct(shares)
	if function == "add" {

	}
	return nil
}

func distributeShares(shares []int, partySize int) {
	for party := 1; party <= partySize; party++ {
		shareBundle := primebundle.PrimeBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: []int{shares[party - 1]},
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

func distributeResult(result []int, partySize int) {

	for party := 1; party <= partySize; party++ {
		if network.GetPartyNumber() != party {
			shareBundle := primebundle.PrimeBundle{
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