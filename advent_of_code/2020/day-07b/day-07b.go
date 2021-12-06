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
    answer := countBagsRecursively(input, "shiny gold")
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

func countBagsRecursively(rules map[string]map[string]int, outerColor string) int {
    retval := 0
    for innerColor, count := range rules[outerColor] {
        fmt.Printf("Outer color: [%s], Inner color: [%s], count: [%d].\n", outerColor, innerColor, count)
        retval += (countBagsRecursively(rules, innerColor) + 1) * count
    }
    return retval
}
