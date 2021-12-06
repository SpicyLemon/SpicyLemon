package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "sort"
    "strings"
    "strconv"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename, preambleLength := getCliInput()
    fmt.Printf("Getting input from [%s].\n", filename)
    fmt.Printf("Preamble length: [%d].\n", preambleLength)
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
    answer, err := findIllegalValue(input, preambleLength)
    if err != nil {
        return err
    }
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getCliInput() (string, int) {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input", 5
    case 1:
        return args[0], 25
    case 2:
        asInt, err := strconv.Atoi(args[1])
        if err != nil {
            panic(err)
        }
        return args[0], asInt
    }
    panic(fmt.Errorf("Invalid command-line arguments: %s.", args))
}

func parseInput(input string) ([]int, error) {
    lines := strings.Split(input, "\n")
    retval := []int{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            asInt, err := strconv.Atoi(line)
            if err != nil {
                return nil, err
            }
            retval = append(retval, asInt)
        }
    }
    return retval, nil
}

func findIllegalValue(allNums []int, preambleLength int) (int, error) {
    for start := 0; start < len(allNums) - preambleLength - 1; start++ {
        window := allNums[start:start+preambleLength]
        target := allNums[start+preambleLength]
        _, _, err := findSum(window, target)
        if err != nil {
            return target, nil
        }
    }
    return 0, fmt.Errorf("No illegal value was found.\n")
}

func findSum(nums []int, target int) (int, int, error) {
    i := 0
    j := 1
    numsLen := len(nums)
    sorted := make([]int, numsLen)
    copy(sorted, nums)
    sort.Ints(sorted)
    fmt.Printf("Looking for [%d] in %v (length: [%d]).\n", target, nums, numsLen)
    for i < numsLen - 1 {
        sum := sorted[i] + sorted[j]
        fmt.Printf("%d + %d = %d\n", sorted[i], sorted[j], sum)
        if sum == target {
            return sorted[i], sorted[j], nil
        } else if sum > target || j >= numsLen - 1 {
            i += 1
            j = i + 1
        } else {
            j += 1
        }
    }
    return 0, 0, fmt.Errorf("No combination found to sum to %d.", target)
}
