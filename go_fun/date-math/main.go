package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// PrintUsage writes a message that describes how to invoke this program to the provided writer (e.g. os.Stdout).
func PrintUsage(stdout io.Writer) {
	fmt.Fprintln(stdout, `date-math: Do calculations with datetimes and durations.

Usage: date-math (<formula>|formats) [flags]

A <formula> has the format <value> <op> (<value>|<formula>)

A <value> can either be a <date>, <epoch>, <dur>, or <num>.
  <time> A datetime string. Multiple formats are supported.
         To see all possible formats, execute: date-math formats
         Datetimes that do not have a time zone are assumed to be local which is
         controllable by setting the TZ environment variable.
  <epoch> A possibly signed number with optional fractional seconds.
          An <epoch> is treated as a <time> for the purposes of the calculations.
  <dur> A possibly signed sequence of decimal numbers, each with optional fraction
        and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
        The "d" and "w" time units are non-standard and represent days and weeks.
        It's assumed that 1w = 7d and 1d = 24h = 1440m = 86400s, even though that
        isn't always the case, e.g. time changes and leap seconds.
  <num> A possibly signed whole number.

A whole number might be either an <epoch> or <num>. By default, a whole number
greater than 1,000,000 or less than -1,000,000 is treated as an <epoch>. A whole
number between -1,000,000 and -1,000,000 (inclusive) is treated as a <num>.
To force a whole number to be an <epoch>, prepend it with 'e', e.g. 'e1000000'.
To force a whole number to be a <num>, prepend it with 'n', e.g. 'n1000001'.

The <op> can be + - x or /. Only the following operations are defined:
  <time> - <time> => <dur>   e.g. 2020-01-09 4:30:00 - 2020-01-09 3:29:28 => 1h2s
                               or 2020-01-09 3:29:28 - 2020-01-09 4:30:00 => -1h2s
  <time> + <dur>  => <time>  e.g. 2020-01-09 4:30:00 + 1h2s => 2020-01-09 5:30:02
  <dur>  + <time> => <time>  e.g. 1h2s + 2020-01-09 4:30:00 => 2020-01-09 5:30:02
  <time> - <dur>  => <time>  e.g. 2020-01-09 4:30:00 - 1h2s => 2020-01-09 3:29:28
  <dur>  + <dur>  => <dur>   e.g. 1h2s + 3m5s => 1h3m7s (communicative)
  <dur>  - <dur>  => <dur>   e.g. 1h2s - 3m5s => 56m57s  or  3m5s - 1h2s => -56m57s
  <dur>  / <dur>  => <num>   e.g. 2h / 40m => 3
  <dur>  x <num>  => <dur>   e.g. 40m x 3 => 2h
  <num>  x <dur>  => <dur>   e.g. 5 x 40m => 3h20m
  <dur>  / <num>  => <dur>   e.g. 2h / 3 => 40m
  <num>  + <num>  => <num>   e.g. 5 + 3 => 8 (communicative)
  <num>  - <num>  => <num>   e.g. 5 - 3 => 2  or  3 - 5 => -2
  <num>  x <num>  => <num>   e.g. 5 x 3 => 15 (communicative)
  <num>  / <num>  => <num>   e.g. 6 / 3 => 2  or  5 / 3 => 1

Notes:
1. Those examples might have slightly different output, but same values.
2. Division is done using integers which will truncate the result.
   A <dur> is handled as an integer amount of nanoseconds. So <dur> / <num>
   will be truncated to the nearest nanosecond.
3. Multiplication is done using x instead of * because shells will expand *,
   and I didn't want to have to always remember to escape it.


The <formula> is calculated from left to right and can have multiple operations.
E.g. 2020-01-09 4:30:00 + 1h2s - 2020-01-02 11:30:18
   = 2020-01-09 5:30:02 - 2020-01-02 11:30:18
   = 6d17h59m44s

If "formats" is provided the list of named datetime format strings is printed.
These are the valid names to provide with the --output flag.

There are a few flags that can also be provided:
  --output-name|-o <name>
        Use the format with the provided <name> to convert a final <time> value
        into the result. Does nothing if the final result isn't a <time>.
        See: date-math formats
  --output-format|-f <format>
        Use the provided <format> to convert a final <time> value into the
        result. Does nothing if the final result isn't a <time>.
        See: https://pkg.go.dev/time#pkg-constants
  --input-name|-i <name>
        Use the format with the provided <name> to parse any provided <time>
        values. When this option is used, none of the other formats will be
        considered for parsing. If the final result is a <time> it will also
        have this format, unless either --output-name or --output-format are used.
        See: date-math formats
  --input-format|-g <format>
        Use the provided input to parse any provided <time> values. When this
        option is used, none of the other formats will be available. If the final
        result is a <time> it will also have this format, unless either
        --output-name or --output-format are used.
  --formats
        Same as providing just "formats"; outputs info on all named formats.
  --pipe|-p
        Read formula args from stdin and run the calculation for each line.
        Each line is inserted in place of the --pipe or -p flag among any other
        formula args that are provided. This allows for piping in values, ops,
        partial formulas, or full formulas. This flag can be omitted if there
        are no other formula args to provide.
  --verbose|-v
        Print debugging information to stderr.
        Can also be enabled by setting the VERBOSE env var.
  --help|-h
        Output this message.`)
}

// calcArgs contains the formula args and info for handling piped in input.
type calcArgs struct {
	// All calculation args with values combined into a single arg.
	All []string
	// HavePipe indicates whether there's a pipe indicator in the All slice.
	HavePipe bool
	// PrePipe is stuff in All that is before the pipe indicator (empty if no pipe indicator).
	PrePipe []string
	// PostPipe is stuff in All that is after the pipe indicator (empty if no pipe indicator).
	PostPipe []string
}

// getArgs handles flags and options and returns the combined formula args and whether to stop early.
// If help or formats are requested, this will print that to the provided writer (e.g. os.Stdout).
func getArgs(argsIn []string, stdout io.Writer) (*calcArgs, bool, error) {
	fArgs, stop, err := processFlags(argsIn, stdout)
	if stop || err != nil {
		return &calcArgs{}, stop, err
	}
	verbosef(" formula args: %q", fArgs)
	return combineArgs(fArgs), false, nil
}

// processFlags will look through argsIn, handle any flags and return just the formula args and whether to stop early.
// If help or formats are requested, this will print that to the provided writer (e.g. os.Stdout).
func processFlags(argsIn []string, stdout io.Writer) ([]string, bool, error) {
	var argsOut []string
	verbosef("Args provided (%d):", len(argsIn))
	for i := 0; i < len(argsIn); i++ {
		rawArg := argsIn[i]
		arg := strings.TrimSpace(rawArg)
		switch {
		case EqualFoldOneOf(arg, "--help", "-h", "help"):
			verbosef("[%d]: help arg identified, %q", i, rawArg)
			PrintUsage(stdout)
			return nil, true, nil

		case EqualFoldOneOf(arg, "--formats", "formats"):
			verbosef("[%d]: formats arg identified, %q", i, rawArg)
			PrintFormats(stdout)
			return nil, true, nil

		case EqualFoldOneOf(arg, "--verbose", "-v"):
			Verbose = true
			verbosef("[%d]: verbose flag identified, %q", i, rawArg)

		case EqualFoldOneOf(arg, "--output-name", "-o"):
			verbosef("[%d]: output-name arg identified, %q", i, rawArg)
			if i+1 >= len(argsIn) {
				return nil, true, fmt.Errorf("no argument provided after %s, expected a format name", arg)
			}
			i++
			verbosef("[%d]: output-name value identified, %q", i, argsIn[i])
			if err := setOutputFormatByName(argsIn[i], stdout); err != nil {
				return nil, true, err
			}

		case EqualFoldOneOf(arg, "--output-format", "-f"):
			verbosef("[%d]: output-format arg identified, %q", i, rawArg)
			if i+1 >= len(argsIn) {
				return nil, true, fmt.Errorf("no argument provided after %s, expected a format string", arg)
			}
			i++
			verbosef("[%d]: output-format value identified, %q", i, argsIn[i])
			if err := setOutputFormatByValue(argsIn[i]); err != nil {
				return nil, true, err
			}

		case EqualFoldOneOf(arg, "--input-name", "-i"):
			verbosef("[%d]: input-name arg identified, %q", i, rawArg)
			if i+1 >= len(argsIn) {
				return nil, true, fmt.Errorf("no argument provided after %s, expected a format name", arg)
			}
			i++
			verbosef("[%d]: input-name value identified, %q", i, argsIn[i])
			if err := setInputFormatByName(argsIn[i], stdout); err != nil {
				return nil, true, err
			}

		case EqualFoldOneOf(arg, "--input-format", "-g"):
			verbosef("[%d]: input-format arg identified, %q", i, rawArg)
			if i+1 >= len(argsIn) {
				return nil, true, fmt.Errorf("no argument provided after %s, expected a format string", arg)
			}
			i++
			verbosef("[%d]: input-format value identified, %q", i, argsIn[i])
			if err := setInputFormatByValue(argsIn[i]); err != nil {
				return nil, true, err
			}

		default:
			verbosef("[%d]: formula arg identified, %q <= %q", i, arg, rawArg)
			argsOut = append(argsOut, arg)
		}
	}

	return argsOut, false, nil
}

// combineArgs processes a slice of strings and returns a new slice that is alternating <value> and <op>.
// This allows users to provide a date and time as multiple CLI args (i.e. without quoting a value).
func combineArgs(argsIn []string) *calcArgs {
	rv := &calcArgs{}
	if len(argsIn) == 0 {
		return rv
	}

	pipeAt := -1
	lastWasVal := false
	for _, arg := range argsIn {
		switch {
		case isPipeInd(arg):
			rv.HavePipe = true
			rv.All = append(rv.All, arg)
			pipeAt = len(rv.All) - 1
			lastWasVal = false
		case IsOp(arg):
			rv.All = append(rv.All, arg)
			lastWasVal = false
		case lastWasVal:
			rv.All[len(rv.All)-1] += " " + arg
		default:
			rv.All = append(rv.All, arg)
			lastWasVal = true
		}
	}

	if rv.HavePipe {
		if pipeAt > 0 {
			rv.PrePipe = rv.All[:pipeAt]
		}
		if pipeAt+1 < len(rv.All) {
			rv.PostPipe = rv.All[pipeAt+1:]
		}
	}

	return rv
}

// isPipeInd returns true if the provided arg is an indicator to use piped in data.
func isPipeInd(arg string) bool {
	return EqualFoldOneOf(arg, "-p", "--pipe")
}

// mainE actually runs this program, printing to the provided writer (e.g. os.Stdout) or returning an error as appropriate.
func mainE(argsIn []string, stdout io.Writer, stdin io.Reader) error {
	if len(argsIn) == 0 {
		if stdin != nil {
			// If we have an stdin, and no other args were provided, we get everything from the pipe.
			argsIn = []string{"--pipe"}
		} else {
			// If don't have stdin, and no args were provided, print help.
			argsIn = []string{"--help"}
		}
	}

	args, stopNow, err := getArgs(argsIn, stdout)
	if stopNow || err != nil {
		return err
	}

	if Verbose {
		stderrPrintf("Input date/time formats (%d):", len(FormatParseOrder))
		for i, nf := range FormatParseOrder {
			stderrPrintf("[%2d]: %s", i, nf)
		}
	}

	var result *DTVal
	if !args.HavePipe {
		result, err = DoCalculation(args.All)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, result.FormattedString())
		return nil
	}

	// There's piped in lines. For each line, put it together and run the calc.
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := scanner.Text()
		pipeArgs := combineArgs(strings.Fields(line))
		formula := make([]string, 0, len(args.PrePipe)+len(pipeArgs.All)+len(args.PostPipe))
		formula = append(formula, args.PrePipe...)
		formula = append(formula, pipeArgs.All...)
		formula = append(formula, args.PostPipe...)

		result, err = DoCalculation(formula)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, result.FormattedString())
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
}

// isCharDev returns true if the provided file is a character device.
// This essentially returns true if there's stuff being piped in.
func isCharDev(stdin *os.File) bool {
	stat, err := stdin.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) == 0
}

// main is the program's entry point.
func main() {
	if val, ok := os.LookupEnv("VERBOSE"); ok {
		Verbose, _ = strconv.ParseBool(val)
		verbosef("verbose environment variable detected")
	}
	var stdin io.Reader
	if isCharDev(os.Stdin) {
		stdin = os.Stdin
	}
	if err := mainE(os.Args[1:], os.Stdout, stdin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
