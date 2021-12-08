package main

import (
	"fmt"
	"os"
	"sort"
)

func main() {
	args := []string{
		"zyxwvutsrqponmlkjihgfedcbaZYXWVUTSRQPONMLKJIHGFEDCBA0123456789`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?",
		"aAbByYzZmMnN0123456789`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?",
		"zyxwvutsrqponmlkjihgfedcba", "azcybxdwevfugthsirjqkplonm", "mnolpkqjrishtgufvewdxbycza",
		"ywusqomkigeca", "zxvtrpnljhfdb",
		"abczyx", "xyzcba",
	}
	args = append(args, os.Args[1:]...)
	df := DigitFormatForMax(len(args))
	for i, arg := range args {
		sorted := SortString(arg)
		fmt.Printf(df+": %q => %q\n", i, arg, sorted)
	}
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d"
// DigitFormatForMax(382920) returns "%6d"
func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
}

type RuneSorter []rune

func (s RuneSorter) Less(i, j int) bool { return s[i] < s[j] }
func (s RuneSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s RuneSorter) Len() int           { return len(s) }

func SortString(str string) string {
	r := []rune(str)
	sort.Sort(RuneSorter(r))
	return string(r)
}
