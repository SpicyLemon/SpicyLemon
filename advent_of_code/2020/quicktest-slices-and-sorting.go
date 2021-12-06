package main

import (
    "fmt"
    "os"
    "sort"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    array := []int{16, 8, 14, 6, 19, 10, 12, 17, 4, 20, 0, 11, 5, 15, 1, 7, 13, 9, 2, 18, 3}
    fmt.Println("Initial:", array)
    fmt.Println("Calling outputSorted on slice with first 5 elements.")
    outputSorted(array[0:5])
    fmt.Println("After call:", array)
    fmt.Println("Calling outputSortedSafe on slice with last 5 elements.")
    outputSortedSafe(array[16:21])
    fmt.Println("After call:", array)
    return nil
}

func outputSorted(nums []int) {
    sort.Ints(nums)
    fmt.Println("sorted:", nums)
}

func outputSortedSafe(nums []int) {
    sorted := make([]int, len(nums))
    copy(sorted, nums)
    sort.Ints(sorted)
    fmt.Println("sorted:", sorted)
}

