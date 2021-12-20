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

const POINTS_PER_LINE = 10

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStarting())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	switch {
	case debug:
		Stderr("Parsed Input:\n%s", input)
	case params.Verbose:
		Stderr("Scanners (%d):\n%s", len(input.Scanners), input.Scanners)
	}
	srots := make(ScannerRotationsList, len(input.Scanners))
	for i, scanner := range input.Scanners {
		srots[i] = NewScannerRotations(scanner)
	}
	if params.Verbose {
		Stdout("All Scanner Rotations (%d):\n%s", len(srots), srots)
	}
	found := make(ScannerLocationList, len(input.Scanners))
	found[0] = NewScannerLocation(0, input.Scanners[0].Points, NewPoint(0, 0, 0))
	newFound := 1
	for newFound > 0 && found.CountNotFound() > 0 {
		newFound = 0
		for nfi := 1; nfi < len(srots); nfi++ {
			if found[nfi] != nil {
				continue
			}
			for fi := 0; fi < len(found); fi++ {
				if found[fi] == nil {
					continue
				}
				for roti, nfirot := range srots[nfi].Rotations {
					Debugf("Comparing found Scanner %d to Scanner %d rotation %d:\n%s\n%s", fi, nfi, roti, found[fi].Points, nfirot)
					if m, ok := found[fi].Points.TranslatesTo(nfirot, params.Count); ok {
						found[nfi] = NewScannerLocation(nfi, nfirot, m.Shift.GetNegative())
						newFound++
						if params.Verbose {
							Stderr("Found match: Scanner %d + %s = Scanner %d. Key Points: %s, %s rot %d", fi, m.Shift, nfi, found[fi].Location, found[nfi].Location, roti)
						}
						break
					}
				}
				if found[nfi] != nil {
					break
				}
			}
		}
	}
	if params.Verbose {
		Stderr("Found (%d of %d):\n%s", len(found)-found.CountNotFound(), len(found), found)
	}
	if f := found.CountNotFound(); f != 0 {
		return "", fmt.Errorf("Could not locate %d scanners!!!", f)
	}
	var answer, s1, s2 int
	centers := make(PointList, len(found))
	for i, f := range found {
		centers[i] = f.Location
	}
	if params.Verbose {
		var msg strings.Builder
		switch {
		case len(centers) > POINTS_PER_LINE:
			msg.WriteByte('\n')
			msg.WriteString(centers.String())
		case len(centers) > 0:
			msg.WriteByte(' ')
			msg.WriteString(centers.String())
		}
		Stderr("Centers (%d):%s", len(centers), msg.String())
	}
	for i := 0; i < len(centers)-1; i++ {
		for j := 1 + 1; j < len(centers); j++ {
			md := GetManhattanDistance(centers[i], centers[j])
			if md > answer {
				answer = md
				s1 = i
				s2 = j
			}
		}
	}
	return fmt.Sprintf("%d from Scanner %d to Scanner %d", answer, s1, s2), nil
}

func GetManhattanDistance(p1 *Point, p2 *Point) int {
	return Abs(p1.X-p2.X) + Abs(p1.Y-p2.Y) + Abs(p1.Z-p2.Z)
}

func Abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ----------  PointListList  -------------------------------------------

type PointListList []PointList

func (l PointListList) String() string {
	lineFmt := DigitFormatForMax(len(l)) + ":"
	var rv strings.Builder
	lastI := len(l) - 1
	for i, pl := range l {
		rv.WriteString(PrefixLines(fmt.Sprintf(lineFmt, i+1), pl.String()))
		if i != lastI {
			rv.WriteByte('\n')
		}
	}
	return rv.String()
}

// ----------  ScannerRotations  ----------------------------------------

type ScannerRotations struct {
	ID        int
	Rotations PointListList
}

func (s ScannerRotations) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("Scanner ID: %d, Rotations: %d:", s.ID, len(s.Rotations)))
	if len(s.Rotations) > 0 {
		rv.WriteByte('\n')
		lineFmt := DigitFormatForMax(len(s.Rotations)) + ": "
		lastI := len(s.Rotations) - 1
		for i, rot := range s.Rotations {
			rv.WriteString(PrefixLines(fmt.Sprintf(lineFmt, i), rot.String()))
			if i != lastI {
				rv.WriteByte('\n')
			}
		}
	}
	return rv.String()
}

func NewScannerRotations(scanner *Scanner) *ScannerRotations {
	return &ScannerRotations{
		ID:        scanner.ID,
		Rotations: scanner.Points.GetAllRotations(),
	}
}

// ----------  ScannerRotationsList  ------------------------------------

type ScannerRotationsList []*ScannerRotations

func (l ScannerRotationsList) String() string {
	lineFmt := DigitFormatForMax(len(l)) + ": "
	var rv strings.Builder
	lastI := len(l) - 1
	for i, srot := range l {
		rv.WriteString(PrefixLines(fmt.Sprintf(lineFmt, srot.ID), srot.String()))
		if i != lastI {
			rv.WriteByte('\n')
		}
	}
	return rv.String()
}

// ----------  ScannerLocation  -----------------------------------------

type ScannerLocation struct {
	Scanner
	Location *Point
}

func NewScannerLocation(id int, points PointList, location *Point) *ScannerLocation {
	rv := ScannerLocation{
		Scanner: Scanner{
			ID:     id,
			Points: make(PointList, len(points)),
		},
		Location: location,
	}
	for i, p := range points {
		rv.Points[i] = p.Plus(location)
	}
	return &rv
}

func (s ScannerLocation) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("Scanner ID: %d at %s, Points (%2d):", s.ID, s.Location, len(s.Points)))
	if len(s.Points) > POINTS_PER_LINE-2 {
		rv.WriteByte('\n')
	} else {
		rv.WriteByte(' ')
	}
	rv.WriteString(s.Points.String())
	return rv.String()
}

// ----------  ScannerLocationList  -------------------------------------

type ScannerLocationList []*ScannerLocation

func (l ScannerLocationList) String() string {
	var rv strings.Builder
	lastI := len(l) - 1
	for i := range l {
		if l[i] == nil {
			rv.WriteString(fmt.Sprintf("Scanner ID %d: <unknown>", i))
		} else {
			rv.WriteString(l[i].String())
		}
		if i != lastI {
			rv.WriteByte('\n')
		}
	}
	return rv.String()
}

func (l ScannerLocationList) CountNotFound() int {
	rv := 0
	for i := range l {
		if l[i] == nil {
			rv++
		}
	}
	return rv
}

// ----------  Match  ---------------------------------------------------

// Match lists points that are equal after a shift. From + Shift = To.
type Match struct {
	From  PointList
	Shift *Point
	To    PointList
}

func NewMatch(from PointList, shift *Point) *Match {
	rv := Match{
		From:  from,
		Shift: shift,
	}
	rv.To = make(PointList, len(rv.From))
	for i, p := range rv.From {
		rv.To[i] = p.Plus(shift)
	}
	return &rv
}

func (m Match) AsReverse() *Match {
	return &Match{
		From:  m.From,
		Shift: m.Shift.GetNegative(),
		To:    m.To,
	}
}

// ----------  Point  ---------------------------------------------------

type Point struct {
	X int
	Y int
	Z int
}

func ParsePoint(str string) (*Point, error) {
	parts := strings.Split(str, ",")
	rv := Point{}
	var err error
	if len(parts) < 2 {
		return nil, fmt.Errorf("could not parse %q to Point: invalid format", str)
	}
	rv.X, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
	}
	rv.Y, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
	}
	if len(parts) > 2 {
		rv.Z, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
		}
	}
	return &rv, nil
}

func NewPoint(x, y, z int) *Point {
	return &Point{
		X: x,
		Y: y,
		Z: z,
	}
}

func (p Point) String() string {
	return fmt.Sprintf("(% 5d,% 5d,% 5d)", p.X, p.Y, p.Z)
}

func (p Point) Equals(pt *Point) bool {
	return p.X == pt.X && p.Y == pt.Y && p.Z == pt.Z
}

func (p Point) GetAllRotations() PointList {
	x, y, z := p.X, p.Y, p.Z
	pts := PointList{
		NewPoint(x, y, z),
		NewPoint(y, -x, z),
		NewPoint(-x, -y, z),
		NewPoint(-y, x, z),
		NewPoint(z, y, -x),
		NewPoint(-z, y, x),
	}
	rv := make(PointList, 0, 24)
	for _, o := range pts {
		rv = append(rv, o)
		rv = append(rv, NewPoint(o.X, -o.Z, o.Y))
		rv = append(rv, NewPoint(o.X, -o.Y, -o.Z))
		rv = append(rv, NewPoint(o.X, o.Z, -o.Y))
	}
	return rv
}

// GetShiftFrom gets the shift to go from the provided point to this point.
// I.e. o + shift = this. I.e. returns this - that
func (p Point) GetShiftFrom(o *Point) *Point {
	return NewPoint(p.X-o.X, p.Y-o.Y, p.Z-o.Z)
}

// GetShiftTo gets the shift to go from this point to the provided point.
// I.e. this + shift = o. I.e. returns that - this.
func (p Point) GetShiftTo(o *Point) *Point {
	return NewPoint(o.X-p.X, o.Y-p.Y, o.Z-p.Z)
}

func (p Point) Plus(o *Point) *Point {
	return NewPoint(p.X+o.X, p.Y+o.Y, p.Z+o.Z)
}

func (p Point) GetNegative() *Point {
	return NewPoint(-p.X, -p.Y, -p.Z)
}

func (p Point) Compare(o *Point) int {
	switch {
	case p.X < o.X:
		return -1
	case p.X > o.X:
		return 1
	case p.Y < o.Y:
		return -1
	case p.Y > o.Y:
		return 1
	case p.Z < o.Z:
		return -1
	case p.Z > o.Z:
		return 1
	default:
		return 0
	}
}

// ----------  PointList  -----------------------------------------------

type PointList []*Point

func (l PointList) String() string {
	leadFmt := "  " + DigitFormatForMax(len(l)) + ":"
	lastI := len(l) - 1
	var rv strings.Builder
	for i, p := range l {
		if i%POINTS_PER_LINE == 0 && len(l) > POINTS_PER_LINE {
			rv.WriteString(fmt.Sprintf(leadFmt, i))
		}
		rv.WriteByte(' ')
		rv.WriteString(p.String())
		if i != lastI && i%POINTS_PER_LINE == POINTS_PER_LINE-1 {
			rv.WriteByte('\n')
		}
	}
	return rv.String()
}

func (l PointList) Contains(pt *Point) bool {
	for _, p := range l {
		if p.Equals(pt) {
			return true
		}
	}
	return false
}

func AppendIfNewPL(l PointList, pts ...*Point) PointList {
	for _, pt := range pts {
		if !l.Contains(pt) {
			l = append(l, pt)
		}
	}
	return l
}

func (l PointList) GetAllRotations() []PointList {
	rv := make([]PointList, 24)
	for i := range rv {
		rv[i] = make(PointList, len(l))
	}
	for j, pj := range l {
		for i, pi := range pj.GetAllRotations() {
			rv[i][j] = pi
		}
	}
	return rv
}

func (l PointList) TranslatesTo(o PointList, count int) (*Match, bool) {
	defer FuncEnding(FuncStarting())
	for si, spti := range l {
		for sj, sptj := range o {
			shift := spti.GetShiftTo(sptj)
			pl := PointList{}
			for i, pti := range l {
				if len(l)-i+len(pl) < count {
					break
				}
				for _, ptj := range o {
					if pti.GetShiftTo(ptj).Equals(shift) {
						pl = append(pl, pti)
						break
					}
				}
			}
			Debugf("%2d%s - %2d%s = shift %s, matches: %d", si, sj, spti, sptj, shift, len(pl))
			if len(pl) >= count {
				return NewMatch(pl, shift), true
			}
		}
	}
	return nil, false
}

func (a PointList) Len() int           { return len(a) }
func (a PointList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PointList) Less(i, j int) bool { return a[i].Compare(a[j]) < 0 }

// ----------  Scanner --------------------------------------------------

type Scanner struct {
	ID     int
	Points PointList
}

func (s Scanner) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("Scanner ID: %d, Points (%2d):", s.ID, len(s.Points)))
	if len(s.Points) > POINTS_PER_LINE-1 {
		rv.WriteByte('\n')
	} else {
		rv.WriteByte(' ')
	}
	rv.WriteString(s.Points.String())
	return rv.String()
}

func ParseScanner(lines []string) (*Scanner, error) {
	if len(lines) == 0 {
		return nil, errors.New("cannot create scanner from empty line slice")
	}
	var err error
	rv := Scanner{}
	rv.ID, err = strconv.Atoi(strings.Trim(lines[0], "- scaner"))
	if err != nil {
		return nil, err
	}
	rv.Points = make(PointList, len(lines)-1)
	for i, line := range lines[1:] {
		rv.Points[i], err = ParsePoint(line)
		if err != nil {
			return nil, err
		}
	}
	return &rv, err
}

// ----------  ScannerList ----------------------------------------------

type ScannerList []*Scanner

func (l ScannerList) String() string {
	var rv strings.Builder
	lastI := len(l) - 1
	for i, s := range l {
		rv.WriteString(s.String())
		if i != lastI {
			rv.WriteByte('\n')
		}
	}
	return rv.String()
}

// ----------  Input  ---------------------------------------------------

type Input struct {
	Scanners ScannerList
}

func ParseInput(lines []string) (*Input, error) {
	rv := Input{}
	scannerLines := [][]string{}
	cur := -1
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "---"):
			cur++
			scannerLines = append(scannerLines, []string{line})
		case len(line) != 0:
			scannerLines[cur] = append(scannerLines[cur], line)
		}
	}
	var err error
	rv.Scanners = make([]*Scanner, len(scannerLines))
	for i, slines := range scannerLines {
		rv.Scanners[i], err = ParseScanner(slines)
		if err != nil {
			return nil, err
		}
	}
	for i, s := range rv.Scanners {
		if i != s.ID {
			return nil, fmt.Errorf("scanner at index %d has id %d", i, s.ID)
		}
	}
	return &rv, nil
}

func (i Input) String() string {
	return i.Scanners.String()
}

// -------------------------------------------------------------------------------------
// -------------------------------  Some generic stuff  --------------------------------
// -------------------------------------------------------------------------------------

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
	rv.Count = MIN_INT
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
	if rv.Count == MIN_INT {
		rv.Count = 12
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
	Stdout("Reading file: %s", filename)
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
	if debug {
		StderrAs(name, "Starting.")
	}
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
	if debug {
		StderrAs(name, done_fmt, time.Since(start))
	}
	if funcDepth > -1 {
		funcDepth--
	}
}

// FuncEndingAlways decrements the function depth and outputs how long a function took.
// If debug is on, output is to stderr, otherwise to stdout.
//
// This differs from FuncEnding in that this will always do the output (regardless of degub state).
//
// Usage: defer FuncEndingAlways(FuncStarting())
func FuncEndingAlways(start time.Time, name string) {
	if debug {
		StderrAs(name, done_fmt, time.Since(start))
	} else {
		StdoutAs(name, done_fmt, time.Since(start))
	}
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
