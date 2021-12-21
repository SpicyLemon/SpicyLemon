package main

import (
	"fmt"
)

func main() {
	cur := 1
	for n := 1; cur > 0; n++ {
		cur *= 7
		fmt.Printf("7^%-2d = %d\n", n, cur)
	}
}
