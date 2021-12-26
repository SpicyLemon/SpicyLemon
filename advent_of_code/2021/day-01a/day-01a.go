package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// debug is a flag for whether or not debug messages should be displayed.
var debug bool

// -------------------------------------------------------------------------------------
// ------------------  Stuff that I will probably want to customize  -------------------
// -------------------------------------------------------------------------------------

// solve is the main entry point to finding a solution.
// The string it returns should be the answer.
func solve(input Input) (string, error) {
	defer timeFunc(time.Now(), "solve")
	count := 0
	for i := 1; i < len(input.Depths); i++ {
		if input.Depths[i] > input.Depths[i-1] {
			count++
		}
	}
	return fmt.Sprintf("%d", count), nil
}

// Input is a struct containing the parsed input file.
type Input struct {
	Depths []int
}

// String creates a mutli-line string representation of this Input.
func (i Input) String() string {
	lineFmt := "%" + fmt.Sprintf("%d", len(fmt.Sprintf("%d", len(i.Depths)))) + "d: %d\n"
	var rv strings.Builder
	for i, d := range i.Depths {
		rv.WriteString(fmt.Sprintf(lineFmt, i, d))
	}
	return rv.String()
}

// parseInput parses the contents of an input file into usable pieces.
func parseInput(fileData []byte) (Input, error) {
	defer debugTimeFunc(time.Now(), "parseInput")
	rv := Input{}
	for _, line := range strings.Split(string(fileData), "\n") {
		if len(line) > 0 {
			d, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				return rv, err
			}
			rv.Depths = append(rv.Depths, int(d))
		}
	}
	return rv, nil
}

// ApplyParams sets input based on CLI params.
func (i *Input) ApplyParams(params CliParams) {
	// TODO: If there are any command-line arguments to apply to the puzzle input, pass them through in here.
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
}

// String creates a multi-line string representing this CliParams
func (c CliParams) String() string {
	nameFmt := "%20s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", c.Debug),
		fmt.Sprintf(nameFmt+"%t", "Help Printed", c.HelpPrinted),
		fmt.Sprintf(nameFmt+"%q", "Errors", c.Errors),
		fmt.Sprintf(nameFmt+"%s", "Input File", c.InputFile),
	}
	return strings.Join(lines, "\n")
}

// getCliParams parses the provided args into the command's params.
func getCliParams(args []string) CliParams {
	defer debugTimeFunc(time.Now(), "getCliParams")
	var err error
	rv := CliParams{}
	debug, err = getEnvVarBool("DEBUG")
	rv.Debug = debug
	rv.AppendError(err)
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
		case hasOneOfPrefixes(args[i], "--input", "-i"):
			var extraI int
			rv.InputFile, extraI, err = parseFlagString(args[i:])
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

// ApplyDefaults applies any default tag values to fields that haven't been set yet.
func (c *CliParams) ApplyDefaults() {
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

// getCmdName returns the name of this program by parsing os.Args[0].
func getCmdName() string {
	_, name := filepath.Split(os.Args[0])
	return name
}

// debugf outputs to stderr if the debug flag is set.
func debugf(format string, a ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

// notDebugf outputs to stdout if the debug flag is NOT set.
func notDebugf(format string, a ...interface{}) {
	if !debug {
		fmt.Printf(format, a...)
	}
}

// debugTimeFunc is a function for outputting (to stderr if we're debugging) how long a function takes.
// Usage: defer debugTimeFunc(time.Now(), "function name")
func debugTimeFunc(start time.Time, name string) {
	if debug {
		d := time.Since(start)
		fmt.Fprintf(os.Stderr, "Run time [%s]: %s\n", name, d)
	}
}

// timeFunc is a function for outputting how long a function takes.
// Usage: defer timeFunc(time.Now(), "function name")
func timeFunc(start time.Time, name string) {
	d := time.Since(start)
	fmt.Printf("Run time [%s]: %s\n", name, d)

}

// Run does all the primary coordination for this program.
// It's basically a replacement for main() that returns an error.
func Run() error {
	defer timeFunc(time.Now(), "Run")
	params := getCliParams(os.Args[1:])
	debugf("CLI Params:\n%s\n", params)
	if params.HelpPrinted {
		return nil
	}
	if params.HasError() {
		return &params
	}
	notDebugf("CLI Params:\n%s\n", params)
	fmt.Printf("Reading input from file: %s\n", params.InputFile)
	dat, err := ioutil.ReadFile(params.InputFile)
	if err != nil {
		return err
	}
	debugf("Input File Contents:\n%s\n", dat)
	input, err := parseInput(dat)
	if err != nil {
		return err
	}
	input.ApplyParams(params)
	debugf("Parsed Input:\n%s\n", input)
	answer, err := solve(input)
	if err != nil {
		return err
	}
	fmt.Printf("Answer: %s\n", answer)
	return nil
}

// main is the main function that gets run for this file.
func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
