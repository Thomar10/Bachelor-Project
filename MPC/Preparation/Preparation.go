package Preparation

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"MPC/Circuit"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	network "MPC/Network"
	secretsharing "MPC/Secret-Sharing"
	"MPC/Secret-Sharing/Shamir"
	"fmt"
	"math/big"
	"sync"

	"github.com/google/uuid"
)

type Preparation struct {
}

type Receiver struct {
}

var matrix [][]finite.Number

var x = make(map[int]finite.Number)
var y = make(map[int]finite.Number)
var z = make(map[int]finite.Number)
var zMutex = &sync.Mutex{}

var r = make(map[int]finite.Number)
var r2t = make(map[int]finite.Number)

var checkShareXMutex = &sync.Mutex{}
var checkShareYMutex = &sync.Mutex{}
var checkShareMapX = make(map[int]map[int]map[string][]finite.Number)
var checkShareMapY = make(map[int]map[int]map[string][]finite.Number)

var prepMutex = &sync.Mutex{}
var prepShares = make(map[int]map[string][]finite.Number)
var r2tShares = make(map[int]finite.Number)
var r2tMapMutex = &sync.Mutex{}
var r2tMap = make(map[int]map[int]finite.Number)
var r2tOpenMutex = &sync.Mutex{}
var r2tOpen = make(map[int]finite.Number)
var partySize int
var myPartyNumber int
var field finite.Finite

func (r Receiver) Receive(bundle bundle.Bundle) {
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "PrepShare" {
			prepMutex.Lock()
			randomMap := prepShares[match.Gate]
			if randomMap == nil {
				initPrepShares(match.Gate)
			}
			randomMap = prepShares[match.Gate]
			list := randomMap[match.Random]
			list[match.From-1] = match.Shares[0]
			randomMap[match.Random] = list
			prepShares[match.Gate] = randomMap
			prepMutex.Unlock()
		} else if match.Type == "R2TShare" {
			r2tMapMutex.Lock()
			r2tShares = r2tMap[match.Gate]
			if r2tShares == nil {
				r2tShares = make(map[int]finite.Number)
			}
			r2tShares[match.From] = match.Shares[0]
			r2tMap[match.Gate] = r2tShares
			r2tMapMutex.Unlock()
		} else if match.Type == "R2TResult" {
			r2tOpenMutex.Lock()
			r2tOpen[match.Gate] = match.Shares[0]
			r2tOpenMutex.Unlock()
		} else if match.Type == "CheckSharesXY" {
			checkShareXMutex.Lock()
			checkXMap := checkShareMapX[match.Gate]
			if checkXMap == nil {
				initCheckShares(match.Gate, checkShareMapX)
			}
			checkXMap = checkShareMapX[match.Gate]

			for i, v := range match.Shares {
				if i == partySize {
					break
				}
				checkStringMap := checkXMap[i+1]
				list := checkStringMap[match.Random]
				list[match.From-1] = v
				checkStringMap[match.Random] = list
				checkXMap[i+1] = checkStringMap
			}
			checkShareMapX[match.Gate] = checkXMap

			checkShareXMutex.Unlock()
			if len(match.Shares) > partySize {
				checkShareYMutex.Lock()
				checkYMap := checkShareMapY[match.Gate]
				if checkYMap == nil {
					initCheckShares(match.Gate, checkShareMapY)
				}
				checkYMap = checkShareMapY[match.Gate]
				for i, v := range match.Shares[partySize:] {
					checkStringMap := checkYMap[i+1]
					list := checkStringMap[match.Random]
					list[match.From-1] = v
					checkStringMap[match.Random] = list
					checkYMap[i+1] = checkStringMap
				}
				checkShareMapY[match.Gate] = checkYMap
				checkShareYMutex.Unlock()
			}
		} else if match.Type == "Panic" {
			panic("Someone tried to cheat in the protocol!")
		}
	}
}

//Reset Prep for testing
func ResetPrep() {
	x = make(map[int]finite.Number)
	y = make(map[int]finite.Number)
	z = make(map[int]finite.Number)
	r = make(map[int]finite.Number)
	r2t = make(map[int]finite.Number)

	prepShares = make(map[int]map[string][]finite.Number)
	r2tShares = make(map[int]finite.Number)
	r2tMap = make(map[int]map[int]finite.Number)
	r2tOpen = make(map[int]finite.Number)

}

//Initialize the CheckShares map
func initCheckShares(gate int, mapToInit map[int]map[int]map[string][]finite.Number) {
	randomMap := mapToInit[gate]
	if randomMap == nil {
		randomMap = make(map[int]map[string][]finite.Number)
	}
	for i := 1; i <= partySize; i++ {
		randomMap[i] = initCheckSharesString()
	}
	mapToInit[gate] = randomMap
}

//Initialize the map to have lists filled with 'garbage' values
//on the map entries x, y, r, r2t
func initCheckSharesString() map[string][]finite.Number {
	checkShareStringMap := make(map[string][]finite.Number)
	checkShareStringMap["x"] = listUnFilled(partySize)
	checkShareStringMap["y"] = listUnFilled(partySize)
	checkShareStringMap["r"] = listUnFilled(partySize)
	checkShareStringMap["r2t"] = listUnFilled(partySize)
	return checkShareStringMap
}

//Initialize the map for entry gate to have lists filled with 'garbage' values
//on the map entries x, y, r, r2t
func initPrepShares(gate int) {
	randomMap := prepShares[gate]
	if randomMap == nil {
		randomMap = make(map[string][]finite.Number)
	}
	randomMap["x"] = listUnFilled(partySize)
	randomMap["y"] = listUnFilled(partySize)
	randomMap["r"] = listUnFilled(partySize)
	randomMap["r2t"] = listUnFilled(partySize)
	prepShares[gate] = randomMap
}

//Register a receiver
func RegisterReceiver() {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)
}

//Creates all the triples needed to run the MPC-protocol with active or passive corrupt parties
func Prepare(circuit Circuit.Circuit, f finite.Finite, corrupts int, shamir secretsharing.Secret_Sharing, active bool) {
	field = f
	partySize = circuit.PartySize
	myPartyNumber = network.GetPartyNumber()

	createHyperMatrix(partySize)
	multiGates := countMultiGates(circuit)
	var wg sync.WaitGroup
	if active {
		triplesActive(multiGates, corrupts)
	} else {
		wg.Add(1)
		go triplesPassive(multiGates, corrupts, &wg)
	}
	wg.Wait()
	shamir.SetTriple(x, y, z)
}

//Computes the triples if the corrupts is active - therefore it also checks
//if any of the parties is cheating
func triplesActive(multiGates int, corrupts int) {
	for i := 1; i <= multiGates; i += corrupts * partySize * (partySize - corrupts - 1) {
		for j := 1; j <= partySize; j++ {
			//Gate can be seen as an unique identifier.
			//We simply call it gate to be consistent with naming
			gate := j + partySize*(i-1)
			random := createRandomNumber()
			yList := createRandomTuple(partySize, corrupts, gate, random, "y", true)
			random = createRandomNumber()
			xList := createRandomTuple(partySize, corrupts, gate, random, "x", true)
			random = createRandomNumber()
			rList := createRandomTuple(partySize, corrupts, gate, random, "r", true)
			r2tList := createRandomTuple(2*partySize, corrupts, gate, random, "r2t", true)
			//check if the lists are consistent
			checkConsistency(corrupts, gate, yList, "y", j)
			checkConsistency(corrupts, gate, xList, "x", j)
			checkConsistency(corrupts, gate, rList, "r", j)
			checkConsistency(corrupts, gate, r2tList, "r2t", j)
			//All the lists were consistent if we got to here. Put them into the correct maps from 1, ..., needed gates
			for k, _ := range yList[2*corrupts:] {
				inputPlace := k + (partySize-corrupts-1)*(j-1) + (i - 1) + 1
				y[inputPlace] = yList[k]
				x[inputPlace] = xList[k]
				r[inputPlace] = rList[k]
				r2t[inputPlace] = r2tList[k]
			}
			//No reason to create too many tuples
			if len(y) >= multiGates {
				break
			}
		}
	}
	//Calculate z in the triple
	computeZ()
}

//Checks if the shares are shared on a t-degree polynomial. If p <= corrupts it also checks for the result
//value of the y-list (list gotten from multiplying share onto the HIM)
func checkConsistency(corrupts int, gate int, yList []finite.Number, randomType string, p int) {
	//Distributes the shares as described in the protocol in the report
	if p <= 2*corrupts {
		prepMutex.Lock()
		xCheckShare := prepShares[gate][randomType]
		listToSend := append(xCheckShare, yList...)
		prepMutex.Unlock()
		distributeCheckShares(listToSend, p, gate, randomType)
	} else {
		prepMutex.Lock()
		xCheckShare := prepShares[gate][randomType]
		prepMutex.Unlock()
		distributeCheckShares(xCheckShare, p, gate, randomType)
	}
	//Check if the polynomial is consistent for x-shares and the result list of multiplication on the matrix (y-list)
	if myPartyNumber <= 2*corrupts && myPartyNumber == p {
		for {
			checkShareYMutex.Lock()
			consistencyCheckOnMap(corrupts, gate, randomType, checkShareMapY)
			checkShareYMutex.Unlock()
			break
		}
		//Check if the polynomial is consistent for x-shares (those checking y-list will also go in here)
	} else if p == myPartyNumber {
		for {
			checkShareXMutex.Lock()
			consistencyCheckOnMap(corrupts, gate, randomType, checkShareMapX)
			checkShareXMutex.Unlock()
			break
		}
	}
}

func consistencyCheckOnMap(corrupts int, gate int, randomType string, checkMap map[int]map[int]map[string][]finite.Number) {
	for _, shareList := range checkMap[gate] {
		if listFilledUp(shareList[randomType]) {
			reconstructMap := make(map[int]finite.Number)
			for i, v := range shareList[randomType] {
				reconstructMap[i+1] = v
			}
			var Polynomial []finite.Number
			if randomType == "r2t" {
				Polynomial = Shamir.ReconstructPolynomial(reconstructMap, 2*corrupts)
			} else {
				Polynomial = Shamir.ReconstructPolynomial(reconstructMap, corrupts)
			}
			for i, v := range shareList[randomType] {
				if !Shamir.ShareIsOnPolynomial(v, Polynomial, i+1) {
					distributePanic()
				}
			}
		}
	}
}

//Computes the triples for passive case
func triplesPassive(multiGates int, corrupts int, wg *sync.WaitGroup) {
	defer wg.Done()
	//Compute the values / triples we need to create our triple (x, y, z)
	for i := 1; i <= multiGates; i += partySize - corrupts {
		random := createRandomNumber()
		yList := createRandomTuple(partySize, corrupts, i, random, "y", false)
		random = createRandomNumber()
		xList := createRandomTuple(partySize, corrupts, i, random, "x", false)
		random = createRandomNumber()
		rList := createRandomTuple(partySize, corrupts, i, random, "r", false)
		r2tList := createRandomTuple(2*partySize, corrupts, i, random, "r2t", false)
		var wg sync.WaitGroup

		for j := range yList {
			wg.Add(1)
			y[j+i] = yList[j]
			x[j+i] = xList[j]
			r[j+i] = rList[j]
			r2t[j+i] = r2tList[j]
			fmt.Println("Calling newZ", j+i)
			go newcomputeZ(yList[j], xList[j], r2tList[j], rList[j], j+i, &wg)
		}
		wg.Wait()

	}

	//Calculate z in the triple
	//computeZ()
}
func newcomputeZ(y, x, r2t, r finite.Number, index int, wg *sync.WaitGroup) {
	defer wg.Done()
	xy := field.Mul(x, y)
	r2tInv := field.Mul(r2t, finite.Number{
		Prime:  big.NewInt(-1),
		Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
	})
	xyr2t := field.Add(xy, r2tInv)
	fmt.Println("Calling reconstruct for ", index)
	reconstructR2T(xyr2t, index)
	for {
		fmt.Println("locking r2topen ", index)
		r2tOpenMutex.Lock()
		xyr, found := r2tOpen[index]
		fmt.Println("unlocking r2topen ", index)
		r2tOpenMutex.Unlock()
		if found {
			resZ := field.Add(r, xyr)
			fmt.Println("Locking Z")
			zMutex.Lock()
			z[index] = resZ
			fmt.Println("Unlocking z")
			zMutex.Unlock()
			break
		}
	}
}

//Computes the last muskets of the triple (z)
func computeZ() {
	for i := 1; i <= len(y); i++ {
		xy := field.Mul(x[i], y[i])
		r2tInv := field.Mul(r2t[i], finite.Number{
			Prime:  big.NewInt(-1),
			Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
		})
		xyr2t := field.Add(xy, r2tInv)
		reconstructR2T(xyr2t, i)
		for {
			r2tOpenMutex.Lock()
			xyr, found := r2tOpen[i]
			r2tOpenMutex.Unlock()
			if found {
				resZ := field.Add(r[i], xyr)
				z[i] = resZ
				break
			}
		}
	}
}

//Reconstructs the polynomial R2T and thereafter distributes the
//open value of the R2T polynomial
func reconstructR2T(xyr2t finite.Number, i int) {
	distributeR2T(xyr2t, i, false)
	if (i%network.GetParties())+1 == myPartyNumber {
		for {
			r2tMapMutex.Lock()
			r2tMapLength := len(r2tMap[i])
			r2tMapMutex.Unlock()
			if r2tMapLength >= partySize {
				xyr := Shamir.Reconstruct(r2tMap[i])
				distributeR2T(xyr, i, true)
				break
			}
		}
	}
}

//Count the number of multiplication gates in the circuit as we
//dont want to create way to many triples for no reason
func countMultiGates(circuit Circuit.Circuit) int {
	result := 0
	for _, g := range circuit.Gates {
		if g.Operation == "Multiplication" {
			result++
		}
	}
	return result
}

//Fills the list with 'garbage' values so we can see when we have
//all the correct values
func listUnFilled(size int) []finite.Number {
	list := make([]finite.Number, size)
	for i := 0; i < size; i++ {
		list[i] = finite.Number{
			Prime:  big.NewInt(-1),
			Binary: []int{-1},
		}
	}
	return list
}

//Computes the shares for the random value and distributes the value accordingly
//Waits thereafter for having received all the shares to compute the random values from the matrix
func createRandomTuple(partySize int, corrupts int, i int, number finite.Number, randomType string, active bool) []finite.Number {
	randomShares := field.ComputeShares(partySize, number, corrupts)
	//randomShares[len(randomShares) - 1] = finite.Number{Prime: big.NewInt(10)}
	distributeShares(randomShares, i, randomType)
	for {
		isFilledUp := false
		prepMutex.Lock()
		isFilledUp = listFilledUp(prepShares[i][randomType])
		prepMutex.Unlock()
		if isFilledUp {
			break
		}
	}
	var xShares []finite.Number
	prepMutex.Lock()
	xShares = prepShares[i][randomType]
	prepMutex.Unlock()
	randomness := extractRandomness(xShares, matrix, corrupts, active)
	return randomness
}

//Checks if the list is filled up with correct values
func listFilledUp(list []finite.Number) bool {
	return field.FilledUp(list)
}

//Creates a random number
func createRandomNumber() finite.Number {
	return field.CreateRandomNumber()
}

//Create the Hyper invertible matrix as described in the report

func createHyperMatrix(size int) {
	a := make([]finite.Number, size)
	b := make([]finite.Number, size)
	//Compute distinct a and b values. a = 1, ..., partySize. b = partySize + 1, ... , 2 * partySize + 1
	for i := 1; i <= size; i++ {
		a[i-1] = finite.Number{
			Prime:  big.NewInt(int64(i)),
			Binary: Binary.ConvertXToByte(i),
		}
		b[i-1] = finite.Number{
			Prime:  big.NewInt(int64(i + size + 1)),
			Binary: Binary.ConvertXToByte(i + size + 1),
		}
	}
	//Initialize the matrix to be filled with 1 on all indexes
	matrix = make([][]finite.Number, size)
	for i, _ := range matrix {
		matrix[i] = make([]finite.Number, size)
		for j, _ := range matrix {
			matrix[i][j] = finite.Number{
				Prime:  big.NewInt(1),
				Binary: Binary.ConvertXToByte(1),
			}
		}
	}
	//Compute the matrix a described in the report
	for i, _ := range matrix {
		for j, _ := range matrix {
			for k := 0; k < size; k++ {
				if k == j {
					continue
				} else {
					//ak-neg
					ak := field.Add(a[k], finite.Number{
						Prime:  field.GetSize().Prime,
						Binary: Binary.ConvertXToByte(0),
					})
					biak := field.Add(b[i], ak)
					ajak := field.Add(a[j], ak)
					ajakInverse := field.FindInverse(ajak, field.GetSize())
					biakajak := field.Mul(biak, ajakInverse)
					matrix[i][j] = field.Mul(biakajak, matrix[i][j])
				}
			}
		}
	}
}

//Returns an array as the result of multiplying the vector of shares received onto the matrix
//If its for passive corrupts returns y_1, ..., y_n-t
//If its for active corrupts return y_1, ..., y_n
func extractRandomness(xVec []finite.Number, matrix [][]finite.Number, corrupts int, active bool) []finite.Number {
	ye := make([]finite.Number, len(xVec))
	for i := 0; i < len(matrix); i++ {
		ye[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
		for j := 0; j < len(matrix[i]); j++ {
			ye[i] = field.Add(ye[i], field.Mul(matrix[i][j], xVec[j]))
		}
	}
	if active {
		return ye
	} else {
		return ye[:len(xVec)-corrupts]
	}
}

//Distribute all the shares used to check if its a consistent polynomial
//If the length of shares is larger than party size its because the list
//also contains the y-list.
func distributeCheckShares(shares []finite.Number, party int, gate int, randomType string) {
	shareBundle := numberbundle.NumberBundle{
		ID:     uuid.Must(uuid.NewRandom()).String(),
		Type:   "CheckSharesXY",
		Shares: shares,
		From:   myPartyNumber,
		Gate:   gate,
		Random: randomType,
	}
	if myPartyNumber == party {
		checkShareXMutex.Lock()
		checkXMap := checkShareMapX[gate]
		if checkXMap == nil {
			initCheckShares(gate, checkShareMapX)
		}
		checkXMap = checkShareMapX[gate]
		for i, v := range shares {
			if i == partySize {
				break
			}
			checkStringMap := checkXMap[i+1]
			list := checkStringMap[randomType]
			list[myPartyNumber-1] = v
			checkStringMap[randomType] = list
			checkXMap[i+1] = checkStringMap
		}
		checkShareMapX[gate] = checkXMap

		checkShareXMutex.Unlock()

		if len(shares) > partySize {
			checkShareYMutex.Lock()
			checkYMap := checkShareMapY[gate]
			if checkYMap == nil {
				initCheckShares(gate, checkShareMapY)
			}
			checkYMap = checkShareMapY[gate]
			for i, v := range shares[partySize:] {
				checkStringMap := checkYMap[i+1]
				list := checkStringMap[randomType]
				list[myPartyNumber-1] = v
				checkStringMap[randomType] = list
				checkYMap[i+1] = checkStringMap
			}
			checkShareMapY[gate] = checkYMap
			checkShareYMutex.Unlock()
		}
	} else {
		network.Send(shareBundle, party)
	}
}

//Distributes the random shares used to compute the triples
func distributeShares(shares []finite.Number, gate int, randomType string) {
	for party := 1; party <= partySize; party++ {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "PrepShare",
			Shares: []finite.Number{shares[party-1]},
			From:   myPartyNumber,
			Gate:   gate,
			Random: randomType,
		}
		//Send to itself
		if myPartyNumber == party {
			prepMutex.Lock()
			randomMap := prepShares[gate]
			if randomMap == nil {
				initPrepShares(gate)
			}
			randomMap = prepShares[gate]
			list := randomMap[randomType]
			list[party-1] = shares[party-1]
			randomMap[randomType] = list
			prepShares[gate] = randomMap
			prepMutex.Unlock()
		} else {
			network.Send(shareBundle, party)
		}
	}
}

//Distribute the R2T shares and the open R2T value from reconstruction
func distributeR2T(share finite.Number, gate int, forAll bool) {
	if forAll {
		for party := 1; party <= partySize; party++ {
			shareBundle := numberbundle.NumberBundle{
				ID:     uuid.Must(uuid.NewRandom()).String(),
				Type:   "R2TResult",
				Shares: []finite.Number{share},
				From:   myPartyNumber,
				Gate:   gate,
			}
			//Send to itself
			if myPartyNumber == party {
				r2tOpenMutex.Lock()
				r2tOpen[gate] = share
				r2tOpenMutex.Unlock()
			} else {
				network.Send(shareBundle, party)
			}
		}
	} else {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "R2TShare",
			Shares: []finite.Number{share},
			From:   myPartyNumber,
			Gate:   gate,
		}
		//Send to itself
		if myPartyNumber == (gate%network.GetParties())+1 {
			r2tMapMutex.Lock()
			r2tShares = r2tMap[gate]
			if r2tShares == nil {
				r2tShares = make(map[int]finite.Number)
			}
			r2tShares[myPartyNumber] = share
			r2tMap[gate] = r2tShares
			r2tMapMutex.Unlock()
		} else {
			network.Send(shareBundle, (gate%network.GetParties())+1)
		}
	}
}
func distributePanic() {
	for party := 1; party <= network.GetParties(); party++ {
		shareBundle := numberbundle.NumberBundle{
			ID:   uuid.Must(uuid.NewRandom()).String(),
			Type: "Panic",
		}
		if network.GetPartyNumber() == party {
			//To nothing
		} else {
			network.Send(shareBundle, party)
		}
	}
	panic("Someone tried to cheat in the protocol!")
}
