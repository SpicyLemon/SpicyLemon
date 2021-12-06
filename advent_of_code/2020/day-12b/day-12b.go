package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type instruction struct {
    action byte
    value int
}

type ship struct {
    x int
    y int
    wx int
    wy int
}

func (this *ship) executeInstruction(inst instruction) {
    switch inst.action {
    case 'N':
        fmt.Printf("Moving waypoint NORTH %d from [%d, %d] to ", inst.value, this.wx, this.wy)
        this.wy += inst.value
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'S':
        fmt.Printf("Moving waypoint SOUTH %d from [%d, %d] to ", inst.value, this.wx, this.wy)
        this.wy -= inst.value
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'E':
        fmt.Printf("Moving waypoint EAST %d from [%d, %d] to ", inst.value, this.wx, this.wy)
        this.wx += inst.value
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'W':
        fmt.Printf("Moving waypoint WEST %d from [%d, %d] to ", inst.value, this.wx, this.wy)
        this.wx -= inst.value
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'L':
        fmt.Printf("Rotating waypoint LEFT %d degrees from [%d, %d] to ", inst.value, this.wx, this.wy)
        iter := inst.value / 90
        for i := 0; i < iter; i++ {
            tmpwx := this.wx
            this.wx = -1 * this.wy
            this.wy = tmpwx
        }
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'R':
        // TODO: Solve
        fmt.Printf("Rotating waypoint RIGHT %d degrees from [%d, %d] to ", inst.value, this.wx, this.wy)
        iter := inst.value / 90
        for i := 0; i < iter; i++ {
            tmpwx := this.wx
            this.wx = this.wy
            this.wy = -1 * tmpwx
        }
        fmt.Printf("[%d, %d].\n", this.wx, this.wy)
    case 'F':
        fmt.Printf("Traveling toward waypoint %d times (waypoint: [%d, %d]) from [%d, %d] to ", inst.value, this.wx, this.wy, this.x, this.y)
        this.x += this.wx * inst.value
        this.y += this.wy * inst.value
        fmt.Printf("[%d, %d].\n", this.x, this.y)
    }
}

func (this *ship) getManhattanDistance() int {
    return abs(this.x) + abs(this.y)
}

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
    fmt.Printf("Instructions: %v\n", input)
    ferry := ship{wx: 10, wy: 1}
    travel(&ferry, input)
    fmt.Printf("Ferry: %v\n", ferry)
    answer := ferry.getManhattanDistance()
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

func parseInput(input string) ([]instruction, error) {
    lines := strings.Split(input, "\n")
    retval := []instruction{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            act := line[0]
            valStr := line[1:]
            val, err := strconv.Atoi(valStr)
            if err != nil {
                return nil, err
            }
            retval = append(retval, instruction{action: act, value: val})
        }
    }
    return retval, nil
}

func travel(ferry *ship, instructions []instruction) {
    for _, inst := range instructions {
        ferry.executeInstruction(inst)
    }
}

func abs(val int) int {
    if val < 0 {
        return val * -1
    }
    return val
}
