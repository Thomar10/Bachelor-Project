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
var prepShares = make(map[int][]finite.Number)
var r2tShares = make(map[int]finite.Number)
var r2tMapMutex = &sync.Mutex{}
var r2tMap = make(map[int]map[int]finite.Number)

func (r Receiver) Receive(bundle bundle.Bundle) {
	//fmt.Println("I have received bundle prep:", bundle)
	switch match := bundle.(type) {
	case numberbundle.NumberBundle:
		if match.Type == "PrepShare"{
			fmt.Println("The prepShare I got", match)
			prepMutex.Lock()
			list := prepShares[match.Gate]
			list[match.From - 1] = match.Shares[0]
			prepShares[match.Gate] = list
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
		}/* else {
			panic("Given type is unknown, got type: " +  match.Type)
		}*/
	}
}



func Prepare(circuit Circuit.Circuit, field finite.Finite, corrupts int, shamir secretsharing.Secret_Sharing) {
	receiver := Receiver{}
	network.RegisterReceiver(receiver)

	partySize := circuit.PartySize
	createHyperMatrix(partySize, field)
	multiGates := countMultiGates(circuit)
	//randomTriplesNeeded := multiGates / corrupts + 1
	//Calculate (x,y) in the (x, y, z) triple
	for i := 1; i <= multiGates; i += partySize - corrupts {
		prepMutex.Lock()
		prepShares[i] = listUnFilled(network.GetParties())
		prepMutex.Unlock()
	}

	for i := 1; i <= multiGates; i += partySize - corrupts {
		random := createRandomNumber(field)
		yList := createRandomTuple(partySize, field, corrupts, i, random)
		random = createRandomNumber(field)
		xList := createRandomTuple(partySize, field, corrupts, i, random)
		random = createRandomNumber(field)
		rList := createRandomTuple(partySize, field, corrupts, i, random)
		r2tList := createRandomTuple(2*partySize, field, corrupts, i, random)
		panic("Only one run")
		for j, _ := range yList {
			y[j + i] = yList[j]
			x[j + i] = xList[j]
			r[j + i] = rList[j]
			r2t[j + i] = r2tList[j]
		}

		//y = append(y, createRandomTuple(partySize, field, corrupts)...)
		//x = append(x, createRandomTuple(partySize, field, corrupts)...)

		//r = append(y, createRandomTuple(partySize, field, corrupts)...)
		//r2t = append(x, createRandomTuple(2*partySize, field, corrupts)...)
	}

	//Making sure there is enough y's
	for {
		if len(y) >= multiGates {
			break
		}
	}
	//Calculate z in the triple
	fmt.Println("X", x)
	fmt.Println("Y", y)
	k := len(y)
	for i := 1; i <= k; i++ {
		xy := field.Mul(x[i], y[i])
		r2tInv := field.Mul(r2t[i], finite.Number{
			Prime: big.NewInt(-1),
			Binary: []int{0, 0, 0, 0, 0, 0, 0, 1}, //-1
		})
		xyr2t := field.Add(xy, r2tInv)
		distributeR2T(xyr2t, partySize, i)
		for {
			r2tMapMutex.Lock()
			r2tMapLength := len(r2tMap[i])
			r2tMapMutex.Unlock()
			if r2tMapLength >= partySize {
				xyr := Shamir.Reconstruct(r2tMap[i])
				//fmt.Println("ab-r", xyr)
				//fmt.Println("r", r[i])
				//fmt.Println("r2t", r2t[i])
				//Add xyr to r to get xy=[z]_t
				resZ := field.Add(r[i], xyr)
				z[i] = resZ
				break
			}
		}

	}
	fmt.Println("Preparation is done!")
	fmt.Println("x", x)
	fmt.Println("y", y)
	fmt.Println("z", z)
	fmt.Println("r", r)
	fmt.Println("r2t", r2t)

	panic("im gay hehe")
	shamir.SetTriple(x, y, z)
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


func createRandomTuple(partySize int, field finite.Finite, corrupts int, i int, number finite.Number) []finite.Number  {
	randomShares := field.ComputeShares(partySize, number)
	distributeShares(randomShares, network.GetParties(), i)
	for {
		isFilledUp := false
		prepMutex.Lock()
		isFilledUp = listFilledUp(prepShares[i], field)
		prepMutex.Unlock()
		if isFilledUp {
			break
		}
	}
	var xShares []finite.Number
	prepMutex.Lock()
	xShares = prepShares[i]
	prepMutex.Unlock()

	fmt.Println("The random shares from all parties for ", i, xShares)
	return extractRandomness(xShares, matrix, field, corrupts)

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
	randomNumber = field.Add(randomNumber, finite.Number{Prime: big.NewInt(0)})
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
					//((b[i] - a[k]) / (a[j] - a[k]) * matrix[i][j]) % 17
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
	return ye//ye[:len(x) - corrupts]
}

func distributeShares(shares []finite.Number, partySize int, gate int) {
	fmt.Println("My random shares for ", gate , shares)
	fmt.Println("")
	for party := 1; party <= partySize; party++ {
		//fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "PrepShare",
			Shares: []finite.Number{shares[party-1]},
			From:   network.GetPartyNumber(),
			Gate: 	gate,
		}

		if network.GetPartyNumber() == party {
			prepMutex.Lock()
			list := prepShares[gate]
			list[party - 1] = shares[party-1]
			prepShares[gate] = list
			prepMutex.Unlock()
		} else {
			network.Send(shareBundle, party)
		}
	}
}
func distributeR2T(share finite.Number, partySize int, gate int) {
	for party := 1; party <= partySize; party++ {
		//fmt.Println("Im sending shares! Im party", network.GetPartyNumber())
		shareBundle := numberbundle.NumberBundle{
			ID:     uuid.Must(uuid.NewRandom()).String(),
			Type:   "R2TShare",
			Shares: []finite.Number{share},
			From:   network.GetPartyNumber(),
			Gate: 	gate,
		}

		if network.GetPartyNumber() == party {
			r2tMapMutex.Lock()
			r2tShares = r2tMap[gate]
			if r2tShares == nil {
				r2tShares = make(map[int]finite.Number)
			}
			r2tShares[network.GetPartyNumber()] = share
			r2tMap[gate] = r2tShares
			r2tMapMutex.Unlock()
		} else {
			network.Send(shareBundle, party)
		}
	}
}