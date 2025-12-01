package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// main is the main function that gets run for this file.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var widthGiven, heightGiven bool
	var width, height int
	var colors, hl []XY
	var entryVals []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		thisArg := args[i]
		switch thisArg {
		case "--help", "-h", "help":
			fmt.Printf(`Usage: create-index-grid-string [<width> [<height>]] [{-v|--value} <values>]
                                [{-c|--color} <points>] [{-m|--mark} <points>]

Default <width> is 6.
Default <height> is <width>.

The <values> are the strings to put in the cells of the grid.
They are used in reading order starting in the upper left cell.
Once all <values> are exhausted, the list is repeated.
If no <values> are given, cells will be numbered starting at "0".

The -c or --color flag causes the provided <point> entries to be colored light-blue in the output.
If no -c or --color points are provided, the left column and every other column are colored.

The -m or --mark flag causes the provided <point> entries to have reversed foreground and background.
If no -m or --mark points are provided, the top row and every other row are marked.

A <point> has the format "<x>,<y>" where <x> and <y> are whole numbers.
<points> is one or more <point> entries separated by whitespace.

`)
			return nil
		case "-c", "--color":
			hadPoint := false
			for ; i+1 < len(args) && args[i+1][0] != '-'; i++ {
				for _, arg := range strings.Fields(args[i+1]) {
					p, err := ParsePoint(arg)
					if err != nil {
						return err
					}
					colors = append(colors, p)
					hadPoint = true
				}
			}
			if !hadPoint {
				return fmt.Errorf("no <points> provided after [%s]", thisArg)
			}
		case "-m", "--mark":
			hadPoint := false
			for ; i+1 < len(args) && args[i+1][0] != '-'; i++ {
				for _, arg := range strings.Fields(args[i+1]) {
					p, err := ParsePoint(arg)
					if err != nil {
						return err
					}
					hl = append(hl, p)
					hadPoint = true
				}
			}
			if !hadPoint {
				return fmt.Errorf("no <points> provided after [%s]", thisArg)
			}
		case "-v", "--val", "--vals", "--value", "--values":
			hadVal := false
			for ; i+1 < len(args) && args[i+1][0] != '-'; i++ {
				if len(strings.TrimSpace(args[i+1])) == 0 {
					entryVals = append(entryVals, args[i+1])
				} else {
					entryVals = append(entryVals, strings.Fields(args[i+1])...)
				}
				hadVal = true
			}
			if !hadVal {
				return fmt.Errorf("no <values> provided after [%s]", thisArg)
			}
		default:
			switch {
			case !widthGiven:
				var err error
				width, err = strconv.Atoi(thisArg)
				if err != nil {
					return err
				}
				widthGiven = true
			case !heightGiven:
				var err error
				height, err = strconv.Atoi(thisArg)
				if err != nil {
					return err
				}
				heightGiven = true
			default:
				return fmt.Errorf("unknown argument: [%s]", thisArg)
			}
		}
	}

	if !widthGiven {
		width = 6
	}
	if !heightGiven {
		height = width
	}

	vals := make([][]string, height)
	i := 0
	for y := range vals {
		vals[y] = make([]string, width)
		for x := range vals[y] {
			if len(entryVals) > 0 {
				vals[y][x] = entryVals[i%len(entryVals)]
			} else {
				vals[y][x] = fmt.Sprintf("%d", i)
			}
			i++
		}
	}

	if len(colors) == 0 {
		for h := 0; h < height; h += 2 {
			for w := 0; w < width; w++ {
				colors = append(colors, Point{w, h})
			}
		}
	}

	if len(hl) == 0 {
		for h := 0; h < height; h++ {
			for w := 0; w < width; w += 2 {
				hl = append(hl, Point{w, h})
			}
		}
	}

	fmt.Println(CreateIndexedGridString(vals, colors, hl))
	return nil
}

// ParsePoint parses a string of the format "<x>,<y>" into a Point.
func ParsePoint(str string) (Point, error) {
	parts := strings.Split(str, ",")
	if len(parts) != 2 {
		return Point{}, fmt.Errorf("unable to parse point %q: expected format \"<x>,<y>\"", str)
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return Point{}, fmt.Errorf("unable to parse point %q: invalid <x>: %w", str, err)
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return Point{}, fmt.Errorf("unable to parse point %q: invalid <y>: %w", str, err)
	}
	return Point{x, y}, nil
}

// Integer is each of the integer types.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// #############################################################################
// #############################################################################
// ######################   Everything Below is Copied   #######################
// #############################################################################
// #############################################################################

// CreateIndexedGridStringBz is for [][]byte
// CreateIndexedGridStringNums is for [][]int or [][]uint or [][]int16 etc.
// CreateIndexedGridString is for [][]string
// All of them have the signature (vals, color, highlight)
// CreateIndexedGridStringFunc is for any other [][]; signature = (vals, converter, color, highlight)

// A Point contains an X and Y value.
type Point struct {
	X int
	Y int
}

// NewPoint creates a new Point with the given coordinates.
func NewPoint(x, y int) *Point {
	return &Point{X: x, Y: y}
}

// String returns a string of this point in the format "(x,y)".
func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// GetX gets this Point's X value.
func (p Point) GetX() int {
	return p.X
}

// GetY gets this Point's Y value.
func (p Point) GetY() int {
	return p.Y
}

// GetXY gets this Point's (X, Y) values.
func (p Point) GetXY() (int, int) {
	return p.X, p.Y
}

// AddPoints returns a new point that is the sum of the provided points.
func AddPoints(points ...*Point) *Point {
	rv := NewPoint(0, 0)
	for _, p := range points {
		rv.X += p.X
		rv.Y += p.Y
	}
	return rv
}

// XY is something that has an X and Y value.
type XY interface {
	GetX() int
	GetY() int
	GetXY() (int, int)
}

// CreateIndexedGridStringBz creates a string of the provided bytes matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringBz[M ~[][]B, B byte | rune, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = string(val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridStringNums creates a string of the provided numbers matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringNums[M ~[][]N, N Integer, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = fmt.Sprintf("%d", val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridStringFunc creates a string of the provided matrix.
// The converter should take in a cell's value and output the string to use for that cell.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringFunc[M ~[][]G, G any, S ~[]E, E XY](vals M, converter func(G) string, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = converter(val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridString creates a string of the provided strings matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridString[S ~[]E, E XY](vals [][]string, colorPoints S, highlightPoints S) string {
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
				cellLen = utf8.RuneCountInString(c)
			}
		}
	}
	// Add an extra space if there's two or more characters per cell.
	if cellLen > 1 {
		cellLen++
	}

	// Define the format that each line will start with and for each cell.
	leadFmt := fmt.Sprintf("%%%dd:", len(fmt.Sprintf("%d", height)))
	blankLead := strings.Repeat(" ", len(fmt.Sprintf(leadFmt, 0)))
	cellFmt := fmt.Sprintf("%%%ds", cellLen)

	// If none of the rows have anything, just print out the row numbers.
	if width == 0 {
		lines := make([]string, len(vals))
		for y := range vals {
			lines[y] = fmt.Sprintf(leadFmt, y)
		}
		return strings.Join(lines, "\n")
	}

	// Create the index numbers across the top.
	dCount := len(fmt.Sprintf("%d", width-1))
	dLen := width * cellLen
	digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	topIndexLines := make([]string, dCount+1)
	topIndexLines[dCount] = strings.Repeat("-", dLen)
	rep := 1
	for l := 1; l <= dCount; l++ {
		first := " "
		if l == 1 {
			first = "0"
		}
		first = strings.Repeat(fmt.Sprintf(cellFmt, first), rep)

		var sb strings.Builder
		for _, s := range digits {
			if len(first)+sb.Len() >= dLen {
				break
			}
			sb.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), rep))
		}

		rep *= 10
		line := first + strings.Repeat(sb.String(), 1+width/rep)
		topIndexLines[dCount-l] = line[:dLen]
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
		if p.GetY() < height && p.GetX() < width && textFmt[p.GetY()][p.GetX()] <= 1 {
			textFmt[p.GetY()][p.GetX()] += 2
		}
	}

	// Start with the top index lines shifted right a bit to account for row indexes in the lines to follow.
	var rv strings.Builder
	for _, l := range topIndexLines {
		rv.WriteString(fmt.Sprintf("%s%s\n", blankLead, l))
	}

	// Add all the line numbers, and cells (with the desired coloring/marking).
	for y, r := range vals {
		rv.WriteString(fmt.Sprintf(leadFmt, y))
		for x := 0; x < width; x++ {
			v := ""
			if x < len(r) {
				v = r[x]
			}
			cell := fmt.Sprintf(cellFmt, v)
			switch textFmt[y][x] {
			case 0: // default look.
				rv.WriteString(cell)
			case 1: // color only.
				rv.WriteString("\033[94m" + cell + "\033[0m") // Light-blue text.
			case 2: // highlight only
				rv.WriteString("\033[7m" + cell + "\033[0m") // Foreground<->Background Reversed.
			case 3: // color and highlight
				rv.WriteString("\033[94;7m" + cell + "\033[0m") // Light-blue background after fg<->bg reversed.
			default: // Unknown, make it ugly.
				rv.WriteString("\033[93;41m" + cell + "\033[0m") // Bright yellow text on a red background.
			}
		}
		rv.WriteByte('\n')
	}

	return rv.String()
}
