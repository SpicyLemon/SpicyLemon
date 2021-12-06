package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strings"
)

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
    fmt.Printf("The input has %d entries.\n", len(input))
    fmt.Printf("Seats: %s\n", gridString(input))
    i1 := iterate(input)
    fmt.Printf("Step 1: %s\n", gridString(i1))
    i2 := iterate(i1)
    fmt.Printf("Step 2: %s\n", gridString(i2))
    iterations, final := countIterationsToStatic(input)
    fmt.Printf("Final (Step %d+): %s\n", iterations, gridString(final))
    answer := countTakenSeats(final)
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

func parseInput(input string) ([][]byte, error) {
    lines := strings.Split(input, "\n")
    retval := [][]byte{}
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for _, line := range lines {
        if hasDataRegex.MatchString(line) {
            retval = append(retval, []byte(line))
        }
    }
    return retval, nil
}

func iterate(grid [][]byte) [][]byte {
    h := len(grid)
    w := len(grid[0])
    retval := make([][]byte, h)
    for r := range(grid) {
        retval[r] = make([]byte, w)
        for c, spot := range grid[r] {
            newval := spot
            switch spot {
            case 'L':
                if countVisibleTaken(grid, r, c) == 0 {
                    newval = '#'
                }
            case '#':
                if countVisibleTaken(grid, r, c) >= 5 {
                    newval = 'L'
                }
            }
            retval[r][c] = newval
        }
    }
    return retval
}

func countVisibleTaken(grid [][]byte, row int, col int) int {
    h := len(grid)
    w := len(grid[0])
    retval := 0
    for i := -1; i <= 1; i++ {
        for j := -1; j <= 1; j++ {
            if i == 0 && j == 0 {
                j++
            }
            d := 1
            for {
                r := row + i*d
                c := col + j*d
                if (r < 0 || r >= h || c < 0 || c >= w) {
                    break
                }
                spot := grid[r][c]
                if spot == '.' {
                    d += 1
                } else if spot == '#' {
                    retval += 1
                    break
                } else if spot == 'L' {
                    break
                }
            }
        }
    }
    return retval
}

func countAdjTaken(grid [][]byte, row int, col int) int {
    retval := 0
    for i := -1; i <= 1; i++ {
        for j := -1; j <= 1; j++ {
            if i == 0 && j == 0 {
                j++
            }
            if grid[row+i][col+j] == '#' {
                retval += 1
            }
        }
    }
    return retval
}

func countIterationsToStatic(initialGrid [][]byte) (int, [][]byte) {
    previous := initialGrid
    i := 0
    keepGoing := true
    for keepGoing {
        i += 1
        next := iterate(previous)
        if areEqualGrids(previous, next) {
            keepGoing = false
        } else {
            previous = next
        }
    }
    return i-1, previous
}

func countTakenSeats(grid [][]byte) int {
    retval := 0
    for _, row := range(grid) {
        for _, spot := range(row) {
            if spot == '#' {
                retval += 1
            }
        }
    }
    return retval
}

func areEqualGrids(grid1 [][]byte, grid2 [][]byte) bool {
    if len(grid1) != len(grid2) {
        return false
    }
    for c1, v1 := range(grid1) {
        if ! areEqualRows(v1, grid2[c1]) {
            return false
        }
    }
    return true
}

func areEqualRows(row1 []byte, row2 []byte) bool {
    if len(row1) != len(row2) {
        return false
    }
    for r1, v1 := range(row1) {
        if v1 != row2[r1] {
            return false
        }
    }
    return true
}

func gridString(nums [][]byte) string {
    var sb strings.Builder
    sb.WriteString("[\n")
    for _, arr := range nums {
        fmt.Fprintf(&sb, "  %s\n", arr)
    }
    sb.WriteString("]")
    return sb.String()
}
