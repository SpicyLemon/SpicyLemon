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
    answer := countAnswers(input)
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

func parseInput(input string) ([][][]byte, error) {
    lines := strings.Split(input, "\n")
    retval := [][][]byte{}
    group := [][]byte{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            group = append(group, []byte(line))
        } else {
            retval = append(retval, group)
            group = [][]byte{}
        }
    }
    if len(group) > 0 {
        retval = append(retval, group)
    }
    return retval, nil
}

func getIntersection(input [][]byte) []byte {
    if len(input) == 1 {
        return input[0]
    }
    retval := []byte{}
    for _, c := range input[0] {
        allHave := true
        for i := 1; i < len(input); i++ {
            if ! contains(input[i], c) {
                allHave = false
                break
            }
        }
        if allHave {
            retval = append(retval, c)
        }
    }
    return retval
}

func contains(array []byte, value byte) bool {
    for _, arrayVal := range array {
        if arrayVal == value {
            return true
        }
    }
    return false
}

func countAnswers(input [][][]byte) int {
    retval := 0
    for _, group := range input {
        allContain := getIntersection(group)
        fmt.Printf("Intersection: [%s] from group [%s].\n", string(allContain), group)
        retval += len(allContain)
    }
    return retval
}
