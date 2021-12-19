package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	var sum *Pair
	for _, p := range input.Pairs {
		if sum == nil {
			sum = p
		} else {
			sum = AddPairs(sum, p)
		}
	}
	if debug {
		Stderr("Sum:\n%s", sum.FancyString())
	}
	answer := sum.GetMagnitude()
	return fmt.Sprintf("%d", answer), nil
}

func AddPairs(p1, p2 *Pair) *Pair {
	rv := Pair{
		Val: NAN,
		X:   p1,
		Y:   p2,
	}
	return rv.Reduce()
}

func (p *Pair) Reduce() *Pair {
	orig := ""
	if debug {
		orig = p.FancyString()
	}
	for {
		if debug {
			Stderr("Attempting to reduce:\n%s", p.FancyString())
		}
		if p.DoExplode() {
			Debugf("Explosion made.")
			continue
		}
		if p.DoSplit() {
			Debugf("Split made.")
			continue
		}
		break
	}
	if debug {
		Stderr("Finished reducing.\nWas:\n%sIs Now:\n%s", orig, p.FancyString())
	}
	return p
}

func (p *Pair) DoExplode() bool {
	pairs := p.Enumerate(0)
	pivot := NAN
	var x, y int
	for i, pair := range pairs {
		if pair.Depth == 4 && pair.Val == NAN {
			pivot = i
			x = pair.X.Val
			y = pair.Y.Val
			pair.Val = 0
			pair.X = nil
			pair.Y = nil
			break
		}
	}
	if pivot == NAN {
		return false
	}
	xto, yto := -1, -1
	for i := pivot - 1; i >= 0; i-- {
		if pairs[i].Val != NAN {
			pairs[i].Val += x
			xto = i
			break
		}
	}
	for i := pivot + 3; i < len(pairs); i++ {
		if pairs[i].Val != NAN {
			pairs[i].Val += y
			yto = i
			break
		}
	}
	Debugf("Exploded at %d which spread to %d and %d", pivot, xto, yto)
	return true
}

func (p *Pair) Enumerate(depth int) []*Pair {
	p.Depth = depth
	rv := []*Pair{p}
	if p.Val == NAN {
		rv = append(rv, p.X.Enumerate(depth+1)...)
		rv = append(rv, p.Y.Enumerate(depth+1)...)
	}
	return rv
}

// DoSplit recursively checks for and applies a split if needed. It returns true if a split was made.
func (p *Pair) DoSplit() bool {
	if p.Val != NAN {
		if p.Val > 9 {
			p.X = &Pair{Val: p.Val / 2}
			p.Y = &Pair{Val: p.Val/2 + p.Val%2}
			p.Val = NAN
			return true
		}
		return false
	}
	return p.X.DoSplit() || p.Y.DoSplit()
}

func (p Pair) GetMagnitude() int {
	if p.Val != NAN {
		return p.Val
	}
	return p.X.GetMagnitude()*3 + p.Y.GetMagnitude()*2
}

// -------------------------------------------------------------------------------------
// ----------------------  Input data structures and definitions  ----------------------
// -------------------------------------------------------------------------------------

const NAN = -9223372036854775808

type Pair struct {
	Val   int
	X     *Pair
	Y     *Pair
	Depth int
}

func NewPair(str string) (*Pair, error) {
	pair, rest, err := ParsePairString(str)
	if err != nil {
		return nil, err
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("the ParsePairString func returned with leftovers: [%s]", rest)
	}
	return pair, nil
}

func ParsePairString(str string) (*Pair, string, error) {
	rv := Pair{
		Val: NAN,
	}
	switch str[0] {
	case '1':
		rv.Val = 1
	case '2':
		rv.Val = 2
	case '3':
		rv.Val = 3
	case '4':
		rv.Val = 4
	case '5':
		rv.Val = 5
	case '6':
		rv.Val = 6
	case '7':
		rv.Val = 7
	case '8':
		rv.Val = 8
	case '9':
		rv.Val = 9
	case '0':
		rv.Val = 0
	case '[':
		var err error
		rv.X, str, err = ParsePairString(str[1:])
		if err != nil {
			return nil, "", err
		}
		if str[0] != ',' {
			return nil, "", fmt.Errorf("unable to parse %q expected ',', found %q", str, str[0])
		}
		rv.Y, str, err = ParsePairString(str[1:])
		if err != nil {
			return nil, "", err
		}
		if str[0] != ']' {
			return nil, "", fmt.Errorf("unable to parse %q expected ']', found %q", str, str[0])
		}
	}
	return &rv, str[1:], nil
}

func (p Pair) String() string {
	if p.Val != NAN {
		return fmt.Sprintf("%d", p.Val)
	}
	return fmt.Sprintf("[%s,%s]", p.X, p.Y)
}

var hlRx = regexp.MustCompile(`(\d\d)`)

func (p Pair) FancyString() string {
	defer FuncEndingAlways(FuncStarting())
	str := p.String()
	blankLine := func() []byte {
		rv := make([]byte, len(str))
		for i := range rv {
			rv[i] = ' '
		}
		return rv
	}
	lines := make([][]byte, 7)
	for i := range lines {
		lines[i] = blankLine()
	}
	maxDepth := 0
	curDepth := 0
	for i := range str {
		b := str[i]
		lines[curDepth][i] = b
		switch b {
		case '[', ',':
			curDepth += 1
			if curDepth > maxDepth {
				maxDepth = curDepth
			}
			if curDepth == len(lines) {
				lines = append(lines, blankLine())
			}
		case ']':
			curDepth -= 1
		default:
			switch str[i+1] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// Do nothing.
			default:
				curDepth -= 1
			}
		}
	}
	var rv strings.Builder
	rv.WriteString("  ")
	rv.WriteString(str)
	rv.WriteByte('\n')
	var cb, ce string
	for i := 0; i <= maxDepth; i++ {
		switch i {
		case 4:
			cb = "\033[97m" // Bright white
			ce = "\033[0m"  // Color off
		case 5:
			cb = "\033[95m" // Bright Purple
		case 6:
			cb = "\033[31m" // Red
		case 7:
			cb = "\033[41;97m" // Bright white text with red background
		}
		rv.WriteString(cb)
		rv.WriteString(fmt.Sprintf("%d:", i))
		rv.Write(hlRx.ReplaceAll(lines[i], []byte("\033[7m$1\033[27m")))
		rv.WriteString(ce)
		rv.WriteByte('\n')
	}
	return rv.String()
}

// Input is a struct containing the parsed input file.
type Input struct {
	Verbose bool
	Pairs   []*Pair
}

// String creates a mutli-line string representation of this Input.
func (i Input) String() string {
	lineFmt := DigitFormatForMax(len(i.Pairs)) + ": %s\n"
	var rv strings.Builder
	for i, v := range i.Pairs {
		rv.WriteString(fmt.Sprintf(lineFmt, i, v))
	}
	return rv.String()
}

// ParseInput parses the contents of an input file into usable pieces.
func ParseInput(fileData []byte) (Input, error) {
	defer FuncEndingAlways(FuncStarting())
	rv := Input{}
	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			pair, err := NewPair(line)
			if err != nil {
				return rv, err
			}
			rv.Pairs = append(rv.Pairs, pair)
		}
	}
	return rv, nil
}

// ApplyParams sets input based on CLI params.
func (i *Input) ApplyParams(params CliParams) error {
	if params.Verbose {
		i.Verbose = true
	}
	if len(params.Custom) > 0 {
		i.Pairs = make([]*Pair, len(params.Custom))
		var err error
		for j, custom := range params.Custom {
			i.Pairs[j], err = NewPair(custom)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	Count  int
	Custom []string
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
	}
	return strings.Join(lines, "\n") + "\n"
}

const default_input_file = "example.input"

// GetCliParams parses the provided args into the command's params.
func GetCliParams(args []string) CliParams {
	defer FuncEnding(FuncStarting())
	var err error
	rv := CliParams{}
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
		case HasOneOfPrefixesFold(args[i], "--custom", "-p", "--pair"):
			Debugf("Custom option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			var custom string
			custom, extraI, err = ParseFlagString(args[i:])
			rv.Custom = append(rv.Custom, custom)
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
	err = input.ApplyParams(params)
	if err != nil {
		return err
	}
	Debugf("Parsed Input:\n%s", input)
	answer, err := Solve(input)
	if err != nil {
		return err
	}
	Stdout("Answer: %s", answer)
	return nil
}
