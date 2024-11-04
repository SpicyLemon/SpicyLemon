package main

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	DtFmtDefault = NewNamedFormat("Default", "2006-01-02 15:04:05.999999999 -0700 MST") // Same as what's used in Time.String().

	DtFmtLayout      = NewNamedFormat("Layout", time.Layout)           // time.Layout      = "01/02 03:04:05PM '06 -0700"
	DtFmtANSIC       = NewNamedFormat("ANSIC", time.ANSIC)             // time.ANSIC       = "Mon Jan _2 15:04:05 2006"
	DtFmtUnixDate    = NewNamedFormat("UnixDate", time.UnixDate)       // time.UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	DtFmtRubyDate    = NewNamedFormat("RubyDate", time.RubyDate)       // time.RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	DtFmtRFC822      = NewNamedFormat("RFC822", time.RFC822)           // time.RFC822      = "02 Jan 06 15:04 MST"
	DtFmtRFC822Z     = NewNamedFormat("RFC822Z", time.RFC822Z)         // time.RFC822Z     = "02 Jan 06 15:04 -0700"
	DtFmtRFC850      = NewNamedFormat("RFC850", time.RFC850)           // time.RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	DtFmtRFC1123     = NewNamedFormat("RFC1123", time.RFC1123)         // time.RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	DtFmtRFC1123Z    = NewNamedFormat("RFC1123Z", time.RFC1123Z)       // time.RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700"
	DtFmtRFC3339     = NewNamedFormat("RFC3339", time.RFC3339)         // time.RFC3339     = "2006-01-02T15:04:05Z07:00"
	DtFmtRFC3339Nano = NewNamedFormat("RFC3339Nano", time.RFC3339Nano) // time.RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	DtFmtKitchen     = NewNamedFormat("Kitchen", time.Kitchen)         // time.Kitchen     = "3:04PM"
	DtFmtStamp       = NewNamedFormat("Stamp", time.Stamp)             // time.Stamp       = "Jan _2 15:04:05"
	DtFmtStampMilli  = NewNamedFormat("StampMilli", time.StampMilli)   // time.StampMilli  = "Jan _2 15:04:05.000"
	DtFmtStampMicro  = NewNamedFormat("StampMicro", time.StampMicro)   // time.StampMicro  = "Jan _2 15:04:05.000000"
	DtFmtStampNano   = NewNamedFormat("StampNano", time.StampNano)     // time.StampNano   = "Jan _2 15:04:05.000000000"
	DtFmtDateTime    = NewNamedFormat("DateTime", time.DateTime)       // time.DateTime    = "2006-01-02 15:04:05"
	DtFmtDateOnly    = NewNamedFormat("DateOnly", time.DateOnly)       // time.DateOnly    = "2006-01-02"
	DtFmtTimeOnly    = NewNamedFormat("TimeOnly", time.TimeOnly)       // time.TimeOnly    = "15:04:05"

	DtFmtDateTimeZone  = NewNamedFormat("DateTimeZone", "2006-01-02 15:04:05.999999999 -0700")
	DtFmtDateTimeZone2 = NewNamedFormat("DateTimeZone2", "2006-01-02 15:04:05.999999999Z0700")

	FormatParseOrder = []*NamedFormat{
		DtFmtDateTimeZone,
		DtFmtDateTimeZone2,
		DtFmtUnixDate,
		DtFmtRFC3339Nano,
		DtFmtDateTime,
		DtFmtRFC1123,
		DtFmtRFC1123Z,
		DtFmtRubyDate,
		DtFmtANSIC,
		DtFmtRFC850,
	}

	// OutputFormat is the output format that has been requested via args.
	OutputFormat string

	// InputFormat is the user-supplied input format.
	InputFormat *NamedFormat
)

// PrintFormats will print out the names and format strings of all the formats to the provided writer (e.g. os.Stdout).
func PrintFormats(stdout io.Writer) {
	names := make([]string, 0, len(NamedFormatMap))
	nameLen := 0
	for name := range NamedFormatMap {
		names = append(names, name)
		if len(name) > nameLen {
			nameLen = len(name)
		}
	}
	slices.Sort(names)
	nw := strconv.Itoa(nameLen)

	fmt.Fprintf(stdout, "Formats (%d): * = possible input format\n", len(names))
	for i, name := range names {
		nf := NamedFormatMap[name]
		parseInd := " "
		if slices.ContainsFunc(FormatParseOrder, FormatHasNameFn(name)) {
			parseInd = "*"
		}
		fmt.Fprintf(stdout, "%4d: %s %"+nw+"s = %q\n", i+1, parseInd, nf.Name, nf.Format)
	}
}

// NamedFormat associates a name with a format.
type NamedFormat struct {
	// Name is the name we give to this format.
	Name string
	// Format is the string used for formatting and parsing datetimes.
	Format string
	// HasDate indicates whether the format has either year, month, day; or year and day-of-year.
	HasDate bool
	// HasTime indicates whether the format has hours, minutes, and seconds.
	HasTime bool
	// HasZone indicates whether the format has a timezone.
	HasZone bool
	// HasDoW indicates whether the format has the day of the week.
	HasDoW bool
}

// NamedFormatMap maps a format name string with the corresponding NamedFormat.
var NamedFormatMap = make(map[string]*NamedFormat)

// fmtSecondRx is a regexp that will match "5" or "05", but not "15".
var fmtSecondRx = regexp.MustCompile(`(^|[^1])5`)

// NewNamedFormat creates a new NamedFormat with the given name and format and adds it to NamedFormatMap.
func NewNamedFormat(name, format string) *NamedFormat {
	rv := makeNamedFormat(name, format)
	if nf, known := NamedFormatMap[name]; known && nf.Format != format {
		panic(fmt.Errorf("format names must be unique: %q created with %q then %q", name, format, nf.Format))
	}
	NamedFormatMap[name] = rv
	return rv
}

// makeNamedFormat creates a new NamedFormat.
func makeNamedFormat(name, format string) *NamedFormat {
	lcFmt := strings.ToLower(format)
	return &NamedFormat{
		Name:   name,
		Format: format,
		// year and day and either a month or day-of-year (the day-of-year doesn't have to be different from the first day we found).
		HasDate: HasAllOf(lcFmt, "6", "2") && HasOneOf(lcFmt, "1", "jan", "002", "__2"),
		// hours (24 or 12 w/pm), minutes, and seconds.
		HasTime: (HasOneOf(lcFmt, "15") || HasAllOf(lcFmt, "3", "pm")) && HasOneOf(lcFmt, "4") && fmtSecondRx.MatchString(lcFmt),
		HasZone: HasOneOf(lcFmt, "7", "mst"),
		HasDoW:  HasOneOf(lcFmt, "mon"),
	}
}

// String returns a string representation of this named format.
func (f *NamedFormat) String() string {
	if f == nil {
		return NilStr
	}
	return fmt.Sprintf("{%s(%s%s%s%s)=%q}",
		f.Name, StrIf(f.HasDate, "d"), StrIf(f.HasTime, "t"), StrIf(f.HasZone, "z"), StrIf(f.HasDoW, "w"), f.Format)
}

// IsComplete returns true if this format has a timezone, and uniquely identifies a date and time to at least seconds.
func (f *NamedFormat) IsComplete() bool {
	return f != nil && f.HasZone && f.HasDate && f.HasTime
}

// EqualName returns true if the name of this NamedFormat equals the name of the one provided.
func (f *NamedFormat) EqualName(g *NamedFormat) bool {
	if f == nil || g == nil {
		return false
	}
	return f.Name == g.Name
}

// UsedInputFormats is a list of all the formats successfully used to parse datetime input arguments.
var UsedInputFormats []*NamedFormat

// RecordUsedInputFormat adds the provided format to UsedInputFormats if it's not already in there.
func RecordUsedInputFormat(nf *NamedFormat) {
	verboseStepf(stepFormat, "%s = %q", nf.Name, nf.Format)
	if !slices.ContainsFunc(UsedInputFormats, nf.EqualName) {
		UsedInputFormats = append(UsedInputFormats, nf)
	}
}

// GetFormatByName will get a format from the NamedFormatMap with the given name (ignoring case).
func GetFormatByName(toFind string) *NamedFormat {
	trimmed := strings.TrimSpace(toFind)
	for name, nf := range NamedFormatMap {
		if strings.EqualFold(trimmed, name) {
			return nf
		}
	}
	return nil
}

// FormatHasNameFn returns af function that will return true if a NamedFormat has the provided name (ignoring case).
func FormatHasNameFn(name string) func(*NamedFormat) bool {
	return func(nf *NamedFormat) bool {
		return nf != nil && strings.EqualFold(nf.Name, name)
	}
}
