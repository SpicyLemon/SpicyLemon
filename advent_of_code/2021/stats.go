package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	var vals []int
	if len(args) == 0 {
		vals = []int{1, 1, 2, 3, 4, 5, 4, 8, 22, 3, 3, 3}
	} else {
		vals = make([]int, len(args))
		var err error
		for i, arg := range args {
			vals[i], err = strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return
			}
		}
	}
	fmt.Printf("GetStats(%v) =\n%s", vals, GetStats(vals))
}

// Stats contains statistical values on a collection of integers.
type Stats struct {
	Vals   []int
	Count  int
	Sum    int
	Min    int
	Max    int
	Ave    float64
	Median float64
	Counts map[int]int
	Mode   []int
}

// String outputs a multi-line string of this Stats struct.
func (s Stats) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("  Vals: %v\n", s.Vals))
	rv.WriteString(fmt.Sprintf(" Count: %d\n", s.Count))
	rv.WriteString(fmt.Sprintf("   Sum: %d\n", s.Sum))
	rv.WriteString(fmt.Sprintf("   Min: %d\n", s.Min))
	rv.WriteString(fmt.Sprintf("   Max: %d\n", s.Max))
	rv.WriteString(fmt.Sprintf("   Ave: %f\n", s.Ave))
	rv.WriteString(fmt.Sprintf("Median: %f\n", s.Median))
	rv.WriteString(fmt.Sprintf("Counts: %v\n", s.Counts))
	rv.WriteString(fmt.Sprintf("  Mode: %v\n", s.Mode))
	return rv.String()
}

// GetStats creates a Stats struct given a collection of integers.
func GetStats(ints []int) Stats {
	rv := Stats{}
	rv.Vals = append(rv.Vals, ints...)
	sort.Ints(rv.Vals)
	rv.Count = len(rv.Vals)
	rv.Counts = map[int]int{}
	if rv.Count > 0 {
		rv.Min = rv.Vals[0]
		rv.Max = rv.Vals[rv.Count-1]
		for _, v := range rv.Vals {
			rv.Sum += v
			rv.Counts[v]++
		}
		rv.Ave = float64(rv.Sum) / float64(rv.Count)
		if rv.Count%2 == 0 {
			rv.Median = float64(rv.Vals[rv.Count/2]+rv.Vals[rv.Count/2-1]) / 2
		} else {
			rv.Median = float64(rv.Vals[rv.Count/2])
		}
		var modeCount int
		for v, c := range rv.Counts {
			switch {
			case len(rv.Mode) == 0 || modeCount < c:
				rv.Mode = []int{v}
				modeCount = c
			case modeCount == c:
				rv.Mode = append(rv.Mode, v)
			}
		}
		sort.Ints(rv.Mode)
	}
	return rv
}
