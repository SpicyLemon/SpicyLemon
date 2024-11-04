package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// Verbose keeps track of whether verbose output is enabled.
	Verbose bool
	// CurStep is a counter that keeps track of how many operations have been processed.
	CurStep int
)

const (
	NilStr   = "<nil>"
	EmptyStr = "<empty>"
)

// stepName defines various parts of a step (used for verbose output).
type stepName string

const (
	stepOp     stepName = "op"
	stepFormat stepName = "format"
	stepValue  stepName = "value"
	stepResult stepName = "result"
)

// verbosef prints the provided message to stderr if verbose output is enabled. If not enabled, this is a no-op.
func verbosef(format string, args ...interface{}) {
	if Verbose {
		stderrPrintf(format, args...)
	}
}

// verboseStepf prints a step message to stderr if verbose output is enabled. If not enabled, this is a no-op.
func verboseStepf(name stepName, format string, args ...interface{}) {
	if Verbose {
		stderrPrintf("step %d %s: "+format, append([]interface{}{CurStep, name}, args...)...)
	}
}

// stderrPrintf prints the provided stuff to stderr.
func stderrPrintf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// setOutputFormatByName sets the OutputFormat variable based on the provided argument.
// If it's not known, the available formats are printed to the provided writer (e.g. os.Stdout) and an error is returned.
func setOutputFormatByName(arg string, stdout io.Writer) error {
	nf := GetFormatByName(arg)
	if nf == nil {
		PrintFormats(stdout)
		return fmt.Errorf("unknown output format name %q", arg)
	}
	OutputFormat = nf.Format
	verbosef("output format set by name %q: %q", arg, OutputFormat)
	return nil
}

// setOutputFormatByValue sets the OutputFormat variable to the one provided, making sure it's not a name.
func setOutputFormatByValue(format string) error {
	if len(strings.TrimSpace(format)) == 0 {
		return fmt.Errorf("empty output format string not allowed")
	}
	if nf := GetFormatByName(format); nf != nil {
		return fmt.Errorf("output format string %q cannot be a named format (did you mean to use --output-name instead)", format)
	}
	OutputFormat = format
	verbosef("output format set as provided: %q", OutputFormat)
	return nil
}

// setInputFormatByName sets the FormatParseOrder variable base on the provided argument.
// If it's not known, the available formats are printed to the provided writer (e.g. os.Stdout) and an error is returned.
func setInputFormatByName(arg string, stdout io.Writer) error {
	nf := GetFormatByName(arg)
	if nf == nil {
		PrintFormats(stdout)
		return fmt.Errorf("unknown input format name %q", arg)
	}
	InputFormat = nf
	FormatParseOrder = []*NamedFormat{InputFormat}
	verbosef("input format set by name %q: %s", arg, nf)
	return nil
}

// setInputFormatByValue sets the FormatParseOrder variable to the provided format, making sure it's not a name.
func setInputFormatByValue(format string) error {
	if len(strings.TrimSpace(format)) == 0 {
		return fmt.Errorf("empty input format string not allowed")
	}
	if nf := GetFormatByName(format); nf != nil {
		return fmt.Errorf("input format string %q cannot be a named format (did you mean to use --input-name instead)", format)
	}
	InputFormat = makeNamedFormat("User", format)
	FormatParseOrder = []*NamedFormat{InputFormat}
	verbosef("input format set as provided: %s", InputFormat)
	return nil
}

// HasOneOf returns true of the provided val contains any of the provided options.
func HasOneOf(val string, options ...string) bool {
	for _, opt := range options {
		if strings.Contains(val, opt) {
			return true
		}
	}
	return false
}

// HasAllOf returns true of the provided val contains all of the provided options.
func HasAllOf(val string, options ...string) bool {
	for _, opt := range options {
		if !strings.Contains(val, opt) {
			return false
		}
	}
	return true
}

// StrIf returns ifTrue if val is true, otherwise returns an empty string
func StrIf(val bool, ifTrue string) string {
	if val {
		return ifTrue
	}
	return ""
}

// EqualFoldOneOf returns true of one of the provided options is equal to the arg (ignoring case).
func EqualFoldOneOf(arg string, options ...string) bool {
	for _, opt := range options {
		if strings.EqualFold(arg, opt) {
			return true
		}
	}
	return false
}
