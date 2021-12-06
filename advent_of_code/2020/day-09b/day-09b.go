package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
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
    filename, target := getCliInput()
    fmt.Printf("Getting input from [%s].\n", filename)
    fmt.Printf("Target: [%d].\n", target)
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
    entries, err := findContiguousSum(input, target)
    if err != nil {
        return err
    }
    answer := min(entries) + max(entries)
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getCliInput() (string, int) {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input", 127
    case 1:
        return args[0], 29221323
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

func findContiguousSum(nums []int, target int) ([]int, error) {
    numsLen := len(nums)
    for i := 0; i < numsLen; i++ {
        sum := nums[i]
        if sum == target {
            return nums[i:i+1], nil
        }
        for j := i + 1; j < numsLen; j++ {
            sum += nums[j]
            if sum == target {
                return nums[i:j+1], nil
            } else if sum > target {
                break
            }
        }
    }
    return []int{}, fmt.Errorf("No combination found to sum to %d.", target)
}

func min(nums []int) int {
    retval := nums[0]
    for i := 1; i < len(nums); i++ {
        if nums[i] < retval {
            retval = nums[i]
        }
    }
    return retval
}

func max(nums []int) int {
    retval := nums[0]
    for i := 1; i < len(nums); i++ {
        if nums[i] > retval {
            retval = nums[i]
        }
    }
    return retval
}
