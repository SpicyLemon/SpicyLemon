package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type Card struct {
    value int
    next *Card
}

type Deck struct {
    top *Card
    bottom *Card
    size int
}

func NewDeck() *Deck {
    this := Deck{}
    this.top = nil
    this.bottom = nil
    this.size = 0
    return &this
}

func (this *Deck) removeTop() int {
    retval := this.top.value
    this.top = this.top.next
    this.size -= 1
    return retval
}

func (this *Deck) addBottom(value int) {
    newCard := Card{value: value}
    if this.bottom != nil {
        this.bottom.next = &newCard
    }
    this.bottom = &newCard
    if this.top == nil {
        this.top = &newCard
    }
    this.size += 1
}

func (this *Deck) toList() []int {
    retval := make([]int, this.size)
    c := this.top
    for i := 0; i < this.size; i++ {
        retval[i] = c.value
        c = c.next
    }
    return retval
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
    deck1, deck2, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("Deck 1 has %d entries: %v\n", deck1.size, deck1.toList())
    fmt.Printf("Deck 2 has %d entries: %v\n", deck2.size, deck2.toList())
    rounds := playGame(deck1, deck2)
    fmt.Printf("After round %d:\n", rounds)
    fmt.Printf("Deck 1 has %d entries: %v\n", deck1.size, deck1.toList())
    fmt.Printf("Deck 2 has %d entries: %v\n", deck2.size, deck2.toList())
    answer := calculateScore(deck1, deck2)
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

func parseInput(input string) (*Deck, *Deck, error) {
    lines := strings.Split(input, "\n")
    decks := make([]*Deck, 2)
    decks[0] = NewDeck()
    decks[1] = NewDeck()
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    d := 0
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            i, err := strconv.Atoi(line)
            if err == nil {
                decks[d].addBottom(i)
            }
        } else {
            d += 1
        }
    }
    return decks[0], decks[1], nil
}

func playGame(deck1 *Deck, deck2 *Deck) int {
    retval := 0
    for deck1.size > 0 && deck2.size > 0 {
        retval += 1
        playRound(deck1, deck2)
    }
    return retval
}

func playRound(deck1 *Deck, deck2 *Deck) {
    v1 := deck1.removeTop()
    v2 := deck2.removeTop()
    if v1 > v2 {
        deck1.addBottom(v1)
        deck1.addBottom(v2)
    } else {
        deck2.addBottom(v2)
        deck2.addBottom(v1)
    }
}

func calculateScore(deck1 *Deck, deck2 *Deck) int {
    deck := deck1
    if deck2.size > 0 {
        deck = deck2
    }
    vals := deck.toList()
    size := len(vals)
    retval := 0
    for i := 0; i < size; i++ {
        retval += (size - i) * vals[i]
    }
    return retval
}
