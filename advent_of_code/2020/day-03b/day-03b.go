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
    answer_1_1 := countTrees(survey, 1, 1)
    answer_3_1 := countTrees(survey, 3, 1)
    answer_5_1 := countTrees(survey, 5, 1)
    answer_7_1 := countTrees(survey, 7, 1)
    answer_1_2 := countTrees(survey, 1, 2)
    answer := answer_1_1 * answer_3_1 * answer_5_1 * answer_7_1 * answer_1_2
    fmt.Printf("Answer right 1 down 1: %d\n", answer_1_1)
    fmt.Printf("Answer right 3 down 1: %d\n", answer_3_1)
    fmt.Printf("Answer right 5 down 1: %d\n", answer_5_1)
    fmt.Printf("Answer right 7 down 1: %d\n", answer_7_1)
    fmt.Printf("Answer right 1 down 2: %d\n", answer_1_2)
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

func countTrees(survey []string, right int, down int) int {
    lineLen := len(survey[0])
    retval := 0
    i := 0
    lastI := len(survey) - down
    for i < lastI {
        i += down
        if survey[i][i / down * right % lineLen] == '#' {
            retval += 1
        }
    }
    return retval
}
