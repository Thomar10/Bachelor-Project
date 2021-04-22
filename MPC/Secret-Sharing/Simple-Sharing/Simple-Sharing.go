package Simple_Sharing

import (
	"MPC/Bundle/Modules/Add"
	"MPC/Bundle/Modules/Multiplication"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	network "MPC/Network"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"sort"
)

var field finite.Finite
var function string

type Simple_Sharing struct {
}

func (s Simple_Sharing) ResetSecretSharing() {
	panic("lul")
}

func (s Simple_Sharing) RegisterReceiver() {
	panic("lul")
}

func (s Simple_Sharing) SetField(f finite.Finite) {
	field = f
}

func (s Simple_Sharing) ComputeFunction(shares map[int][]finite.Number, party int) []finite.Number {
	resultSize := len(shares[1])
	result := make([]*big.Int, resultSize)
	//result consists of <nil> - make it consist of big.Ints(0)
	for i, _ := range result {
		result[i] = big.NewInt(0)
	}
	if function == "Addition" {
		for i := 0; i < resultSize; i++ {
			for _, share := range shares {
				//result[i] += share[i]
				result[i].Add(result[i], share[i].Prime)
			}
			//result[i] = result[i] % field.GetSize()
			result[i].Mod(result[i], field.GetSize().Prime)
		}
	} else if function == "Multiplication" {
		keys := reflect.ValueOf(shares).MapKeys()
		var keysArray []int
		for _, k := range keys {
			keysArray = append(keysArray, (k.Interface()).(int))
		}
		sort.Ints(keysArray)
		size := keysArray[0]
		//Everyone needs to have same party order
		party1 := keysArray[0]
		party2 := keysArray[1]
		i := party - 1
		for j := 0; j < len(shares[size]); j++ {
			//Sidste party
			if i == len(shares[size]) {
				//result[0] += (shares[party1][0] * shares[party2][j]) % field.GetSize()
				mulBig := new(big.Int).Mul(shares[party1][0].Prime, shares[party2][j].Prime)
				modBig := new(big.Int).Mod(mulBig, field.GetSize().Prime)
				result[0].Add(result[0], modBig)
			} else {
				//result[0] += (shares[party1][i] * shares[party2][j]) % field.GetSize()
				mulBig := new(big.Int).Mul(shares[party1][i].Prime, shares[party2][j].Prime)
				modBig := new(big.Int).Mod(mulBig, field.GetSize().Prime)
				result[0].Add(result[0], modBig)
			}

		}
		if i-1 < 0 {
			//result[0] += (shares[party1][len(shares[size]) +  (i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
			mulBig := new(big.Int).Mul(shares[party1][len(shares[size])+(i-1)].Prime, shares[party2][len(shares[size])+(i-2)].Prime)
			modBig := new(big.Int).Mod(mulBig, field.GetSize().Prime)
			result[0].Add(result[0], modBig)
		} else if i-2 < 0 {
			//result[0] += (shares[party1][(i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
			mulBig := new(big.Int).Mul(shares[party1][(i-1)].Prime, shares[party2][len(shares[size])+(i-2)].Prime)
			modBig := new(big.Int).Mod(mulBig, field.GetSize().Prime)
			result[0].Add(result[0], modBig)
		} else {
			//result[0] += (shares[party1][(i - 1)] * shares[party2][(i - 2)]) % field.GetSize()
			mulBig := new(big.Int).Mul(shares[party1][(i-1)].Prime, shares[party2][(i-2)].Prime)
			modBig := new(big.Int).Mod(mulBig, field.GetSize().Prime)
			result[0].Add(result[0], modBig)
		}

		result[0].Mod(result[0], field.GetSize().Prime)
		r := result[0].Cmp(big.NewInt(0))
		if r < 0 {
			result[0].Add(field.GetSize().Prime, result[0])
			result[0].Mod(result[0], field.GetSize().Prime)
		}
	}
	var numberResult []finite.Number
	for _, share := range result {
		numberResult = append(numberResult, finite.Number{Prime: share})
	}
	return numberResult
}

func (s Simple_Sharing) SetFunction(f string) {
	function = f
}

func (s Simple_Sharing) SetTriple(xMap, yMap, zMap map[int]finite.Number) {
	panic("Omegalul")
}

func (s Simple_Sharing) TheOneRing(circuit Circuit.Circuit, secret finite.Number, preprocessed bool) finite.Number {
	var result = finite.Number{Prime: big.NewInt(0)}

	partyNumber := network.GetPartyNumber()
	partySize := network.GetParties()
	gate := circuit.Gates[0]

	switch gate.Operation {
	case "Addition":
		function = "Addition"
		result = Add.Add(secret, s, partySize)
	case "Multiplication":
		if partyNumber != gate.Input_one && partyNumber != gate.Input_two {
			//This party should not participate
			secret = finite.Number{Prime: big.NewInt(int64(-1))}
		}
		function = "Multiplication"
		multiplyResult := Multiplication.Multiply(secret, s, partySize)
		function = "Addition"
		result = Add.Add(multiplyResult, s, partySize)
	default:
		panic("Unknown operation")
	}

	return result
}

func (s Simple_Sharing) ComputeShares(parties int, secret finite.Number) []finite.Number {
	var prime = field.GetSize()
	var shares []*big.Int
	lastShare := secret.Prime
	//Create the n - 1 random shares
	for s := 1; s < parties; s++ {
		share, err := crand.Int(crand.Reader, prime.Prime)
		if err != nil {
			panic("Could not compute share!")
		}
		shares = append(shares, share)
	}
	//Create the nth share
	for _, share := range shares {
		lastShare.Sub(lastShare, share) //lastShare -= share
	}
	//Remove negative number
	// x cmp y
	// r = -1 if x < y
	// r = 0  if x = y
	// r = 1  if x > y
	r := lastShare.Cmp(big.NewInt(0))
	if r < 0 {
		//lastShare = prime + lastShare % prime
		lastShare.Add(prime.Prime, lastShare)
		lastShare.Mod(lastShare, prime.Prime)
	}
	//shares = append(shares, lastShare % prime)
	shares = append(shares, lastShare.Mod(lastShare, prime.Prime))
	numberShares := make([]finite.Number, len(shares))
	for i, share := range shares {
		numberShares[i] = finite.Number{Prime: share}
	}
	return numberShares
}

func (s Simple_Sharing) ComputeResult(results []finite.Number) finite.Number {
	result := big.NewInt(0)
	fmt.Println("results", results)
	for _, r := range results {
		//result += r
		result.Add(result, r.Prime)
	}
	return finite.Number{Prime: result.Mod(result, field.GetSize().Prime)} //result % field.GetSize()
}

//
//func (s Simple_Sharing) SetField(f finite.Finite) {
//	field = f
//}
//
//
//func (s Simple_Sharing) ComputeShares(parties int, secret *big.Int) []*big.Int {
//	rand.Seed(time.Now().UnixNano())
//	var prime = field.GetSize()
//	var shares []*big.Int
//	lastShare := secret
//	//Create the n - 1 random shares
//	for s := 1; s < parties; s++ {
//		share, err := crand.Int(crand.Reader, prime)
//		if err != nil {
//			panic("Could not compute share!")
//		}
//		shares = append(shares, share)
//	}
//	//Create the nth share
//	for _, share := range shares {
//		lastShare.Sub(lastShare, share) //lastShare -= share
//	}
//	//Remove negative number
//	// x cmp y
//	// r = -1 if x < y
//	// r = 0  if x = y
//	// r = 1  if x > y
//	r := lastShare.Cmp(big.NewInt(0))
//	if r < 0 {
//		//lastShare = prime + lastShare % prime
//		lastShare.Add(prime, lastShare)
//		lastShare.Mod(lastShare, prime)
//	}
//	//shares = append(shares, lastShare % prime)
//	shares = append(shares, lastShare.Mod(lastShare, prime))
//	return shares
//}
//
//
//func (s Simple_Sharing) SetFunction(f string) {
//	function = f
//}
//
//func (s Simple_Sharing) TheOneRing(circuit Circuit.Circuit, secret int) int {
//	result := 0
//
//	partyNumber := network.GetPartyNumber()
//	partySize := network.GetParties()
//	gate := circuit.Gates[0]
//
//	switch gate.Operation {
//		case "Addition":
//			function = "Addition"
//			result = Add.Add(big.NewInt(int64(secret)), s, partySize)
//		case "Multiplication":
//			if partyNumber != gate.Input_one && partyNumber != gate.Input_two {
//				//This party should not participate
//				secret = -1
//			}
//			function = "Multiplication"
//			multiplyResult := Multiplication.Multiply(big.NewInt(int64(secret)), s, partySize)
//			function = "Addition"
//			result = Add.Add(multiplyResult, s, partySize)
//		default:
//			panic("Unknown operation")
//	}
//
//	return result
//}
//
//func (s Simple_Sharing) ComputeFunction(shares map[int][]*big.Int, party int) []*big.Int {
//	resultSize := len(shares[1])
//	result := make([]*big.Int, resultSize)
//	//result consists of <nil> - make it consist of big.Ints(0)
//	for i, _ := range result {
//		result[i] = big.NewInt(0)
//	}
//	if function == "Addition" {
//		for i := 0; i < resultSize; i++ {
//			for _, share := range shares {
//				//result[i] += share[i]
//				result[i].Add(result[i], share[i])
//			}
//			//result[i] = result[i] % field.GetSize()
//			result[i].Mod(result[i], field.GetSize())
//		}
//	} else if function == "Multiplication" {
//		keys := reflect.ValueOf(shares).MapKeys()
//		var keysArray []int
//		for _, k := range keys {
//			keysArray = append(keysArray, (k.Interface()).(int))
//		}
//		sort.Ints(keysArray)
//		size := keysArray[0]
//		//Everyone needs to have same party order
//		party1 := keysArray[0]
//		party2 := keysArray[1]
//		i := party - 1
//		for j := 0; j < len(shares[size]); j++ {
//			//Sidste party
//			if i == len(shares[size]) {
//				//result[0] += (shares[party1][0] * shares[party2][j]) % field.GetSize()
//				mulBig := new(big.Int).Mul(shares[party1][0], shares[party2][j])
//				modBig := new(big.Int).Mod(mulBig, field.GetSize())
//				result[0].Add(result[0], modBig)
//			}else {
//				//result[0] += (shares[party1][i] * shares[party2][j]) % field.GetSize()
//				mulBig := new(big.Int).Mul(shares[party1][i], shares[party2][j])
//				modBig := new(big.Int).Mod(mulBig, field.GetSize())
//				result[0].Add(result[0], modBig)
//			}
//
//		}
//		if i - 1 < 0 {
//			//result[0] += (shares[party1][len(shares[size]) +  (i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
//			mulBig := new(big.Int).Mul(shares[party1][len(shares[size]) +  (i - 1)], shares[party2][len(shares[size]) + (i - 2)])
//			modBig := new(big.Int).Mod(mulBig, field.GetSize())
//			result[0].Add(result[0], modBig)
//		} else if i - 2 < 0 {
//			//result[0] += (shares[party1][(i - 1)] * shares[party2][len(shares[size]) + (i - 2)]) % field.GetSize()
//			mulBig := new(big.Int).Mul(shares[party1][(i - 1)], shares[party2][len(shares[size]) + (i - 2)])
//			modBig := new(big.Int).Mod(mulBig, field.GetSize())
//			result[0].Add(result[0], modBig)
//		} else {
//			//result[0] += (shares[party1][(i - 1)] * shares[party2][(i - 2)]) % field.GetSize()
//			mulBig := new(big.Int).Mul(shares[party1][(i - 1)], shares[party2][(i - 2)])
//			modBig := new(big.Int).Mod(mulBig, field.GetSize())
//			result[0].Add(result[0], modBig)
//		}
//
//		//result[0] = result[0] % field.GetSize()
//		result[0].Mod(result[0], field.GetSize())
///*		if result[0] < 0 {
//			result[0] = field.GetSize() + result[0]
//		}*/
//		r := result[0].Cmp(big.NewInt(0))
//		if r < 0 {
//			result[0].Add(field.GetSize(), result[0])
//			result[0].Mod(result[0], field.GetSize())
//		}
//		fmt.Println("result is ", result[0])
//	}
//	return result
//}
//
//func (s Simple_Sharing) ComputeResult(results []*big.Int) int {
//	result := big.NewInt(0)
//	for _, r := range results {
//		//result += r
//		result.Add(result, r)
//	}
//	return int(result.Mod(result, field.GetSize()).Int64())//result % field.GetSize()
//}
