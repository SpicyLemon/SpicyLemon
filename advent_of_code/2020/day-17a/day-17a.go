package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "sort"
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
    fmt.Printf("The input is %d by %d by %d.\n", len(input), len(input[0]), len(input[0][0]))
    fmt.Printf("Initial State:\n%s", spaceString(input))
    cycle6 := doCycles(input, 6)
    active, inactive := countSpace(cycle6)
    fmt.Printf("active: %d, inactive: %d\n", active, inactive)
    answer := active
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

func parseInput(input string) (map[int]map[int]map[int]bool, error) {
    lines := strings.Split(input, "\n")
    retval := make(map[int]map[int]map[int]bool)
    retval[0] = make(map[int]map[int]bool)
    hasDataRegex := regexp.MustCompile("[^[:space:]]")
    for y, line := range lines {
        retval[0][y] = make(map[int]bool)
        if hasDataRegex.MatchString(line) {
            for x, c := range line {
                if c == '#' {
                    retval[0][y][x] = true
                } else if c == '.' {
                    retval[0][y][x] = false
                } else {
                    return nil, fmt.Errorf("Unknown character in input: [%s].", string(c))
                }
            }
        }
    }
    return retval, nil
}

func spaceString(space map[int]map[int]map[int]bool) string {
    zs, ys, xs := keys(space)
    var sb strings.Builder
    for _, z := range zs {
        fmt.Fprintf(&sb, "z: %d\n", z)
        for _, y := range ys {
            for _, x := range xs {
                if space[z][y][x] {
                    sb.WriteByte('#')
                } else {
                    sb.WriteByte('.')
                }
            }
            sb.WriteString("\n")
        }
        sb.WriteString("\n")
    }
    return sb.String()
}

func keys(m map[int]map[int]map[int]bool) ([]int, []int, []int) {
    zs := keysZ(m)
    ys := keysY(m[zs[0]])
    xs := keysX(m[zs[0]][ys[0]])
    return zs, ys, xs
}

func keysX(m map[int]bool) []int {
    retval := make([]int, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    sort.Ints(retval)
    return retval
}

func keysY(m map[int]map[int]bool) []int {
    retval := make([]int, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    sort.Ints(retval)
    return retval
}

func keysZ(m map[int]map[int]map[int]bool) []int {
    retval := make([]int, len(m))
    i := 0
    for key := range m {
        retval[i] = key
        i++
    }
    sort.Ints(retval)
    return retval
}

func countSpace(space map[int]map[int]map[int]bool) (int, int) {
    zs, ys, xs := keys(space)
    active := 0
    inactive := 0
    for _, z := range zs {
        for _, y := range ys {
            for _, x := range xs {
                if space[z][y][x] {
                    active += 1
                } else {
                    inactive += 1
                }
            }
        }
    }
    return active, inactive
}

func min(l []int) int {
    retval := l[0]
    for _, n := range l {
        if n < retval {
            retval = n
        }
    }
    return retval
}

func max(l []int) int {
    retval := l[0]
    for _, n := range l {
        if n > retval {
            retval = n
        }
    }
    return retval
}

func getNextCyle(space map[int]map[int]map[int]bool) map[int]map[int]map[int]bool {
    zs, ys, xs := keys(space)
    retval := make(map[int]map[int]map[int]bool)
    zs = append(zs, min(zs) - 1, max(zs) + 1)
    ys = append(ys, min(ys) - 1, max(ys) + 1)
    xs = append(xs, min(xs) - 1, max(xs) + 1)
    for _, z := range zs {
        retval[z] = make(map[int]map[int]bool)
        for _, y := range ys {
            retval[z][y] = make(map[int]bool)
            for _, x := range xs {
                retval[z][y][x] = getNewState(space, z, y, x)
            }
        }
    }
    return retval
}

func doCycles(space map[int]map[int]map[int]bool, count int) map[int]map[int]map[int]bool {
    retval := space
    for i := 0; i < count; i++ {
        fmt.Printf("Calculating cycle %d.\n", i + 1)
        retval = getNextCyle(retval)
    }
    return retval
}

func getNewState(space map[int]map[int]map[int]bool, z int, y int, x int) bool {
    active := 0
    for k := -1; k <= 1; k++ {
        _, okZ := space[z+k]
        if okZ {
            for j := -1; j <= 1; j++ {
                _, okZY := space[z+k][y+j]
                if okZY {
                    for i := -1; i <= 1; i++ {
                        if (i == 0 && j == 0 && k == 0) {
                            i += 1
                        }
                        state, okZYX := space[z+k][y+j][x+i]
                        if okZYX && state {
                            active += 1
                        }
                    }
                }
            }
        }
    }
    current, currentOk := space[z][y][x]
    if (!currentOk || !current) && active == 3 {
        return true
    }
    if currentOk && current && (active == 2 || active == 3) {
        return true
    }
    return false
}
