package main_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

func TestPrintUsage(t *testing.T) {
	// I don't really want to require specific things of the usage other than it has the program name in it.
	var w bytes.Buffer
	testFunc := func() {
		PrintUsage(&w)
	}
	require.NotPanics(t, testFunc, "PrintUsage(w)")
	printed := w.String()
	assert.Contains(t, printed, "date-math", "usage message should contain \"date-math\"")
}

func TestGetArgs(t *testing.T) {
	tests := []struct {
		name       string
		argsIn     []string
		expArgs    *CalcArgs
		expBool    bool
		expErr     string
		expInPrint []string
	}{
		{
			name:       "nil args",
			argsIn:     nil,
			expArgs:    nil,
			expBool:    false,
			expErr:     "",
			expInPrint: nil,
		},
		{
			name:       "empty args",
			argsIn:     []string{},
			expArgs:    nil,
			expBool:    false,
			expErr:     "",
			expInPrint: nil,
		},
		{
			name:       "args with help",
			argsIn:     []string{"args", "with", "help"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:    "invalid flag",
			argsIn:  []string{"arg", "-f"},
			expBool: true,
			expErr:  "no argument provided after -f, expected a format string",
		},
		{
			name:    "formula with args to combine",
			argsIn:  []string{"2006-04-12", "17:04:55", "-", "2006-04-12", "17:03:12"},
			expArgs: &CalcArgs{All: []string{"2006-04-12 17:04:55", "-", "2006-04-12 17:03:12"}},
		},
		{
			name:    "flag in the middle of formula",
			argsIn:  []string{"23m", "x", "-v", "44"},
			expArgs: &CalcArgs{All: []string{"23m", "x", "44"}},
		},
		{
			name:    "just --pipe",
			argsIn:  []string{"--pipe"},
			expArgs: &CalcArgs{All: []string{"--pipe"}, HavePipe: true},
		},
		{
			name:   "formula with --pipe at start",
			argsIn: []string{"--pipe", "+", "15m"},
			expArgs: &CalcArgs{
				All:      []string{"--pipe", "+", "15m"},
				HavePipe: true,
				PrePipe:  nil,
				PostPipe: []string{"+", "15m"},
			},
		},
		{
			name:   "formula with --pipe at end",
			argsIn: []string{"15m", "+", "--pipe"},
			expArgs: &CalcArgs{
				All:      []string{"15m", "+", "--pipe"},
				HavePipe: true,
				PrePipe:  []string{"15m", "+"},
				PostPipe: nil,
			},
		},
		{
			name:   "formula with -p in middle",
			argsIn: []string{"15m", "+", "--pipe", "44"},
			expArgs: &CalcArgs{
				All:      []string{"15m", "+", "--pipe", "44"},
				HavePipe: true,
				PrePipe:  []string{"15m", "+"},
				PostPipe: []string{"44"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			defer SuppressStderrFn()()
			Verbose = false

			if tc.expArgs == nil {
				tc.expArgs = &CalcArgs{}
			}

			var w bytes.Buffer
			var actArgs *CalcArgs
			var actBool bool
			var err error
			testFunc := func() {
				actArgs, actBool, err = GetArgs(tc.argsIn, &w)
			}
			require.NotPanics(t, testFunc, "getArgs(%q, w)", tc.argsIn)
			printed := w.String()
			AssertEqualError(t, tc.expErr, err, "getArgs(%q, w) error", tc.argsIn)
			assert.Equal(t, tc.expArgs, actArgs, "getArgs(%q, w) args", tc.argsIn)
			assert.Equal(t, tc.expBool, actBool, "getArgs(%q, w) bool", tc.argsIn)
			for i, exp := range tc.expInPrint {
				assert.Contains(t, printed, exp, "[%d]: Printed text should have %q", i, exp)
			}
		})
	}
}

func TestProcessFlags(t *testing.T) {
	tests := []struct {
		name       string
		argsIn     []string
		expArgs    []string
		expBool    bool
		expErr     string
		expInPrint []string
		expV       bool
		expOutFmt  string
		expPO      []*NamedFormat // defaults to FormatParseOrder if nil.
	}{
		{
			name:    "nil args",
			argsIn:  nil,
			expArgs: nil,
			expBool: false,
			expErr:  "",
		},
		{
			name:    "empty",
			argsIn:  []string{},
			expArgs: nil,
			expBool: false,
			expErr:  "",
		},

		{
			name:       "--help",
			argsIn:     []string{"--help"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:       "-h",
			argsIn:     []string{"-h"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:       "help",
			argsIn:     []string{"help"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:       "Help",
			argsIn:     []string{"Help"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:       "HELP",
			argsIn:     []string{"HELP"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},
		{
			name:       "arg --help arg",
			argsIn:     []string{"arg1", "--help", "num2"},
			expBool:    true,
			expInPrint: []string{"date-math"},
		},

		{
			name:       "--formats",
			argsIn:     []string{"--formats"},
			expBool:    true,
			expInPrint: []string{"* = possible input format"},
		},
		{
			name:       "formats",
			argsIn:     []string{"formats"},
			expBool:    true,
			expInPrint: []string{"* = possible input format"},
		},
		{
			name:       "Formats",
			argsIn:     []string{"Formats"},
			expBool:    true,
			expInPrint: []string{"* = possible input format"},
		},
		{
			name:       "FORMATS",
			argsIn:     []string{"FORMATS"},
			expBool:    true,
			expInPrint: []string{"* = possible input format"},
		},
		{
			name:       "arg --formats arg",
			argsIn:     []string{"arg1", "--formats", "val2"},
			expBool:    true,
			expInPrint: []string{"* = possible input format"},
		},

		{
			name:    "--verbose",
			argsIn:  []string{"--verbose"},
			expArgs: nil,
			expBool: false,
			expV:    true,
		},
		{
			name:    "-v",
			argsIn:  []string{"-v"},
			expArgs: nil,
			expBool: false,
			expV:    true,
		},
		{
			name:    "arg --verbose arg",
			argsIn:  []string{"num1", "--verbose", "thing2"},
			expArgs: []string{"num1", "thing2"},
			expBool: false,
			expV:    true,
		},
		{
			name:    "arg -v arg",
			argsIn:  []string{"num1", "-v", "thing2"},
			expArgs: []string{"num1", "thing2"},
			expBool: false,
			expV:    true,
		},

		{
			name:      "--output-name without arg",
			argsIn:    []string{"--output-name"},
			expArgs:   nil,
			expBool:   true,
			expErr:    "no argument provided after --output-name, expected a format name",
			expOutFmt: "",
		},
		{
			name:      "-o without arg",
			argsIn:    []string{"-o"},
			expArgs:   nil,
			expBool:   true,
			expErr:    "no argument provided after -o, expected a format name",
			expOutFmt: "",
		},
		{
			name:      "--output-name with unknown name",
			argsIn:    []string{"--output-name", "nope"},
			expArgs:   nil,
			expBool:   true,
			expErr:    "unknown output format name \"nope\"",
			expOutFmt: "",
		},
		{
			name:      "-o with unknown name",
			argsIn:    []string{"-o", "weird"},
			expArgs:   nil,
			expBool:   true,
			expErr:    "unknown output format name \"weird\"",
			expOutFmt: "",
		},
		{
			name:      "--output-name with known name exact case",
			argsIn:    []string{"--output-name", DtFmtRFC850.Name},
			expArgs:   nil,
			expOutFmt: DtFmtRFC850.Format,
		},
		{
			name:      "-o with known name exact case",
			argsIn:    []string{"-o", DtFmtRFC850.Name},
			expArgs:   nil,
			expOutFmt: DtFmtRFC850.Format,
		},
		{
			name:      "--output-name with known name different case",
			argsIn:    []string{"--output-name", "unIxdAte"},
			expArgs:   nil,
			expOutFmt: DtFmtUnixDate.Format,
		},
		{
			name:      "-o with known name different case",
			argsIn:    []string{"-o", "unIxdAte"},
			expArgs:   nil,
			expOutFmt: DtFmtUnixDate.Format,
		},
		{
			name:      "arg --output-name name arg",
			argsIn:    []string{"thing1", "--output-name", "default", "stuff2"},
			expArgs:   []string{"thing1", "stuff2"},
			expOutFmt: DtFmtDefault.Format,
		},
		{
			name:      "arg -o name arg",
			argsIn:    []string{"stuff1", "-o", "datetime", "thing2"},
			expArgs:   []string{"stuff1", "thing2"},
			expOutFmt: DtFmtDateTime.Format,
		},

		{
			name:      "--output-format without arg",
			argsIn:    []string{"--output-format"},
			expBool:   true,
			expErr:    "no argument provided after --output-format, expected a format string",
			expOutFmt: "",
		},
		{
			name:      "-f without arg",
			argsIn:    []string{"-f"},
			expBool:   true,
			expErr:    "no argument provided after -f, expected a format string",
			expOutFmt: "",
		},
		{
			name:      "--output-format with name arg",
			argsIn:    []string{"--output-format", DtFmtRFC3339Nano.Name},
			expBool:   true,
			expErr:    "output format string \"" + DtFmtRFC3339Nano.Name + "\" cannot be a named format (did you mean to use --output-name instead)",
			expOutFmt: "",
		},
		{
			name:      "-f with name arg",
			argsIn:    []string{"-f", DtFmtANSIC.Name},
			expBool:   true,
			expErr:    "output format string \"" + DtFmtANSIC.Name + "\" cannot be a named format (did you mean to use --output-name instead)",
			expOutFmt: "",
		},
		{
			name:      "--output-format with format",
			argsIn:    []string{"--output-format", "Jan 02, 2006"},
			expOutFmt: "Jan 02, 2006",
		},
		{
			name:      "-f with format",
			argsIn:    []string{"-f", "13:14:15.999"},
			expOutFmt: "13:14:15.999",
		},
		{
			name:      "arg --output-format format arg",
			argsIn:    []string{"val1", "--output-format", "2006 03:04:05", "SecondVal"},
			expArgs:   []string{"val1", "SecondVal"},
			expOutFmt: "2006 03:04:05",
		},
		{
			name:      "arg -f format arg",
			argsIn:    []string{"val1", "-f", "Mon 02/03", "SecondVal"},
			expArgs:   []string{"val1", "SecondVal"},
			expOutFmt: "Mon 02/03",
		},

		{
			name:    "--input-name without arg",
			argsIn:  []string{"--input-name"},
			expBool: true,
			expErr:  "no argument provided after --input-name, expected a format name",
			expPO:   nil,
		},
		{
			name:    "-i without arg",
			argsIn:  []string{"-i"},
			expBool: true,
			expErr:  "no argument provided after -i, expected a format name",
			expPO:   nil,
		},
		{
			name:    "--input-name with unknown name",
			argsIn:  []string{"--input-name", "crazy"},
			expBool: true,
			expErr:  "unknown input format name \"crazy\"",
			expPO:   nil,
		},
		{
			name:    "-i with unknown name",
			argsIn:  []string{"-i", "OffTheWall"},
			expBool: true,
			expErr:  "unknown input format name \"OffTheWall\"",
			expPO:   nil,
		},
		{
			name:   "--input-name with name exact case",
			argsIn: []string{"--input-name", DtFmtRFC1123Z.Name},
			expPO:  []*NamedFormat{DtFmtRFC1123Z},
		},
		{
			name:   "-i with name exact case",
			argsIn: []string{"-i", DtFmtKitchen.Name},
			expPO:  []*NamedFormat{DtFmtKitchen},
		},
		{
			name:   "--input-name with name different case",
			argsIn: []string{"--input-name", "rfc822Z"},
			expPO:  []*NamedFormat{DtFmtRFC822Z},
		},
		{
			name:   "-i with name different case",
			argsIn: []string{"-i", "ansic"},
			expPO:  []*NamedFormat{DtFmtANSIC},
		},
		{
			name:    "arg --input-name name arg",
			argsIn:  []string{"num1", "--input-name", "layout", "arg2"},
			expArgs: []string{"num1", "arg2"},
			expPO:   []*NamedFormat{DtFmtLayout},
		},
		{
			name:    "arg -i name arg",
			argsIn:  []string{"arg1", "-i", "timeonly", "num2"},
			expArgs: []string{"arg1", "num2"},
			expPO:   []*NamedFormat{DtFmtTimeOnly},
		},

		{
			name:    "--input-format without arg",
			argsIn:  []string{"--input-format"},
			expBool: true,
			expErr:  "no argument provided after --input-format, expected a format string",
			expPO:   nil,
		},
		{
			name:    "-g without arg",
			argsIn:  []string{"-g"},
			expBool: true,
			expErr:  "no argument provided after -g, expected a format string",
			expPO:   nil,
		},
		{
			name:    "--input-format with name arg",
			argsIn:  []string{"--input-format", DtFmtStampMicro.Name},
			expBool: true,
			expErr:  "input format string \"" + DtFmtStampMicro.Name + "\" cannot be a named format (did you mean to use --input-name instead)",
			expPO:   nil,
		},
		{
			name:    "-g with name arg",
			argsIn:  []string{"-g", DtFmtRubyDate.Name},
			expBool: true,
			expErr:  "input format string \"" + DtFmtRubyDate.Name + "\" cannot be a named format (did you mean to use --input-name instead)",
			expPO:   nil,
		},
		{
			name:   "--input-format with format arg",
			argsIn: []string{"--input-format", "Jan 02 03:04"},
			expPO:  []*NamedFormat{MakeNamedFormat("User", "Jan 02 03:04")},
		},
		{
			name:   "-g with format arg",
			argsIn: []string{"-g", "4:05.999999"},
			expPO:  []*NamedFormat{MakeNamedFormat("User", "4:05.999999")},
		},
		{
			name:    "arg --input-format format arg",
			argsIn:  []string{"time1", "--input-format", "Jan 02", "num2"},
			expArgs: []string{"time1", "num2"},
			expPO:   []*NamedFormat{MakeNamedFormat("User", "Jan 02")},
		},
		{
			name:    "arg -g format arg",
			argsIn:  []string{"dur1", "-g", "Mon 03:04 PM", "dur2"},
			expArgs: []string{"dur1", "dur2"},
			expPO:   []*NamedFormat{MakeNamedFormat("User", "Mon 03:04 PM")},
		},

		{
			name:    "just an empty string",
			argsIn:  []string{""},
			expArgs: []string{""},
		},
		{
			name:    "simple formula",
			argsIn:  []string{"3m", "+", "15s"},
			expArgs: []string{"3m", "+", "15s"},
		},
		{
			name: "formula with -v and -g and -f in it",
			argsIn: []string{"2020-03-15T07:42:12Z", "-v", "+",
				"-g", "02 Jan 06 15:04:05 -0700",
				"1d5m", "-",
				"-f", "02 Jan 06",
				"2020-02-01T00:00:00Z"},
			expArgs:   []string{"2020-03-15T07:42:12Z", "+", "1d5m", "-", "2020-02-01T00:00:00Z"},
			expV:      true,
			expOutFmt: "02 Jan 06",
			expPO:     []*NamedFormat{MakeNamedFormat("User", "02 Jan 06 15:04:05 -0700")},
		},

		{
			name:    "just --pipe",
			argsIn:  []string{"--pipe"},
			expArgs: []string{"--pipe"},
		},
		{
			name:    "just -p",
			argsIn:  []string{"-p"},
			expArgs: []string{"-p"},
		},
		{
			name:    "arg --pipe arg",
			argsIn:  []string{"arg1", "--pipe", "arg2"},
			expArgs: []string{"arg1", "--pipe", "arg2"},
		},
		{
			name:    "arg -p arg",
			argsIn:  []string{"arg1", "-p", "arg2"},
			expArgs: []string{"arg1", "-p", "arg2"},
		},
		{
			name:    "-v -p",
			argsIn:  []string{"-v", "-p"},
			expArgs: []string{"-p"},
			expV:    true,
		},
		{
			name:    "-p -v",
			argsIn:  []string{"-p", "-v"},
			expArgs: []string{"-p"},
			expV:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			defer SuppressStderrFn()()
			Verbose = false
			OutputFormat = ""

			if tc.expPO == nil {
				tc.expPO = copySlice(FormatParseOrder)
			}

			var w bytes.Buffer
			var actArgs []string
			var actBool bool
			var err error
			testFunc := func() {
				actArgs, actBool, err = ProcessFlags(tc.argsIn, &w)
			}
			require.NotPanics(t, testFunc, "processFlags(%q, w)", tc.argsIn)
			printed := w.String()
			AssertEqualError(t, tc.expErr, err, "processFlags(%q, w) error", tc.argsIn)
			assert.Equal(t, tc.expArgs, actArgs, "processFlags(%q, w) args", tc.argsIn)
			assert.Equal(t, tc.expBool, actBool, "processFlags(%q, w) bool", tc.argsIn)
			assert.Equal(t, tc.expV, Verbose, "Verbose global variable")
			assert.Equal(t, tc.expOutFmt, OutputFormat, "OutputFormat global variable")
			assert.Equal(t, tc.expPO, FormatParseOrder, "FormatParseOrder global variable")
			for i, exp := range tc.expInPrint {
				assert.Contains(t, printed, exp, "[%d]: Printed text should have %q", i, exp)
			}
		})
	}
}

func TestCombineArgs(t *testing.T) {
	tests := []struct {
		name   string
		argsIn []string
		exp    *CalcArgs
	}{
		{
			name:   "nil",
			argsIn: nil,
			exp:    nil,
		},
		{
			name:   "empty",
			argsIn: []string{},
			exp:    nil,
		},
		{
			name:   "one arg: op",
			argsIn: []string{"+"},
			exp:    &CalcArgs{All: []string{"+"}},
		},
		{
			name:   "arg op arg op arg",
			argsIn: []string{"1", "+", "2", "+", "3"},
			exp:    &CalcArgs{All: []string{"1", "+", "2", "+", "3"}},
		},
		{
			name:   "three args, op, two more",
			argsIn: []string{"1", "2", "3", "+", "4", "5"},
			exp:    &CalcArgs{All: []string{"1 2 3", "+", "4 5"}},
		},
		{
			name:   "three args, op, two more, op",
			argsIn: []string{"1", "2", "3", "+", "4", "5", "+"},
			exp:    &CalcArgs{All: []string{"1 2 3", "+", "4 5", "+"}},
		},

		{
			name:   "just pipe",
			argsIn: []string{"--pipe"},
			exp:    &CalcArgs{All: []string{"--pipe"}, HavePipe: true},
		},
		{
			name:   "arg pipe",
			argsIn: []string{"1", "2", "-p"},
			exp: &CalcArgs{
				All:      []string{"1 2", "-p"},
				HavePipe: true,
				PrePipe:  []string{"1 2"},
				PostPipe: nil,
			},
		},
		{
			name:   "arg op pipe",
			argsIn: []string{"1", "2", "+", "-p"},
			exp: &CalcArgs{
				All:      []string{"1 2", "+", "-p"},
				HavePipe: true,
				PrePipe:  []string{"1 2", "+"},
				PostPipe: nil,
			},
		},
		{
			name:   "pipe arg",
			argsIn: []string{"--pipe", "1", "2"},
			exp: &CalcArgs{
				All:      []string{"--pipe", "1 2"},
				HavePipe: true,
				PrePipe:  nil,
				PostPipe: []string{"1 2"},
			},
		},
		{
			name:   "pipe op arg",
			argsIn: []string{"--pipe", "+", "1", "2"},
			exp: &CalcArgs{
				All:      []string{"--pipe", "+", "1 2"},
				HavePipe: true,
				PrePipe:  nil,
				PostPipe: []string{"+", "1 2"},
			},
		},
		{
			name:   "arg pipe arg",
			argsIn: []string{"1", "2", "-p", "3", "4"},
			exp: &CalcArgs{
				All:      []string{"1 2", "-p", "3 4"},
				HavePipe: true,
				PrePipe:  []string{"1 2"},
				PostPipe: []string{"3 4"},
			},
		},
		{
			name:   "arg pipe op arg",
			argsIn: []string{"1", "2", "-p", "+", "3", "4"},
			exp: &CalcArgs{
				All:      []string{"1 2", "-p", "+", "3 4"},
				HavePipe: true,
				PrePipe:  []string{"1 2"},
				PostPipe: []string{"+", "3 4"},
			},
		},
		{
			name:   "arg op pipe arg",
			argsIn: []string{"1", "2", "+", "-p", "3", "4"},
			exp: &CalcArgs{
				All:      []string{"1 2", "+", "-p", "3 4"},
				HavePipe: true,
				PrePipe:  []string{"1 2", "+"},
				PostPipe: []string{"3 4"},
			},
		},
		{
			name:   "arg op pipe op arg",
			argsIn: []string{"1", "2", "+", "-p", "x", "3", "4"},
			exp: &CalcArgs{
				All:      []string{"1 2", "+", "-p", "x", "3 4"},
				HavePipe: true,
				PrePipe:  []string{"1 2", "+"},
				PostPipe: []string{"x", "3 4"},
			},
		},
		{
			name:   "arg pipe arg pipe arg",
			argsIn: []string{"1", "2", "-p", "3", "4", "-p", "5", "6"},
			exp: &CalcArgs{
				All:      []string{"1 2", "-p", "3 4", "-p", "5 6"},
				HavePipe: true,
				PrePipe:  []string{"1 2", "-p", "3 4"},
				PostPipe: []string{"5 6"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.exp == nil {
				tc.exp = &CalcArgs{}
			}

			var act *CalcArgs
			testFunc := func() {
				act = CombineArgs(tc.argsIn)
			}
			require.NotPanics(t, testFunc, "combineArgs(%q)", tc.argsIn)
			assert.Equal(t, tc.exp, act, "combineArgs(%q) result", tc.argsIn)
		})
	}
}

func TestIsPipeInd(t *testing.T) {
	tests := []struct {
		arg string
		exp bool
	}{
		{arg: "", exp: false},
		{arg: "-p", exp: true},
		{arg: "-P", exp: true},
		{arg: "--pipe", exp: true},
		{arg: "--PIPE", exp: true},
		{arg: "--Pipe", exp: true},
		{arg: "--pope", exp: false},
		{arg: "pipe", exp: false},
		{arg: "p", exp: false},
		{arg: "15m", exp: false},
		{arg: "74", exp: false},
		{arg: "2025-01-10", exp: false},
		{arg: "16:30:15", exp: false},
		{arg: "2025-01-10 16:30:15", exp: false},
	}

	for _, tc := range tests {
		name := tc.arg
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = IsPipeInd(tc.arg)
			}
			require.NotPanics(t, testFunc, "IsPipeInd(%q)", tc.arg)
			assert.Equal(t, tc.exp, act, "IsPipeInd(%q) result", tc.arg)
		})
	}
}

// TODO: func TestIsCharDev(t *testing.T)

func TestMainE(t *testing.T) {
	_, pipeArgErr := ParseDTVal("--pipe")

	tests := []struct {
		name        string
		argsIn      []string
		stdin       string
		expErr      string
		expResult   string
		expInStdout []string
		expInStderr []string
	}{
		{
			name:        "nil args",
			argsIn:      nil,
			expInStdout: []string{"date-math"},
		},
		{
			name:        "empty args",
			argsIn:      []string{},
			expInStdout: []string{"date-math"},
		},
		{
			name:        "formats",
			argsIn:      []string{"formats"},
			expInStdout: []string{"* = possible input format"},
		},
		{
			name:        "verbose formula",
			argsIn:      []string{"-v", "3", "+", "8"},
			expResult:   "11",
			expInStderr: []string{"Input date/time formats"},
		},
		{
			name:   "calc error",
			argsIn: []string{"3m", "+", "8"},
			expErr: "cannot apply operation 3m0s + 8: operation <dur> + <num> not defined",
		},
		{
			name:      "formula with time result",
			argsIn:    []string{"2020-10-08", "08:33:15", "-", "4h15m", "-g", "2006-01-02 03:04:05"},
			expResult: "2020-10-08 04:18:15",
		},
		{
			name:      "formula with dur result",
			argsIn:    []string{"2022-09-23", "06:07:08", "-", "2022-09-22", "06:07:08"},
			expResult: "1d",
		},
		{
			name:      "formula with num result",
			argsIn:    []string{"3m", "/", "15s"},
			expResult: "12",
		},
		{
			name: "fancy formula",
			argsIn: []string{
				"1730873499", "-", "1730873400", // => 99s
				"x", "3", // => 297s = 4m57s
				"/", "5", // => 59.4s
				"+", "2002-05-08", "04:20:00", "-0000", // => 2002-05-08 04:20:59.4 -0000
			},
			expResult: "2002-05-08 04:20:59.4 +0000",
		},
		{
			name:      "no args, two formulas piped in",
			argsIn:    nil,
			stdin:     "15m - 5s\n15m + 5s",
			expResult: "14m55s\n15m5s",
		},
		{
			name:   "piped in partial formulas",
			argsIn: []string{"2002-05-08", "4:20:00", "-0000", "-p", "+", "5s"},
			stdin: strings.Join([]string{
				"+ 15s",
				"- 2002-05-08 1:15:03 -0000 x 2",
			}, "\n"),
			expResult: strings.Join([]string{
				"2002-05-08 04:20:20 +0000",
				"6h9m59s",
			}, "\n"),
		},
		{
			name:   "two pipe args",
			argsIn: []string{"5m", "+", "--pipe", "-", "--pipe"},
			stdin:  "10s\n11s",
			expErr: pipeArgErr.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			origStderr := os.Stderr
			stderrR, stderrW, _ := os.Pipe()
			resetStderr := func() {
				os.Stderr = origStderr
				if stderrW != nil {
					stderrW.Close()
					stderrW = nil
				}
			}
			defer func() {
				resetStderr()
			}()
			os.Stderr = stderrW

			var stdin io.Reader
			if len(tc.stdin) > 0 {
				var stdinB bytes.Buffer
				stdinB.WriteString(tc.stdin)
				stdin = &stdinB
			}

			var stdoutB bytes.Buffer
			var err error
			testFunc := func() {
				err = MainE(tc.argsIn, &stdoutB, stdin)
			}
			require.NotPanics(t, testFunc, "mainE(%q, w)", tc.argsIn)
			stdout := stdoutB.String()
			resetStderr()
			stderrBz, _ := io.ReadAll(stderrR)
			stderrR.Close()
			stderr := string(stderrBz)

			AssertEqualError(t, tc.expErr, err, "mainE(%q, w) error", tc.argsIn)
			for i, exp := range tc.expInStdout {
				assert.Contains(t, stdout, exp, "[%d]: stdout should contain %q", i, exp)
			}
			if len(tc.expResult) > 0 {
				assert.Equal(t, tc.expResult+"\n", stdout, "stdout should only have the result")
			}
			if len(tc.expInStdout) == 0 && len(tc.expResult) == 0 {
				assert.Empty(t, stdout, "nothing is expected in stdout")
			}

			for i, exp := range tc.expInStderr {
				assert.Contains(t, stderr, exp, "[%d]: stderr should contain %q", i, exp)
			}
			if len(tc.expInStderr) == 0 {
				assert.Empty(t, stderr, "nothing is expected in stderr")
			}

			if t.Failed() {
				t.Logf("stdout:\n%s", stdout)
				t.Logf("stderr:\n%s", stderr)
			}
		})
	}
}
