package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// debug is a flag for whether or not debug messages should be displayed.
var debug bool

// startTime is the time when the program started.
var startTime time.Time

// funcDepth is a global counter keeping track of function depth by the starting/ending function functions.
var funcDepth int

// -------------------------------------------------------------------------------------
// ----------------------------  Solver specific functions  ----------------------------
// -------------------------------------------------------------------------------------

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(input Input) (string, error) {
	defer FuncEndingAlways(FuncStarting())
	if input.Verbose || input.DX != NOT_GIVEN || input.DY != NOT_GIVEN {
		dx := input.DX
		dy := input.DY
		if dx == NOT_GIVEN {
			dx = int(math.Ceil(math.Sqrt(2.0*float64(input.Target.MinX)+0.25) - 0.5))
		}
		if dy == NOT_GIVEN {
			dy = 0 - input.Target.MinY - 1
		}
		probe := NewProbe(dx, dy)
		Stderr("Initial Probe: %s", probe)
		path, isHit := probe.FireAt(input.Target)
		Stderr("Final Probe: %s", probe)
		Stderr("Path: %s", path)
		if isHit {
			Stdout("It's a hit!")
		} else {
			Stdout("It's a miss.")
		}
		if debug || input.Verbose {
			Stderr("Drawing:\n%s", Draw(path, input.Target))
		}
	}
	dxs := GetRangeDX(input.Target)
	dys := GetRangeDY(input.Target)
	hits := Points{}
	actualDXs := []int{}
	actualDYs := []int{}
	for _, dy := range dys {
		for _, dx := range dxs {
			probe := NewProbe(dx, dy)
			_, isHit := probe.FireAt(input.Target)
			if isHit {
				hits = append(hits, probe.Velocity)
				actualDXs = appendIfNew(actualDXs, dx)
				actualDYs = appendIfNew(actualDYs, dy)
			}
		}
	}
	answer := len(hits)
	if debug || input.Verbose {
		sort.Ints(actualDXs)
		sort.Ints(actualDYs)
		Stderr("Hits (%d): %s", answer, hits)
		Stderr("Possible DX values (%d): %v", len(dxs), dxs)
		Stderr("Actual DX values (%d): %v", len(actualDXs), actualDXs)
		Stderr("Possible DY values (%d): %v", len(dys), dys)
		Stderr("Actual DY values (%d): %v", len(actualDYs), actualDYs)
	}
	return fmt.Sprintf("%d", answer), nil
}

func appendIfNew(vals []int, val int) []int {
	for _, v := range vals {
		if val == v {
			return vals
		}
	}
	return append(vals, val)
}

func GetRangeDX(target *TargetArea) []int {
	// x(i) = if i <= 0 => 0
	//        if i >= dx => dx*(dx+1)/2
	//        else i*(2dx - i + 1)/2
	// The overall minimum valid dx is when dx*(dx+1)/2 = minx
	// Solving for dx, that's dx = sqrt(2*minx+0.25) - 0.5
	// Since we only care about integers, we need to ceiling that
	// Once i gets to that, there won't be any more unique entries.
	// The overall max valid dx is simply maxx.
	// Not all values between are valid, though.
	// But we can solve that function for dx and loop through the i's to get all we need.
	// dx = (i - 1 + x*2/i)/2
	// So for each i:
	//    mindxi = (i - 1 + minx*2/i)/2
	//    maxdxi = (i - 1 + maxx*2/i)/2
	fmin := float64(target.MinX)
	fmax := float64(target.MaxX)
	maxI := math.Ceil(math.Sqrt(2.0*fmin+0.25) - 0.5)
	rv := []int{}
	for i := float64(1); i < maxI; i++ {
		mindi := int(math.Ceil((i - 1 + fmin*2/i) / 2))
		maxdi := int(math.Floor((i - 1 + fmax*2/i) / 2))
		Debugf("i: %f, mindi: %d, maxdi: %d", i, mindi, maxdi)
		for val := mindi; val <= maxdi; val++ {
			rv = appendIfNew(rv, val)
		}
	}
	sort.Ints(rv)
	return rv
}

func GetRangeDY(target *TargetArea) []int {
	// y(i) = if i <= 0 => 0
	//        else i*(2dy - i + 1)/2
	// The last i that will give a valid dy is when one step goes from y = 0 to y = miny.
	// That happens when it goes to the highest possible point.
	// We know that from part 1 to be when dy = 0 - miny - 1.
	// For positive dy's, it reaches a max after dy steps, then hits 0 at another dy + 1.
	// And one more step takes it to the target area.
	// That's a total of 2dy+2 steps.
	// So maxI is 2*(0 - miny - 1) + 2 => -2 * miny
	// The min and max are then the same as from dx.
	fmin := float64(target.MinY)
	fmax := float64(target.MaxY)
	maxI := float64(-2 * target.MinY)
	rv := []int{}
	for i := float64(1); i <= maxI; i++ {
		mindi := int(math.Ceil((i - 1 + fmin*2/i) / 2))
		maxdi := int(math.Floor((i - 1 + fmax*2/i) / 2))
		Debugf("i: %f, mindi: %d, maxdi: %d", i, mindi, maxdi)
		for val := mindi; val <= maxdi; val++ {
			rv = appendIfNew(rv, val)
		}
	}
	sort.Ints(rv)
	return rv
}

func Draw(pts Points, target *TargetArea) string {
	var minX, maxX, minY, maxY int
	if target.MinX < minX {
		minX = target.MinX
	}
	if target.MaxX > maxX {
		maxX = target.MaxX
	}
	if target.MinY < minY {
		minY = target.MinY
	}
	if target.MaxY > maxY {
		maxY = target.MaxY
	}
	for _, pt := range pts {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	height := maxY - minY + 1
	width := maxX - minX + 1
	grid := make([][]string, height)
	for y := range grid {
		grid[y] = make([]string, width)
		for x := range grid[y] {
			grid[y][x] = "."
		}
	}
	grid[height+minY-1][0-minX] = "S"
	for y := target.MinY; y <= target.MaxY; y++ {
		for x := target.MinX; x <= target.MaxX; x++ {
			grid[height-y+minY-1][x-minX] = "T"
		}
	}
	xys := make([]XY, len(pts))
	for i, pt := range pts {
		grid[height-pt.Y+minY-1][pt.X-minX] = "#"
		xys[i] = &Point{pt.X - minX, height - pt.Y + minY - 1}
	}
	hl := []XY{}
	last := pts[len(pts)-1]
	if target.IsHit(last) {
		hl = append(hl, &Point{last.X - minX, height - last.Y + minY - 1})
	}
	return CreateIndexedGridString(grid, xys, hl)
}

type Probe struct {
	Position *Point
	Velocity *Point
	Path     Points
}

func NewProbe(dx, dy int) *Probe {
	rv := Probe{
		Position: &Point{0, 0},
		Velocity: &Point{dx, dy},
	}
	return &rv
}

func (p Probe) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("Position: %s, Velocity: %s", p.Position, p.Velocity))
	if len(p.Path) > 0 {
		rv.WriteString(", Path: ")
		rv.WriteString(p.Path.String())
	}
	return rv.String()
}

func (p *Probe) FireAt(target *TargetArea) (Points, bool) {
	for !target.IsHit(p.Position) && !target.WasMissedBy(p.Position) {
		p.Move()
	}
	return p.Path, target.IsHit(p.Position)
}

func (p *Probe) Move() {
	p.Position = &Point{p.Position.X + p.Velocity.X, p.Position.Y + p.Velocity.Y}
	p.Path = append(p.Path, p.Position)
	if p.Velocity.X > 0 {
		p.Velocity.X -= 1
	}
	p.Velocity.Y -= 1
}

type Point struct {
	X int
	Y int
}

func (p Point) String() string {
	return fmt.Sprintf("(%3d,%3d)", p.X, p.Y)
}

func (p Point) GetX() int {
	return p.X
}

func (p Point) GetY() int {
	return p.Y
}

type Points []*Point

func (p Points) String() string {
	var rv strings.Builder
	rv.WriteString(fmt.Sprintf("{%d}", len(p)))
	for _, pt := range p {
		rv.WriteByte(' ')
		rv.WriteString(pt.String())
	}
	return rv.String()
}

// -------------------------------------------------------------------------------------
// ----------------------  Input data structures and definitions  ----------------------
// -------------------------------------------------------------------------------------

type TargetArea struct {
	MinX int
	MaxX int
	MinY int
	MaxY int
}

func (t TargetArea) String() string {
	return fmt.Sprintf("x=%d..%d, y=%d..%d", t.MinX, t.MaxX, t.MinY, t.MaxY)
}

func (t TargetArea) IsHit(p *Point) bool {
	return p.X >= t.MinX && p.X <= t.MaxX && p.Y >= t.MinY && p.Y <= t.MaxY
}

func (t TargetArea) WasMissedBy(p *Point) bool {
	return p.X > t.MaxX || p.Y < t.MinY
}

// Input is a struct containing the parsed input file.
type Input struct {
	Verbose bool
	Target  *TargetArea
	DX      int
	DY      int
}

// String creates a mutli-line string representation of this Input.
func (i Input) String() string {
	return fmt.Sprintf("Target Area: %s", i.Target)
}

// ParseInput parses the contents of an input file into usable pieces.
func ParseInput(fileData []byte) (Input, error) {
	defer FuncEndingAlways(FuncStarting())
	rv := Input{}
	line := strings.TrimPrefix(strings.TrimSpace(string(fileData)), "target area: ")
	xy := strings.Split(line, ", ")
	xs := strings.Split(strings.TrimPrefix(xy[0], "x="), "..")
	ys := strings.Split(strings.TrimPrefix(xy[1], "y="), "..")
	rv.Target = &TargetArea{}
	var err error
	rv.Target.MinX, err = strconv.Atoi(xs[0])
	if err != nil {
		return rv, err
	}
	rv.Target.MaxX, err = strconv.Atoi(xs[1])
	if err != nil {
		return rv, err
	}
	rv.Target.MinY, err = strconv.Atoi(ys[0])
	if err != nil {
		return rv, err
	}
	rv.Target.MaxY, err = strconv.Atoi(ys[1])
	if err != nil {
		return rv, err
	}
	if rv.Target.MinX > rv.Target.MaxX {
		rv.Target.MinX, rv.Target.MaxX = rv.Target.MaxX, rv.Target.MinX
	}
	if rv.Target.MinY > rv.Target.MaxY {
		rv.Target.MinY, rv.Target.MaxY = rv.Target.MaxY, rv.Target.MinY
	}
	return rv, nil
}

const NOT_GIVEN = -9223372036854775808

// ApplyParams sets input based on CLI params.
func (i *Input) ApplyParams(params CliParams) {
	if params.Verbose {
		i.Verbose = true
	}
	i.DX = params.DX
	i.DY = params.DY
}

// -------------------------------------------------------------------------------------
// -----------------------------  CLI options and parsing  -----------------------------
// -------------------------------------------------------------------------------------

// CliParams contains anything that might be provided via command-line arguments.
type CliParams struct {
	// Debug is whether or not to output debug messages.
	Debug bool
	// Verbose is a flag indicating some extra output is desired.
	Verbose bool
	// HelpPrinted is whether or not the help message was printed.
	HelpPrinted bool
	// Errors is a list of errors encountered while parsing the arguments.
	Errors []error
	// InputFile is the file that contains the puzzle data to solve.
	InputFile string
	// Count is just a generic int that can be provided.
	Count int
	DX    int
	DY    int
}

// String creates a multi-line string representing this CliParams
func (c CliParams) String() string {
	nameFmt := "%20s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", c.Debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", c.Verbose),
		fmt.Sprintf(nameFmt+"%t", "Help Printed", c.HelpPrinted),
		fmt.Sprintf(nameFmt+"%q", "Errors", c.Errors),
		fmt.Sprintf(nameFmt+"%s", "Input File", c.InputFile),
		fmt.Sprintf(nameFmt+"%d", "Count", c.Count),
		fmt.Sprintf(nameFmt+"%d", "DX", c.DX),
		fmt.Sprintf(nameFmt+"%d", "DY", c.DY),
	}
	return strings.Join(lines, "\n") + "\n"
}

const default_input_file = "example.input"

// GetCliParams parses the provided args into the command's params.
func GetCliParams(args []string) CliParams {
	defer FuncEnding(FuncStarting())
	var err error
	rv := CliParams{}
	rv.DX = NOT_GIVEN
	rv.DY = NOT_GIVEN
	for i := 0; i < len(args); i++ {
		switch {
		// Flag cases go first.
		case IsOneOfStrFold(args[i], "--help", "-h", "help"):
			Debugf("Help flag found: [%s].", args[i])
			// Using fmt.Printf here instead of my stdout function because the extra formatting is annoying with help text.
			fmt.Printf("Usage: %s [<input file>]\n", GetCmdName())
			fmt.Printf("Default <input file> is %s\n", default_input_file)
			rv.HelpPrinted = true
		case HasPrefixFold(args[i], "--debug"):
			Debugf("Debug option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.Debug, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
			if err == nil {
				switch {
				case !debug && rv.Debug:
					debug = rv.Debug
					Stderr("Debugging enabled by CLI arguments.")
				case debug && !rv.Debug:
					Stderr("Debugging disabled by CLI arguments.")
					debug = rv.Debug
				}
			}
		case HasOneOfPrefixesFold(args[i], "--input", "--input-file"):
			Debugf("Input file option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.InputFile, extraI, err = ParseFlagString(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--count", "-c", "-n"):
			Debugf("Count option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.Count, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--verbose", "-v"):
			Debugf("Verbose option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.Verbose, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--dx", "-x", "-dx"):
			Debugf("DX option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.DX, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--dy", "-y", "-dy"):
			Debugf("DY option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			rv.DY, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)

		// Positional args go last in the order they're expected.
		case len(rv.InputFile) == 0 && len(args[i]) > 0 && args[i][0] != '-':
			Debugf("Input File argument: [%s].", args[i])
			rv.InputFile = args[i]
		default:
			Debugf("Unknown argument found: [%s], args left: %q.", args[i], args[i:])
			rv.AppendError(fmt.Errorf("unknown argument %d: [%s]", i+1, args[i]))
		}
	}
	rv.Debug = debug
	if len(rv.InputFile) == 0 {
		rv.InputFile = default_input_file
	}
	return rv
}

// -------------------------------------------------------------------------------------
// ----------  Still CLI parsing stuff, but stuff that should need changing  -----------
// -------------------------------------------------------------------------------------

// AppendError adds an error to this CliParams as long as the error is not nil.
func (c *CliParams) AppendError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// HasError returns true if this CliParams has one or more errors.
func (c CliParams) HasError() bool {
	return len(c.Errors) != 0
}

// Error flattens the Errors slice into a single string.
// It also makes the CliParams struct satisfy the error interface.
func (c *CliParams) Error() string {
	switch len(c.Errors) {
	case 0:
		return ""
	case 1:
		return c.Errors[0].Error()
	default:
		lines := []string{fmt.Sprintf("Found %d errors:", len(c.Errors))}
		for i, err := range c.Errors {
			lines = append(lines, fmt.Sprintf("  %d: %s", i, err.Error()))
		}
		return strings.Join(lines, "\n")
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
			for _, c := range []string{`'`, `"`} {
				if parts[1][:1] == c && parts[1][len(parts[1])-1:] == c {
					return parts[1][1 : len(parts[1])-1], 0, nil
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

// GetCmdName returns the name of this program by parsing os.Args[0].
func GetCmdName() string {
	_, name := filepath.Split(os.Args[0])
	return name
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d"
// DigitFormatForMax(382920) returns "%6d"
func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
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
	name := GetFuncName(1, a...)
	if debug {
		StderrAs(name, "Starting.")
	}
	funcDepth++
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
	if funcDepth > 0 {
		funcDepth--
	}
	if debug {
		StderrAs(name, done_fmt, time.Since(start))
	}
}

// FuncEndingAlways decrements the function depth and outputs how long a function took.
// If debug is on, output is to stderr, otherwise to stdout.
//
// This differs from FuncEnding in that this will always do the output (regardless of degub state).
//
// Usage: defer FuncEndingAlways(FuncStarting())
func FuncEndingAlways(start time.Time, name string) {
	if funcDepth > 0 {
		funcDepth--
	}
	if debug {
		StderrAs(name, done_fmt, time.Since(start))
	} else {
		StdoutAs(name, done_fmt, time.Since(start))
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
		if p.GetY() < height && p.GetX() < width && textFmt[p.GetY()][p.GetX()] <= 1 {
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
	cellFmt := fmt.Sprintf("%%%ds", cellLen)
	var digits strings.Builder
	for _, s := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		digits.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), 10))
	}
	rv := strings.Repeat(fmt.Sprintf(cellFmt, " "), 10) + strings.Repeat(digits.String(), 1+count/100)
	return rv[:count*cellLen]
}

func CreateIndexLinesHundreds(count, cellLen int) string {
	cellFmt := fmt.Sprintf("%%%ds", cellLen)
	var digits strings.Builder
	for _, s := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", " "} {
		digits.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), 100))
	}
	rv := strings.Repeat(fmt.Sprintf(cellFmt, " "), 100) + strings.Repeat(digits.String(), 1+count/1000)
	return rv[:count*cellLen]
}

// -------------------------------------------------------------------------------------
// ---------------------------------  Output wrappers  ---------------------------------
// -------------------------------------------------------------------------------------

// GetOutputPrefix gets the prefix to add to all output.
func GetOutputPrefix(funcName string) string {
	tabs := ""
	if debug {
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

// ReadFile is a wrapper on ioutil.ReadFile(filename) that adds output and timing.
func ReadFile(filename string) ([]byte, error) {
	defer FuncEndingAlways(FuncStarting(filename))
	Stdout("Reading input from file: %s", filename)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		Stderr("error reading file: %v", err)
	}
	return dat, err
}

// run does all the primary coordination for this program.
// It's basically a replacement for main() that returns an error.
func Run() error {
	defer FuncEndingAlways(FuncStarting())
	params := GetCliParams(os.Args[1:])
	Debugf("CLI Params:\n%s", params)
	if params.HelpPrinted {
		return nil
	}
	if params.HasError() {
		return &params
	}
	dat, err := ReadFile(params.InputFile)
	if err != nil {
		return err
	}
	Debugf("Input File Contents:\n%s", dat)
	input, err := ParseInput(dat)
	if err != nil {
		return err
	}
	input.ApplyParams(params)
	Debugf("Parsed Input:\n%s", input)
	answer, err := Solve(input)
	if err != nil {
		return err
	}
	Stdout("Answer: %s", answer)
	return nil
}
