package main

import (
	"fmt"
	"strings"
)

// main is the main function that gets run for this file.
func main() {
	vals := [][]string{
		{"one", "two", "three", "four"},
		{"five", "six", "seven", "eight", "nine", "ten"},
		{"eleven"},
		{},
		{"twelve", "thirteen"},
		{"fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"},
	}
	colors := []XY{
		Point{0, 0},
		Point{1, 1},
		Point{2, 2},
		Point{3, 3},
		Point{4, 4},
		Point{5, 5},
		Point{0, 10},
		Point{10, 0},
	}
	hl := []XY{
		Point{5, 0},
		Point{4, 1},
		Point{3, 2},
		Point{2, 3},
		Point{1, 4},
		Point{0, 5},
		Point{0, 0},
		Point{4, 4},
		Point{0, 10},
		Point{10, 0},
	}
	fmt.Println(CreateIndexedGridString(vals, colors, hl))
}

type Point struct {
	X int
	Y int
}

func (p Point) GetX() int {
	return p.X
}

func (p Point) GetY() int {
	return p.Y
}

type XY interface {
	GetX() int
	GetY() int
}

func CreateIndexedGridString(vals [][]string, colorPoints []XY, highlightPoints []XY) string {
	// Get the height. If it's zero, there's nothing to return.
	height := len(vals)
	if height == 0 {
		return ""
	}
	// Get the max cell length and the max row width.
	cellLen := 0
	width := len(vals[0])
	for _, r := range vals {
		if len(r) > width {
			width = len(r)
		}
		for _, c := range r {
			if len(c) > cellLen {
				cellLen = len(c)
			}
		}
	}
	leadFmt := fmt.Sprintf("%%%dd:", len(fmt.Sprintf("%d", height)))
	var rv strings.Builder
	// If none of the rows have anything, just print out the row numbers.
	if width == 0 {
		for y := range vals {
			rv.WriteString(fmt.Sprintf(leadFmt, y))
			rv.WriteByte('\n')
		}
		return rv.String()
	}
	// Create a matrix indicating desired text formats.
	textFmt := make([][]int, height)
	for y := range textFmt {
		textFmt[y] = make([]int, width)
	}
	for _, p := range colorPoints {
		if p.GetY() < height && p.GetX() < width {
			textFmt[p.GetY()][p.GetX()] = 1
		}
	}
	for _, p := range highlightPoints {
		if p.GetY() < height && p.GetX() < width {
			textFmt[p.GetY()][p.GetX()] += 2
		}
	}
	// Create the index numbers accross the top.
	if cellLen > 1 {
		cellLen++
	}
	cellFmt := fmt.Sprintf("%%%ds", cellLen)
	blankLead := strings.Repeat(" ", len(fmt.Sprintf(leadFmt, 0)))
	topIndexLines := CreateTopIndexLines(width, cellLen)
	for _, l := range topIndexLines {
		rv.WriteString(fmt.Sprintf("%s%s\n", blankLead, l))
	}
	// Add all the line numbers, cells and extra formatting.
	for y, r := range vals {
		rv.WriteString(fmt.Sprintf(leadFmt, y))
		for x := 0; x < width; x++ {
			v := ""
			if x < len(r) {
				v = r[x]
			}
			cell := fmt.Sprintf(cellFmt, v)
			switch textFmt[y][x] {
			case 1: // color only
				rv.WriteString("\033[32m" + cell + "\033[0m") // Green text
			case 2: // highlight only
				rv.WriteString("\033[7m" + cell + "\033[0m") // Reversed (grey back, black text)
			case 3: // color and highlight
				rv.WriteString("\033[97;42m" + cell + "\033[0m") // Green background, white text
			default:
				rv.WriteString(cell)
			}
		}
		rv.WriteByte('\n')
	}
	return rv.String()
}

func CreateTopIndexLines(count, cellLen int) []string {
	rv := []string{}
	if count > 100 {
		rv = append(rv, CreateIndexLinesHundreds(count, cellLen))
	}
	if count > 10 {
		rv = append(rv, CreateIndexLinesTens(count, cellLen))
	}
	if count > 0 {
		rv = append(rv, CreateIndexLineOnes(count, cellLen))
		rv = append(rv, strings.Repeat("-", count*cellLen))
	}
	return rv
}

func CreateIndexLineOnes(count, cellLen int) string {
	cellFmt := fmt.Sprintf("%%%dd", cellLen)
	digits := fmt.Sprintf(strings.Repeat(cellFmt, 10), 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	rv := strings.Repeat(digits, 1+count/10)
	return rv[:count*cellLen]
}

func CreateIndexLinesTens(count, cellLen int) string {
	cellFmt := fmt.Sprintf("%%%dc", cellLen)
	var digits strings.Builder
	for _, s := range []string{" ", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		digits.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), 10))
	}
	rv := strings.Repeat(digits.String(), 1+count/100)
	return rv[:count*cellLen]
}

func CreateIndexLinesHundreds(count, cellLen int) string {
	cellFmt := fmt.Sprintf("%%%dc", cellLen)
	var digits strings.Builder
	for _, s := range []string{" ", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		digits.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), 100))
	}
	rv := strings.Repeat(digits.String(), 1+count/1000)
	return rv[:count*cellLen]
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d"
// DigitFormatForMax(382920) returns "%6d"
func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
}
