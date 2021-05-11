package Secret_Sharing

import (
	"MPC/Circuit"
	finite "MPC/Finite-fields"
)

type Secret_Sharing interface {
	SetField(field finite.Finite)
	TheOneRing(circuit Circuit.Circuit, secret finite.Number, preprocessed bool, corrupts int) finite.Number
	ComputeShares(parties int, secret finite.Number) []finite.Number
	SetTriple(xMap, yMap, zMap map[int]finite.Number)
	RegisterReceiver()
	ResetSecretSharing()
}
