package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_COUNT = 0

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	answer := DoDigPlan(input.Plan)
	return fmt.Sprintf("%d", answer), nil
}

func DoDigPlan(plan []Step) int {
	defer FuncEnding(FuncStarting())
	trench := DigTrench(plan)
	trenchLen := trench.GetLength()
	Debugf("Trench (%d):\n%s", trenchLen, trench)

	xMin, yMin, xMax, yMax := trench.MinMax()
	width := xMax - xMin + 1
	height := yMax - yMin + 1
	xShift := 0 - xMin
	yShift := 0 - yMin
	Debugf("Min: (%d,%d), Max: (%d,%d), Width: %d, Height: %d, Shift: (%d,%d)",
		xMin, yMin, xMax, yMax, width, height, xShift, yShift)

	if debug {
		sTrench := make(Lines, 0, len(trench))
		for _, line := range trench {
			from := NewPoint(line.From.X+xShift, line.From.Y+yShift)
			to := NewPoint(line.To.X+xShift, line.To.Y+yShift)
			sTrench = append(sTrench, NewLine(from, to))
		}

		sTrenchLen := sTrench.GetLength()
		Debugf("Shifted Trench (%d):\n%s", sTrenchLen, sTrench)

		ShrinkAndDraw(sTrench, width, height)

		// CheckCuts()
	}

	wholeBox := NewBox(NewPoint(xMin, yMin), NewPoint(xMax, yMax), Unknown)
	boxes := Boxes{wholeBox}
	for _, line := range trench {
		nBoxes := Boxes{}
		for _, b := range boxes {
			nBoxes = append(nBoxes, b.Cut(line)...)
		}
		boxes = nBoxes
	}

	var trenchBoxes, unknownBoxes Boxes
	for _, box := range boxes {
		if box.Type == Trench {
			trenchBoxes = append(trenchBoxes, box)
		} else {
			unknownBoxes = append(unknownBoxes, box)
		}
	}

	Debugf("After cutting, there are %d trench boxes, and %d unknown boxes.", len(trenchBoxes), len(unknownBoxes))
	trenchArea := trenchBoxes.TotalArea()
	Debugf("Trench area: %d", trenchArea)

	var outsideBoxes Boxes
	keepGoing := true
	for keepGoing {
		keepGoing = false
		var stillUnknown Boxes
		for _, box := range unknownBoxes {
			if IsExternal(box, boxes) {
				box.Type = External
				outsideBoxes = append(outsideBoxes, box)
				keepGoing = true
			} else {
				stillUnknown = append(stillUnknown, box)
			}
		}
		unknownBoxes = stillUnknown
	}

	internalArea := unknownBoxes.TotalArea()
	rv := trenchArea + internalArea
	Debugf("%d = %d (trench) %d (internal)", rv, trenchArea, internalArea)
	return rv
}

func IsExternal(box *Box, boxes Boxes) bool {
	if box.Type == External {
		return true
	}
	if box.Type != Unknown {
		return false
	}

	// Check along the top edge for an external box.
	for x := box.XMin; x <= box.XMax; x++ {
		topBox := boxes.Find(x, box.YMin-1)
		if topBox == nil || topBox.Type == External {
			return true
		}
		x = topBox.XMax
	}

	// Check along the bottom edge.
	for x := box.XMin; x <= box.XMax; x++ {
		bottomBox := boxes.Find(x, box.YMax+1)
		if bottomBox == nil || bottomBox.Type == External {
			return true
		}
		x = bottomBox.XMax
	}

	// Check along the left side.
	for y := box.YMin; y <= box.YMax; y++ {
		leftBox := boxes.Find(box.XMin-1, y)
		if leftBox == nil || leftBox.Type == External {
			return true
		}
		y = leftBox.YMax
	}

	// And the right side.
	for y := box.YMin; y <= box.YMax; y++ {
		rightBox := boxes.Find(box.XMax+1, y)
		if rightBox == nil || rightBox.Type == External {
			return true
		}
		y = rightBox.YMax
	}

	return false
}

const (
	Empty    = byte('.')
	Trench   = byte('#')
	Internal = byte('I')
	External = byte(' ')
	Unknown  = byte('x')
)

type Box struct {
	XMin int
	YMin int
	XMax int
	YMax int
	Type byte
}

func NewBox(p1, p2 *Point, t byte) *Box {
	rv := &Box{
		Type: t,
	}
	rv.XMin, rv.XMax = PutInOrder(p1.X, p2.X)
	rv.YMin, rv.YMax = PutInOrder(p1.Y, p2.Y)
	return rv
}

func PutInOrder(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func (b Box) String() string {
	return fmt.Sprintf("[%d,%d]-[%d,%d]", b.XMin, b.YMin, b.XMax, b.YMax)
}

func (b Box) Contains(x, y int) bool {
	return b.XMin <= x && x <= b.XMax && b.YMin <= y && y <= b.YMax
}

func (b Box) Width() int {
	return b.XMax - b.XMin + 1
}

func (b Box) Height() int {
	return b.YMax - b.YMin + 1
}

func (b Box) Area() int {
	return b.Width() * b.Height()
}

func (b Box) Restrict(limit Box) *Box {
	return &Box{
		XMin: Max(b.XMin, limit.XMin),
		XMax: Min(b.XMax, limit.XMax),
		YMin: Max(b.YMin, limit.YMin),
		YMax: Min(b.YMax, limit.YMax),
		Type: b.Type,
	}
}

func (b Box) IsValid() bool {
	return b.XMin <= b.XMax && b.YMin <= b.YMax
}

func (b Box) Cut(line Line) Boxes {
	isH := line.From.Y == line.To.Y
	isV := line.From.X == line.To.X
	if !isH && !isV {
		panic(fmt.Errorf("cannot handle diagonal line %s", line))
	}

	// Create a box to represent this line (the trench) and restrict it to this box.
	tBox := NewBox(&line.From, &line.To, Trench).Restrict(b)
	if !tBox.IsValid() {
		// The line is completely outside the box.
		return Boxes{&b}
	}

	// The line intersects the box in some way.
	// Cut into: 1: Just the line, 2: Upper left, 3: Upper right, 4: Lower left, 5: Lower right.
	rv := make(Boxes, 1, 5)
	// 1: Just the line (already created above).
	rv[0] = tBox
	if tBox.XMin > b.XMin { // 2: Upper left.
		rv = append(rv, NewBox(NewPoint(b.XMin, b.YMin), NewPoint(tBox.XMin-1, tBox.YMax), b.Type))
	}
	if tBox.YMin > b.YMin { // 3: Upper right.
		rv = append(rv, NewBox(NewPoint(tBox.XMin, b.YMin), NewPoint(b.XMax, tBox.YMin-1), b.Type))
	}
	if tBox.YMax < b.YMax { // 4: Lower left.
		rv = append(rv, NewBox(NewPoint(b.XMin, tBox.YMax+1), NewPoint(tBox.XMax, b.YMax), b.Type))
	}
	if tBox.XMax < b.XMax { // 5: Lower right.
		rv = append(rv, NewBox(NewPoint(tBox.XMax+1, tBox.YMin), NewPoint(b.XMax, b.YMax), b.Type))
	}
	return rv
}

func CheckCuts() {
	box := NewBox(NewPoint(1, 1), NewPoint(10, 10), Unknown)
	tests := []struct {
		name string
		line Line
	}{
		{name: "V All Outside Left", line: NewLine(NewPoint(0, 0), NewPoint(0, 11))},
		{name: "V Left Edge", line: NewLine(NewPoint(1, 0), NewPoint(1, 11))},
		{name: "V Inside", line: NewLine(NewPoint(5, 0), NewPoint(5, 11))},
		{name: "V Right Edge", line: NewLine(NewPoint(10, 0), NewPoint(10, 11))},
		{name: "V All Outside Right", line: NewLine(NewPoint(10, 0), NewPoint(10, 11))},
		{name: "V Top Only", line: NewLine(NewPoint(5, 0), NewPoint(5, 5))},
		{name: "V Bottom Only", line: NewLine(NewPoint(5, 5), NewPoint(5, 11))},
		{name: "V Both Inside", line: NewLine(NewPoint(5, 3), NewPoint(5, 8))},

		{name: "H All Outside Top", line: NewLine(NewPoint(0, 0), NewPoint(11, 0))},
		{name: "H Top Edge", line: NewLine(NewPoint(0, 1), NewPoint(11, 1))},
		{name: "H Inside", line: NewLine(NewPoint(0, 5), NewPoint(11, 5))},
		{name: "H Bottom Edge", line: NewLine(NewPoint(0, 10), NewPoint(11, 10))},
		{name: "H All Outside Bottom", line: NewLine(NewPoint(0, 10), NewPoint(11, 10))},
		{name: "H Left Only", line: NewLine(NewPoint(0, 5), NewPoint(5, 5))},
		{name: "H Right Only", line: NewLine(NewPoint(5, 5), NewPoint(11, 5))},
		{name: "H Both Inside", line: NewLine(NewPoint(3, 5), NewPoint(8, 5))},

		{name: "P Outside", line: NewLine(NewPoint(0, 5), NewPoint(0, 5))},
		{name: "P Corner TL", line: NewLine(NewPoint(1, 1), NewPoint(1, 1))},
		{name: "P Corner TR", line: NewLine(NewPoint(10, 1), NewPoint(10, 1))},
		{name: "P Corner BL", line: NewLine(NewPoint(1, 10), NewPoint(1, 10))},
		{name: "P Corner BR", line: NewLine(NewPoint(10, 10), NewPoint(10, 10))},
		{name: "P Corner T", line: NewLine(NewPoint(5, 1), NewPoint(5, 1))},
		{name: "P Corner B", line: NewLine(NewPoint(5, 10), NewPoint(5, 10))},
		{name: "P Corner L", line: NewLine(NewPoint(1, 5), NewPoint(1, 5))},
		{name: "P Corner R", line: NewLine(NewPoint(10, 5), NewPoint(10, 5))},
		{name: "P Inside", line: NewLine(NewPoint(5, 5), NewPoint(5, 5))},
	}

	for n, tc := range tests {
		Stderrf("[%d]: %q %s|%s", n, tc.name, box, tc.line)
		boxes := box.Cut(tc.line)
		Stderrf("[%d]: = %s", n, boxes)

		origArea := box.Area()
		newArea := boxes.TotalArea()
		switch {
		case origArea > newArea:
			Stderrf("[%d]:  AREA SHRANK from %d to %d!!", n, origArea, newArea)
		case origArea < newArea:
			Stderrf("[%d]:  AREA GREW from %d to %d!!", n, origArea, newArea)
		default:
			// Debugf("[%d]:  Area stayed the same %d and %d.", n, origArea, newArea)
		}

		for i, b := range boxes {
			if !b.IsValid() {
				Stderrf("[%d]:  NEW BOX [%d] = %s: Invalid!", n, i, b)
			}
			if b.XMin < box.XMin {
				Stderrf("[%d]:  NEW BOX [%d] = %s: XMin %d less than original %d", n, i, b, b.XMin, box.XMin)
			}
			if b.XMax > box.XMax {
				Stderrf("[%d]:  NEW BOX [%d] = %s: XMax %d more than original %d", n, i, b, b.XMax, box.XMax)
			}
			if b.YMin < box.YMin {
				Stderrf("[%d]:  NEW BOX [%d] = %s: YMin %d less than original %d", n, i, b, b.YMin, box.YMin)
			}
			if b.YMax > box.YMax {
				Stderrf("[%d]:  NEW BOX [%d] = %s: YMax %d more than original %d", n, i, b, b.YMax, box.YMax)
			}
		}

		for i, b1 := range boxes {
			for j, b2 := range boxes {
				if i == j {
					continue
				}
				r := b1.Restrict(*b2)
				if r.IsValid() {
					Stderrf("[%d]:  OVERLAP in result[%d] = %s and result[%d] = %s", n, i, b1, j, b2)
				}
			}
		}
	}
}

type Boxes []*Box

func (b Boxes) String() string {
	return strings.Join(MapSlice(b, (*Box).String), " ")
}

func (b Boxes) TotalArea() int {
	rv := 0
	for _, box := range b {
		rv += box.Area()
	}
	return rv
}

func (b Boxes) Find(x, y int) *Box {
	for _, box := range b {
		if box.Contains(x, y) {
			return box
		}
	}
	return nil
}

func ShrinkAndDraw(trench Lines, width, height int) {
	sDim := 120
	xDiv := float64(width) / float64(sDim)
	yDiv := float64(height) / float64(sDim)
	shrunken := make(Lines, 0, len(trench))
	for _, line := range trench {
		newLine := Line{
			From: Point{
				X: int(math.Round(float64(line.From.X) / xDiv)),
				Y: int(math.Round(float64(line.From.Y) / yDiv)),
			},
			To: Point{
				X: int(math.Round(float64(line.To.X) / xDiv)),
				Y: int(math.Round(float64(line.To.Y) / yDiv)),
			},
		}
		shrunken = append(shrunken, newLine)
	}

	Debugf("Shrunken:\n%s", shrunken)

	pit := make([][]byte, sDim+1)
	for y := range pit {
		pit[y] = []byte(strings.Repeat(string(Empty), sDim+1))
	}

	var tPoints []Point
	for _, line := range shrunken {
		length := line.Length()
		if length == 0 {
			continue
		}
		dx, dy := 0, 0
		switch {
		case line.From.X < line.To.X:
			dx = 1
		case line.From.X > line.To.X:
			dx = -1
		case line.From.Y < line.To.Y:
			dy = 1
		case line.From.Y > line.To.Y:
			dy = -1
		default:
			panic(fmt.Errorf("cannot figure dx/dy for %s", line))
		}
		for i := 1; i <= length; i++ {
			newP := Point{
				X: line.From.X + (dx * i),
				Y: line.From.Y + (dy * i),
			}
			pit[newP.Y][newP.X] = Trench
			tPoints = append(tPoints, newP)
		}
	}

	Debugf("Pit:\n%s", CreateIndexedGridStringBz(pit, nil, tPoints))
}

func DigTrench(plan []Step) Lines {
	defer FuncEnding(FuncStarting())
	rv := make(Lines, 0, len(plan))
	cur := Point{X: 0, Y: 0}
	for _, step := range plan {
		newLine := step.Follow(cur)
		cur = newLine.To
		rv = append(rv, newLine)
	}
	return rv
}

type Lines []Line

func (l Lines) String() string {
	return StringNumberJoin(l, 1, "\n")
}

func (l Lines) MinMax() (xMin, yMin, xMax, yMax int) {
	if len(l) == 0 {
		return
	}
	xMin, xMax = MAX_INT, 0
	yMin, yMax = MAX_INT, 0
	for _, line := range l {
		xMin, xMax = MinMax(xMin, line.From.X, xMax)
		xMin, xMax = MinMax(xMin, line.To.X, xMax)
		yMin, yMax = MinMax(yMin, line.From.Y, yMax)
		yMin, yMax = MinMax(yMin, line.To.Y, yMax)
	}
	return
}

func MinMax(min, val, max int) (int, int) {
	if val < min {
		return val, max
	}
	if val > max {
		return min, val
	}
	return min, max
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (l Lines) GetLength() int {
	rv := 0
	for _, line := range l {
		rv += line.Length()
	}
	return rv
}

type Line struct {
	From Point
	To   Point
}

func NewLine(from, to *Point) Line {
	return Line{From: *from, To: *to}
}

func (l Line) String() string {
	return fmt.Sprintf("%s-%s", l.From, l.To)
}

func (l Line) Length() int {
	return Abs(l.From.X-l.To.X) + Abs(l.From.Y-l.To.Y)
}

func Abs(a int) int {
	if a > 0 {
		return a
	}
	return -1 * a
}

const (
	Right = byte('R')
	Left  = byte('L')
	Up    = byte('U')
	Down  = byte('D')
)

var DirMap = map[byte]byte{
	'0': Right,
	'1': Down,
	'2': Left,
	'3': Up,
}

type Step struct {
	Dir    byte
	Len    int
	Color  string
	OldDir byte
	OldLen int
}

func (s Step) String() string {
	return fmt.Sprintf("%c %d <= %s (was %c %d)", s.Dir, s.Len, s.Color, s.OldDir, s.OldLen)
}

func (s Step) Follow(cur XY) Line {
	dx, dy := 0, 0
	switch s.Dir {
	case Right:
		dx = 1
	case Left:
		dx = -1
	case Up:
		dy = -1
	case Down:
		dy = 1
	}
	from := NewPoint(cur.GetX(), cur.GetY())
	to := NewPoint(from.X+(dx*s.Len), from.Y+(dy*s.Len))
	return NewLine(from, to)
}

func ParseStep(line string) (*Step, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unable to split line %q into 3 parts, found %d", line, len(parts))
	}

	if len(parts[0]) != 1 || !strings.Contains("RLUD", parts[0]) { //nolint:gocritic // This order is correct.
		return nil, fmt.Errorf("unexpected direction %q from line %q", parts[0], line)
	}
	oldLength, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to convert length %q from line %q to int: %w", parts[1], line, err)
	}

	if len(parts[2]) != 9 || !strings.HasPrefix(parts[2], "(#") || !strings.HasSuffix(parts[2], ")") {
		return nil, fmt.Errorf("color part %q of line %q has unknown format", parts[2], line)
	}
	color := []byte(strings.TrimSuffix(strings.TrimPrefix(parts[2], "(#"), ")"))

	length, err := strconv.ParseUint(string(color[:5]), 16, 64)
	if err != nil {
		return nil, fmt.Errorf("could not convert color %q to length from line %q: %w", string(color[:5]), line, err)
	}

	dir, found := DirMap[color[5]]
	if !found {
		return nil, fmt.Errorf("unknown direction byte %q from line %q", string(color[5]), line)
	}

	return &Step{
		Dir:    dir,
		Len:    int(length),
		Color:  string(color),
		OldDir: parts[0][0],
		OldLen: oldLength,
	}, nil
}

type Input struct {
	Plan []Step
}

func (i Input) String() string {
	return fmt.Sprintf("Dig Plan (%d):\n%s", len(i.Plan), StringNumberJoin(i.Plan, 1, "\n"))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Plan: make([]Step, len(lines))}
	// TODO: Update this to parse the lines and create the puzzle input.
	for i, line := range lines {
		step, err := ParseStep(line)
		if err != nil {
			return nil, err
		}
		rv.Plan[i] = *step
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------  Some generic stuff  --------------------------------------
// -------------------------------------------------------------------------------------------------

const MIN_INT8 = int8(-128)
const MAX_INT8 = int8(127)
const MIN_INT16 = int16(-32_768)
const MAX_INT16 = int16(32_767)
const MIN_INT32 = int32(-2_147_483_648)
const MAX_INT32 = int32(2_147_483_647)
const MIN_INT64 = int64(-9_223_372_036_854_775_808)
const MAX_INT64 = int64(9_223_372_036_854_775_807)
const MIN_INT = -9_223_372_036_854_775_808
const MAX_INT = 9_223_372_036_854_775_807

const MAX_UINT8 = uint8(255)
const MAX_UINT16 = uint16(65_535)
const MAX_UINT32 = uint32(4_294_967_295)
const MAX_UINT64 = uint64(18_446_744_073_709_551_615)
const MAX_UINT = uint(18_446_744_073_709_551_615)

// A Point contains an X and Y value.
type Point struct {
	X int
	Y int
}

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

// XY is something that has an X and Y value.
type XY interface {
	GetX() int
	GetY() int
	GetXY() (int, int)
}

// CreateIndexedGridString creates a string of the provided vals bytes matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringBz[S ~[]E, E XY](vals [][]byte, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = string(val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// IntType is each of the integer types.
type IntType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// CreateIndexedGridString creates a string of the provided vals bytes matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringNums[M ~[][]N, N IntType, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = fmt.Sprintf("%d", val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridString creates a string of the provided vals matrix.
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
				cellLen = len(c)
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

// SplitParseInts splits a string using the given separator and converts each part into an int.
// Uses strings.Split(s, sep) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseInts(s string, sep string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Split(s, sep) {
		if len(entry) > 0 {
			i, err := strconv.Atoi(strings.TrimSpace(entry))
			if err != nil {
				return rv, err
			}
			rv = append(rv, i)
		}
	}
	return rv, nil
}

// StringNumberJoin maps the slice to strings, numbers them and joins them.
func StringNumberJoin[S ~[]E, E Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
}

// AddLineNumbers adds line numbers to each string.
func AddLineNumbers(lines []string, startAt int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	lineFmt := DigitFormatForMax(len(lines)) + ": %s"
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf(lineFmt, i+startAt, line)
	}
	return rv
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d".
// DigitFormatForMax(382920) returns "%6d".
func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
}

// PrefixLines splits each provided string on \n then adds a prefix to each line, then puts it all back together.
func PrefixLines(pre string, strs ...string) string {
	var rv strings.Builder
	lastI := len(strs) - 1
	for i, str := range strs {
		lines := strings.Split(str, "\n")
		lastJ := len(lines) - 1
		for j, line := range lines {
			rv.WriteString(pre)
			rv.WriteString(line)
			if i != lastI || j != lastJ {
				rv.WriteByte('\n')
			}
		}
	}
	return rv.String()
}

// Stringer is an interface for something that can be turned into a string.
type Stringer interface {
	String() string
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

// MapPSlice returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapPSlice[S ~[]*E, E any, R any](slice S, mapper func(E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		rv[i] = mapper(*e)
	}
	return rv
}

// MapSliceP returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapSliceP[S ~[]E, E any, R any](slice S, mapper func(*E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		e := e
		rv[i] = mapper(&e)
	}
	return rv
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  CLI params and input parsing  ---------------------------------
// -------------------------------------------------------------------------------------------------

// Params contains anything that might be provided via command-line arguments.
type Params struct {
	// Verbose is a flag indicating some extra output is desired.
	Verbose bool
	// HelpPrinted is whether or not the help message was printed.
	HelpPrinted bool
	// Errors is a list of errors encountered while parsing the arguments.
	Errors []error
	// Count is just a generic int that can be provided.
	Count int
	// InputFile is the file that contains the puzzle data to solve.
	InputFile string
	// Input is the contents of the input file split on newlines.
	Input []string
	// Custom is a set of custom strings to provide as input.
	Custom []string
}

// String creates a multi-line string representing this Params.
func (c Params) String() string {
	defer FuncEnding(FuncStarting())
	nameFmt := "%10s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", c.Verbose),
		fmt.Sprintf(nameFmt+"%d", "Errors", len(c.Errors)),
		fmt.Sprintf(nameFmt+"%d", "Count", c.Count),
		fmt.Sprintf(nameFmt+"%s", "Input File", c.InputFile),
		fmt.Sprintf(nameFmt+"%d lines", "Input", len(c.Input)),
		fmt.Sprintf(nameFmt+"%d lines", "Custom", len(c.Custom)),
	}
	if len(c.Errors) > 0 {
		lines = append(lines, fmt.Sprintf("Errors (%d):", len(c.Errors)))
		errors := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errors[i] = err.Error()
		}
		lines = append(lines, AddLineNumbers(errors, 1)...)
	}
	if len(c.Input) > 0 {
		lines = append(lines, fmt.Sprintf("Input (%d):", len(c.Input)))
		lines = append(lines, AddLineNumbers(c.Input, 0)...)
	}
	if len(c.Custom) > 0 {
		lines = append(lines, fmt.Sprintf("Custom Input (%d):", len(c.Custom)))
		lines = append(lines, AddLineNumbers(c.Custom, 0)...)
	}
	return strings.Join(lines, "\n")
}

// DEFAULT_INPUT_FILE is the default input filename.
const DEFAULT_INPUT_FILE = "example.input"

// GetParams parses the provided args into the command's params.
func GetParams(args []string) *Params {
	defer FuncEnding(FuncStarting())
	var err error
	rv := Params{}
	countGiven := false
	verboseGiven := false
	for i := 0; i < len(args); i++ {
		switch {
		// Flag cases go first.
		case IsOneOfStrFold(args[i], "--help", "-h", "help"):
			Debugf("Help flag found: [%s].", args[i])
			lines := []string{
				fmt.Sprintf("Usage: %s [<input file>] [<flags>]", GetMyExe()),
				fmt.Sprintf("Default <input file> is %s", DEFAULT_INPUT_FILE),
				"Flags:",
				"  --debug       Turns on debugging.",
				"  --verbose|-v  Turns on verbose output.",
				"",
				"Single Options:",
				"  Providing these multiple times will overwrite the previously provided value.",
				"  --input|-i <input file>  An option to define the input file.",
				"  --count|-n <number>      Defines a count.",
				"",
				"Repeatable Options:",
				"  Providing these multiple times will add to previously provided values.",
				"  Values are read until the next one starts with a dash.",
				"  To provide entries that start with a dash, you can use --flag='<value>' syntax.",
				"  --lines|-l <value 1> [<value 2> ...]  Defines custom input lines.",
				"",
			}
			// Not using Stdoutf() here because the extra formatting is annoying with help text.
			fmt.Println(strings.Join(lines, "\n"))
			rv.HelpPrinted = true
		case HasPrefixFold(args[i], "--debug"):
			Debugf("Debug option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			oldDebug := debug
			debug, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
			if err == nil {
				switch {
				case !oldDebug && debug:
					Stderrf("Debugging enabled by CLI arguments.")
				case oldDebug && !debug:
					Stderrf("Debugging disabled by CLI arguments.")
				}
			}
		case HasOneOfPrefixesFold(args[i], "--verbose", "-v"):
			Debugf("Verbose option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.Verbose, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
			verboseGiven = true
		case HasOneOfPrefixesFold(args[i], "--input", "--input-file"):
			Debugf("Input file option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.InputFile, extraI, err = ParseFlagString(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--count", "-c", "-n"):
			Debugf("Count option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.Count, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)
			countGiven = true
		case HasOneOfPrefixesFold(args[i], "--line", "--lines", "-l", "--custom", "--val"):
			Debugf("Custom option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			var vals []string
			vals, extraI, err = ParseRepeatedFlagString(args[i:])
			rv.Custom = append(rv.Custom, vals...)
			i += extraI
			rv.AppendError(err)

		// Positional args go last in the order they're expected.
		case len(rv.InputFile) == 0 && len(args[i]) > 0 && args[i][0] != '-':
			Debugf("Input File argument: [%s], args after: %q", args[i], args[i:])
			rv.InputFile = args[i]
		default:
			Debugf("Unknown argument found: [%s], args after: %q.", args[i], args[i:])
			rv.AppendError(fmt.Errorf("unknown argument %d: [%s]", i+1, args[i]))
		}
	}
	if len(rv.InputFile) == 0 {
		rv.InputFile = DEFAULT_INPUT_FILE
	}
	if !verboseGiven {
		rv.Verbose = debug
	}
	if !countGiven {
		rv.Count = DEFAULT_COUNT
	}
	return &rv
}

// AppendError adds an error to this Params as long as the error is not nil.
func (c *Params) AppendError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// HasError returns true if this Params has one or more errors.
func (c Params) HasError() bool {
	return len(c.Errors) != 0
}

// Error flattens the Errors slice into a single string.
// It also makes the Params struct satisfy the error interface.
func (c Params) GetError() error {
	switch len(c.Errors) {
	case 0:
		return nil
	case 1:
		return c.Errors[0]
	default:
		errs := make([]error, 1, 1+len(c.Errors))
		errs[0] = fmt.Errorf("Found %d errors:", len(c.Errors)) //nolint:stylecheck,revive // punct okay here.
		for i, err := range c.Errors {
			errs = append(errs, fmt.Errorf("  %d: %w", i+1, err))
		}
		return errors.Join(errs...)
	}
}

// IsOneOfStrFold tests if the given string is equal (ignoring case) to one of the given options.
func IsOneOfStrFold(str string, opts ...string) bool {
	for _, opt := range opts {
		if strings.EqualFold(str, opt) {
			return true
		}
	}
	return false
}

// HasPrefixFold tests if the given string starts with the given prefix (ignoring case).
func HasPrefixFold(str, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[0:len(prefix)], prefix)
}

// HasOneOfPrefixesFold tests if the given string has one of the given prefixes.
func HasOneOfPrefixesFold(str string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if HasPrefixFold(str, pre) {
			return true
		}
	}
	return false
}

// ParseBool converts a string into a bool.
// First return bool is the parsed value.
// Second return bool is whether or not the parsing was successful.
func ParseBool(str string) (val bool, isBool bool) {
	// Note: Not using strconv.ParseBool because I want it a bit looser (any casing) and to allow yes/no/off/on values.
	lstr := strings.ToLower(strings.TrimSpace(str))
	switch lstr {
	case "false", "f", "0", "no", "n", "off":
		isBool = true
	case "true", "t", "1", "yes", "y", "on":
		val = true
		isBool = true
	}
	return
}

// ParseFlagString parses a string flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's string value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagString(args []string) (string, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) == 2 {
			if len(parts[1]) > 1 {
				for _, c := range []string{`'`, `"`} {
					if parts[1][:1] == c && parts[1][len(parts[1])-1:] == c {
						return parts[1][1 : len(parts[1])-1], 0, nil
					}
				}
			}
			return parts[1], 0, nil
		}
		return "", 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
	}
	if len(args) > 1 {
		return args[1], 1, nil
	}
	return "", 0, fmt.Errorf("no value provided after %s flag", args[0])
}

// ParseRepeatedFlagString parses a flag that allows providing multiple strings.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's string value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseRepeatedFlagString(args []string) ([]string, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) != 2 {
			return []string{}, 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
		}
		if len(parts[1]) > 1 {
			for _, c := range []string{`'`, `"`} {
				if parts[1][:1] == c && parts[1][len(parts[1])-1:] == c {
					parts[1] = parts[1][1 : len(parts[1])-1]
				}
			}
		}
		return parts[1:], 0, nil
	}
	rv := []string{}
	for _, arg := range args[1:] {
		if arg[0] == '-' {
			return rv, len(rv), nil
		}
		rv = append(rv, arg)
	}
	if len(rv) > 0 {
		return rv, len(rv), nil
	}
	return rv, 0, fmt.Errorf("no values provided after %s flag", args[0])
}

// ParseFlagBool parses a boolean flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and parsed.
// Otherwise, if args[1] is a boolean string value, that is parsed.
// Otherwise, the flag defaults to true.
//
// The first return value is the parsed boolean value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagBool(args []string) (bool, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) == 2 {
			val, isBool := ParseBool(parts[1])
			if !isBool {
				return false, 0, fmt.Errorf("invalid %s bool value: [%s]", parts[0], parts[1])
			}
			return val, 0, nil
		}
		return false, 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
	}
	if len(args) > 1 {
		val, isBool := ParseBool(args[1])
		if isBool {
			return val, 1, nil
		}
	}
	return true, 0, nil
}

// ParseFlagInt parses an int flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's int value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagInt(args []string) (int, int, error) {
	rvStr, used, err := ParseFlagString(args)
	if err != nil {
		return 0, used, err
	}
	var rv int
	rv, err = strconv.Atoi(rvStr)
	if err != nil {
		return 0, used, err
	}
	return rv, used, nil
}

// ReadFile reads a file and splits it into lines.
func ReadFile(filename string) ([]string, error) {
	defer FuncEndingAlways(FuncStarting(filename))
	DebugAlwaysf("Reading file: %s", filename)
	dat, err := os.ReadFile(filename)
	if err != nil {
		Stderrf("error reading file: %v", err)
		return []string{}, err
	}
	rv := strings.Split(string(dat), "\n")
	for len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	return rv, nil
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Environment Variable Handling  --------------------------------
// -------------------------------------------------------------------------------------------------

// HandleEnvVars looks at specific environment variables and sets global variables appropriately.
func HandleEnvVars() error {
	var err error
	debug, err = GetEnvVarBool("DEBUG")
	if debug {
		Stderrf("Debugging enabled via environment variable.")
	}
	return err
}

// GetEnvVarBool gets the environment variable with the given name and converts it to a bool.
func GetEnvVarBool(name string) (bool, error) {
	str := os.Getenv(name)
	if len(str) == 0 {
		return false, nil
	}
	val, isBool := ParseBool(str)
	if !isBool {
		return false, fmt.Errorf("invalid %s env var boolean value: [%s]", name, str)
	}
	return val, nil
}

// -------------------------------------------------------------------------------------------------
// ------------------------------  Function start/stop timing stuff  -------------------------------
// -------------------------------------------------------------------------------------------------

// If all you want is starting/ending messages when debug is on, use:
//    defer FuncEnding(FuncStarting())
// If, when debug is on, you want starting/ending messages,
// but when debug is off, you still want the function duration, then use:
//    defer FuncEndingAlways(FuncStarting())

// FuncStarting outputs that a function is starting (if debug is true).
// It returns the params needed by FuncEnding or FuncEndingAlways.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Only provide minimal values needed to differentiate start/stop output lines.
// Long strings and complex structs should be avoided as args.
//
// Example 1: In a function named "foo", you have this:
//
//	  FuncStarting()
//	The printed message will note that "foo" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//
//	  FuncStarting(3 * time.Second)
//	The printed message will note that "bar: 3s" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//
//	  func sum(ints ...int) {
//	      FuncStarting(ints...)
//	  }
//	  sum(1, 2, 3, 4, 20, 21, 22)
//	The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer FuncEnding(FuncStarting())
//
//	Or: defer FuncEndingAlways(FuncStarting())
func FuncStarting(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugAsf(name, "Starting.")
	return time.Now(), name
}

// FuncStartingAlways is the same as FuncStarting except if debug is off, output will go to stdout.
//
// This differs from FuncStarting in that this will always do the output (regardless of debug state).
//
// Usage: defer FuncEndingAlways(FuncStartingAlways())
func FuncStartingAlways(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugAlwaysAsf(name, "Starting.")
	return time.Now(), name
}

const DONE_FMT = "Done. Duration: [%s]."

var panicPrinted bool

// FuncEnding decrements the function depth and, if debug is on, outputs to stderr how long a function took.
// Args will usually come from FuncStarting().
//
// This differs from FuncEndingAlways in that this only outputs something if debugging is turned on.
//
// Usage: defer FuncEnding(FuncStarting())
func FuncEnding(start time.Time, name string) {
	if !panicPrinted {
		if r := recover(); r != nil {
			DebugAlwaysAsf(name, "PANIC")
			panicPrinted = true
			defer func() {
				panic(r)
			}()
		}
	}
	if !panicPrinted {
		DebugAsf(name, DONE_FMT, time.Since(start))
	}
	if funcDepth > -1 {
		funcDepth--
	}
}

// FuncEndingAlways is the same as FuncEnding except if debug is off, output will go to stdout.
//
// This differs from FuncEnding in that this will always do the output (regardless of debug state).
//
// Usage: defer FuncEndingAlways(FuncStarting())
func FuncEndingAlways(start time.Time, name string) {
	if !panicPrinted {
		if r := recover(); r != nil {
			DebugAlwaysAsf(name, "PANIC")
			panicPrinted = true
			defer func() {
				panic(r)
			}()
		}
	}
	if !panicPrinted {
		DebugAlwaysAsf(name, DONE_FMT, time.Since(start))
	}
	if funcDepth > -1 {
		funcDepth--
	}
}

// DurClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNNs", e.g. "12:01:02.000000000".
// - If less than one hour, format is "M:SS.NNNNNNNNNs",   e.g. "34:00.000000789".
// - If less than one minute, format is "S.NNNNNNNNNs",    e.g. "56.000456000".
// - If less than one second, format is "0.NNNNNNNNNs",    e.g. "0.123000000".
func DurClock(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	s := int(d.Seconds())
	n := int(d.Nanoseconds()) - 1000000000*s
	s -= 60 * m
	m -= 60 * h
	switch {
	case h > 0:
		return fmt.Sprintf("%d:%02d:%02d.%09d", h, m, s, n)
	case m > 0:
		return fmt.Sprintf("%d:%02d.%09d", m, s, n)
	default:
		return fmt.Sprintf("%d.%09d", s, n)
	}
}

// GetFuncName gets the name of the function at the given depth.
//
// Depth 0 = the function calling GetFuncName.
// Depth 1 = the function calling the function calling GetFuncName.
// Etc.
//
// Extra arguments provided will be converted to stings using %v and included as part of the function name.
// Only values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
func GetFuncName(depth int, a ...interface{}) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	for more && depth > 0 {
		frame, more = frames.Next()
		depth--
	}
	name := strings.TrimPrefix(frame.Function, "main.")
	// Using a switch to prevent calling strings.Join for small (common) use cases. Saves a little mem and processing.
	switch len(a) {
	case 0:
		// do nothing
	case 1:
		name += fmt.Sprintf(": %v", a[0])
	case 2:
		name += fmt.Sprintf(": %v, %v", a[0], a[1])
	case 3:
		name += fmt.Sprintf(": %v, %v, %v", a[0], a[1], a[2])
	default:
		args := make([]string, len(a))
		for i, arg := range a {
			args[i] = fmt.Sprintf("%v", arg)
		}
		name += fmt.Sprintf(": %s", strings.Join(args, ", "))
	}
	return name
}

// GetMyExe returns how to execute this program by parsing os.Args[0].
func GetMyExe() string {
	_, name := filepath.Split(os.Args[0])
	if i := strings.Index(os.Args[0], "/go-build"); i == -1 {
		name = "./" + name
	} else {
		name = fmt.Sprintf("go run %s.go", name)
	}
	return name
}

// -------------------------------------------------------------------------------------------------
// ---------------------------------------  Output wrappers  ---------------------------------------
// -------------------------------------------------------------------------------------------------

// GetOutputPrefix gets the prefix to add to all output.
func GetOutputPrefix(funcName string) string {
	tabs := ""
	if debug && funcDepth > 0 {
		tabs = strings.Repeat("  ", funcDepth)
	}
	return fmt.Sprintf("(%14s) %s[%s] ", DurClock(time.Since(startTime)), tabs, funcName)
}

// Stdoutf outputs to stdout with a prefixed run duration and automatic function name.
func Stdoutf(format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// Stderrf outputs to stderr with a prefixed run duration and automatic function name.
func Stderrf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// StdoutAsf outputs to stdout with a prefixed run duration and provided function name.
func StdoutAsf(funcName, format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(funcName)+format+"\n", a...)
}

// StderrAsf outputs to stderr with a prefixed run duration and provided function name.
func StderrAsf(funcName, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(funcName)+format+"\n", a...)
}

// Debugf is like Stderrf if the debug flag is set; otherwise it does nothing.
func Debugf(format string, a ...interface{}) {
	if debug {
		StderrAsf(GetFuncName(1), format, a...)
	}
}

// DebugAsf is like StderrAsf if the debug flag is set; otherwise it does nothing.
func DebugAsf(funcName, format string, a ...interface{}) {
	if debug {
		StderrAsf(funcName, format, a...)
	}
}

// DebugAlwaysf is like Stderrf if the debug flag is set; otherwise it's like Stdoutf.
func DebugAlwaysf(format string, a ...interface{}) {
	if debug {
		StderrAsf(GetFuncName(1), format, a...)
	} else {
		StdoutAsf(GetFuncName(1), format, a...)
	}
}

// DebugAlwaysAsf is like StderrAsf if the debug flag is set; otherwise it's like StdoutAsf.
func DebugAlwaysAsf(funcName, format string, a ...interface{}) {
	if debug {
		StderrAsf(funcName, format, a...)
	} else {
		StdoutAsf(funcName, format, a...)
	}
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Primary Program Running Parts  --------------------------------
// -------------------------------------------------------------------------------------------------

// debug is a flag for whether or not debug messages should be displayed.
var debug bool

// startTime is the time when the program started.
var startTime time.Time

// funcDepth is a global counter keeping track of function depth by the starting/ending function functions.
var funcDepth int

func init() {
	funcDepth = -1
}

// main is the main function that gets run for this file.
func main() {
	startTime = time.Now()
	// Handle the env vars before calling into Run().
	// That way, if debug is on, we will get the start message for Run().
	err := HandleEnvVars()
	if err == nil {
		err = Run()
	}
	if err != nil {
		// Not using Stderrf(...) here because I don't want the time and function prefix on this.
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run does all the primary coordination for this program.
// It's basically a replacement for main() that returns an error.
func Run() error {
	defer FuncEndingAlways(FuncStarting())
	params := GetParams(os.Args[1:])
	if params.HelpPrinted {
		return nil
	}
	if !params.HasError() {
		var err error
		params.Input, err = ReadFile(params.InputFile)
		params.AppendError(err)
	}
	Debugf("Params:\n%s", params)
	if params.HasError() {
		return params.GetError()
	}
	answer, err := Solve(params)
	if err != nil {
		return err
	}
	Stdoutf("Answer: %s", answer)
	return nil
}
