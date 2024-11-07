package main_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

func TestNewTimeVal(t *testing.T) {
	theTime := time.Unix(1234567890, 55) // 2009-02-13 23:31:30.000000055 +0000 (UTC) Friday
	var val *DTVal
	testFunc := func() {
		val = NewTimeVal(theTime)
	}
	require.NotPanics(t, testFunc, "NewTimeVal")
	require.NotNil(t, val, "NewTimeVal result")
	assert.Equal(t, &theTime, val.Time, "result.Time")
	assert.NoError(t, val.Validate(), "Validate")
}

func TestNewDurVal(t *testing.T) {
	theDur := time.Hour + time.Minute*20
	var val *DTVal
	testFunc := func() {
		val = NewDurVal(theDur)
	}
	require.NotPanics(t, testFunc, "NewDurVal")
	require.NotNil(t, val, "NewDurVal result")
	assert.Equal(t, &theDur, val.Dur, "result.Dur")
	assert.NoError(t, val.Validate(), "Validate")
}

func TestNewNumVal(t *testing.T) {
	theNum := 12
	var val *DTVal
	testFunc := func() {
		val = NewNumVal(theNum)
	}
	require.NotPanics(t, testFunc, "NewNumVal")
	require.NotNil(t, val, "NewNumVal result")
	assert.Equal(t, &theNum, val.Num, "result.Num")
	assert.NoError(t, val.Validate(), "Validate")
}

func TestDTVal_IsTime(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  bool
	}{
		{name: "nil", val: nil, exp: false},
		{name: "empty", val: &DTVal{}, exp: false},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: true},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: false},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: false},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.val.IsTime()
			}
			require.NotPanics(t, testFunc, "%s.IsTime()", tc.val)
			assert.Equal(t, tc.exp, act, "IsTime() result")
		})
	}
}

func TestDTVal_IsDur(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  bool
	}{
		{name: "nil", val: nil, exp: false},
		{name: "empty", val: &DTVal{}, exp: false},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: false},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: true},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: false},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.val.IsDur()
			}
			require.NotPanics(t, testFunc, "%s.IsDur()", tc.val)
			assert.Equal(t, tc.exp, act, "IsDur() result")
		})
	}
}

func TestDTVal_IsNum(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  bool
	}{
		{name: "nil", val: nil, exp: false},
		{name: "empty", val: &DTVal{}, exp: false},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: false},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: false},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: true},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.val.IsNum()
			}
			require.NotPanics(t, testFunc, "%s.IsNum()", tc.val)
			assert.Equal(t, tc.exp, act, "IsNum() result")
		})
	}
}

func TestDTVal_TimeString(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: NilStr},
		{name: "empty", val: &DTVal{}, exp: NilStr},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: theTime.String()},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: NilStr},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: NilStr},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: theTime.String()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.TimeString()
			}
			require.NotPanics(t, testFunc, "%s.TimeString()", tc.val)
			assert.Equal(t, tc.exp, act, "TimeString() result")
		})
	}
}

func TestDTVal_DurString(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: NilStr},
		{name: "empty", val: &DTVal{}, exp: NilStr},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: NilStr},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: theDur.String()},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: NilStr},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: theDur.String()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.DurString()
			}
			require.NotPanics(t, testFunc, "%s.DurString()", tc.val)
			assert.Equal(t, tc.exp, act, "DurString() result")
		})
	}
}

func TestDTVal_NumString(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: NilStr},
		{name: "empty", val: &DTVal{}, exp: NilStr},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: NilStr},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: NilStr},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: strconv.Itoa(theNum)},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: strconv.Itoa(theNum)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.NumString()
			}
			require.NotPanics(t, testFunc, "%s.NumString()", tc.val)
			assert.Equal(t, tc.exp, act, "NumString() result")
		})
	}
}

func TestDTVal_String(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: NilStr},
		{name: "empty", val: &DTVal{}, exp: EmptyStr},
		{name: "only time", val: &DTVal{Time: &theTime}, exp: theTime.String()},
		{name: "only duration", val: &DTVal{Dur: &theDur}, exp: theDur.String()},
		{name: "only number", val: &DTVal{Num: &theNum}, exp: strconv.Itoa(theNum)},
		{
			name: "time and duration",
			val:  &DTVal{Time: &theTime, Dur: &theDur},
			exp:  "{" + theTime.String() + "|" + theDur.String() + "}",
		},
		{
			name: "time and number",
			val:  &DTVal{Time: &theTime, Num: &theNum},
			exp:  "{" + theTime.String() + "|" + strconv.Itoa(theNum) + "}",
		},
		{
			name: "duration and number",
			val:  &DTVal{Dur: &theDur, Num: &theNum},
			exp:  "{" + theDur.String() + "|" + strconv.Itoa(theNum) + "}",
		},
		{
			name: "all",
			val:  &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum},
			exp:  "{" + theTime.String() + "|" + theDur.String() + "|" + strconv.Itoa(theNum) + "}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.String()
			}
			require.NotPanics(t, testFunc, "String()")
			assert.Equal(t, tc.exp, act, "String() result")
		})
	}
}

func TestDTVal_Validate(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: "cannot be nil"},
		{name: "empty", val: &DTVal{}, exp: "cannot be empty"},
		{name: "only time", val: &DTVal{Time: &theTime}},
		{name: "only duration", val: &DTVal{Dur: &theDur}},
		{name: "only number", val: &DTVal{Num: &theNum}},
		{
			name: "time and duration",
			val:  &DTVal{Time: &theTime, Dur: &theDur},
			exp: "can only have one of datetime (" + theTime.String() + ") " +
				"or duration (" + theDur.String() + ") or number (" + NilStr + ")",
		},
		{
			name: "time and number",
			val:  &DTVal{Time: &theTime, Num: &theNum},
			exp: "can only have one of datetime (" + theTime.String() + ") " +
				"or duration (" + NilStr + ") or number (" + strconv.Itoa(theNum) + ")",
		},
		{
			name: "duration and number",
			val:  &DTVal{Dur: &theDur, Num: &theNum},
			exp: "can only have one of datetime (" + NilStr + ") " +
				"or duration (" + theDur.String() + ") or number (" + strconv.Itoa(theNum) + ")",
		},
		{
			name: "all",
			val:  &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum},
			exp: "can only have one of datetime (" + theTime.String() + ") " +
				"or duration (" + theDur.String() + ") or number (" + strconv.Itoa(theNum) + ")",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.val.Validate()
			}
			require.NotPanics(t, testFunc, "Validate()")
			AssertEqualError(t, tc.exp, err, "%s.Validate()", tc.val)
		})
	}
}

func TestDTVal_TypeString(t *testing.T) {
	theTime := time.Unix(1234567890, 55)
	theDur := time.Hour + time.Minute*20
	theNum := 12

	tests := []struct {
		name string
		val  *DTVal
		exp  string
	}{
		{name: "nil", val: nil, exp: NilStr},
		{name: "empty", val: &DTVal{}, exp: EmptyStr},
		{name: "NewTimeVal", val: NewTimeVal(theTime), exp: "<time>"},
		{name: "NewDurVal", val: NewDurVal(theDur), exp: "<dur>"},
		{name: "NewNumVal", val: NewNumVal(theNum), exp: "<num>"},
		{name: "all", val: &DTVal{Time: &theTime, Dur: &theDur, Num: &theNum}, exp: "<time>"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.TypeString()
			}
			require.NotPanics(t, testFunc, "%s.TypeString()", tc.val)
			assert.Equal(t, tc.exp, act, "TypeString() result")
		})
	}
}

func TestDTVal_FormattedString(t *testing.T) {
	tests := []struct {
		name         string
		inputFormat  string
		outputFormat string
		usedFormats  []*NamedFormat
		val          *DTVal
		exp          string
	}{
		{
			name: "nil",
			val:  nil,
			exp:  "invalid result: cannot be nil",
		},
		{
			name: "invalid",
			val:  &DTVal{},
			exp:  "invalid result: cannot be empty",
		},

		{
			name: "positive number",
			val:  NewNumVal(62),
			exp:  "62",
		},
		{
			name: "negative number",
			val:  NewNumVal(-16),
			exp:  "-16",
		},

		{
			name:         "time: with output format",
			outputFormat: "4 minutes and 5 seconds past 3 PM on day 2 of 1 in 2006",
			val:          NewTimeVal(time.Date(2024, 3, 7, 11, 15, 42, 0, time.Local)),
			exp:          "15 minutes and 42 seconds past 11 AM on day 7 of 3 in 2024",
		},
		{
			name: "time: no used formats",
			val:  NewTimeVal(time.Date(2023, 4, 15, 20, 7, 18, 0, time.UTC)),
			exp:  "2023-04-15 20:07:18 +0000 UTC",
		},
		{
			name:        "time: one used format, not complete",
			usedFormats: []*NamedFormat{DtFmtDateTime},
			val:         NewTimeVal(time.Date(2023, 4, 15, 20, 7, 18, 0, time.UTC)),
			exp:         "2023-04-15 20:07:18 +0000 UTC",
		},
		{
			name:        "time: one used format, complete",
			usedFormats: []*NamedFormat{DtFmtRFC3339Nano},
			val:         NewTimeVal(time.Date(2023, 4, 15, 20, 7, 18, 0, time.UTC)),
			exp:         "2023-04-15T20:07:18Z",
		},
		{
			name:        "time: two used formats",
			usedFormats: []*NamedFormat{DtFmtRFC3339Nano, DtFmtDateTime},
			val:         NewTimeVal(time.Date(2015, 2, 6, 9, 55, 1, 550000000, time.UTC)),
			exp:         "2015-02-06 09:55:01.55 +0000 UTC",
		},
		{
			name:        "time: with input format",
			inputFormat: "2006-01-02 03:04 PM",
			val:         NewTimeVal(time.Date(1983, 6, 6, 7, 7, 32, 55, time.UTC)),
			exp:         "1983-06-06 07:07 AM",
		},

		{
			name: "duration zero",
			val:  NewDurVal(0),
			exp:  "0s",
		},
		{
			name: "duration less than one day",
			val:  NewDurVal(time.Hour*23 + time.Minute*59),
			exp:  "23h59m",
		},
		{
			name: "duration more than one day",
			val:  NewDurVal(time.Hour*24 + time.Minute),
			exp:  "1d0h1m",
		},
		{
			name: "duration more than a week",
			val:  NewDurVal(time.Hour*24*8 + time.Minute),
			exp:  "8d0h1m",
		},
		{
			name: "duration only days",
			val:  NewDurVal(time.Hour * 24 * 8),
			exp:  "8d",
		},
		{
			name: "duration with fractional seconds.",
			val:  NewDurVal(time.Hour + time.Minute*2 + time.Second*3 + time.Millisecond*40),
			exp:  "1h2m3.04s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			OutputFormat = tc.outputFormat
			UsedInputFormats = tc.usedFormats
			if len(tc.inputFormat) > 0 {
				InputFormat = MakeNamedFormat("User", tc.inputFormat)
			}
			Verbose = false

			var act string
			testFunc := func() {
				act = tc.val.FormattedString()
			}
			require.NotPanics(t, testFunc, "%s.FormattedString()", tc.val)
			assert.Equal(t, tc.exp, act, "%s.FormattedString() result", tc.val)
		})
	}
}

func TestParseDTVal(t *testing.T) {
	origLocal := time.Local
	defer func() {
		time.Local = origLocal
	}()
	time.Local = time.UTC

	tests := []struct {
		name     string
		arg      string
		expVal   *DTVal
		expErr   string
		expInErr []string
	}{
		{
			name:   "empty arg",
			arg:    "",
			expVal: nil,
			expErr: "empty value argument not allowed",
		},
		{
			name:   "time",
			arg:    "2024-07-23T14:16:18Z",
			expVal: NewTimeVal(time.Date(2024, 7, 23, 14, 16, 18, 0, time.UTC)),
		},
		{
			name:   "epoch with fractional seconds",
			arg:    "1700000012",
			expVal: NewTimeVal(time.Date(2023, 11, 14, 22, 13, 32, 0, time.Local)),
		},
		{
			name:   "duration",
			arg:    "1w5s",
			expVal: NewDurVal(time.Hour*24*7 + time.Second*5),
		},
		{
			name:   "number 14",
			arg:    "14",
			expVal: NewNumVal(14),
		},
		{
			name:   "number 1,000,000",
			arg:    "1000000",
			expVal: NewNumVal(1_000_000),
		},
		{
			name:   "number 1,000,001",
			arg:    "n1000001",
			expVal: NewNumVal(1000001),
		},
		{
			name:   "epoch 1,000,001",
			arg:    "1000001",
			expVal: NewTimeVal(time.Date(1970, 1, 12, 13, 46, 41, 0, time.Local)),
		},
		{
			name:   "epoch 1,000,000",
			arg:    "e1000000",
			expVal: NewTimeVal(time.Date(1970, 1, 12, 13, 46, 40, 0, time.Local)),
		},
		{
			name: "invalid short",
			arg:  "short",
			expInErr: []string{
				"could not convert \"short\" to either a datetime, epoch, duration, or number",
				"RubyDate",
				"duration: ",
				"epoch: ",
				"number: ",
			},
		},
		{
			name: "invalid long",
			arg:  strings.Repeat("x", 61),
			expInErr: []string{
				"could not convert \"" + strings.Repeat("x", 61) + "\" to either a datetime, epoch, duration, or number",
				"Did you use * instead of x?",
			},
		},
		// I'm not sure how to make it be multiple things.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			Verbose = false

			var actVal *DTVal
			var err error
			testFunc := func() {
				actVal, err = ParseDTVal(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseDTVal(%q)", tc.arg)
			if len(tc.expInErr) > 0 {
				if assert.Error(t, err, "ParseDTVal(%q) error", tc.arg) {
					for i, exp := range tc.expInErr {
						assert.ErrorContains(t, err, exp, "[%d]: ParseDTVal(%q) error", i, tc.arg)
					}
				}
			} else {
				AssertEqualError(t, tc.expErr, err, "ParseDTVal(%q) error", tc.arg)
			}
			assert.Equal(t, tc.expVal.String(), actVal.String(), "ParseDTVal(%q) result", tc.arg)
		})
	}
}

func TestParseTime(t *testing.T) {
	// The time stuff heavily depends on system config and there's no guarantee
	// that it has any knowledge of any timezones other than local and UTC.
	// So we only use local and UTC in these tests and just assume it's actually
	// handling timezones correctly.
	localDTInDST := time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)
	localDTNoDST := time.Date(2024, 7, 1, 0, 0, 0, 0, time.Local)
	localTZInDST := localDTInDST.Format("MST")
	localTZNoDST := localDTNoDST.Format("MST")
	localOffsetInDST := localDTInDST.Format("-0700")
	localOffsetNoDST := localDTNoDST.Format("-0700")

	tests := []struct {
		name    string
		arg     string
		expTime time.Time
		expErr  bool
		fmtName string
	}{
		{
			name:   "empty",
			arg:    "",
			expErr: true,
		},
		// DateTimeZone "2006-01-02 15:04:05.999999999 -0700"
		{
			name:    "DateTimeZone",
			arg:     "2004-05-01 03:15:42.1 -0500",
			expTime: time.Date(2004, 5, 1, 3, 15, 42, 100000000, time.FixedZone("EST", -5*60*60)),
			fmtName: "DateTimeZone",
		},
		// DateTimeZone2 "2006-01-02 15:04:05.999999999Z0700"
		{
			name:    "DateTimeZone2",
			arg:     "2033-11-04 17:08:19-0700",
			expTime: time.Date(2033, 11, 4, 17, 8, 19, 0, time.FixedZone("CDT", -7*60*60)),
			fmtName: "DateTimeZone2",
		},
		// UnixDate    "Mon Jan _2 15:04:05 MST 2006"
		{
			name:    "UnixDate",
			arg:     "Sun Nov  3 20:23:30 " + localTZInDST + " 2024",
			expTime: time.Date(2024, 11, 3, 20, 23, 30, 0, time.Local),
			fmtName: "UnixDate",
		},
		// RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"
		{
			name:    "RFC3339Nano utc",
			arg:     "2004-02-29T23:11:51.999666333Z",
			expTime: time.Date(2004, 2, 29, 23, 11, 51, 999666333, time.UTC),
			fmtName: "RFC3339Nano",
		},
		{
			name:    "RFC3339Nano with offset",
			arg:     "2004-02-29T23:11:51.123321-05:00",
			expTime: time.Date(2004, 2, 29, 23, 11, 51, 123321000, time.FixedZone("EST", -5*60*60)),
			fmtName: "RFC3339Nano",
		},
		{
			name:    "RFC3339Nano no fractional seconds",
			arg:     "2004-02-29T23:11:51-08:00",
			expTime: time.Date(2004, 2, 29, 23, 11, 51, 0, time.FixedZone("PST", -8*60*60)),
			fmtName: "RFC3339Nano",
		},
		// DateTime    "2006-01-02 15:04:05"
		{
			name:    "DateTime",
			arg:     "2032-04-12 13:06:44",
			expTime: time.Date(2032, 4, 12, 13, 6, 44, 0, time.Local),
			fmtName: "DateTime",
		},
		{
			name:    "DateTime with fractional seconds",
			arg:     "2032-04-12 13:06:44.1002003",
			expTime: time.Date(2032, 4, 12, 13, 6, 44, 100200300, time.Local),
			fmtName: "DateTime",
		},
		// RFC1123     "Mon, 02 Jan 2006 15:04:05 MST"
		{
			name:    "RFC1123",
			arg:     "Fri, 09 Jan 1981 03:57:12 " + localTZInDST,
			expTime: time.Date(1981, 1, 9, 3, 57, 12, 0, time.Local),
			fmtName: "RFC1123",
		},
		// RFC1123Z    "Mon, 02 Jan 2006 15:04:05 -0700"
		{
			name:    "RFC1123Z",
			arg:     "Sun, 06 May 2001 02:57:12 " + localOffsetNoDST,
			expTime: time.Date(2001, 5, 6, 2, 57, 12, 0, time.Local),
			fmtName: "RFC1123Z",
		},
		// RubyDate    "Mon Jan 02 15:04:05 -0700 2006"
		{
			name:    "RubyDate",
			arg:     "Thu Nov 29 07:09:11 " + localOffsetInDST + " 2012",
			expTime: time.Date(2012, 11, 29, 7, 9, 11, 0, time.Local),
			fmtName: "RubyDate",
		},
		// ANSIC       "Mon Jan _2 15:04:05 2006"
		{
			name:    "ANSIC",
			arg:     "Tue Aug  4 21:15:03 2015",
			expTime: time.Date(2015, 8, 4, 21, 15, 3, 0, time.Local),
			fmtName: "ANSIC",
		},
		// RFC850      "Monday, 02-Jan-06 15:04:05 MST"
		{
			name:    "RFC850",
			arg:     "Thursday, 08-Oct-20 17:10:59 " + localTZNoDST,
			expTime: time.Date(2020, 10, 8, 17, 10, 59, 0, time.Local),
			fmtName: "RFC850",
		},
		// unknown
		{
			name:   "unknown",
			arg:    "unknown",
			expErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			UsedInputFormats = nil

			var actTime time.Time
			var err error
			testFunc := func() {
				actTime, err = ParseTime(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseTime(%q)", tc.arg)

			if tc.expErr {
				if assert.Error(t, err, "ParseTime(%q) error", tc.arg) {
					assert.ErrorContains(t, err, tc.arg, "ParseTime(%q) error should contain the arg", tc.arg)
					for i, nf := range FormatParseOrder {
						assert.ErrorContains(t, err, nf.Format, "ParseTime(%q) error should contain FormatParseOrder[%d].Format", tc.arg, i)
						assert.ErrorContains(t, err, nf.Name, "ParseTime(%q) error should contain FormatParseOrder[%d].Name", tc.arg, i)
					}
				}
			} else {
				assert.NoError(t, err, "ParseTime(%q) error", tc.arg)
			}

			AssertEqualTime(t, tc.expTime, actTime, "ParseTime(%q) result (string)", tc.arg)

			if len(tc.fmtName) > 0 {
				if assert.Len(t, UsedInputFormats, 1, "UsedInputFormats") {
					assert.Equal(t, tc.fmtName, UsedInputFormats[0].Name, "UsedInputFormats[0].Name")
				}
			} else {
				assert.Empty(t, UsedInputFormats, "UsedInputFormats")
			}
		})
	}
}

func TestParseEpoch(t *testing.T) {
	// time.Unix returns time in the local timezone.
	// But that's going to change depending on who's running this test.
	// So we swap out the "local" one for UTC, then put it back when done.
	origLocal := time.Local
	defer func() {
		time.Local = origLocal
	}()
	time.Local = time.UTC

	tests := []struct {
		name    string
		arg     string
		expTime time.Time
		expErr  string
	}{
		{
			name:   "empty",
			arg:    "",
			expErr: "empty string not allowed",
		},
		{
			name:    "whole number",
			arg:     "1650428400",
			expTime: time.Date(2022, 4, 20, 4, 20, 0, 0, time.Local),
		},
		{
			name:    "e whole number",
			arg:     "e1650428400",
			expTime: time.Date(2022, 4, 20, 4, 20, 0, 0, time.Local),
		},
		{
			name:    "with 3 fractional seconds",
			arg:     "2000000000.246",
			expTime: time.Date(2033, 5, 18, 3, 33, 20, 246000000, time.Local),
		},
		{
			name:    "e with 3 fractional seconds",
			arg:     "e2000000000.246",
			expTime: time.Date(2033, 5, 18, 3, 33, 20, 246000000, time.Local),
		},
		{
			name:    "with 6 fractional seconds",
			arg:     "1357900000.086427",
			expTime: time.Date(2013, 1, 11, 10, 26, 40, 86427000, time.Local),
		},
		{
			name:    "e with 6 fractional seconds",
			arg:     "e1357900000.086427",
			expTime: time.Date(2013, 1, 11, 10, 26, 40, 86427000, time.Local),
		},
		{
			name:    "with 9 fractional seconds",
			arg:     "1876543209.000000002",
			expTime: time.Date(2029, 6, 19, 6, 0, 9, 2, time.Local),
		},
		{
			name:    "e with 9 fractional seconds",
			arg:     "e1876543209.000000002",
			expTime: time.Date(2029, 6, 19, 6, 0, 9, 2, time.Local),
		},
		{
			name:    "known",
			arg:     "981173106",
			expTime: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC).In(time.Local),
		},
		{
			name:    "e known",
			arg:     "e981173106",
			expTime: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC).In(time.Local),
		},
		{
			name:   "n known",
			arg:    "n981173106",
			expErr: "could not parse seconds from \"n981173106\": strconv.ParseInt: parsing \"n981173106\": invalid syntax",
		},
		{
			name:    "zero",
			arg:     "0",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).In(time.Local),
		},
		{
			name:    "e zero",
			arg:     "e0",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).In(time.Local),
		},
		{
			name:    "ends with decimal",
			arg:     "981173106.",
			expTime: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC).In(time.Local),
		},
		{
			name:    "e ends with decimal",
			arg:     "e981173106.",
			expTime: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC).In(time.Local),
		},
		{
			name:    "starts with decimal",
			arg:     ".123",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 123000000, time.UTC).In(time.Local),
		},
		{
			name:    "e starts with decimal",
			arg:     "e.123",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 123000000, time.UTC).In(time.Local),
		},
		{
			name:   "only decimal",
			arg:    ".",
			expErr: "invalid number: \".\"",
		},
		{
			name:   "e only decimal",
			arg:    "e.",
			expErr: "invalid number: \".\"",
		},
		{
			name:   "just e",
			arg:    "e",
			expErr: "no value provided after epoch designator 'e'",
		},
		{
			name:   "invalid whole number",
			arg:    "123four",
			expErr: "could not parse seconds from \"123four\": strconv.ParseInt: parsing \"123four\": invalid syntax",
		},
		{
			name:   "e invalid whole number",
			arg:    "e123four",
			expErr: "could not parse seconds from \"123four\": strconv.ParseInt: parsing \"123four\": invalid syntax",
		},
		{
			name:   "invalid fractional part",
			arg:    "123.4five6",
			expErr: "could not parse nanoseconds from \"123.4five6\": strconv.ParseInt: parsing \"4five6000\": invalid syntax",
		},
		{
			name:   "e invalid fractional part",
			arg:    "e123.4five6",
			expErr: "could not parse nanoseconds from \"123.4five6\": strconv.ParseInt: parsing \"4five6000\": invalid syntax",
		},
		{
			name:    "zero with fractional",
			arg:     "0.975",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 975000000, time.UTC).In(time.Local),
		},
		{
			name:    "e zero with fractional",
			arg:     "e0.975",
			expTime: time.Date(1970, 1, 1, 0, 0, 0, 975000000, time.UTC).In(time.Local),
		},
		{
			name:    "zero fractional",
			arg:     "123.000000000",
			expTime: time.Date(1970, 1, 1, 0, 2, 3, 0, time.UTC).In(time.Local),
		},
		{
			name:    "e zero fractional",
			arg:     "e123.000000000",
			expTime: time.Date(1970, 1, 1, 0, 2, 3, 0, time.UTC).In(time.Local),
		},
		{
			name:    "e one",
			arg:     "e1",
			expTime: time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC).In(time.Local),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actTime time.Time
			var err error
			testFunc := func() {
				actTime, err = ParseEpoch(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseEpoch(%q)", tc.arg)
			AssertEqualError(t, tc.expErr, err, "ParseEpoch(%q) error", tc.arg)
			AssertEqualTime(t, tc.expTime, actTime, "ParseEpoch(%q) result", tc.arg)
		})
	}
}

func TestParseDur(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		expDur time.Duration
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			expDur: 0,
		},
		{
			name:   "just digits",
			arg:    "55",
			expDur: 0,
			expErr: "invalid duration \"55\": time: missing unit in duration \"55\"",
		},
		{
			name:   "3h10m5.432s",
			arg:    "3h10m5.432s",
			expDur: time.Hour*3 + time.Minute*10 + time.Second*5 + time.Millisecond*432,
		},
		{
			name:   "-3h10m5.432s",
			arg:    "-3h10m5.432s",
			expDur: -1 * (time.Hour*3 + time.Minute*10 + time.Second*5 + time.Millisecond*432),
		},
		{
			name:   "1w",
			arg:    "1w",
			expDur: time.Hour * 24 * 7,
		},
		{
			name:   "-1w",
			arg:    "-1w",
			expDur: -1 * time.Hour * 24 * 7,
		},
		{
			name:   "1d",
			arg:    "1d",
			expDur: time.Hour * 24,
		},
		{
			name:   "-1d",
			arg:    "-1d",
			expDur: -1 * time.Hour * 24,
		},
		{
			name:   "1w5d",
			arg:    "1w5d",
			expDur: time.Hour * 24 * 12,
		},
		{
			name:   "-1w5d",
			arg:    "-1w5d",
			expDur: -1 * time.Hour * 24 * 12,
		},
		{
			name:   "2w3d5h10m",
			arg:    "2w3d5h10m",
			expDur: time.Hour*24*time.Duration(2*7+3) + time.Hour*5 + time.Minute*10,
		},
		{
			name:   "-2w3d5h10m",
			arg:    "-2w3d5h10m",
			expDur: -1 * (time.Hour*24*time.Duration(2*7+3) + time.Hour*5 + time.Minute*10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actDur time.Duration
			var err error
			testFunc := func() {
				actDur, err = ParseDur(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseDur(%q)", tc.arg)
			AssertEqualError(t, tc.expErr, err, "ParseDur(%q) error", tc.arg)
			assert.Equal(t, tc.expDur, actDur, "ParseDur(%q) duration\nExpected: %s,  Actual: %s", tc.arg, tc.expDur, actDur)
		})
	}
}

func TestParseNum(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		expNum int
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			expErr: "strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name:   "just n",
			arg:    "n",
			expErr: "strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name:   "zero",
			arg:    "0",
			expNum: 0,
		},
		{
			name:   "n zero",
			arg:    "n0",
			expNum: 0,
		},
		{
			name:   "e zero",
			arg:    "e0",
			expErr: "strconv.Atoi: parsing \"e0\": invalid syntax",
		},
		{
			name:   "one",
			arg:    "1",
			expNum: 1,
		},
		{
			name:   "n one",
			arg:    "n1",
			expNum: 1,
		},
		{
			name:   "e one",
			arg:    "e1",
			expErr: "strconv.Atoi: parsing \"e1\": invalid syntax",
		},
		{
			name:   "big number",
			arg:    "987654321",
			expNum: 987654321,
		},
		{
			name:   "n big number",
			arg:    "n987654321",
			expNum: 987654321,
		},
		{
			name:   "e big number",
			arg:    "e987654321",
			expErr: "strconv.Atoi: parsing \"e987654321\": invalid syntax",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actNum int
			var err error
			testFunc := func() {
				actNum, err = ParseNum(tc.arg)
			}
			require.NotPanics(t, testFunc, "ParseNum(%q)", tc.arg)
			AssertEqualError(t, tc.expErr, err, "ParseNum(%q) error", tc.arg)
			assert.Equal(t, tc.expNum, actNum, "ParseNum(%q) number", tc.arg)
		})
	}
}
