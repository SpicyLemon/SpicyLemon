package main

import (
    "fmt"
    "os"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    arr := [4]int{0, 1, 2, 3}
    fmt.Printf("Initial: %v\n", arr)
    arr[0], arr[1], arr[2], arr[3] = arr[1], arr[2], arr[3], arr[0]
    fmt.Printf("Shift Left: %v\n", arr)
    arr[0], arr[1], arr[2], arr[3] = arr[3], arr[2], arr[1], arr[0]
    fmt.Printf("Reversed: %v\n", arr)
    arr[0], arr[1], arr[2], arr[3] = arr[0] + 1, arr[1] + 1, arr[2] + 1, arr[3] + 1
    fmt.Printf("Plus 1: %v\n", arr)
    return nil
}

