package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strings"
)

const WHITE = 0
const BLACK = 1

const SW = "sw"
const W = "w"
const NW = "nw"
const NE = "ne"
const E = "e"
const SE = "se"

type Hex struct {
    id int
    color int
    neighbor map[string]*Hex
}

func NewHex() *Hex {
    this := Hex{}
    this.color = WHITE
    this.neighbor = make(map[string]*Hex)
    this.neighbor[SW] = nil
    this.neighbor[W] = nil
    this.neighbor[NW] = nil
    this.neighbor[NE] = nil
    this.neighbor[E] = nil
    this.neighbor[SE] = nil
    return &this
}

func (this *Hex) Get(direction string) *Hex {
    return this.neighbor[direction]
}

func (this *Hex) ToggleColor() {
    this.color = (this.color + 1) % 2
}

func (this *Hex) ConnectNeighbors() {
    dirs := clockwiseDirectionsFrom(W)
    for i, dir := range dirs {
        if this.neighbor[dir] != nil {
            this.neighbor[dir].neighbor[dirs[(i+3)%6]] = this
        }
    }
}

func (this *Hex) String() string {
    var retval strings.Builder
    fmt.Fprintf(&retval, "[%d] is [%s]:", this.id, intToColor(this.color))
    for _, dir := range clockwiseDirectionsFrom(W) {
        fmt.Fprintf(&retval, " %2s=", dir)
        if this.neighbor[dir] != nil {
            fmt.Fprintf(&retval, "[%4d]", this.neighbor[dir].id)
        } else {
            retval.WriteString(" nil  ")
        }
    }
    return retval.String()
}

type Floor struct {
    size int
    hexes []*Hex
}

func NewFloor() *Floor {
    this := Floor{}
    this.hexes = make([]*Hex, 35000) // hexes with a radius of 106 = 34027, this is close enough
    this.size = 0
    this.recordHex(NewHex())
    for i := 1; i <= 25; i++ {
        this.extendFloor()
    }
    return &this
}

func (this *Floor) recordHex(hex *Hex) {
    hex.ConnectNeighbors()
    hex.id = this.size
    // fmt.Printf("Storing hex %d: %s\n", this.size, hex.String())
    this.hexes[this.size] = hex
    this.size += 1
}

func (this *Floor) extendFloor() {
    // go west until there isn't anything
    current := this.hexes[0]
    currentRadius := 0
    for current.neighbor[W] != nil {
        current = current.neighbor[W]
        currentRadius += 1
    }
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(W))
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(NW))
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(NE))
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(E))
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(SE))
    current = this.addEdge(current, currentRadius, clockwiseDirectionsFrom(SW))
}

func (this *Floor) extendFloorIfEdgeTileIsBlack() {
    dirs := clockwiseDirectionsFrom(W)
    blackEdge := false
    for i := this.size - 1; i >= 0 && ! blackEdge; i-- {
        if this.hexes[i].color == 1 {
            for _, dir := range dirs {
                if this.hexes[i].neighbor[dir] == nil {
                    blackEdge = true
                    break
                }
            }
        }
    }
    if blackEdge {
        this.extendFloor()
    }
}

func (this *Floor) addEdge(current *Hex, currentRadius int, hexDir []string) *Hex {
    // hexDir directions should be clockwise. E.g. w, nw, ne, e, se, sw.
    // hexDir[0] = current -> new Hex
    // hexDir[1] = 
    // hexDir[2] = 
    // hexDir[3] = new Hex -> current: opposite of 0
    // hexDir[4] = opposite of 1
    // hexDir[5] = opposite of 2
    cornerHex := NewHex()
    cornerHex.neighbor[hexDir[3]] = current
    if current.neighbor[hexDir[5]] != nil {
        cornerHex.neighbor[hexDir[4]] = current.neighbor[hexDir[5]]
    }
    if current.neighbor[hexDir[1]] != nil {
        cornerHex.neighbor[hexDir[2]] = current.neighbor[hexDir[1]]
    }
    this.recordHex(cornerHex)
    for i := 0; i < currentRadius; i++ {
        current = current.neighbor[hexDir[2]]
        newHex := NewHex()
        newHex.neighbor[hexDir[3]] = current
        newHex.neighbor[hexDir[4]] = current.neighbor[hexDir[5]]
        newHex.neighbor[hexDir[5]] = current.neighbor[hexDir[5]].neighbor[hexDir[0]]
        if current.neighbor[hexDir[1]] != nil {
            newHex.neighbor[hexDir[2]] = current.neighbor[hexDir[1]]
        }
        this.recordHex(newHex)
    }
    return current
}

func (this *Floor) CountBlack() int {
    retval := 0
    for _, hex := range this.hexes {
        if hex != nil {
            retval += hex.color
        }
    }
    return retval
}

func (this *Floor) FlipTile(directions []string) {
    current := this.hexes[0]
    for _, dir := range directions {
        current = current.neighbor[dir]
    }
    current.ToggleColor()
}

func (this *Floor) FlipTiles(directionsList [][]string) {
    for _, directions := range directionsList {
        this.FlipTile(directions)
    }
}

func (this *Floor) Evolve() {
    toFlip := []int{}
    dirs := clockwiseDirectionsFrom(W)
    for id := 0; id < this.size; id++ {
        hex := this.hexes[id]
        blackNeighbors := 0     // :nervous-laughter:
        for _, dir := range dirs {
            if hex.neighbor[dir] != nil {
                blackNeighbors += hex.neighbor[dir].color
            }
        }
        switch hex.color {
        case WHITE:
            if blackNeighbors == 2 {
                toFlip = append(toFlip, id)
            }
        case BLACK:
            if blackNeighbors == 0 || blackNeighbors > 2 {
                toFlip = append(toFlip, id)
            }
        }
    }
    for _, f := range toFlip {
        this.hexes[f].ToggleColor()
    }
}

func (this *Floor) RunFor(days int) {
    for i := 0; i < days; i++ {
        this.extendFloorIfEdgeTileIsBlack()
        this.Evolve()
    }
}

func clockwiseDirectionsFrom(dir string) []string {
    switch dir {
    case SW: return []string{SW, W, NW, NE, E, SE}
    case W: return []string{W, NW, NE, E, SE, SW}
    case NW: return []string{NW, NE, E, SE, SW, W}
    case NE: return []string{NE, E, SE, SW, W, NW}
    case E: return []string{E, SE, SW, W, NW, NE}
    case SE: return []string{SE, SW, W, NW, NE, E}
    }
    panic(fmt.Errorf("Unknown direction string [%s].", dir))
}

func intToColor(color int) string {
    switch color {
    case WHITE: return "WHITE"
    case BLACK: return "BLACK"
    }
    panic(fmt.Errorf("Unknown color [%d].", color))
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

func parseInput(input string) ([][]string, error) {
    lines := strings.Split(input, "\n")
    retval := [][]string{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    directionRegex := regexp.MustCompile("(sw|w|nw|ne|e|se)")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            matches := directionRegex.FindAllStringSubmatch(line, -1)
            if matches != nil && len(matches) > 0 {
                subRetval := make([]string, len(matches))
                for i, match := range matches {
                    subRetval[i] = match[1]
                }
                retval = append(retval, subRetval)
            }
        }
    }
    return retval, nil
}

func run() error {
    filename := getCliParams()
    fmt.Printf("Getting input from [%s].\n", filename)
    dat, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    // fmt.Println(string(dat))
    input, err := parseInput(string(dat))
    if err != nil {
        return err
    }
    maxLen := 0
    for i, line := range input {
        if len(line) > maxLen {
            maxLen = len(line)
        }
        fmt.Printf("%3d: %v\n", i, line)
    }
    fmt.Printf("Max length: %d\n", maxLen)
    floor := NewFloor()
    floor.FlipTiles(input)
    floor.RunFor(100)
    answer := floor.CountBlack()
    fmt.Printf("Answer: %d\n", answer)
    return nil
}
