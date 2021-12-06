package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename := getCliParams()
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
    answer := 0
    for _, expr := range input {
        exprAns, err := solve(expr)
        if err != nil {
            fmt.Printf("Error encountered on [%s].\n", expr)
            return err
        }
        fmt.Printf("%d = %s\n", exprAns, expr)
        answer += exprAns
    }
    if (err != nil) {
        return err
    }
    fmt.Printf("Answer: %d\n", answer)
    return nil
}

func getCliParams() string {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input"
    case 1:
        return args[0]
    }
    panic(fmt.Errorf("Invalid command-line arguments: %s.", args))
}

func parseInput(input string) ([]string, error) {
    lines := strings.Split(input, "\n")
    retval := []string{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            retval = append(retval, line)
        }
    }
    return retval, nil
}

func solve(expr string) (int, error) {
    parenRegex := regexp.MustCompile("\\([^()]+\\)")
    fmt.Printf("Evaluating [%s].\n", expr)
    for {
        parenMatch := parenRegex.FindStringSubmatchIndex(expr)
        if parenMatch == nil {
            break
        }
        fmt.Printf("Has parens: [%s].\n", expr)
        subval, err := solve(expr[parenMatch[0]+1:parenMatch[1]-1])
        if err != nil {
            return -1, err
        }
        expr = substringReplace(expr, parenMatch[0], parenMatch[1], fmt.Sprintf("%d", subval))
        fmt.Printf("Expression is now [%s].\n", expr)
    }
    plusRegex := regexp.MustCompile("([[:digit:]]+) \\+ ([[:digit:]]+)")
    for {
        plusMatch := plusRegex.FindStringSubmatchIndex(expr)
        if plusMatch == nil {
            break
        }
        fmt.Printf("Has plus: [%s].\n", expr)
        split := strings.Split(expr[plusMatch[0]:plusMatch[1]], " ")
        v1, err := strconv.Atoi(split[0])
        if err != nil {
            return -1, err
        }
        v2, err := strconv.Atoi(split[2])
        if err != nil {
            return -1, err
        }
        fmt.Printf("Adding [%d] to [%d].\n", v1, v2)
        subval := v1 + v2
        expr = substringReplace(expr, plusMatch[0], plusMatch[1], fmt.Sprintf("%d", subval))
        fmt.Printf("Expression is now [%s].\n", expr)
    }
    mathParts := strings.Split(expr, " ")
    fmt.Printf("Math parts: %v.\n", mathParts)
    retval, err := strconv.Atoi(mathParts[0])
    if err != nil {
        return -1, err
    }
    for i := 1; i < len(mathParts); i += 2 {
        switch mathParts[i] {
        case "+":
            v, err := strconv.Atoi(mathParts[i+1])
            if err != nil {
                return -1, err
            }
            fmt.Printf("Adding [%d] to [%d].\n", retval, v)
            retval += v
        case "*":
            v, err := strconv.Atoi(mathParts[i+1])
            if err != nil {
                return -1, err
            }
            fmt.Printf("Multiplying [%d] by [%d].\n", retval, v)
            retval *= v
        }
    }
    return retval, nil
}

func substringReplace(orig string, startI int, endI int, repl string) string {
    return orig[0:startI] + repl + orig[endI:]
}
