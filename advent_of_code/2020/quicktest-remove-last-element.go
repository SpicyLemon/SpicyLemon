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
    array := []string{}
    array = append(array, "zero")
    array = append(array, "one")
    array = append(array, "two")
    array = append(array, "three")
    fmt.Println("element 0:", array[0])
    fmt.Println("element 1:", array[1])
    fmt.Println("element 2:", array[2])
    fmt.Println("element 3:", array[3])
    fmt.Println("initial:", array)
    fmt.Println(" length:", len(array))
    array = array[:3]
    fmt.Println("chopped:", array)
    fmt.Println(" length:", len(array))
    return nil
}

