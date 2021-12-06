package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename := getCliParams()
    fmt.Printf("Getting input from [%s].\n", filename)
    dat, err := ioutil.ReadFile(filename)
    if (err != nil) {
        return err
    }
    // fmt.Println(string(dat))
    input, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("The input has %d entries.\n", len(input))
    fmt.Printf("input: %v\n", input)
    answer := runGame(input, 2020)
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getCliParams() string {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input"
    case 1:
        return args[0]
    }
    panic(fmt.Errorf("Invalid command-line arguments: %s.", args))
}

func parseInput(input string) ([]int, error) {
    lines := strings.Split(input, "\n")
    retval := []int{}
    for _, n := range strings.Split(lines[0], ",") {
        v, err := strconv.Atoi(n)
        if err != nil {
            return nil, err
        }
        retval = append(retval, v)
    }
    return retval, nil
}

func initializeGame(nums []int) map[int][]int {
    retval := make(map[int][]int)
    for i, n := range nums {
        retval[n] = []int{i + 1}
        // fmt.Printf("Turn %4d: %4d - %v\n", i + 1, n, retval)
    }
    return retval
}

func runGame(nums []int, turns int) int {
    mem := initializeGame(nums)
    lastNum := nums[len(nums)-1]
    for turn := len(nums) + 1; turn <= turns; turn++ {
        thisNum := 0
        turnsLastSaid, ok := mem[lastNum]
        if ok && len(turnsLastSaid) >= 2 {
            thisNum = turnsLastSaid[len(turnsLastSaid)-1] - turnsLastSaid[len(turnsLastSaid)-2]
        }
        lastNum = thisNum
        mem[thisNum] = append(mem[thisNum], turn)
        // fmt.Printf("Turn %4d: %4d - %v\n", turn, thisNum, mem)
    }
    return lastNum
}
