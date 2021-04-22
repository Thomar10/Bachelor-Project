package Circuit

type Circuit struct {
	PartySize int `json:"PartySize"`
	Field string `json:"Field"`
	SecretSharing string `json:"SecretSharing"`
	Preprocessing bool `json:"Preprocessing"`
	Gates []Gate `json:"Gate"`
}

type Gate struct {
	GateNumber int `json:"GateNumber"`
	Input_one int `json:"Input_one"`
	Input_two int `json:"Input_two"`
	Input_constant int `json:"Input_constant"`
	Output int `json:"Output"`
	Operation string `json:"Operation"`
}
