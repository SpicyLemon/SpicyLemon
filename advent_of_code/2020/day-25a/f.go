package main

import (
    "fmt"
    "os"
    "strconv"
)

func findLoopSize(result int) int {
    retval := 0
    val := 1
    subjectNumber := 7
    for val != result {
        val = val * subjectNumber % 20201227
        retval += 1
    }
    return retval
}

func transform(subjectNumber int, loopSize int) int {
    retval := 1
    for i := 0; i < loopSize; i++ {
        retval = retval * subjectNumber % 20201227
    }
    return retval
}

func fStraight(loopSize int, subjectNumber int, magicNumber int) []int {
    retval := make([]int, loopSize + 1)
    retval[0] = 1
    for n := 1; n <= loopSize; n++ {
        retval[n] = (retval[n-1] * subjectNumber) % magicNumber
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

func getCliParams() (int, int, int, error) {
    args := os.Args[1:]
    if len(args) < 1 || len(args) > 3 {
        return 0, 0, 0, fmt.Errorf("Invalid command-line arguments: %s. Expecting [loopSize (subjectNumber = 7) (magicNumber = 20201227)]", args)
    }
    loopSize, err := strconv.Atoi(args[0])
    if err != nil {
        return 0, 0, 0, err
    }
    subjectNumber := 7
    if len(args) >= 2 {
        subjectNumber, err = strconv.Atoi(args[1])
        if err != nil {
            return 0, 0, 0, err
        }
    }
    magicNumber := 20201227
    if len(args) >= 3 {
        magicNumber, err = strconv.Atoi(args[2])
        if err != nil {
            return 0, 0, 0, err
        }
    }
    return loopSize, subjectNumber, magicNumber, nil
}

func run() error {
    loopSize, subjectNumber, magicNumber, err := getCliParams()
    if err != nil {
        return err
    }
    vals := fStraight(loopSize, subjectNumber, magicNumber)
    lineFormat := fmt.Sprintf("f(%%%dd, %%d, %%d) = %%d\n", len(strconv.Itoa(loopSize)))
    for n, val := range vals {
        fmt.Printf(lineFormat, n, subjectNumber, magicNumber, val)
    }
    return nil
}
