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
    filename := getCliInput()
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
    gapCounts := countGaps(input)
    answer := gapCounts[1] * gapCounts[3]
    fmt.Printf("gapCounts[1]: %d\n", gapCounts[1])
    fmt.Printf("gapCounts[3]: %d\n", gapCounts[3])
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getCliInput() string {
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
    retval := []int{0}
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
    sort.Ints(retval)
    return retval, nil
}

func countGaps(nums []int) map[int]int {
    retval := make(map[int]int)
    retval[3] = 1   // Count the final gap to device
    for i := 0; i < len(nums) -1; i++ {
        gap := nums[i+1] - nums[i]
        _, ok := retval[gap]
        if ok {
            retval[gap]++
        } else {
            retval[gap] = 1
        }
    }
    return retval
}
