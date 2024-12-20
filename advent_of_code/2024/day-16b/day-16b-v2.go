package main

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const DEFAULT_COUNT = 0

// Flag specifics:
// --count <maxCost>   -> The default is based on the name of the input file to match what's known.
// --lines <point>     -> provide any number of points to output extra details about those points.
// --lines (x,y)-(x,y) -> Define a window of the grid to show after linking things up.
// --lines stop        -> Stop the program after wiring up the initial maze and outputting point details.

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)

	// Some extra params parsing.
	poi, windows, stopEarly, err := ParseCustom(params)
	_ = windows
	if err != nil {
		return "", err
	}
	maxCost := GetMaxCost(params)
	Debugf("Max Cost: %d", maxCost)

	// Create the wired-up maze.
	maze := CreateMaze(input.Maze)
	if verbose {
		PrintMaze(input, maze)
	}
	PrintPointsOfInterest(poi, maze)
	PrintWindowedMazes(windows, input, maze)
	if stopEarly {
		return "Stopping early due to custom option.", nil
	}

	// Find all the paths.
	paths := FindPaths(input.Start, input.End, maxCost, maze)
	PrintPathInfo(paths, input.Maze, maze)

	// Get all the deduplicated points on good paths.
	points := GetPointsOnPaths(paths)
	if verbose {
		Stderrf("Good spots (%d):\n%s", len(points), SolutionGridString(input.Maze, maze, points))
	}
	answer := len(points)
	return fmt.Sprintf("%d", answer), nil
}

func GetMaxCost(params *Params) int {
	if params != nil && params.Count > 0 {
		return params.Count
	}
	switch params.InputFile {
	case DEFAULT_INPUT_FILE:
		return 7036
	case "example2.input":
		return 11048
	}
	return 73432 // My answer from part 1.
}

type Window struct {
	Min *Point
	Max *Point
}

func (w *Window) String() string {
	if w == nil {
		return NilStr
	}
	return fmt.Sprintf("%s-%s", w.Min, w.Max)
}

var windowRx = regexp.MustCompile(`^\((\d+), ?(\d+)\)-\((\d+), ?(\d+)\)$`)

func ParseWindow(line string) (*Window, error) {
	parts := windowRx.FindStringSubmatch(line)
	if len(parts) != 5 {
		return nil, nil
	}
	var err error
	coords := make([]int, 4)
	for i, part := range parts[1:] {
		coords[i], err = strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid %s number %q from window %q: %w", Ith(i), part, line, err)
		}
	}
	return &Window{
		Min: NewPoint(min(coords[0], coords[2]), min(coords[1], coords[3])),
		Max: NewPoint(max(coords[0], coords[2]), max(coords[1], coords[3])),
	}, nil
}

func (w *Window) Validate() (err error) {
	if w == nil {
		return nil
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: invalid window", err)
		}
	}()

	if w.Min == nil {
		return errors.New("min cannot be nil")
	}
	if w.Max == nil {
		return errors.New("max cannot be nil")
	}
	if w.Min.X < 0 {
		return fmt.Errorf("min X value %d cannot be negative", w.Min.X)
	}
	if w.Min.Y < 0 {
		return fmt.Errorf("min Y value %d cannot be negative", w.Min.Y)
	}
	if w.Min.X > w.Max.X {
		return fmt.Errorf("min X %d cannot be greater than max X %d", w.Min.X, w.Max.X)
	}
	if w.Min.Y > w.Max.Y {
		return fmt.Errorf("min Y %d cannot be greater than max Y %d", w.Min.Y, w.Max.Y)
	}
	return nil
}

func (w *Window) Contains(point XY) bool {
	if w == nil || point == nil || w.Min == nil || w.Max == nil {
		return false
	}
	x, y := point.GetXY()
	return w.Min.X <= x && w.Min.Y <= y && w.Max.X >= x && w.Max.Y >= y
}

func ApplyWindowToGrid[V any](window *Window, vals [][]V) ([][]V, error) {
	if window == nil {
		return nil, errors.New("window cannot be nil")
	}
	if err := window.Validate(); err != nil {
		return nil, err
	}

	if window.Min.Y > len(vals) {
		return nil, fmt.Errorf("window min Y %d is greater than the height %d of the vals provided", window.Max.Y, len(vals))
	}
	if window.Max.Y > len(vals) {
		return nil, fmt.Errorf("window max Y %d is greater than the height %d of the vals provided", window.Max.Y, len(vals))
	}

	width := 0
	for _, row := range vals {
		if len(row) > width {
			width = len(row)
		}
	}
	if window.Min.X > width {
		return nil, fmt.Errorf("window min X %d is greater than the width %d of the vals provided", window.Max.Y, width)
	}
	if window.Max.X > width {
		return nil, fmt.Errorf("window max X %d is greater than the width %d of the vals provided", window.Max.Y, width)
	}

	rv := make([][]V, window.Max.Y-window.Min.Y+1)
	for y := window.Min.Y; y <= window.Max.Y; y++ {
		cy := y - window.Min.Y
		rv[cy] = make([]V, window.Max.X-window.Min.X+1)
		for x := window.Min.X; x <= window.Max.X && x < len(vals[y]); x++ {
			rv[cy][x-window.Min.X] = vals[y][x]
		}
	}

	return rv, nil
}

func ApplyWindowToPoints[S ~[]E, E XY](window *Window, points S) ([]*Point, error) {
	if window == nil {
		return nil, errors.New("window cannot be nil")
	}
	if err := window.Validate(); err != nil {
		return nil, err
	}

	shifter := NewPoint(window.Min.X*-1, window.Min.Y*-1)
	var rv []*Point
	for _, point := range points {
		if window.Contains(point) {
			rv = append(rv, AddXYs(point, shifter))
		}
	}
	return rv, nil
}

// ParseCustom returns all the points and windows from the custom lines and whether the program should stop early.
func ParseCustom(params *Params) ([]*Point, []*Window, bool, error) {
	if params == nil || len(params.Custom) == 0 {
		return nil, nil, false, nil
	}

	var poi []*Point      //nolint:prealloc // There's no decent size to pre-allocate this to.
	var windows []*Window //nolint:prealloc // There's no decent size to pre-allocate this to.
	for i, line := range params.Custom {
		if line == "stop" {
			return poi, windows, true, nil
		}

		win, err := ParseWindow(line)
		if err != nil {
			return nil, nil, false, err
		}
		if win != nil {
			windows = append(windows, win)
			continue
		}

		point, err := ParsePoint(line)
		if err != nil {
			return nil, nil, false, fmt.Errorf("invalid point[%d] option: %w", i, err)
		}
		poi = append(poi, point)
	}
	return poi, windows, false, nil
}

func PrintPointsOfInterest(points []*Point, maze [][]*Node[Cell]) {
	for _, point := range points {
		node, ok := GetB(maze, point)
		if !ok {
			Stderrf("Provided point %s is not in the maze of size %d x %d.", point, len(maze), len(maze[0]))
			continue
		}
		if node == nil {
			Stderrf("Provided point %s is not an intersection.", point)
			continue
		}
		Stderrf("%s", NodeDetailsString(node))
	}
}

func PrintWindowedMazes(windows []*Window, input *Input, maze [][]*Node[Cell]) {
	// TODO: Write this: CreateWindowedIndexedGridString
	_, _, _ = windows, input, maze
}

func PrintPathInfo(paths []*Path, base [][]byte, maze [][]*Node[Cell]) {
	if !verbose && !debug {
		return
	}
	Stderrf("Found %d paths.", len(paths))
	if !debug || len(paths) >= 20 {
		Stderrf("Paths (%d):\n%s", len(paths), StringProgressJoin(paths))
		return
	}

	iFmt := DigitFormatForMax(len(paths))
	iMax := fmt.Sprintf(iFmt, len(paths))
	for i, path := range paths {
		turns := 0
		for _, dir := range path.Steps {
			if dir == Turn {
				turns++
			}
		}
		Stderrf("["+iFmt+"/%s]: (%d) = %d turns and %d steps:\n%s\n%s", i+1, iMax, path.Cost, turns, len(path.Points)-1,
			path.MuxedString(),
			SolutionGridString(base, maze, path.Points))
	}
}

func NodeDetailsString(node *Node[Cell]) string {
	// node.Point
	// node.Next  map[Direction]*Node[V]
	// node.Value.Path       map[Direction]*Path
	// node.Value.Next       map[Direction]*Node[Cell]
	// node.Value.Visited    bool
	// node.Value.Queued     bool
	// node.Value.Cost       int
	// node.Value.MinCostOff int
	// node.Value.PathsTo    []*Path
	// node.Value.NewPathsTo []*Path

	label := fmt.Sprintf("(%3d,%3d)[%s%s]", node.Point.X, node.Point.Y,
		Ternary(node.Value.Queued, "Q", " "), Ternary(node.Value.Visited, "V", " "))
	labelLen := len(label)
	costs := fmt.Sprintf("Cost=%d, MinCostOff=%d", node.Value.Cost, node.Value.MinCostOff)

	allDirs := make([]Direction, len(Dirs))
	copy(allDirs, Dirs)
	allDirs = AppendDirIfNew(allDirs, DirKeys(node.Next)...)
	allDirs = AppendDirIfNew(allDirs, DirKeys(node.Value.Next)...)
	allDirs = AppendDirIfNew(allDirs, DirKeys(node.Value.Path)...)
	for _, dir := range allDirs {
		name := DirNames[dir]
		if len(name) == 0 {
			name = fmt.Sprintf("Dir('%c')", dir)
		}
		name = PadLeft(name, labelLen)
		// TODO: finish this and then call it using params.Custom stuff.
		_ = name
	}

	// TODO: Put it all together.
	_, _, _ = label, labelLen, costs
	return "TODO"
}

// TODO: Delete this once I make use of path.AsMoveString().
func ShortPathString(path *Path) string {
	return path.AsMoveString()
}

func AppendDirIfNew(dirs []Direction, newDirs ...Direction) []Direction {
	for _, dir := range newDirs {
		if !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func DirKeys[V any](m map[Direction]V) []Direction {
	rv := make([]Direction, 0, len(m))
	for key := range m {
		rv = append(rv, key)
	}
	return rv
}

func SolutionGridString(grid [][]byte, maze [][]*Node[Cell], points []*Point) string {
	base := CopyGrid(grid)

	var colors []*Point
	for y := range maze {
		for x := range maze[y] {
			if maze[y][x] != nil {
				colors = append(colors, NewPoint(x, y))
			}
		}
	}
	base = SetPoints(colors, '+', base)
	base = SetPoints(points, 'O', base)

	return CreateIndexedGridStringBz(base, colors, points)
}

func CopyGrid(grid [][]byte) [][]byte {
	rv := make([][]byte, len(grid))
	for y := range grid {
		rv[y] = make([]byte, len(grid[y]))
		copy(rv[y], grid[y])
	}
	return rv
}

func SetPoints[S ~[]E, E XY](points S, char byte, grid [][]byte) [][]byte {
	for _, p := range points {
		x, y := p.GetXY()
		if IsIn(grid, p) && grid[y][x] != Start && grid[y][x] != End {
			grid[y][x] = char
		}
	}
	return grid
}

var (
	BadChar = byte('~')

	ReallyBad     = "~x~X~~X~x~"
	WallStr       = "##########"
	UpStr         = "    ^^    "
	DownStr       = "    vv    "
	LeftStr       = " <===     "
	RightStr      = "     ===> "
	UpDown        = "    ||    "
	LeftRight     = " <======> "
	UpRight       = "    ^^==> "
	DownRight     = "    vv==> "
	UpLeft        = " <==^^    "
	DownLeft      = " <==vv    "
	UpDownRight   = "    ||==> "
	UpDownLeft    = " <==||    "
	UpLeftRight   = " <==^^==> "
	DownLeftRight = " <==vv==> "
	AllDir        = " <==||==> "

	DirCells = map[string]string{
		// "^":    UpStr,
		// "v":    DownStr,
		// "<":    LeftStr,
		// ">":    RightStr,
		"^v":   UpDown,
		"^<":   UpLeft,
		"^>":   UpRight,
		"v<":   DownLeft,
		"v>":   DownRight,
		"<>":   LeftRight,
		"^v<":  UpDownLeft,
		"^v>":  UpDownRight,
		"^<>":  UpLeftRight,
		"v<>":  DownLeftRight,
		"^v<>": AllDir,
	}
)

func CleanCell(cur string) (string, bool) {
	if len(cur) <= 10 || strings.Contains(cur, string(BadChar)) {
		return cur, true
	}

	var key string
	for _, dir := range Dirs {
		if strings.Contains(cur, string(dir)) {
			key += string(dir)
		}
	}

	rv, ok := DirCells[key]
	if !ok {
		return AsBadCell(cur), false
	}
	return rv, true
}

func AsBadCell(cur string) string {
	if len(cur) == 0 {
		return ReallyBad
	}
	m := Ternary(len(cur) < 9, 2, 3)
	var rv strings.Builder
	for i, b := range []byte(cur) {
		rv.WriteByte(Ternary(i%m == 1, BadChar, b))
	}
	return rv.String()
}

var muxWallRx = regexp.MustCompile(`[^^v<>0-9]`)

func PrintMaze(input *Input, maze [][]*Node[Cell]) {
	colors := make([]*Point, 0, 2)
	if input.Start != nil {
		colors = append(colors, input.Start)
	}
	if input.End != nil {
		colors = append(colors, input.End)
	}
	highs := CopyAppend(colors)
	cellMap := MapGrid(maze, CellNodeShortString)
	base := input.Maze // Just used for reference, so no need to copy it.
	for y := range base {
		for x, space := range base[y] {
			if space == Wall {
				cur := cellMap[y][x]
				if len(cur) == 0 {
					cellMap[y][x] = string(Wall)
				} else {
					highs = append(highs, NewPoint(x, y))
				}
			}
		}
	}
	Stderrf("Maze:\n%s", CreateIndexedGridString(cellMap, colors, highs))

	if !debug || len(input.Maze) > 20 {
		return
	}

	// Start over, wider, with just the walls.
	cellMap = MapGrid(base, MorphWall)

	// Add the intersection points.
	var intersections []*Node[Cell]
	for y := range maze {
		for x, node := range maze[y] {
			if node == nil {
				continue
			}

			intersections = append(intersections, node)
			cellVal := CellNodeDirString(node)
			if cellMap[y][x] == "" {
				cellMap[y][x] = cellVal
			} else {
				// Can only be a wall right now.
				cellMap[y][x] = muxWallRx.ReplaceAllString(cellVal, string(Wall))
			}
		}
	}

	// Now, set the cells next to the intersections.
	for _, node := range intersections {
		for dir, next := range node.Next {
			x, y := node.Point.Move(dir).GetXY()
			newVal := fmt.Sprintf("%c(%3d,%3d)", dir, next.Point.X, next.Point.Y)
			curVal := cellMap[y][x]
			if len(curVal) == 0 {
				cellMap[y][x] = newVal
				continue
			}
			if len(curVal) == 10 && strings.Contains(curVal, "(") {
				// Path cell next to at least one more intersection.
				// Append this dir to it and we'll clean that up once we've marked all the nexts.
				cellMap[y][x] += string(dir)
				continue
			}
			Stderrf("cannot combine cell into at (%d,%d): is %q, also %q", x, y, cellMap[y][x], newVal)
			cellMap[y][x] = ReallyBad
			highs = append(highs, NewPoint(x, y))
		}
	}

	// And lastly, combine the path cells that are next to two or more intersections.
	var ok bool
	for y := range cellMap {
		for x := range cellMap {
			cellMap[y][x], ok = CleanCell(cellMap[y][x])
			if !ok {
				highs = append(highs, NewPoint(x, y))
			}
		}
	}

	Stderrf("Maze Intersection Details:\n%s", CreateIndexedGridString(cellMap, colors, highs))
}

func MorphWall(b byte) string {
	if b == Wall {
		return strings.Repeat("#", 10)
	}
	return ""
}

func GetPointsOnPaths(paths []*Path) []*Point {
	var rv []*Point
	known := make(map[int]map[int]bool)
	for _, path := range paths {
		for _, point := range path.Points {
			if known[point.Y] != nil && known[point.Y][point.X] {
				continue
			}
			if known[point.Y] == nil {
				known[point.Y] = make(map[int]bool)
			}
			known[point.Y][point.X] = true
			rv = append(rv, point)
		}
	}
	return rv
}

func FindPaths(start, end *Point, maxCost int, maze [][]*Node[Cell]) []*Path {
	Verbosef("Finding all paths that cost %d or less.", maxCost)
	queue := make([]*Node[Cell], 0, len(maze)*2) // Will probably grow a bit, but at least I tried.
	enqueue := func(node *Node[Cell]) {
		switch {
		// case node.Value.Visited:
		//	Debugf("Not queuing already visited node: %s", node)
		case node.Value.Queued:
			Debugf("Already queued: %s", node)
		case IsSameXY(node, end):
			Debugf("Not queing end node: %s", node)
		default:
			Debugf("Adding to queue: %s", node)
			node.Value.Queued = true
			queue = append(queue, node)
		}
	}
	dequeue := func() *Node[Cell] {
		slices.SortFunc(queue, CompareCells)
		rv := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		rv.Value.Queued = false
		return rv
	}

	totalNodes := 0
	for y := range maze {
		for x := range maze[y] {
			if maze[y][x] != nil {
				totalNodes++
			}
		}
	}

	startNode := Get(maze, start)
	endNode := Get(maze, end)
	if startNode == nil || endNode == nil {
		Stderrf("Start: %s = %s", start, startNode)
		Stderrf("  End: %s = %s", end, endNode)
		Stderrf("Cannot proceed.")
		return nil
	}
	startNode.Value.PathsTo = []*Path{NewPath(start)}

	Verbosef("Start: %s = %s", start, startNode)
	Verbosef("  End: %s = %s", end, endNode)
	Verbosef("Count: %d", totalNodes)

	enqueue(Get(maze, start))
	checked := 0
	for len(queue) > 0 && checked < totalNodes*5 {
		checked++
		cur := dequeue()
		Verbosef("[%d/%d]: cur = %s", checked, len(queue), cur)

		var nextPaths []*Path
		if cur.Value.Visited {
			nextPaths = cur.Value.NewPathsTo
		} else {
			nextPaths = cur.Value.PathsTo
		}

		for i, path := range nextPaths {
			if path.Cost > maxCost {
				Debugf("[%d/%d]:[%d/%d]: Path already too long: %s", checked, len(queue), i+1, len(cur.Value.PathsTo), path)
				continue
			}
			Debugf("[%d/%d]:[%d/%d]: Extending %s", checked, len(queue), i+1, len(cur.Value.PathsTo), path)

			for dir, pathToNext := range cur.Value.Path {
				nextPath, err := CombinePaths(path, pathToNext)
				if err != nil {
					Debugf("[%d/%d]:[%d/%d]'%c': Cannot combine path to %s: %v",
						checked, len(queue), i+1, len(cur.Value.PathsTo), dir, cur.Value.Next[dir].Point, err)
					continue
				}
				Debugf("[%d/%d]:[%d/%d]'%c': Added path %s",
					checked, len(queue), i+1, len(cur.Value.PathsTo), dir, pathToNext)
				Debugf("[%d/%d]:[%d/%d]'%c': New path to: %s",
					checked, len(queue), i+1, len(cur.Value.PathsTo), dir, nextPath)

				nextNode := cur.Value.Next[dir]

				nextNode.Value.Cost = cur.Value.Cost + 1

				minCostOff := nextPath.Cost + 1
				if nextNode.Value.Next[dir] == nil {
					minCostOff += 1000
				}
				if nextNode.Value.MinCostOff == 0 || minCostOff < nextNode.Value.MinCostOff {
					nextNode.Value.MinCostOff = minCostOff
				}

				AppendPathIfNew(nextNode, nextPath)

				Debugf("[%d/%d]:[%d/%d]'%c': Updated next node: %s", checked, len(queue), i+1, len(cur.Value.PathsTo), dir, nextNode)
				enqueue(nextNode)
			}
		}

		cur.Value.Visited = true
		cur.Value.PathsTo = append(cur.Value.PathsTo, cur.Value.NewPathsTo...)
		cur.Value.NewPathsTo = nil
	}

	rv := make([]*Path, 0, len(endNode.Value.PathsTo)+len(endNode.Value.NewPathsTo))
	for _, path := range CopyAppend(endNode.Value.PathsTo, endNode.Value.NewPathsTo...) {
		if path.Cost <= maxCost {
			rv = append(rv, path)
		}
	}
	return rv
}

func AppendPathIfNew(node *Node[Cell], newPath *Path) {
	for _, path := range node.Value.PathsTo {
		if PathsAreEqual(newPath, path) {
			return
		}
	}
	for _, path := range node.Value.NewPathsTo {
		if PathsAreEqual(newPath, path) {
			return
		}
	}
	if node.Value.Visited {
		node.Value.NewPathsTo = append(node.Value.NewPathsTo, newPath)
	} else {
		node.Value.PathsTo = append(node.Value.PathsTo, newPath)
	}
	Debugf("New path added to %s: %s", node, newPath)
}

func CompareCells(a, b *Node[Cell]) int {
	if a == b {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	if rv := CmpInts(a.Value.MinCostOff, b.Value.MinCostOff); rv != 0 {
		return rv
	}
	if rv := CmpInts(a.Point.X, b.Point.X); rv != 0 {
		return rv
	}
	return CmpInts(a.Point.Y, b.Point.Y) * -1
}

func CmpInts(a, b int) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

func CreateMaze(grid [][]byte) [][]*Node[Cell] {
	rv := make([][]*Node[Cell], len(grid))
	var intersections []*Node[Cell]

	for y := range grid {
		rv[y] = make([]*Node[Cell], len(grid[y]))
		for x, c := range grid[y] {
			cur := NewPoint(x, y)
			nextDirs, keep := GetAdjacentOpenDirs(grid, cur)
			if !keep {
				continue
			}
			node := NewCellNode(x, y, c)
			for _, dir := range nextDirs {
				// Set them to nil for now and we'll create them later, once we have all the intersections identified.
				node.Value.Path[dir] = nil
				node.Value.Next[dir] = nil
			}
			node.Next = node.Value.Next
			rv[y][x] = node
			intersections = append(intersections, node)
		}
	}

	for _, node := range intersections {
		var rmDir []Direction
		for dir := range node.Value.Next {
			path, err := WalkPath(dir, grid, NewPath(node))
			if err != nil || rv == nil {
				rmDir = append(rmDir, dir)
				continue
			}
			nextPoint := path.GetLastPoint()
			if nextPoint == nil {
				Stderrf("Path:\n%s", path)
				panic(errors.New("empty path returned"))
			}
			next := Get(rv, nextPoint)
			if next == nil {
				Stderrf("Path:\n%s", path)
				panic(errors.New("path did not end at an intersection"))
			}
			node.Value.Path[dir] = path
			node.Value.Next[dir] = next
		}
		for _, key := range rmDir {
			delete(node.Value.Path, key)
			delete(node.Value.Next, key)
		}
		node.Next = node.Value.Next
	}

	Verbosef("There are %d intersections:\n%s", len(intersections), IntersectionsString(intersections))
	return rv
}

func WalkPath(dir Direction, grid [][]byte, rv *Path) (*Path, error) {
	err := rv.AddStep(dir)
	if err != nil {
		return nil, err
	}
	nexts, isStop := GetAdjacentOpenDirs(grid, rv.GetLastPoint())
	if isStop {
		return rv, nil
	}

	for _, next := range nexts {
		rv2, err2 := WalkPath(next, grid, rv)
		if err2 == nil {
			// Success
			return rv2, nil
		}
	}
	return nil, errors.New("dead end")
}

// GetAdjacentOpenDirs returns the list of possible directions and whether this is an intersection.
func GetAdjacentOpenDirs(grid [][]byte, point *Point) ([]Direction, bool) {
	c, ok := GetB(grid, point)
	if !ok || c == Wall {
		return nil, false
	}

	rv := make([]Direction, 0, 4)
	for dir, space := range GetAdjacent(grid, point) {
		if space != Wall {
			rv = append(rv, dir)
		}
	}

	return rv, len(rv) >= 3 || c == Start || c == End
}

func IntersectionsString(nodes []*Node[Cell]) string {
	parts := make([]string, len(nodes))
	for i, node := range nodes {
		dirs := ""
		for _, dir := range Dirs {
			if node.Value.Next[dir] != nil {
				dirs += string(dir)
			}
		}
		parts[i] = fmt.Sprintf("(%d,%d)%d%s", node.Point.X, node.Point.Y, len(node.Value.Next), dirs)
	}
	parts = ToEqualLengthStrings(parts)

	var lines []string
	nextLine := make([]string, 0, 10)
	for i, part := range parts {
		if i != 0 && i%10 == 0 {
			lines = append(lines, strings.Join(nextLine, "  "))
			nextLine = make([]string, 0, 10)
		}
		nextLine = append(nextLine, part)
	}
	if len(nextLine) != 0 {
		lines = append(lines, strings.Join(nextLine, "  "))
	}

	return strings.Join(lines, "\n")
}

func NewCellNode(x int, y int, space byte) *Node[Cell] {
	if space == Wall {
		return nil
	}
	return NewNode(x, y, *NewCell())
}

func CellNodeShortString(node *Node[Cell]) string {
	if node == nil {
		return ""
	}
	return strconv.Itoa(len(node.Value.Path))
}

func CellNodeDirString(node *Node[Cell]) string {
	if node == nil {
		return ""
	}
	dirs := make([]string, 4)
	for i, dir := range Dirs {
		dirs[i] = Ternary(node.Next[dir] != nil, string(dir), " ")
	}
	return "  [" + strings.Join(dirs, "") + "]  "
}

// A Cell is a type to put in a Node in order to process and find paths.
type Cell struct {
	Path       map[Direction]*Path
	Next       map[Direction]*Node[Cell]
	Visited    bool
	Queued     bool
	Cost       int
	MinCostOff int
	PathsTo    []*Path
	NewPathsTo []*Path
}

func NewCell() *Cell {
	return &Cell{
		Path: make(map[Direction]*Path),
		Next: make(map[Direction]*Node[Cell]),
	}
}

func (c Cell) String() string {
	parts := make([]string, len(Dirs))
	for i, dir := range Dirs {
		parts[i] = string(dir)
		path := c.Path[dir]
		next := c.Next[dir]
		if path == nil && next == nil {
			parts[i] += na
			continue
		}
		pathStr := "?<?>"
		if path != nil {
			pathStr = fmt.Sprintf("%d<%d>", len(path.Points)-1, path.Cost)
		}
		nextStr := "(?,?)"
		if next != nil {
			nextStr = next.Point.String()
		}
		parts[i] += pathStr + nextStr
	}
	return fmt.Sprintf("%d[%s%s]{%s}(%d)(%d)", c.MinCostOff, Ternary(c.Visited, "V", " "), Ternary(c.Queued, "Q", " "),
		strings.Join(parts, ";"), len(c.PathsTo), len(c.NewPathsTo))
}

// Path represents a series of steps between points.
// The first point is where the path start. Then take the first step to get to the send point.
// The last point is the ultimate destination.
// There *should* be one more points than steps.
// The Cost is how much it is to traverse the path starting at the first point facing in the same direction as the first step.
// I.e. If you have to turn to start down the path, that is NOT included in the path's cost (since it's outside the path).
type Path struct {
	Points    []*Point
	PointsMap map[int]map[int]*Point
	Steps     []Direction
	Cost      int
}

func NewPath(start XY) *Path {
	rv := &Path{PointsMap: make(map[int]map[int]*Point)}
	rv.addPoints(NewPoint(start.GetX(), start.GetY())) //nolint:errcheck // It's empty, so it can't give an error.
	return rv
}

func (p *Path) String() string {
	if p == nil {
		return NilStr
	}
	return fmt.Sprintf("(%d): %s", p.Cost, p.MuxedString())
}

func (p *Path) MuxedString() string {
	if p == nil {
		return NilStr
	}

	var rv strings.Builder
	pc, sc := len(p.Points), len(p.Steps) // "Point Count" and "Step Count"
	pi, si := 0, 0                        // "Point Index" and "Step Index"

	for pi < pc || si < sc {
		np := "(?,?)"
		ns := "?"
		if pi < pc {
			np = p.Points[pi].String()
		}
		if si < sc {
			ns = string(p.Steps[si])
			if p.Steps[si] == Turn {
				si++
				if si < sc {
					ns += string(p.Steps[si])
				}
			}
		}
		rv.WriteString(np)
		rv.WriteString(ns)
		pi++
		si++
	}
	str := rv.String()
	if str[len(str)-1] == '?' {
		return str[:len(str)-1]
	}
	return str
}

func (p *Path) Contains(p2 XY) bool {
	y := p2.GetY()
	return p != nil && p.PointsMap != nil && p.PointsMap[y] != nil && p.PointsMap[y][p2.GetX()] != nil
}

func (p *Path) GetFirstPoint() *Point {
	return GetFirst(p.Points, nil)
}

func (p *Path) GetLastPoint() *Point {
	return GetLast(p.Points, nil)
}

func (p *Path) GetFirstStep() Direction {
	return GetFirst(p.Steps, UnknownDir)
}

func (p *Path) GetLastStep() Direction {
	return GetLast(p.Steps, UnknownDir)
}

var (
	unknownPoint = "(?,?)"
	na           = "n/a"
)

func (p *Path) GetFirstMove() string {
	if p == nil {
		return na
	}

	var step string
	if len(p.Steps) > 0 {
		step = string(p.Steps[0])
	}
	if step == string(Turn) && len(p.Steps) > 1 {
		step += string(p.Steps[1])
	}

	switch {
	case len(p.Points) == 0:
		if len(step) == 0 {
			return "<empty>"
		}
		return unknownPoint + step + unknownPoint
	case len(p.Points) == 1:
		lp := p.Points[0].String()
		if len(step) > 0 {
			return lp + step + unknownPoint
		}
		return lp
	}

	if len(step) == 0 {
		step = "?"
	}
	return p.Points[0].String() + step + p.Points[1].String()
}

func (p *Path) GetLastMove() string {
	if p == nil {
		return na
	}

	var step string
	if len(p.Steps) > 1 && p.Steps[len(p.Steps)-2] == Turn {
		step = string(Turn)
	}
	if len(p.Steps) > 0 {
		step += string(p.Steps[len(p.Steps)-1])
	}

	switch {
	case len(p.Points) == 0:
		if len(step) == 0 {
			return "<empty>"
		}
		return unknownPoint + step + unknownPoint
	case len(p.Points) == 1:
		rp := p.Points[len(p.Points)-1].String()
		if len(step) > 0 {
			return unknownPoint + step + rp
		}
		return rp
	}

	if len(step) == 0 {
		step = "?"
	}
	return p.Points[len(p.Points)-2].String() + step + p.Points[len(p.Points)-1].String()
}

func (p *Path) AsMoveString() string {
	return fmt.Sprintf("%s...%s...%s", p.GetFirstMove(), p.GetCostString(), p.GetLastMove())
}

func (p *Path) GetCost() int {
	if p == nil {
		return MAX_INT
	}
	return p.Cost
}

func (p *Path) GetCostString() string {
	if p == nil {
		return NilStr
	}
	return strconv.Itoa(p.Cost)
}

func (p *Path) AddStep(nextDir Direction) error {
	lastDir := p.GetLastStep()
	switch {
	case lastDir == nextDir || lastDir == UnknownDir:
		p.Cost++
		p.Steps = append(p.Steps, nextDir)
	case IsTurn(lastDir, nextDir):
		p.Cost += 1001
		p.Steps = append(p.Steps, Turn, nextDir)
	default:
		return fmt.Errorf("unexpected next step %c following %c", nextDir, lastDir)
	}
	return p.addPoints(p.GetLastPoint().Move(nextDir))
}

// addPoints is not what you're looking for. It adds a point, but NOT the direction or cost for it. Use AddStep.
func (p *Path) addPoints(points ...*Point) error {
	// Don't change the path unless we know it's okay to do so.
	for _, point := range points {
		if p.PointsMap != nil && p.PointsMap[point.Y] != nil && p.PointsMap[point.Y][point.X] != nil {
			return fmt.Errorf("path already contains %s", point)
		}
	}

	p.Points = append(p.Points, points...)
	for _, point := range points {
		if p.PointsMap == nil {
			p.PointsMap = make(map[int]map[int]*Point)
		}
		if p.PointsMap[point.Y] == nil {
			p.PointsMap[point.Y] = make(map[int]*Point)
		}
		p.PointsMap[point.Y][point.X] = point
	}

	return nil
}

func (p *Path) Copy() *Path {
	if p == nil {
		return nil
	}
	rv := &Path{
		// Extra capacity in the expectation that we're about to add a point and step or two, and want to limit growth.
		Points:    make([]*Point, len(p.Points), len(p.Points)+1),
		PointsMap: make(map[int]map[int]*Point),
		Steps:     make([]Direction, len(p.Steps), len(p.Steps)+2),
		Cost:      p.Cost,
	}

	copy(rv.Points, p.Points)

	for y := range p.PointsMap {
		rv.PointsMap[y] = make(map[int]*Point)
		for x, v := range p.PointsMap[y] {
			rv.PointsMap[y][x] = v
		}
	}

	copy(rv.Steps, p.Steps)

	return rv
}

func (p *Path) CopyAddStep(nextDir Direction) (*Path, error) {
	rv := p.Copy()
	err := rv.AddStep(nextDir)
	return rv, err
}

func (p *Path) Append(next *Path) error {
	thisLastPoint := p.GetLastPoint()
	nextFirstPoint := next.GetFirstPoint()
	if !IsSameXY(thisLastPoint, nextFirstPoint) {
		return fmt.Errorf("this path's last point %s is not equal to the next path's "+
			"first point %s: cannot append path", thisLastPoint, nextFirstPoint)
	}

	thisLastStep := p.GetLastStep()
	nextFirstStep := next.GetFirstStep()
	isTurn := IsTurn(thisLastStep, nextFirstStep)
	if !isTurn && thisLastStep != nextFirstStep && thisLastStep != UnknownDir && nextFirstStep != UnknownDir {
		return fmt.Errorf("unexpected next step %c following %c: cannot append path", nextFirstStep, thisLastStep)
	}

	// if addPoints returns an error, it shouldn't have changed p at all.
	err := p.addPoints(next.Points[1:]...)
	if err != nil {
		return fmt.Errorf("%w: cannot append path", err)
	}
	if isTurn {
		p.Cost += 1000
		p.Steps = append(p.Steps, Turn)
	}
	p.Cost += next.Cost
	p.Steps = append(p.Steps, next.Steps...)
	return nil
}

func CombinePaths(path *Path, paths ...*Path) (*Path, error) {
	rv := path.Copy()
	for _, p := range paths {
		err := rv.Append(p)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func PathsAreEqual(a, b *Path) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a.Points) != len(b.Points) || len(a.PointsMap) != len(b.PointsMap) || len(a.Steps) != len(b.Steps) || a.Cost != b.Cost {
		return false
	}

	for i := range a.Points {
		if !IsSameXY(a.Points[i], b.Points[i]) {
			return false
		}
	}

	for y, aMap := range a.PointsMap {
		bMap := b.PointsMap[y]
		if aMap == nil && bMap == nil {
			continue
		}
		if aMap == nil || bMap == nil {
			return false
		}
		for x, aPoint := range aMap {
			bPoint := bMap[x]
			if aPoint == nil && bPoint == nil {
				continue
			}
			if aPoint == nil || bPoint == nil {
				return false
			}
			if !IsSameXY(aPoint, bPoint) {
				return false
			}
		}
	}

	for i := range a.Steps {
		if a.Steps[i] != b.Steps[i] {
			return false
		}
	}

	return true
}

// Move will return a new point 1 unit in the direction provided from this point.
func (p *Point) Move(dir Direction) *Point {
	if p == nil {
		return nil
	}
	d, ok := DDirs[dir]
	if !ok {
		panic(fmt.Errorf("cannot move %c from %s: unknown direction", dir, p))
	}
	return AddPoints(p, d)
}

// GetFirst will get the first element of a slice, or the provided default if it's empty.
func GetFirst[S ~[]E, E any](vals S, def E) E {
	if len(vals) == 0 {
		return def
	}
	return vals[0]
}

// GetLast will get the last element of a slice, or the provided default if it's empty.
func GetLast[S ~[]E, E any](vals S, def E) E {
	if len(vals) == 0 {
		return def
	}
	return vals[len(vals)-1]
}

// Equals returns true if the provided point equals this one.
func (p *Point) Equals(p2 *Point) bool {
	return IsSameXY(p, p2)
}

// MapGridXY creates a new map by running the provided mapper on the x,y values and each element of the provided grid.
func MapGridXY[G ~[][]E, E any, R any](grid G, mapper func(int, int, E) R) [][]R {
	if grid == nil {
		return nil
	}
	rv := make([][]R, len(grid))
	for y := range grid {
		rv[y] = make([]R, len(grid[y]))
		for x := range rv[y] {
			rv[y][x] = mapper(x, y, grid[y][x])
		}
	}
	return rv
}

// IsTurn returns true if cur and next are along different axes.
func IsTurn(cur, next Direction) bool {
	switch cur {
	case Up, Down:
		return next == Left || next == Right
	case Left, Right, UnknownDir:
		return next == Up || next == Down
	}
	return false
}

func ParsePoint(line string) (*Point, error) {
	str := line
	if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
		str = line[1 : len(line)-1]
	}

	parts := strings.Split(str, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unknown number of parts (%d) in %q: expected 2: cannot parse as point", len(parts), line)
	}

	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid x value %q in %q: %w", parts[0], line, err)
	}

	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid y value %q in %q: %w", parts[1], line, err)
	}

	return NewPoint(x, y), nil
}

// AddProgressNumbers adds progress numbers to each string in the format of "[i/max]: <string>" where i starts at 1.
func AddProgressNumbers(lines []string) []string {
	if len(lines) == 0 {
		return []string{}
	}
	iMax := len(lines)
	iFmt := DigitFormatForMax(iMax)
	lineFmt := "[" + iFmt + "/" + iFmt + "]: %s"
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf(lineFmt, i+1, iMax, line)
	}
	return rv
}

func StringProgressJoin[S ~[]E, E fmt.Stringer](vals S) string {
	return strings.Join(AddProgressNumbers(MapSlice(vals, E.String)), "\n")
}

func Ith(i int) string {
	sfx := "th"
	iStr := strconv.Itoa(i)
	switch {
	case strings.HasSuffix(iStr, "1") && !strings.HasSuffix(iStr, "11"):
		sfx = "st"
	case strings.HasSuffix(iStr, "2") && !strings.HasSuffix(iStr, "12"):
		sfx = "nd"
	case strings.HasSuffix(iStr, "3") && !strings.HasSuffix(iStr, "13"):
		sfx = "rd"
	}
	return iStr + sfx
}

const (
	Wall  = byte('#')
	Open  = byte('.')
	Start = byte('S')
	End   = byte('E')

	Turn       = Direction('T')
	UnknownDir = Direction('x')
)

type Input struct {
	Maze  [][]byte
	Start *Point
	End   *Point
}

func (i Input) String() string {
	// StringNumberJoin(slice, startAt, sep) string
	// StringNumberJoinFunc(slice, stringer, startAt, sep) string
	// SliceToStrings(slice) []string
	// AddLineNumbers(lines, startAt) []string
	// MapSlice(slice, mapper) slice  or  MapPSlice  or  MapSliceP
	// CreateIndexedGridString(grid, color, highlight) string  or  CreateIndexedGridStringBz  or  CreateIndexedGridStringNums
	// CreateIndexedGridStringFunc(grid, converter, color, highlight)
	var c, h []*Point
	if i.Start != nil {
		c = append(c, i.Start)
		h = append(h, i.Start)
	}
	if i.End != nil {
		h = append(h, i.End)
	}
	return fmt.Sprintf("Start: %s\nEnd: %s\n", i.Start, i.End) + CreateIndexedGridStringBz(i.Maze, c, h)
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Maze: make([][]byte, len(lines))}
	for i, line := range lines {
		rv.Maze[i] = []byte(line)
		for x, b := range rv.Maze[i] {
			switch b {
			case Start:
				rv.Start = NewPoint(x, i)
			case End:
				rv.End = NewPoint(x, i)
			}
		}
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------  Some generic stuff  --------------------------------------
// -------------------------------------------------------------------------------------------------

const (
	MIN_INT8  = int8(-128)
	MAX_INT8  = int8(127)
	MIN_INT16 = int16(-32_768)
	MAX_INT16 = int16(32_767)
	MIN_INT32 = int32(-2_147_483_648)
	MAX_INT32 = int32(2_147_483_647)
	MIN_INT64 = int64(-9_223_372_036_854_775_808)
	MAX_INT64 = int64(9_223_372_036_854_775_807)
	MIN_INT   = -9_223_372_036_854_775_808
	MAX_INT   = 9_223_372_036_854_775_807

	MAX_UINT8  = uint8(255)
	MAX_UINT16 = uint16(65_535)
	MAX_UINT32 = uint32(4_294_967_295)
	MAX_UINT64 = uint64(18_446_744_073_709_551_615)
	MAX_UINT   = uint(18_446_744_073_709_551_615)

	NilStr = "<nil>"
)

// SplitParseInts splits a string on whitespace and converts each part into an int.
// Uses strings.Fields(s) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseInts(s string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Fields(s) {
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

// SplitParseIntsD splits a string on the provided delimiter and converts each part into an int.
// Uses strings.Fields(s) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseIntsD(s, d string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Split(s, d) {
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

// StringJoin maps the slice to strings and joins them.
func StringJoin[S ~[]E, E fmt.Stringer](slice S, sep string) string {
	return strings.Join(MapSlice(slice, E.String), sep)
}

// StringNumberJoin maps the slice to strings, numbers them, and joins them.
func StringNumberJoin[S ~[]E, E fmt.Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
}

// StringNumberJoinFunc maps the slice to strings using the provided stringer, numbers them, and joins them.
func StringNumberJoinFunc[S ~[]E, E any](slice S, stringer func(E) string, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, stringer), startAt), sep)
}

// SliceToStrings runs String() on each entry of the provided slice.
func SliceToStrings[S ~[]E, E fmt.Stringer](slice S) []string {
	return MapSlice(slice, E.String)
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

// MakeZeroGrid creates a 2-d matrix with the given width and height, each entry having the zero value for the type.
// Usage: grid := MakeZeroGrid[byte](10, 10)
func MakeZeroGrid[V any](width, height int) [][]V {
	rv := make([][]V, height)
	for y := range rv {
		rv[y] = make([]V, width)
	}
	return rv
}

// MakeGrid creates a new 2-d matrix with the given width and height, each entry having the provided value.
func MakeGrid[V any](width, height int, value V) [][]V {
	rv := make([][]V, height)
	for y := range rv {
		rv[y] = make([]V, width)
		for x := range rv[y] {
			rv[y][x] = value
		}
	}
	return rv
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

// Abs returns the absolute value of the provided number.
func Abs[V Number](v V) V {
	var zero V
	if v < zero {
		return zero - v
	}
	return v
}

// Ternary returns ifTrue if test == true, otherwise, returns ifFalse.
func Ternary[E any](test bool, ifTrue, ifFalse E) E {
	if test {
		return ifTrue
	}
	return ifFalse
}

// CopyAppend returns a copy of s with the other provided entries appended.
func CopyAppend[S ~[]E, E any](s S, es ...E) S {
	rv := make(S, len(s)+len(es))
	copy(rv, s)
	copy(rv[len(s):], es)
	return rv
}

// Alternates: ©®¬ÆæØøÞþ

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

// Float is a constraint of float types. Same as golang.org/x/exp/constraints.Float.
type Float interface {
	~float32 | ~float64
}

// Ordered is a constraint for types that can be compared using > etc. Same as golang.org/x/exp/constraints.Ordered.
type Ordered interface {
	Integer | Float | ~string
}

// Complex is a constraint of complex types. Same as golang.org/x/exp/constraints.Complex.
type Complex interface {
	~complex64 | ~complex128
}

// Number is a constraint of integers and floats.
type Number interface {
	Integer | Float
}

// -----------------------------------------------------------------------------
// -------------------------------  Coordinates  -------------------------------
// -----------------------------------------------------------------------------

// XY is something that has an X and Y value.
type XY interface {
	GetX() int
	GetY() int
	GetXY() (int, int)
}

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
// See also: AddXYs, SumXYs.
func AddPoints(points ...*Point) *Point {
	rv := NewPoint(0, 0)
	for _, p := range points {
		rv.X += p.X
		rv.Y += p.Y
	}
	return rv
}

// AddXYs returns a new point that is the sum of the provided points.
// See also: AddPoints, SumXYs.
func AddXYs(points ...XY) *Point {
	return SumXYs(points)
}

// SumXYs returns a new point that is the sum of the provided points.
// See also: AddXYs, AddPoints.
func SumXYs[S ~[]E, E XY](points S) *Point {
	rv := NewPoint(0, 0)
	for _, p := range points {
		rv.X += p.GetX()
		rv.Y += p.GetY()
	}
	return rv
}

// IsSameXY returns true if a and b have the same x and y.
func IsSameXY(a, b XY) bool {
	return a != nil && b != nil && a.GetX() == b.GetX() && a.GetY() == b.GetY()
}

// AsPoints converts a slice of something with an X and Y into a slice of points.
func AsPoints[S ~[]E, E XY](vals S) []*Point {
	rv := make([]*Point, len(vals))
	for i, val := range vals {
		rv[i] = NewPoint(val.GetX(), val.GetY())
	}
	return rv
}

// HasPoints returns true if there's a point in path with the same (x,y) as the point provided.
func HasPoint[S ~[]E, E XY](path S, point XY) bool {
	x, y := point.GetXY()
	for _, p := range path {
		if x == p.GetX() && y == p.GetY() {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// -----------------------------  Node Grid Stuff  -----------------------------
// -----------------------------------------------------------------------------

// Key stuff: Node, Direction,
// AsNodeGrid[V any](vals [][]V) [][]*Node[V]
// GroupByValue[V comparable](vals [][]*Node[V]) map[V][]*Node[V]
// CreateByValueMapString[V cmp.Ordered](vals map[V][]*Node[V]) string
// CreateEnhancedByValueMapString[K cmp.Ordered, V ~[]W, W XY, S ~[]E, E XY](vals map[K]V, colorPoints, highlightPoints S) string

// AsNodeGrid creates a grid of nodes with the provide values, all the nodes are linked up with their neighbors.
func AsNodeGrid[V any](vals [][]V) [][]*Node[V] {
	rv := make([][]*Node[V], len(vals))
	for y := range vals {
		rv[y] = make([]*Node[V], len(vals[y]))
		for x := range vals[y] {
			rv[y][x] = NewNode(x, y, vals[y][x])
		}
	}
	LinkNodes(rv)
	return rv
}

// LinkNodes creates all of the Next maps linking nodes to their neighbors.
func LinkNodes[V any](grid [][]*Node[V]) {
	for y := range grid {
		for x := range grid[y] {
			cur := grid[y][x]
			if cur != nil {
				cur.Next = GetAdjacent(grid, cur)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// ----------------------------------  Node  -----------------------------------
// -----------------------------------------------------------------------------

// Node[V] has an x,y position, value, and knows its neighbors in a 2d grid.
type Node[V any] struct {
	Point
	Value V
	Next  map[Direction]*Node[V]
}

// NewNode creates a new Node at the given point with the given value (and no neighbors).
func NewNode[V any](x, y int, value V) *Node[V] {
	return &Node[V]{Point: Point{X: x, Y: y}, Value: value, Next: make(map[Direction]*Node[V])}
}

// String gets a string of this node.
func (n *Node[V]) String() string {
	if debug {
		return n.FullString()
	}
	return n.ShortString()
}

// String gets a string of this node that contains the point and value.
func (n *Node[V]) ShortString() string {
	if n == nil {
		return NilStr
	}
	return fmt.Sprintf("%s=%s", n.Point, GenericValueString(n.Value))
}

// FullString converts this node into a string with the format "(<x>,<y>)=<value>:[<neighbor flags>]".
// If a node has all four neighbors, the <neighbor flags> will be "UDLR".
// Any neighbor directions the node does NOT have are replaced with a space in that string.
// E.g the node in the upper right corner of the grid only has neighbors to the right and down, so it's " D R".
func (n *Node[V]) FullString() string {
	if n == nil {
		return NilStr
	}
	dirs := Ternary(n.Next[Up] != nil, "U", " ") +
		Ternary(n.Next[Down] != nil, "D", " ") +
		Ternary(n.Next[Right] != nil, "R", " ") +
		Ternary(n.Next[Left] != nil, "L", " ")
	return fmt.Sprintf("%s=%s:[%s]", n.Point, GenericValueString(n.Value), dirs)
}

// PointString returns the "(<x>,<y>)" for this node.
func (n *Node[V]) PointString() string {
	if n == nil {
		return NilStr
	}
	return n.Point.String()
}

// GetValue is a nil-safe way to get this node's value.
func (n *Node[V]) GetValue() V {
	if n != nil {
		return n.Value
	}
	var rv V
	return rv
}

// GetUp is a nil-safe way to get the node up from this one.
func (n *Node[V]) GetUp() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Next[Up]
}

// GetDown is a nil-safe way to get the node down from this one.
func (n *Node[V]) GetDown() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Next[Down]
}

// GetLeft is a nil-safe way to get the node to the left of this one.
func (n *Node[V]) GetLeft() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Next[Left]
}

// GetRight is a nil-safe way to get the node to the right of this one.
func (n *Node[V]) GetRight() *Node[V] {
	if n == nil {
		return nil
	}
	return n.Next[Right]
}

// Go gets the node in the requested direction from this one.
func (n *Node[V]) Go(dir Direction) *Node[V] {
	if n == nil {
		return nil
	}
	return n.Next[dir]
}

// CanGo returns true if you can go the given direction from this node.
func (n *Node[V]) CanGo(dir Direction) bool {
	return n != nil && n.Next[dir] != nil
}

// Unlink will go through all next nodes and make them not point at this node, then remove all next entries in this node.
func (n *Node[V]) Unlink() {
	if n == nil {
		return
	}
	for dir, node := range n.Next {
		if node != nil {
			delete(node.Next, DirOpposites[dir])
		}
	}
	n.Next = make(map[Direction]*Node[V])
}

// -----------------------------------------------------------------------------
// -------------------------------  Directions  --------------------------------
// -----------------------------------------------------------------------------

// Direction is a typed byte used to indicate a direction of travel.
type Direction byte

const (
	Up    = Direction('^')
	Down  = Direction('v')
	Left  = Direction('<')
	Right = Direction('>')
)

var (
	// Dirs are all of the directions available.
	Dirs = []Direction{Up, Down, Left, Right}
	// DirNames are a map of direction to a string naming it.
	DirNames = map[Direction]string{
		Up:    "Up",
		Down:  "Down",
		Left:  "Left",
		Right: "Right",
	}
	// DirOpposites is a map of direction to the direction going the other way.
	DirOpposites = map[Direction]Direction{
		Up:    Down,
		Down:  Up,
		Left:  Right,
		Right: Left,
	}

	DUp    = NewPoint(0, -1)
	DDown  = NewPoint(0, 1)
	DLeft  = NewPoint(-1, 0)
	DRight = NewPoint(1, 0)

	// DDirs is a map of direction to a Point that, when added to another Point, will move in that direction.
	DDirs = map[Direction]*Point{
		Up:    DUp,
		Down:  DDown,
		Left:  DLeft,
		Right: DRight,
	}
)

func IsDir(dir Direction) bool {
	return slices.Contains(Dirs, dir)
}

// GetAdjacentPoints gets the points that are adjacent to the given one.
func GetAdjacentPoints(p *Point) map[Direction]*Point {
	rv := make(map[Direction]*Point)
	if p == nil {
		return rv
	}
	for _, dir := range Dirs {
		rv[dir] = AddPoints(p, DDirs[dir])
	}
	return rv
}

// GetAdjacent gets the elemnts adjacent to the provided point in the provided grid.
func GetAdjacent[V any](grid [][]V, p XY) map[Direction]V {
	rv := make(map[Direction]V)
	if !IsIn(grid, p) {
		return rv
	}
	for _, dir := range Dirs {
		if v, ok := GetB(grid, AddXYs(p, DDirs[dir])); ok {
			rv[dir] = v
		}
	}
	return rv
}

// IsIn returns true if the provided point exists in the provided grid.
func IsIn[E any](grid [][]E, p XY) bool {
	if p == nil {
		return false
	}
	x, y := p.GetXY()
	return y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y])
}

// Get will safely get the element of the grid at the provided point.
// If the point is outside the grid, the zero-value is returned.
func Get[E any](grid [][]E, p XY) E {
	rv, _ := GetB(grid, p)
	return rv
}

// GetB will safely get the element of the grid at the provided point and whether it is in the grid.
// If the point is outside the grid, the zero-value and false is returned.
func GetB[E any](grid [][]E, p XY) (E, bool) {
	if IsIn(grid, p) {
		return grid[p.GetY()][p.GetX()], true
	}
	var rv E
	return rv, false
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

// -----------------------------------------------------------------------------
// ---------------------------  New Generic Helpers  ---------------------------
// -----------------------------------------------------------------------------

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

// GenericValueString returns a string representation of the provided value that's a little better than just fmt.Sprintf("%v", value).
// Specifically, byte and rune types are converted to their character instead of just their number value.
func GenericValueString[T any](value T) string {
	switch v := any(value).(type) {
	case string:
		return v
	case byte, rune:
		return fmt.Sprintf("%c", v)
	case fmt.Stringer:
		return v.String()
	}
	return fmt.Sprintf("%v", value)
}

// -----------------------------------------------------------------------------
// ------------------------------  String Makers  ------------------------------
// -----------------------------------------------------------------------------

// CreateIndexedGridStringBz is for [][]byte
// CreateIndexedGridStringNums is for [][]int or [][]uint or [][]int16 etc.
// CreateIndexedGridString is for [][]string
// All of them have the signature (vals, color, highlight)
// CreateIndexedGridStringFunc is for any other [][]; signature = (vals, converter, color, highlight)
// PointsString is simpler than PathString.
// EnhancedPathString is a similar signature as CreateIndexedGridString.

// PointString returns the "(%d,%d)" string for the provided XY.
// The generic here might seem silly, but without it, its use in PointsString gives a syntax error.
func PointString[V XY](p V) string {
	return fmt.Sprintf("(%d,%d)", p.GetX(), p.GetY())
}

// PointsString creates a string of the provided points, e.g. "(0,0) (0,1) (1,1)".
func PointsString[S ~[]E, E XY](points S) string {
	return strings.Join(MapSlice(points, PointString), " ")
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
	return CreateIndexedGridStringAtOffset(vals, colorPoints, highlightPoints, nil)
}

// CreateWindowedIndexedGridString will creagte a string of a window of the provided vals.
func CreateWindowedIndexedGridString[S ~[]E, E XY](vals [][]string, colorPoints S, highlightPoints S, window *Window) string {
	subVals, err := ApplyWindowToGrid(window, vals)
	if err != nil {
		panic(err.Error())
	}
	colors, err := ApplyWindowToPoints(window, colorPoints)
	if err != nil {
		panic(err.Error())
	}
	highs, err := ApplyWindowToPoints(window, highlightPoints)
	if err != nil {
		panic(err.Error())
	}
	return CreateIndexedGridStringAtOffset(subVals, colors, highs, window.Min)
}

// CreateIndexedGridStringAtOffset creates a string of the provided strings matrix.
// The result will have row and column indexes shifted by the provided offset and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringAtOffset[S ~[]E, E XY](vals [][]string, colorPoints S, highlightPoints S, offset *Point) string {
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

	x0, y0 := 0, 0
	if offset != nil {
		x0, y0 = offset.GetXY()
	}

	// Define the format that each line will start with and for each cell.
	leadFmt := fmt.Sprintf("%%%dd:", len(fmt.Sprintf("%d", height+y0)))
	blankLead := strings.Repeat(" ", len(fmt.Sprintf(leadFmt, 0)))
	cellFmt := fmt.Sprintf("%%%ds", cellLen)

	// If none of the rows have anything, just print out the row numbers.
	if width == 0 {
		lines := make([]string, len(vals))
		for y := range vals {
			lines[y] = fmt.Sprintf(leadFmt, y+y0)
		}
		return strings.Join(lines, "\n")
	}

	// Create the index numbers across the top.
	dCount := len(fmt.Sprintf("%d", width-1+x0))
	dLen := (width + x0) * cellLen
	digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	topIndexLines := make([]string, dCount+1)
	topIndexLines[dCount] = strings.Repeat("-", width*cellLen)
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
		topIndexLines[dCount-l] = line[x0*cellLen : dLen]
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
		rv.WriteString(fmt.Sprintf(leadFmt, y+y0))
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

// -------------------------------------------------------------------------------------------------
// --------------------------------  CLI params and input parsing  ---------------------------------
// -------------------------------------------------------------------------------------------------

// Params contains anything that might be provided via command-line arguments.
type Params struct {
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
		fmt.Sprintf(nameFmt+"%t", "Verbose", verbose),
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
		case HasOneOfPrefixesFold(args[i], "--debug", "-vv"):
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
			verbose, extraI, err = ParseFlagBool(args[i:])
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
		verbose = debug
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

// HasCustom returns true if the provided string was given as a custom arg.
func (p Params) HasCustom(str string) bool {
	for _, cust := range p.Custom {
		if cust == str {
			return true
		}
	}
	return false
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
	if debug || verbose {
		StderrAsf(GetFuncName(1), format, a...)
	} else {
		StdoutAsf(GetFuncName(1), format, a...)
	}
}

// DebugAlwaysAsf is like StderrAsf if the debug flag is set; otherwise it's like StdoutAsf.
func DebugAlwaysAsf(funcName, format string, a ...interface{}) {
	if debug || verbose {
		StderrAsf(funcName, format, a...)
	} else {
		StdoutAsf(funcName, format, a...)
	}
}

// Verbosef outputs to Stderr if the verbose flag was provided. Does nothing otherwise.
func Verbosef(format string, a ...interface{}) {
	if verbose {
		StderrAsf(GetFuncName(1), format, a...)
	}
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Primary Program Running Parts  --------------------------------
// -------------------------------------------------------------------------------------------------

var (
	// Debug is a flag for whether or not debug messages should be displayed.
	debug bool
	// Verbose is a flag for whether or not verbose messages should be displayed.
	verbose bool
)

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
