package Secret_Sharing

import finite "MPC/Finite-fields"

type Secret_Sharing interface {
	SetField(field finite.Finite)
	ComputeFunction(map[int][]int, int) []int
	SetFunction(f string)
	ComputeShares(parties, secret int) []int
	ComputeResult([]int) int
}