package main_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

func TestDoCalculation(t *testing.T) {
	tests := []struct {
		name    string
		formula []string
		expVal  *DTVal
		expErr  string
		expStep int
	}{
		{
			name:    "nil formula",
			formula: nil,
			expVal:  nil,
			expErr:  "no formula provided",
			expStep: 0,
		},
		{
			name:    "empty formula",
			formula: []string{},
			expVal:  nil,
			expErr:  "no formula provided",
			expStep: 0,
		},
		{
			name:    "just time",
			formula: []string{"2001-02-03 04:05:06"},
			expVal:  NewTimeVal(time.Date(2001, 2, 3, 4, 5, 6, 0, time.Local)),
			expStep: 0,
		},
		{
			name:    "just epoch",
			formula: []string{"981173106.789"},
			expVal:  NewTimeVal(time.Date(2001, 2, 3, 4, 5, 6, 789000000, time.UTC).In(time.Local)),
			expStep: 0,
		},
		{
			name:    "just dur",
			formula: []string{"1h3m"},
			expVal:  NewDurVal(time.Hour + time.Minute*3),
			expStep: 0,
		},
		{
			name:    "just num",
			formula: []string{"3"},
			expVal:  NewNumVal(3),
			expStep: 0,
		},
		{
			name:    "ends with op",
			formula: []string{"3", "+", "4", "x", "1h3m", "-"},
			expErr:  "formula ends with operation \"-\": must end in value",
			expStep: 3,
		},
		{
			name:    "three ops: all valid",
			formula: []string{"2020-03-15 16:20:00", "-", "2020-03-14 22:40:15", "/", "30m"},
			// => "17h39m45s" / "30m" => 63585s / 1800s => 35.325
			expVal:  NewNumVal(35),
			expStep: 2,
		},
		{
			name:    "three ops: first invalid",
			formula: []string{"30m", "-", "5", "x", "2"},
			expErr:  "cannot apply operation 30m0s - 5: operation <dur> - <num> not defined",
			expStep: 1,
		},
		{
			name:    "three ops: second invalid",
			formula: []string{"30m", "x", "5", "-", "15"},
			expErr:  "cannot apply operation 2h30m0s - 15: operation <dur> - <num> not defined",
			expStep: 2,
		},
		{
			name: "one op",
			formula: []string{
				"2010-12-23 08:05:00", "+", "1d15h55m", // => 2010-12-25 00:00:00
			},
			expVal:  NewTimeVal(time.Date(2010, 12, 25, 0, 0, 0, 0, time.Local)),
			expStep: 1,
		},
		{
			name: "two ops",
			formula: []string{
				"2010-12-23 08:05:00", "+", "1d15h55m", // => 2010-12-25 00:00:00
				"-", "2010-12-24 20:00:15", // => 3h59m45s = 14385s = 3 * 5 * 7 * 137 seconds
			},
			expVal:  NewDurVal(time.Hour*3 + time.Minute*59 + time.Second*45),
			expStep: 2,
		},
		{
			name: "three ops",
			formula: []string{
				"2010-12-23 08:05:00", "+", "1d15h55m", // => 2010-12-25 00:00:00
				"-", "2010-12-24 20:00:15", // => 3h59m45s = 14385s = 3 * 5 * 7 * 137 seconds
				"/", "2m17s", // => 105
			},
			expVal:  NewNumVal(105),
			expErr:  "",
			expStep: 3,
		},
		{
			name: "four ops",
			formula: []string{
				"2010-12-23 08:05:00", "+", "1d15h55m", // => 2010-12-25 00:00:00
				"-", "2010-12-24 20:00:15", // => 3h59m45s = 14385s = 3 * 5 * 7 * 137 seconds
				"/", "2m17s", // => 105
				"x", "1d", // => 105d = 2520h
			},
			expVal:  NewDurVal(time.Hour * 2520),
			expStep: 4,
		},
		{
			name: "five ops",
			formula: []string{
				"2010-12-23 08:05:00", "+", "1d15h55m", // => 2010-12-25 00:00:00
				"-", "2010-12-24 20:00:15", // => 3h59m45s = 14385s = 3 * 5 * 7 * 137 seconds
				"/", "2m17s", // => 105
				"x", "1d", // => 105d = 2520h
				"+", "2019-11-16 15:16:17", // => 2020-02-29 15:16:17
			},
			expVal:  NewTimeVal(time.Date(2020, 2, 29, 15, 16, 17, 0, time.Local)),
			expStep: 5,
		},
		{
			name: "fancy formula",
			// 1730873499 - 1730873400 x 3 / 5 + 2002-05-08 04:20:00 -0000 =  2002-05-08 04:20:59.4 +0000
			formula: []string{
				"1730873499", "-", "1730873400", // => 99s
				"x", "3", // => 297s = 4m57s
				"/", "5", // => 59.4s
				"+", "2002-05-08 04:20:00 -0000", // => 2002-05-08 04:20:59.4 -0000
			},
			expVal:  NewTimeVal(time.Date(2002, 5, 8, 4, 20, 59, 400000000, time.FixedZone("+0000", 0))),
			expErr:  "",
			expStep: 4,
		},
		{
			name:    "example: time - time 1",
			formula: []string{"2020-01-09 4:30:00", "-", "2020-01-09 3:29:28"},
			expVal:  NewDurVal(time.Hour + time.Second*32),
			expStep: 1,
		},
		{
			name:    "example: time - time 2",
			formula: []string{"2020-01-09 3:29:28", "-", "2020-01-09 4:30:00"},
			expVal:  NewDurVal(-1 * (time.Hour + time.Second*32)),
			expStep: 1,
		},
		{
			name:    "example: time + dur",
			formula: []string{"2020-01-09 4:30:00", "+", "1h2s"},
			expVal:  NewTimeVal(time.Date(2020, 1, 9, 5, 30, 2, 0, time.Local)),
			expStep: 1,
		},
		{
			name:    "example: dur + time",
			formula: []string{"1h2s", "+", "2020-01-09 4:30:00"},
			expVal:  NewTimeVal(time.Date(2020, 1, 9, 5, 30, 2, 0, time.Local)),
			expStep: 1,
		},
		{
			name:    "example: time - dur",
			formula: []string{"2020-01-09 4:30:00", "-", "1h2s"},
			expVal:  NewTimeVal(time.Date(2020, 1, 9, 3, 29, 58, 0, time.Local)),
			expStep: 1,
		},
		{
			name:    "example: dur + dur 1",
			formula: []string{"1h2s", "+", "3m5s"},
			expVal:  NewDurVal(time.Hour + time.Minute*3 + time.Second*7),
			expStep: 1,
		},
		{
			name:    "example: dur + dur 2",
			formula: []string{"3m5s", "+", "1h2s"},
			expVal:  NewDurVal(time.Hour + time.Minute*3 + time.Second*7),
			expStep: 1,
		},
		{
			name:    "example: dur - dur 1",
			formula: []string{"1h2s", "-", "3m5s"},
			expVal:  NewDurVal(time.Minute*56 + time.Second*57),
			expStep: 1,
		},
		{
			name:    "example: dur - dur 2",
			formula: []string{"3m5s", "-", "1h2s"},
			expVal:  NewDurVal(-1 * (time.Minute*56 + time.Second*57)),
			expStep: 1,
		},
		{
			name:    "example: dur / dur",
			formula: []string{"2h", "/", "40m"},
			expVal:  NewNumVal(3),
			expStep: 1,
		},
		{
			name:    "example: dur x num",
			formula: []string{"40m", "x", "3"},
			expVal:  NewDurVal(time.Hour * 2),
			expStep: 1,
		},
		{
			name:    "example: num x dur",
			formula: []string{"5", "x", "40m"},
			expVal:  NewDurVal(time.Hour*3 + time.Minute*20),
			expStep: 1,
		},
		{
			name:    "example: dur / num",
			formula: []string{"2h", "/", "3"},
			expVal:  NewDurVal(time.Minute * 40),
			expStep: 1,
		},
		{
			name:    "example: num + num 1",
			formula: []string{"5", "+", "3"},
			expVal:  NewNumVal(8),
			expStep: 1,
		},
		{
			name:    "example: num + num 2",
			formula: []string{"3", "+", "5"},
			expVal:  NewNumVal(8),
			expStep: 1,
		},
		{
			name:    "example: num - num 1",
			formula: []string{"5", "-", "3"},
			expVal:  NewNumVal(2),
			expStep: 1,
		},
		{
			name:    "example: num - num 2",
			formula: []string{"3", "-", "5"},
			expVal:  NewNumVal(-2),
			expStep: 1,
		},
		{
			name:    "example: num x num 1",
			formula: []string{"5", "x", "3"},
			expVal:  NewNumVal(15),
			expStep: 1,
		},
		{
			name:    "example: num x num 2",
			formula: []string{"3", "x", "5"},
			expVal:  NewNumVal(15),
			expStep: 1,
		},
		{
			name:    "example: num / num 1",
			formula: []string{"6", "/", "3"},
			expVal:  NewNumVal(2),
			expStep: 1,
		},
		{
			name:    "example: num / num 2",
			formula: []string{"5", "/", "3"},
			expVal:  NewNumVal(1),
			expStep: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			var actVal *DTVal
			var err error
			testFunc := func() {
				actVal, err = DoCalculation(tc.formula)
			}
			require.NotPanics(t, testFunc, "DoCalculation(%q)", tc.formula)
			AssertEqualError(t, tc.expErr, err, "DoCalculation(%q) error", tc.formula)
			assert.Equal(t, tc.expVal.String(), actVal.String(), "DoCalculation(%q) result", tc.formula)
			assert.Equal(t, tc.expStep, CurStep, "CurStep")
		})
	}
}

func TestApplyOperation(t *testing.T) {
	type testCase struct {
		name     string
		leftVal  *DTVal
		op       Operation
		rightVal *DTVal
		expVal   *DTVal
		expErr   string
		errW     func(testCase) string // Defaults to sWrapper.
	}
	// sWrapper returns a new expected error string with the added context using unquoted values.
	// If there's no expected error, this will return an empty string.
	sWrapper := func(tc testCase) string {
		if len(tc.expErr) == 0 {
			return tc.expErr
		}
		return "cannot apply operation " + tc.leftVal.String() + " " +
			tc.op.String() + " " + tc.rightVal.String() + ": " + tc.expErr
	}
	// qWrapper returns a new expected error string with the added context using quoted values.
	// If there's no expected error, this will return an empty string.
	qWrapper := func(tc testCase) string {
		if len(tc.expErr) == 0 {
			return tc.expErr
		}
		return "cannot apply operation \"" + tc.leftVal.String() + "\" " +
			"\"" + tc.op.String() + "\" \"" + tc.rightVal.String() + "\": " + tc.expErr
	}

	tests := []testCase{
		{
			name:     "invalid left value",
			leftVal:  nil,
			op:       "+",
			rightVal: NewNumVal(3),
			expErr:   "invalid left value: cannot be nil",
			errW:     qWrapper,
		},
		{
			name:     "invalid op",
			leftVal:  NewNumVal(3),
			op:       "*",
			rightVal: NewNumVal(3),
			expErr:   "invalid operation: unknown operation \"*\": must be either \"+\" or \"-\" or \"x\" or \"/\"",
			errW:     qWrapper,
		},
		{
			name:     "invalid right value",
			leftVal:  NewNumVal(3),
			op:       "+",
			rightVal: nil,
			expErr:   "invalid right value: cannot be nil",
			errW:     qWrapper,
		},

		{
			name:     "dur + dur",
			leftVal:  NewDurVal(time.Minute + time.Second*3),
			op:       "+",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expVal:   NewDurVal(time.Hour + time.Minute*9 + time.Second*3),
		},
		{
			name:     "dur + time",
			leftVal:  NewDurVal(time.Minute + time.Second*3),
			op:       "+",
			rightVal: NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			expVal:   NewTimeVal(time.Date(2001, 3, 14, 15, 17, 20, 18, time.UTC)),
		},
		{
			name:     "time + dur",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "+",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expVal:   NewTimeVal(time.Date(2001, 3, 14, 16, 24, 17, 18, time.UTC)),
		},
		{
			name:     "num + num",
			leftVal:  NewNumVal(3),
			op:       "+",
			rightVal: NewNumVal(8),
			expVal:   NewNumVal(11),
		},
		{
			name:     "time + time",
			leftVal:  NewTimeVal(time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)),
			op:       "+",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <time> + <time> not defined",
		},
		{
			name:     "time + num",
			leftVal:  NewTimeVal(time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)),
			op:       "+",
			rightVal: NewNumVal(3),
			expErr:   "operation <time> + <num> not defined",
		},
		{
			name:     "dur  + num",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "+",
			rightVal: NewNumVal(8),
			expErr:   "operation <dur> + <num> not defined",
		},
		{
			name:     "num  + time",
			leftVal:  NewNumVal(3),
			op:       "+",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <num> + <time> not defined",
		},
		{
			name:     "num  + dur",
			leftVal:  NewNumVal(3),
			op:       "+",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <num> + <dur> not defined",
		},

		{
			name:     "dur - dur",
			leftVal:  NewDurVal(time.Minute + time.Second*3),
			op:       "-",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expVal:   NewDurVal(-1 * (time.Hour + time.Minute*6 + time.Second*57)),
		},
		{
			name:     "time - dur",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "-",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expVal:   NewTimeVal(time.Date(2001, 3, 14, 14, 8, 17, 18, time.UTC)),
		},
		{
			name:     "time - time",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "-",
			rightVal: NewTimeVal(time.Date(2001, 3, 13, 14, 15, 16, 17, time.UTC)),
			expVal:   NewDurVal(time.Hour*25 + time.Minute + time.Second + time.Nanosecond),
		},
		{
			name:     "num - num",
			leftVal:  NewNumVal(3),
			op:       "-",
			rightVal: NewNumVal(8),
			expVal:   NewNumVal(-5),
		},
		{
			name:     "time - num",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "-",
			rightVal: NewNumVal(3),
			expErr:   "operation <time> - <num> not defined",
		},
		{
			name:     "dur  - time",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "-",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <dur> - <time> not defined",
		},
		{
			name:     "dur  - num",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "-",
			rightVal: NewNumVal(8),
			expErr:   "operation <dur> - <num> not defined",
		},
		{
			name:     "num  - time",
			leftVal:  NewNumVal(3),
			op:       "-",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <num> - <time> not defined",
		},
		{
			name:     "num  - dur",
			leftVal:  NewNumVal(3),
			op:       "-",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <num> - <dur> not defined",
		},

		{
			name:     "dur x num",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "x",
			rightVal: NewNumVal(8),
			expVal:   NewDurVal(time.Hour*8 + time.Minute*24),
		},
		{
			name:     "num x dur",
			leftVal:  NewNumVal(3),
			op:       "x",
			rightVal: NewDurVal(time.Hour + time.Minute*40),
			expVal:   NewDurVal(time.Hour * 5),
		},
		{
			name:     "num x num",
			leftVal:  NewNumVal(3),
			op:       "x",
			rightVal: NewNumVal(8),
			expVal:   NewNumVal(24),
		},
		{
			name:     "time x time",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "x",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <time> x <time> not defined",
		},
		{
			name:     "time x dur",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "x",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <time> x <dur> not defined",
		},
		{
			name:     "time x num",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "x",
			rightVal: NewNumVal(8),
			expErr:   "operation <time> x <num> not defined",
		},
		{
			name:     "dur  x time",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "x",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <dur> x <time> not defined",
		},
		{
			name:     "dur  x dur",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "x",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <dur> x <dur> not defined",
		},
		{
			name:     "num  x time",
			leftVal:  NewNumVal(3),
			op:       "x",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <num> x <time> not defined",
		},

		{
			name:     "dur / num",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "/",
			rightVal: NewNumVal(8),
			expVal:   NewDurVal(time.Minute*7 + time.Second*52 + time.Millisecond*500),
		},
		{
			name:     "dur / dur",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "/",
			rightVal: NewDurVal(time.Second * 30),
			expVal:   NewNumVal(126),
		},
		{
			name:     "num / num",
			leftVal:  NewNumVal(8),
			op:       "/",
			rightVal: NewNumVal(3),
			expVal:   NewNumVal(2),
		},
		{
			name:     "time / time",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "/",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <time> / <time> not defined",
		},
		{
			name:     "time / dur",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "/",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <time> / <dur> not defined",
		},
		{
			name:     "time / num",
			leftVal:  NewTimeVal(time.Date(2001, 3, 14, 15, 16, 17, 18, time.UTC)),
			op:       "/",
			rightVal: NewNumVal(8),
			expErr:   "operation <time> / <num> not defined",
		},
		{
			name:     "dur  / time",
			leftVal:  NewDurVal(time.Hour + time.Minute*3),
			op:       "/",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <dur> / <time> not defined",
		},
		{
			name:     "num  / time",
			leftVal:  NewNumVal(3),
			op:       "/",
			rightVal: NewTimeVal(time.Date(2002, 3, 4, 5, 6, 7, 8, time.UTC)),
			expErr:   "operation <num> / <time> not defined",
		},
		{
			name:     "num  / dur",
			leftVal:  NewNumVal(3),
			op:       "/",
			rightVal: NewDurVal(time.Hour + time.Minute*8),
			expErr:   "operation <num> / <dur> not defined",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errW == nil {
				tc.errW = sWrapper
			}
			expErr := tc.errW(tc)

			var actVal *DTVal
			var err error
			testFunc := func() {
				actVal, err = ApplyOperation(tc.leftVal, tc.op, tc.rightVal)
			}
			require.NotPanics(t, testFunc, "ApplyOperation(%s, %s, %s)", tc.leftVal, tc.op, tc.rightVal)
			AssertEqualError(t, expErr, err, "ApplyOperation(%s, %s, %s) error", tc.leftVal, tc.op, tc.rightVal)
			// Comparing the vals as strings because the Location in the Time values makes it
			// impossible to create an expected Time that deep-equals one parsed from a string.
			assert.Equal(t, tc.expVal.String(), actVal.String(), "ApplyOperation(%s, %s, %s) result", tc.leftVal, tc.op, tc.rightVal)
		})
	}
}
