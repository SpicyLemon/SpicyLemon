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

const DEFAULT_COUNT = 26501365

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	if params.HasCustom("count") {
		CountSpaces(input.Garden)
		return "STOP", nil
	}
	Debugf("Parsed Input:\n%s", input)
	solver := NewSolver(input.Garden, params.Count, input.Start)
	solver.Wander(params)
	if params.HasCustom("grow") {
		WatchGrow(solver)
	}
	if params.Verbose {
		spots := solver.GetSolutions(params.Count)
		grid := input.Garden.Copy()
		var col []XY
		var shown, notShown int
		for _, p := range spots {
			if grid.Has(p) {
				grid[p.Y][p.X] = 'O'
				col = append(col, p)
				shown++
			} else {
				notShown++
			}
		}
		hl := []XY{input.Start}
		Stdoutf("Solution:\n%s", CreateIndexedGridStringBz(grid, col, hl))
		Stdoutf("    Shown: %d", shown)
		Stdoutf("Not Shown: %d", notShown)
		Stdoutf("  Max Min: %d", solver.GetMaxMin())
	}
	answer := solver.ExpandAndCount(params)
	return fmt.Sprintf("%d", answer), nil
}

func (s *Solver) ExpandAndCount(params *Params) int {
	ms := s.GetMinsSections()
	if params.HasCustom("mins") {
		s.PrintMinsSections(ms)
	}
	minGrids := make(map[int]map[int][][]int)
	xLim := InitMinMax()
	yLim := InitMinMax()
	haveMatrix := false
	for y, row := range ms {
		minGrids[y] = make(map[int][][]int)
		yLim.Include(y)
		for x, cell := range row {
			minGrids[y][x] = cell.ToMatrix(s.Width, s.Height)
			xLim.Include(x)
			haveMatrix = true
		}
	}
	if !haveMatrix {
		Stderrf("No grids identified.")
		return 0
	}

	leftDiffs := make([]map[int][][]int, s.Extra)
	rightDiffs := make([]map[int][][]int, s.Extra)
	for d := 0; d < s.Extra; d++ {
		leftDiffs[d] = make(map[int][][]int)
		rightDiffs[d] = make(map[int][][]int)
		yLim.Iter(func(y int) {
			leftDiffs[d][y] = MSub(minGrids[y][xLim.Min+d], minGrids[y][xLim.Min+d+1])
			rightDiffs[d][y] = MSub(minGrids[y][xLim.Max-d], minGrids[y][xLim.Max-d-1])
		})
	}

	upDiffs := make([]map[int][][]int, s.Extra)
	downDiffs := make([]map[int][][]int, s.Extra)
	for d := 0; d < s.Extra; d++ {
		upDiffs[d] = make(map[int][][]int)
		downDiffs[d] = make(map[int][][]int)
		xLim.Iter(func(x int) {
			upDiffs[d][x] = MSub(minGrids[yLim.Min+d][x], minGrids[yLim.Min+d+1][x])
			downDiffs[d][x] = MSub(minGrids[yLim.Max-d][x], minGrids[yLim.Max-d-1][x])
		})
	}

	if params.HasCustom("diffs") {
		upRows := make([]string, len(upDiffs))
		for d, diffs := range upDiffs {
			sects := make([]string, 0, xLim.Count())
			xLim.Iter(func(x int) {
				sects = append(sects, CreateIndexedGridStringMins(diffs[x], []XY{}, nil))
			})
			upRows[d] = JoinSections(sects)
		}

		downRows := make([]string, len(downDiffs))
		for d, diffs := range downDiffs {
			sects := make([]string, 0, xLim.Count())
			xLim.Iter(func(x int) {
				sects = append(sects, CreateIndexedGridStringMins(diffs[x], []XY{}, nil))
			})
			downRows[len(downRows)-d-1] = JoinSections(sects)
		}

		lineCount := strings.Count(upRows[0], "\n") + 1
		spacerLines := make([]string, lineCount)
		for i := range spacerLines {
			spacerLines[i] = strings.Repeat(" ", 4)
		}
		spacerSect := strings.Join(spacerLines, "\n")

		lrRows := make([]string, 0, yLim.Count())
		yLim.Iter(func(y int) {
			sects := make([]string, 0, len(leftDiffs)+len(rightDiffs)+1)
			for _, diffs := range leftDiffs {
				sects = append(sects, CreateIndexedGridStringMins(diffs[y], []XY{}, nil))
			}
			sects = append(sects, spacerSect)
			for d := range rightDiffs {
				sects = append(sects, CreateIndexedGridStringMins(rightDiffs[len(rightDiffs)-d-1][y], []XY{}, nil))
			}
			lrRows = append(lrRows, JoinSections(sects))
		})

		ups := strings.Join(upRows, "\n")
		downs := strings.Join(downRows, "\n")
		lrs := strings.Join(lrRows, "\n")

		Stdoutf("Up Diffs:\n%s", ups)
		Stdoutf("Down Diffs:\n%s", downs)
		Stdoutf("Left and Right Diffs:\n%s", lrs)
	}

	if params.Verbose {
		rv := 0
		secCount := 0
		for _, row := range minGrids {
			for _, minGrid := range row {
				rv += CountVals(minGrid, s.MaxSteps)
				secCount++
			}
		}

		Stdoutf("In the initial %d sections, there are %d valid spots.", secCount, rv)
	}

	var notUniform []*Point
	var sectionDiff int
	addNotUniform := func(mm *MinMax, x, y int) {
		if mm.Count() != 1 {
			notUniform = append(notUniform, NewPoint(x, y))
		} else {
			sectionDiff = mm.Min
		}
	}
	xLim.Iter(func(x int) {
		addNotUniform(MMinMax(upDiffs[0][x]), x, -1*s.Extra)
		addNotUniform(MMinMax(downDiffs[0][x]), x, s.Extra)
	})
	yLim.Iter(func(y int) {
		addNotUniform(MMinMax(leftDiffs[0][y]), -1*s.Extra, y)
		addNotUniform(MMinMax(rightDiffs[0][y]), s.Extra, y)
	})

	if params.Verbose {
		if len(notUniform) > 0 {
			panic(fmt.Errorf("the following sections aren't uniform: %s", notUniform))
		} else {
			Stdoutf("All sections are uniform.")
		}
	}

	// Find all the full sections.
	xFullRanges := make(map[int]*MinMax)
	yFullRanges := make(map[int]*MinMax)
	xHasSpotsRanges := make(map[int]*MinMax)
	yHasSpotsRanges := make(map[int]*MinMax)
	updateRangeMap := func(rmap map[int]*MinMax, ind, val int) {
		r := rmap[ind]
		if r != nil {
			r.Include(val)
		} else {
			rmap[ind] = NewMinMax(0, val)
		}
	}
	noteHasSpots := func(x, y int) {
		updateRangeMap(xHasSpotsRanges, y, x)
		updateRangeMap(yHasSpotsRanges, x, y)
	}
	noteFull := func(x, y int) {
		updateRangeMap(xFullRanges, y, x)
		updateRangeMap(yFullRanges, x, y)
		noteHasSpots(x, y)
	}

	params.Verbosef("Finding full sections horizontally.")
	// Start horizontally.
	yLim.Iter(func(y int) {
		fullNl, partNl := s.FindNs(minGrids[y][xLim.Min], sectionDiff)
		fullNr, partNr := s.FindNs(minGrids[y][xLim.Max], sectionDiff)
		noteFull(xLim.Min-fullNl, y)
		noteFull(xLim.Max+fullNr, y)
		noteHasSpots(xLim.Min-partNl, y)
		noteHasSpots(xLim.Max+partNr, y)
	})

	// Expand them upwards.
	params.Verbosef("Expanding upwards from x = %s and y = %d.", xHasSpotsRanges[yLim.Min], yLim.Min)
	xHasSpotsRanges[yLim.Min].Iter(func(x int) {
		var mins [][]int
		switch {
		case x < xLim.Min:
			mins = MAddFlat(minGrids[yLim.Min][xLim.Min], Abs(x-xLim.Min)*sectionDiff)
		case x > xLim.Max:
			mins = MAddFlat(minGrids[yLim.Min][xLim.Max], Abs(x-xLim.Max)*sectionDiff)
		default:
			mins = minGrids[yLim.Min][x]
		}
		fullN, partN := s.FindNs(mins, sectionDiff)
		if fullN > 0 {
			noteFull(x, yLim.Min-fullN)
		}
		if partN > 0 {
			noteHasSpots(x, yLim.Min-partN)
		}
	})

	// And now downwards.
	params.Verbosef("Expanding downwards from x = %s and y = %d.", xHasSpotsRanges[yLim.Max], yLim.Max)
	xHasSpotsRanges[yLim.Max].Iter(func(x int) {
		var mins [][]int
		switch {
		case x < xLim.Min:
			mins = MAddFlat(minGrids[yLim.Max][xLim.Min], Abs(x-xLim.Min)*sectionDiff)
		case x > xLim.Max:
			mins = MAddFlat(minGrids[yLim.Max][xLim.Max], Abs(x-xLim.Max)*sectionDiff)
		default:
			mins = minGrids[yLim.Max][x]
		}
		fullN, partN := s.FindNs(mins, sectionDiff)
		if fullN > 0 {
			noteFull(x, yLim.Max+fullN)
		}
		if partN > 0 {
			noteHasSpots(x, yLim.Max+partN)
		}
	})

	if debug {
		xs := IntKeys(yFullRanges)
		ys := IntKeys(xFullRanges)

		xFulls := make([]string, 0, len(xFullRanges))
		for _, y := range ys {
			xFulls = append(xFulls, fmt.Sprintf("y[% d]: %s", y, xFullRanges[y]))
		}
		yFulls := make([]string, 0, len(yFullRanges))
		for _, x := range xs {
			yFulls = append(yFulls, fmt.Sprintf("x[ %d]: %s", x, yFullRanges[x]))
		}
		Stdoutf("X Full Ranges:\n%s", strings.Join(xFulls, "\n"))
		Stdoutf("Y Full Ranges:\n%s", strings.Join(yFulls, "\n"))
	}

	fullEven := CountVals(minGrids[0][0], s.MaxSteps)
	fullOdd := CountVals(minGrids[0][1], s.MaxSteps)
	params.Verbosef("A full even grid (e.g. 0, 0) has %d valid spots", fullEven)
	params.Verbosef("A full  odd grid (e.g. 0, 1) has %d valid spots", fullOdd)

	totalFulls := 0
	for _, xr := range xFullRanges {
		evens := (xr.Max/2)*2 + 1 // Need the integer truncation on that division.
		odds := ((xr.Max + 1) / 2) * 2
		totalFulls += evens*fullEven + odds*fullOdd
	}

	params.Verbosef("The full grids combine to a total of %d", totalFulls)

	// And (hopefully) lastly, count all the partials.
	totalPartial := 0
	xs := IntKeys(yHasSpotsRanges)
	for _, x := range xs {
		ypr := yHasSpotsRanges[x]
		var ys []int
		if yr := yFullRanges[x]; yr != nil {
			for y := ypr.Min; y < yr.Min; y++ {
				ys = append(ys, y)
			}
			for y := yr.Max + 1; y <= ypr.Max; y++ {
				ys = append(ys, y)
			}
		} else {
			ypr.Iter(func(y int) {
				ys = append(ys, y)
			})
		}

		Debugf("Partials for x = %d: y = %v", x, ys)
		for _, y := range ys {
			var bx, by int
			var dx, dy int
			switch {
			case x < xLim.Min:
				bx = xLim.Min
				dx = Abs(x - bx)
			case x > xLim.Max:
				bx = xLim.Max
				dx = Abs(x - bx)
			default:
				bx = x
			}
			switch {
			case y < yLim.Min:
				by = yLim.Min
				dy = Abs(y - by)
			case y > yLim.Max:
				by = yLim.Max
				dy = Abs(y - by)
			default:
				by = y
			}

			totalDiff := (dx + dy) * sectionDiff
			mins := MAddFlat(minGrids[by][bx], totalDiff)
			mr := MMinMax(mins)
			nv := CountVals(mins, s.MaxSteps)
			Debugf("Partial: %d, %d = (%d, %d)+(%d, %d) with total diff %d it ends up with %s and has %d spots.",
				x, y, bx, by, dx, dy, totalDiff, mr, nv)
			if s.MaxSteps < mr.Min || mr.Max <= s.MaxSteps {
				Stderrf("Skipping partial %d, %d with range %s", x, y, mr)
				continue
			}
			if debug && x == 5 && y == 457 {
				Stderrf("Section %d, %d:\n%s", x, y, CreateIndexedGridStringMins(mins, []XY{}, nil))
			}
			totalPartial += nv
		}
	}

	params.Verbosef("The partial grids combine to a total of %d", totalPartial)

	return totalFulls + totalPartial
}

func (s Solver) FindNs(mins [][]int, sectionDiff int) (int, int) {
	// Start by going left to find the last full section.
	sRange := MMinMax(mins)
	n := (s.MaxSteps - sRange.Max) / sectionDiff
	sRange = sRange.Add(n * sectionDiff) // probably a section or two low.
	// Back it up if for some reason, it's not a full section. Probably doesn't happen, though.
	for s.MaxSteps < sRange.Max {
		n--
		sRange = sRange.Sub(sectionDiff)
	}
	// Move it forward again until we find the first partial.
	for s.MaxSteps >= sRange.Max {
		n++
		sRange = sRange.Add(sectionDiff)
	}
	// It's now a partial, so the previous one was the last full one.
	fullN := n - 1
	for s.MaxSteps > sRange.Min {
		n++
		sRange = sRange.Add(sectionDiff)
	}
	// It's now completely empty, so the previous one was the last partial one.
	partN := n - 1
	return fullN, partN
}

func OutputMatrix(label string, vals [][]int) {
	Stdoutf("%s %s:\n%s", label, MMinMax(vals), CreateIndexedGridStringMins(vals, []XY{}, nil))
}

func CountVals(vals [][]int, max int) int {
	rv := 0
	parity := max % 2
	for _, row := range vals {
		for _, val := range row {
			if val > 0 && val <= max && val%2 == parity {
				rv++
			}
		}
	}
	return rv
}

func MMinMax(vals [][]int) *MinMax {
	rv := InitMinMax()
	for _, row := range vals {
		for _, val := range row {
			if val != 0 {
				rv.Include(val)
			}
		}
	}
	return rv
}

func (s *Solver) Wander(params *Params) {
	defer FuncEnding(FuncStarting())
	i := 0
	for !s.IsDone() {
		i++
		s.CalculateNext()
		if params.Verbose && i%1_000_000 == 0 {
			Stderrf("After %d million moves, solver = %s", i/1_000_000, s)
		}
	}
	Stdoutf("Solver finished after %d moves: %s", i, s)
}

func (s *Solver) CalculateNext() {
	defer FuncEnding(FuncStarting())
	keyNode, _ := heap.Pop(&s.Unvisited).(*PQNode)
	s.Last = keyNode
	if !s.IsBetter(keyNode) {
		return
	}

	nextMoves := s.GetNextMoves(keyNode)
	for _, move := range nextMoves {
		if !s.IsInLimits(move) {
			continue
		}

		cur := s.GetMin(move)
		if !IsBetter(move, cur) {
			continue
		}

		if cur != nil {
			s.Unvisited.Update(cur, move.Cost, keyNode)
			continue
		}

		s.SetMin(move)
		if move.Cost < s.MaxSteps {
			heap.Push(&s.Unvisited, move)
		}
	}
}

const (
	Garden = byte('.')
	Rocks  = byte('#')
)

func (s Solver) GetNextMoves(cur *PQNode) []*PQNode {
	defer FuncEnding(FuncStarting())
	possibles := cur.GetNextSteps()
	rv := make([]*PQNode, 0, len(possibles))
	for _, p := range possibles {
		if s.IsGarden(p) {
			rv = append(rv, p)
		}
	}
	return rv
}

type Solver struct {
	Garden    Grid
	Start     *PQNode
	MaxSteps  int
	Mins      map[int]map[int]*PQNode
	Width     int
	Height    int
	Unvisited PriorityQueue
	Last      *PQNode
	Solution  *PQNode
	Extra     int
	XLimit    *MinMax
	YLimit    *MinMax
}

func NewSolver(garden Grid, maxSteps int, start *Point) *Solver {
	rv := &Solver{
		Garden:   garden,
		MaxSteps: maxSteps,
		Mins:     make(map[int]map[int]*PQNode),
	}
	rv.Width, rv.Height = garden.GetWH()
	rv.Start = NewPQNode(start.X, start.Y)
	rv.Unvisited.Push(rv.Start)
	rv.SetMin(rv.Start)
	// Through some trial and error, I found that:
	// With actual.input and Extra = 2, all the edge diffs are uniform.
	// With exampe.input and Extra = 4, all the edge diffs are uniform.
	rv.Extra = 2
	if rv.Width < 20 {
		// test.input
		rv.Extra = 4
	}

	rv.XLimit = NewMinMax(-1*rv.Extra*rv.Width, (rv.Extra+1)*rv.Width-1)
	rv.YLimit = NewMinMax(-1*rv.Extra*rv.Height, (rv.Extra+1)*rv.Height-1)
	Debugf("Initializing Solver: Max Steps: %d %s", rv.MaxSteps, rv)
	heap.Init(&rv.Unvisited)
	return rv
}

func (s Solver) String() string {
	return fmt.Sprintf("Unvisited: %d, Last; %s, Solution: %s", len(s.Unvisited), s.Last, s.Solution)
}

func (s Solver) IsDone() bool {
	return s.Solution != nil || len(s.Unvisited) == 0
}

func (s Solver) GetMin(p XY) *PQNode {
	x, y := p.GetXY()
	if s.Mins[y] == nil {
		return nil
	}
	return s.Mins[y][x]
}

func (s Solver) IsBetter(node *PQNode) bool {
	cur := s.GetMin(node)
	return IsBetter(node, cur)
}

func IsBetter(node, cur *PQNode) bool {
	if cur == nil || cur == node {
		return true
	}
	if cur.Visited {
		return false
	}
	return node.Cost < cur.Cost
}

func (s Solver) SetMin(node *PQNode) {
	if s.Mins[node.Y] == nil {
		s.Mins[node.Y] = make(map[int]*PQNode)
	}
	s.Mins[node.Y][node.X] = node
}

func (s Solver) IsGarden(node *PQNode) bool {
	x := ShiftCoord(node.X, s.Width)
	y := ShiftCoord(node.Y, s.Height)
	return s.Garden[y][x] == Garden
}

func ShiftCoord(val, max int) int {
	rv := val % max
	if rv < 0 {
		return rv + max
	}
	return rv
}

func (s Solver) IsInLimits(node *PQNode) bool {
	return s.XLimit.Contains(node.X) && s.YLimit.Contains(node.Y)
}

func (s Solver) GetMaxMin() int {
	rv := 0
	for y := range s.Mins {
		for x := range s.Mins[y] {
			if s.Mins[y][x].Cost > rv {
				rv = s.Mins[y][x].Cost
			}
		}
	}
	return rv
}

func (s Solver) GetSolutions(maxSteps int) []*PQNode {
	var rv []*PQNode
	for _, row := range s.Mins {
		for _, min := range row {
			if min.Cost <= maxSteps && min.Cost%2 == maxSteps%2 {
				rv = append(rv, min)
			}
		}
	}
	return rv
}

type MinsSections map[int]map[int]MinGrid

func (s Solver) GetMinsSections() MinsSections {
	rv := make(map[int]map[int]MinGrid)
	for y := -1 * s.Extra; y <= s.Extra; y++ {
		rv[y] = make(map[int]MinGrid)
		for x := -1 * s.Extra; x <= s.Extra; x++ {
			rv[y][x] = make(MinGrid)
		}
	}

	for y := range s.Mins {
		for x := range s.Mins[y] {
			sx, nx := SectionAndCoord(x, s.Width)
			sy, ny := SectionAndCoord(y, s.Height)
			rv[sy][sx].Set(nx, ny, s.Mins[y][x].Cost)
		}
	}

	return rv
}

func (s Solver) GetMinMatrix() [][]int {
	rv := make([][]int, s.Height)
	for r := range rv {
		rv[r] = make([]int, s.Width)
	}
	s.YLimit.Iter(func(y int) {
		s.XLimit.Iter(func(x int) {
			cell := s.GetMin(Point{X: x, Y: y})
			if cell != nil {
				r := ShiftCoord(y, s.Height)
				c := ShiftCoord(x, s.Width)
				rv[r][c] = cell.Cost
			}
		})
	})
	return rv
}

func SectionAndCoord(val, size int) (int, int) {
	coord := ShiftCoord(val, size)
	if val < 0 {
		return (val - size + 1) / size, coord
	}
	return val / size, coord
}

type MinGrid map[int]map[int]int

func (m MinGrid) Get(x, y int) int {
	if m[y] == nil {
		return MAX_INT
	}
	rv, ok := m[y][x]
	if !ok {
		return MAX_INT
	}
	return rv
}

func (m MinGrid) Set(x, y, min int) {
	if m[y] == nil {
		m[y] = make(map[int]int)
	}
	m[y][x] = min
}

func (m MinGrid) String() string {
	return StringNumberJoin(m.PointVals(), 1, " ")
}

func (m MinGrid) PointVals() PointVals {
	var pvs PointVals
	ys := IntKeys(m)
	for _, y := range ys {
		xs := IntKeys(m[y])
		for _, x := range xs {
			pvs = append(pvs, NewPointVal(x, y, m[y][x]))
		}
	}
	return pvs
}

func (m MinGrid) ToMatrix(width, height int) [][]int {
	xLim := InitMinMax()
	yLim := InitMinMax()
	pvs := m.PointVals()
	for _, p := range pvs {
		xLim.Include(p.X)
		yLim.Include(p.Y)
	}
	if xLim.Min < 0 || yLim.Min < 0 {
		panic(fmt.Errorf("cannot make matrix from negative points: %s", pvs))
	}
	if xLim.Max >= width || yLim.Max >= height {
		panic(fmt.Errorf("cannot make matrix with points over %d x %d: %s", width, height, pvs))
	}
	rv := make([][]int, height)
	for y := range rv {
		rv[y] = make([]int, width)
	}

	for _, p := range pvs {
		rv[p.Y][p.X] = p.Val
	}
	return rv
}

func IntKeys[V any](m map[int]V) []int {
	if m == nil {
		return nil
	}
	rv := make([]int, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	slices.Sort(rv)
	return rv
}

type PointVals []*PointVal

func (p PointVals) String() string {
	return fmt.Sprintf("[%s]", strings.Join(SliceToStrings(p), " "))
}

type PointVal struct {
	Point
	Val int
}

func NewPointVal(x, y, val int) *PointVal {
	return &PointVal{Point: Point{X: x, Y: y}, Val: val}
}

func (p PointVal) String() string {
	return fmt.Sprintf("(%d,%d)=%d", p.X, p.Y, p.Val)
}

func Abs(v int) int {
	if v < 0 {
		return -1 * v
	}
	return v
}

type MinMax struct {
	Min int
	Max int
}

func NewMinMax(min, max int) *MinMax {
	if min > max {
		min, max = max, min
	}
	return &MinMax{Min: min, Max: max}
}

func InitMinMax() *MinMax {
	return &MinMax{
		Min: MAX_INT,
		Max: MIN_INT,
	}
}

func (m MinMax) Add(d int) *MinMax {
	return NewMinMax(m.Min+d, m.Max+d)
}

func (m MinMax) Sub(d int) *MinMax {
	return NewMinMax(m.Min-d, m.Max-d)
}

func (m MinMax) String() string {
	return fmt.Sprintf("[%d, %d]", m.Min, m.Max)
}

func (m MinMax) Contains(val int) bool {
	return m.Min <= val && val <= m.Max
}

func (m *MinMax) Include(val int) {
	if val < m.Min {
		m.Min = val
	}
	if val > m.Max {
		m.Max = val
	}
}

func (m MinMax) Count() int {
	return m.Max - m.Min + 1
}

func (m MinMax) Iter(op func(n int)) {
	for n := m.Min; n <= m.Max; n++ {
		op(n)
	}
}

type PQNode struct {
	Point
	Cost    int
	Index   int
	Visited bool
	Queued  bool
	Prev    *PQNode
}

func NewPQNode(x, y int) *PQNode {
	return &PQNode{
		Point: Point{X: x, Y: y},
		Index: -1,
	}
}

const (
	Up    = byte('U')
	Down  = byte('D')
	Left  = byte('L')
	Right = byte('R')
)

func (n PQNode) Step(direction byte) *PQNode {
	rv := &PQNode{
		Point: Point{X: n.X, Y: n.Y},
		Cost:  n.Cost + 1,
		Index: -1,
		Prev:  &n,
	}
	switch direction {
	case Up:
		rv.Y--
	case Down:
		rv.Y++
	case Left:
		rv.X--
	case Right:
		rv.X++
	default:
		panic(fmt.Errorf("unknown direction: %q", direction))
	}
	return rv
}

func (n PQNode) GetNextSteps() []*PQNode {
	return []*PQNode{
		n.Step(Up),
		n.Step(Right),
		n.Step(Down),
		n.Step(Left),
	}
}

func (n PQNode) String() string {
	return fmt.Sprintf("%s = %d", n.Point, n.Cost)
}

func (n *PQNode) SetPrev(prev *PQNode) {
	n.Prev = prev
}

func (n *PQNode) Equal(n2 *PQNode) bool {
	return SamePoints(n, n2) && n.Cost == n2.Cost
}

func SamePoints(n1, n2 *PQNode) bool {
	return n1 != nil && n2 != nil && n1.X == n2.X && n1.Y == n2.Y
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

// TraceRev gets the points of the previous nodes up to the count provided.
// If count is zero or negative, all previous points are retrieved.
// They are ordered with the starting node first, and the one before this one last.
func (n PQNode) TraceRev(count int) []*PQNode {
	nodes := n.Trace(count)
	slices.Reverse(nodes)
	return nodes
}

// GetRoute gets the route from the start to this node (inclusive).
func (n PQNode) GetRoute() []*PQNode {
	rv := n.TraceRev(0)
	rv = append(rv, &n)
	return rv
}

type PriorityQueue []*PQNode

func (q PriorityQueue) String() string {
	return StringNumberJoin(q, 0, "\n")
}

func (q PriorityQueue) Len() int {
	return len(q)
}

func (q PriorityQueue) Less(i, j int) bool {
	cc := CompInt(q[i].Cost, q[j].Cost)
	if cc != 0 {
		return cc < 0
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
	node.Visited = false
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

func (q *PriorityQueue) Update(node *PQNode, cost int, prev *PQNode) {
	node.Cost = cost
	node.SetPrev(prev)
	heap.Fix(q, node.Index)
}

type Grid [][]byte

func (g Grid) Copy() Grid {
	rv := make(Grid, len(g))
	for y, row := range g {
		rv[y] = make([]byte, len(row))
		copy(rv[y], row)
	}
	return rv
}

func (g Grid) Replicate(xMul, yMul int) Grid {
	w, h := g.GetWH()
	rv := make(Grid, h*yMul)
	for y := range rv {
		rv[y] = make([]byte, w*xMul)
		for x := range rv[y] {
			rv[y][x] = g[y%h][x%w]
		}
	}
	return rv
}

func (g Grid) String() string {
	return CreateIndexedGridStringBz(g, []XY{}, nil)
}

func (g Grid) GetWidth() int {
	if len(g) == 0 {
		return 0
	}
	return len(g[0])
}

func (g Grid) GetHeight() int {
	return len(g)
}

func (g Grid) GetWH() (int, int) {
	w := len(g)
	if w == 0 {
		return 0, 0
	}
	return w, len(g[0])
}

func (g Grid) Has(p XY) bool {
	x, y := p.GetXY()
	return x >= 0 && x < g.GetWidth() && y >= 0 && y < g.GetHeight()
}

func MSub(a, b [][]int) [][]int {
	assertDims(a, b, "subtract")
	if a == nil || b == nil {
		return nil
	}

	rv := make([][]int, len(a))
	for y := range a {
		rv[y] = make([]int, len(a[y]))
		for x := range rv[y] {
			rv[y][x] = a[y][x] - b[y][x]
		}
	}

	return rv
}

func MSubFlat(a [][]int, val int) [][]int {
	if a == nil {
		return nil
	}

	rv := make([][]int, len(a))
	for y := range a {
		rv[y] = make([]int, len(a[y]))
		for x := range rv[y] {
			if a[y][x] != 0 {
				rv[y][x] = a[y][x] - val
			}
		}
	}

	return rv
}

func MAddFlat(a [][]int, val int) [][]int {
	if a == nil {
		return nil
	}

	rv := make([][]int, len(a))
	for y := range a {
		rv[y] = make([]int, len(a[y]))
		for x := range rv[y] {
			if a[y][x] != 0 {
				rv[y][x] = a[y][x] + val
			}
		}
	}

	return rv
}

func MMul(a [][]int, m int) [][]int {
	if a == nil {
		return nil
	}

	rv := make([][]int, len(a))
	for y, row := range a {
		rv[y] = make([]int, len(row))
		for x, v := range row {
			rv[y][x] = v * m
		}
	}

	return rv
}

func MCopy(a [][]int) [][]int {
	return MMul(a, 1)
}

func assertDims(a, b [][]int, action string) {
	if len(a) != len(b) {
		panic(fmt.Errorf("cannot %s matrixes with different numbers of rows: %d vs %d\na: %v\n b: %v",
			action, len(a), len(b), a, b))
	}
	if len(a) > 0 && len(a[0]) != len(b[0]) {
		panic(fmt.Errorf("cannot %s matrixes with different numbers of columns: %d vs %d\na: %v\n b: %v",
			action, len(a[0]), len(b[0]), a, b))
	}
}

type Input struct {
	Garden Grid
	Start  *Point
}

func (i Input) String() string {
	return fmt.Sprintf("Start: %s\nGarden (%dx%d):\n%s", i.Start,
		len(i.Garden[0]), len(i.Garden), CreateIndexedGridStringBz(i.Garden, nil, []*Point{i.Start}))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Garden: make(Grid, len(lines))}
	for y, line := range lines {
		rv.Garden[y] = []byte(line)
		x := strings.Index(line, "S")
		if x >= 0 {
			rv.Start = NewPoint(x, y)
			rv.Garden[y][x] = Garden
		}
	}
	return &rv, nil
}

func WatchGrow(solver *Solver) {
	odd := solver.MaxSteps % 2
	maxSteps := 2
	if odd == 1 {
		maxSteps = 1
	}
	spots := solver.GetSolutions(solver.MaxSteps)
	grid := solver.Garden.Replicate(3, 3)
	for y, row := range grid {
		for x, cell := range row {
			if cell == Garden && (x+y)%2 == odd {
				grid[y][x] = 'O'
			}
		}
	}
	hl := []XY{ShiftPoint(solver.Start, 0, 0)}
	counts := make([]int, 0, solver.MaxSteps/2+1)
	for ; maxSteps <= solver.MaxSteps; maxSteps += 2 {
		var col []XY
		var shown, notShown int
		for _, spot := range spots {
			if spot.Cost > maxSteps {
				continue
			}
			p := ShiftPoint(spot, 0, 0)
			if grid.Has(p) {
				col = append(col, p)
				shown++
			} else {
				notShown++
			}
		}

		tot := shown + notShown
		counts = append(counts, tot)
		if maxSteps <= 20 {
			Stdoutf("Max Steps: %d, Total: %d, Shown: %d, Not Shown: %d\n%s",
				maxSteps, tot, shown, notShown, CreateIndexedGridStringBz(grid, col, hl))
		} else {
			Stdoutf("Max Steps: %d, Total: %d, Shown: %d, Not Shown: %d",
				maxSteps, tot, shown, notShown)
		}
	}
	allSame := make(map[int]map[int][]int)
	Stdoutf("   Counts: %v", counts)
	showDiffs := len(counts) <= 50
	for e := 0; e < len(counts); e++ {
		diffs := DiffEvery(counts, e)
		if len(diffs) <= 2 {
			break
		}
		if showDiffs {
			Stdoutf("Diff e%dd%d: %v", e, 0, diffs)
		}
		if AreAllSame(diffs) {
			if showDiffs {
				Stdoutf("%s", strings.Repeat("^", 100))
			} else {
				Stdoutf("Diff e%dd%d: %v", e, 0, diffs)
			}
			if allSame[e] == nil {
				allSame[e] = make(map[int][]int)
			}
			allSame[e][0] = diffs
		}
		dl := len(diffs)
		for d := 1; d < dl; d++ {
			diffs = GetDiffs(diffs)
			if len(diffs) <= 2 {
				break
			}
			if showDiffs {
				Stdoutf("Diff e%dd%d: %s%v", e, d, strings.Repeat(" ", d), diffs)
			}
			if AreAllSame(diffs) {
				if showDiffs {
					Stdoutf("%s", strings.Repeat("^", 100))
				} else {
					Stdoutf("Diff e%dd%d: %s%v", e, d, strings.Repeat(" ", d), diffs)
				}
				if allSame[e] == nil {
					allSame[e] = make(map[int][]int)
				}
				allSame[e][d] = diffs
			}
		}
	}
	for e := range allSame {
		for d := range allSame[e] {
			Stdoutf("All Same: e: %d, d: %d", e, d)
		}
	}
}

func AreAllSame(vals []int) bool {
	if len(vals) <= 1 {
		return true
	}
	for i := 1; i < len(vals); i++ {
		if vals[0] != vals[i] {
			return false
		}
	}
	return true
}

func GetDiffs(vals []int) []int {
	return DiffEvery(vals, 1)
}

func DiffEvery(vals []int, every int) []int {
	if every == 0 {
		return vals
	}
	if len(vals) <= every {
		return nil
	}
	rv := make([]int, len(vals)-every)
	for i := 0; i < len(rv); i++ {
		rv[i] = vals[i+every] - vals[i]
	}
	return rv
}

func ShiftPoint(p XY, dx, dy int) *Point {
	return &Point{
		X: p.GetX() + dx,
		Y: p.GetY() + dy,
	}
}

func CountSpaces(garden Grid) {
	// The only spots possible are the ones where (x + y) %2 == maxSteps %2
	var plots, rocks, hits, hitPlots, hitRocks int
	for y, row := range garden {
		for x, cell := range row {
			isHit := (x+y)%2 == 0
			if isHit {
				hits++
			}
			switch cell {
			case Garden:
				plots++
				if isHit {
					hitPlots++
				}
			case Rocks:
				rocks++
				if isHit {
					hitRocks++
				}
			default:
				panic(fmt.Errorf("unknown cell %q at (%d, %d)", cell, x, y))
			}
		}
	}
	w, h := garden.GetWH()
	tot := w * h
	Stdoutf("The garden is %d x %d = %d spaces.", w, h, w*h)
	Stdoutf("It has %d garden plots, and %d rocks.", plots, rocks)
	Stdoutf("With an even number of steps:")
	Stdoutf("Without rocks, there would be %d possible spots.", hits)
	Stdoutf("But there are %d rocks in the way, leaving only %d possible spots.", hitRocks, hitPlots)
	Stdoutf("With an odd number of steps:")
	Stdoutf("Without rocks, there would be %d possible spots.", tot-hits)
	Stdoutf("But there are %d rocks in the way, leaving only %d possible spots.", rocks-hitRocks, plots-hitPlots)
}

func (s Solver) PrintMinsSections(ms MinsSections) {
	ys := IntKeys(ms)
	rows := make([]string, len(ys))
	for i, y := range ys {
		xs := IntKeys(ms[y])
		sections := make([]string, len(xs))
		for j, x := range xs {
			grid := ms[y][x].ToMatrix(s.Width, s.Height)
			var hl []XY
			if x == 0 && y == 0 && s.Start != nil {
				hl = append(hl, s.Start)
			}
			var col []XY
			for my, row := range grid {
				for mx, cell := range row {
					if cell > 0 && cell%2 == s.MaxSteps%2 {
						col = append(col, Point{X: mx, Y: my})
					}
				}
			}
			gridStr := CreateIndexedGridStringMins(grid, col, hl)
			sections[j] = gridStr
		}
		rows[i] = JoinSections(sections)
	}
	Stdoutf("Mins:\n%s", strings.Join(rows, "\n"))
}

func CreateIndexedGridStringMins[S ~[]E, E XY](vals [][]int, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			if val > 0 {
				strs[y][x] = fmt.Sprintf("%d", val)
			}
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

func JoinSections(sections []string) string {
	sectionLines := make([][]string, len(sections))
	for i, section := range sections {
		sectionLines[i] = strings.Split(section, "\n")
	}
	lines := make([]string, len(sectionLines[0]))
	for i := range lines {
		parts := make([]string, len(sectionLines))
		for s := range sectionLines {
			parts[s] = sectionLines[s][i]
		}
		lines[i] = strings.Join(parts, "  ")
	}
	return strings.Join(lines, "\n")
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

// StringNumberJoin maps the slice to strings, numbers them, and joins them.
func StringNumberJoin[S ~[]E, E Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
}

// StringNumberJoinFunc maps the slice to strings using the provided stringer, numbers them, and joins them.
func StringNumberJoinFunc[S ~[]E, E any](slice S, stringer func(E) string, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, stringer), startAt), sep)
}

// SliceToStrings runs String() on each entry of the provided slice.
func SliceToStrings[S ~[]E, E Stringer](slice S) []string {
	return MapSlice(slice, E.String)
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
func (p Params) String() string {
	defer FuncEnding(FuncStarting())
	nameFmt := "%10s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", p.Verbose),
		fmt.Sprintf(nameFmt+"%d", "Errors", len(p.Errors)),
		fmt.Sprintf(nameFmt+"%d", "Count", p.Count),
		fmt.Sprintf(nameFmt+"%s", "Input File", p.InputFile),
		fmt.Sprintf(nameFmt+"%d lines", "Input", len(p.Input)),
		fmt.Sprintf(nameFmt+"%d lines", "Custom", len(p.Custom)),
	}
	if len(p.Errors) > 0 {
		lines = append(lines, fmt.Sprintf("Errors (%d):", len(p.Errors)))
		errors := make([]string, len(p.Errors))
		for i, err := range p.Errors {
			errors[i] = err.Error()
		}
		lines = append(lines, AddLineNumbers(errors, 1)...)
	}
	if len(p.Input) > 0 {
		lines = append(lines, fmt.Sprintf("Input (%d):", len(p.Input)))
		lines = append(lines, AddLineNumbers(p.Input, 0)...)
	}
	if len(p.Custom) > 0 {
		lines = append(lines, fmt.Sprintf("Custom Input (%d):", len(p.Custom)))
		lines = append(lines, AddLineNumbers(p.Custom, 0)...)
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
func (p *Params) AppendError(err error) {
	if err != nil {
		p.Errors = append(p.Errors, err)
	}
}

// HasError returns true if this Params has one or more errors.
func (p Params) HasError() bool {
	return len(p.Errors) != 0
}

// Error flattens the Errors slice into a single string.
// It also makes the Params struct satisfy the error interface.
func (p Params) GetError() error {
	switch len(p.Errors) {
	case 0:
		return nil
	case 1:
		return p.Errors[0]
	default:
		errs := make([]error, 1, 1+len(p.Errors))
		errs[0] = fmt.Errorf("Found %d errors:", len(p.Errors)) //nolint:stylecheck,revive // punct okay here.
		for i, err := range p.Errors {
			errs = append(errs, fmt.Errorf("  %d: %w", i+1, err))
		}
		return errors.Join(errs...)
	}
}

func (p Params) HasCustom(val string) bool {
	return slices.Contains(p.Custom, val)
}

func (p Params) Verbosef(format string, a ...interface{}) {
	if p.Verbose {
		Stderrf(format, a...)
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
