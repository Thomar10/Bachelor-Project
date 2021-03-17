package Circuit

type Circuit struct {
	Gates []Gate `json:"Gate"`
}

type Gate struct {
	GateNumber int `json:"GateNumber"`
	Input_one int `json:"Input_one"`
	Input_two int `json:"Input_two"`
	Operation string `json:"Operation"`
}
