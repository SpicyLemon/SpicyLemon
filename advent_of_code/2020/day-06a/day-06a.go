package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// This is just a wrapper for the shell script I wrote for this.
	// I'm just including it so my generic timing stuff will work.
	cmd := exec.Command("/Users/danielwedul/git/SpicyLemon/advent_of_code/2020/day-06a/day-06a.sh", os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
