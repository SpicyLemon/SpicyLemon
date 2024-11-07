package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DTVal is a struct for holding either a time.Time or time.Duration or int.
type DTVal struct {
	Time *time.Time
	Dur  *time.Duration
	Num  *int
}

// NewTimeVal creates a new DTVal with the provided Time.
func NewTimeVal(t time.Time) *DTVal {
	return &DTVal{Time: &t}
}

// NewDurVal creates a new DTVal with the provided Duration.
func NewDurVal(d time.Duration) *DTVal {
	return &DTVal{Dur: &d}
}

// NewNumVal creates a new DTVal with the provided number.
func NewNumVal(i int) *DTVal {
	return &DTVal{Num: &i}
}

// IsTime returns true if this DTVal has a Time.
func (v *DTVal) IsTime() bool {
	return v != nil && v.Time != nil
}

// IsDur returns true if this DTVal has a Duration.
func (v *DTVal) IsDur() bool {
	return v != nil && v.Dur != nil
}

// IsNum returns true if this DTVal has a Number.
func (v *DTVal) IsNum() bool {
	return v != nil && v.Num != nil
}

// TimeString returns this DTVal's Time as a string using the default format (or "<nil>").
func (v *DTVal) TimeString() string {
	if v == nil || v.Time == nil {
		return NilStr
	}
	return v.Time.String()
}

// DurString returns this DTVal's Duration as a string (or "<nil>").
func (v *DTVal) DurString() string {
	if v == nil || v.Dur == nil {
		return NilStr
	}
	return v.Dur.String()
}

// NumString returns this DTVal's Number as a string (or "<nil>").
func (v *DTVal) NumString() string {
	if v == nil || v.Num == nil {
		return NilStr
	}
	return strconv.Itoa(*v.Num)
}

// String returns a string that represents this DTVal such that these strings can be used to test equality.
func (v *DTVal) String() string {
	if v == nil {
		return NilStr
	}
	parts := make([]string, 0, 3)
	if v.Time != nil {
		parts = append(parts, v.Time.String())
	}
	if v.Dur != nil {
		parts = append(parts, v.Dur.String())
	}
	if v.Num != nil {
		parts = append(parts, strconv.Itoa(*v.Num))
	}
	switch len(parts) {
	case 0:
		return EmptyStr
	case 1:
		return parts[0]
	}
	return "{" + strings.Join(parts, "|") + "}"
}

// Validate returns an error if there's something wrong with this DTVal.
func (v *DTVal) Validate() error {
	if v == nil {
		return errors.New("cannot be nil")
	}

	n := 0
	if v.Time != nil {
		n++
	}
	if v.Dur != nil {
		n++
	}
	if v.Num != nil {
		n++
	}

	if n == 0 {
		return errors.New("cannot be empty")
	}

	if n > 1 {
		return fmt.Errorf("can only have one of datetime (%s) or duration (%s) or number (%s)",
			v.TimeString(), v.DurString(), v.NumString())
	}

	return nil
}

// TypeString returns either "<time>" "<dur>" or "<num>" (or "<nil>" or "<empty>").
func (v *DTVal) TypeString() string {
	switch {
	case v == nil:
		return NilStr
	case v.Time != nil:
		return "<time>"
	case v.Dur != nil:
		return "<dur>"
	case v.Num != nil:
		return "<num>"
	}
	return EmptyStr
}

// FormattedString returns a string of this DTVal with some extra formatting applied.
// If it's a Number, this returns it as a string.
// If it's a Time, it's formatted using either OutputFormat or the single input format used (or default format).
// If it's a Duration, hours are converted to days and hours and ending zero-values are removed.
func (v *DTVal) FormattedString() string {
	if err := v.Validate(); err != nil {
		return fmt.Sprintf("invalid result: %v", err)
	}

	if v.Num != nil {
		verbosef("result is number")
		return strconv.Itoa(*v.Num)
	}

	if v.Time != nil {
		verbosef("result is datetime")
		var format string
		switch {
		case len(OutputFormat) > 0:
			format = OutputFormat
			verbosef("using requested format: %q", format)
		case InputFormat != nil:
			format = InputFormat.Format
			verbosef("using provided input format: %q", format)
		case len(UsedInputFormats) == 1 && UsedInputFormats[0].IsComplete():
			format = UsedInputFormats[0].Format
			verbosef("using same format as input: %s", UsedInputFormats[0])
		default:
			format = DtFmtDefault.Format
			verbosef("using default format: %s", DtFmtDefault)
		}
		return v.Time.Format(format)
	}

	if v.Dur != nil {
		// Start with the standard string, then we'll clean it up.
		dur := v.Dur.String()
		verbosef("result is duration: %q", dur)

		// Convert hours to days and hours.
		if parts := hourRx.FindStringSubmatch(dur); len(parts) == 2 {
			hours, err := strconv.Atoi(parts[1])
			if err == nil && hours >= 24 {
				days := hours / 24
				hours = hours % 24
				newStr := fmt.Sprintf("%dd%dh", days, hours)
				dur = strings.Replace(dur, parts[0], newStr, 1)
				verbosef("converted hours %q to days and hours %q, result is now %q", parts[0], newStr, dur)
			}
		}

		// Remove ending zero values.
		for {
			parts := endingZeroValueRx.FindStringSubmatch(dur)
			if len(parts) != 3 {
				break
			}
			dur = strings.TrimSuffix(dur, parts[2])
			verbosef("removed %q from the end, result is now %q", parts[2], dur)
		}
		if len(dur) == 0 {
			dur = "0s" // time.Duration.String() returns "0s" when the duration is zero.
			verbosef("result is now empty, switching to default %q", dur)
		}

		return dur
	}

	return fmt.Sprintf("unknown result type %s = %s "+v.TypeString(), v.String())
}

var (
	// dayRx is a regexp that matches the custom "d" (days) time-unit section in a duration string.
	dayRx = regexp.MustCompile(`([[:digit:]]+)d`)
	// weekRx is a regexp that matches the custom "w" (weeks) time-unit section in a duration string.
	weekRx = regexp.MustCompile(`([[:digit:]]+)w`)
	// hourRx is a regexp that matches the standard "h" hours time-unit section in a duration string.
	hourRx = regexp.MustCompile(`([[:digit:]]+)h`)
	// endingZeroValueRx is a regexp that matches zero-value time-units at the end of a duration string.
	endingZeroValueRx = regexp.MustCompile(`(^|[^[:digit:]])(0[^[:digit:]])$`)
)

// ParseDTVal attempts to convert an arg into either a datetime, epoch, duration, or int and returns it as a DTVal.
func ParseDTVal(arg string) (*DTVal, error) {
	if len(arg) == 0 {
		return nil, errors.New("empty value argument not allowed")
	}

	t, errT := ParseTime(arg)
	e, errE := ParseEpoch(arg)
	d, errD := ParseDur(arg)
	i, errI := ParseNum(arg)

	// Make bools for these to make stuff easier to read.
	var isT, isE, isD, isI bool
	okCount := 0
	if errT == nil {
		isT = true
		okCount++
	}
	if errE == nil {
		isE = true
		okCount++
	}
	if errD == nil {
		isD = true
		okCount++
	}
	if errI == nil {
		isI = true
		okCount++
	}

	if okCount == 1 {
		switch {
		case isT:
			return NewTimeVal(t), nil
		case isE:
			return NewTimeVal(e), nil
		case isD:
			return NewDurVal(d), nil
		case isI:
			return NewNumVal(i), nil
		default:
			panic(fmt.Errorf("unhandled single argument type for %q", arg))
		}
	}

	if isE && isI && okCount == 2 {
		if i <= 1_000_000 {
			return NewNumVal(i), nil
		}
		return NewTimeVal(e), nil
	}

	if okCount == 0 {
		// It's natural to use * for multiplication. But unescaped, the terminal will expand it with all the files in the dir.
		// If that happens, and there's more than a few files in the dir, the error message is unusable.
		// So, if the arg is more than 60 chars (enough for all standard datetime formats), just use a 1-line error message.
		errMain := fmt.Errorf("could not convert %q to either a datetime, epoch, duration, or number", arg)
		if len(arg) > 60 {
			return nil, errors.Join(errMain, errors.New("Did you use * instead of x?"))
		}
		return nil, errors.Join(
			errMain,
			errT, // This will be multi-line with the format name at the start of all but the first.
			fmt.Errorf("duration: %w", errD),
			fmt.Errorf("epoch: %w", errE),
			fmt.Errorf("number: %w", errI),
		)
	}

	parts := make([]string, 0, okCount)
	if isT {
		parts = append(parts, fmt.Sprintf("datetime (%s)", t))
	}
	if isE {
		parts = append(parts, fmt.Sprintf("epoch (%s)", e))
	}
	if isD {
		parts = append(parts, fmt.Sprintf("duration (%s)", d))
	}
	if isI {
		parts = append(parts, fmt.Sprintf("number (%d)", i))
	}
	return nil, fmt.Errorf("ambiguous argument %q: can either be a %s", arg, strings.Join(parts, " or "))
}

// ParseTime attempts to convert the provided arg to a Time using the entries of FormatParseOrder.
func ParseTime(arg string) (time.Time, error) {
	errs := make([]error, len(FormatParseOrder))
	var rv time.Time
	for i, nf := range FormatParseOrder {
		if nf.HasZone {
			rv, errs[i] = time.Parse(nf.Format, arg)
		} else {
			rv, errs[i] = time.ParseInLocation(nf.Format, arg, time.Local)
		}
		if errs[i] == nil {
			RecordUsedInputFormat(nf)
			return rv, nil
		}
		if !strings.Contains(errs[i].Error(), nf.Format) {
			errs[i] = fmt.Errorf("%w: format %q", errs[i], nf.Format)
		}
		errs[i] = fmt.Errorf("%s: %w", nf.Name, errs[i])
	}

	return rv, errors.Join(errs...)
}

// ParseEpoch parses an epoch string (with possible fractional seconds) into a time.Time.
func ParseEpoch(arg string) (time.Time, error) {
	if len(arg) == 0 {
		return time.Time{}, errors.New("empty string not allowed")
	}
	if arg == "e" {
		return time.Time{}, errors.New("no value provided after epoch designator 'e'")
	}
	if arg[0] == 'e' {
		arg = arg[1:]
	}
	parts := strings.Split(arg, ".")
	var err error
	var s, ns int64
	var haveS, haveNS bool
	if len(parts[0]) > 0 {
		s, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("could not parse seconds from %q: %w", arg, err)
		}
		haveS = true
	}
	if len(parts) > 1 && len(parts[1]) > 0 {
		ns, err = strconv.ParseInt((parts[1] + "000000000")[:9], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("could not parse nanoseconds from %q: %w", arg, err)
		}
		haveNS = true
	}
	if !haveS && !haveNS {
		return time.Time{}, fmt.Errorf("invalid number: %q", arg)
	}
	return time.Unix(s, ns), nil
}

// ParseDur extends time.ParseDuration to allow for the "d" (days) and "w" (weeks) time units.
func ParseDur(arg string) (time.Duration, error) {
	orig := arg
	var days, weeks int
	var err error
	if parts := weekRx.FindStringSubmatch(arg); len(parts) == 2 {
		weeks, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid weeks %q (in %q): %w", parts[0], orig, err)
		}
		arg = strings.Replace(arg, parts[0], "", 1)
	}
	if parts := dayRx.FindStringSubmatch(arg); len(parts) == 2 {
		days, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid days %q (in %q): %w", parts[0], orig, err)
		}
		arg = strings.Replace(arg, parts[0], "", 1)
	}

	wdDur := time.Hour * 24 * time.Duration(days+7*weeks)
	switch arg {
	case "":
		return wdDur, nil
	case "-":
		return -1 * wdDur, nil
	}

	baseDur, err := time.ParseDuration(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", orig, err)
	}

	if wdDur == 0 {
		return baseDur, nil
	}
	if baseDur >= 0 {
		return baseDur + wdDur, nil
	}
	return baseDur - wdDur, nil
}

// ParseNum parses a number, allowing it to be prefixed with 'n'.
func ParseNum(arg string) (int, error) {
	if len(arg) > 0 && arg[0] == 'n' {
		arg = arg[1:]
	}
	return strconv.Atoi(arg)
}
