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

func (this *Deck) copyOfFirstCards(size int) *Deck {
    retval := NewDeck()
    card := this.top
    for retval.size < size {
        retval.addBottom(card.value)
        card = card.next
    }
    return retval
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

func (this *Deck) isEqualTo(cards []int) bool {
    if this.size != len(cards) {
        return false
    }
    card := this.top
    for _, value := range cards {
        if value != card.value {
            return false
        }
        card = card.next
    }
    return true
}

type DeckRecord struct {
    deck1 []int
    deck2 []int
}

func NewDeckRecord(deck1 *Deck, deck2 *Deck) *DeckRecord {
    this := DeckRecord{}
    this.deck1 = deck1.toList()
    this.deck2 = deck2.toList()
    return &this
}

func (this *DeckRecord) equalsDecks(deck1 *Deck, deck2 *Deck) bool {
    return deck1.isEqualTo(this.deck1) && deck2.isEqualTo(this.deck2)
}

type GameRecord struct {
    deckRecord *DeckRecord
    winner int
}

func NewGameRecord(deck1 *Deck, deck2 *Deck) *GameRecord {
    this := GameRecord{}
    this.deckRecord = NewDeckRecord(deck1, deck2)
    return &this
}

func (this *GameRecord) knownWinner(deck1 *Deck, deck2 *Deck) (int, bool) {
    if this.deckRecord.equalsDecks(deck1, deck2) {
        return this.winner, true
    }
    return 0, false
}

type Game struct {
    deck1 *Deck
    deck2 *Deck
    deckRecords []*DeckRecord
    gameRecords []*GameRecord
}

func NewGame(deck1 *Deck, deck2 *Deck, gameRecords []*GameRecord) *Game {
    this := Game{}
    this.deck1 = deck1
    this.deck2 = deck2
    this.deckRecords = []*DeckRecord{}
    this.gameRecords = gameRecords
    return &this
}

func (this *Game) haveSeenTheseDecksBefore() bool {
    for _, dr := range this.deckRecords {
        if dr.equalsDecks(this.deck1, this.deck2) {
            return true
        }
    }
    return false
}

func (this *Game) haveSeenThisGameBefore(deck1 *Deck, deck2 *Deck) (bool, int) {
    for _, gameRecord := range this.gameRecords {
        winner, seen := gameRecord.knownWinner(deck1, deck2)
        if seen {
            return true, winner
        }
    }
    return false, 0
}

func (this *Game) playRound() {
    winner := 0 // 0 = no one yet.
    if this.haveSeenTheseDecksBefore() {
        this.deck2 = NewDeck()
    } else {
        this.deckRecords = append(this.deckRecords, NewDeckRecord(this.deck1, this.deck2))
        v1 := this.deck1.removeTop()
        v2 := this.deck2.removeTop()
        if v1 <= this.deck1.size && v2 <= this.deck2.size {
            newDeck1 := this.deck1.copyOfFirstCards(v1)
            newDeck2 := this.deck2.copyOfFirstCards(v2)
            alreadyPlayed, previousWinner := this.haveSeenThisGameBefore(newDeck1, newDeck2)
            if alreadyPlayed {
                winner = previousWinner
            } else {
                gameRecord := NewGameRecord(newDeck1, newDeck2)
                newGame := NewGame(newDeck1, newDeck2, this.gameRecords)
                winner = newGame.play()
                gameRecord.winner = winner
                this.gameRecords = append(newGame.gameRecords, gameRecord)
            }
        } else if v1 > v2 {
            winner = 1
        } else {
            winner = 2
        }
        switch winner {
        case 1:
            this.deck1.addBottom(v1)
            this.deck1.addBottom(v2)
        case 2:
            this.deck2.addBottom(v2)
            this.deck2.addBottom(v1)
        default:
            panic(fmt.Errorf("No winner found! v1:[%d], deck1:%v - v2:[%d], deck2:%v.", v1, this.deck1.toList(), v2, this.deck2.toList()))
        }
    }
}

func (this *Game) play() int {
    for this.deck1.size > 0 && this.deck2.size > 0 {
        this.playRound()
    }
    if this.deck1.size > 0 {
        return 1
    } else if this.deck2.size > 0 {
        return 2
    }
    return 0
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
    game := NewGame(deck1, deck2, []*GameRecord{})
    game.play()
    fmt.Printf("Deck 1 has %d entries: %v\n", game.deck1.size, game.deck1.toList())
    fmt.Printf("Deck 2 has %d entries: %v\n", game.deck2.size, game.deck2.toList())
    answer := calculateScore(game.deck1, game.deck2)
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
