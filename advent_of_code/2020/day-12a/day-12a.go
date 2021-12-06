package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

const directions = "NESW"

type instruction struct {
    action byte
    value int
}

type ship struct {
    x int
    y int
    facing byte
}

func (this *ship) executeInstruction(inst instruction) {
    switch inst.action {
    case 'N':
        fmt.Printf("Moving North %d from %d to ", inst.value, this.y)
        this.y -= inst.value
        fmt.Printf("%d.\n", this.y)
    case 'S':
        fmt.Printf("Moving South %d from %d to ", inst.value, this.y)
        this.y += inst.value
        fmt.Printf("%d.\n", this.y)
    case 'W':
        fmt.Printf("Moving West %d from %d to ", inst.value, this.x)
        this.x -= inst.value
        fmt.Printf("%d.\n", this.x)
    case 'E':
        fmt.Printf("Moving East %d from %d to ", inst.value, this.x)
        this.x += inst.value
        fmt.Printf("%d.\n", this.x)
    case 'L':
        fmt.Printf("Turning Left %d degrees from %s to ", inst.value, string(this.facing))
        this.facing = directions[(4 + getDirectionI(this.facing) - inst.value / 90) % 4]
        fmt.Printf("%s.\n", string(this.facing))
    case 'R':
        fmt.Printf("Turning Right %d degrees from %s to ", inst.value, string(this.facing))
        this.facing = directions[(getDirectionI(this.facing) + inst.value / 90) % 4]
        fmt.Printf("%s.\n", string(this.facing))
    case 'F':
        fmt.Printf("Moving forward %d - ", inst.value)
        this.executeInstruction(instruction{action: this.facing, value: inst.value})
    }
}

func getDirectionI(facing byte) int {
    for retval, d := range directions {
        if byte(d) == facing {
            return retval
        }
    }
    return -1
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
    ferry := ship{facing: 'E'}
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
