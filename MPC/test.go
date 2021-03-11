package main

import (
	"fmt"
	"math"
)

/*
Test fil til at teste go kode uden at køre hele programmet xd
 */
func main() {
	fmt.Println(math.Pow(2,3))
	fmt.Println(findInverse(-2, 11))
	/*
	a := []int{1, 2, 3}

	fmt.Println("Party 1 shares")
	fmt.Println(a[:0])
	fmt.Println(a[1:])

	fmt.Println("Party 2 shares")
	fmt.Println(a[:1])
	fmt.Println(a[2:])

	fmt.Println("Party 3 shares")
	fmt.Println(a[:2])
	fmt.Println(a[3:])
	*/

}

func findInverse(a int, prime int) int {
	if a < 0 {
		a = prime + a
	}
	return int(math.Pow(float64(a), float64(prime - 2))) % prime
}

//Itoh–Tsujii inversion algorithm
func findItohInverse(a int, prime int) int {
	var r = (prime - 1) / (prime - 1)
	var ar1 = int(math.Pow(float64(a), float64(r - 1))) % prime
	var ar = ar1 * a
	var arp1 = (1 / ar) % prime
	var a1 = arp1 * ar1
	return a1
}
