package Shamir

import (
	"math"
	"reflect"
)

func Reconstruct(shares map[int][]int) {
	keys := reflect.ValueOf(shares).MapKeys()
	var keysArray []int
	for _, k := range keys {
		keysArray = append(keysArray, (k.Interface()).(int))
	}

	for key := range keysArray {

	}
}

func computeDelta(key int, keys []int) {
	var talker = 1
	for j := range keys {
		if j == key {
			continue
		}

		talker *= (key - j)
	}

	var inverseTalker = findInverse(talker, field.GetSize())

}

func findInverse(a int, prime int) int {
	if a < 0 {
		a = prime + a
	}
	return int(math.Pow(float64(a), float64(prime - 2))) % prime
}

//Itoh–Tsujii inversion algorithm
/*
func findItohInverse(a int, prime int) int {
	var r = (prime - 1) / (prime - 1)
	var ar1 = int(math.Pow(float64(a), float64(r - 1))) % prime
	var ar = ar1 * a
	var arp1 = (1 / ar) % prime
	var a1 = arp1 * ar1
	return a1
}
 */
