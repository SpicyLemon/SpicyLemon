package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
)

// PrintUsage outputs a multi-line string with info on how to run this program.
func PrintUsage(stdout io.Writer) {
	fmt.Fprintf(stdout, `big-sum: Add a bunch of numbers together with nearly infinite precision.

Usage: big-sum <number 1> [<number 2> ...] [--pipe|-] [--pretty|-p] [--verbose|-v]
  or : <stuff> | big-sum

The --pipe or - flag is implied if there are no arguments provided.
The --pretty or -p flag will add commas to the result.

Warning: Floating point numbers may result in unwanted rounding.'
`)
}

// Sum will parse each arg as a number and return a sum of all those numbers as a converted to a string.
func Sum(args []string) (string, error) {
	var totalInt *big.Int
	var totalFloat *big.Float
	wholeDigits := 0
	fractionalDigits := 0

	prec := calculatePrec(args)

	for _, arg := range args {
		orig := arg
		// Remove all commas so that people can provide numbers with the commas in them.
		arg = strings.ReplaceAll(arg, ",", "")

		if equalFoldOneOf(arg, "", "0", "0.0", ".0", "0.") {
			verbosef("Ignoring empty or zero arg: %q.", orig)
			continue
		}

		if !strings.Contains(arg, ".") {
			// Number doesn't have a ".", Parse it as an integer.
			// Note that by using 0 as the base, the arg can also handle underscore separators, e.g. 123_456!
			val, ok := new(big.Int).SetString(arg, 0)
			if !ok {
				return "", fmt.Errorf("could not parse %q as integer", orig)
			}
			if totalInt == nil {
				verbosef("Ints:   %40s  %q", val, arg)
				totalInt = val
			} else {
				verbosef("Ints: + %40s  %q", val, arg)
				totalInt.Add(totalInt, val)
				verbosef("Ints: = %40s", totalInt)
			}
		} else {
			// Number has a ".", parse it as a float.
			argWholeLen, argFractLen := countDigits(arg)
			if argWholeLen > wholeDigits {
				wholeDigits = argWholeLen
			}
			if argFractLen > fractionalDigits {
				fractionalDigits = argFractLen
			}

			val, _, err := big.ParseFloat(arg, 0, prec, big.ToNearestEven)
			if err != nil {
				return "", fmt.Errorf("could not parse %q as float: %w", orig, err)
			}
			if totalFloat == nil {
				verbosef("Floats:   %40s  from %q (prec=%d, acc=%s) (%d,%d)",
					val.Text('f', argFractLen)+strings.Repeat(" ", fractionalDigits-argFractLen), arg,
					val.Prec(), val.Acc(), argWholeLen, argFractLen)
				totalFloat = val
			} else {
				verbosef("Floats: + %40s  from %q (prec=%d, acc=%s) (%d,%d)",
					val.Text('f', argFractLen)+strings.Repeat(" ", fractionalDigits-argFractLen), arg,
					val.Prec(), val.Acc(), argWholeLen, argFractLen)
				totalFloat = new(big.Float).SetPrec(prec).Add(totalFloat, val)
				verbosef("Floats: = %40s (prec=%d, acc=%s) (%d,%d)",
					totalFloat.Text('f', fractionalDigits), totalFloat.Prec(),
					totalFloat.Acc(), wholeDigits, fractionalDigits)
			}
		}
	}

	if totalFloat != nil {
		if totalInt != nil {
			sumInts := totalInt.String()
			if len(sumInts) > wholeDigits {
				wholeDigits = len(sumInts)
			}
			verbosef("Sum Ints:     %40s", sumInts+strings.Repeat(" ", fractionalDigits+1))
			verbosef("Sum Floats: + %40s", totalFloat.Text('f', fractionalDigits))
			totalFloat = new(big.Float).SetPrec(precForLen(wholeDigits+fractionalDigits+1)).Add(totalFloat, new(big.Float).SetInt(totalInt))
			verbosef("Grand Sum:  = %40s", totalFloat.Text('f', fractionalDigits))
		}
		return totalFloat.Text('f', fractionalDigits), nil
	}
	if totalInt != nil {
		return totalInt.String(), nil
	}
	return "0", nil
}

// calculatePrec will calculate the precision needed to add all the floats together accurately.
func calculatePrec(args []string) uint {
	var wMax, fMax int
	for _, arg := range args {
		w, f := countDigits(arg)
		if w > wMax {
			wMax = w
		}
		if f > fMax {
			fMax = f
		}
	}
	// Add one to the length for extra growth room.
	// Add one to the length for each 3 args too for the same reason.
	return precForLen(wMax + fMax + 1 + len(args)/3)
}

// countDigits returns the number of whole and fractional digits in the provided number string.
func countDigits(arg string) (whole int, fractional int) {
	parts := strings.Split(arg, ".")
	if len(parts) > 0 {
		whole = len(parts[0])
		if strings.HasPrefix(parts[0], "-") {
			whole--
		}
	}
	if len(parts) > 1 {
		fractional = len(parts[1])
	}
	return
}

// precForLen returns a safe precision that can be used to represent the provided number of digits.
func precForLen(digits int) uint {
	// Through trial and error, it seems like precision should go up 7 for each digit provided.
	// Once I got to 7, all my unit tests finally passed. Before that, there were rounding errors
	// affecting up to 5 digits. Also, I'm erring on the side of too much precision since I'm
	// pretty sure more precision means it's more likely to get the correct answer.
	return uint(digits * 7)
}

// mainE is the actual runner of this program, possibly returning an error.
func mainE(argsIn []string, stdout io.Writer, stdin io.Reader) error {
	args, stopNow, err := processFlags(argsIn, stdout, stdin)
	if stopNow || err != nil {
		return err
	}

	answer, err := Sum(args.Values)
	if err != nil {
		return err
	}
	if args.Pretty {
		answer = MakeNumberPretty(answer)
	}
	fmt.Fprintln(stdout, answer)
	return nil
}

// sumParams are the parameters defined by command-line arguments on how to behave and execute.
type sumParams struct {
	Values []string
	Pretty bool
}

// processFlags will handle all the flags in the provided args. It will also read stdin if called for.
func processFlags(argsIn []string, stdout io.Writer, stdin io.Reader) (*sumParams, bool, error) {
	rv := &sumParams{}
	verbosef("Args provided (%d):", len(argsIn))
	for i := 0; i < len(argsIn); i++ {
		rawArg := argsIn[i]
		arg := strings.TrimSpace(rawArg)
		switch {
		case equalFoldOneOf(arg, "--help", "-h", "help"):
			verbosef("[%d]: help arg identified, %q", i, rawArg)
			PrintUsage(stdout)
			return nil, true, nil
		case equalFoldOneOf(arg, "--pretty", "-p"):
			verbosef("[%d]: pretty arg identified, %q", i, rawArg)
			rv.Pretty = true
		case equalFoldOneOf(arg, "--verbose", "-v"):
			Verbose = true
			verbosef("[%d]: verbose flag identified, %q", i, rawArg)
		case equalFoldOneOf(arg, "--pipe", "-p", "-"):
			verbosef("[%d]: pipe flag identified, %q", i, rawArg)
			newArgs, err := readStdin(stdin)
			if err != nil {
				return nil, true, err
			}
			stdin = nil
			rv.Values = append(rv.Values, newArgs...)
		default:
			verbosef("[%d]: number identified, %q", i, rawArg)
			rv.Values = append(rv.Values, strings.Fields(arg)...)
		}
	}

	if len(rv.Values) == 0 {
		if stdin != nil {
			// If we have stdin, and no other args were provided, we get everything from the pipe.
			verbosef("no args provided, using pipe.")
			newArgs, err := readStdin(stdin)
			if err != nil {
				return nil, true, err
			}
			rv.Values = append(rv.Values, newArgs...)
		} else {
			// If don't have stdin, and no args were provided, print help.
			verbosef("no args provided, and no pipe either.")
			PrintUsage(stdout)
			return nil, true, nil
		}
	}

	return rv, false, nil
}

// readStdin reads all possible info from sdtdin.
func readStdin(stdin io.Reader) ([]string, error) {
	if stdin == nil {
		return nil, errors.New("no stdin available")
	}

	var rv []string
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := scanner.Text()
		rv = append(rv, strings.Fields(line)...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from stdin: %w", err)
	}

	return rv, nil
}

// equalFoldOneOf returns true of one of the provided options is equal to the arg (ignoring case).
func equalFoldOneOf(arg string, options ...string) bool {
	for _, opt := range options {
		if strings.EqualFold(arg, opt) {
			return true
		}
	}
	return false
}

// MakeNumberPretty takes in a number string and adds commas to the whole part.
// Examples: "1234567" -> "1,234,567", "12345.678901" -> "12,345.678901"
// If the string already has commas, or has more than one period, the provided value is returned unchanged.
func MakeNumberPretty(val string) string {
	if len(val) <= 3 || strings.Contains(val, ",") {
		return val
	}
	parts := strings.Split(val, ".")
	if len(parts) == 0 || len(parts) > 2 {
		return val
	}

	wholePart := parts[0]
	hasNeg := len(wholePart) > 0 && wholePart[0] == '-'
	if hasNeg {
		wholePart = wholePart[1:]
	}

	if len(wholePart) > 3 {
		lenLhs := len(wholePart)
		lhs := make([]rune, 0, lenLhs+(lenLhs-1)/3+1)
		if hasNeg {
			lhs = append(lhs, '-')
		}
		for i, digit := range wholePart {
			if i > 0 && (lenLhs-i)%3 == 0 {
				lhs = append(lhs, ',')
			}
			lhs = append(lhs, digit)
		}
		parts[0] = string(lhs)
	}

	return strings.Join(parts, ".")
}

// Verbose keeps track of whether verbose output is enabled.
var Verbose bool

// verbosef prints the provided message to stderr if verbose output is enabled. If not enabled, this is a no-op.
func verbosef(format string, args ...interface{}) {
	if Verbose {
		stderrPrintf(format, args...)
	}
}

// stderrPrintf prints the provided stuff to stderr.
func stderrPrintf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// isCharDev returns true if the provided file is a character device.
// This essentially returns true if there's stuff being piped in.
func isCharDev(stdin *os.File) bool {
	stat, err := stdin.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) == 0
}

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
