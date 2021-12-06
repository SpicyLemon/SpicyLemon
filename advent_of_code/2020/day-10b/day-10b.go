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
    fmt.Printf("input: %v\n", input)
    nextMap := getPossibleNexts(input)
    fmt.Printf("next map: %s", mapString(nextMap))
    optionGroups := getOptionGroups(nextMap)
    fmt.Printf("Option groups: %s\n", gridString(optionGroups))
    answer := getPathCount(optionGroups)
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

func getPossibleNexts(nums []int) map[int][]int {
    retval := make(map[int][]int)
    numsLen := len(nums)
    for i := 0; i < numsLen - 1; i++ {
        val := nums[i]
        retval[val] = []int{}
        max := val + 3
        for j := 1; j <= 3; j++ {
            if i + j < len(nums) && nums[i+j] <= max {
                retval[val] = append(retval[val], nums[i+j])
            }
        }
    }
    lastNum := nums[numsLen - 1]
    retval[lastNum] = []int{lastNum + 3}
    return retval
}

func getOptionGroups(nums map[int][]int) [][]int {
    breakNums := []int{}
    keyNums := keys(nums)
    for _, key := range keyNums {
        if len(nums[key]) == 1 {
            breakNums = append(breakNums, nums[key][0])
        }
    }
    fmt.Printf("Break Numbers: %v\n", breakNums)
    retval := [][]int{}
    group := []int{}
    for _, key := range keys(nums) {
        if len(nums[key]) > 1 {
            for _, num := range nums[key] {
                if ! contains(group, num) && ! contains(breakNums, num) {
                    group = append(group, num)
                }
            }
        } else if len(group) > 0 {
            retval = append(retval, group)
            group = []int{}
        }
    }
    return retval
}

func getPathCount(optionGroups [][]int) int {
    retval := 1
    for _, group := range optionGroups {
        switch len(group) {
        case 1:
            retval *= 2
        case 2:
            retval *= 4
        case 3:
            retval *= 7
        default:
            fmt.Printf("Not sure what to do with this group: %v\n", group)
        }
    }
    return retval
}

func contains(nums []int, num int) bool {
    for _, v := range nums {
        if v == num {
            return true
        }
    }
    return false
}

func gridString(nums [][]int) string {
    var sb strings.Builder
    sb.WriteString("[\n")
    for _, arr := range nums {
        fmt.Fprintf(&sb, "  %v\n", arr)
    }
    sb.WriteString("]\n")
    return sb.String()
}

func mapString(nums map[int][]int) string {
    var sb strings.Builder
    sb.WriteString("[\n")
    for _, key := range keys(nums) {
        fmt.Fprintf(&sb, "  %d: %v\n", key, nums[key])
    }
    sb.WriteString("]\n")
    return sb.String()
}

func keys(m map[int][]int) []int {
    retval := make([]int, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    sort.Ints(retval)
    return retval
}
