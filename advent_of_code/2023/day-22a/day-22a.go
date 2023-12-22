package main

import (
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
	bricks, space := Drop(input.Bricks)
	Debugf("After Drops:\n%s", bricks)
	answer := CountDeletable(bricks, space)
	return fmt.Sprintf("%d", answer), nil
}

func CountDeletable(bricks Bricks, space SpaceMap[*Brick]) int {
	relations := BuildBrickRelations(bricks, space)
	rv := 0
	for _, brick := range relations {
		if brick.IsDeletable() {
			rv++
		}
	}
	return rv
}

func BuildBrickRelations(bricks Bricks, space SpaceMap[*Brick]) []*BrickRelation {
	rv := MapSlice(bricks, NewBrickRelation)
	brickMap := make(map[string]*BrickRelation)
	keyer := func(brick *Brick) string {
		return brick.String()
	}
	for _, brick := range rv {
		brickMap[keyer(brick.Brick)] = brick
	}
	for _, brick := range rv {
		for _, above := range GetAbove(brick, space) {
			br, ok := brickMap[keyer(above)]
			if !ok {
				panic(fmt.Errorf("could not find above brick %s in map", above))
			}
			brick.Above = append(brick.Above, br)
		}
		for _, below := range GetBelow(brick, space) {
			br, ok := brickMap[keyer(below)]
			if !ok {
				panic(fmt.Errorf("could not find below brick %s in map", below))
			}
			brick.Below = append(brick.Below, br)
		}
	}
	return rv
}

type BrickRelation struct {
	Brick *Brick
	Above []*BrickRelation
	Below []*BrickRelation
}

func NewBrickRelation(brick *Brick) *BrickRelation {
	return &BrickRelation{Brick: brick}
}

func (b BrickRelation) String() string {
	return fmt.Sprintf("%s has %d above: %v and %d below: %v", b.Brick,
		len(b.Above), MapSlice(b.Above, (*BrickRelation).GetName),
		len(b.Below), MapSlice(b.Below, (*BrickRelation).GetName))
}

func (b BrickRelation) GetName() string {
	return b.Brick.GetName()
}

func (b BrickRelation) IsDeletable() bool {
	for _, above := range b.Above {
		if len(above.Below) <= 1 {
			return false
		}
	}
	return true
}

func GetBelow(brick *BrickRelation, space SpaceMap[*Brick]) Bricks {
	var rv Bricks
	for _, block := range brick.Brick.Blocks {
		x, y, z := block.GetXYZ()
		below := space.GetRaw(x, y, z-1)
		if below != nil && below != brick.Brick && !rv.Contains(below) {
			rv = append(rv, below)
		}
	}
	return rv
}

func GetAbove(brick *BrickRelation, space SpaceMap[*Brick]) Bricks {
	var rv Bricks
	for _, block := range brick.Brick.Blocks {
		x, y, z := block.GetXYZ()
		above := space.GetRaw(x, y, z+1)
		if above != nil && above != brick.Brick && !rv.Contains(above) {
			rv = append(rv, above)
		}
	}
	return rv
}

func Drop(startingBricks Bricks) (Bricks, SpaceMap[*Brick]) {
	defer FuncEnding(FuncStarting())
	bricks := MapSlice(startingBricks, (*Brick).Copy)
	space := NewSpaceMap[*Brick]()

	setBrick := func(brick *Brick) {
		for _, block := range brick.Blocks {
			already := space.Get(block)
			if already != nil {
				panic(fmt.Errorf("cannot move put brick %s in place because it would overlap with %s", brick, already))
			}
			space.Set(block, brick)
		}
	}
	remBrick := func(brick *Brick) {
		for _, block := range brick.Blocks {
			space.Rem(block)
		}
	}
	dropBrick := func(brick *Brick) {
		Debugf("Trying to drop brick %s", brick)
		dist := -1
		canDrop := true
		for canDrop {
			dist++
			for _, block := range brick.Blocks {
				x, y, z := block.GetXYZ()
				z -= dist
				if z == 1 {
					Debugf("Brick %s has the ground %d below it.", brick, dist)
					canDrop = false
					break
				}
				below := space.GetRaw(x, y, z-1)
				if below != nil && below != brick {
					Debugf("Brick %s has brick %s %d below it.", brick, below, dist)
					canDrop = false
					break
				}
			}
		}
		if dist > 0 {
			Debugf("Dropping brick %s by %d", brick, dist)
			remBrick(brick)
			brick.Drop(dist)
			setBrick(brick)
		}
	}

	Debugf("Setting up space")
	for _, brick := range bricks {
		setBrick(brick)
	}

	_, _, zs := space.GetDims()
	if zs == nil {
		panic(errors.New("could not get space dimensions"))
	}
	Debugf("zs: %s", zs)

	zs.Iter(func(z int) {
		var toDrop Bricks
		for _, ym := range space[z] {
			for _, brick := range ym {
				toDrop = append(toDrop, brick)
			}
		}
		Debugf("z = %d, bricks (%d): %s", z, len(toDrop), toDrop)

		for _, brick := range toDrop {
			dropBrick(brick)
		}
	})

	return bricks, space
}

func IntKeys[M ~map[int]E, E any](m M) []int {
	rv := make([]int, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	slices.Sort(rv)
	return rv
}

type SpaceMap[E any] map[int]map[int]map[int]E

func NewSpaceMap[E any]() SpaceMap[E] {
	return make(SpaceMap[E])
}

type XYZ interface {
	GetX() int
	GetY() int
	GetZ() int
	GetXYZ() (int, int, int)
}

func (m SpaceMap[E]) Set(p XYZ, val E) {
	x, y, z := p.GetXYZ()
	m.SetRaw(x, y, z, val)
}

func (m SpaceMap[E]) SetRaw(x, y, z int, val E) {
	if m[z] == nil {
		m[z] = make(map[int]map[int]E)
	}
	if m[z][y] == nil {
		m[z][y] = make(map[int]E)
	}
	m[z][y][x] = val
}

func (m SpaceMap[E]) Rem(p XYZ) {
	x, y, z := p.GetXYZ()
	m.RemRaw(x, y, z)
}

func (m SpaceMap[E]) RemRaw(x, y, z int) {
	_, has := m.SafeGetRaw(x, y, z)
	if has {
		delete(m[z][y], x)
	}
	if len(m[z][y]) == 0 {
		delete(m[z], y)
	}
	if len(m[z]) == 0 {
		delete(m, z)
	}
}

func (m SpaceMap[E]) Get(p XYZ) E {
	x, y, z := p.GetXYZ()
	return m.GetRaw(x, y, z)
}

func (m SpaceMap[E]) GetRaw(x, y, z int) E {
	if m[z] == nil || m[z][y] == nil {
		var rv E
		return rv
	}
	return m[z][y][x]
}

func (m SpaceMap[E]) SafeGet(p XYZ) (E, bool) {
	x, y, z := p.GetXYZ()
	return m.SafeGetRaw(x, y, z)
}

func (m SpaceMap[E]) SafeGetRaw(x, y, z int) (E, bool) {
	if m[z] == nil || m[z][y] == nil {
		var rv E
		return rv, false
	}
	return m[z][y][x], true
}

func (m SpaceMap[E]) GetPointVals() []*PointVal[E] {
	var rv []*PointVal[E]
	for z, zm := range m {
		for y, ym := range zm {
			for x, v := range ym {
				rv = append(rv, NewPointVal(x, y, z, v))
			}
		}
	}
	return rv
}

func (m SpaceMap[E]) GetDims() (*MinMax, *MinMax, *MinMax) {
	rvx := InitMinMax()
	rvy := InitMinMax()
	rvz := InitMinMax()
	for z, zm := range m {
		rvz.Include(z)
		for y, ym := range zm {
			rvy.Include(y)
			for x := range ym {
				rvx.Include(x)
			}
		}
	}
	if !rvx.IsValid() {
		return nil, nil, nil
	}
	return rvx, rvy, rvz
}

type PointVal[E any] struct {
	Point
	Val E
}

func NewPointVal[E any](x, y, z int, val E) *PointVal[E] {
	return &PointVal[E]{
		Point: Point{
			X: x,
			Y: y,
			Z: z,
		},
		Val: val,
	}
}

func WithVal[E any](p XYZ, val E) *PointVal[E] {
	return &PointVal[E]{
		Point: Point{
			X: p.GetX(),
			Y: p.GetY(),
			Z: p.GetZ(),
		},
		Val: val,
	}
}

func (p PointVal[E]) GetVal() E {
	return p.Val
}

type Point struct {
	X int
	Y int
	Z int
}

func NewPoint(x, y, z int) *Point {
	return &Point{X: x, Y: y, Z: z}
}

func (p Point) Copy() *Point {
	return &Point{X: p.X, Y: p.Y, Z: p.Z}
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d,%d)", p.X, p.Y, p.Z)
}

func (p Point) GetX() int {
	return p.X
}

func (p Point) GetY() int {
	return p.Y
}

func (p Point) GetZ() int {
	return p.Z
}

func (p Point) GetXY() (int, int) {
	return p.X, p.Y
}

func (p Point) GetXZ() (int, int) {
	return p.X, p.Z
}

func (p Point) GetYZ() (int, int) {
	return p.Y, p.Z
}

func (p Point) GetXYZ() (int, int, int) {
	return p.X, p.Y, p.Z
}

type Path []*Point

func (p Path) String() string {
	return fmt.Sprintf("[%s]", StringNumberJoin(p, 1, "; "))
}

type Brick struct {
	Name   byte
	P1     *Point
	P2     *Point
	Blocks Path
}

func NewBrick(name byte, p1, p2 *Point) *Brick {
	rv := &Brick{Name: name, P1: p1, P2: p2}
	xr := NewRanger(p1.X, p2.X)
	yr := NewRanger(p1.Y, p2.Y)
	zr := NewRanger(p1.Z, p2.Z)
	xr.Iter(func(x int) {
		yr.Iter(func(y int) {
			zr.Iter(func(z int) {
				rv.Blocks = append(rv.Blocks, NewPoint(x, y, z))
			})
		})
	})
	return rv
}

func (b Brick) Copy() *Brick {
	rv := &Brick{
		Name: b.Name,
		P1:   b.P1.Copy(),
		P2:   b.P2.Copy(),
	}
	if b.Blocks != nil {
		rv.Blocks = make([]*Point, len(b.Blocks))
		for i, p := range b.Blocks {
			rv.Blocks[i] = p.Copy()
		}
	}
	return rv
}

func ParseBrick(name byte, line string) (*Brick, error) {
	parts := strings.Split(line, "~")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse brick %q", line)
	}

	p1, err := SplitParseInts(parts[0], ",")
	if err != nil {
		return nil, fmt.Errorf("could not parse brick %q point %q: %w", line, parts[0], err)
	}
	if len(p1) != 3 {
		return nil, fmt.Errorf("brick %q point %q has %d ints %v: expected 3", line, parts[0], len(p1), p1)
	}

	p2, err := SplitParseInts(parts[1], ",")
	if err != nil {
		return nil, fmt.Errorf("could not parse brick %q point %q: %w", line, parts[1], err)
	}
	if len(p2) != 3 {
		return nil, fmt.Errorf("brick %q point %q has %d ints %v: expected 3", line, parts[0], len(p2), p2)
	}

	return NewBrick(name, NewPoint(p1[0], p1[1], p1[2]), NewPoint(p2[0], p2[1], p2[2])), nil
}

func (b Brick) String() string {
	return fmt.Sprintf("%c=%s~%s:%s", b.Name, b.P1, b.P2, b.Blocks)
}

func (b Brick) GetName() string {
	return string(b.Name)
}

func (b *Brick) Drop(dist int) {
	b.P1.Z -= dist
	b.P2.Z -= dist
	for _, block := range b.Blocks {
		block.Z -= dist
	}
}

type Bricks []*Brick

func (b Bricks) String() string {
	return StringNumberJoin(b, 1, "\n")
}

func (b Bricks) Contains(brick *Brick) bool {
	for _, known := range b {
		if known == brick {
			return true
		}
	}
	return false
}

// CompInts returns -1 when a < b, 0 when a == b, and 1 when a > b.
func CompInts(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func OrderInts(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func Abs(i int) int {
	if i < 0 {
		return -1 * i
	}
	return i
}

type MinMax struct {
	Min int
	Max int
}

func NewMinMax(a, b int) *MinMax {
	min, max := OrderInts(a, b)
	return &MinMax{Min: min, Max: max}
}

func InitMinMax() *MinMax {
	return &MinMax{Min: MAX_INT, Max: MIN_INT}
}

func (m MinMax) Copy() *MinMax {
	return &MinMax{Min: m.Min, Max: m.Max}
}

func (m MinMax) String() string {
	return fmt.Sprintf("[%d-%d]", m.Min, m.Max)
}

func (m MinMax) IsValid() bool {
	return m.Min <= m.Max
}

func (m MinMax) Count() int {
	return m.Max - m.Min + 1
}

func (m MinMax) Contains(n int) bool {
	return m.Min <= n && n <= m.Max
}

func (m MinMax) Iter(runner func(n int)) {
	for n := m.Min; n <= m.Max; n++ {
		runner(n)
	}
}

func (m MinMax) Values() []int {
	rv := make([]int, 0, m.Count())
	m.Iter(func(n int) {
		rv = append(rv, n)
	})
	return rv
}

func (m *MinMax) Include(n int) {
	if n < m.Min {
		m.Min = n
	}
	if n > m.Max {
		m.Max = n
	}
}

type Ranger struct {
	Start int
	End   int
}

func NewRanger(start, end int) *Ranger {
	return &Ranger{Start: start, End: end}
}

func (r Ranger) Copy() *Ranger {
	return &Ranger{Start: r.Start, End: r.End}
}

func (r Ranger) String() string {
	return fmt.Sprintf("[%d,%d]", r.Start, r.End)
}

func (r Ranger) Contains(n int) bool {
	min, max := OrderInts(r.Start, r.End)
	return min <= n && n <= max
}

func (r Ranger) Count() int {
	return Abs(r.End-r.Start) + 1
}

func (r Ranger) Iter(runner func(n int)) {
	d := CompInts(r.Start, r.End)
	switch d {
	case 0:
		runner(r.Start)
	case -1:
		for n := r.Start; n <= r.End; n++ {
			runner(n)
		}
	case 1:
		for n := r.Start; n >= r.End; n-- {
			runner(n)
		}
	default:
		panic(fmt.Errorf("could not determine iterator for %s with d = %d", r, d))
	}
}

func (r Ranger) Values() []int {
	rv := make([]int, 0, r.Count())
	r.Iter(func(n int) {
		rv = append(rv, n)
	})
	return rv
}

type Input struct {
	Bricks Bricks
}

func (i Input) String() string {
	return fmt.Sprintf("Bricks (%d):\n%s", len(i.Bricks), StringNumberJoin(i.Bricks, 1, "\n"))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	names := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rv := Input{Bricks: make(Bricks, len(lines))}
	var err error
	for i, line := range lines {
		rv.Bricks[i], err = ParseBrick(names[i%len(names)], line)
		if err != nil {
			return nil, err
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
