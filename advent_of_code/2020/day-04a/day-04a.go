package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
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
    // fmt.Println("input:", input)
    required := []string{"byr", "iyr", "eyr", "hgt", "hcl", "ecl", "pid"}
    // optional := [1]string{"cid"}
    answer := countValid(input, required)
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

func parseInput(input string) ([]map[string]string, error) {
    lines := strings.Split(input, "\n")
    retval := []map[string]string{}
    currentData := make(map[string]string)
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        lineHasData := hasDataRegex.FindStringSubmatch(line) != nil
        if (lineHasData) {
            for _, entry := range strings.Split(line, " ") {
                keyValue := strings.Split(entry, ":")
                currentData[keyValue[0]] = keyValue[1]
            }
        } else {
            retval = append(retval, currentData)
            currentData = make(map[string]string)
        }
    }
    if len(currentData) > 0 {
        retval = append(retval, currentData)
    }
    return retval, nil
}

func isValid(toCheck map[string]string, required []string) bool {
    for _, key := range required {
        _, ok := toCheck[key]
        if ! ok {
            return false
        }
    }
    return true
}

func countValid(entries []map[string]string, required []string) int {
    retval := 0
    for _, entry := range entries {
        if (isValid(entry, required)) {
            retval += 1
        }
    }
    return retval
}
