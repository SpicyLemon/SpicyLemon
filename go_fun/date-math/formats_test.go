package main_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/SpicyLemon/date-math"
)

func TestPrintFormats(t *testing.T) {
	expInPrinted := []string{
		fmt.Sprintf("Formats (%d):", len(NamedFormatMap)),
		"* = possible input format",
	}
	nameLen := 13 // DateTimeZone2
	for name, nf := range NamedFormatMap {
		expInPrinted = append(expInPrinted, name+" = \""+nf.Format+"\"")
		for _, ponf := range FormatParseOrder {
			if name == ponf.Name {
				expInPrinted = append(expInPrinted, "* "+(strings.Repeat(" ", nameLen) + name)[len(name):])
				break
			}
		}
	}

	var w bytes.Buffer
	testFunc := func() {
		PrintFormats(&w)
	}
	require.NotPanics(t, testFunc, "PrintFormats(w)")
	printed := w.String()

	for _, exp := range expInPrinted {
		t.Run(exp, func(t *testing.T) {
			assert.Contains(t, printed, exp, "Expected: %q", exp)
		})
	}

	if t.Failed() {
		t.Logf("Full output:\n%s", printed)
	}
}

func TestNewNamedFormat(t *testing.T) {
	// Year: "2006" "06"
	// Month: "Jan" "January" "01" "1"
	// Day of the week: "Mon" "Monday"
	// Day of the month: "2" "_2" "02"
	// Day of the year: "__2" "002"
	// Hour: "15" "3" "03" (PM or AM)
	// Minute: "4" "04"
	// Second: "5" "05"
	// AM/PM mark: "PM"
	// timezone: "-0700" "-07:00" "-07" "-070000" "-07:00:00"
	//           "Z0700" "Z07:00" "Z07" "Z070000" "Z07:00:00"

	tests := []struct {
		testName string
		name     string
		format   string
		expNF    *NamedFormat
		expPanic string
		expGNF   *NamedFormat // entry from NamedFormatMap (if different from what's expected from NewNamedFormat).
	}{
		{
			testName: "date: year, month name, day of month",
			name:     "TestDate",
			format:   "2006 Jan 2",
			expNF:    &NamedFormat{Name: "TestDate", Format: "2006 Jan 2", HasDate: true},
		},
		{
			testName: "date: year, month num, day of month",
			name:     "TestDate",
			format:   "02 of 1 in 06",
			expNF:    &NamedFormat{Name: "TestDate", Format: "02 of 1 in 06", HasDate: true},
		},
		{
			testName: "date: year, day of year",
			name:     "TestDate",
			format:   "__2 2006",
			expNF:    &NamedFormat{Name: "TestDate", Format: "__2 2006", HasDate: true},
		},
		{
			testName: "date: year, long month name, and day",
			name:     "TestDate",
			format:   "January 2, 2006",
			expNF:    &NamedFormat{Name: "TestDate", Format: "January 2, 2006", HasDate: true},
		},
		{
			testName: "date: year and three digit day of year",
			name:     "TestDate",
			format:   "06002",
			expNF:    &NamedFormat{Name: "TestDate", Format: "06002", HasDate: true},
		},
		{
			testName: "date: just year and day of month",
			name:     "TestDate",
			format:   "_2 in some part of 2006",
			expNF:    &NamedFormat{Name: "TestDate", Format: "_2 in some part of 2006"},
		},
		// long month name and 002
		{
			testName: "time: 24 hour",
			name:     "TestTime",
			format:   "5 sec and 4 min past 1500",
			expNF:    &NamedFormat{Name: "TestTime", Format: "5 sec and 4 min past 1500", HasTime: true},
		},
		{
			testName: "time: 12 hour",
			name:     "TestTime",
			format:   "3:04:05 PM",
			expNF:    &NamedFormat{Name: "TestTime", Format: "3:04:05 PM", HasTime: true},
		},
		{
			testName: "time: missing hours",
			name:     "TestTime",
			format:   "04:05 PM",
			expNF:    &NamedFormat{Name: "TestTime", Format: "04:05 PM"},
		},
		{
			testName: "time: missing minutes",
			name:     "TestTime",
			format:   "03:  :05 PM",
			expNF:    &NamedFormat{Name: "TestTime", Format: "03:  :05 PM"},
		},
		{
			testName: "time: missing seconds",
			name:     "TestTime",
			format:   "15 + 04",
			expNF:    &NamedFormat{Name: "TestTime", Format: "15 + 04"},
		},
		{
			testName: "time: 12 hour without am/pm",
			name:     "TestTime",
			format:   "03:04:05",
			expNF:    &NamedFormat{Name: "TestTime", Format: "03:04:05"},
		},
		{
			testName: "only zone: number",
			name:     "TestZone",
			format:   "Z0700",
			expNF:    &NamedFormat{Name: "TestZone", Format: "Z0700", HasZone: true},
		},
		{
			testName: "only zone: name",
			name:     "TestZone",
			format:   "Mst",
			expNF:    &NamedFormat{Name: "TestZone", Format: "Mst", HasZone: true},
		},
		{
			testName: "only dow: short",
			name:     "TestDoW",
			format:   "Mon",
			expNF:    &NamedFormat{Name: "TestDoW", Format: "Mon", HasDoW: true},
		},
		{
			testName: "only dow: long",
			name:     "TestDoW",
			format:   "Monday",
			expNF:    &NamedFormat{Name: "TestDoW", Format: "Monday", HasDoW: true},
		},
		{
			testName: "date and time",
			name:     "Testing",
			format:   "06002 150405",
			expNF:    &NamedFormat{Name: "Testing", Format: "06002 150405", HasDate: true, HasTime: true},
		},
		{
			testName: "date and zone",
			name:     "Testing",
			format:   "2006-01-02 MST",
			expNF:    &NamedFormat{Name: "Testing", Format: "2006-01-02 MST", HasDate: true, HasZone: true},
		},
		{
			testName: "date and dow",
			name:     "Testing",
			format:   "Monday 02/01/06",
			expNF:    &NamedFormat{Name: "Testing", Format: "Monday 02/01/06", HasDate: true, HasDoW: true},
		},
		{
			testName: "time and zone",
			name:     "Testing",
			format:   "3:04:05Z070000 PM",
			expNF:    &NamedFormat{Name: "Testing", Format: "3:04:05Z070000 PM", HasTime: true, HasZone: true},
		},
		{
			testName: "time and dow",
			name:     "TestingTD",
			format:   "Mon at 15:04:05",
			expNF:    &NamedFormat{Name: "TestingTD", Format: "Mon at 15:04:05", HasTime: true, HasDoW: true},
		},
		{
			testName: "zone and dow",
			name:     "Testing",
			format:   "MST, a Mon",
			expNF:    &NamedFormat{Name: "Testing", Format: "MST, a Mon", HasZone: true, HasDoW: true},
		},
		{
			testName: "date and time and zone",
			name:     "Test3",
			format:   "2006x01x02x03x04x05xPMx-0700",
			expNF: &NamedFormat{Name: "Test3", Format: "2006x01x02x03x04x05xPMx-0700",
				HasDate: true, HasTime: true, HasZone: true},
		},
		{
			testName: "date and time and dow",
			name:     "Test3",
			format:   "Mon Jan 2, 2006 at 03:04:05.999 PM",
			expNF: &NamedFormat{Name: "Test3", Format: "Mon Jan 2, 2006 at 03:04:05.999 PM",
				HasDate: true, HasTime: true, HasDoW: true},
		},
		{
			testName: "date and zone and dow",
			name:     "Test3",
			format:   "Mon 06-Jan-02 Z07",
			expNF: &NamedFormat{Name: "Test3", Format: "Mon 06-Jan-02 Z07",
				HasDate: true, HasZone: true, HasDoW: true},
		},
		{
			testName: "time and zone and dow",
			name:     "Test3",
			format:   "15:04:05 -0700 on Mon",
			expNF: &NamedFormat{Name: "Test3", Format: "15:04:05 -0700 on Mon",
				HasTime: true, HasZone: true, HasDoW: true},
		},
		{
			testName: "date and time and zone and dow",
			name:     "TestAll",
			format:   "Jan 2, 2006 (a Monday) at 15:04:05 Z0700",
			expNF: &NamedFormat{Name: "TestAll", Format: "Jan 2, 2006 (a Monday) at 15:04:05 Z0700",
				HasDate: true, HasTime: true, HasZone: true, HasDoW: true},
		},
		{
			testName: "name already known: same format",
			name:     DtFmtStamp.Name,
			format:   DtFmtStamp.Format,
			expNF:    DtFmtStamp,
		},
		{
			testName: "name already known: different format",
			name:     DtFmtDateTime.Name,
			format:   DtFmtDateTime.Format + " -07:00",
			expPanic: "format names must be unique: \"" + DtFmtDateTime.Name + "\" " +
				"created with \"" + DtFmtDateTime.Format + " -07:00\" then \"" + DtFmtDateTime.Format + "\"",
			expGNF: DtFmtDateTime,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			defer ResetGlobalsFn()()

			if tc.expGNF == nil {
				tc.expGNF = tc.expNF
			}

			var actNF *NamedFormat
			testFunc := func() {
				actNF = NewNamedFormat(tc.name, tc.format)
			}

			if len(tc.expPanic) > 0 {
				require.PanicsWithError(t, tc.expPanic, testFunc, "NewNamedFormat(%q, %q)", tc.name, tc.format)
			} else {
				require.NotPanics(t, testFunc, "NewNamedFormat(%q, %q)", tc.name, tc.format)
			}

			assert.Equal(t, tc.expNF, actNF, "NewNamedFormat(%q, %q) result", tc.name, tc.format)
			mapNF := NamedFormatMap[tc.name]
			assert.Equal(t, tc.expGNF, mapNF, "NamedFormatMap[%q]", tc.name)
		})
	}
}

func TestNamedFormat_String(t *testing.T) {
	tests := []struct {
		name string
		nf   *NamedFormat
		exp  string
	}{
		{
			name: "nil",
			nf:   nil,
			exp:  NilStr,
		},
		{
			name: "empty",
			nf:   &NamedFormat{},
			exp:  "{()=\"\"}",
		},
		{
			name: "only date",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true},
			exp:  "{nnn(d)=\"fff\"}",
		},
		{
			name: "only time",
			nf:   &NamedFormat{Name: "nNn", Format: "fFf", HasTime: true},
			exp:  "{nNn(t)=\"fFf\"}",
		},
		{
			name: "only zone",
			nf:   &NamedFormat{Name: "NNNN", Format: "FFFF", HasZone: true},
			exp:  "{NNNN(z)=\"FFFF\"}",
		},
		{
			name: "only dow",
			nf:   &NamedFormat{Name: "NNNN", Format: "FFFF", HasDoW: true},
			exp:  "{NNNN(w)=\"FFFF\"}",
		},
		{
			name: "date and time",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasTime: true},
			exp:  "{nnn(dt)=\"fff\"}",
		},
		{
			name: "date and zone",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasZone: true},
			exp:  "{nnn(dz)=\"fff\"}",
		},
		{
			name: "date and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasDoW: true},
			exp:  "{nnn(dw)=\"fff\"}",
		},
		{
			name: "time and zone",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasTime: true, HasZone: true},
			exp:  "{nnn(tz)=\"fff\"}",
		},
		{
			name: "time and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasTime: true, HasDoW: true},
			exp:  "{nnn(tw)=\"fff\"}",
		},
		{
			name: "zone and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasZone: true, HasDoW: true},
			exp:  "{nnn(zw)=\"fff\"}",
		},
		{
			name: "date and time and zone",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasTime: true, HasZone: true},
			exp:  "{nnn(dtz)=\"fff\"}",
		},
		{
			name: "date and time and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasTime: true, HasDoW: true},
			exp:  "{nnn(dtw)=\"fff\"}",
		},
		{
			name: "date and zone and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasZone: true, HasDoW: true},
			exp:  "{nnn(dzw)=\"fff\"}",
		},
		{
			name: "time and zone and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasTime: true, HasZone: true, HasDoW: true},
			exp:  "{nnn(tzw)=\"fff\"}",
		},
		{
			name: "date and time and zone and dow",
			nf:   &NamedFormat{Name: "nnn", Format: "fff", HasDate: true, HasTime: true, HasZone: true, HasDoW: true},
			exp:  "{nnn(dtzw)=\"fff\"}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.nf.String()
			}
			require.NotPanics(t, testFunc, "String()")
			assert.Equal(t, tc.exp, act, "String() result")
		})
	}
}

func TestNamedFormat_IsComplete(t *testing.T) {
	tests := []struct {
		name string
		nf   *NamedFormat
		exp  bool
	}{
		{
			name: "nil",
			nf:   nil,
			exp:  false,
		},
		{
			name: "empty",
			nf:   &NamedFormat{},
			exp:  false,
		},
		{
			name: "only date",
			nf:   &NamedFormat{HasDate: true},
			exp:  false,
		},
		{
			name: "only time",
			nf:   &NamedFormat{HasTime: true},
			exp:  false,
		},
		{
			name: "only zone",
			nf:   &NamedFormat{HasZone: true},
			exp:  false,
		},
		{
			name: "only dow",
			nf:   &NamedFormat{HasDoW: true},
			exp:  false,
		},
		{
			name: "date and time",
			nf:   &NamedFormat{HasDate: true, HasTime: true},
			exp:  false,
		},
		{
			name: "date and zone",
			nf:   &NamedFormat{HasDate: true, HasZone: true},
			exp:  false,
		},
		{
			name: "date and dow",
			nf:   &NamedFormat{HasDate: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "time and zone",
			nf:   &NamedFormat{HasTime: true, HasZone: true},
			exp:  false,
		},
		{
			name: "time and dow",
			nf:   &NamedFormat{HasTime: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "zone and dow",
			nf:   &NamedFormat{HasZone: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "date and time and zone",
			nf:   &NamedFormat{HasDate: true, HasTime: true, HasZone: true},
			exp:  true,
		},
		{
			name: "date and time and dow",
			nf:   &NamedFormat{HasDate: true, HasTime: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "date and zone and dow",
			nf:   &NamedFormat{HasDate: true, HasZone: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "time and zone and dow",
			nf:   &NamedFormat{HasTime: true, HasZone: true, HasDoW: true},
			exp:  false,
		},
		{
			name: "date and time and zone and dow",
			nf:   &NamedFormat{HasDate: true, HasTime: true, HasZone: true, HasDoW: true},
			exp:  true,
		},
		{
			name: "DtFmtDateTimeZone",
			nf:   DtFmtDateTimeZone,
			exp:  true,
		},
		{
			name: "DtFmtDateTimeZone2",
			nf:   DtFmtDateTimeZone2,
			exp:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.nf.IsComplete()
			}
			require.NotPanics(t, testFunc, "%s.IsComplete()", tc.nf)
			assert.Equal(t, tc.exp, act, "%s.IsComplete() result", tc.nf)
		})
	}
}

func TestNamedFormat_EqualName(t *testing.T) {
	tests := []struct {
		name string
		f    *NamedFormat
		g    *NamedFormat
		exp  bool
	}{
		{
			name: "nil v nil",
			f:    nil,
			g:    nil,
			exp:  false,
		},
		{
			name: "nil v empty",
			f:    nil,
			g:    &NamedFormat{Name: ""},
			exp:  false,
		},
		{
			name: "empty v nil",
			f:    &NamedFormat{Name: ""},
			g:    nil,
			exp:  false,
		},
		{
			name: "empty v empty",
			f:    &NamedFormat{Name: ""},
			g:    &NamedFormat{Name: ""},
			exp:  true,
		},
		{
			name: "empty v not",
			f:    &NamedFormat{Name: ""},
			g:    &NamedFormat{Name: "not"},
			exp:  false,
		},
		{
			name: "not v empty",
			f:    &NamedFormat{Name: "not"},
			g:    &NamedFormat{Name: ""},
			exp:  false,
		},
		{
			name: "different names",
			f:    &NamedFormat{Name: "name1"},
			g:    &NamedFormat{Name: "name2"},
			exp:  false,
		},
		{
			name: "same names",
			f:    &NamedFormat{Name: "Bananas"},
			g:    &NamedFormat{Name: "Bananas"},
			exp:  true,
		},
		{
			name: "same names, different casing",
			f:    &NamedFormat{Name: "Bananas"},
			g:    &NamedFormat{Name: "bananas"},
			exp:  false,
		},
		{
			name: "first contains second",
			f:    &NamedFormat{Name: "Bananas"},
			g:    &NamedFormat{Name: "anana"},
			exp:  false,
		},
		{
			name: "second contains first",
			f:    &NamedFormat{Name: "anana"},
			g:    &NamedFormat{Name: "Bananas"},
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.f.EqualName(tc.g)
			}
			require.NotPanics(t, testFunc, "%s.EqualName(%s)", tc.f, tc.g)
			assert.Equal(t, tc.exp, act, "%s.EqualName(%s) result", tc.f, tc.g)
		})
	}
}

func TestRecordInputUsedFormat(t *testing.T) {
	tests := []struct {
		name    string
		iniUsed []*NamedFormat
		nf      *NamedFormat
		expUsed []*NamedFormat
	}{
		{
			name:    "no used yet",
			iniUsed: nil,
			nf:      DtFmtRFC822,
			expUsed: []*NamedFormat{DtFmtRFC822},
		},
		{
			name:    "one used, new is same",
			iniUsed: []*NamedFormat{DtFmtLayout},
			nf:      &NamedFormat{Name: DtFmtLayout.Name},
			expUsed: []*NamedFormat{DtFmtLayout},
		},
		{
			name:    "one used, new is new",
			iniUsed: []*NamedFormat{DtFmtRFC822Z},
			nf:      DtFmtKitchen,
			expUsed: []*NamedFormat{DtFmtRFC822Z, DtFmtKitchen},
		},
		{
			name:    "three used, new is first",
			iniUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
			nf:      DtFmtStampMilli,
			expUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
		},
		{
			name:    "three used, new is second",
			iniUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
			nf:      DtFmtStampMicro,
			expUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
		},
		{
			name:    "three used, new is third",
			iniUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
			nf:      DtFmtStampNano,
			expUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
		},
		{
			name:    "three used, new is new",
			iniUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano},
			nf:      DtFmtDateOnly,
			expUsed: []*NamedFormat{DtFmtStampMilli, DtFmtStampMicro, DtFmtStampNano, DtFmtDateOnly},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer ResetGlobalsFn()()
			UsedInputFormats = tc.iniUsed

			testFunc := func() {
				RecordUsedInputFormat(tc.nf)
			}
			require.NotPanics(t, testFunc, "RecordUsedInputFormat(%s)", tc.nf)
			assert.Equal(t, tc.expUsed, UsedInputFormats, "UsedInputFormats")
		})
	}
}

func TestGetFormatByName(t *testing.T) {
	tests := []struct {
		name   string
		toFind string
		expNF  *NamedFormat
	}{
		{
			name:   "known name with same case",
			toFind: "TimeOnly",
			expNF:  DtFmtTimeOnly,
		},
		{
			name:   "known name with different case",
			toFind: "default",
			expNF:  DtFmtDefault,
		},
		{
			name:   "unknown",
			toFind: "bananas",
			expNF:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actNF *NamedFormat
			testFunc := func() {
				actNF = GetFormatByName(tc.toFind)
			}
			require.NotPanics(t, testFunc, "GetFormatByName(%s)", tc.toFind)
			assert.Equal(t, tc.expNF, actNF, "GetFormatByName(%s) result", tc.toFind)
		})
	}
}

func TestFormatHasNameFn(t *testing.T) {
	type expResult struct {
		nf  *NamedFormat
		exp bool
	}
	tests := []struct {
		testName string
		name     string
		results  []expResult
	}{
		{
			testName: "empty name",
			name:     "",
			results: []expResult{
				{nf: nil, exp: false},
				{nf: DtFmtStamp, exp: false},
				{nf: &NamedFormat{}, exp: true},
			},
		},
		{
			testName: "all lower-case name",
			name:     "bananas",
			results: []expResult{
				{nf: nil, exp: false},
				{nf: DtFmtRFC850, exp: false},
				{nf: &NamedFormat{Name: "bananas"}, exp: true},
				{nf: &NamedFormat{Name: "BANANAS"}, exp: true},
				{nf: &NamedFormat{Name: "Bananas"}, exp: true},
				{nf: &NamedFormat{Name: "baNaNas"}, exp: true},
			},
		},
		{
			testName: "all upper-case name",
			name:     "BANANAS",
			results: []expResult{
				{nf: nil, exp: false},
				{nf: DtFmtANSIC, exp: false},
				{nf: &NamedFormat{Name: "bananas"}, exp: true},
				{nf: &NamedFormat{Name: "BANANAS"}, exp: true},
				{nf: &NamedFormat{Name: "Bananas"}, exp: true},
				{nf: &NamedFormat{Name: "baNaNas"}, exp: true},
			},
		},
		{
			testName: "mixed case name",
			name:     "bAnAnAs",
			results: []expResult{
				{nf: nil, exp: false},
				{nf: DtFmtDateTime, exp: false},
				{nf: &NamedFormat{Name: "bananas"}, exp: true},
				{nf: &NamedFormat{Name: "BANANAS"}, exp: true},
				{nf: &NamedFormat{Name: "Bananas"}, exp: true},
				{nf: &NamedFormat{Name: "baNaNas"}, exp: true},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var checker func(*NamedFormat) bool
			testFunc := func() {
				checker = FormatHasNameFn(tc.name)
			}
			require.NotPanics(t, testFunc, "FormatHasNameFn(%q)", tc.name)

			for i, result := range tc.results {
				var act bool
				testFunc = func() {
					act = checker(result.nf)
				}
				require.NotPanics(t, testFunc, "[%d]: FormatHasNameFn(%q)(%s)", i, tc.name, result.nf)
				assert.Equal(t, result.exp, act, "[%d]: FormatHasNameFn(%q)(%s) result", i, tc.name, result.nf)
			}
		})
	}
}
