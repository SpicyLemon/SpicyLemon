package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
)

const TOP = 0
const RIGHT = 1
const BOTTOM = 2
const LEFT = 3

type coord struct {
    x int
    y int
}

type pictureTile struct {
    number int
    contents [][]byte
    neighbors [4]int
}

func NewPictureTile(number int, contents [][]byte) pictureTile {
    this := pictureTile{}
    this.number = number
    this.contents = contents
    this.neighbors = [4]int{0, 0, 0, 0}
    return this
}

func (this *pictureTile) getTopEdge() []byte {
    return getCopy(this.contents[0])
}

func (this *pictureTile) getRightEdge() []byte {
    last := len(this.contents[0]) - 1
    retval := make([]byte, len(this.contents))
    for i := 0; i < len(this.contents); i++ {
        retval[i] = this.contents[i][last]
    }
    return retval
}

func (this *pictureTile) getBottomEdge() []byte {
    return getCopy(this.contents[len(this.contents) - 1])
}

func (this *pictureTile) getLeftEdge() []byte {
    retval := make([]byte, len(this.contents))
    for i := 0; i < len(this.contents); i++ {
        retval[i] = this.contents[i][0]
    }
    return retval
}

func (this *pictureTile) getEdge(side int) []byte {
    switch side {
    case TOP:
        return this.getTopEdge()
    case RIGHT:
        return this.getRightEdge()
    case BOTTOM:
        return this.getBottomEdge()
    case LEFT:
        return this.getLeftEdge()
    }
    return []byte{}
}

func (this *pictureTile) getEdges() [4][]byte {
    retval := [4][]byte{}
    retval[TOP] = this.getTopEdge()
    retval[RIGHT] = this.getRightEdge()
    retval[BOTTOM] = this.getBottomEdge()
    retval[LEFT] = this.getLeftEdge()
    return retval
}

func (this *pictureTile) rotateTileClockwise() {
    this.contents = rotateClockwise(this.contents)
    this.neighbors[RIGHT], this.neighbors[BOTTOM], this.neighbors[LEFT], this.neighbors[TOP] = this.neighbors[TOP], this.neighbors[RIGHT], this.neighbors[BOTTOM], this.neighbors[LEFT]
}

func (this *pictureTile) rotateTileCounterClockwise() {
    this.contents = rotateCounterClockwise(this.contents)
    this.neighbors[LEFT], this.neighbors[TOP], this.neighbors[RIGHT], this.neighbors[BOTTOM] = this.neighbors[TOP], this.neighbors[RIGHT], this.neighbors[BOTTOM], this.neighbors[LEFT]
}

func (this *pictureTile) rotateTile180() {
    this.contents = rotate180(this.contents)
    this.neighbors[BOTTOM], this.neighbors[LEFT], this.neighbors[TOP], this.neighbors[RIGHT] = this.neighbors[TOP], this.neighbors[RIGHT], this.neighbors[BOTTOM], this.neighbors[LEFT]
}

func (this *pictureTile) flipTileHorizontally() {
    this.contents = flipHorizontally(this.contents)
    this.neighbors[LEFT], this.neighbors[RIGHT] = this.neighbors[RIGHT], this.neighbors[LEFT]
}

func (this *pictureTile) flipTileVertically() {
    this.contents = flipVertically(this.contents)
    this.neighbors[TOP], this.neighbors[BOTTOM] = this.neighbors[BOTTOM], this.neighbors[TOP]
}

func (this *pictureTile) addNeighbor(side int, neighbor int) {
    fmt.Printf("Tile %d's %s neighbor is tile %d\n", this.number, sideString(side), neighbor)
    this.neighbors[side] = neighbor
}

func (this *pictureTile) countNeighbors() int {
    retval := 0
    for _, n := range this.neighbors {
        if n > 0 {
            retval += 1
        }
    }
    return retval
}

func (this *pictureTile) getEdgeForNeighbor(neighbor int) ([]byte, int) {
    for i, n := range this.neighbors {
        if n == neighbor {
            return this.getEdge(i), i
        }
    }
    return []byte{}, -1
}

func (this *pictureTile) getPicture() [][]byte {
    retval := [][]byte{}
    for i := 1; i < len(this.contents) - 1; i++ {
        retval = append(retval, this.contents[i][1:len(this.contents[i])-1])
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
    tiles, err := parseInput(string(dat))
    if (err != nil) {
        return err
    }
    fmt.Printf("There are %d tiles.\n", len(tiles))
    findNeighbors(tiles)
    corners, sides, middles := sortTiles(tiles)
    fmt.Printf("Corner tiles: %v\n", corners)
    fmt.Printf("Side tiles: %v\n", sides)
    fmt.Printf("Middle tiles: %v\n", middles)
    picture := toPicture(tiles)
    fmt.Printf("Picture:\n%s\n", matrixString(picture))
    startingPounds := countPounds(picture)
    withMonsters, monsterCount := findMonsters(picture)
    fmt.Printf("Picture:\n%s\n", matrixString(withMonsters))
    fmt.Printf("There are %d monsters.\n", monsterCount)
    answer := countPounds(withMonsters)
    fmt.Printf("Answer: %d = %d - 15 * %d\n", answer, startingPounds, monsterCount)
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
            retval = append(retval, NewPictureTile(tileNumber, tileContents))
            tileContents = [][]byte{}
        }
    }
    if len(tileContents) != 0 {
        retval = append(retval, NewPictureTile(tileNumber, tileContents))
    }
    return retval, nil
}

func monsterRight() []coord {
    return []coord{
        coord{y: 0, x: 18},
        coord{y: 1, x: 0},
        coord{y: 1, x: 5},
        coord{y: 1, x: 6},
        coord{y: 1, x: 11},
        coord{y: 1, x: 12},
        coord{y: 1, x: 17},
        coord{y: 1, x: 18},
        coord{y: 1, x: 19},
        coord{y: 2, x: 1},
        coord{y: 2, x: 4},
        coord{y: 2, x: 7},
        coord{y: 2, x: 10},
        coord{y: 2, x: 13},
        coord{y: 2, x: 16}}
}

func monsterLeft() []coord {
    return []coord{
        coord{y: 0, x: 1},
        coord{y: 1, x: 0},
        coord{y: 1, x: 1},
        coord{y: 1, x: 2},
        coord{y: 1, x: 7},
        coord{y: 1, x: 8},
        coord{y: 1, x: 13},
        coord{y: 1, x: 18},
        coord{y: 1, x: 19},
        coord{y: 2, x: 3},
        coord{y: 2, x: 6},
        coord{y: 2, x: 9},
        coord{y: 2, x: 12},
        coord{y: 2, x: 15},
        coord{y: 2, x: 18}}
}

func testMutations() {
    sample := make([][]byte, 5)
    sample[0] = []byte("12345")
    sample[1] = []byte("abcde")
    sample[2] = []byte("lmnop")
    sample[3] = []byte("vwxyz")
    sample[4] = []byte(".-~_+")
    fmt.Printf("Sample:\n%s\n", matrixString(sample))
    rot90 := rotateClockwise(sample)
    rot180 := rotate180(sample)
    rot270 := rotateCounterClockwise(sample)
    fliph := flipHorizontally(sample)
    flipv := flipVertically(sample)
    fmt.Printf("Sample:\n%s\n", matrixString(sample))
    fmt.Printf("Rot 90:\n%s\n", matrixString(rot90))
    fmt.Printf("Rot 180:\n%s\n", matrixString(rot180))
    fmt.Printf("Rot 270:\n%s\n", matrixString(rot270))
    fmt.Printf("Flip H:\n%s\n", matrixString(fliph))
    fmt.Printf("Flip V:\n%s\n", matrixString(flipv))
    fmt.Println(sideBySideString(sample, rot90))
    fmt.Println(sideBySideString(sample, rot180))
    fmt.Println(sideBySideString(sample, rot270))
    fmt.Println(sideBySideString(sample, fliph))
    fmt.Println(sideBySideString(sample, flipv))
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
            if ! containsInt(tile1.neighbors[0:], tile2.number) {
                tile1Side := -1
                tile2Side := -1
                isMatch := false
                for e1, edge1 := range tile1.getEdges() {
                    for e2, edge2 := range tile2.getEdges() {
                        if canBeNeighbors(edge1, edge2) {
                            isMatch = true
                            tile1Side = e1
                            tile2Side = e2
                            break
                        }
                    }
                    if isMatch {
                        break
                    }
                }
                if isMatch {
                    tiles[i1].addNeighbor(tile1Side, tile2.number)
                    tiles[i2].addNeighbor(tile2Side, tile1.number)
                }
            }
        }
    }
}

func toPicture(tiles []pictureTile) [][]byte {
    tileMap := make(map[int]pictureTile)
    for _, tile := range tiles {
        tileMap[tile.number] = tile
    }
    corners, _, _ := sortTiles(tiles)
    picture := [][]pictureTile{}
    picture = append(picture, []pictureTile{})
    // Get the top right corner oriented.
    topLeftCorner := tileMap[corners[0]]
    for topLeftCorner.neighbors[RIGHT] == 0 || topLeftCorner.neighbors[BOTTOM] == 0 {
        fmt.Printf("Top left corner [%d] neighbors: %v. Rotating clockwise.\n", topLeftCorner.number, topLeftCorner.neighbors)
        topLeftCorner.rotateTileClockwise()
    }
    fmt.Printf("Top left corner [%d] neighbors: %v.\n", topLeftCorner.number, topLeftCorner.neighbors)
    row := 0
    picture[row] = append(picture[row], topLeftCorner)
    for {
        // put together the rest of this row
        for {
            tile := picture[row][len(picture[row])-1]
            rightEdge := tile.getRightEdge()
            nextTile := tileMap[tile.neighbors[RIGHT]]
            nextTileEdge, nextTileEdgeSide := nextTile.getEdgeForNeighbor(tile.number)
            fmt.Printf("%d  RIGHT: [%s]\n", tile.number, string(rightEdge))
            fmt.Printf("%d % 6s: [%s]", nextTile.number, sideString(nextTileEdgeSide), string(nextTileEdge))
            switch nextTileEdgeSide {
            case TOP:
                nextTile.rotateTileCounterClockwise()
                fmt.Printf(", rotating counter clockwise -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                if areEqual(rightEdge, nextTileEdge) {
                    nextTile.flipTileVertically()
                    fmt.Printf(", flipping vertically -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                }
            case RIGHT:
                if areEqual(rightEdge, nextTileEdge) {
                    nextTile.flipTileHorizontally()
                    fmt.Printf(", flipping horizontally -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                } else {
                    nextTile.rotateTile180()
                    fmt.Printf(", rotating 180 -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                }
            case BOTTOM:
                nextTile.rotateTileClockwise()
                fmt.Printf(", rotating clockwise -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                if areReversies(rightEdge, nextTileEdge) {
                    nextTile.flipTileVertically()
                    fmt.Printf(", flipping vertically -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                }
            case LEFT:
                if areReversies(rightEdge, nextTileEdge) {
                    nextTile.flipTileVertically()
                    fmt.Printf(", flipping vertically -> LEFT:[%s]", string(nextTile.getLeftEdge()))
                }
            }
            fmt.Printf(" -- LEFT:[%s]\n", string(nextTile.getLeftEdge()))
            picture[row] = append(picture[row], nextTile)
            if nextTile.neighbors[RIGHT] == 0 {
                break
            }
        }
        topTile := picture[row][0]
        // if the left-most piece of this row has no bottom neighbor, we're done.
        if topTile.neighbors[BOTTOM] <= 0 {
            break
        }
        // Create the new row and align the first piece in it.
        topTileBottomEdge := topTile.getBottomEdge()
        newRowTile := tileMap[topTile.neighbors[BOTTOM]]
        newRowTileEdge, newRowTileEdgeSide := newRowTile.getEdgeForNeighbor(topTile.number)
        fmt.Printf("%d BOTTOM: [%s]\n", topTile.number, string(topTileBottomEdge))
        fmt.Printf("%d % 6s: [%s]", newRowTile.number, sideString(newRowTileEdgeSide), string(newRowTileEdge))
        switch newRowTileEdgeSide {
        case TOP:
            if areReversies(topTileBottomEdge, newRowTileEdge) {
                newRowTile.flipTileHorizontally()
                fmt.Printf(", flipping horizontally -> TOP:[%s]", string(newRowTile.getTopEdge()))
            }
        case RIGHT:
            newRowTile.rotateTileCounterClockwise()
            fmt.Printf(", rotating counter clockwise -> TOP:[%s]", string(newRowTile.getTopEdge()))
            if areReversies(topTileBottomEdge, newRowTileEdge) {
                newRowTile.flipTileHorizontally()
                fmt.Printf(", flipping horizontaly -> TOP:[%s]", string(newRowTile.getTopEdge()))
            }
        case BOTTOM:
            if areEqual(topTileBottomEdge, newRowTileEdge) {
                newRowTile.flipTileVertically()
                fmt.Printf(", flipping vertically -> TOP:[%s]", string(newRowTile.getTopEdge()))
            } else {
                newRowTile.rotateTile180()
                fmt.Printf(", rotating 180 -> TOP:[%s]", string(newRowTile.getTopEdge()))
            }
        case LEFT:
            newRowTile.rotateTileClockwise()
            fmt.Printf(", rotating clockwise -> TOP:[%s]", string(newRowTile.getTopEdge()))
            if areEqual(topTileBottomEdge, newRowTileEdge) {
                newRowTile.flipTileHorizontally()
                fmt.Printf(", flipping horizontally -> TOP:[%s]", string(newRowTile.getTopEdge()))
            }
        }
        fmt.Printf(" -- TOP:[%s]\n", string(newRowTile.getTopEdge()))
        picture = append(picture, []pictureTile{})
        row += 1
        picture[row] = append(picture[row], newRowTile)
    }

    tileWidth := len(tiles[0].contents) - 2
    retval := make([][]byte, len(picture) * tileWidth)
    for i := range picture {
        for j := 0; j < tileWidth; j++ {
            retval[i*tileWidth+j] = make([]byte, len(picture[i])*tileWidth)
        }
    }
    for i, tileRow := range picture {
        for j, tile := range tileRow {
            subPic := tile.getPicture()
            for si, subPicRow := range subPic {
                for sj, c := range subPicRow {
                    retval[i*tileWidth+si][j*tileWidth+sj] = c
                }
            }
        }
    }
    return retval
}

func findMonsters(picture [][]byte) ([][]byte, int) {
    monsterWidth := 20
    monsterHeight := 3
    monster := monsterRight()
    pictures := make([][][]byte, 8)
    pictures[0] = picture
    pictures[1] = rotateClockwise(picture)
    pictures[2] = rotate180(picture)
    pictures[3] = rotateCounterClockwise(picture)
    pictures[4] = flipHorizontally(pictures[0])
    pictures[5] = flipHorizontally(pictures[1])
    pictures[6] = flipHorizontally(pictures[2])
    pictures[7] = flipHorizontally(pictures[3])
    monsterCounts := make([]int, 8)
    for o := 0; o < 8; o++ {
        pic := pictures[o]
        for i := 0; i < len(pic) - monsterHeight; i++ {
            for j := 0; j < len(pic[i]) - monsterWidth; j++ {
                monsterFound := true
                for _, c := range monster {
                    fmt.Printf("o:[%d], i:[%d], c.y:[%d], i+c.y:[%d], j:[%d], c.x:[%d], j+c.x[%d]\n", o, i, c.y, i+c.y, j, c.x, j+c.x)
                    if pic[i+c.y][j+c.x] != '#' {
                        monsterFound = false
                        break
                    }
                }
                if monsterFound {
                    monsterCounts[o] += 1
                    for _, c := range monster {
                        pic[i+c.y][j+c.x] = 'O'
                    }
                }
            }
        }
    }
    fmt.Printf("Unchanged: %d\n", monsterCounts[0])
    fmt.Printf("   rot 90: %d\n", monsterCounts[1])
    fmt.Printf("  rot 180: %d\n", monsterCounts[2])
    fmt.Printf("  rot 270: %d\n", monsterCounts[3])
    fmt.Printf("  flipped: %d\n", monsterCounts[4])
    fmt.Printf(" rot 90 f: %d\n", monsterCounts[5])
    fmt.Printf("rot 180 f: %d\n", monsterCounts[6])
    fmt.Printf("rot 270 f: %d\n", monsterCounts[7])
    for o, monsterCount := range monsterCounts {
        if monsterCount > 0 {
            return pictures[o], monsterCount
        }
    }
    return picture, 0
}

func countPounds(picture [][]byte) int {
    retval := 0
    for _, row := range picture {
        for _, c := range row {
            if c == '#' {
                retval += 1
            }
        }
    }
    return retval
}

func sortTiles(tiles []pictureTile) ([]int, []int, []int) {
    corners := []int{}
    sides := []int{}
    middles := []int{}
    for _, tile := range tiles {
        switch tile.countNeighbors() {
        case 2:
            corners = append(corners, tile.number)
        case 3:
            sides = append(sides, tile.number)
        case 4:
            middles = append(middles, tile.number)
        default:
            panic(fmt.Errorf("Tile %d has %d neighbor(s)!", tile.number, tile.countNeighbors()))
        }
    }
    return corners, sides, middles
}

func rotateClockwise(image [][]byte) [][]byte {
    retval := makeEmptyCopy(image)
    l := len(image)
    for i := 0; i < l; i++ {
        for j := 0; j < l; j++ {
            retval[i][j] = image[l-j-1][i]
        }
    }
    return retval
}

func rotateCounterClockwise(image [][]byte) [][]byte {
    retval := makeEmptyCopy(image)
    l := len(image)
    for i := 0; i < l; i++ {
        for j := 0; j < l; j++ {
            retval[i][j] = image[j][l-i-1]
        }
    }
    return retval
}

func rotate180(image [][]byte) [][]byte {
    retval := makeEmptyCopy(image)
    l := len(image)
    for i := 0; i < l; i++ {
        for j := 0; j < l; j++ {
            retval[i][j] = image[l-i-1][l-j-1]
        }
    }
    return retval
}

func flipVertically(image [][]byte) [][]byte {
    retval := makeEmptyCopy(image)
    l := len(image)
    for i := 0; i < l; i++ {
        for j := 0; j < l; j++ {
            retval[i][j] = image[l-i-1][j]
        }
    }
    return retval
}

func flipHorizontally(image [][]byte) [][]byte {
    retval := makeEmptyCopy(image)
    l := len(image)
    for i := 0; i < l; i++ {
        for j := 0; j < l; j++ {
            retval[i][j] = image[i][l-j-1]
        }
    }
    return retval
}

func makeEmptyCopy(image [][]byte) [][]byte {
    retval := make([][]byte, len(image))
    for i := range image {
        retval[i] = make([]byte, len(image[i]))
    }
    return retval
}

func matrixString(tile [][]byte) string {
    var sb strings.Builder
    for _, row := range tile {
        fmt.Fprintf(&sb, "[%s]\n", string(row))
    }
    return sb.String()
}

func sideBySideString(tile1 [][]byte, tile2 [][]byte) string {
    var sb strings.Builder
    for i := range tile1 {
        fmt.Fprintf(&sb, "[%s]    [%s]\n", string(tile1[i]), string(tile2[i]))
    }
    return sb.String()
}

func containsInt(nums []int, number int) bool {
    for _, num := range nums {
        if num == number {
            return true
        }
    }
    return false
}

func containsTileNumber(tiles []pictureTile, number int) bool {
    for _, tile := range tiles {
        if tile.number == number {
            return true
        }
    }
    return false
}

func sideString(side int) string {
    switch side {
    case TOP: return "TOP"
    case RIGHT: return "RIGHT"
    case BOTTOM: return "BOTTOM"
    case LEFT: return"LEFT"
    }
    return "UNKNOWN"
}

func getCopy(orig []byte) []byte {
    retval := make([]byte, len(orig))
    copy(retval, orig)
    return retval
}

