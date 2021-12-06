package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type pictureTile struct {
    number int
    contents [][]byte
    edges [4][]byte
    neighbors []int
}

func (this *pictureTile) getTopEdge() []byte {
    this.ensureEdges()
    return this.edges[0]
}

func (this *pictureTile) getRightEdge() []byte {
    this.ensureEdges()
    return this.edges[1]
}

func (this *pictureTile) getBottomEdge() []byte {
    this.ensureEdges()
    return this.edges[2]
}

func (this *pictureTile) getLeftEdge() []byte {
    this.ensureEdges()
    return this.edges[3]
}

func (this *pictureTile) ensureEdges() {
    // Top edge
    if this.edges[0] == nil || len(this.edges[0]) == 0 {
        this.edges[0] = make([]byte, len(this.contents[0]))
        for i, b := range this.contents[0] {
            this.edges[0][i] = b
        }
    }
    // Right edge
    lastByte := len(this.contents[0]) - 1
    if this.edges[1] == nil || len(this.edges[1]) == 0 {
        this.edges[1] = make([]byte, len(this.contents[lastByte]))
        for i := 0; i < len(this.contents); i++ {
            this.edges[1][i] = this.contents[i][lastByte]
        }
    }
    // Bottom edge
    lastLine := len(this.contents) - 1
    if this.edges[2] == nil || len(this.edges[2]) == 0 {
        this.edges[2] = make([]byte, len(this.contents[lastLine]))
        for i, b := range this.contents[lastLine] {
            this.edges[2][i] = b
        }
    }
    // Left edge
    if this.edges[3] == nil || len(this.edges[3]) == 0 {
        this.edges[3] = make([]byte, len(this.contents))
        for i := 0; i < len(this.contents); i++ {
            this.edges[3][i] = this.contents[i][0]
        }
    }
}

func (this *pictureTile) getEdges() [4][]byte {
    this.ensureEdges()
    return this.edges
}

func (this *pictureTile) addNeighbor(neighbor int) {
    this.neighbors = append(this.neighbors, neighbor)
}

func (this *pictureTile) String() string {
    var sb strings.Builder
    fmt.Fprintf(&sb, "Tile %d:\n", this.number)
    for _, line := range this.contents {
        fmt.Fprintf(&sb, "[%s]\n", string(line))
    }
    return sb.String()
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
    tiles, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("There are %d tiles.\n", len(tiles))
    findNeighbors(tiles)
    corners := getCorners(tiles)
    fmt.Printf("There are %d possible corner pieces: %v\n", len(corners), corners)
    answer := multiply(corners)
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

func parseInput(input string) ([]pictureTile, error) {
    lines := strings.Split(input, "\n")
    retval := []pictureTile{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    tileLineRegex := regexp.MustCompile("^Tile ([[:digit:]]+):")
    tileNumber := 0
    tileContents := [][]byte{}
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            tileLineMatch := tileLineRegex.FindStringSubmatch(line)
            if tileLineMatch != nil {
                tileNum, err := strconv.Atoi(tileLineMatch[1])
                if err != nil {
                    return nil, err
                }
                tileNumber = tileNum
            } else {
                tileContents = append(tileContents, []byte(line))
            }
        } else {
            retval = append(retval, pictureTile{number: tileNumber, contents: tileContents})
            tileContents = [][]byte{}
        }
    }
    if len(tileContents) != 0 {
        retval = append(retval, pictureTile{number: tileNumber, contents: tileContents})
    }
    return retval, nil
}

func areEqual(l1 []byte, l2 []byte) bool {
    if len(l1) != len(l2) {
        return false
    }
    for i, b1 := range l1 {
        if l2[i] != b1 {
            return false
        }
    }
    return true
}

func areReversies(l1 []byte, l2 []byte) bool {
    if len(l1) != len(l2) {
        return false
    }
    for i, b1 := range l1 {
        if l2[len(l1)-i-1] != b1 {
            return false
        }
    }
    return true
}

func canBeNeighbors(l1 []byte, l2 []byte) bool {
    return areEqual(l1, l2) || areReversies(l1, l2)
}

func findNeighbors(tiles []pictureTile) {
    for i1 := 0; i1 < len(tiles) - 1; i1++ {
        tile1 := tiles[i1]
        for i2 := i1+1; i2 < len(tiles); i2++ {
            tile2 := tiles[i2]
            isMatch := false
            for _, edge1 := range tile1.getEdges() {
                for _, edge2 := range tile2.getEdges() {
                    if canBeNeighbors(edge1, edge2) {
                        isMatch = true
                        break
                    }
                }
                if isMatch {
                    break
                }
            }
            if isMatch {
                tiles[i1].addNeighbor(tile2.number)
                tiles[i2].addNeighbor(tile1.number)
                // fmt.Printf("Tiles %d and %d can be neighbors!\n", tile1.number, tile2.number)
            } else {
                // fmt.Printf("tiles %d and %d do not go together.\n", tile1.number, tile2.number)
            }
        }
    }
}

func getCorners(tiles []pictureTile) []int {
    retval := []int{}
    for _, tile := range tiles {
        if len(tile.neighbors) == 2 {
            retval = append(retval, tile.number)
        }
    }
    return retval
}

func multiply(nums []int) int {
    retval := 1
    for _, num := range nums {
        retval *= num
    }
    return retval
}
