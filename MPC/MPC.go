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
	"time"

	"github.com/google/uuid"
)

type Receiver struct {

}

func (r Receiver) Receive(bundle bundle.Bundle) {
	fmt.Println("I have received bundle:", bundle)
}

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

	receiver := Receiver{}
	network.RegisterReceiver(receiver)
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
			if network.GetParties() == party_size - 1 {
				//TODO make better implementation where we don't have to sleep
				//Ie. make network gob decoding more robust
				time.Sleep(time.Second)
				for i := 0; i < party_size - 1; i++ {
					network.Send(bundle_type, i)
				}
				break
			}
		}
	}

	for{

	}
}
