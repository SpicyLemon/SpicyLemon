package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strings"
)

type satRule struct {
    num string
    originalValue string
    hasOptions bool
    currentValue string
}

func (this *satRule) setCurrentValue(curVal string) {
    this.currentValue = curVal
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
    rules, input, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("Rules: %v\n", rules)
    fmt.Printf("Input: %v\n", input)
    numberRegex := regexp.MustCompile("([[:digit:]]+)")
    rules = resolveValue(rules, "0", numberRegex, 0)
    fmt.Printf("Rules: %v\n", rules)
    fmt.Printf("Rule 0: %s\n", rules["0"].currentValue)
    fmt.Printf("Rule 8: %s\n", rules["8"].currentValue)
    fmt.Printf("Rule 11: %s\n", rules["11"].currentValue)
    maxMessageLength := maxLength(input)
    fmt.Printf("Max Message Length = %d\n", maxMessageLength)
    answer := countValid(rules, input)
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

func parseInput(input string) (map[string]satRule, []string, error) {
    lines := strings.Split(input, "\n")
    rules := make(map[string]satRule)
    messages := []string{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    quotesRegex := regexp.MustCompile("\"")
    mode := 0
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            switch mode {
            case 0:
                lineSplit := strings.Split(line, ": ")
                rule := satRule{}
                rule.num = lineSplit[0]
                rule.originalValue = quotesRegex.ReplaceAllLiteralString(lineSplit[1], "")
                rule.currentValue = rule.originalValue
                rule.hasOptions = stringContains(rule.originalValue, '|')
                rules[lineSplit[0]] = rule
            case 1:
                messages = append(messages, line)
            }
        } else {
            mode += 1
        }
    }
    return rules, messages, nil
}

func resolveValue(rules map[string]satRule, key string, numberRegex *regexp.Regexp, indent int) map[string]satRule {
    mainRule := rules[key]
    ind := strings.Repeat("  ", indent)
    fmt.Printf("%sResolving %s: \"%s\"\n", ind, key, mainRule.currentValue)
    for {
        numberMatch := numberRegex.FindStringSubmatchIndex(mainRule.currentValue)
        if numberMatch == nil {
            break
        }
        subKey := mainRule.currentValue[numberMatch[0]:numberMatch[1]]
        subValue := "UNKNOWN"
        if subKey == key {
            switch key {
            case "8":
                subValue = "EIGHT"
            case "11":
                subValue = "ELEVEN"
            default:
                panic(fmt.Errorf("Unexpected recursive key: [%s].", key))
            }
        } else {
            rules = resolveValue(rules, subKey, numberRegex, indent + 1)
            subValue = rules[subKey].currentValue
        }
        if rules[subKey].hasOptions {
            subValue = "(" + subValue + ")"
        }
        newValue := substringReplace(mainRule.currentValue, numberMatch[0], numberMatch[1], subValue)
        fmt.Printf("%s  %s: \"%s\" replacing \"%s\" with \"%s\" to become \"%s\"\n", ind, key, mainRule.currentValue, subKey, subValue, newValue)
        mainRule.setCurrentValue(newValue)
        rules[key] = mainRule
    }
    fmt.Printf("%s%s resolved to \"%s\"\n", ind, key, mainRule.currentValue)
    return rules
}

func stringContains(str string, c rune) bool {
    for _, s := range str {
        if s == c {
            return true
        }
    }
    return false
}

func substringReplace(orig string, startI int, endI int, repl string) string {
    return orig[0:startI] + repl + orig[endI:]
}

func countValid(rules map[string]satRule, messages []string) int {
    ruleRegex := regexp.MustCompile("^" + makeRuleString(rules) + "$")
    retval := 0
    for _, message := range messages {
        if ruleRegex.MatchString(message) {
            retval += 1
        }
    }
    return retval
}

func maxLength(messages []string) int {
    retval := 0
    for _, m := range messages {
        if len(m) > retval {
            retval = len(m)
        }
    }
    return retval
}

func makeRuleString(rules map[string]satRule) string {
    spaceRegex := regexp.MustCompile("[[:space:]]")
    eightRegex := regexp.MustCompile("\\(EIGHT\\)")
    elevenRegex := regexp.MustCompile("\\(ELEVEN\\)")
    eightStr := "(" + spaceRegex.ReplaceAllLiteralString(rules["8"].currentValue, "") + ")"
    elevenStr := "(" + spaceRegex.ReplaceAllLiteralString(rules["11"].currentValue, "") + ")"
    retval := spaceRegex.ReplaceAllLiteralString(rules["0"].currentValue, "")
    for i := 0; i < 15; i++ {
        retval = eightRegex.ReplaceAllLiteralString(retval, eightStr)
    }
    retval = eightRegex.ReplaceAllLiteralString(retval, "")
    for i := 0; i < 8; i++ {
        retval = elevenRegex.ReplaceAllLiteralString(retval, elevenStr)
    }
    retval = elevenRegex.ReplaceAllLiteralString(retval, "")
    return retval
}
