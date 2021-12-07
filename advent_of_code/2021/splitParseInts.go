package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = append(args, "1,2,3,4,5")
	}
	lineNumFmt := "%" + fmt.Sprintf("%d", len(fmt.Sprintf("%d", len(args)))) + "d"
	for i, arg := range args {
		lineNum := fmt.Sprintf(lineNumFmt, i+1)
		fmt.Printf("%s: input: \"%s\", ", lineNum, arg)
		ints, err := SplitParseInts(arg, ",")
		fmt.Printf("output: %v, error: %v\n", ints, err)
	}
}

// SplitParseInts splits a string using the given separator and converts each part into an int.
// Uses strings.Split(s, sep) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseInts(s string, sep string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Split(s, sep) {
		if len(entry) > 0 {
			i, err := strconv.Atoi(strings.TrimSpace(entry))
			if err != nil {
				return rv, err
			}
			rv = append(rv, i)
		}
	}
	return rv, nil
}
