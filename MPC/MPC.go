package main

import (
	bundle "MPC/Bundle"
	Prime_bundle "MPC/Bundle/Prime-bundle"
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"MPC/Finite-fields/Prime"
	network "MPC/Network"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)


var finite_field finite.Finite
var bundle_type bundle.Bundle
var party_size int

func main() {
	if os.Args[1] == "-p" {
		finite_field = Prime.Prime{}
		bundle_type = Prime_bundle.PrimeBundle{}
	}else {
		finite_field = Binary.Binary{}
		//TODO Add binary bundle
	}
	party_size, _ = strconv.Atoi(os.Args[2])


	isFirst := network.Init()
	if isFirst {
		finiteSize := finite_field.GenerateField()
		switch bundle_type.(type) {
		case Prime_bundle.PrimeBundle:
			bundle_type = Prime_bundle.PrimeBundle{
				ID: uuid.Must(uuid.NewRandom()).String(),
				Prime: finiteSize,
			}
		default:
			fmt.Println(":(")
		}
		for {
			if network.GetParties() == party_size {
				for i := 0; i <= party_size; i++ {
					network.Send(bundle_type, i)
				}
				break
			}
		}
	}

	for{

	}
}
