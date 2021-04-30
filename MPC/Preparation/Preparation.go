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
	crand "crypto/rand"
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

var r = make(map[int]finite.Number)
var r2t = make(map[int]finite.Number)

var prepMutex = &sync.Mutex{}
var prepShares = make(map[int]map[string][]finite.Number)
var r2tShares = make(map[int]finite.Number)
var r2tMapMutex = &sync.Mutex{}
var r2tMap = make(map[int]map[int]finite.Number)
var r2tOpenMutex = &sync.Mutex{}
var r2tOpen = make(map[int]finite.Number)
var bundleCounter = 1

func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle prep:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "PrepShare" {
			//fmt.Println("I have received bundle prep:", bundle)
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
		}
	}
}

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
	bundleCounter = 1
}

func initPrepShares(gate int) {
	randomMap := prepShares[gate]
	if randomMap == nil {
		randomMap = make(map[string][]finite.Number)
	}
	randomMap["x"] = listUnFilled(network.GetParties())
	randomMap["y"] = listUnFilled(network.GetParties())
	randomMap["r"] = listUnFilled(network.GetParties())
	randomMap["r2t"] = listUnFilled(network.GetParties())
	prepShares[gate] = randomMap
}

func RegisterReceiver() {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)
}



func Prepare(circuit Circuit.Circuit, field finite.Finite, corrupts int, shamir secretsharing.Secret_Sharing, active bool) {
	partySize := circuit.PartySize
	createHyperMatrix(partySize, field)
	multiGates := countMultiGates(circuit)

	if active {
		triplesActive(multiGates, partySize, corrupts, field)
	}else {
		triplesPassive(multiGates, partySize, corrupts, field)
	}

	fmt.Println("The triples")
	fmt.Println("x", x)
	fmt.Println("y", y)
	fmt.Println("z", z)
	shamir.SetTriple(x, y, z)
}

func triplesActive(multiGates int, partySize int, corrupts int, field finite.Finite) {
	for i := 1; i <= multiGates; i += corrupts * partySize * (partySize - corrupts - 1) {
		for j := 1; j <= partySize; j++ {
			random := createRandomNumber(field)
			yList := createRandomTuple(partySize, field, corrupts, j + partySize * (i - 1), random, "y", true)
			random = createRandomNumber(field)
			xList := createRandomTuple(partySize, field, corrupts, i, random, "x", true)
			random = createRandomNumber(field)
			rList := createRandomTuple(partySize, field, corrupts, i, random, "r", true)
			r2tList := createRandomTuple(2*partySize, field, corrupts, i, random, "r2t", true)
			//check om listerne er konsistente

			//en liste er ikke consistent (dø)

			//alle er konsistente
			for k, _ := range yList[2 * corrupts:] {
				inputPlace :=  k + (partySize - corrupts - 1) * (j - 1) + (i - 1) + 1
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
		//No reason to create too many tuples
		if len(y) >= multiGates {
			break
		}

	}

	//Calculate z in the triple
	k := len(y)
	for i := 1; i <= k; i++ {
		xy := field.Mul(x[i], y[i])
		r2tInv := field.Mul(r2t[i], finite.Number{
			Prime:  big.NewInt(-1),
			Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
		})
		xyr2t := field.Add(xy, r2tInv)
		reconstructR2T(xyr2t, partySize, i)
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

func triplesPassive(multiGates int, partySize int, corrupts int, field finite.Finite) {
	for i := 1; i <= multiGates; i += partySize - corrupts {
		random := createRandomNumber(field)
		yList := createRandomTuple(partySize, field, corrupts, i, random, "y", false)
		random = createRandomNumber(field)
		xList := createRandomTuple(partySize, field, corrupts, i, random, "x", false)
		random = createRandomNumber(field)
		rList := createRandomTuple(partySize, field, corrupts, i, random, "r", false)
		r2tList := createRandomTuple(2*partySize, field, corrupts, i, random, "r2t", false)
		for j, _ := range yList {
			y[j+i] = yList[j]
			x[j+i] = xList[j]
			r[j+i] = rList[j]
			r2t[j+i] = r2tList[j]
		}
	}
	//Calculate z in the triple
	k := len(y)
	for i := 1; i <= k; i++ {
		xy := field.Mul(x[i], y[i])
		r2tInv := field.Mul(r2t[i], finite.Number{
			Prime:  big.NewInt(-1),
			Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
		})
		xyr2t := field.Add(xy, r2tInv)
		reconstructR2T(xyr2t, partySize, i)
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

func reconstructR2T(xyr2t finite.Number, partySize int, i int) {
	distributeR2T(xyr2t, partySize, i, false)
	if bundleCounter == network.GetPartyNumber() {
		for {
			r2tMapMutex.Lock()
			r2tMapLength := len(r2tMap[i])
			r2tMapMutex.Unlock()
			if r2tMapLength >= partySize {
				xyr := Shamir.Reconstruct(r2tMap[i])
				distributeR2T(xyr, partySize, i, true)
				break
			}
		}
	}
	bundleCounter++
	if bundleCounter > partySize {
		bundleCounter = 1
	}
}

func countMultiGates(circuit Circuit.Circuit) int {
	result := 0
	for _, g := range circuit.Gates {
		if g.Operation == "Multiplication" {
			result++
		}
	}
	return result
}

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

func createRandomTuple(partySize int, field finite.Finite, corrupts int, i int, number finite.Number, randomType string, active bool) []finite.Number {
	randomShares := field.ComputeShares(partySize, number)
	distributeShares(randomShares, network.GetParties(), i, randomType)
	for {
		isFilledUp := false
		prepMutex.Lock()
		isFilledUp = listFilledUp(prepShares[i][randomType], field)
		prepMutex.Unlock()
		if isFilledUp {
			break
		}
	}
	var xShares []finite.Number
	prepMutex.Lock()
	xShares = prepShares[i][randomType]
	prepMutex.Unlock()

	randomness := extractRandomness(xShares, matrix, field, corrupts, active)
	return randomness
}



func listFilledUp(list []finite.Number, field finite.Finite) bool {
	return field.FilledUp(list)
}

func createRandomNumber(field finite.Finite) finite.Number {
	randomNumber := finite.Number{}
	randomPrime, err := crand.Prime(crand.Reader, 32)
	if err != nil {
		panic("Unable to compute random number")
	}
	randomNumber.Prime = randomPrime
	randomNumber.Binary = Binary.CreateRandomByte()
	//Make sure random number is in the field
	randomNumber = field.Add(randomNumber, finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)})
	return randomNumber
}

func createHyperMatrix(partySize int, field finite.Finite) {
	a := make([]finite.Number, partySize)
	b := make([]finite.Number, partySize)
	for i := 1; i <= partySize; i++ {
		a[i-1] = finite.Number{
			Prime:  big.NewInt(int64(i)),
			Binary: Binary.ConvertXToByte(i),
		}
		b[i-1] = finite.Number{
			Prime:  big.NewInt(int64(i + partySize + 1)),
			Binary: Binary.ConvertXToByte(i + partySize + 1),
		}
	}
	matrix = make([][]finite.Number, partySize)
	for i, _ := range matrix {
		matrix[i] = make([]finite.Number, partySize)
		for j, _ := range matrix {
			matrix[i][j] = finite.Number{
				Prime:  big.NewInt(1),
				Binary: Binary.ConvertXToByte(1),
			}
		}
	}
	for i, _ := range matrix {
		for j, _ := range matrix {
			for k := 0; k < partySize; k++ {
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

func extractRandomness(xVec []finite.Number, matrix [][]finite.Number, field finite.Finite, corrupts int, active bool) []finite.Number {
	ye := make([]finite.Number, len(xVec))
	for i := 0; i < len(matrix); i++ {
		ye[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
		for j := 0; j < len(matrix[i]); j++ {
			ye[i] = field.Add(ye[i], field.Mul(matrix[i][j], xVec[j]))
			//y[i] = y[i] + matrix[i][j] * xVec[j]
		}
	}
	if active {
		return ye
	}else {
		return ye[:len(xVec) - corrupts]
	}
}

func distributeShares(shares []finite.Number, partySize int, gate int, randomType string) {
	for party := 1; party <= partySize; party++ {
		//fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "PrepShare",
			Shares: []finite.Number{shares[party-1]},
			From:   network.GetPartyNumber(),
			Gate:   gate,
			Random: randomType,
		}

		if network.GetPartyNumber() == party {
			//fmt.Println("Im sending", shareBundle, "to", party, "from", network.GetPartyNumber())
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
			//fmt.Println("Im sending", shareBundle, "to", party, "from", network.GetPartyNumber())
			network.Send(shareBundle, party)
		}
	}
}
func distributeR2T(share finite.Number, partySize int, gate int, forAll bool) {
	if forAll {
		for party := 1; party <= partySize; party++ {
			//fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
			shareBundle := numberbundle.NumberBundle{
				ID:     uuid.Must(uuid.NewRandom()).String(),
				Type:   "R2TResult",
				Shares: []finite.Number{share},
				From:   network.GetPartyNumber(),
				Gate:   gate,
			}

			if network.GetPartyNumber() == party {
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
			From:   network.GetPartyNumber(),
			Gate:   gate,
		}

		if network.GetPartyNumber() == bundleCounter {
			r2tMapMutex.Lock()
			r2tShares = r2tMap[gate]
			if r2tShares == nil {
				r2tShares = make(map[int]finite.Number)
			}
			r2tShares[network.GetPartyNumber()] = share
			r2tMap[gate] = r2tShares
			r2tMapMutex.Unlock()
		} else {
			network.Send(shareBundle, bundleCounter)
		}
	}

}
