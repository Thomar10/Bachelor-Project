package Circuit

type Circuit struct {
	Gates []Gate `json:"Circuit"`
}

type Gate struct {
	GateNumber int
	Input_one int
	Input_two int
	Operation string
	IntermediateRes int
}
