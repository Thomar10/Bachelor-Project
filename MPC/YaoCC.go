package main

import (
	"MPC/Circuit"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	bitLengthString := os.Args[1]
	bitLength, _ := strconv.Atoi(bitLengthString)
	file, _ := json.MarshalIndent(createCircuit(bitLength), "", " ")
	_ = ioutil.WriteFile("YaoBits"+ bitLengthString +".json", file, 0644)

}

//The following code contains high "job security"!
func createCircuit(bits int) Circuit.Circuit {
	circuitCreated := Circuit.Circuit{
		Gates: createGates(bits),
	}
	return circuitCreated
}

func create2BitComparator(gateStart, input1, input2 int) ([]Circuit.Gate, int, int, int, []int) {
	var gates []Circuit.Gate
	var openGates []int
	//XOR-gate
	gates = append(gates, Circuit.Gate{
		GateNumber: gateStart,
		Input_one: input1,
		Input_two: input2,
		Operation: "Addition",
	})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: input1,
			Input_two: gateStart - 1,
			Operation: "Multiplication",
		})
	openGates = append(openGates, gateStart)
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: input2,
			Input_two: gateStart - 2,
			Operation: "Multiplication",
		})
	openGates = append(openGates, gateStart)
	gateStart++

	return gates, input1 + 1, input2 + 1, gateStart, openGates
}

func createGates(bits int) []Circuit.Gate {
	gateStartNumber := bits * 2 + 1
	party1Bit := 1
	party2Bit := bits + 1
	var gates []Circuit.Gate
	var openGates [][]int
	for i := 1; i <= bits; i++ {
		var gateBlock []Circuit.Gate
		var openGatesBlock []int
		gateBlock, party1Bit, party2Bit, gateStartNumber, openGatesBlock = create2BitComparator(gateStartNumber, party1Bit, party2Bit)
		gates = append(gates, gateBlock...)
		openGates = append(openGates, openGatesBlock)
	}
	if bits >= 2 {
		//Create blocking blocks
		for i := 1; i <= bits - 1; i++ {
			var gateBlock []Circuit.Gate
			var openGatesBlock []int
			//Take the first to open gates arrays out of openGates and replace them with one array
			gateBlock, gateStartNumber, openGatesBlock = createBlockingBlock(gateStartNumber, openGates[:2])
			gates = append(gates, gateBlock...)
			tempOpenGates := make([][]int, 1)
			tempOpenGates[0] = openGatesBlock
			openGates = append(tempOpenGates, openGates[2:]...)

		}
		gates = append(gates, createOutputGates(gateStartNumber, openGates[0])...) //TODO giv rigtige output gates med
	} else {
		gates = append(gates, createOutputGates(gateStartNumber, openGates[0])...)
	}
	return gates
}

func createOutputGates(gateStart int, openGates []int) []Circuit.Gate {
	var gates []Circuit.Gate
	for i:= 1; i <= 2; i++ {
		gates = append(gates,
			Circuit.Gate{
				GateNumber: gateStart,
				Input_one: openGates[i - 1],
				Output: i,
				Operation: "Output",
			})
		gateStart++
	}
	return gates
}

func createOrGate(gateStart, input1, input2 int) ([]Circuit.Gate, int) {
	var gates []Circuit.Gate
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: input1,
			Input_two: input2,
			Operation: "Multiplication",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: input1,
			Input_two: input2,
			Operation: "Addition",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gateStart - 1,
			Input_two: gateStart - 2,
			Operation: "Addition",
		})
	gateStart++
	return gates, gateStart
}
func createBlockingBlock(gateStart int, gatesToClose[][]int) ([]Circuit.Gate, int, []int) {
	//Gates to close contains an array of two arrays
	//First array could be the "OR" gates from another blocking block or AND gates from the inputs block
	//Second always contains AND gates to extend and blocking block on
	var gates []Circuit.Gate
	var openGates []int
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gatesToClose[0][0],
			Input_two: gatesToClose[0][1],
			Operation: "Addition",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gateStart - 1,
			Input_two: gatesToClose[1][0],
			Operation: "Addition",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gateStart - 2,
			Input_two: gatesToClose[1][1],
			Operation: "Addition",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gateStart - 2,
			Input_two: gatesToClose[1][0],
			Operation: "Multiplication",
		})
	gateStart++
	gates = append(gates,
		Circuit.Gate{
			GateNumber: gateStart,
			Input_one: gateStart - 2,
			Input_two: gatesToClose[1][1],
			Operation: "Multiplication",
		})
	gateStart++
	var orGates []Circuit.Gate
	orGates, gateStart = createOrGate(gateStart, gateStart - 2, gatesToClose[0][0])
	gates = append(gates, orGates...)
	openGates = append(openGates, gateStart - 1)
	orGates, gateStart = createOrGate(gateStart, gateStart - 4, gatesToClose[0][1])
	gates = append(gates, orGates...)
	openGates = append(openGates, gateStart - 1)
	gateStart++

	return gates, gateStart, openGates
}