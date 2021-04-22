package main

import (
	finite "MPC/Finite-fields"
	"math/big"
)


func main() {
	secretToTest := finite.Number{Prime: big.NewInt(5)}
	MPCTest("Circuit", secretToTest, "")
	MPCTest("Circuit", secretToTest, ":40445")
}