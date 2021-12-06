package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// debug is a flag for whether or not debug messages should be displayed.
var debug bool

// startTime is the time when the program started.
var startTime time.Time

// indentOn is a flag indicating whether or not to indent start/stop messages based on function depth.
var indentOn bool

// funcDepth is a global counter keeping track of function depth.
// It is only updated if debug == true and indentOn == true, and only updated during funcStarting and funcEnding.
var funcDepth int

func init() {
	startTime = time.Now()
	indentOn = true
}

// -------------------------------------------------------------------------------------
// ------------------  Stuff that I will probably want to customize  -------------------
// -------------------------------------------------------------------------------------

// solve is the main entry point to finding a solution.
// The string it returns should be the answer.
func solve(input Input) (string, error) {
	defer funcTimeEnding(funcStarting())
	dayMap := map[int]int{}
	for i := 0; i <= 8; i++ {
		dayMap[i] = 0
	}
	for _, c := range input.DaysLeft {
		dayMap[c]++
	}
	debugf("Initial Day Counts: %#v", dayMap)
	evolve := func() {
		zeros := dayMap[0]
		for i := 1; i <= 8; i++ {
			dayMap[i-1] = dayMap[i]
		}
		dayMap[8] = zeros
		dayMap[6] += zeros
	}
	for i := 0; i < input.DayCount; i++ {
		evolve()
	}
	debugf("Final Day Counts: %#v", dayMap)
	answer := 0
	for _, v := range dayMap {
		answer += v
	}
	return fmt.Sprintf("%d", answer), nil
}

// Input is a struct containing the parsed input file.
type Input struct {
	DaysLeft []int
	DayCount int
}

// String creates a mutli-line string representation of this Input.
func (i Input) String() string {
	//lineFmt := "%" + fmt.Sprintf("%d", len(fmt.Sprintf("%d", len(i.Lines)))) + "d: %s\n"
	//var rv strings.Builder
	//for i, v := range i.Lines {
	//	rv.WriteString(fmt.Sprintf(lineFmt, i, v))
	//}
	//return rv.String()
	return fmt.Sprintf("Day Count: %d: %v (%d)", i.DayCount, i.DaysLeft, len(i.DaysLeft))
}

// parseInput parses the contents of an input file into usable pieces.
func parseInput(fileData []byte) (Input, error) {
	defer funcTimeEnding(funcStarting())
	rv := Input{}
	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			nums := strings.Split(line, ",")
			for _, num := range nums {
				i, err := strconv.Atoi(num)
				if err != nil {
					return rv, err
				}
				rv.DaysLeft = append(rv.DaysLeft, i)
			}
		}
	}
	return rv, nil
}

// ApplyParams sets input based on CLI params.
func (i *Input) ApplyParams(params CliParams) {
	// TODO: If there are any command-line arguments to apply to the puzzle input, pass them through in here.
	switch {
	case params.DayCount != 0:
		i.DayCount = params.DayCount
	case i.DayCount == 0:
		i.DayCount = 256
	}
}

// CliParams contains anything that might be provided via command-line arguments.
type CliParams struct {
	// Debug is whether or not to output debug messages.
	Debug bool
	// HelpPrinted is whether or not the help message was printed.
	HelpPrinted bool
	// Errors is a list of errors encountered while parsing the arguments.
	Errors []error
	// InputFile is the file that contains the puzzle data to solve.
	InputFile string `default:"example.input"`
	// DayCount is the number of days to go through in the simulation (for day 6).
	DayCount int
}

// String creates a multi-line string representing this CliParams
func (c CliParams) String() string {
	nameFmt := "%20s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", c.Debug),
		fmt.Sprintf(nameFmt+"%t", "Help Printed", c.HelpPrinted),
		fmt.Sprintf(nameFmt+"%q", "Errors", c.Errors),
		fmt.Sprintf(nameFmt+"%s", "Input File", c.InputFile),
		fmt.Sprintf(nameFmt+"%d", "Days Count", c.DayCount),
	}
	return strings.Join(lines, "\n")
}

// getCliParams parses the provided args into the command's params.
func getCliParams(args []string) CliParams {
	defer funcEnding(funcStarting())
	var err error
	rv := CliParams{}
	rv.Debug = debug
	for i := 0; i < len(args); i++ {
		switch {
		// Flag cases go first.
		case isOneOfStrFold(args[i], "--help", "-h", "help"):
			fmt.Printf("Usage: %s [<input file>]\n", getCmdName())
			fmt.Printf("Default <input file> is example.input\n")
			rv.HelpPrinted = true
		case strings.HasPrefix(args[i], "--debug"):
			var extraI int
			debug, extraI, err = parseFlagBool(args[i:])
			rv.Debug = debug
			i += extraI
			rv.AppendError(err)
			funcDepth = 2
		case hasOneOfPrefixes(args[i], "--input", "-i"):
			var extraI int
			rv.InputFile, extraI, err = parseFlagString(args[i:])
			i += extraI
			rv.AppendError(err)
		case hasOneOfPrefixes(args[i], "--day-count", "-n"):
			var extraI int
			rv.DayCount, extraI, err = parseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)

		// Positional args go last in the order they're expected.
		case len(rv.InputFile) == 0:
			rv.InputFile = args[i]
		default:
			rv.AppendError(fmt.Errorf("unknown argument %d: [%s]", i+1, args[i]))
		}
	}
	rv.ApplyDefaults()
	return rv
}

// -------------------------------------------------------------------------------------
// --------------  Standard stuff that I hopefully don't need to change  ---------------
// -------------------------------------------------------------------------------------

// handleEnvVars looks at specific environment variables and sets global variables appropriately.
func handleEnvVars() error {
	var err error
	debug, err = getEnvVarBool("DEBUG")
	return err
}

// ApplyDefaults applies any default tag values to fields that haven't been set yet.
func (c *CliParams) ApplyDefaults() {
	defer funcEnding(funcStarting())
	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(c)
	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		sf := t.Field(i)
		dtag := sf.Tag.Get("default")
		if len(dtag) == 0 {
			continue
		}
		fv := v.Elem().Field(i)
		if !fv.IsZero() {
			continue
		}
		debugf("Applying default value [%s] to field [%s] of kind %s.\n", dtag, sf.Name, fv.Kind())
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(dtag)
		case reflect.Int, reflect.Int64:
			fv.SetInt(mustParseInt(dtag, 10, 64))
		case reflect.Int32:
			fv.SetInt(mustParseInt(dtag, 10, 32))
		case reflect.Int16:
			fv.SetInt(mustParseInt(dtag, 10, 16))
		case reflect.Int8:
			fv.SetInt(mustParseInt(dtag, 10, 8))
		case reflect.Uint, reflect.Uint64:
			fv.SetUint(mustParseUint(dtag, 10, 64))
		case reflect.Uint32:
			fv.SetUint(mustParseUint(dtag, 10, 32))
		case reflect.Uint16:
			fv.SetUint(mustParseUint(dtag, 10, 16))
		case reflect.Uint8:
			fv.SetUint(mustParseUint(dtag, 10, 8))
		default:
			panic(fmt.Sprintf("Cannot define default tag value for field %s.%s with kind %s.", t.Name(), fv.Type().Name(), fv.Kind()))
		}
	}
}

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

// isOneOfStrFold tests if the given string is equal (ignoring case) to one of the given options.
func isOneOfStrFold(str string, opts ...string) bool {
	for _, opt := range opts {
		if strings.EqualFold(str, opt) {
			return true
		}
	}
	return false
}

// hasOneOfPrefixes tests if the given string has one of the given prefixes.
func hasOneOfPrefixes(str string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(str, pre) {
			return true
		}
	}
	return false
}

// mustParseInt is the same as strconv.ParseInt except this panics on error.
func mustParseInt(str string, base, bitSize int) int64 {
	rv, err := strconv.ParseInt(str, base, bitSize)
	if err != nil {
		panic(err)
	}
	return rv
}

// mustParseUint is the same as strconv.ParseUint except this panics on error.
func mustParseUint(str string, base, bitSize int) uint64 {
	rv, err := strconv.ParseUint(str, base, bitSize)
	if err != nil {
		panic(err)
	}
	return rv
}

// parseBool converts a string into a bool.
// First return bool is the parsed value.
// Second return bool is whether or not the parsing was successful.
func parseBool(str string) (val bool, isBool bool) {
	lstr := strings.ToLower(strings.TrimSpace(str))
	switch lstr {
	case "false", "f", "no", "n", "0":
		isBool = true
	case "true", "t", "yes", "y", "1":
		val = true
		isBool = true
	}
	return
}

// parseFlagBool parses a boolean flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and parsed.
// Otherwise, if args[1] is a boolean string value, that is parsed.
// Otherwise, the flag defaults to true.
//
// The first return value is the parsed boolean value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func parseFlagBool(args []string) (bool, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) == 2 {
			val, isBool := parseBool(parts[1])
			if !isBool {
				return false, 0, fmt.Errorf("invalid %s bool value: [%s]", parts[0], parts[1])
			}
			return val, 0, nil
		}
		return false, 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
	}
	if len(args) > 1 {
		val, isBool := parseBool(args[1])
		if isBool {
			return val, 1, nil
		}
	}
	return true, 0, nil
}

// parseFlagString parses a string flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's string value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func parseFlagString(args []string) (string, int, error) {
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

// parseFlagInt parses an int flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's int value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func parseFlagInt(args []string) (int, int, error) {
	rvStr, used, err := parseFlagString(args)
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

// getEnvVarBool gets the environment variable with the given name and converts it to a bool.
func getEnvVarBool(name string) (bool, error) {
	str := os.Getenv(name)
	if len(str) == 0 {
		return false, nil
	}
	val, isBool := parseBool(str)
	if !isBool {
		return false, fmt.Errorf("invalid %s env var boolean value: [%s]", name, str)
	}
	return val, nil
}

// stdout outputs to stdout with a prefixed run duration prefix.
func stdout(format string, a ...interface{}) {
	args := append([]interface{}{durClock(time.Since(startTime))}, a...)
	fmt.Printf("(%14s) "+format+"\n", args...)
}

// stderr outputs to stderr with a prefixed run duration prefix.
func stderr(format string, a ...interface{}) {
	args := append([]interface{}{durClock(time.Since(startTime))}, a...)
	fmt.Fprintf(os.Stderr, "(%14s) "+format+"\n", args...)
}

// debugf outputs to stderr if the debug flag is set.
func debugf(format string, a ...interface{}) {
	if debug {
		stderr(format, a...)
	}
}

// funcStarting outputs that a function is starting (if debug is true).
// It returns the params needed by funcEnding.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Only values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
//
// Example 1: In a function named "foo", you have this:
//     funcStarting()
//   The printed message will note that "foo" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//     funcStarting(3 * time.Second)
//   The printed message will note that "bar: 3s" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//     func sum(ints ...int) {
//         funcStarting(ints...)
//     }
//     sum(1, 2, 3, 4, 20, 21, 22)
//   The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer funcEnding(funcStarting())
func funcStarting(a ...interface{}) (time.Time, string) {
	if debug {
		name := getFuncName(1)
		switch len(a) {
		case 0:
			// do nothing
		case 1:
			name += fmt.Sprintf(": %v", a[0])
		case 2:
			name += fmt.Sprintf(": %v, %v", a[0], a[1])
		case 3:
			name += fmt.Sprintf(": %v, %v, %v", a[0], a[1], a[2])
		case 4:
			name += fmt.Sprintf(": %v, %v, %v, %v", a[0], a[1], a[2], a[3])
		default:
			args := make([]string, len(a))
			for i, arg := range a {
				args[i] = fmt.Sprintf("%v", arg)
			}
			name += fmt.Sprintf(": %s", strings.Join(args, ", "))
		}
		tabs := ""
		if indentOn {
			tabs = strings.Repeat("  ", funcDepth)
			funcDepth++
		}
		stdout("%s[%s] Starting.", tabs, name)
		return time.Now(), name
	}
	return time.Now(), ""
}

// funcEnding outputs that a function is ending (if debugging).
// Args will usually come from funcStarting().
//
// Use this when you want a start/done debug message but don't want the time duration otherwise.
//
// Standard Usage: defer funcEnding(funcStarting())
func funcEnding(start time.Time, name string) {
	if debug {
		d := time.Since(start)
		if len(name) == 0 {
			// Can happen if debug is turned on after the funcStarting() call was made.
			name = getFuncName(1)
		}
		tabs := ""
		if indentOn {
			funcDepth--
			if funcDepth < 0 {
				funcDepth = 0
			}
			tabs = strings.Repeat("  ", funcDepth)
		}
		stdout("%s[%s] Done. Duration: [%s].", tabs, name, d)
	}
}

// funcTimeEnding calls funcEnding if debug is on, otherwise it calls timeFunc.
//
// Use this when you always want the time printed, but if debug is on, use the start/done messages.
//
// Usage: defer funcTimeEnding(funcStarting())
func funcTimeEnding(start time.Time, name string) {
	if len(name) == 0 {
		name = getFuncName(1)
	}
	if debug {
		funcEnding(start, name)
	} else {
		timeFunc(start, name)
	}
}

// timeFunc is a function for outputting how long a function takes.
// If the function name is given as "", it will be reflectively looked up.
//
// Use this when you don't want the debug start/done messages but do want a function timed.
//
// Usage: defer timeFunc(time.Now(), "function name")
func timeFunc(start time.Time, name string) {
	d := time.Since(start)
	if len(name) == 0 {
		name = getFuncName(1)
	}
	stdout("[%s] Done. Duration: [%s].", name, d)
}

// durClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNNs", e.g. "12:01:02.000000000"
// - If less than one hour, format is "M:SS.NNNNNNNNNs", e.g. "34:00.000000789"
// - If less than one minute, format is "S.NNNNNNNNNs", e.g. "56.000456000"
// - If less than one second, format is "0.NNNNNNNNNs", e.g. "0.123000000"
func durClock(d time.Duration) string {
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

// getFuncName gets the name of the function at the given depth.
//
// depth 0 = the function calling getFuncName.
// depth 1 = the function calling the function calling getFuncName.
// etc.
func getFuncName(depth int) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	for more && depth > 0 {
		frame, more = frames.Next()
		depth--
	}
	return strings.TrimPrefix(frame.Function, "main.")
}

// getCmdName returns the name of this program by parsing os.Args[0].
func getCmdName() string {
	_, name := filepath.Split(os.Args[0])
	return name
}

// readFile is a wrapper on ioutil.ReadFile(filename) that adds output and timing.
func readFile(filename string) ([]byte, error) {
	defer funcTimeEnding(funcStarting(filename))
	stdout("Reading input from file: %s", filename)
	return ioutil.ReadFile(filename)
}

// Run does all the primary coordination for this program.
// It's basically a replacement for main() that returns an error.
func Run() error {
	defer funcTimeEnding(funcStarting())
	params := getCliParams(os.Args[1:])
	debugf("CLI Params:\n%s", params)
	if params.HelpPrinted {
		return nil
	}
	if params.HasError() {
		return &params
	}
	dat, err := readFile(params.InputFile)
	if err != nil {
		return err
	}
	debugf("Input File Contents:\n%s", dat)
	input, err := parseInput(dat)
	if err != nil {
		return err
	}
	input.ApplyParams(params)
	debugf("Parsed Input:\n%s", input)
	answer, err := solve(input)
	if err != nil {
		return err
	}
	stdout("Answer: %s", answer)
	return nil
}

// main is the main function that gets run for this file.
func main() {
	err := handleEnvVars()
	if err == nil {
		err = Run()
	}
	if err != nil {
		stderr("error: %v", err)
		os.Exit(1)
	}
}
