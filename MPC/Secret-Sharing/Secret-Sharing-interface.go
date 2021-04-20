package Secret_Sharing

import (
	"MPC/Circuit"
	finite "MPC/Finite-fields"
)

type Secret_Sharing interface {
	SetField(field finite.Finite)
	ComputeFunction(map[int][]finite.Number, int) []finite.Number
	SetFunction(f string)
	TheOneRing(circuit Circuit.Circuit, secret finite.Number, preprocessed bool) finite.Number
	ComputeShares(parties int, secret finite.Number) []finite.Number
	ComputeResult([]finite.Number) finite.Number
	SetTriple(xMap, yMap, zMap map[int]finite.Number)
	RegisterReceiver()
}
