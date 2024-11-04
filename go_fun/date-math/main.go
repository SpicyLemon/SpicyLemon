package main

import (
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
  <num> A possibly signed whole number. Limited to -1,000,000 and 1,000,000
        (inclusive), otherwise it's treated as an epoch.

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
  --verbose|-v
        Print debugging information to stderr.
        Can also be enabled by setting the VERBOSE env var.
  --help|-h
        Output this message.`)
}

// getArgs handles flags and options and returns the combined formula args and whether to stop early.
// If help or formats are requested, this will print that to the provided writer (e.g. os.Stdout).
func getArgs(argsIn []string, stdout io.Writer) ([]string, bool, error) {
	fArgs, stop, err := processFlags(argsIn, stdout)
	if stop || err != nil {
		return nil, stop, err
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
func combineArgs(argsIn []string) []string {
	if len(argsIn) == 0 {
		return nil
	}
	argsOut := make([]string, 0, len(argsIn))
	var i int
	if !IsOp(argsIn[0]) {
		var arg0 string
		arg0, i = getNextValueArg(argsIn)
		argsOut = append(argsOut, arg0)
	}
	for i < len(argsIn) {
		argsOut = append(argsOut, argsIn[i])
		i++
		if i >= len(argsIn) {
			break
		}
		valArg, n := getNextValueArg(argsIn[i:])
		argsOut = append(argsOut, valArg)
		i += n
	}

	verbosef("combined args: %q", argsOut)
	return argsOut
}

// getNextValueArg returns a string with the next value arg in it and the number of args entries it spans.
func getNextValueArg(args []string) (string, int) {
	for i, arg := range args {
		if IsOp(arg) {
			// E.g. args = [a b c +], i = 3 at +, so args[:i] = [a b c] and we're using 3 entries.
			return strings.Join(args[:i], " "), i
		}
	}
	return strings.Join(args, " "), len(args)
}

// mainE actually runs this program, printing to the provided writer (e.g. os.Stdout) or returning an error as appropriate.
func mainE(argsIn []string, stdout io.Writer) error {
	if len(argsIn) == 0 {
		argsIn = []string{"--help"}
	}
	formula, stopNow, err := getArgs(argsIn, stdout)
	if stopNow || err != nil {
		return err
	}

	if Verbose {
		stderrPrintf("Input date/time formats (%d):", len(FormatParseOrder))
		for i, nf := range FormatParseOrder {
			stderrPrintf("[%2d]: %s", i, nf)
		}
	}

	result, err := DoCalculation(formula)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, result.FormattedString())
	return nil
}

// main is the program's entry point.
func main() {
	if val, ok := os.LookupEnv("VERBOSE"); ok {
		Verbose, _ = strconv.ParseBool(val)
		verbosef("verbose environment variable detected")
	}
	if err := mainE(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
