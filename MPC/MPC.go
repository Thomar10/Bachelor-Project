package main

import (
	shamir "MPC/Secret-Sharing/Shamir"
	simple "MPC/Secret-Sharing/Simple-Sharing"
	"fmt"
)

func main() {
	fmt.Println(shamir.Test())
	fmt.Print(simple.SimpleSharing())
}
