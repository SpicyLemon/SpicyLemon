package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type fieldRule struct {
    name string
    min1 int
    max1 int
    min2 int
    max2 int
}

func (this *fieldRule) String() string {
    return fmt.Sprintf("%20s: %3d - %3d or %3d - %3d", this.name, this.min1, this.max1, this.min2, this.max2)
}

func (this *fieldRule) conforms(value int) bool {
    return (this.min1 <= value && value <= this.max1) || (this.min2 <= value && value <= this.max2)
}

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
    rules, myTicket, otherTickets, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("My ticket has %d fields: %s\n", len(myTicket), intsString(myTicket))
    fmt.Printf("There are %d other tickets: %s\n", len(otherTickets), gridString(otherTickets))
    fmt.Print(rulesString(rules))
    problemFields := validateTickets(otherTickets, rules)
    fmt.Printf("Problems: %s\n", gridString(problemFields))
    answer := sumAll(problemFields)
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

func parseInput(input string) ([]fieldRule, []int, [][]int, error) {
    lines := strings.Split(input, "\n")
    fieldRules := []fieldRule{}
    myTicket := []int{}
    otherTickets := [][]int{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    ruleRegex := regexp.MustCompile("^([^:]+): ([[:digit:]]+)-([[:digit:]]+) or ([[:digit:]]+)-([[:digit:]]+)")
    section := 0
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            switch section {
            case 0:
                ruleMatch := ruleRegex.FindStringSubmatch(line)
                if ruleMatch == nil {
                    return nil, nil, nil, fmt.Errorf("Could not parse rule line: [%s].", line)
                }
                min1, err := strconv.Atoi(ruleMatch[2])
                if err != nil {
                    return nil, nil, nil, err
                }
                max1, err := strconv.Atoi(ruleMatch[3])
                if err != nil {
                    return nil, nil, nil, err
                }
                min2, err := strconv.Atoi(ruleMatch[4])
                if err != nil {
                    return nil, nil, nil, err
                }
                max2, err := strconv.Atoi(ruleMatch[5])
                if err != nil {
                    return nil, nil, nil, err
                }
                fieldRules = append(fieldRules, fieldRule{name: ruleMatch[1], min1: min1, max1: max1, min2: min2, max2: max2})
            case 1:
                if line != "your ticket:" {
                    ints, err := csvToInts(line)
                    if err != nil {
                        return nil, nil, nil, err
                    }
                    myTicket = append(myTicket, ints...)
                }
            case 2:
                if line != "nearby tickets:" {
                    ints, err := csvToInts(line)
                    if err != nil {
                        return nil, nil, nil, err
                    }
                    otherTickets = append(otherTickets, ints)
                }
            }
        } else {
            section += 1
        }
    }
    return fieldRules, myTicket, otherTickets, nil
}

func csvToInts(csv string) ([]int, error) {
    retval := []int{}
    for _, v := range strings.Split(csv, ",") {
        i, err := strconv.Atoi(v)
        if err != nil {
            return nil, err
        }
        retval = append(retval, i)
    }
    return retval, nil
}

func gridString(nums [][]int) string {
    var sb strings.Builder
    sb.WriteString("[\n")
    for _, arr := range nums {
        fmt.Fprintf(&sb, "  %s\n", intsString(arr))
    }
    sb.WriteString("]")
    return sb.String()
}

func rulesString(rules []fieldRule) string {
    var sb strings.Builder
    fmt.Fprintf(&sb, "There are %d rules:\n", len(rules))
    for _, r := range rules {
        fmt.Fprintf(&sb, "  %s\n", r.String())
    }
    return sb.String()
}

func intsString(nums []int) string {
    var sb strings.Builder
    sb.WriteString("[")
    for i, n := range nums {
        if i > 0 {
            sb.WriteString(", ")
        }
        fmt.Fprintf(&sb, "%3d", n)
    }
    sb.WriteString("]")
    return sb.String()
}

func validateTickets(tickets [][]int, rules []fieldRule) [][]int {
    retval := make([][]int, len(tickets))
    for i, ticket := range tickets {
        retval[i] = getInvalidFields(ticket, rules)
    }
    return retval
}

func getInvalidFields(ticket []int, rules []fieldRule) []int {
    retval := []int{}
    for _, f := range ticket {
        ok := false
        for _, rule := range rules {
            if rule.conforms(f) {
                ok = true
                break
            }
        }
        if ! ok {
            retval = append(retval, f)
        }
    }
    return retval
}

func sumAll(nums [][]int) int {
    retval := 0
    for _, arr := range nums {
        for _, n := range arr {
            retval += n
        }
    }
    return retval
}
