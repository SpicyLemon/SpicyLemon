package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "strings"
    "strconv"
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
    if false {
        for i, entry := range input {
            fmt.Printf("input[%s] = [%s].\n", i, entry)
        }
    }
    canBeIn := makeCanBeInMap(input)
    if false {
        for i, entry := range canBeIn {
            fmt.Printf("canBeIn[%s] = [%s].\n", i, entry)
        }
    }
    answers := getCanBeInsRecursively(canBeIn, "shiny gold")
    answer := len(answers)
    fmt.Printf("Answer: %d - %s\n", answer, answers)
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
func parseInput(input string) (map[string]map[string]int, error) {
    lines := strings.Split(input, "\n")
    retval := make(map[string]map[string]int)
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    ruleRegex := regexp.MustCompile("^([[:digit:]]+|no) (.+) bags?$")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            parts := strings.Split(line, " bags contain ")
            if len(parts) != 2 {
                fmt.Printf("Could not initially split line: [%s].\n", line)
            } else {
                outerColor := parts[0]
                retval[outerColor] = make(map[string]int)
                for _, rule := range strings.Split(parts[1][:len(parts[1])-1], ", ") {
                    match := ruleRegex.FindStringSubmatch(rule)
                    if len(match) != 3 {
                        fmt.Printf("Failed to parse rule [%s].\n", rule)
                    } else {
                        if match[1] != "no" {
                            count, err := strconv.Atoi(match[1])
                            if err != nil {
                                fmt.Printf("Failed to parse rule [%s]: %s\n", rule, err)
                            }
                            retval[outerColor][match[2]] = count
                        }
                    }
                }
            }
        }
    }
    return retval, nil
}

func makeCanBeInMap(input map[string]map[string]int) map[string][]string {
    retval := make(map[string][]string)
    for outerColor, requirements := range(input) {
        for innerColor, _ := range(requirements) {
            _, ok := retval[innerColor]
            if ! ok {
                retval[innerColor] = []string{}
            }
            if ! contains(retval[innerColor], outerColor) {
                retval[innerColor] = append(retval[innerColor], outerColor)
            }
        }
    }
    return retval
}

func contains(array []string, value string) bool {
    for _, arrayVal := range array {
        if arrayVal == value {
            return true
        }
    }
    return false
}

func getCanBeInsRecursively(canBeIn map[string][]string, innerColor string) []string {
    retval := canBeIn[innerColor][:]
    for _, outerColor := range canBeIn[innerColor] {
        for _, secondOuterColor := range getCanBeInsRecursively(canBeIn, outerColor) {
            if ! contains(retval, secondOuterColor) {
                retval = append(retval, secondOuterColor)
            }
        }
    }
    return retval
}
