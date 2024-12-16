package main

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// If copying this stuff, you'll want to copy all the stuff between the BeginCopy and EndCopy comments.
// Should get it all: getlines node-grid.go 18-423
// BeginCopy

// -----------------------------------------------------------------------------
// -----------------------------  Node Grid Stuff  -----------------------------
// -----------------------------------------------------------------------------

// Key stuff:
// AsNodeGrid[V any](vals [][]V) [][]*Node[V]
// GroupByValue[V comparable](vals [][]*Node[V]) map[V][]*Node[V]
// CreateByValueMapString[V cmp.Ordered](vals map[V][]*Node[V]) string
// CreateEnhancedByValueMapString[K cmp.Ordered, V ~[]W, W XY, S ~[]E, E XY](vals map[K]V, colorPoints, highlightPoints S) string
// PathString[S ~[]E, E XY](path S) string
// EnhancedPathString[P ~[]Q, Q XY, S ~[]E, E XY](path P, colorPoints, highlightPoints S) string
// MapGrid[G ~[][]E, E any, R any](grid G, mapper func(E) R) [][]R

// AsNodeGrid creates a grid of nodes with the provide values, all the nodes are linked up with their neighbors.
func AsNodeGrid[V any](vals [][]V) [][]*Node[V] {
	rv := make([][]*Node[V], len(vals))
	for y := range vals {
		rv[y] = make([]*Node[V], len(vals[y]))
		for x := range vals[y] {
			rv[y][x] = NewNode(x, y, vals[y][x])
		}
	}

	for y := range vals {
		for x := range vals[y] {
			curP := NewPoint(x, y)
			cur := Get(rv, curP)
			u, d, l, r := GetUDLR(curP)
			if n := Get(rv, u); n != nil {
				cur.Up = n
			}
			if n := Get(rv, d); n != nil {
				cur.Down = n
			}
			if n := Get(rv, l); n != nil {
				cur.Left = n
			}
			if n := Get(rv, r); n != nil {
				cur.Right = n
			}
		}
	}

	return rv
}

var (
	// DUp can be added to another point to get the point above it.
	DUp = NewPoint(0, -1)
	// DDown can be added to another point to get the point below it.
	DDown = NewPoint(0, 1)
	// DLeft can be added to another point to get the point to the left of it.
	DLeft = NewPoint(-1, 0)
	// DRight can be added to another point to get the point to the right of it.
	DRight = NewPoint(1, 0)
)

// GetUDLR gets the points that are up, down, left and right of the given one.
func GetUDLR(p *Point) (*Point, *Point, *Point, *Point) {
	return AddPoints(p, DUp), AddPoints(p, DDown), AddPoints(p, DLeft), AddPoints(p, DRight)
}

// IsIn returns true if the provided point exists in the provided grid.
func IsIn[E any](grid [][]E, p *Point) bool {
	return p != nil && p.Y >= 0 && p.Y < len(grid) && p.X >= 0 && p.X < len(grid[p.Y])
}

// Get will safely get the element of the grid at the provided point.
// If the point is outside the grid, the zero-value is returned.
func Get[E any](grid [][]E, p *Point) E {
	if !IsIn(grid, p) {
		var rv E
		return rv
	}
	return grid[p.Y][p.X]
}

// Node[V] has an x,y position, value, and knows its neighbors in a 2d grid.
type Node[V any] struct {
	Point
	Value V
	Up    *Node[V]
	Down  *Node[V]
	Left  *Node[V]
	Right *Node[V]
}

// NewNode creates a new Node at the given point with the given value (and no neighbors).
func NewNode[V any](x, y int, value V) *Node[V] {
	return &Node[V]{Point: Point{X: x, Y: y}, Value: value}
}

// String gets a string of this node that contains the point and value.
func (n *Node[V]) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s", n.Point, GenericValueString(n.Value))
}

// FullString converts this node into a string with the format "(<x>,<y>)=<value>:[<neighbor flags>]".
// If a node has all four neighbors, the <neighbor flags> will be "UDLR".
// Any neighbor directions the node does NOT have are replaced with a space in that string.
// E.g the node in the upper right corner of the grid only has neighbors to the right and down, so it's " D R".
func (n *Node[V]) FullString() string {
	if n == nil {
		return "<nil>"
	}
	dirs := Ternary(n.Up != nil, "U", " ") +
		Ternary(n.Down != nil, "D", " ") +
		Ternary(n.Right != nil, "R", " ") +
		Ternary(n.Left != nil, "L", " ")
	return fmt.Sprintf("%s=%s:[%s]", n.Point, GenericValueString(n.Value), dirs)
}

// PointString returns the "(<x>,<y>)" for this node.
func (n *Node[V]) PointString() string {
	if n == nil {
		return "<nil>"
	}
	return n.Point.String()
}

// GetUp is a nil-safe way to get the node up from this one.
func (n *Node[V]) GetUp() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Up
}

// GetDown is a nil-safe way to get the node down from this one.
func (n *Node[V]) GetDown() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Down
}

// GetLeft is a nil-safe way to get the node to the left of this one.
func (n *Node[V]) GetLeft() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Left
}

// GetRight is a nil-safe way to get the node to the right of this one.
func (n *Node[V]) GetRight() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Right
}

// Get returns the node that is at this one plus the provided d.
func (n *Node[V]) Get(d XY) *Node[V] {
	if n == nil {
		return nil
	}
	if d == nil {
		return n
	}
	dx, dy := d.GetXY()

	cur := n
	switch {
	case dy > 0:
		for y := 0; y < dy; y++ {
			cur = cur.Down
			if cur == nil {
				return nil
			}
		}
	case dy < 0:
		for y := 0; y > dy; y-- {
			cur = cur.Up
			if cur == nil {
				return nil
			}
		}
	}

	switch {
	case dx > 0:
		for x := 0; x < dx; x++ {
			cur = cur.Right
			if cur == nil {
				return nil
			}
		}
	case dx < 0:
		for x := 0; x > dx; x-- {
			cur = cur.Left
			if cur == nil {
				return nil
			}
		}
	}

	return cur
}

// Follow returns the node reached when starting at the provide node and moving the directions.
func Follow[V any, S ~[]E, E XY](start *Node[V], directions S) *Node[V] {
	rv := start
	for _, point := range directions {
		rv = rv.Get(point)
		if rv == nil {
			return nil
		}
	}
	return rv
}

// -----------------------------------------------------------------------------
// ---------------------------  By-Value Map Stuff  ----------------------------
// -----------------------------------------------------------------------------

// GroupByValue will create a map of value to nodes with that value.
func GroupByValue[V comparable](vals [][]*Node[V]) map[V][]*Node[V] {
	rv := make(map[V][]*Node[V])
	for y := range vals {
		for x := range vals[y] {
			rv[vals[y][x].Value] = append(rv[vals[y][x].Value], vals[y][x])
		}
	}
	return rv
}

// CreateByValueMapString creates a multi-line string, one line per key.
func CreateByValueMapString[V cmp.Ordered](vals map[V][]*Node[V]) string {
	keys := slices.Sorted(maps.Keys(vals))
	keyStrs := ToEqualLengthStrings(keys)

	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = fmt.Sprintf("[%s]: %s", keyStrs[i], PathString(vals[k]))
	}
	return strings.Join(lines, "\n") + "\n"
}

// CreateEnhancedByValueMapString creates a multi-line string, one line per key and colors and highlights the points as provided.
func CreateEnhancedByValueMapString[K cmp.Ordered, V ~[]W, W XY, S ~[]E, E XY](vals map[K]V, colorPoints, highlightPoints S) string {
	keys := slices.Sorted(maps.Keys(vals))
	keyStrs := ToEqualLengthStrings(keys)

	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = fmt.Sprintf("[%s]: %s", keyStrs[i], EnhancedPathString(vals[k], colorPoints, highlightPoints))
	}
	return strings.Join(lines, "\n") + "\n"
}

// ToEqualLengthStrings converts each val to a string using fmt.Sprintf("%v", val) and pads them to the same length.
// Numbers are left-padded, everything else is right-padded.
// The longest ones won't have any padding.
func ToEqualLengthStrings[E any](vals []E) []string {
	if vals == nil {
		return nil
	}
	if len(vals) == 0 {
		return []string{}
	}
	rv := make([]string, len(vals))
	maxLen := 0
	for i, val := range vals {
		rv[i] = GenericValueString(val)
		if len(rv[i]) > maxLen {
			maxLen = len(rv[i])
		}
	}
	padder := PadRight
	switch any(vals[0]).(type) {
	// Both byte and uint8 will match either of those. Same with rune and int32.
	// There's no easy way to identify when its one or the other, though.
	// And since I use byte and rune way more than uint8 or int32, we'll still pad the right side for those.
	case int, int8, int16, int32, int64, uint, uint16, uint64:
		padder = PadLeft
	}
	for i, str := range rv {
		if len(str) < maxLen {
			rv[i] = padder(rv[i], maxLen)
		}
	}
	return rv
}

// PathString returns a one-line string containing all the points in the provided slice of XY. E.g. "{3}[0:(0,0);1:(0,1);2:(1,1)]".
func PathString[S ~[]E, E XY](path S) string {
	lines := MapSlice(path, PointString)
	for i, line := range lines {
		lines[i] = strconv.Itoa(i) + ":" + line
	}
	return fmt.Sprintf("{%d}[%s]", len(path), strings.Join(lines, ";"))
}

// EnhancedPathString creates a string of the provided path, coloring and/or highlighting points as provided.
func EnhancedPathString[P ~[]Q, Q XY, S ~[]E, E XY](path P, colorPoints, highlightPoints S) string {
	fmts := make([]int, len(path))
	for i, point := range path {
		if HasPoint(colorPoints, point) {
			fmts[i] = 1
		}
		if HasPoint(highlightPoints, point) {
			fmts[i] += 2
		}
	}

	var rv strings.Builder
	for i, point := range path {
		if i != 0 {
			rv.WriteByte(';')
		}
		cell := fmt.Sprintf("%d:(%d,%d)", i, point.GetX(), point.GetY())
		switch fmts[i] {
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

	return fmt.Sprintf("{%d}[%s]", len(path), rv.String())
}

// HasPoints returns true if there's a point in path with the same (x,y) as the point provided.
func HasPoint[S ~[]E, E XY, P XY](path S, point P) bool {
	x, y := point.GetXY()
	for _, p := range path {
		if x == p.GetX() && y == p.GetY() {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// ---------------------------  New Generic Helpers  ---------------------------
// -----------------------------------------------------------------------------

// PointString returns the "(%d,%d)" string for the provided XY.
func PointString[V XY](p V) string {
	return fmt.Sprintf("(%d,%d)", p.GetX(), p.GetY())
}

// GenericValueString returns a string representation of the provided value that's a little better than just fmt.Sprintf("%v", value).
// Specifically, byte and rune types are converted to their character instead of just their number value.
func GenericValueString[T any](value T) string {
	switch v := any(value).(type) {
	case string:
		return v
	case byte, rune:
		return fmt.Sprintf("%c", v)
	}
	return fmt.Sprintf("%v", value)
}

// PadLeft will return a string with spaces added to the left of the provided one up to the provided length.
func PadLeft(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return strings.Repeat(" ", length-len(str)) + str
}

// PadRight will return a string with spaces added to the right of the provided one up to the provided length.
func PadRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}

// Ternary returns ifTrue if test == true, otherwise, returns ifFalse.
func Ternary[E any](test bool, ifTrue, ifFalse E) E {
	if test {
		return ifTrue
	}
	return ifFalse
}

// MapGrid creates a new map by running the provided mapper on each element of the provided grid.
func MapGrid[G ~[][]E, E any, R any](grid G, mapper func(E) R) [][]R {
	if grid == nil {
		return nil
	}
	rv := make([][]R, len(grid))
	for y := range grid {
		rv[y] = make([]R, len(grid[y]))
		for x := range rv[y] {
			rv[y][x] = mapper(grid[y][x])
		}
	}
	return rv
}

// EndCopy
// Should get it all: getlines node-grid.go 18-423
// If copying this stuff, you'll want to copy all the stuff between the BeginCopy and EndCopy comments.

// #############################################################################
// #########################  Stuff From the Template  #########################
// #############################################################################

// StringNumberJoinFunc maps the slice to strings using the provided stringer, numbers them, and joins them.
func StringNumberJoinFunc[S ~[]E, E any](slice S, stringer func(E) string, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, stringer), startAt), sep)
}

// AddLineNumbers adds line numbers to each string.
func AddLineNumbers(lines []string, startAt int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	lineFmt := DigitFormatForMax(len(lines)-1+startAt) + ": %s"
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf(lineFmt, i+startAt, line)
	}
	return rv
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d".
// DigitFormatForMax(382920) returns "%6d".
func DigitFormatForMax(maximum int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", maximum)))
}

// MapSlice returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapSlice[S ~[]E, E any, R any](slice S, mapper func(E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		rv[i] = mapper(e)
	}
	return rv
}

// ConversionRunes are some chars used to represent numbers for smaller output. See also: GetRune.
var ConversionRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz~-_+=|¦:;!@#$¢£¥%^&*()[]{}<>«»/?¿÷°§¶¤")

// GetRune returns the rune used to represent the provided number for smaller output.
// The runes will repeat every 100. E.g. IntRune(3) returns the same as IntRune(103).
func GetRune(i int) rune {
	return ConversionRunes[i%len(ConversionRunes)]
}

// Signed is a constraint of signed integer types. Same as golang.org/x/exp/constraints.Signed.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Signed is a constraint of unsigned integer types. Same as golang.org/x/exp/constraints.Unsigned.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint of integer types. Same as golang.org/x/exp/constraints.Integer.
type Integer interface {
	Signed | Unsigned
}

// -----------------------------------------------------------------------------
// -------------------------  CreateIndexedGridString  -------------------------
// -----------------------------------------------------------------------------

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
func CreateIndexedGridStringStringer[M ~[][]G, G fmt.Stringer, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = val.String()
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

// #############################################################################
// ##################################  Main  ###################################
// #############################################################################

func main() {
	// Allow for usage of: go run <file> [<size> [<max value>]]
	args := os.Args[1:]
	size := 10
	maxVal := 13
	if len(args) > 0 {
		var err error
		size, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("invalid size arg %q: %v", args[0], err)
			return
		}
	}
	if len(args) > 1 {
		var err error
		maxVal, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("invalid maxVal arg %q: %v", args[1], err)
			return
		}
	}

	// Identify the points that we'll want colored and highlighted.
	// The points on the x are colored.
	// The points on the + and in the corners are highlighted.
	mid1 := size / 2
	mid2 := (size - 1) / 2
	var colors, bolds []*Point
	for i := 0; i < size; i++ {
		colors = append(colors, NewPoint(i, i), NewPoint(i, size-i-1)) // x shape
		bolds = append(bolds, NewPoint(i, mid1), NewPoint(mid1, i))    // + shape
		if mid1 != mid2 {
			bolds = append(bolds, NewPoint(i, mid2), NewPoint(mid2, i)) // + shape is thick when size%2==0
		}
	}
	bolds = append(bolds, NewPoint(0, 0), NewPoint(0, size-1), NewPoint(size-1, 0), NewPoint(size-1, size-1)) // Also the 4 corners.

	// Create the value map.
	values := make([][]int, size)
	for y := range values {
		values[y] = make([]int, size)
		for x := range values[y] {
			values[y][x] = 1 + (y+size*x)%maxVal
		}
	}
	fmt.Printf("Values (int):\n%s\n", CreateIndexedGridStringNums(values, colors, bolds))

	// Convert it to a node map.
	nodeMap := AsNodeGrid(values)
	fmt.Printf("Node Map (int):\n%s\n", CreateIndexedGridStringStringer(nodeMap, colors, bolds))

	// And organize the nodes by value.
	byVal := GroupByValue(nodeMap)
	fmt.Printf("By Value (int):\n%s\n", CreateByValueMapString(byVal))
	fmt.Printf("By Value (int, enhanced):\n%s\n", CreateEnhancedByValueMapString(byVal, colors, bolds))

	// ------------

	// Now let's try this stuff with a rune map.
	runeVals := MapGrid(values, GetRune)
	fmt.Printf("Values (rune):\n%s\n", CreateIndexedGridStringBz(runeVals, colors, bolds))

	// Runes as a node map.
	runeNodeMap := AsNodeGrid(runeVals)
	fmt.Printf("Node Map (rune):\n%s\n", CreateIndexedGridStringStringer(runeNodeMap, colors, bolds))

	// And grouped up.
	runesByVal := GroupByValue(runeNodeMap)
	fmt.Printf("By Value (rune):\n%s\n", CreateByValueMapString(runesByVal))
	fmt.Printf("By Value (rune, enhanced):\n%s\n", CreateEnhancedByValueMapString(runesByVal, colors, bolds))

	// ------------
	intStringer := func(i int) string {
		return fmt.Sprintf("<%d>", i)
	}

	// Now let's try this stuff with a string map.
	strVals := MapGrid(values, intStringer)
	fmt.Printf("Values (string):\n%s\n", CreateIndexedGridString(strVals, colors, bolds))

	// Runes as a node map.
	strNodeMap := AsNodeGrid(strVals)
	fmt.Printf("Node Map (string):\n%s\n", CreateIndexedGridStringStringer(strNodeMap, colors, bolds))

	// And grouped up.
	strsByVal := GroupByValue(strNodeMap)
	fmt.Printf("By Value (string):\n%s\n", CreateByValueMapString(strsByVal))
	fmt.Printf("By Value (string, enhanced):\n%s\n", CreateEnhancedByValueMapString(strsByVal, colors, bolds))
}
