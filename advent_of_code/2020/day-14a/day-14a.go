package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

type instruction struct {
    addr int
    value uint64
    maskedValue uint64
}

type instructionGroup struct {
    maskStr string
    maskOrStr string
    maskOr uint64
    maskAndNotStr string
    maskAndNot uint64
    instructions []instruction
}

func (this *instructionGroup) addInstruction(addr int, value uint64) {
    this.instructions = append(this.instructions, instruction{addr: addr, value: value, maskedValue: maskValue(value, this.maskOr, this.maskAndNot)})
}

func (this *instructionGroup) runInstructions(mem map[int]uint64) {
    for _, inst := range this.instructions {
        mem[inst.addr] = maskValue(inst.value, this.maskOr, this.maskAndNot)
    }
}

func (this *instruction) String() string {
    return fmt.Sprintf("mem[%d] = %d -> %d", this.addr, this.value, this.maskedValue)
}

func (this *instructionGroup) String() string {
    var sb strings.Builder
    fmt.Fprintf(&sb, "        Mask: %s\n", this.maskStr)
    fmt.Fprintf(&sb, "     Mask or: %s = %d\n", this.maskOrStr, this.maskOr)
    fmt.Fprintf(&sb, "Mask and not: %s = %d\n", this.maskAndNotStr, this.maskAndNot)
    fmt.Fprintf(&sb, "Instructions: [\n")
    for i, inst := range this.instructions {
        fmt.Fprintf(&sb, "  %d: %s\n", i, inst.String())
    }
    fmt.Fprintf(&sb, "]")
    return sb.String()
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
    fmt.Printf("input:\n%s\n", instructionGroupsString(input))
    mem := make(map[int]uint64)
    runAllInstructions(mem, input)
    fmt.Printf("Mem: %s\n", memString(mem))
    answer := sumMem(mem)
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

func parseInput(input string) ([]instructionGroup, error) {
    lines := strings.Split(input, "\n")
    retval := []instructionGroup{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    maskLineRegex := regexp.MustCompile("^mask[[:space:]]*=[[:space:]]*([X01]+)")
    memLineRegex := regexp.MustCompile("^mem\\[([[:digit:]]+)\\][[:space:]]*=[[:space:]]*([[:digit:]]+)")
    xRegex := regexp.MustCompile("X")
    zeroRegex := regexp.MustCompile("0")
    oneRegex := regexp.MustCompile("1")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            maskMatch := maskLineRegex.FindStringSubmatch(line)
            if maskMatch != nil {
                mask := maskMatch[1]
                orMaskStr := xRegex.ReplaceAllString(zeroRegex.ReplaceAllString(mask, "X"), "0")
                andNotMaskStr := xRegex.ReplaceAllString(zeroRegex.ReplaceAllString(oneRegex.ReplaceAllString(mask, "X"), "1"), "0")
                orMask, err := strconv.ParseUint(orMaskStr, 2, 64)
                if (err != nil) {
                    return nil, err
                }
                andNotMask, err := strconv.ParseUint(andNotMaskStr, 2, 64)
                if (err != nil) {
                    return nil, err
                }
                retval = append(retval, instructionGroup{maskStr: mask, maskOr: orMask, maskAndNot: andNotMask, maskOrStr: orMaskStr, maskAndNotStr: andNotMaskStr})
            } else {
                memMatch := memLineRegex.FindStringSubmatch(line)
                if memMatch == nil {
                    return nil, fmt.Errorf("Could not parse line [%s].\n", line)
                }
                addr, err := strconv.Atoi(memMatch[1])
                if err != nil {
                    return nil, err
                }
                value, err := strconv.ParseUint(memMatch[2], 10, 64)
                if err != nil {
                    return nil, err
                }
                retval[len(retval)-1].addInstruction(addr, value)
            }
        }
    }
    return retval, nil
}

func instructionGroupsString(input []instructionGroup) string {
    var sb strings.Builder
    for _, instg := range input {
        fmt.Fprintf(&sb, "%s\n", instg.String())
    }
    return sb.String()
}

func maskValue(value uint64, maskOr uint64, maskAndNot uint64) uint64 {
    return (value | maskOr) &^ maskAndNot
}

func runAllInstructions(mem map[int]uint64, instructionGroups []instructionGroup) {
    for _, instg := range instructionGroups {
        instg.runInstructions(mem)
    }
}

func sumMem(mem map[int]uint64) uint64 {
    retval := uint64(0)
    for _, val := range mem {
        retval += val
    }
    return retval
}

func keys(m map[int]uint64) []int {
    retval := make([]int, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    sort.Ints(retval)
    return retval
}

func memString(mem map[int]uint64) string {
    var sb strings.Builder
    sb.WriteString("{\n")
    for _, key := range keys(mem) {
        fmt.Fprintf(&sb, "  %4d: %13d\n", key, mem[key])
    }
    sb.WriteString("}")
    return sb.String()
}
