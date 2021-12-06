package main

import (
    "fmt"
    "io/ioutil"
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
    survey, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    // fmt.Println(survey)
    answer := countTrees(survey)
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

func parseInput(input string) ([]string, error) {
    lines := strings.Split(input, "\n")
    retval := []string{}
    for _, line := range lines {
        if len(line) > 2 {
            retval = append(retval, line)
        }
    }
    return retval, nil
}

func countTrees(survey []string) int {
    lineLen := len(survey[0])
    retval := 0
    for i, line := range survey {
        if line[i*3%lineLen] == '#' {
            retval += 1
        }
    }
    return retval
}
