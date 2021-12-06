package main

import (
    "fmt"
    "io/ioutil"
    "sort"
    "strconv"
    "strings"
    "os"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename := getInputFilename()
    dat, err := ioutil.ReadFile(filename)
    if (err != nil) {
        return err
    }
    // fmt.Println(string(dat))
    ints, err2 := toIntList(string(dat))
    if (err2 != nil) {
        return err2
    }
    fmt.Println(ints)
    int1, int2, int3, err3 := findSum(ints, 2020)
    if (err3 != nil) {
        return err3
    }
    fmt.Printf("%d + %d + %d = 2020\n", int1, int2, int3)
    answer := int1 * int2 * int3
    fmt.Printf("%d * %d * %d = %d\n", int1, int2, int3, answer)
    return nil
}

func getInputFilename() string {
    args := os.Args[1:]
    if len(args) == 0 {
        return "example.input"
    }
    return args[0]
}

func toIntList(input string) ([]int, error) {
    lines := strings.Split(input, "\n")
    retval := []int{}
    for _, line := range lines {
        asInt, err := strconv.Atoi(line)
        if err != nil {
            return nil, err
        }
        retval = append(retval, asInt)
    }
    sort.Ints(retval)
    return retval, nil
}

func findSum(nums []int, target int) (int, int, int, error) {
    i := 0
    j := 1
    k := 2
    numsLen := len(nums)
    for i <= numsLen - 2 {
        sum := nums[i] + nums[j] + nums[k]
        if sum == target {
            return nums[i], nums[j], nums[k], nil
        } else if sum > target {
            if j < k - 1 {
                j += 1
                k = j + 1
            } else {
                i += 1
                j = i + 1
                k = i + 2
            }
        } else {
            k += 1
        }
    }
    return 0, 0, 0, fmt.Errorf("No combination found to sum to %d.", target)
}
