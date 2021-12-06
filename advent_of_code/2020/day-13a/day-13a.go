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
    now, busses, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("The input has %d entries.\n", len(busses))
    fmt.Printf("Now: [%d]. Busses: %v.\n", now, busses)
    waitLength, forBus := waitForBusses(now, busses)
    fmt.Printf("In %d minutes (at %d) bus %d will arrive.\n", waitLength, now + waitLength, forBus)
    answer := waitLength * forBus
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

func parseInput(input string) (int, []int, error) {
    lines := strings.Split(input, "\n")
    now, err := strconv.Atoi(lines[0])
    if err != nil {
        return -1, nil, err
    }
    busses := []int{}
    for _, bus := range strings.Split(lines[1], ",") {
        if bus != "x" {
            busVal, err := strconv.Atoi(bus)
            if err != nil {
                return now, nil, err
            }
            busses = append(busses, busVal)
        }
    }
    return now, busses, nil
}

func waitForBusses(now int, busses []int) (int, int) {
    i := 0
    for {
        for _, bus := range busses  {
            if (now + i) % bus == 0 {
                return i, bus
            }
        }
        i += 1
    }
    return -1, -1
}
