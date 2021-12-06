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
    str := "This is a string"
    strThis := str[0:4]
    strIsA := str[5:9]
    strString := str[len(str)-6:]
    strThisIsA := str[:len(str)-7]
    fmt.Printf("Full string: [%s].\n", str)
    fmt.Printf("  This: [%s].\n", strThis)
    fmt.Printf("  is a: [%s].\n", strIsA)
    fmt.Printf("string: [%s].\n", strString)
    fmt.Printf("This is a: [%s].\n", strThisIsA)
    str2b := []byte(str)
    str2b[4] = '-'
    str2 := string(str2b)
    fmt.Printf("After change: [%s].\n", str2)
    return nil
}

