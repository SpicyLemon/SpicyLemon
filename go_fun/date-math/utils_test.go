package main_test

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

// AssertEqualError checks that the expErr matches theErr.
func AssertEqualError(t *testing.T, expErr string, theErr error, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(expErr) == 0 {
		return assert.NoError(t, theErr, msgAndArgs...)
	}
	return assert.EqualError(t, theErr, expErr, msgAndArgs...)
}

const assertTimeFormat = "2006-01-02 15:04:05.999999999 -0700"

func AssertEqualTime(t *testing.T, expected, actual time.Time, msgAndArgs ...interface{}) bool {
	t.Helper()
	// If compared as structs, the locations always cause a failure.
	// So we do the comparison using a formatted string that contains all the important bits.
	expDT := expected.Format(assertTimeFormat)
	actDT := actual.Format(assertTimeFormat)
	return assert.Equal(t, expDT, actDT, msgAndArgs...)
}

// SuppressStderrFn switches os.Stderr to a pipe and returns a function that will put it back and close the pipe.
// Standard usage: defer SuppressStderrFn()()
func SuppressStderrFn() func() {
	stderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w
	return func() {
		os.Stderr = stderr
		w.Close()
	}
}

// ResetGlobalsFn returns a function that will return the global variables back to their current values.
// Standard usage: defer ResetGlobalsFn()()
func ResetGlobalsFn() func() {
	origFormatParseOrder := copySlice(FormatParseOrder)
	origOutputFormat := OutputFormat
	origInputFormat := InputFormat
	origNamedFormatMap := copyMap(NamedFormatMap)
	origInputFormatsUsed := copySlice(UsedInputFormats)
	origVerbose := Verbose
	origCurStep := CurStep
	return func() {
		FormatParseOrder = origFormatParseOrder
		OutputFormat = origOutputFormat
		InputFormat = origInputFormat
		NamedFormatMap = origNamedFormatMap
		UsedInputFormats = origInputFormatsUsed
		Verbose = origVerbose
		CurStep = origCurStep
	}
}

// LogGlobals logs all of the global variables that ResetGlobalsFn manages.
func LogGlobals(t *testing.T) {
	// Do not delete this function just because it's not being called.
	// It's handy to add to a unit test when things go weird, but might not keep the call once done.
	parseOrder := make([]string, len(FormatParseOrder))
	for i, format := range FormatParseOrder {
		parseOrder[i] = fmt.Sprintf(" [%d]: %s", i, format)
	}
	t.Logf("FormatParseOrder (%d):\n%s", len(FormatParseOrder), strings.Join(parseOrder, "\n"))

	t.Logf("OutputFormat: %q", OutputFormat)

	t.Logf("InputFormat: %s", InputFormat)

	namedFormats := make([]string, len(NamedFormatMap))
	for i, name := range slices.Sorted(maps.Keys(NamedFormatMap)) {
		namedFormats[i] = fmt.Sprintf(" [%d]: %q => %s", i, name, NamedFormatMap[name])
	}
	t.Logf("NamedFormatMap (%d):\n%s", len(NamedFormatMap), strings.Join(namedFormats, "\n"))

	// UsedInputFormats
	inputFmts := make([]string, len(UsedInputFormats))
	for i, format := range UsedInputFormats {
		inputFmts[i] = fmt.Sprintf(" [%d]: %s", i, format)
	}
	t.Logf("UsedInputFormats (%d):\n%s", len(UsedInputFormats), strings.Join(inputFmts, "\n"))

	t.Logf("Verbose: %t", Verbose)
	t.Logf("CurStep: %d", CurStep)
}

// copySlice returns a shallow copy of a slice.
func copySlice[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	rv := make(S, len(s))
	for i, e := range s {
		rv[i] = e
	}
	return rv
}

// copyMap returns a shallow copy of a map.
func copyMap[M ~map[K]V, K comparable, V any](m M) M {
	if m == nil {
		return nil
	}
	rv := make(M, len(m))
	for k, v := range m {
		rv[k] = v
	}
	return rv
}

func TestSetOutputFormatByName(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		expErr string
		expFmt string
	}{
		{
			name:   "empty arg",
			arg:    "",
			expErr: "unknown output format name \"\"",
		},
		{
			name:   "unknown name",
			arg:    "not known",
			expErr: "unknown output format name \"not known\"",
		},
		{
			name:   "UnixDate",
			arg:    "UnixDate",
			expFmt: DtFmtUnixDate.Format,
		},
		{
			name:   "lowercase unixdate",
			arg:    "unixdate",
			expFmt: DtFmtUnixDate.Format,
		},
		{
			name:   "RFC1123",
			arg:    "RFC1123",
			expFmt: DtFmtRFC1123.Format,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			OutputFormat = ""

			var w bytes.Buffer
			var err error
			testFunc := func() {
				err = SetOutputFormatByName(tc.arg, &w)
			}
			require.NotPanics(t, testFunc, "setOutputFormatByName(%q)", tc.name)
			printed := w.String()
			AssertEqualError(t, tc.expErr, err, "setOutputFormatByName(%q) error", tc.name)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "setOutputFormatByName(%q) error", tc.name)
				assert.NotEmpty(t, printed, "things printed during setOutputFormatByName")
			} else {
				assert.NoError(t, err, "setOutputFormatByName(%q) error", tc.name)
				assert.Empty(t, printed, "things printed during setOutputFormatByName")
			}
			assert.Equal(t, tc.expFmt, OutputFormat, "OutputFormat global variable")

			if t.Failed() {
				t.Logf("printed:\n%s", printed)
			}
		})
	}
}

func TestSetOutputFormatByValue(t *testing.T) {
	LogGlobals(t)

	tests := []struct {
		name   string
		format string
		expErr string
	}{
		{
			name:   "empty",
			format: "",
			expErr: "empty output format string not allowed",
		},
		{
			name:   "only spaces",
			format: "   \t  ",
			expErr: "empty output format string not allowed",
		},
		{
			name:   "name of format",
			format: DtFmtRFC3339.Name,
			expErr: "output format string \"" + DtFmtRFC3339.Name + "\" cannot be a named format (did you mean to use --output-name instead)",
		},
		{
			name:   "just a year",
			format: "2006",
		},
		{
			name:   "full format",
			format: "02/01/06 04:05 after 03 -0700",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			Verbose = false
			OutputFormat = ""

			expOutputFormat := StrIf(len(tc.expErr) == 0, tc.format)

			var err error
			testFunc := func() {
				err = SetOutputFormatByValue(tc.format)
			}
			require.NotPanics(t, testFunc, "setOutputFormatByValue(%q)", tc.format)
			AssertEqualError(t, tc.expErr, err, "setOutputFormatByValue(%q) error", tc.format)
			assert.Equal(t, expOutputFormat, OutputFormat, "OutputFormat global variable")
		})
	}
}

func TestSetInputFormatByName(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		expErr string
		expFmt *NamedFormat
	}{
		{
			name:   "empty arg",
			arg:    "",
			expErr: "unknown input format name \"\"",
		},
		{
			name:   "unknown name",
			arg:    "not known",
			expErr: "unknown input format name \"not known\"",
		},
		{
			name:   "UnixDate",
			arg:    "UnixDate",
			expFmt: DtFmtUnixDate,
		},
		{
			name:   "lowercase unixdate",
			arg:    "unixdate",
			expFmt: DtFmtUnixDate,
		},
		{
			name:   "RFC1123",
			arg:    "RFC1123",
			expFmt: DtFmtRFC1123,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			InputFormat = nil
			FormatParseOrder = nil
			var expPO []*NamedFormat
			if tc.expFmt != nil {
				expPO = append(expPO, tc.expFmt)
			}

			var w bytes.Buffer
			var err error
			testFunc := func() {
				err = SetInputFormatByName(tc.arg, &w)
			}
			require.NotPanics(t, testFunc, "setInputFormatByName(%q)", tc.name)
			printed := w.String()
			AssertEqualError(t, tc.expErr, err, "setInputFormatByName(%q) error", tc.name)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "setInputFormatByName(%q) error", tc.name)
				assert.NotEmpty(t, printed, "things printed during setInputFormatByName")
			} else {
				assert.NoError(t, err, "setInputFormatByName(%q) error", tc.name)
				assert.Empty(t, printed, "things printed during setInputFormatByName")
			}
			assert.Equal(t, tc.expFmt, InputFormat, "InputFormat global variable")
			assert.Equal(t, expPO, FormatParseOrder, "FormatParseOrder global variable")

			if t.Failed() {
				t.Logf("printed:\n%s", printed)
			}
		})
	}
}

func TestSetInputFormatByValue(t *testing.T) {
	LogGlobals(t)

	tests := []struct {
		name   string
		format string
		expErr string
	}{
		{
			name:   "empty",
			format: "",
			expErr: "empty input format string not allowed",
		},
		{
			name:   "only spaces",
			format: "   \t  ",
			expErr: "empty input format string not allowed",
		},
		{
			name:   "name of format",
			format: DtFmtRFC3339.Name,
			expErr: "input format string \"" + DtFmtRFC3339.Name + "\" cannot be a named format (did you mean to use --input-name instead)",
		},
		{
			name:   "just a year",
			format: "2006",
		},
		{
			name:   "full format",
			format: "02/01/06 04:05 after 03 -0700",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			Verbose = false
			InputFormat = nil
			FormatParseOrder = nil

			var expInputFormat *NamedFormat
			var expPO []*NamedFormat
			if len(tc.expErr) == 0 {
				expInputFormat = MakeNamedFormat("User", tc.format)
				expPO = append(expPO, expInputFormat)
			}

			var err error
			testFunc := func() {
				err = SetInputFormatByValue(tc.format)
			}
			require.NotPanics(t, testFunc, "setInputFormatByValue(%q)", tc.format)
			AssertEqualError(t, tc.expErr, err, "setInputFormatByValue(%q) error", tc.format)
			assert.Equal(t, expInputFormat, InputFormat, "InputFormat global variable")
			assert.Equal(t, expPO, FormatParseOrder, "FormatParseOrder global variable")
		})
	}
}

func TestHasOneOf(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		options []string
		exp     bool
	}{
		{
			name:    "empty string, no options",
			val:     "",
			options: nil,
			exp:     false,
		},
		{
			name:    "empty string, one empty option",
			val:     "",
			options: []string{""},
			exp:     true,
		},
		{
			name:    "empty string, one non-empty option",
			val:     "",
			options: []string{"x"},
			exp:     false,
		},
		{
			name:    "no options",
			val:     "this is a value",
			options: nil,
			exp:     false,
		},
		{
			name:    "one option at start of val",
			val:     "this is a value",
			options: []string{"this"},
			exp:     true,
		},
		{
			name:    "one option in middle of val",
			val:     "this is a value",
			options: []string{"is a"},
			exp:     true,
		},
		{
			name:    "one option at end of val",
			val:     "this is a value",
			options: []string{"value"},
			exp:     true,
		},
		{
			name:    "one option is whole val",
			val:     "this is a value",
			options: []string{"this is a value"},
			exp:     true,
		},
		{
			name:    "one option not in val",
			val:     "this is a value",
			options: []string{"sis"},
			exp:     false,
		},
		{
			name:    "one empty option",
			val:     "this is a value",
			options: []string{""},
			exp:     true,
		},
		{
			name:    "three options, first in val",
			val:     "1234567890",
			options: []string{"23", "32", "79"},
			exp:     true,
		},
		{
			name:    "three options, second in val",
			val:     "1234567890",
			options: []string{"32", "90", "79"},
			exp:     true,
		},
		{
			name:    "three options, third in val",
			val:     "1234567890",
			options: []string{"32", "79", "456"},
			exp:     true,
		},
		{
			name:    "three options, none in val",
			val:     "1234567890",
			options: []string{"32", "79", "42"},
			exp:     false,
		},
		{
			name:    "three options, has first and second",
			val:     "1234567890",
			options: []string{"12", "23", "zero"},
			exp:     true,
		},
		{
			name:    "three options, has first and third",
			val:     "1234567890",
			options: []string{"12", "zero", "678"},
			exp:     true,
		},
		{
			name:    "three options, has second and third",
			val:     "1234567890",
			options: []string{"zero", "456", "678"},
			exp:     true,
		},
		{
			name:    "three options, has all",
			val:     "1234567890",
			options: []string{"0", "456", "678"},
			exp:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = HasOneOf(tc.val, tc.options...)
			}
			require.NotPanics(t, testFunc, "HasOneOf(%q, %q)", tc.val, tc.options)
			assert.Equal(t, tc.exp, act, "HasOneOf(%q, %q) result", tc.val, tc.options)
		})
	}
}

func TestHasAllOf(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		options []string
		exp     bool
	}{
		{
			name:    "empty string, no options",
			val:     "",
			options: nil,
			exp:     true,
		},
		{
			name:    "empty string, one empty option",
			val:     "",
			options: []string{""},
			exp:     true,
		},
		{
			name:    "empty string, one non-empty option",
			val:     "",
			options: []string{"x"},
			exp:     false,
		},
		{
			name:    "no options",
			val:     "this is a value",
			options: nil,
			exp:     true,
		},
		{
			name:    "one option at start of val",
			val:     "this is a value",
			options: []string{"this"},
			exp:     true,
		},
		{
			name:    "one option in middle of val",
			val:     "this is a value",
			options: []string{"is a"},
			exp:     true,
		},
		{
			name:    "one option at end of val",
			val:     "this is a value",
			options: []string{"value"},
			exp:     true,
		},
		{
			name:    "one option is whole val",
			val:     "this is a value",
			options: []string{"this is a value"},
			exp:     true,
		},
		{
			name:    "one option not in val",
			val:     "this is a value",
			options: []string{"sis"},
			exp:     false,
		},
		{
			name:    "one empty option",
			val:     "this is a value",
			options: []string{""},
			exp:     true,
		},
		{
			name:    "three options, first in val",
			val:     "1234567890",
			options: []string{"23", "32", "79"},
			exp:     false,
		},
		{
			name:    "three options, second in val",
			val:     "1234567890",
			options: []string{"32", "90", "79"},
			exp:     false,
		},
		{
			name:    "three options, third in val",
			val:     "1234567890",
			options: []string{"32", "79", "456"},
			exp:     false,
		},
		{
			name:    "three options, none in val",
			val:     "1234567890",
			options: []string{"32", "79", "42"},
			exp:     false,
		},
		{
			name:    "three options, has first and second",
			val:     "1234567890",
			options: []string{"12", "23", "zero"},
			exp:     false,
		},
		{
			name:    "three options, has first and third",
			val:     "1234567890",
			options: []string{"12", "zero", "678"},
			exp:     false,
		},
		{
			name:    "three options, has second and third",
			val:     "1234567890",
			options: []string{"zero", "456", "678"},
			exp:     false,
		},
		{
			name:    "three options, has all",
			val:     "1234567890",
			options: []string{"0", "456", "678"},
			exp:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = HasAllOf(tc.val, tc.options...)
			}
			require.NotPanics(t, testFunc, "HasAllOf(%q, %q)", tc.val, tc.options)
			assert.Equal(t, tc.exp, act, "HasAllOf(%q, %q) result", tc.val, tc.options)
		})
	}
}

func TestStrIf(t *testing.T) {
	tests := []struct {
		name   string
		val    bool
		ifTrue string
		exp    string
	}{
		{
			name:   "true, empty",
			val:    true,
			ifTrue: "",
			exp:    "",
		},
		{
			name:   "true, not empty",
			val:    true,
			ifTrue: "something",
			exp:    "something",
		},
		{
			name:   "false, empty",
			val:    false,
			ifTrue: "",
			exp:    "",
		},
		{
			name:   "false, not empty",
			val:    false,
			ifTrue: "something",
			exp:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = StrIf(tc.val, tc.ifTrue)
			}
			require.NotPanics(t, testFunc, "StrIf(%t, %q)", tc.val, tc.ifTrue)
			assert.Equal(t, tc.exp, act, "StrIf(%t, %q)", tc.val, tc.ifTrue)
		})
	}
}

func TestEqualFoldOneOf(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		options []string
		exp     bool
	}{
		{
			name:    "empty arg, no options",
			arg:     "",
			options: nil,
			exp:     false,
		},
		{
			name:    "empty arg, one empty option",
			arg:     "",
			options: []string{""},
			exp:     true,
		},
		{
			name:    "empty arg, one non-empty option",
			arg:     "",
			options: []string{"x"},
			exp:     false,
		},
		{
			name:    "no options",
			arg:     "This is a LITTLE sentEnce.",
			options: nil,
			exp:     false,
		},
		{
			name:    "one option: equals arg",
			arg:     "This is a LITTLE sentEnce.",
			options: []string{"This is a LITTLE sentEnce."},
			exp:     true,
		},
		{
			name:    "one option: contains arg",
			arg:     "This is a LITTLE sentEnce.",
			options: []string{"This is a LITTLE sentEnce"},
			exp:     false,
		},
		{
			name:    "one option: totally different",
			arg:     "This is a LITTLE sentEnce.",
			options: []string{"And now for something completely different."},
			exp:     false,
		},
		{
			name:    "one option: alternately cased",
			arg:     "This is a LITTLE sentEnce.",
			options: []string{"tHIS IS A little SENTeNCE."},
			exp:     true,
		},
		{
			name:    "three options: matches first",
			arg:     "Bananas",
			options: []string{"bananas", "oranges", "apples"},
			exp:     true,
		},
		{
			name:    "three options: matches second",
			arg:     "Bananas",
			options: []string{"oranges", "bAnAnAs", "apples"},
			exp:     true,
		},
		{
			name:    "three options: matches third",
			arg:     "Bananas",
			options: []string{"oranges", "apples", "baNaNas"},
			exp:     true,
		},
		{
			name:    "three options: matches none",
			arg:     "Bananas",
			options: []string{"oranges", "apples", " Bananas "},
			exp:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = EqualFoldOneOf(tc.arg, tc.options...)
			}
			require.NotPanics(t, testFunc, "EqualFoldOneOf(%q, %q)", tc.arg, tc.options)
			assert.Equal(t, tc.exp, act, "EqualFoldOneOf(%q, %q)", tc.arg, tc.options)
		})
	}
}
