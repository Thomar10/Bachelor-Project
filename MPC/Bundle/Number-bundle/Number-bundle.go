package Number_bundle

import (
	Finite_fields "MPC/Finite-fields"
)

type NumberBundle struct {
	ID string
	Type string
	Prime Finite_fields.Number
	Shares []Finite_fields.Number
	Result Finite_fields.Number
	From int
	Gate int
}
