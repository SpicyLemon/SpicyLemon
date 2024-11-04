package main_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

func TestOperation_Validate(t *testing.T) {
	expErr := func(val string) string {
		return `unknown operation "` + val + `": must be either "+" or "-" or "x" or "/"`
	}
	tests := []struct {
		name string
		op   Operation
		exp  string
	}{
		{name: "empty", op: "", exp: expErr("")},
		{name: "OpAdd", op: OpAdd},
		{name: "OpSub", op: OpSub},
		{name: "OpMul", op: OpMul},
		{name: "OpDiv", op: OpDiv},
		{name: "+", op: Operation("+")},
		{name: "-", op: Operation("-")},
		{name: "x", op: Operation("x")},
		{name: "/", op: Operation("/")},
		{name: "++", op: "++", exp: expErr("++")},
		{name: "other", op: "other", exp: expErr("other")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.op.Validate()
			}
			require.NotPanics(t, testFunc, "Validate()")
			AssertEqualError(t, tc.exp, err, "Validate() result")
		})
	}
}

func TestOperation_Is(t *testing.T) {
	tests := []struct {
		op  Operation
		arg string
		exp bool
	}{
		{op: OpAdd, arg: "+", exp: true},
		{op: OpAdd, arg: "-", exp: false},
		{op: OpAdd, arg: "x", exp: false},
		{op: OpAdd, arg: "/", exp: false},
		{op: OpAdd, arg: "*", exp: false},
		{op: OpAdd, arg: "", exp: false},
		{op: OpAdd, arg: "other", exp: false},

		{op: OpSub, arg: "+", exp: false},
		{op: OpSub, arg: "-", exp: true},
		{op: OpSub, arg: "x", exp: false},
		{op: OpSub, arg: "/", exp: false},
		{op: OpSub, arg: "*", exp: false},
		{op: OpSub, arg: "", exp: false},
		{op: OpSub, arg: "other", exp: false},

		{op: OpMul, arg: "+", exp: false},
		{op: OpMul, arg: "-", exp: false},
		{op: OpMul, arg: "x", exp: true},
		{op: OpMul, arg: "/", exp: false},
		{op: OpMul, arg: "*", exp: false},
		{op: OpMul, arg: "", exp: false},
		{op: OpMul, arg: "other", exp: false},

		{op: OpDiv, arg: "+", exp: false},
		{op: OpDiv, arg: "-", exp: false},
		{op: OpDiv, arg: "x", exp: false},
		{op: OpDiv, arg: "/", exp: true},
		{op: OpDiv, arg: "*", exp: false},
		{op: OpDiv, arg: "", exp: false},
		{op: OpDiv, arg: "other", exp: false},
	}

	for _, tc := range tests {
		argName := tc.arg
		if len(argName) == 0 {
			argName = "empty"
		}

		t.Run(tc.op.Name()+" "+argName, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.op.Is(tc.arg)
			}
			require.NotPanics(t, testFunc, "%q.Is(%q)", tc.op, tc.arg)
			assert.Equal(t, tc.exp, act, "%q.Is(%q) result", tc.op, tc.arg)
		})
	}
}

func TestOperation_String(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		exp  string
	}{
		{name: "OpAdd", op: OpAdd, exp: "+"},
		{name: "OpSub", op: OpSub, exp: "-"},
		{name: "OpMul", op: OpMul, exp: "x"},
		{name: "OpDiv", op: OpDiv, exp: "/"},
		{name: "+", op: Operation("+"), exp: "+"},
		{name: "-", op: Operation("-"), exp: "-"},
		{name: "x", op: Operation("x"), exp: "x"},
		{name: "/", op: Operation("/"), exp: "/"},
		{name: "empty", op: Operation(""), exp: ""},
		{name: "other", op: Operation("other"), exp: "other"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.op.String()
			}
			require.NotPanics(t, testFunc, "String()")
			assert.Equal(t, tc.exp, act, "String()")
			act2 := fmt.Sprintf("%s", tc.op)
			assert.Equal(t, tc.exp, act2, "from Sprintf")
		})
	}
}

func TestOperation_Name(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		exp  string
	}{
		{name: "OpAdd", op: OpAdd, exp: "OpAdd"},
		{name: "OpSub", op: OpSub, exp: "OpSub"},
		{name: "OpMul", op: OpMul, exp: "OpMul"},
		{name: "OpDiv", op: OpDiv, exp: "OpDiv"},
		{name: "+", op: Operation("+"), exp: "OpAdd"},
		{name: "-", op: Operation("-"), exp: "OpSub"},
		{name: "x", op: Operation("x"), exp: "OpMul"},
		{name: "/", op: Operation("/"), exp: "OpDiv"},
		{name: "empty", op: Operation(""), exp: "Operation(\"\")"},
		{name: "other", op: Operation("other"), exp: "Operation(\"other\")"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.op.Name()
			}
			require.NotPanics(t, testFunc, "%s.Name()", tc.exp)
			assert.Equal(t, tc.exp, act, "%s.Name()", tc.exp)
		})
	}
}

func TestIsOp(t *testing.T) {
	tests := []struct {
		arg string
		exp bool
	}{
		{arg: "+", exp: true},
		{arg: "-", exp: true},
		{arg: "x", exp: true},
		{arg: "/", exp: true},
		{arg: "", exp: false},
		{arg: "other", exp: false},
		{arg: "*", exp: false},
	}

	for _, tc := range tests {
		name := tc.arg
		if len(name) == 0 {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = IsOp(tc.arg)
			}
			require.NotPanics(t, testFunc, "IsOp(%q)", tc.arg)
			assert.Equal(t, tc.exp, act, "IsOp(%q) result", tc.arg)
		})
	}
}

func TestParseOperation(t *testing.T) {
	expErr := func(arg string) string {
		return `unknown operation "` + arg + `": must be either "+" or "-" or "x" or "/"`
	}
	tests := []struct {
		arg    string
		expOp  Operation
		expErr string
	}{
		{arg: "", expOp: "", expErr: "empty operation argument not allowed"},
		{arg: "+", expOp: OpAdd},
		{arg: "-", expOp: OpSub},
		{arg: "x", expOp: OpMul},
		{arg: "/", expOp: OpDiv},
		{arg: "other", expOp: "other", expErr: expErr("other")},
		{arg: "*", expOp: "*", expErr: expErr("*")},
	}

	for _, tc := range tests {
		name := tc.arg
		if len(name) == 0 {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			var actOp Operation
			var err error
			testFunc := func() {
				actOp, err = ParseOperation(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseOperation(%q)", tc.arg)
			AssertEqualError(t, tc.expErr, err, "ParseOperation(%q) error", tc.arg)
			assert.Equal(t, tc.expOp, actOp, "ParseOperation(%q) Operation", tc.arg)
		})
	}
}
