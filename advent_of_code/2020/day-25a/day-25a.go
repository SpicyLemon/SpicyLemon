package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
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

func parseInput(input string) (int, int, error) {
    lines := strings.Split(input, "\n")
    doorPubKey, err := strconv.Atoi(lines[0])
    if err != nil {
        return 0, 0, err
    }
    cardPubKey, err := strconv.Atoi(lines[1])
    if err != nil {
        return 0, 0, err
    }
    return doorPubKey, cardPubKey, nil
}

func run() error {
    filename := getCliParams()
    fmt.Printf("Getting input from [%s].\n", filename)
    dat, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    // fmt.Println(string(dat))
    doorPubKey, cardPubKey, err := parseInput(string(dat))
    if err != nil {
        return err
    }
    fmt.Printf("Door public key: %d\n", doorPubKey)
    fmt.Printf("Card public key: %d\n", cardPubKey)
    doorLoopSize := findLoopSize(doorPubKey)
    cardLoopSize := findLoopSize(cardPubKey)
    fmt.Printf("Door loop size: %d\n", doorLoopSize)
    fmt.Printf("Card loop size: %d\n", cardLoopSize)
    answer1 := transform(doorPubKey, cardLoopSize)
    answer2 := transform(cardPubKey, doorLoopSize)
    fmt.Printf("Answer1: %d\n", answer1)
    fmt.Printf("Answer2: %d\n", answer2)
    return nil
}
