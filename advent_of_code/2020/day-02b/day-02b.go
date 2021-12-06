package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "strconv"
    "strings"
    "os"
)

type pwpolicy struct {
    min int
    max int
    letter byte
    password string
}

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
    pwpolicies, err2 := toPwpolicyList(string(dat))
    if (err2 != nil) {
        return err2
    }
    // fmt.Println(pwpolicies)
    answer := countValid(pwpolicies)
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

func toPwpolicyList(input string) ([]pwpolicy, error) {
    lines := strings.Split(input, "\n")
    retval := []pwpolicy{}
    lineRegex := regexp.MustCompile("^(\\d+)-(\\d+) ([a-zA-Z]): (\\S+).*$")
    for _, line := range lines {
        match := lineRegex.FindStringSubmatch(line)
        if match == nil {
            fmt.Printf("Line: [%s] could not be parsed.\n", line)
        } else {
            minInt, err1 := strconv.Atoi(match[1])
            maxInt, err2 := strconv.Atoi(match[2])
            if err1 != nil {
                return nil, err1
            } else if err2 != nil {
                return nil, err2
            }
            retval = append(retval, pwpolicy{min: minInt, max: maxInt, letter: match[3][0], password: match[4]})
        }
    }
    return retval, nil
}

func isValid(toCheck pwpolicy) bool {
    letter1 := toCheck.password[toCheck.min - 1]
    letter2 := toCheck.password[toCheck.max - 1]
    return letter1 != letter2 && (letter1 == toCheck.letter || letter2 == toCheck.letter)
}

func countValid(pwpolicies []pwpolicy) int {
    retval := 0
    for _, pol := range pwpolicies {
        if (isValid(pol)) {
            retval += 1
        }
    }
    return retval
}
