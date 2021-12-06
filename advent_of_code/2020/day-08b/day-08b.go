package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "strings"
    "strconv"
    "os"
)

type instruction struct {
    operation string
    argument int
}

func NewInstruction(op string, arg int) instruction {
    this := instruction{}
    this.operation = op
    this.argument = arg
    return this
}

func (this *instruction) copyOf() instruction {
    return NewInstruction(this.operation, this.argument)
}

type program struct {
    instructions []instruction
    linesRun []int
    accumulator int
}

func NewProgram() program {
    this := program{}
    this.instructions = []instruction{}
    this.linesRun = []int{}
    this.accumulator = 0
    return this
}

func (this *program) copyFresh() program {
    retval := NewProgram()
    for _, inst := range this.instructions {
        retval.instructions = append(retval.instructions, inst.copyOf())
    }
    return retval
}

func (this *program) copyProgramWithChange(lineNum int) program {
    retval := this.copyFresh()
    lineOp := retval.instructions[lineNum].operation
    switch lineOp {
    case "nop":
        fmt.Printf("Changing line [%d] from nop to jmp.\n", lineNum)
        retval.instructions[lineNum].operation = "jmp"
    case "jmp":
        fmt.Printf("Changing line [%d] from jmp to nop.\n", lineNum)
        retval.instructions[lineNum].operation = "nop"
    }
    return retval
}

func (this *program) addInstruction(op string, arg string) error {
    argInt, err := strconv.Atoi(arg)
    if err != nil {
        return err
    }
    this.instructions = append(this.instructions, NewInstruction(op, argInt))
    return nil
}

func (this *program) runInstruction(lineNum int, counter int) (int, error) {
    if contains(this.linesRun, lineNum) {
        return -1, fmt.Errorf("Line number [%d] has already been run. The accumulator is currently [%d].", lineNum, this.accumulator)
    }
    inst := this.instructions[lineNum]
    fmt.Printf("[%d]: Running line [%d]: %s %d\n", counter, lineNum, inst.operation, inst.argument)
    this.linesRun = append(this.linesRun, lineNum)
    switch inst.operation {
    case "nop":
        return lineNum + 1, nil
    case "jmp":
        return lineNum + inst.argument, nil
    case "acc":
        this.accumulator += inst.argument
        return lineNum + 1, nil
    }
    return -1, fmt.Errorf("Unknown instruction operation [%s] on line [%d].", inst.operation, lineNum)
}

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename := getInputFilename()
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
    answer := -1
    targetLine := len(input.instructions)
    fmt.Printf("Target line: [%d].\n", targetLine)
    for changeLine := targetLine - 1; changeLine >= 0; changeLine-- {
        if input.instructions[changeLine].operation != "acc" {
            prog := input.copyProgramWithChange(changeLine)
            nextLine := 0
            counter := 0
            err = nil
            for err == nil && nextLine != targetLine {
                counter += 1
                nextLine, err = prog.runInstruction(nextLine, counter)
            }
            if nextLine == targetLine {
                answer = prog.accumulator
                break
            }
        } else {
            fmt.Printf("Line [%d] is an acc line. Moving on.\n", changeLine)
        }
    }
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getInputFilename() string {
    args := os.Args[1:]
    if len(args) == 0 {
        return "example.input"
    }
    return args[0]
}

// first string is outer color. Second is inner color. int is required count.
func parseInput(input string) (program, error) {
    lines := strings.Split(input, "\n")
    retval := NewProgram()
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    lineRegex := regexp.MustCompile("(nop|acc|jmp)[[:space:]]+((?:\\+|-)?[[:digit:]]+)")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            lineMatch := lineRegex.FindStringSubmatch(line)
            if len(lineMatch) != 3 {
                return retval, fmt.Errorf("Could not parse line [%s].", line)
            }
            err := retval.addInstruction(lineMatch[1], lineMatch[2])
            if err != nil {
                return retval, err
            }
        }
    }
    return retval, nil
}

func contains(array []int, value int) bool {
    for _, arrayVal := range array {
        if arrayVal == value {
            return true
        }
    }
    return false
}
