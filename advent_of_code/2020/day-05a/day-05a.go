package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "sort"
    "strings"
    "strconv"
    "os"
)

type seat struct {
    code string
    rowCode string
    columnCode string
    row int
    column int
    id int
}

func NewSeat(code string) (seat, error) {
    this := seat{}
    this.code = code
    this.splitCode()
    err := this.convertRowCode()
    if (err != nil) {
        return this, err
    }
    err = this.convertColumnCode()
    if (err != nil) {
        return this, err
    }
    this.setId()
    return this, nil
}

func (this *seat) splitCode() {
    this.rowCode = this.code[0:7]
    this.columnCode = this.code[7:10]
}

func (this *seat) convertRowCode() error {
    bin := strings.ReplaceAll(strings.ReplaceAll(this.rowCode, "F", "0"), "B", "1")
    val, err := strconv.ParseInt(bin, 2, 0)
    if (err != nil) {
        return err
    }
    this.row = int(val)
    return nil
}

func (this *seat) convertColumnCode() error {
    bin := strings.ReplaceAll(strings.ReplaceAll(this.columnCode, "L", "0"), "R", "1")
    val, err := strconv.ParseInt(bin, 2, 0)
    if (err != nil) {
        return err
    }
    this.column = int(val)
    return nil
}

func (this *seat) setId() {
    this.id = this.row * 8 + this.column
}


// ------ End of type seat ------

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
    answer := input[0].id
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

func parseInput(input string) ([]seat, error) {
    lines := strings.Split(input, "\n")
    retval := []seat{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            s, err := NewSeat(line)
            if (err != nil) {
                fmt.Printf("Failed to parse [%s].\n    Error: %s", line, err)
                return nil, err
            } else {
                retval = append(retval, s)
            }
        }
    }
    sort.Slice(retval[:], func(i int, j int) bool {
        return retval[i].id > retval[j].id
    })
    return retval, nil
}
