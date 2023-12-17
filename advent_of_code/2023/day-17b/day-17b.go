package main

import (
	"container/heap"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
	solver := NewSolver(input.Blocks)
	solver.Run(params)
	answer := -999999999999999999
	if solver.Solution != nil {
		answer = solver.Solution.MinCost
		path := solver.Solution.GetRoute()
		Stdoutf("Solution (%d):\n%s", len(path), CreateIndexedGridStringInts(input.Blocks, path, nil))
	}
	return fmt.Sprintf("%d", answer), nil
}

type Solver struct {
	KnownMins [][]int
	Blocks    [][]int
	Nodes     [][][]*PQNode
	Width     int
	Height    int
	Unvisited PriorityQueue
	Last      *PQNode
	Solution  *PQNode
}

func (s Solver) String() string {
	return fmt.Sprintf("Unvisited: %d, Solution: %s", len(s.Unvisited), s.Solution)
}

func (s Solver) IsDone() bool {
	return s.Solution != nil || len(s.Unvisited) == 0
}

func NewSolver(blocks [][]int) *Solver {
	rv := &Solver{
		Blocks:    blocks,
		Width:     len(blocks[0]),
		Height:    len(blocks),
		Unvisited: make(PriorityQueue, 0, 100),
	}
	rv.KnownMins = make([][]int, len(blocks))
	rv.Nodes = make([][][]*PQNode, len(blocks))
	for y, row := range blocks {
		rv.KnownMins[y] = make([]int, len(row))
		rv.Nodes[y] = make([][]*PQNode, len(row))
		for x := range row {
			rv.KnownMins[y][x] = 1009 // Picked because that was my first answer that it said was too high.
		}
	}
	rv.KnownMins[0][0] = 0
	start := NewPQNode(0, 0, blocks)
	start.MinCost = 0
	rv.Unvisited = append(rv.Unvisited, start)
	Debugf("Initializing Solver: %s", rv)
	heap.Init(&rv.Unvisited)
	return rv
}

func (s *Solver) Run(params *Params) {
	defer FuncEnding(FuncStarting())
	i := 0
	for !s.IsDone() {
		i++
		s.CalculateNext()
		Debugf("After %d moves. Solver: %s", i, s)
		if params.Verbose && i%250_000 == 0 && !debug {
			Stderrf("After %d moves. Solver: %s", i, s)
			path := s.Last.TracePath(0)
			Stderrf("Last: %s", s.Last)
			Stderrf("Path (%d): %s", len(path), path)
			if len(s.Blocks) < 20 {
				Stderrf("%s", CreateIndexedGridStringInts(s.KnownMins, path, []Point{s.Last.Point}))
			}
		}
		if params.Count > 0 && params.Count <= i {
			break
		}
	}
	Stdoutf("After %d moves. Solver: %s", i, s)
}

func (s Solver) FindSamePrevs(node *PQNode) *PQNode {
	return FindSamePrevs(s.Nodes[node.Y][node.X], node)
}

func (s *Solver) CalculateNext() {
	defer FuncEnding(FuncStarting())
	keyNode, _ := heap.Pop(&s.Unvisited).(*PQNode)
	s.Last = keyNode
	if debug {
		path := keyNode.TracePath(0)
		Stderrf("Key Node: %s", keyNode)
		Stderrf("    Path: %s\n%s", path, CreateIndexedGridStringInts(s.KnownMins, path, []Point{keyNode.Point}))
	}

	nextMoves := s.GetNextMoves(keyNode)
	Debugf("Next Moves: %v", nextMoves)

	for _, move := range nextMoves {
		Debugf(" Checking move: %s", move)
		if move.Visited {
			Debugf("  Already visited.")
			continue
		}
		if keyNode.InPast(move) {
			Debugf("  In Past.")
			continue
		}

		if move.MinCost < s.KnownMins[move.Y][move.X] {
			s.KnownMins[move.Y][move.X] = move.MinCost
		}

		if move.X == s.Width-1 && move.Y == s.Height-1 {
			s.Solution = move
			Debugf("  Found the end: %s", s.Solution)
			break
		}

		match := s.FindSamePrevs(move)
		if match != nil && match.MinCost <= move.MinCost {
			Debugf("  Not better than existing entry: %s", match)
			continue
		}

		Debugf("  Adding to queue: %s", move)
		heap.Push(&s.Unvisited, move)
		s.Nodes[move.Y][move.X] = append(s.Nodes[move.Y][move.X], move)
	}
}

func (s Solver) GetNextMoves(cur *PQNode) []*PQNode {
	defer FuncEnding(FuncStarting())
	Debugf("Cur: %s From: %s", cur, cur.Prev)
	prevs := cur.Trace(10)
	Debugf("Last 10: %s", prevs)
	canTurn, mustTurn := false, false
	if len(prevs) >= 4 {
		dx := Abs(cur.X - prevs[3].X)
		dy := Abs(cur.Y - prevs[3].Y)
		Debugf("Last  4: dx: %d, dy: %d", dx, dy)
		if (dx == 0 && dy >= 4) || (dx >= 4 && dy == 0) {
			canTurn = true
			Debugf("Can turn.")
		}
	}
	if len(prevs) >= 10 {
		dx := Abs(cur.X - prevs[9].X)
		dy := Abs(cur.Y - prevs[9].Y)
		Debugf("Last 10: dx: %d, dy: %d", dx, dy)
		if (dx == 0 && dy >= 10) || (dx >= 10 && dy == 0) {
			mustTurn = true
			Debugf("Must turn.")
		}
	}

	canD := cur.Y < s.Height-1
	canU := cur.Y > 0
	canL := cur.X > 0
	canR := cur.X < s.Width-1
	if debug {
		if !canD {
			Debugf("Cannot go down because there's nothing there.")
		}
		if !canU {
			Debugf("Cannot go up because there's nothing there.")
		}
		if !canL {
			Debugf("Cannot go left because there's nothing there.")
		}
		if !canR {
			Debugf("Cannot go right because there's nothing there.")
		}
	}

	if cur.Prev != nil {
		dx := cur.X - cur.Prev.X
		dy := cur.Y - cur.Prev.Y
		switch {
		case dx == 0 && dy == -1:
			Debugf("Cannot go down because that's where we came from.")
			canD = false
			if !canTurn {
				Debugf("Cannot go left or right because we haven't gone far enough yet.")
				canL = false
				canR = false
			}
			if mustTurn {
				Debugf("Cannot go up because we must turn.")
				canU = false
			}
		case dx == 0 && dy == 1:
			Debugf("Cannot go up because that's where we came from.")
			canU = false
			if !canTurn {
				Debugf("Cannot go left or right because we haven't gone far enough yet.")
				canL = false
				canR = false
			}
			if mustTurn {
				Debugf("Cannot go down because we must turn.")
				canD = false
			}
		case dx == -1 && dy == 0:
			Debugf("Cannot go right because that's where we came from.")
			canR = false
			if !canTurn {
				Debugf("Cannot go up or down because we haven't gone far enough yet.")
				canU = false
				canD = false
			}
			if mustTurn {
				Debugf("Cannot go left because we must turn.")
				canL = false
			}
		case dx == 1 && dy == 0:
			Debugf("Cannot go left because that's where we came from.")
			canL = false
			if !canTurn {
				Debugf("Cannot go up or down because we haven't gone far enough yet.")
				canU = false
				canD = false
			}
			if mustTurn {
				Debugf("Cannot go right because we must turn.")
				canR = false
			}
		default:
			panic(fmt.Errorf("cannot determine previous step: Cur: %s, Prev: %s, dx: %d, dy: %d",
				cur, cur.Prev, dx, dy))
		}
	}

	rv := make([]*PQNode, 0, 4)
	if canU {
		Debugf("Can go up.")
		rv = append(rv, NewPQNode(cur.X, cur.Y-1, s.Blocks))
	}
	if canD {
		Debugf("Can go down.")
		rv = append(rv, NewPQNode(cur.X, cur.Y+1, s.Blocks))
	}
	if canL {
		Debugf("Can go left.")
		rv = append(rv, NewPQNode(cur.X-1, cur.Y, s.Blocks))
	}
	if canR {
		Debugf("Can go right.")
		rv = append(rv, NewPQNode(cur.X+1, cur.Y, s.Blocks))
	}

	for _, v := range rv {
		v.SetPrev(cur)
		v.MinCost = cur.MinCost + v.Cost
	}

	return rv
}

func Abs(x int) int {
	if x > 0 {
		return x
	}
	return -1 * x
}

type PQNode struct {
	Point
	Cost    int
	MinCost int
	Index   int
	Visited bool
	Queued  bool
	Length  int
	Prev    *PQNode
	Prev10  []*PQNode
}

func NewPQNode(x, y int, costs [][]int) *PQNode {
	return &PQNode{
		Point:   Point{X: x, Y: y},
		Cost:    costs[y][x],
		MinCost: MAX_INT,
		Index:   -1,
	}
}

func (n PQNode) String() string {
	prev := "Prev: <nil>"
	if n.Prev != nil {
		prev = fmt.Sprintf("Prev: %s", n.Prev.Point)
	}
	return fmt.Sprintf("%s%d = %d, %s", n.Point, n.Cost, n.MinCost, prev)
}

func (n *PQNode) SetPrev(prev *PQNode) {
	n.Prev = prev
	n.Length = GetLength(n)
	n.Prev10 = n.Trace(10)
}

func SamePoints(n1, n2 *PQNode) bool {
	return n1 != nil && n2 != nil && n1.X == n2.X && n1.Y == n2.Y
}

func SamePrevs(n1, n2 *PQNode) bool {
	if n1 == nil || n2 == nil || !SamePoints(n1, n2) || len(n1.Prev10) != len(n2.Prev10) {
		return false
	}
	for i := range n1.Prev10 {
		if !SamePoints(n1.Prev10[i], n2.Prev10[i]) {
			return false
		}
	}
	return true
}

func FindSamePrevs(existing []*PQNode, toFind *PQNode) *PQNode {
	for _, node := range existing {
		if SamePrevs(node, toFind) {
			return node
		}
	}
	return nil
}

func GetLength(node *PQNode) int {
	rv := 0
	cur := node
	for cur != nil {
		rv++
		cur = cur.Prev
	}
	return rv
}

// Trace gets the previous nodes up to the provided count.
// If count is zero or negative, all previous nodes are retrieved.
// The first returned node is the node just before this one, the last is the start.
// I.e. They are in reverse order.
func (n PQNode) Trace(count int) []*PQNode {
	var rv []*PQNode
	if count > 0 {
		rv = make([]*PQNode, 0, count)
	}
	cur := &n
	for cur.Prev != nil && (count <= 0 || len(rv) < count) {
		cur = cur.Prev
		rv = append(rv, cur)
	}
	return rv
}

// TracePath gets the points of the previous nodes up to the count provided.
// If count is zero or negative, all previous points are retrieved.
// They are ordered with the starting node first, and the one before this one last.
func (n PQNode) TracePath(count int) Path {
	nodes := n.Trace(count)
	rv := make(Path, len(nodes))
	for i, n := range nodes {
		rv[len(nodes)-i-1] = n.Point
	}
	return rv
}

func (n PQNode) GetRoute() []*PQNode {
	rv := n.Trace(0)
	slices.Reverse(rv)
	rv = append(rv, &n)
	return rv
}

func (n PQNode) InPast(toFind *PQNode) bool {
	cur := &n
	for cur != nil {
		if cur.X == toFind.X && cur.Y == toFind.Y {
			return true
		}
		cur = cur.Prev
	}
	return false
}

type PriorityQueue []*PQNode

func (q PriorityQueue) String() string {
	return StringNumberJoin(q, 0, "\n")
}

func (q PriorityQueue) Len() int {
	return len(q)
}

func (q PriorityQueue) Less(i, j int) bool {
	cm := CompInt(q[i].MinCost, q[j].MinCost)
	if cm != 0 {
		return cm < 0
	}

	cl := CompInt(q[i].Length, q[j].Length)
	if cl != 0 {
		return cl > 0
	}

	cx := CompInt(q[i].X, q[j].X)
	if cx != 0 {
		return cx > 0
	}

	cy := CompInt(q[i].Y, q[j].Y)
	if cy != 0 {
		return cy > 0
	}

	return q[i].Index < q[j].Index
}

// CompInt returns -1 if a < b, 0 if a == b, or 1 if a > b.
func CompInt(a, b int) int {
	if a < b {
		return -1
	}
	if a == b {
		return 0
	}
	return 1
}

func (q PriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Index = i
	q[j].Index = j
}

func (q *PriorityQueue) Push(x interface{}) {
	node, _ := x.(*PQNode)
	node.Index = len(*q)
	node.Queued = true
	*q = append(*q, node)
}

func (q *PriorityQueue) Pop() interface{} {
	old := *q
	n := len(old)
	node := old[n-1]
	old[n-1] = nil  // avoid memory leak
	node.Index = -1 // for safety
	node.Visited = true
	*q = old[0 : n-1]
	return node
}

func (q *PriorityQueue) Update(node *PQNode, minCost int, prev *PQNode) {
	node.MinCost = minCost
	node.SetPrev(prev)
	heap.Fix(q, node.Index)
}

type Input struct {
	Blocks [][]int
}

func (i Input) String() string {
	return fmt.Sprintf("Blocks (%d, %d):\n%s", len(i.Blocks[0]), len(i.Blocks),
		CreateIndexedGridStringInts(i.Blocks, ([]XY)(nil), nil))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Blocks: make([][]int, len(lines))}
	var err error
	for i, line := range lines {
		rv.Blocks[i], err = SplitParseInts(line, "")
		if err != nil {
			return nil, fmt.Errorf("failed to parts line %q: %w", line, err)
		}
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

type Path []Point

func (p Path) String() string {
	return StringNumberJoin(p, 1, " ")
}

// A Point contains an X and Y value.
type Point struct {
	X int
	Y int
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

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// XY is something that has an X and Y value.
type XY interface {
	GetX() int
	GetY() int
	GetXY() (int, int)
	String() string
}

func CreateIndexedGridStringInts[S ~[]E, E XY](vals [][]int, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, cell := range row {
			strs[y][x] = fmt.Sprintf("%d", cell)
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
