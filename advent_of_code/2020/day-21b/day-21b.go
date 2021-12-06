package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "sort"
    "strings"
)

type inputLine struct {
    ingredients []string
    alergens []string
}

func (this *inputLine) String() string {
    return fmt.Sprintf("%v comes from %v.", this.alergens, this.ingredients)
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
    input, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("There are %d input lines.\n", len(input))
    ingredients, alergens := getLists(input)
    fmt.Printf("There are %d ingredients: %v\n", len(ingredients), ingredients)
    fmt.Printf("There are %d alergens: %v\n", len(alergens), alergens)
    alergensByIngredients := intersectIngredientsByAlergen(input)
    for alergen, alergenIngredients := range alergensByIngredients {
        fmt.Printf("%s comes from %v.\n", alergen, alergenIngredients)
    }
    alergenIngredients := reduceAlergenIngredients(alergensByIngredients)
    for alergen, ingredient := range alergenIngredients {
        fmt.Printf("%s comes from %s.\n", alergen, ingredient)
    }
    badIngredients := values(alergenIngredients)
    goodIngredients := removeElements(ingredients, badIngredients)
    fmt.Println("Bad Ingredients: %v\n", badIngredients)
    fmt.Println("Good Ingredients: %v\n", goodIngredients)
    answer := canoncialBadIngredients(alergenIngredients)
    fmt.Printf("Answer: %s\n", answer)
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

func parseInput(input string) ([]inputLine, error) {
    lines := strings.Split(input, "\n")
    retval := []inputLine{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    lineRegex := regexp.MustCompile("^([^(]+) \\(contains ([^)]+)\\)$")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            lineMatch := lineRegex.FindStringSubmatch(line)
            if lineMatch == nil {
                return nil, fmt.Errorf("Could not parse line [%s].", line)
            }
            retval = append(retval, inputLine{ingredients: strings.Split(lineMatch[1], " "), alergens: strings.Split(lineMatch[2], ", ")})
        }
    }
    return retval, nil
}

func getLists(inputs []inputLine) ([]string, []string) {
    ingredients := []string{}
    alergens := []string{}
    for _, input := range inputs {
        ingredients = appendUnique(ingredients, input.ingredients...)
        alergens = appendUnique(alergens, input.alergens...)
    }
    return ingredients, alergens
}

func intersectIngredientsByAlergen(inputs []inputLine) map[string][]string {
    retval := make(map[string][]string)
    for _, input := range inputs {
        for _, alergen := range input.alergens {
            if _, okay := retval[alergen]; okay {
                retval[alergen] = intersection(retval[alergen], input.ingredients)
            } else {
                retval[alergen] = input.ingredients
            }
        }
    }
    return retval
}

func reduceAlergenIngredients(alergenIngredients map[string][]string) map[string]string {
    retval := make(map[string]string)
    madeProgress := true
    i := 0
    for len(retval) < len(alergenIngredients) && madeProgress {
        i += 1
        madeProgress = false
        for alergen, ingredients := range alergenIngredients {
            if _, known := retval[alergen]; !known && len(ingredients) == 1 {
                retval[alergen] = ingredients[0]
                madeProgress = true
            }
        }
        for _, ingredient := range retval {
            for alergen, _ := range alergenIngredients {
                alergenIngredients[alergen] = removeElement(alergenIngredients[alergen], ingredient)
            }
        }
    }
    return retval
}

func canoncialBadIngredients(alergenIngredient map[string]string) string {
    alergens := keys(alergenIngredient)
    ingredients := make([]string, len(alergens))
    for i, alergen := range alergens {
        ingredients[i] = alergenIngredient[alergen]
    }
    return strings.Join(ingredients, ",")
}

func countOccurances(inputs []inputLine, ingredients []string) int {
    retval := 0
    for _, input := range inputs {
        retval += len(intersection(input.ingredients, ingredients))
    }
    return retval
}

func appendUnique(slice []string, elems ...string) []string {
    for _, elem := range elems {
        if ! contains(slice, elem) {
            slice = append(slice, elem)
        }
    }
    return slice
}

func contains(slice []string, elem string) bool {
    for _, e := range slice {
        if e == elem {
            return true
        }
    }
    return false
}

func intersection(slice1 []string, slice2 []string) []string {
    retval := []string{}
    for _, e := range slice1 {
        if contains(slice2, e) {
            retval = append(retval, e)
        }
    }
    return retval
}

func removeElement(slice []string, element string) []string {
    retval := []string{}
    for _, e := range slice {
        if e != element {
            retval = append(retval, e)
        }
    }
    return retval
}

func removeElements(slice []string, elements []string) []string {
    retval := []string{}
    for _, e := range slice {
        if ! contains(elements, e) {
            retval = append(retval, e)
        }
    }
    return retval
}

func keys(m map[string]string) []string {
    retval := make([]string, len(m))
    i := 0
    for k, _ := range m {
        retval[i] = k
        i += 1
    }
    sort.Strings(retval)
    return retval
}

func values(m map[string]string) []string {
    retval := make([]string, len(m))
    i := 0
    for _, v := range m {
        retval[i] = v
        i += 1
    }
    sort.Strings(retval)
    return retval
}
