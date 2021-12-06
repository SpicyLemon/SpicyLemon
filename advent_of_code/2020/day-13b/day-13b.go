package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "sort"
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
    startingGuess, schedule, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("The input has %d entries.\n", len(schedule))
    fmt.Printf("Starting guess: [%d]. Schedule: %v.\n", startingGuess, schedule)
    busMap := reverseMap(schedule)
    fmt.Printf("Bus map: %v.\n", busMap)
    answer := findSolution(busMap, startingGuess)
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

func parseInput(input string) (uint64, map[uint64]uint64, error) {
    lines := strings.Split(input, "\n")
    start, err := strconv.ParseUint(lines[0], 10, 64)
    if err != nil {
        return 0, nil, err
    }
    busses := make(map[uint64]uint64)
    for i, bus := range strings.Split(lines[1], ",") {
        if bus != "x" {
            busVal, err := strconv.Atoi(bus)
            if err != nil {
                return start, nil, err
            }
            busses[uint64(i)] = uint64(busVal)
        }
    }
    return start, busses, nil
}

func findSolution(busMap map[uint64]uint64, startingGuess uint64) uint64 {
    busses := reverse(keys(busMap))
    slowBus := busses[0]
    otherBusses := busses[1:]
    m := startingGuess / slowBus
    keepGoing := true
    retval := uint64(0)
    for keepGoing {
        m += 1
        keepGoing = false
        retval = slowBus * m - busMap[slowBus]
        if m % 100000000 == 0 {
            fmt.Printf("m: [%d], checking [%d].\n", m, retval)
        }
        for _, bus := range otherBusses {
            if (retval + busMap[bus]) % bus != 0 {
                keepGoing = true
                break
            }
        }
    }
    return retval
}

func reverseMap(m map[uint64]uint64) map[uint64]uint64 {
    retval := make(map[uint64]uint64)
    for key, value := range m {
        retval[value] = key
    }
    return retval
}

func keys(m map[uint64]uint64) []uint64 {
    retval := make([]uint64, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    // sort.Slice(dirRange, func(i, j int) bool { return dirRange[i] < dirRange[j] })
    sort.Slice(retval, func(i, j int) bool { return retval[i] < retval[j] })
    return retval
}

func reverse(s []uint64) []uint64 {
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }
    return s
}
