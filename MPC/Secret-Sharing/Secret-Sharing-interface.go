package Secret_Sharing

import (
	"MPC/Circuit"
	finite "MPC/Finite-fields"
)

type Secret_Sharing interface {
	SetField(field finite.Finite)
	ComputeFunction(map[int][]int, int) []int
	SetFunction(f string)
	TheOneRing(circuit Circuit.Circuit, secret int) int
	ComputeShares(parties, secret int) []int
	ComputeResult([]int) int
}