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
	"github.com/google/uuid"
	"math/big"
	"sync"
	"time"
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

var xList = make(map[int][]finite.Number)
var yList = make(map[int][]finite.Number)
var rList = make(map[int][]finite.Number)
var r2tList = make(map[int][]finite.Number)



var prepMutex = &sync.Mutex{}
var xMutex = &sync.Mutex{}
var yMutex = &sync.Mutex{}
var zMutex = &sync.Mutex{}
var rMutex = &sync.Mutex{}
var r2tMutex = &sync.Mutex{}

var xListMutex = &sync.Mutex{}
var yListMutex = &sync.Mutex{}
var rListMutex = &sync.Mutex{}
var r2tListMutex = &sync.Mutex{}


var prepShares = make(map[int]map[string][]finite.Number)
var r2tShares = make(map[int]finite.Number)
var r2tMapMutex = &sync.Mutex{}
var r2tMap = make(map[int]map[int]finite.Number)
var r2tOpenMutex = &sync.Mutex{}
var r2tOpen = make(map[int]finite.Number)


func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle prep:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "PrepShare"{
			prepMutex.Lock()
			randomMap := prepShares[match.Gate]
			if randomMap == nil {
				randomMap = make(map[string][]finite.Number)
			}
			list := randomMap[match.Random]
			listLen := len(list)
			prepMutex.Unlock()
			if listLen == 0 {
				go initPrepShares(match.Gate)
			}
			prepMutex.Lock()
			//Tager den ud igen fordi jeg er nød til at unlock inden funktionskaldet
			//Grundet initPrep vil have en lock på samme mutex
			list = randomMap[match.Random]
			list[match.From - 1] = match.Shares[0]
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

func initPrepShares(gate int) {
	prepMutex.Lock()
	randomMap := prepShares[gate]
	if randomMap == nil {
		randomMap = make(map[string][]finite.Number)
	}
	randomMap["x"] = listUnFilled(network.GetParties())
	randomMap["y"] = listUnFilled(network.GetParties())
	randomMap["r"] = listUnFilled(network.GetParties())
	randomMap["r2t"] = listUnFilled(network.GetParties())
	prepShares[gate] = randomMap
	prepMutex.Unlock()
}

func Prepare(circuit Circuit.Circuit, field finite.Finite, corrupts int, shamir secretsharing.Secret_Sharing) {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)

	partySize := circuit.PartySize
	multiGates := countMultiGates(circuit)
	k := 0
	//randomTriplesNeeded := multiGates / corrupts + 1
	for i := 1; i <= multiGates; i += partySize - corrupts {
		go initPrepShares(i)
		k++
	}

	createHyperMatrix(partySize, field)


	//Calculate (x,y) in the (x, y, z) triple


	for i := 1; i <= multiGates; i += partySize - corrupts {
		go generateTripleValues(field, partySize, corrupts, i)
	}


	//Calculate z in the triple
	for i := 1; i <= k * (partySize - corrupts); i++ {
		go calculateZ(field, i, partySize)
	}

	fmt.Println("Waiting for triple")
	for {
		zMutex.Lock()
		zLen := len(z)
		zMutex.Unlock()
		if zLen >= k * (partySize - corrupts) {
			break
		}
	}
	fmt.Println("Got all triples")
	shamir.SetTriple(x, y, z)
}

func calculateZ(field finite.Finite, i int, partySize int) {
	startTimee := time.Now()
	var xValue, yValue finite.Number
	for {
		xMutex.Lock()
		xVal, foundX := x[i]
		xMutex.Unlock()
		yMutex.Lock()
		yVal, foundY := y[i]
		yMutex.Unlock()
		if foundX && foundY {
			xValue = xVal
			yValue = yVal
			break
		}
	}

	xy := field.Mul(xValue, yValue)
	var r2tValue finite.Number
	for {
		r2tMutex.Lock()
		r2tVal, found := r2t[i]
		fmt.Println("found r2tVal", found)
		r2tMutex.Unlock()
		if found {
			r2tValue = r2tVal
			break
		}
	}

	r2tInv := field.Mul(r2tValue, finite.Number{
		Prime:  big.NewInt(-1),
		Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
	})
	xyr2t := field.Add(xy, r2tInv)
	reconstructR2T(xyr2t, partySize, i)
	fmt.Println("have reconstructed xyr2t")
	for {
		r2tOpenMutex.Lock()
		xyr, found := r2tOpen[i]
		r2tOpenMutex.Unlock()
		rMutex.Lock()
		rValue, foundR := r[i]
		rMutex.Unlock()
		if found && foundR {
			resZ := field.Add(rValue, xyr)
			zMutex.Lock()
			z[i] = resZ
			zMutex.Unlock()
			break
		}
	}
	fmt.Println("It took", time.Since(startTimee), "to calculate Z")
}

func generateTripleValues(field finite.Finite, partySize int, corrupts int, i int) {
	startTime := time.Now()
	var wg sync.WaitGroup
	random := createRandomNumber(field)
	wg.Add(1)
	go createRandomTuple(partySize, field, corrupts, i, random, "y", &wg)
	random = createRandomNumber(field)
	wg.Add(1)
	go createRandomTuple(partySize, field, corrupts, i, random, "x", &wg)
	random = createRandomNumber(field)
	wg.Add(1)
	go createRandomTuple(partySize, field, corrupts, i, random, "r", &wg)
	wg.Add(1)
	go createRandomTuple(2*partySize, field, corrupts, i, random, "r2t", &wg)
	wg.Wait()

	yListMutex.Lock()
	yListLen := len(yList[i])
	yListMutex.Unlock()
	for j := 0; j < yListLen; j ++ {
		yListMutex.Lock()
		yListValue := yList[i][j]
		yListMutex.Unlock()
		yMutex.Lock()
		y[j+i] = yListValue
		yMutex.Unlock()

		xListMutex.Lock()
		xListValue := xList[i][j]
		xListMutex.Unlock()
		xMutex.Lock()
		x[j+i] = xListValue
		xMutex.Unlock()

		rListMutex.Lock()
		rListValue := rList[i][j]
		rListMutex.Unlock()
		rMutex.Lock()
		r[j+i] = rListValue
		rMutex.Unlock()

		r2tListMutex.Lock()
		r2tListValue := r2tList[i][j]
		r2tListMutex.Unlock()
		r2tMutex.Lock()
		r2t[j+i] = r2tListValue
		r2tMutex.Unlock()
	}
	fmt.Println("It took me", time.Since(startTime))
}

func reconstructR2T(xyr2t finite.Number, partySize int, i int) {
	distributeR2T(xyr2t, partySize, i, false)
	whosTurn := i % partySize + 1
	if whosTurn == network.GetPartyNumber()  {
		for {
			r2tMapMutex.Lock()
			r2tMapLength := len(r2tMap[i])
			r2tMapValue := r2tMap[i]
			r2tMapMutex.Unlock()
			if r2tMapLength >= partySize {
				xyr := Shamir.Reconstruct(r2tMapValue)
				distributeR2T(xyr, partySize, i, true)
				break
			}
		}
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

func listUnFilled(size int) []finite.Number{
	list := make([]finite.Number, size)
	for i := 0; i < size; i++ {
		list[i] = finite.Number{
			Prime: big.NewInt(-1),
			Binary: []int{-1},
		}
	}
	return list
}


func createRandomTuple(partySize int, field finite.Finite, corrupts int, i int, number finite.Number, randomType string, wg *sync.WaitGroup)   {
	defer wg.Done()
	startTime := time.Now()
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
	if randomType == "y" {
		yListMutex.Lock()
		yList[i] = extractRandomness(xShares, matrix, field, corrupts)
		yListMutex.Unlock()
	} else if randomType == "x" {
		xListMutex.Lock()
		xList[i] = extractRandomness(xShares, matrix, field, corrupts)
		xListMutex.Unlock()
	} else if randomType == "r" {
		rListMutex.Lock()
		rList[i] = extractRandomness(xShares, matrix, field, corrupts)
		rListMutex.Unlock()
	} else {
		r2tListMutex.Lock()
		r2tList[i] = extractRandomness(xShares, matrix, field, corrupts)
		r2tListMutex.Unlock()
	}
	fmt.Println("It took me ", time.Since(startTime), "to calculate randomTuple for", randomType)

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
		a[i - 1] = finite.Number{
			Prime: big.NewInt(int64(i)),
			Binary: Binary.ConvertXToByte(i),
		}
		b[i - 1] = finite.Number{
			Prime: big.NewInt(int64(i + partySize + 1)),
			Binary: Binary.ConvertXToByte(i + partySize + 1),
		}
	}
	matrix = make([][]finite.Number, partySize)
	for i, _ := range matrix {
		matrix[i] = make([]finite.Number, partySize)
		for j, _ := range matrix {
			matrix[i][j] = finite.Number{
				Prime: big.NewInt( 1),
				Binary: Binary.ConvertXToByte(1),
			}
		}
	}
	for i, _ := range matrix {
		for j, _ := range matrix {
			for k := 0; k < partySize; k++ {
				if k == j {
					continue
				}else {
					//ak-neg
					ak := field.Add(a[k], finite.Number{
						Prime: field.GetSize().Prime,
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

func extractRandomness(x []finite.Number, matrix [][]finite.Number, field finite.Finite, corrupts int) []finite.Number{
	ye := make([]finite.Number, len(x))
	for i := 0; i < len(matrix); i++ {
		ye[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
		for j := 0; j < len(matrix[i]); j++ {
			ye[i] = field.Add(ye[i], field.Mul(matrix[i][j], x[j]))
			//y[i] = y[i] + matrix[i][j] * x[j]
		}
	}
	return ye[:len(x) - corrupts]
}

func distributeShares(shares []finite.Number, partySize int, gate int, randomType string) {
	for party := 1; party <= partySize; party++ {
		//fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "PrepShare",
			Shares: []finite.Number{shares[party-1]},
			From:   network.GetPartyNumber(),
			Gate: 	gate,
			Random: randomType,
		}

		if network.GetPartyNumber() == party {
			prepMutex.Lock()
			randomMap := prepShares[gate]
			if randomMap == nil {
				randomMap = make(map[string][]finite.Number)
			}
			list := randomMap[randomType]
			list[party - 1] = shares[party-1]
			randomMap[randomType] = list
			prepShares[gate] = randomMap
			prepMutex.Unlock()
		} else {
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
				Gate: 	gate,
			}

			if network.GetPartyNumber() == party {
				r2tOpenMutex.Lock()
				r2tOpen[gate] = share
				r2tOpenMutex.Unlock()
			} else {
				network.Send(shareBundle, party)
			}
		}
	}else {
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "R2TShare",
			Shares: []finite.Number{share},
			From:   network.GetPartyNumber(),
			Gate: 	gate,
		}
		whosTurn := gate % partySize + 1
		if network.GetPartyNumber() == whosTurn {
			r2tMapMutex.Lock()
			r2tShares = r2tMap[gate]
			if r2tShares == nil {
				r2tShares = make(map[int]finite.Number)
			}
			r2tShares[network.GetPartyNumber()] = share
			r2tMap[gate] = r2tShares
			r2tMapMutex.Unlock()
		} else {
			network.Send(shareBundle, whosTurn)
		}
	}

}