package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
)

const oneMillion = 1000000
const tenMillion = 10000000
const maxCup = oneMillion

type Cup struct {
    value int
    left *Cup
    right *Cup
}

func NewCup(value int, left *Cup, right *Cup) *Cup {
    this := Cup{}
    this.value = value
    this.left = left
    this.right = right
    return &this
}

func (this *Cup) addLeft(value int) *Cup {
    newCup := NewCup(value, this.left, this)
    newCup.left.right = newCup
    this.left = newCup
    return newCup
}

type Circle struct {
    current *Cup
    size int
    cups map[int]*Cup
}

func NewCircle(values []int) *Circle {
    this := Circle{}
    this.cups = make(map[int]*Cup)
    this.current = nil
    for _, value := range values {
        this.addLeft(value)
    }
    for value := 10; value <= maxCup; value++ {
        this.addLeft(value)
    }
    return &this
}

func (this *Circle) addLeft(value int) {
    if this.current == nil {
        newCup := NewCup(value, nil, nil)
        newCup.left = newCup
        newCup.right = newCup
        this.current = newCup
        this.cups[value] = newCup
    } else {
        this.cups[value] = this.current.addLeft(value)
    }
    this.size += 1
}

func (this *Circle) addRightAfter(toFind int, firstCup *Cup) bool {
    cup := this.find(toFind)
    if cup == nil {
        return false
    }
    lastCup := firstCup
    cupCount := 1
    for lastCup.right != nil {
        lastCup = lastCup.right
        cupCount += 1
    }
    cup.right.left = lastCup
    lastCup.right = cup.right
    firstCup.left = cup
    cup.right = firstCup
    this.size += cupCount
    return true
}

func (this *Circle) moveRightOne() {
    this.current = this.current.right
}

func (this *Circle) find(value int) *Cup {
    return this.cups[value]
}

func (this *Circle) removeRightCups(count int) *Cup {
    firstCup := this.current.right
    lastCup := this.current.right
    for i := 1; i < count; i++ {
        lastCup = lastCup.right
    }
    this.current.right = lastCup.right
    this.current.right.left = this.current
    firstCup.left = nil
    lastCup.right = nil
    this.size -= count
    return firstCup
}

func (this *Circle) String(count int) string {
    length := count
    if count < 0 || count > this.size {
        length = this.size
    }
    var retval strings.Builder
    c := this.current
    fmt.Fprintf(&retval, "(%d)", c.value)
    for i := 1; i < length; i++ {
        c = c.right
        fmt.Fprintf(&retval, " %d", c.value)
    }
    return retval.String()
}

func (this *Circle) getAnswer() (int, int) {
    oneCup := this.find(1)
    star1 := oneCup.right.value
    star2 := oneCup.right.right.value
    return star1, star2
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
//  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func playRounds(circle *Circle, count int) error {
    for i := 1; i <= count; i++ {
        printInfo := i <= 5 || i % 1000 == 0
        if printInfo {
            fmt.Printf("-- move %d --\n", i)
        }
        err := playRound(circle, printInfo)
        if err != nil {
            fmt.Printf("Error encountered in round %d.", i)
            return err
        }
    }
    return nil
}

func playRound(circle *Circle, printInfo bool) error {
    if printInfo {
        fmt.Printf("cups: %s ... [size: %d]\n", circle.String(15), circle.size)
    }
    firstPulledCup := circle.removeRightCups(3)
    if printInfo {
        pulledCupValues := cupValues(firstPulledCup, 3)
        fmt.Printf("pick up: %v\n", pulledCupValues)
    }
    destination := getDestination(circle, firstPulledCup)
    if printInfo {
        fmt.Printf("destination: %d\n", destination)
    }
    added := circle.addRightAfter(destination, firstPulledCup)
    if ! added {
        return fmt.Errorf("Could not find %d in the circle.", destination)
    }
    circle.moveRightOne()
    if printInfo {
        fmt.Printf("cups: %s ... [size: %d]\n\n", circle.String(15), circle.size)
    }
    return nil
}

func getDestination(circle *Circle, firstPulledCup *Cup) int {
    retval := circle.current.value - 1
    if retval <= 0 {
        retval = maxCup
    }
    for contains(firstPulledCup, retval) {
        retval -= 1
        if retval <= 0 {
            retval = maxCup
        }
    }
    return retval
}

func contains(cup *Cup, value int) bool {
    for cup != nil {
        if cup.value == value {
            return true
        }
        cup = cup.right
    }
    return false
}

func cupValues(cup *Cup, max int) []int {
    retval := make([]int, max)
    for i := 0; i < max && cup != nil; i++ {
        retval[i] = cup.value
        cup = cup.right
    }
    return retval
}

// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    filename, count := getCliParams()
    fmt.Printf("Getting input from [%s].\n", filename)
    fmt.Printf("Executing %d rounds.\n", count)
    dat, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    // fmt.Println(string(dat))
    circle, err := parseInput(string(dat))
    if err != nil {
        return err
    }
    fmt.Printf("Starting circle: %s ... [size: %d].\n", circle.String(15), circle.size)
    err = playRounds(circle, count)
    if err != nil {
        return err
    }
    star1, star2 := circle.getAnswer()
    answer := star1 * star2
    fmt.Printf("Answer: %d = %d * %d\n", answer, star1, star2)
    return nil
}

func getCliParams() (string, int) {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input", tenMillion
    case 1:
        return args[0], tenMillion
    case 2:
        count, err := strconv.Atoi(args[1])
        if err != nil {
            panic(err)
        }
        return args[0], count
    }
    panic(fmt.Errorf("Invalid command-line arguments: %s.", args))
}

func parseInput(input string) (*Circle, error) {
    lines := strings.Split(input, "\n")
    vals := make([]int, len(lines[0]))
    for i, d := range strings.Split(lines[0], "") {
        val, err := strconv.Atoi(d)
        if err != nil {
            return nil, err
        }
        vals[i] = val
    }
    retval := NewCircle(vals)
    return retval, nil
}
