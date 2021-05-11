package Simple_Sharing

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"reflect"
	"sort"
	"sync"
)

var field finite.Finite
var receivedShares =  make(map[int][]finite.Number)
var receivedMultShares =  make(map[int][]finite.Number)
var receivedResults []finite.Number
var rSharesMutex = &sync.Mutex{}
var rMultSharesMutex = &sync.Mutex{}
var rResultMutex = &sync.Mutex{}
type Simple_Sharing struct {}
type Receiver struct {}

func (r Receiver) Receive(bundle bundle.Bundle) {
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "Share" {
			rSharesMutex.Lock()
			receivedShares[match.From] = match.Shares
			rSharesMutex.Unlock()
		} else if match.Type == "Result" {
			rResultMutex.Lock()
			if !isNumberInList(match.Result, receivedResults) {
				receivedResults = append(receivedResults, match.Result)
			}
			rResultMutex.Unlock()
		} else if match.Type == "MultShare" {
			rMultSharesMutex.Lock()
			receivedMultShares[match.From] = match.Shares
			rMultSharesMutex.Unlock()
		}
	}
}

func isNumberInList(element finite.Number, list []finite.Number) bool {
	for _, r := range list {
		return field.CompareEqNumbers(element, r)
	}
	return false
}

func (s Simple_Sharing) ResetSecretSharing() {
	panic("lul")
}

func (s Simple_Sharing) RegisterReceiver() {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)
}

func (s Simple_Sharing) SetField(f finite.Finite) {
	field = f
}

func computeAdd(secret finite.Number, s secretsharing.Secret_Sharing) finite.Number {
	shares := s.ComputeShares(network.GetParties(), secret)
	distributeShares(shares, "Share")
	for{
		if network.GetParties() == len(receivedShares) {
			result := make([]finite.Number, len(receivedShares[1]))
			for i, _ := range result {
				result[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
			}
			for i := 0; i < len(receivedShares[1]); i++ {
				for _, share := range receivedShares {
					result[i] = field.Add(result[i], share[i])
				}
			}
			distributeResult(result)
			break
		}
	}

	for {
		if network.GetParties() == len(receivedResults) {
			finalResult := finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
			for _, r := range receivedResults {
				finalResult = field.Add(finalResult, r)
			}
			return finalResult
		}
	}
}

func (s Simple_Sharing) TheOneRing(circuit Circuit.Circuit, secret finite.Number, preprocessed bool, corrupts int) finite.Number {
	var result = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
	partyNumber := network.GetPartyNumber()
	partySize := network.GetParties()
	gate := circuit.Gates[0]

	switch gate.Operation {
	case "Addition":
		result = computeAdd(secret, s)
	case "Multiplication":
		shouldGiveInput := true
		if partyNumber != gate.Input_one && partyNumber != gate.Input_two {
			//This party should not participate
			shouldGiveInput = false
		}
		if shouldGiveInput {
			fmt.Println("I should share secret")
			shares := s.ComputeShares(partySize, secret)
			distributeShares(shares, "MultShare")
		}
		multResult := computeMul()
		fmt.Println("the multRes", multResult)
		//Add the u's together
		result = computeAdd(multResult, s)
	}
	return result
}

func computeMul() finite.Number {
	//Wait for enough shares
	for {
		if len(receivedMultShares) == 2 {
			break
		}
	}
	result := finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
	keys := reflect.ValueOf(receivedMultShares).MapKeys()
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}
	sort.Ints(keysArray)
	size := keysArray[0]
	//Everyone needs to have same party order
	party1 := keysArray[0]
	party2 := keysArray[1]
	i := network.GetPartyNumber() - 1
	for j := 0; j < len(receivedMultShares[size]); j++ {
		//Last party
		if i == len(receivedMultShares[size]) {
			interMult := field.Mul(receivedMultShares[party1][0],  receivedMultShares[party2][j])
			result = field.Add(result, interMult)
		} else {
			interMul := field.Mul(receivedMultShares[party1][i], receivedMultShares[party2][j])
			result = field.Add(result, interMul)
		}
	}
	if i-1 < 0 {
		interMul := field.Mul(receivedMultShares[party1][len(receivedMultShares[size])+(i-1)], receivedMultShares[party2][len(receivedMultShares[size])+(i-2)])
		result = field.Add(result, interMul)
	} else if i-2 < 0 {
		interMul := field.Mul(receivedMultShares[party1][(i-1)], receivedMultShares[party2][len(receivedMultShares[size])+(i-2)])
		result = field.Add(result, interMul)
	} else {
		interMul := field.Mul(receivedMultShares[party1][(i-1)], receivedMultShares[party2][(i-2)])
		result = field.Add(result, interMul)
	}
	return result
}

func (s Simple_Sharing) ComputeShares(parties int, secret finite.Number) []finite.Number {
	var shares []finite.Number
	lastShare := field.ConvertLastShare(secret)
	//Create the n - 1 random shares
	for sh := 1; sh < parties; sh++ {
		shares = append(shares, field.CreateRandomNumber())
	}
	//Create the nth share
	for _, share := range shares {
		inverseShare := field.Mul(finite.Number{Prime: big.NewInt(-1), Binary: Binary.ConvertXToByte(1)}, share)
		lastShare = field.Add(lastShare, inverseShare)
	}
	shares = append(shares, lastShare)
	return shares
}

func distributeShares(shares []finite.Number, shareType string) {
	for party := 1; party <= network.GetParties(); party++ {
		shareCopy := make([]finite.Number, len(shares))
		copy(shareCopy, shares)
		shareSlice := shareCopy[:party - 1]
		shareSlice2 := shareCopy[party:]
		shareSlice = append(shareSlice, shareSlice2...)
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   shareType,
			Shares: shareSlice,
			From:   network.GetPartyNumber(),
		}
		if network.GetPartyNumber() == party && shareType == "Share" {
			rSharesMutex.Lock()
			receivedShares[network.GetPartyNumber()] = shareSlice
			rSharesMutex.Unlock()
		} else if network.GetPartyNumber() == party && shareType == "MultShare" {
			rMultSharesMutex.Lock()
			receivedMultShares[network.GetPartyNumber()] = shareSlice
			rMultSharesMutex.Unlock()
		}else {
			network.Send(shareBundle, party)
		}

	}
}

func distributeResult(result []finite.Number) {
	counter := 0
	for party := 1; party <= network.GetParties(); party++ {
		if network.GetPartyNumber() != party {
			shareBundle := numberbundle.NumberBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Type: "Result",
				Result:  result[counter],
				From: network.GetPartyNumber(),
			}
			counter++
			network.Send(shareBundle, party)
		} else {
			rResultMutex.Lock()
			for _, e := range result {
				if !isNumberInList(e, receivedResults) {
					receivedResults = append(receivedResults, e)
				}
			}
			rResultMutex.Unlock()
		}
	}
}

func (s Simple_Sharing) SetTriple(xMap, yMap, zMap map[int]finite.Number) {
	//Will only be run by Shamir
}
