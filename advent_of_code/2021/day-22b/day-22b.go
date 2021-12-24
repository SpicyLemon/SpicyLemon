package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_COUNT = 500

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	cubes := input.Instructions.ToCubes()
	Debugf("Cubes (%d):\n%s", len(cubes), cubes)
	if len(params.Custom) > 0 && params.Custom[0] == "p1" {
		bounds := NewCube(MinMax{-50, 50}, MinMax{-50, 50}, MinMax{-50, 50}, CUBE_ON)
		Debugf("Filtering cubes to those with something in %s", bounds)
		filtered := CubeList{}
		for _, cube := range cubes {
			if bounds.GetIntersectionCube(cube) != nil {
				filtered = append(filtered, cube)
			} else {
				Debugf("Filtered: %s", cube)
			}
		}
		cubes = filtered
		Debugf("Filtered Cubes (%d):\n%s", len(cubes), cubes)
	}
	finalCubes := CubeList{}
	for _, cube := range cubes {
		for _, fcube := range finalCubes {
			icube := cube.GetIntersectionCube(fcube)
			if icube != nil {
				finalCubes = append(finalCubes, icube)
			}
		}
		if cube.Val == CUBE_ON {
			finalCubes = append(finalCubes, cube)
		}
	}
	Debugf("Final Cubes (%d)", len(finalCubes))
	Debugf("Final Cubes (%d):\n%s", len(finalCubes), finalCubes)
	answer := finalCubes.GetTotal()
	return fmt.Sprintf("%d", answer), nil
}

func (l InstructionList) ToCubes() CubeList {
	rv := make(CubeList, len(l))
	for i, instr := range l {
		val := CUBE_ON
		if !instr.Val {
			val = CUBE_OFF
		}
		rv[i] = NewCube(instr.X, instr.Y, instr.Z, val)
	}
	return rv
}

func (c *Cube) HasIntersection(d *Cube) bool {
	// Check if any corners of c are inside d.
	for _, p := range c.GetCorners() {
		if d.Contains(p) {
			Debugf("Point %s is in %s.", p, d)
			return true
		}
	}
	// If no corners of c are inside d, either all or none of d's corners are in c.
	if c.Contains(NewPoint(d.X.Min, d.Y.Min, d.Z.Min)) {
		Debugf("Min Point %s is in %s.", NewPoint(d.X.Min, d.Y.Min, d.Z.Min), c)
		return true
	}
	// lastly, check if any edges of c intersect sides of d.
	thatSides := d.GetSides()
	for _, e := range c.GetEdges() {
		for _, s := range thatSides {
			if e.Intersects(s) {
				Debugf("Edge %s intersects side %s", e, s)
				return true
			}
		}
	}
	// Nothing left to check.
	return false
}

func (c *Cube) GetIntersectionCube(d *Cube) *Cube {
	// OFF OFF   - don't care.
	// MASK MASK - don't care.
	// ON ON     - standard mask.
	// OFF ON    - standard mask.
	// ON OFF    - standard mask.
	// ON MASK   - undo mask (because the off will apply two negatives).
	// MASK ON   - undo mask (because the off will apply two negatives).
	// OFF MASK  - undo mask (because the off will apply two negatives).
	// MASK OFF  - undo mask (because the off will apply two negatives).
	val := CUBE_MASK
	switch {
	case BothAreOneOf(c.Val, d.Val, CUBE_OFF) || BothAreOneOf(c.Val, d.Val, CUBE_MASK):
		// OFF OFF, MASK MASK
		return nil
	case BothAreOneOf(c.Val, d.Val, CUBE_ON, CUBE_OFF):
		// ON ON, ON OFF, OFF ON. (OFF OFF is handled in the case above.
		// Do nothing (keep going)
	default:
		// ON MASK, MASK ON, OFF MASK, MASK ON
		val = CUBE_ON
	}
	rv := NewCube(
		MinMax{Max(c.X.Min, d.X.Min), Min(c.X.Max, d.X.Max)},
		MinMax{Max(c.Y.Min, d.Y.Min), Min(c.Y.Max, d.Y.Max)},
		MinMax{Max(c.Z.Min, d.Z.Min), Min(c.Z.Max, d.Z.Max)},
		val,
	)
	if rv.Volume() <= 0 {
		return nil
	}
	return rv
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

type CubeList []*Cube

func (l CubeList) String() string {
	lines := make([]string, len(l))
	for i, v := range l {
		lines[i] = v.String()
	}
	return strings.Join(AddLineNumbers(lines, 0), "\n")
}

func (l CubeList) GetTotal() int {
	rv := 0
	for _, instr := range l {
		rv += instr.GetTotal()
	}
	return rv
}

type CubeType int

const CUBE_ON CubeType = 1
const CUBE_OFF CubeType = 0
const CUBE_MASK CubeType = -1

func (t CubeType) String() string {
	switch t {
	case 1:
		return "ON"
	case 0:
		return "OFF"
	case -1:
		return "MASK"
	}
	return fmt.Sprintf("CubeType(%d)", t)
}

func IsOneOf(v CubeType, vals ...CubeType) bool {
	for _, val := range vals {
		if v == val {
			return true
		}
	}
	return false
}

func BothAreOneOf(v1, v2 CubeType, vals ...CubeType) bool {
	return IsOneOf(v1, vals...) && IsOneOf(v2, vals...)
}

type Cube struct {
	Width  int
	Height int
	Depth  int
	X      MinMax
	Y      MinMax
	Z      MinMax
	Val    CubeType
}

func NewCube(x, y, z MinMax, val CubeType) *Cube {
	rv := Cube{
		Width:  x.Max - x.Min + 1,
		Height: y.Max - y.Min + 1,
		Depth:  z.Max - z.Min + 1,
		X:      x,
		Y:      y,
		Z:      z,
		Val:    val,
	}
	if !rv.X.IsOK() {
		rv.Width = rv.X.Max - rv.X.Min - 1
	}
	if !rv.Y.IsOK() {
		rv.Height = rv.Y.Max - rv.Y.Min - 1
	}
	if !rv.Z.IsOK() {
		rv.Depth = rv.Z.Max - rv.Z.Min - 1
	}
	return &rv
}

func (c Cube) String() string {
	return fmt.Sprintf("x=%s (% 6d), y=%s (% 6d), z=%s (% 6d) = % 16d cells %4s = % 16d", c.X, c.Width, c.Y, c.Height, c.Z, c.Depth, c.Volume(), c.Val, c.GetTotal())
}

func (c Cube) Volume() int {
	rv := c.Width * c.Height * c.Depth
	if c.X.IsOK() && c.Y.IsOK() && c.Z.IsOK() {
		return rv
	}
	if rv < 0 {
		return rv
	}
	return -rv
}

func (c Cube) GetCorners() PointList {
	return PointList{
		NewPoint(c.X.Min, c.Y.Min, c.Z.Min),
		NewPoint(c.X.Min, c.Y.Min, c.Z.Max),
		NewPoint(c.X.Min, c.Y.Max, c.Z.Min),
		NewPoint(c.X.Max, c.Y.Min, c.Z.Min),
		NewPoint(c.X.Min, c.Y.Max, c.Z.Max),
		NewPoint(c.X.Max, c.Y.Min, c.Z.Max),
		NewPoint(c.X.Max, c.Y.Max, c.Z.Min),
		NewPoint(c.X.Max, c.Y.Max, c.Z.Max),
	}
}

func (c Cube) GetEdges() EdgeList {
	return EdgeList{
		NewEdgeX(c.X, c.Y.Min, c.Z.Min),
		NewEdgeX(c.X, c.Y.Min, c.Z.Max),
		NewEdgeX(c.X, c.Y.Max, c.Z.Min),
		NewEdgeX(c.X, c.Y.Max, c.Z.Max),
		NewEdgeY(c.X.Min, c.Y, c.Z.Min),
		NewEdgeY(c.X.Min, c.Y, c.Z.Max),
		NewEdgeY(c.X.Max, c.Y, c.Z.Min),
		NewEdgeY(c.X.Max, c.Y, c.Z.Max),
		NewEdgeZ(c.X.Min, c.Y.Min, c.Z),
		NewEdgeZ(c.X.Min, c.Y.Max, c.Z),
		NewEdgeZ(c.X.Max, c.Y.Min, c.Z),
		NewEdgeZ(c.X.Max, c.Y.Max, c.Z),
	}
}

func (c Cube) GetSides() SideList {
	return SideList{
		NewSideX(c.X.Min, c.Y, c.Z),
		NewSideX(c.X.Max, c.Y, c.Z),
		NewSideY(c.X, c.Y.Min, c.Z),
		NewSideY(c.X, c.Y.Max, c.Z),
		NewSideZ(c.X, c.Y, c.Z.Min),
		NewSideZ(c.X, c.Y, c.Z.Max),
	}
}

func (c Cube) Contains(p Point) bool {
	return c.X.Contains(p.X) && c.Y.Contains(p.Y) && c.Z.Contains(p.Z)
}

func (c Cube) GetTotal() int {
	return c.Volume() * int(c.Val)
}

type Point struct {
	X int
	Y int
	Z int
}

func (p Point) String() string {
	return fmt.Sprintf("(% 6d,% 6d,% 6d)", p.X, p.Y, p.Z)
}

func NewPoint(x, y, z int) Point {
	return Point{
		X: x,
		Y: y,
		Z: z,
	}
}

type PointList []Point

func (l PointList) String() string {
	lines := make([]string, len(l))
	for i, v := range l {
		lines[i] = v.String()
	}
	return strings.Join(AddLineNumbers(lines, 1), ", ")
}

type Edge struct {
	D1     int
	D2     int
	DV     MinMax
	DVAxis byte
}

func (e Edge) String() string {
	switch e.DVAxis {
	case 'x':
		return fmt.Sprintf("(%s,%d,%d)", e.DV, e.D1, e.D2)
	case 'y':
		return fmt.Sprintf("(%d,%s,%d)", e.D1, e.DV, e.D2)
	case 'z':
		return fmt.Sprintf("(%d,%d,%s)", e.D1, e.D2, e.DV)
	}
	return fmt.Sprintf("DVAxis: %c, DV: %s, D1: %d, D2: %d", e.DVAxis, e.DV, e.D1, e.D2)
}

func NewEdgeX(x MinMax, y, z int) Edge {
	return Edge{
		D1:     y,
		D2:     z,
		DV:     x,
		DVAxis: 'x',
	}
}

func NewEdgeY(x int, y MinMax, z int) Edge {
	return Edge{
		D1:     x,
		D2:     z,
		DV:     y,
		DVAxis: 'y',
	}
}

func NewEdgeZ(x, y int, z MinMax) Edge {
	return Edge{
		D1:     x,
		D2:     y,
		DV:     z,
		DVAxis: 'z',
	}
}

func (e Edge) Intersects(s Side) bool {
	if e.DVAxis != s.DCAxis {
		return false
	}
	return e.DV.Contains(s.DC) && s.D1.Contains(e.D1) && s.D2.Contains(e.D2)
}

func (e Edge) MinPoint() Point {
	switch e.DVAxis {
	case 'x':
		return NewPoint(e.DV.Min, e.D1, e.D2)
	case 'y':
		return NewPoint(e.D1, e.DV.Min, e.D2)
	case 'z':
		return NewPoint(e.D1, e.D2, e.DV.Min)
	default:
		panic(fmt.Sprintf("Unknown variable side: %c", e.DVAxis))
	}
}

type EdgeList []Edge

func (l EdgeList) String() string {
	lines := make([]string, len(l))
	for i, v := range l {
		lines[i] = v.String()
	}
	if len(l) == 12 {
		return strings.Join(AddLineNumbers(lines[:6], 1), ", ") + "\n" + strings.Join(AddLineNumbers(lines[6:12], 7), ", ")
	}
	return strings.Join(AddLineNumbers(lines, 1), ", ")
}

type Side struct {
	DC     int
	D1     MinMax
	D2     MinMax
	DCAxis byte
}

func (s Side) String() string {
	switch s.DCAxis {
	case 'x':
		return fmt.Sprintf("(%d,%s,%s)", s.DC, s.D1, s.D2)
	case 'y':
		return fmt.Sprintf("(%s,%d,%s)", s.D1, s.DC, s.D2)
	case 'z':
		return fmt.Sprintf("(%s,%s,%d)", s.D1, s.D2, s.DC)
	}
	return fmt.Sprintf("DCAxis: %c, DC: %s, D1: %d, D2: %d", s.DCAxis, s.DC, s.D1, s.D2)
}

func NewSideX(x int, y, z MinMax) Side {
	return Side{
		DC:     x,
		D1:     y,
		D2:     z,
		DCAxis: 'x',
	}
}

func NewSideY(x MinMax, y int, z MinMax) Side {
	return Side{
		DC:     y,
		D1:     x,
		D2:     z,
		DCAxis: 'y',
	}
}

func NewSideZ(x, y MinMax, z int) Side {
	return Side{
		DC:     z,
		D1:     x,
		D2:     y,
		DCAxis: 'z',
	}
}

func (s Side) Contains(p Point) bool {
	switch s.DCAxis {
	case 'x':
		return s.DC == p.X && s.D1.Contains(p.Y) && s.D2.Contains(p.Z)
	case 'y':
		return s.DC == p.Y && s.D1.Contains(p.X) && s.D2.Contains(p.Z)
	case 'z':
		return s.DC == p.Z && s.D1.Contains(p.X) && s.D2.Contains(p.Y)
	}
	panic(fmt.Sprintf("Unknown constant side: %c", s.DCAxis))
}

func (s Side) Intersects(e Edge) bool {
	return e.Intersects(s)
}

type SideList []Side

func (l SideList) String() string {
	lines := make([]string, len(l))
	for i, v := range l {
		lines[i] = v.String()
	}
	return strings.Join(AddLineNumbers(lines, 1), ", ")
}

type Instruction struct {
	Val bool
	X   MinMax
	Y   MinMax
	Z   MinMax
}

func (i Instruction) String() string {
	return fmt.Sprintf("%t x=%s, y=%s, z=%s", i.Val, i.X, i.Y, i.Z)
}

func ParseInstruction(str string) (*Instruction, error) {
	onOff := strings.Split(str, " ")
	if len(onOff) != 2 {
		return nil, fmt.Errorf("could not parse Instruction on/off: %q", str)
	}
	rv := Instruction{}
	rv.Val = onOff[0] == "on"
	xyz := strings.Split(onOff[1], ",")
	if len(xyz) != 3 {
		return nil, fmt.Errorf("could not parse Instruction xyz: %q", str)
	}
	var err error
	rv.X, err = ParseMinMax(xyz[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse Instruction range x: %q: %w", str, err)
	}
	rv.Y, err = ParseMinMax(xyz[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse Instruction range y: %q: %w", str, err)
	}
	rv.Z, err = ParseMinMax(xyz[2])
	if err != nil {
		return nil, fmt.Errorf("could not parse Instruction range z: %q: %w", str, err)
	}
	return &rv, nil
}

type InstructionList []*Instruction

func (l InstructionList) String() string {
	lines := make([]string, len(l))
	for i, v := range l {
		lines[i] = v.String()
	}
	return strings.Join(AddLineNumbers(lines, 1), "\n")
}

type MinMax struct {
	Min int
	Max int
}

func (m MinMax) String() string {
	return fmt.Sprintf("% 7d..% 7d", m.Min, m.Max)
}

func ParseMinMax(str string) (MinMax, error) {
	parts := strings.Split(str, "..")
	if len(parts) != 2 {
		return MinMax{}, fmt.Errorf("could not parse MinMax: %q", str)
	}
	var err error
	rv := MinMax{}
	rv.Min, err = strconv.Atoi(parts[0][2:])
	if err != nil {
		return rv, err
	}
	rv.Max, err = strconv.Atoi(parts[1])
	return rv, err
}

func (m MinMax) Contains(val int) bool {
	return m.Min <= val && val <= m.Max
}

func (m MinMax) IsOK() bool {
	return m.Min <= m.Max
}

func (m MinMax) GetRange() []int {
	rv := make([]int, m.Max-m.Min+1)
	for v := m.Min; v <= m.Max; v++ {
		rv[v-m.Min] = v
	}
	return rv
}

type Input struct {
	Instructions InstructionList
}

func (i Input) String() string {
	return fmt.Sprintf("Instructions (%d):\n%s", len(i.Instructions), i.Instructions)
}

func ParseInput(lines []string) (*Input, error) {
	rv := Input{}
	for i, line := range lines {
		if len(line) > 0 {
			inst, err := ParseInstruction(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i, err)
			}
			rv.Instructions = append(rv.Instructions, inst)
		}
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------
// -------------------------------  Some generic stuff  --------------------------------
// -------------------------------------------------------------------------------------

const MIN_INT8 = int8(-128)
const MAX_INT8 = int8(127)
const MIN_INT16 = int16(-32768)
const MAX_INT16 = int16(32767)
const MIN_INT32 = int32(-2147483648)
const MAX_INT32 = int32(2147483647)
const MIN_INT64 = int64(-9223372036854775808)
const MAX_INT64 = int64(9223372036854775807)
const MIN_INT = -9223372036854775808
const MAX_INT = 9223372036854775807

const MAX_UINT8 = uint8(255)
const MAX_UINT16 = uint16(65535)
const MAX_UINT32 = uint32(4294967295)
const MAX_UINT64 = uint64(18446744073709551615)
const MAX_UINT = uint(18446744073709551615)

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
// E.g. DigitFormatForMax(10) returns "%2d"
// DigitFormatForMax(382920) returns "%6d"
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

// -------------------------------------------------------------------------------------
// --------------------------  CLI params and input parsing  ---------------------------
// -------------------------------------------------------------------------------------

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

// String creates a multi-line string representing this Params
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

// DEFAULT_INPUT_FILE is the default input filename
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
			// Using fmt.Println here instead of my stdout function because the extra formatting is annoying with help text.
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
					Stderr("Debugging enabled by CLI arguments.")
				case oldDebug && !debug:
					Stderr("Debugging disabled by CLI arguments.")
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
		lines := []string{fmt.Sprintf("Found %d errors:", len(c.Errors))}
		for i, err := range c.Errors {
			lines = append(lines, fmt.Sprintf("  %d: %s", i, err.Error()))
		}
		return errors.New(strings.Join(lines, "\n"))
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
	if len(rv) >= 0 {
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
	DebugfAlways("Reading file: %s", filename)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		Stderr("error reading file: %v", err)
		return []string{}, err
	}
	return strings.Split(string(dat), "\n"), nil
}

// -------------------------------------------------------------------------------------
// --------------------------  Environment Variable Handling  --------------------------
// -------------------------------------------------------------------------------------

// HandleEnvVars looks at specific environment variables and sets global variables appropriately.
func HandleEnvVars() error {
	var err error
	debug, err = GetEnvVarBool("DEBUG")
	if debug {
		Stderr("Debugging enabled via environment variable.")
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

// -------------------------------------------------------------------------------------
// ------------------------  Function start/stop timing stuff  -------------------------
// -------------------------------------------------------------------------------------

// If all you want is starting/ending messages when debug is on, use:
//    defer FuncEnding(FuncStarting())
// If, when debug is on, you want starting/ending messages,
// but when debug is off, you still the function duration, then use:
//    defer FuncEndingAlways(FuncStarting())

// FuncStarting outputs that a function is starting (if debug is true).
// It returns the params needed by FuncEnding or FuncEndingAlways.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Minimal values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
//
// Example 1: In a function named "foo", you have this:
//     FuncStarting()
//   The printed message will note that "foo" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//     FuncStarting(3 * time.Second)
//   The printed message will note that "bar: 3s" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//     func sum(ints ...int) {
//         FuncStarting(ints...)
//     }
//     sum(1, 2, 3, 4, 20, 21, 22)
//   The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer FuncEnding(FuncStarting())
//             Or: defer FuncEndingAlways(FuncStarting())
func FuncStarting(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugfAs(name, "Starting.")
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
	DebugfAlwaysAs(name, "Starting.")
	return time.Now(), name
}

const done_fmt = "Done. Duration: [%s]."

// FuncEnding decrements the function depth and, if debug is on, outputs to stderr that how long a function took.
// Args will usually come from FuncStarting().
//
// This differs from FuncEndingAlways in that this only outputs something if debugging is turned on.
//
// Standard Usage: defer FuncEnding(FuncStarting())
func FuncEnding(start time.Time, name string) {
	DebugfAs(name, done_fmt, time.Since(start))
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
	DebugfAlwaysAs(name, done_fmt, time.Since(start))
	if funcDepth > -1 {
		funcDepth--
	}
}

// DurClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNNs", e.g. "12:01:02.000000000"
// - If less than one hour, format is "M:SS.NNNNNNNNNs",   e.g. "34:00.000000789"
// - If less than one minute, format is "S.NNNNNNNNNs",    e.g. "56.000456000"
// - If less than one second, format is "0.NNNNNNNNNs",    e.g. "0.123000000"
func DurClock(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	s := int(d.Seconds())
	n := int(d.Nanoseconds()) - 1000000000*s
	s = s - 60*m
	m = m - 60*h
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
// depth 0 = the function calling GetFuncName.
// depth 1 = the function calling the function calling GetFuncName.
// etc.
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

// -------------------------------------------------------------------------------------
// ---------------------------------  Output wrappers  ---------------------------------
// -------------------------------------------------------------------------------------

// GetOutputPrefix gets the prefix to add to all output.
func GetOutputPrefix(funcName string) string {
	tabs := ""
	if debug && funcDepth > 0 {
		tabs = strings.Repeat("  ", funcDepth)
	}
	return fmt.Sprintf("(%14s) %s[%s] ", DurClock(time.Since(startTime)), tabs, funcName)
}

// Stdout outputs to stdout with a prefixed run duration and automatic function name.
func Stdout(format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// Stderr outputs to stderr with a prefixed run duration and automatic function name.
func Stderr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// StdoutAs outputs to stdout with a prefixed run duration and provided function name.
func StdoutAs(funcName, format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(funcName)+format+"\n", a...)
}

// StderrAs outputs to stderr with a prefixed run duration and provided functio name.
func StderrAs(funcName, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(funcName)+format+"\n", a...)
}

// Debugf outputs to stderr if the debug flag is set.
func Debugf(format string, a ...interface{}) {
	if debug {
		StderrAs(GetFuncName(1), format, a...)
	}
}

// DebugfAs outputs to stderr if the debug flag is set.
func DebugfAs(funcName, format string, a ...interface{}) {
	if debug {
		StderrAs(funcName, format, a...)
	}
}

// DebugfAlways outputs to stderr if the debug flag is set, otherwise to stdout.
func DebugfAlways(format string, a ...interface{}) {
	if debug {
		StderrAs(GetFuncName(1), format, a...)
	} else {
		StdoutAs(GetFuncName(1), format, a...)
	}
}

// DebugfAlways outputs to stderr if the debug flag is set, otherwise to stdout.
func DebugfAlwaysAs(funcName, format string, a ...interface{}) {
	if debug {
		StderrAs(funcName, format, a...)
	} else {
		StdoutAs(funcName, format, a...)
	}
}

// -------------------------------------------------------------------------------------
// --------------------------  Primary Program Running Parts  --------------------------
// -------------------------------------------------------------------------------------

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
		// Not using Stderr(...) here because I don't want the time and function prefix on this.
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
	Stdout("Answer: %s", answer)
	return nil
}
