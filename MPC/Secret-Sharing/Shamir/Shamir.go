package Shamir

import (
	bundle "MPC/Bundle"
	primebundle "MPC/Bundle/Prime-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	network "MPC/Network"
	crand "crypto/rand"
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
	case primebundle.PrimeBundle:
		if match.Type == "Share"{
			wiresMutex.Lock()
			wires[match.From] = match.Shares[0]
			wiresMutex.Unlock()
		} else if match.Type == "MultShare" {
			gateMutex.Lock()
			multMap[match.From] = match.Shares[0]
			gateMult[match.Gate] = multMap
			gateMutex.Unlock()
		} else if match.Type == "Result" {
			receivedResults[match.From] = match.Result
		} else {
			panic("Given type is unknown")
		}
	}
}

var function string
var wires = make(map[int]*big.Int)
var multMap = make(map[int]*big.Int)
var gateMult = make(map[int]map[int]*big.Int)
var wiresMutex = &sync.Mutex{}
var gateMutex = &sync.Mutex{}
var receivedResults = make(map[int]*big.Int)

func (s Shamir) SetFunction(f string) {
	function = f
}

func (s Shamir) ComputeShares(parties int, secret *big.Int) []*big.Int {
	// t should be less than half of connected parties t < 1/2 n
	var t = (parties - 1) / 2 //Integer division rounds down automatically
	//3 + 4x + 2x^2
	//[3, 4, 2]
	var polynomial = make([]*big.Int, t + 1)

	polynomial[0] = secret
	for i := 1; i < t + 1; i++ {
		//TODO Måske gøre så vi kan få error ud og tjekke på (fuck go)
		polynomial[i], _ = crand.Int(crand.Reader, field.GetSize())
	}

	var shares = make([]*big.Int, parties)

	for i := 1; i <= parties; i++ {
		shares[i - 1] = calculatePolynomial(polynomial, i)
	}

	return shares
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

func (s Shamir) TheOneRing(circuit Circuit.Circuit, sec int) int {
	partySize := network.GetParties()

	receiver := Receiver{}

	network.RegisterReceiver(receiver)
	secret := big.NewInt(int64(sec))
	result := 0

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
				var output *big.Int
				switch gate.Operation {
				case "Addition":
					//output = (input1 + input2) % field.GetSize()
					output = new(big.Int).Add(input1, input2)
					output.Mod(output, field.GetSize())
				case "Multiplication":
					interMult := new(big.Int).Mul(input1, input2)
					interMult.Mod(interMult, field.GetSize())
					multShares := s.ComputeShares(partySize, interMult)
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
					someint := Reconstruct(multMaaaaap)
					output = big.NewInt(int64(someint))
				}
				wiresMutex.Lock()
				wires[gate.GateNumber] = output
				wiresMutex.Unlock()
				circuit.Gates = removeGate(circuit, gate, i)
				if len(circuit.Gates) == 0 {
					distributeResult([]*big.Int{output}, partySize)
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

func calculatePolynomial(polynomial []*big.Int, x int) *big.Int {
	var result = big.NewInt(0)

	for i := 0; i < len(polynomial); i++ {
		//result += polynomial[i] * int(math.Pow(float64(x), float64(i)))
		iterres := new(big.Int).Exp(big.NewInt(int64(x)), big.NewInt(int64(i)), nil)
		iterres.Mul(iterres, polynomial[i])
		result.Add(result, iterres)
	}

	return result.Mod(result, field.GetSize())//result % field.GetSize()
}
func distributeMultShares(shares []*big.Int, partySize int, gate int) {
	for party := 1; party <= partySize; party++ {
		shareBundle := primebundle.PrimeBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "MultShare",
			Shares: []*big.Int{shares[party - 1]},
			From:   network.GetPartyNumber(),
			Gate: gate,
		}

		if network.GetPartyNumber() == party {
			gateMutex.Lock()
			multMap[party] = shares[party - 1]
			gateMult[gate] = multMap
			gateMutex.Unlock()
			//receivedShares = append(receivedShares, shareSlice...)
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeShares(shares []*big.Int, partySize int) {
	for party := 1; party <= partySize; party++ {
		shareBundle := primebundle.PrimeBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "Share",
			Shares: []*big.Int{shares[party - 1]},
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

func distributeResult(result []*big.Int, partySize int) {

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

func (s Shamir) ComputeFunction(shares map[int][]*big.Int, party int) []*big.Int {
	//Reconstruct(shares)
	if function == "add" {

	}
	return nil
}