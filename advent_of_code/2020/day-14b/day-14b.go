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
    addr uint64
    value uint64
}

type instructionGroup struct {
    maskStr string
    maskOrStr string
    maskOr uint64
    maskAndNotStr string
    maskAndNot uint64
    floatVals []uint64
    instructions []instruction
}

func (this *instructionGroup) addInstruction(addr uint64, value uint64) {
    this.instructions = append(this.instructions, instruction{addr: addr, value: value})
}

func (this *instructionGroup) runInstructions(mem map[uint64]uint64) {
    for _, inst := range this.instructions {
        baseMem := (inst.addr | this.maskOr) &^ this.maskAndNot
        for _, floater := range this.floatVals {
            mem[baseMem + floater] = inst.value
        }
    }
}

func (this *instruction) String() string {
    return fmt.Sprintf("mem[%d] = %d", this.addr, this.value)
}

func (this *instructionGroup) String() string {
    var sb strings.Builder
    fmt.Fprintf(&sb, "        Mask: %s\n", this.maskStr)
    fmt.Fprintf(&sb, "     Mask or: %s = %d\n", this.maskOrStr, this.maskOr)
    fmt.Fprintf(&sb, "Mask and not: %s = %d\n", this.maskAndNotStr, this.maskAndNot)
    fmt.Fprintf(&sb, "Float Values: %v\n", this.floatVals)
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
    mem := make(map[uint64]uint64)
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
    maskLineRegex := regexp.MustCompile("^mask[[:space:]]*=[[:space:]]*([X01]{36})")
    memLineRegex := regexp.MustCompile("^mem\\[([[:digit:]]+)\\][[:space:]]*=[[:space:]]*([[:digit:]]+)")
    xRegex := regexp.MustCompile("X")
    oneRegex := regexp.MustCompile("1")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            maskMatch := maskLineRegex.FindStringSubmatch(line)
            if maskMatch != nil {
                mask := maskMatch[1]
                orMaskStr := xRegex.ReplaceAllString(mask, "0")
                andNotMaskStrXs := oneRegex.ReplaceAllString(mask, "0")
                andNotMaskStr := xRegex.ReplaceAllString(andNotMaskStrXs, "1")
                orMask, err := strconv.ParseUint(orMaskStr, 2, 64)
                if (err != nil) {
                    return nil, err
                }
                andNotMask, err := strconv.ParseUint(andNotMaskStr, 2, 64)
                if (err != nil) {
                    return nil, err
                }
                floatValues, err := getFloatValues(andNotMaskStrXs)
                if (err != nil) {
                    return nil, err
                }
                retval = append(retval, instructionGroup{maskStr: mask, maskOr: orMask, maskAndNot: andNotMask, maskOrStr: orMaskStr, maskAndNotStr: andNotMaskStr, floatVals: floatValues})
            } else {
                memMatch := memLineRegex.FindStringSubmatch(line)
                if memMatch == nil {
                    return nil, fmt.Errorf("Could not parse line [%s].\n", line)
                }
                addr, err := strconv.ParseUint(memMatch[1], 10, 64)
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

func getFloatValues(mask string) ([]uint64, error) {
    xRegex := regexp.MustCompile("X")
    xMatch := xRegex.FindStringIndex(mask)
    retval := []uint64{}
    if xMatch != nil {
        newMask0 := []byte(mask)
        newMask0[xMatch[0]] = '0'
        newMask1 := []byte(mask)
        newMask1[xMatch[0]] = '1'
        fmt.Printf("    mask: %s\n", mask)
        fmt.Printf("newMask0: %s\n", string(newMask0))
        fmt.Printf("newMask1: %s\n", string(newMask1))
        newFloats0, err := getFloatValues(string(newMask0))
        if err != nil {
            return nil, err
        }
        newFloats1, err := getFloatValues(string(newMask1))
        if err != nil {
            return nil, err
        }
        retval = append(retval, newFloats0...)
        retval = append(retval, newFloats1...)
    } else {
        maskVal, err := strconv.ParseUint(mask, 2, 64)
        if err != nil {
            return nil, err
        }
        retval = append(retval, maskVal)
    }
    return retval, nil
}

func runAllInstructions(mem map[uint64]uint64, instructionGroups []instructionGroup) {
    for _, instg := range instructionGroups {
        instg.runInstructions(mem)
    }
}

func sumMem(mem map[uint64]uint64) uint64 {
    retval := uint64(0)
    for _, val := range mem {
        retval += val
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
    sort.Slice(retval, func(i, j int) bool { return retval[i] < retval[j] })
    return retval
}

func memString(mem map[uint64]uint64) string {
    var sb strings.Builder
    sb.WriteString("{\n")
    for _, key := range keys(mem) {
        fmt.Fprintf(&sb, "  %4d: %13d\n", key, mem[key])
    }
    sb.WriteString("}")
    return sb.String()
}

func pow2(exponent int) uint64 {
    if exponent < 0 {
        panic(fmt.Errorf("Exponent [%d] cannot be less than zero.", exponent))
    }
    if exponent > 64 {
        panic(fmt.Errorf("Exponent [%d] cannot be greater than 64.", exponent))
    }
    switch exponent {
    case 0: return uint64(1)
    case 1: return uint64(2)
    case 2: return uint64(4)
    case 3: return uint64(8)
    case 4: return uint64(16)
    case 5: return uint64(32)
    case 6: return uint64(64)
    case 7: return uint64(128)
    case 8: return uint64(256)
    case 9: return uint64(512)
    case 10: return uint64(1024)
    case 11: return uint64(2048)
    case 12: return uint64(4096)
    case 13: return uint64(8192)
    case 14: return uint64(16384)
    case 15: return uint64(32768)
    case 16: return uint64(65536)
    case 17: return uint64(131072)
    case 18: return uint64(262144)
    case 19: return uint64(524288)
    case 20: return uint64(1048576)
    case 21: return uint64(2097152)
    case 22: return uint64(4194304)
    case 23: return uint64(8388608)
    case 24: return uint64(16777216)
    case 25: return uint64(33554432)
    case 26: return uint64(67108864)
    case 27: return uint64(134217728)
    case 28: return uint64(268435456)
    case 29: return uint64(536870912)
    case 30: return uint64(1073741824)
    case 31: return uint64(2147483648)
    case 32: return uint64(4294967296)
    case 33: return uint64(8589934592)
    case 34: return uint64(17179869184)
    case 35: return uint64(34359738368)
    case 36: return uint64(68719476736)
    case 37: return uint64(137438953472)
    case 38: return uint64(274877906944)
    case 39: return uint64(549755813888)
    default:
        retval := pow2(39)
        for i := 40; i <= exponent; i++ {
            retval *= uint64(2)
        }
        return retval
    }
    return 0
}
