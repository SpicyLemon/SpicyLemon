package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
)

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

type Circle struct {
    current *Cup
    size int
}

func NewCircle(values []int) *Circle {
    this := Circle{}
    this.current = nil
    for _, value := range values {
        this.addLeft(value)
    }
    return &this
}

func (this *Circle) addLeft(value int) {
    newCup := Cup{}
    newCup.value = value
    if this.current != nil {
        newCup.left = this.current.left
        newCup.right = this.current
        this.current.left.right = &newCup
        this.current.left =  &newCup
    } else {
        newCup.left = &newCup
        newCup.right = &newCup
        this.current = &newCup
    }
    this.size += 1
}

func (this *Circle) addRight(value int) {
    newCup := Cup{}
    newCup.value = value
    if this.current != nil {
        newCup.left = this.current
        newCup.right = this.current.right
        this.current.right.left = &newCup
        this.current.right =  &newCup
    } else {
        newCup.left = &newCup
        newCup.right = &newCup
    }
    this.size += 1
}

func (this *Circle) moveLeft() {
    this.current = this.current.left
}

func (this *Circle) moveRight() {
    this.current = this.current.right
}

func (this *Circle) moveTo(value int) bool {
    for i := 0; i < this.size; i++ {
        if this.current.value == value {
            return true
        }
        this.moveRight()
    }
    this.moveRight()
    return false
}

func (this *Circle) removeLeft() (int, bool) {
    if this.size == 1 {
        retval := this.current.value
        this.current = nil
        this.size = 0
        return retval, true
    } else if this.current != nil {
        cupToRemove := this.current.left
        this.current.left = cupToRemove.left
        cupToRemove.left.right = this.current
        cupToRemove.left = nil
        cupToRemove.right = nil
        this.size -= 1
        return cupToRemove.value, true
    }
    return 0, false
}

func (this *Circle) removeRight() (int, bool) {
    if this.size == 1 {
        retval := this.current.value
        this.current = nil
        this.size = 0
        return retval, true
    } else if this.current != nil {
        cupToRemove := this.current.right
        this.current.right = cupToRemove.right
        cupToRemove.right.left = this.current
        cupToRemove.left = nil
        cupToRemove.right = nil
        this.size -= 1
        return cupToRemove.value, true
    }
    return 0, false
}

func (this *Circle) toList() []int {
    retval := make([]int, this.size)
    c := this.current
    for i := 0; i < this.size; i++ {
        retval[i] = c.value
        c = c.right
    }
    return retval
}

func (this *Circle) getAnswer() string {
    if this.current == nil {
        return ""
    }
    c := this.current
    for i := 0; i < this.size && c.value != 1; i++ {
        c = c.right
    }
    c = c.right
    var retval strings.Builder
    for i := 0; i < this.size - 1; i++ {
        retval.WriteString(strconv.Itoa(c.value))
        c = c.right
    }
    return retval.String()
}

func (this *Circle) String() string {
    return fmt.Sprintf("(%d) %s", this.current.value, intsToString(this.toList()[1:], " "))
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
//  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func playRounds(circle *Circle, count int) error {
    for i := 1; i <= count; i++ {
        fmt.Printf("-- move %d --\n", i)
        err := playRound(circle)
        if err != nil {
            return err
        }
    }
    return nil
}

func playRound(circle *Circle) error {
    fmt.Printf("cups: %v\n", circle.String())
    curVal := circle.current.value
    pulledCups, err := pullCups(circle, 3)
    if err != nil {
        return err
    }
    fmt.Printf("Pick up: %v\n", pulledCups)
    destination := getDestination(circle, pulledCups)
    destinationFound := circle.moveTo(destination)
    if ! destinationFound {
        return fmt.Errorf("Could not find %d in the circle: %s\n", destination, circle.String())
    }
    fmt.Printf("destination: %d\n", destination)
    for i := len(pulledCups) - 1; i >= 0; i-- {
        circle.addRight(pulledCups[i])
    }
    circle.moveTo(curVal)
    circle.moveRight()
    fmt.Printf("cups: %v\n\n", circle.String())
    return nil
}

func pullCups(circle *Circle, count int) ([]int, error) {
    retval := make([]int, count)
    for i := 0; i < count; i++ {
        v, found := circle.removeRight()
        if ! found {
            return nil, fmt.Errorf("No cup found to remove.")
        }
        retval[i] = v
    }
    return retval, nil
}

func getDestination(circle *Circle, pulledCups []int) int {
    retval := circle.current.value - 1
    if retval <= 0 {
        retval = 9
    }
    for contains(pulledCups, retval) {
        retval -= 1
        if retval <= 0 {
            retval = 9
        }
    }
    return retval
}

func indexOf(values []int, value int) (int, bool) {
    for i := 0; i < len(values); i++ {
        if values[i] == value {
            return i, true
        }
    }
    return -1, false
}

func intsToString(values []int, delim string) string {
    strs := make([]string, len(values))
    for i, v := range values {
        strs[i] = strconv.Itoa(v)
    }
    return strings.Join(strs, delim)
}

func contains(values []int, value int) bool {
    for _, val := range values {
        if val == value {
            return true
        }
    }
    return false
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
    err = playRounds(circle, count)
    if err != nil {
        return err
    }
    answer := circle.getAnswer()
    fmt.Printf("Answer: %s\n", answer)
    return nil
}

func getCliParams() (string, int) {
    args := os.Args[1:]
    switch len(args) {
    case 0:
        return "example.input", 10
    case 1:
        return args[0], 100
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
