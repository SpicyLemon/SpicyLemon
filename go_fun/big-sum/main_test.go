package main

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMakeNumbers is here to help generate numbers for testing.
// It doesn't check anything, it just logs a bunch of stuff.
func TestMakeNumbers(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	nums := newNumberSet(r, 50)
	for _, set := range nums.AsNamedArgs() {
		logTestNumbers(t, set.Name, set.Args)
	}
}

// makeNums runs the provided maker count times and returns all the results.
func makeNums(count int, maker func() string) []string {
	rv := make([]string, count)
	for i := range rv {
		rv[i] = maker()
	}
	return rv
}

// logTestNumbers will output the provided nums to the test log as quoted strings, comma separated, 5 per line.
func logTestNumbers(t *testing.T, title string, nums []string) {
	perLine := 5
	var toLog strings.Builder
	numsInCurLine := 0
	for _, num := range nums {
		if numsInCurLine == perLine {
			toLog.WriteRune('\n')
			numsInCurLine = 0
		}
		if numsInCurLine != 0 {
			toLog.WriteRune(' ')
		}
		toLog.WriteString(fmt.Sprintf("%q", num))
		toLog.WriteRune(',')
		numsInCurLine++
	}
	t.Logf("%s (%d): []string{\n%s\n}", title, len(nums), toLog.String())
}

// digits are the runes available for numbers.
var digits = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

// randomFloatString generates a random string representing a floating point number.
// It will have between 1 and maxWholeDigits (inclusive) digits to the left of the decimal.
// It will have between 1 and maxFractionalDigits (inclusive) digits to the right of the decimal.
// If maxWholeDigits is 0, there's a 50% chance that either a '0' is used, or nothing.
// If maxFractionalDigits is 0, the strings will all end with a decimal.
func randomFloatString(r *rand.Rand, maxWholeDigits, maxFractionalDigits int, negative bool) string {
	var whole, fractional string
	switch {
	case maxWholeDigits > 0:
		whole = randomIntString(r, maxWholeDigits, negative)
	case randBool(r):
		if negative {
			whole = "-0"
		} else {
			whole = "0"
		}
	case negative:
		whole = "-"
	}
	if maxFractionalDigits > 0 {
		fractional = randDigits(r, r.Intn(maxFractionalDigits)+1)
	}
	return whole + "." + fractional
}

// randomIntString generates a random string representing an integer.
// It will have between 1 and maxWholeDigits (inclusive) digits.
// If maxWholeDigits is zero, an empty string is returned.
func randomIntString(r *rand.Rand, maxWholeDigits int, negative bool) string {
	if maxWholeDigits <= 0 {
		return ""
	}
	// We don't want the first digit to be zero, so we handle it specially.
	length := r.Intn(maxWholeDigits) + 1
	first := string(digits[r.Intn(len(digits)-1)+1]) // Assumes '0' is the 0th element.
	if negative {
		first = "-" + first
	}
	if length == 1 {
		return first
	}
	return first + randDigits(r, length-1)
}

// randDigits creates a string of the provided length filled with random digits.
func randDigits(r *rand.Rand, length int) string {
	if length == 0 {
		return ""
	}
	rv := make([]rune, length)
	for i := range rv {
		rv[i] = digits[r.Intn(len(digits))]
	}
	return string(rv)
}

// randBool has a 50% chance of returning true.
func randBool(r *rand.Rand) bool {
	return r.Intn(2) == 0
}

// numberSetMakers are a set of functions used to create a value in each of the sets of numbers.
type numberSetMakers struct {
	FloatsSmallPositive func() string
	FloatsSmallNegative func() string
	FloatsSmallMixed    func() string
	FloatsLargePositive func() string
	FloatsLargeNegative func() string
	FloatsLargeMixed    func() string
	IntsPositive        func() string
	IntsNegative        func() string
	IntsMixed           func() string
	MixedNumbers        func() string
	MixedPositive       func() string
	MixedNegative       func() string
}

// newNumberSetMakers returns the standard set of number makers.
func newNumberSetMakers(r *rand.Rand) *numberSetMakers {
	fsw, fsf := 0, 30 // "Floats Small [Whole|Fractional]" digit count maxes
	flw, flf := 25, 5 // "Floats Large [Whole|Fractional]" digit count maxes
	iw := 30          // "Integers Whole" digit count max.
	rv := &numberSetMakers{
		FloatsSmallPositive: func() string { return randomFloatString(r, fsw, fsf, false) },
		FloatsSmallNegative: func() string { return randomFloatString(r, fsw, fsf, true) },
		FloatsSmallMixed:    func() string { return randomFloatString(r, fsw, fsf, randBool(r)) },
		FloatsLargePositive: func() string { return randomFloatString(r, flw, flf, false) },
		FloatsLargeNegative: func() string { return randomFloatString(r, flw, flf, true) },
		FloatsLargeMixed:    func() string { return randomFloatString(r, flw, flf, randBool(r)) },
		IntsPositive:        func() string { return randomIntString(r, iw, false) },
		IntsNegative:        func() string { return randomIntString(r, iw, true) },
		IntsMixed:           func() string { return randomIntString(r, iw, randBool(r)) },
	}
	rv.MixedNumbers = newMixedNumbersFunc(r, rv.FloatsSmallMixed, rv.FloatsLargeMixed, rv.IntsMixed)
	rv.MixedPositive = newMixedNumbersFunc(r, rv.FloatsSmallPositive, rv.FloatsLargePositive, rv.IntsPositive)
	rv.MixedNegative = newMixedNumbersFunc(r, rv.FloatsSmallNegative, rv.FloatsLargeNegative, rv.IntsNegative)
	return rv
}

// newMixedNumbersFunc returns a function that randomly chooses one of the provided funcs, runs it and returns the result.
func newMixedNumbersFunc(r *rand.Rand, funcs ...func() string) func() string {
	return func() string {
		return funcs[r.Intn(len(funcs))]()
	}
}

// numberSet contains several different categories of args used for testing the sum functions.
type numberSet struct {
	FloatsSmallPositive []string
	FloatsSmallNegative []string
	FloatsSmallMixed    []string
	FloatsLargePositive []string
	FloatsLargeNegative []string
	FloatsLargeMixed    []string
	IntsPositive        []string
	IntsNegative        []string
	IntsMixed           []string
	MixedNumbers        []string
}

// newNumberSet randomly generates a new numberSet, each with count entries.
func newNumberSet(r *rand.Rand, count int) *numberSet {
	makers := newNumberSetMakers(r)
	rv := &numberSet{
		FloatsSmallPositive: makeNums(count, makers.FloatsSmallPositive),
		FloatsSmallNegative: makeNums(count, makers.FloatsSmallNegative),
		FloatsSmallMixed:    makeNums(count, makers.FloatsSmallMixed),
		FloatsLargePositive: makeNums(count, makers.FloatsLargePositive),
		FloatsLargeNegative: makeNums(count, makers.FloatsLargeNegative),
		FloatsLargeMixed:    makeNums(count, makers.FloatsLargeMixed),
		IntsPositive:        makeNums(count, makers.IntsPositive),
		IntsNegative:        makeNums(count, makers.IntsNegative),
		IntsMixed:           makeNums(count, makers.IntsMixed),
		MixedNumbers:        makeNums(count, makers.MixedNumbers),
	}

	rv.FloatsSmallMixed = ensurePosNeg(rv.FloatsSmallMixed, makers.FloatsSmallPositive, makers.FloatsSmallNegative)
	rv.FloatsLargeMixed = ensurePosNeg(rv.FloatsLargeMixed, makers.FloatsLargePositive, makers.FloatsLargeNegative)
	rv.IntsMixed = ensurePosNeg(rv.IntsMixed, makers.IntsPositive, makers.IntsNegative)
	rv.MixedNumbers = ensurePosNeg(rv.MixedNumbers, makers.MixedPositive, makers.MixedNegative)

	return rv
}

// hasPosNeg looks through the provided vals to see if it has positive and/or negative numbers.
func hasPosNeg(vals []string) (havePos, haveNeg bool) {
	for _, val := range vals {
		if strings.HasPrefix(val, "-") {
			haveNeg = true
		} else {
			havePos = true
		}
		if havePos && haveNeg {
			return
		}
	}
	return
}

// ensurePosNeg will check the provided vals to make sure they have both positive and negative numbers.
// If not, the last entry of vals is replaced with a newly generated value using the appropriate maker.
// If vals doesn't have at least 2 entries, vals is returned unaltered.
func ensurePosNeg(vals []string, posMaker, negMaker func() string) []string {
	if len(vals) < 2 {
		return vals
	}
	havePos, haveNeg := hasPosNeg(vals)
	if havePos && haveNeg {
		return vals
	}

	rv := make([]string, len(vals))
	copy(rv, vals)
	if !havePos {
		rv[len(rv)-1] = posMaker()
	}
	if !haveNeg {
		rv[len(rv)-1] = negMaker()
	}
	return rv
}

// namedArgs gives a name to a set of args.
type namedArgs struct {
	Name  string
	Args  []string
	Mixed bool
}

// AsNamedArgs converts this numberSet to a slice of namedArgs.
func (n *numberSet) AsNamedArgs() []*namedArgs {
	return []*namedArgs{
		{Name: "Positive Small Floats", Args: n.FloatsSmallPositive},
		{Name: "Negative Small Floats", Args: n.FloatsSmallNegative},
		{Name: "Mixed Small Floats", Args: n.FloatsSmallMixed, Mixed: true},
		{Name: "Positive Big Floats", Args: n.FloatsLargePositive},
		{Name: "Negative Big Floats", Args: n.FloatsLargeNegative},
		{Name: "Mixed Big Floats", Args: n.FloatsLargeMixed, Mixed: true},
		{Name: "Positive Integers", Args: n.IntsPositive},
		{Name: "Negative Integers", Args: n.IntsNegative},
		{Name: "Mixed Integers", Args: n.IntsMixed, Mixed: true},
		{Name: "Mixed Numbers", Args: n.MixedNumbers, Mixed: true},
	}
}

// SubSet returns a subset of the args in this namedArgs set.
// The first element will be Args[offset%len(Args)], taking count args total, wrapping back
// to the beginning of Args if needed. If these namedArgs are mixed, the returned args will
// also be mixed by replacing the last entry with the next one from Args with the needed sign.
// Contract: count cannot be more than len(Args) or the test is failed (or panics if tb is nil).
func (a *namedArgs) SubSet(tb testing.TB, offset, count int) []string {
	if count > len(a.Args) {
		msg := fmt.Sprintf("%q.SubSet(t, %d, %d): count is greater than the number of args available (%d)", a.Name, offset, count, len(a.Args))
		if tb != nil {
			tb.Fatal(msg)
		} else {
			panic(msg)
		}
	}
	offset = offset % len(a.Args)
	if count == len(a.Args) {
		if offset == 0 {
			return a.Args
		}
		rv := make([]string, len(a.Args))
		a1 := copy(rv, a.Args[offset:])
		a2 := copy(rv[len(a.Args)-offset:], a.Args)
		if tb != nil {
			require.Equal(tb, count, a1+a2, "%q.SubSet(t, %d, %d): args copied to rv.\nactual = %d + %d", a.Name, offset, count, a1, a2)
		}
		return rv
	}

	rv := make([]string, count)
	copied := copy(rv, a.Args[offset:])
	end := offset + count
	if end > len(a.Args) {
		copied += copy(rv, a.Args[:end%len(a.Args)])
	}
	if tb != nil {
		require.Equal(tb, count, copied, "%q.SubSet(t, %d, %d): args copied to rv", a.Name, offset, count)
	}
	if !a.Mixed {
		return rv
	}

	havePos, haveNeg := hasPosNeg(rv)
	if havePos && haveNeg {
		return rv
	}

	if !havePos {
		for i := range a.Args {
			val := a.Args[(i+offset+end)%len(a.Args)]
			if !strings.HasPrefix(val, "-") {
				rv[count-1] = val
				break
			}
		}
	}

	if !haveNeg {
		for i := range a.Args {
			val := a.Args[(i+offset+end)%len(a.Args)]
			if strings.HasPrefix(val, "-") {
				rv[count-1] = val
				break
			}
		}
	}

	return rv
}

func BenchmarkSumFuncs(b *testing.B) {
	// -893514434144017 chosen randomly, but hard-coded for test consistency.
	numbers := newNumberSet(rand.New(rand.NewSource(-893514434144017)), 10_000)

	argSets := numbers.AsNamedArgs()
	counts := []int{2, 5, 10, 1000, 10_000}
	sumFuncs := []struct {
		name string
		f    func(args []string) (string, error)
	}{
		// {name: "Sum1", f: Sum1},
		{name: "Sum2", f: Sum2},
		{name: "Sum3", f: Sum3},
		{name: "Sum4", f: Sum4},
		{name: "Sum5", f: Sum5},
	}

	type testCase struct {
		name    string
		sumFunc func(args []string) (string, error)
		args    []string
	}
	tests := make([]testCase, 0, len(argSets)*len(counts)*len(sumFuncs))

	for _, args := range argSets {
		offset := 0
		for _, count := range counts {
			for _, sumFunc := range sumFuncs {
				tests = append(tests, testCase{
					name:    fmt.Sprintf("%s %d %s", args.Name, count, sumFunc.name),
					sumFunc: sumFunc.f,
					args:    args.SubSet(b, offset, count),
				})
			}
			offset += count
		}
	}

	b.ResetTimer()
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			Verbose = false
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				oneFloat, oneFloatNeg, fiveFloat, fiveFloatNeg = nil, nil, nil, nil
				_, _ = tc.sumFunc(tc.args)
			}
			b.StopTimer()
		})
	}
}

// RunSumFuncTests runs a set of tests using the provided Sum function.
// If checkErrContents is true, the errors must have the right content.
// If checkErrContents is false, this only checks whether there was an error.
func RunSumFuncTests(t *testing.T, name string, sumFunc func(args []string) (string, error), checkErrContents bool) {
	tests := []struct {
		name     string
		args     []string
		exp      string
		expInErr []string
		verbose  bool // Set this to true to get verbose output in the test logs for a given test.
	}{
		{
			name: "nil args",
			args: nil,
			exp:  "0",
		},
		{
			name: "empty args",
			args: []string{},
			exp:  "0",
		},
		{
			name: "one arg: empty string",
			args: []string{""},
			exp:  "0",
		},
		{
			name:     "one arg: no decimal, not an int",
			args:     []string{"123nope4"},
			expInErr: []string{"could not parse \"123nope4\" as integer"},
		},
		{
			name:     "one arg: with decimal, bad integer part",
			args:     []string{"87ba.d4"},
			expInErr: []string{"could not parse \"87ba.d4\" as float"},
		},
		{
			name:     "one arg: with decimal, bad fractional part",
			args:     []string{"87.3d4"},
			expInErr: []string{"could not parse \"87.3d4\" as float"},
		},
		{
			name: "one arg: int",
			args: []string{"12345678900987654321"},
			exp:  "12345678900987654321",
		},
		{
			name: "one arg: int with commas",
			args: []string{"12,345,678,900,987,654,321"},
			exp:  "12345678900987654321",
		},
		{
			name: "one arg: int with underscores",
			args: []string{"12_345_678_900_987_654_321"},
			exp:  "12345678900987654321",
		},
		{
			name: "one arg: large float with few digits after decimal",
			args: []string{"987656789000000000.12"},
			exp:  "987656789000000000.12",
		},
		{
			name: "one arg: large float with few digits after decimal with commas",
			args: []string{"987,656,789,000,000,000.12"},
			exp:  "987656789000000000.12",
		},
		{
			name: "one arg: large float with few digits after decimal with underscores",
			args: []string{"987_656_789_000_000_000.12"},
			exp:  "987656789000000000.12",
		},
		{
			name: "one arg: float with small with many digits after decimal",
			args: []string{"42.0000000000123000999"},
			exp:  "42.0000000000123000999",
		},
		{
			name: "one arg: negative int",
			args: []string{"-57"},
			exp:  "-57",
		},
		{
			name: "one arg: negative float",
			args: []string{"-15.7"},
			exp:  "-15.7",
		},
		{
			name:     "two args: first is invalid int",
			args:     []string{"eleven", "12"},
			expInErr: []string{"could not parse \"eleven\" as integer"},
		},
		{
			name:     "two args: first is invalid float",
			args:     []string{"eleven.4", "12"},
			expInErr: []string{"could not parse \"eleven.4\" as float"},
		},
		{
			name:     "two args: second is invalid int",
			args:     []string{"11", "twelve"},
			expInErr: []string{"could not parse \"twelve\" as integer"},
		},
		{
			name:     "two args: second is invalid float",
			args:     []string{"3.1", "6.twelve"},
			expInErr: []string{"could not parse \"6.twelve\" as float"},
		},
		{
			name: "two args: int + int",
			args: []string{"8723413", "42823938291913"},
			exp:  "42823947015326",
		},
		{
			name: "two args: int + small float",
			args: []string{"31387", "0.0000000001"},
			exp:  "31387.0000000001",
		},
		{
			name: "two args: int + large float",
			args: []string{"31387", "9000000000000.99"},
			exp:  "9000000031387.99",
		},
		{
			name: "two args: small float + int",
			args: []string{"0.0000000001", "31387"},
			exp:  "31387.0000000001",
		},
		{
			name: "two args: large float + int",
			args: []string{"9000000000000.99", "31387"},
			exp:  "9000000031387.99",
		},
		{
			name: "two args: small float + large float",
			args: []string{"0.000000000001294", "123456000000000.5"},
			exp:  "123456000000000.500000000001294",
		},
		{
			name: "two args: large float + small float",
			args: []string{"123456000000000.5", "0.000000000001293"},
			exp:  "123456000000000.500000000001293",
		},
		{
			name: "two args: large float + large float",
			args: []string{"1515151515151515.15", "999999999999.871515"},
			exp:  "1516151515151515.021515",
		},
		{
			name: "two args: small float + small float",
			args: []string{"303.000000789", "15.12300000000000055"},
			exp:  "318.12300078900000055",
		},
		{
			name: "two args: really large float + really small float",
			args: []string{"432,100,000,000,000,000,000,000.05", "0.0000000000001300000000451"},
			exp:  "432100000000000000000000.0500000000001300000000451",
		},
		{
			name: "two args: really small float + really larg float",
			args: []string{"0.0000000000001300000000451", "432,100,000,000,000,000,000,000.05"},
			exp:  "432100000000000000000000.0500000000001300000000451",
		},
		{
			name: "two args: pos int + neg int: pos result",
			args: []string{"574", "-28"},
			exp:  "546",
		},
		{
			name: "two args: pos int + neg int: neg result",
			args: []string{"5555", "-10001"},
			exp:  "-4446",
		},
		{
			name: "two args: neg int + pos int: pos result",
			args: []string{"-28", "574"},
			exp:  "546",
		},
		{
			name: "two args: neg int + pos int: neg result",
			args: []string{"-10001", "5555"},
			exp:  "-4446",
		},
		{
			name: "two args: neg int + neg int",
			args: []string{"-123456", "-975312468"},
			exp:  "-975435924",
		},
		{
			name: "two args: pos float + neg float: pos result",
			args: []string{"46581.777", "-14.49"},
			exp:  "46567.287",
		},
		{
			name: "two args: pos float + neg float: neg result",
			args: []string{"5,718,222.4", "-11111111111.111"},
			exp:  "-11105392888.711",
		},
		{
			name: "two args: neg float + pos float: pos result",
			args: []string{"-14.49", "46581.777"},
			exp:  "46567.287",
		},
		{
			name: "two args: neg float + pos float: neg result",
			args: []string{"-11111111111.111", "5,718,222.4"},
			exp:  "-11105392888.711",
		},
		{
			name: "two args: neg float + neg float",
			args: []string{"-87.00012300123", "-9000.000001"},
			exp:  "-9087.00012400123",
		},
		{
			name: "two args: pos int + neg float",
			args: []string{"666333", "-32.0001"},
			exp:  "666300.9999",
		},
		{
			name: "two args: neg int + pos float",
			args: []string{"-666333", "32.0001"},
			exp:  "-666300.9999",
		},
		{
			name: "two args: neg int + neg float",
			args: []string{"-8", "-3.15"},
			exp:  "-11.15",
		},
		{
			name: "two args: pos float + neg int",
			args: []string{"32.0001", "-666333"},
			exp:  "-666300.9999",
		},
		{
			name: "two args: neg float + pos int",
			args: []string{"-32.0001", "666333"},
			exp:  "666300.9999",
		},
		{
			name: "two args: neg float + neg int",
			args: []string{"-3.15", "-8"},
			exp:  "-11.15",
		},
		{
			name: "three args: all ints",
			args: []string{"123", "45600", "7890001"},
			exp:  "7935724",
		},
		{
			name: "three args: int + float + int",
			args: []string{"5312", "-1000.55", "43"},
			exp:  "4354.45",
		},
		{
			name: "three args: float + int + float",
			args: []string{"77000000000.05", "-7", "8888.404"},
			exp:  "77000008881.454",
		},
		{
			name: "three args: all floats",
			args: []string{"123.456", "0.00000000000013", "987987987678.991"},
			exp:  "987987987802.44700000000013",
		},
		{
			name:     "three args: first is invalid int",
			args:     []string{"seven", "2", "3"},
			expInErr: []string{"could not parse \"seven\" as integer"},
		},
		{
			name:     "three args: first is invalid float",
			args:     []string{"3.1four1", "2", "3"},
			expInErr: []string{"could not parse \"3.1four1\" as float"},
		},
		{
			name:     "three args: second is invalid int",
			args:     []string{"1", "seven", "3"},
			expInErr: []string{"could not parse \"seven\" as integer"},
		},
		{
			name:     "three args: second is invalid float",
			args:     []string{"1", "3.1four1", "3"},
			expInErr: []string{"could not parse \"3.1four1\" as float"},
		},
		{
			name:     "three args: third is invalid int",
			args:     []string{"1", "2", "seven"},
			expInErr: []string{"could not parse \"seven\" as integer"},
		},
		{
			name:     "three args: third is invalid float",
			args:     []string{"1", "2", "3.1four1"},
			expInErr: []string{"could not parse \"3.1four1\" as float"},
		},
		{
			name: "61 ints",
			args: []string{
				"720937664390779", "1153751913750000000", "449255625000000", "9257437500000000",
				"79852159058958332", "79760186447395832", "267542813541666666", "3370475108458115881",
				"241,187,390,625,000,000", "28,557,455,625,000,000", "503,105,617,106,250,000", "112,500,000",
				"24843758437500000", "100593902343750000", "26334269062500000", "37620384375000000",
				"9375000000", "48906500625000000", "305567614029375000", "1632869491864190",
				"302573", "937500000", "4076086956521", "103125000000",
				"247_980_468_750_000", "297979639524209558", "14671488883568196", "648155382187500000",
				"338709607500000000", "6_853_218_750_000_000", "937500000", "3750000000",
				"75240709687500000", "493421052630", "16_741_598_936_954_135_789", "106970339567812500",
				"2812500000", "9375000000", "478125000000000", "335_313_017_812_500_000",
				"124164986250000", "14211971767512490596", "231272598651562500", "38243580880312500",
				"16765577812500000", "497758540194791665", "4372071283783783", "9375000000",
				"1064191425000000", "1192817887500000", "792873089231250000", "36978218749999999",
				"70434375000", "5625000000", "18973333673622524", "315514722187500000",
				"76181107901956318", "14026875046406250000", "947677951389", "4687500000",
				"703124982187500000",
			},
			exp: "55819669568808150721",
		},
		{
			name: "61 floats",
			args: []string{
				"720,937.664390779", "1,153,751,913.750000000", "449,255.625000000", "9,257,437.500000000",
				"79,852,159.058958332", "79,760,186.447395832", "267,542,813.541666666", "3,370,475,108.458115881",
				"241,187,390.625000000", "28,557,455.625000000", "503,105,617.106250000", "0.112500000",
				"24843758.437500000", "100593902.343750000", "26,334,269.062500000", "37,620,384.375000000",
				"9.375000000", "48,906,500.625000000", "305,567,614.029375000", "1,632,869.491864190",
				"0.000302573", "0.937500000", "4,076.086956521", "103.125000000",
				"247,980.468750000", "297,979,639.524209558", "14,671,488.883568196", "648,155,382.187500000",
				"338_709_607.500000000", "6,853,218.750000000", "0.937500000", "3.750000000",
				"75,240,709.687500000", "493.421052630", "16_741_598_936.954135789", "106,970,339.567812500",
				"2.812500000", "9.375000000", "478,125.000000000", "335313017.812500000",
				"124,164.986250000", "14,211,971,767.512490596", "231,272,598.651562500", "38,243,580.880312500",
				"16,765,577.812500000", "497,758,540.194791665", "4,372,071.283783783", "9.375000000",
				"1,064,191.425000000", "1,192,817.887500000", "792,873,089.231250000", "36,978,218.749999999",
				"70.434375000", "5.625000000", "18_973_333.673622524", "315,514,722.187500000",
				"76,181,107.901956318", "14,026,875,046.406250000", "947.677951389", "4.687500000",
				"703,124,982.187500000",
			},
			exp: "55819669568.808150721",
		},
		{
			name: "one thousand args: all 999,999,999,999,999.000000000005",
			args: slices.Repeat([]string{"999,999,999,999,999.000000000005"}, 1000),
			exp:  "999999999999999000.000000005000",
		},
		{
			name: "1000 positive floats all less than zero",
			args: []string{
				"0.852475611994529885", ".1978922524264933212", ".8345", ".462", "0.706",
				".40", ".710012", "0.54", "0.77021788679683321279", ".13815613436107950400",
				"0.3904519040389379558", "0.8573755604497", "0.56044470499836", ".3317700223309", ".4",
				".56544833586062", ".28911112246473592680", "0.453613", "0.1917974", ".94860777285281890",
				"0.007087559677", ".65600144565666486594", "0.0728087098987001149", ".2524717712940176", ".32199489365455551978",
				"0.85971239872952883239", ".66813993659257", "0.238", ".15", "0.52740502615",
				".0535275475619671", "0.59362145147", ".03343281504330", ".2142835121692931", ".8748329727301",
				"0.643", ".5601", "0.129", "0.361", "0.36823922001",
				"0.408", ".084610617224553947", "0.0896825", "0.0663652926126161208", "0.8130591",
				"0.7378447", "0.917", ".03346689459388801", "0.61480682340826421", ".49382166168253936271",
				".118952149109034986", "0.1340508", "0.784", ".7668", "0.97935",
				".403091", "0.212623470512221", ".901894000515450727", "0.226045109756225335", ".495270617612",
				"0.22350923025288", "0.42276", "0.3794885405", ".758301681025418", "0.0222588",
				"0.814773", "0.06925357", "0.270672478676", "0.124295", ".344805382953742083",
				".7828", ".05", ".4119671078757771", ".19540843596674961258", "0.70",
				".46623801824433", "0.7739088218241106", "0.21588310", "0.4024761", ".0627411730520",
				".2881", ".96614852140", "0.8985979866", ".674786116487", ".688",
				".2696", ".9715", ".37153324373915978628", ".31109770372025605", ".7315332617",
				".084942843547685", "0.944003528532", ".900245625", "0.78918013714790", "0.6",
				".087238806", "0.08563", ".1846987254781534697", "0.700638", "0.4390059898",
				".47976", "0.6", "0.1044539300524068400", "0.4696612382681466", "0.094590672",
				"0.06182", ".06511110520873395838", "0.1218309789082", "0.448717636716102", ".2",
				"0.5095", ".807", ".06075833297", ".08623641", "0.19",
				".764976592", "0.832265517", "0.4099964004671952", ".778128109854", ".89",
				".6149731", ".84437569", ".97975524585", "0.56", "0.41274382855424",
				".702", ".236639", "0.291084915257281", ".5", ".6947931714489219735",
				".66", "0.901540", ".412636", ".3586512", ".5",
				".98371175008859358895", "0.914003629731", ".48943755318", ".813", "0.7480517",
				".3", ".6573799952", "0.764311562", "0.949600589381", ".70902676690992",
				"0.37306", "0.787", ".61099183401", "0.99551752274760", ".4357898904996555",
				".78875", "0.5387043352663784", ".38414898280789117", "0.126946205", ".537740",
				".3673428807", ".17153310", ".7700389455541772", "0.6188649533827739008", ".50473",
				".12269518124", "0.643377308185188", ".432577907950", "0.9684829805", ".704778351680028",
				".4862751082160728", ".489693501505059678", "0.0473891651890335451", ".935988616836907319", "0.97200696796787440",
				".1845201010", "0.29641316728354308", "0.192159072", ".9316723570119", "0.4432682638",
				".63140198002", "0.090640515177880", "0.687796123", "0.86716803", "0.3828",
				"0.02291303", "0.97051930831", ".525419236617", "0.77394048112258360999", ".311175009",
				"0.634814065", ".532167203", ".698313", "0.3943273630430682", "0.62345791204512",
				".994371028613", "0.9251091530799694", "0.32012335130185423", "0.68525181063", ".9",
				".8939894617343906", "0.422972250", "0.517545520308120047", ".448573817687404496", "0.8548638575823316122",
				"0.04568329556", ".587544769831409", "0.930447524525", ".28", "0.3229543442549814411",
				"0.606453834944895", ".59787221260771704", ".91763801469952", ".12128430531885", "0.66785514878659",
				"0.0923251064", "0.6393400397303", "0.18", "0.7", "0.9",
				"0.104446236", "0.85337997", "0.527898286", ".3256", "0.206925238313535",
				"0.1217", ".670922502150483612", ".31531983196", "0.617470245313863", "0.130123813319",
				"0.66545861242083", "0.409610035610", "0.2758260187429", ".069088309299583", "0.509843874663582177",
				"0.6798967", "0.8932802545112465", ".717207391979407", ".35688057108532037152", "0.018137787513580426",
				"0.944448386514932495", "0.0734180752201728", "0.6314857651370984", "0.134141566933613566", "0.2519904430809777425",
				".502878135413", ".8026205066907707", "0.18049782838", "0.17240803", ".62",
				"0.79", "0.335686", ".82871252158727433516", ".81444364226809935", ".3963488492860",
				".11939303778937490713", ".992574620666714", ".5445916724672702", "0.3396999730763949791", "0.922",
				"0.068467175184908827", ".92077632199078", "0.1300632", ".63981919632283139", "0.05023029908",
				".1068", "0.6721053", "0.3078745", "0.919359398881506", "0.94690093858986276943",
				".5674003016156", "0.90534", ".682776169", "0.839", "0.56799031132141",
				"0.25417025067611", "0.9834767393", "0.12541811269576502", ".386727169813298134", "0.33466614901",
				".0283849797316", "0.51", "0.0211450454425387", "0.75706396308964699454", "0.495703871524",
				"0.71241812736651", ".81988237617", ".7017768830371", ".36659482534", ".66074965",
				"0.689146004133616", "0.801020499645", ".243", "0.0032645630133579", "0.5704676822448512",
				"0.89360285491", ".55948126890277399732", "0.3664208560976625733", "0.7152438398134", ".426674366",
				".571426", "0.34133331175618097993", "0.6131514571016", ".703169534", ".130434738194000",
				".109826275981722805", "0.1565308812544152", ".379390083188745133", ".066542", ".9759010478816927495",
				".0841791515", ".899", ".42745807823002", ".08475858188262818693", "0.16",
				".254206284791604882", "0.087284396222", "0.829068984808435", "0.325005791310", "0.4",
				"0.7063495132592026", "0.60900438114862", "0.3", ".419140", ".338202",
				"0.64832537102769650809", "0.728291", ".25", "0.172174", "0.34222186",
				".8399798353543", ".484896", "0.61542386021771757", ".221411", ".326375168149",
				".521272600070", "0.226706965600313131", "0.4309", ".0527165857253505", ".748703926657838672",
				"0.859770990669905212", ".97228106965", ".8", ".404879106", "0.788999533644528511",
				".17", "0.54476541264660529839", ".432887895512", "0.8893302315", ".30091",
				".57", ".73259209742", ".7708209368444985202", "0.22", "0.402996103335368630",
				"0.9001447028052", "0.43613018485301281733", ".55149418", "0.227437642571370", "0.08371636699154739606",
				".19057", ".9443185", ".20842", "0.81785041", "0.5292101289422",
				".13", ".087425421", ".4426", ".0878994431609488", "0.576726933951",
				".1001291732", "0.92969051278773436", "0.3652107998", ".7357988771", ".89461476581",
				".16616780", ".306702905806529", "0.41002434264509", ".9145312", "0.736216708056045153",
				"0.639619", "0.64483843806379485", ".87", ".592250", "0.2101443",
				"0.38111", ".177775254809758712", "0.803", "0.81", "0.78815",
				".09044290643533", "0.5407", ".48715460215162814", "0.25772489406857684", "0.582",
				"0.94440008035065", ".3358859842939689", "0.45696058383278", "0.360935193946", ".5502233250138461",
				".83212836475", "0.16286", "0.83857465306528103846", "0.73808066231934", "0.1535431959233227",
				"0.942546990532892", ".511840247578", "0.62986158691254631702", ".32768841729818513938", ".061967772517389330",
				".3367", "0.029017864309814042", ".4947646940", "0.1453619494588174", ".21773499958937175",
				"0.076", "0.686790292316381362", "0.43431976338", ".8075192109017", "0.500542188663391756",
				"0.5", ".04", "0.71", "0.20126123", ".39377588",
				".55437362259070010", ".448915354754718660", ".972309988713", ".21959843364124011", "0.2077221618835",
				".27245007407", ".93554809209102798", "0.06936389480856", ".99", ".85456340222256508",
				".7544396007", "0.5598760273127", ".5279331717112", ".491534446", ".7416976808599",
				"0.1684934208", "0.4476086492622464", ".022365294688351297", "0.55320744202918", "0.53380012929",
				"0.816", ".1912200455873739761", ".55569", ".1805724372552641657", "0.1050515247148786",
				"0.860961007732", ".12497", ".39493065", ".313571639097668966", ".7697430877760589256",
				".80980938659697", ".1268137560", "0.62403204475121959144", ".9781", "0.420681885065588",
				"0.3", "0.35254139832", "0.0741093497238", "0.736555", "0.31826933062470559",
				"0.8395457396242", ".0126", "0.0", "0.01", "0.33663856891792455",
				"0.1501954", "0.7108", "0.806403440322", ".1", "0.825673713513567372",
				"0.161027874452", ".28947902", "0.4701229117256783", ".20036336722439346", ".2891177304839661",
				"0.544", "0.4576618", ".393422", "0.007", "0.643823",
				"0.16537628291074", "0.7654671050968", "0.055710811174", ".57491500321448301", ".3",
				".2265194", "0.63", ".268322432103", ".492275342175900579", "0.91295",
				".1616795744685403", "0.294566777009", ".32106680449", ".11066101499429057773", "0.8438886",
				"0.3", "0.02812", ".826", ".56948438", ".86641188576912506",
				"0.108618988", ".97452333884816", "0.2210911088922", ".544344", "0.334402321855",
				".0258744820254", "0.396", ".3450", ".48", ".431",
				".1120218794", ".7076", ".44899234650552734", ".16158801470259", ".986353621057882426",
				".654295", "0.649200", "0.54820540930774148779", "0.0377948838", ".070754715197839",
				".853539", "0.0125571416055603", "0.67", "0.40363755522191", ".85412356930184778674",
				".1336", ".9534408", "0.002083403", ".1598", "0.232216725",
				"0.495", "0.4928114263239571", ".6", ".9032117023", ".7",
				".46454823532727474", "0.6375691529325350573", ".59037015621232056", "0.87152727565216161746", "0.15238131",
				"0.39771369", "0.46605518281", ".00382569147851632", ".28", ".01830893",
				".807133", ".4248957005984728336", ".089194", ".5622115319224643332", ".14",
				".52567", ".963806685309642443", ".200357865", ".35331017428308986", "0.93148732594",
				".95650302", ".0372088324912904443", "0.6908", "0.7577661", "0.62226",
				".640328614", "0.7132", "0.97503", ".17495644", ".630154597419",
				"0.67088830427", "0.0730671788620", "0.4709780885", ".156941", ".9341335873",
				".28", ".89065418", ".5078413318622626883", ".181879033940", ".4370414994961",
				"0.731265768564", "0.9202466903365630601", ".5575690168354", ".30965306", "0.946565599027352",
				".582949793562740", "0.46", ".51473985", ".06215005", ".77838848004446570388",
				"0.910214897394", "0.028", "0.9034882921198", ".682393050", "0.80947348",
				".03078267055", "0.3", "0.03060490235965", "0.9370449530", ".8610722238298886609",
				".2444", "0.6632574", ".73467251079943", ".31343213", "0.6672095",
				".74998", "0.237", ".12701", "0.3269022145164379", ".60247501100170",
				".76553533202173", "0.3472893", ".17495631581244681", ".9", "0.812790994",
				"0.041895370990", "0.659350335", ".3446979", ".396319347", ".39",
				".29649516516122503", "0.07193747060885772", "0.65", ".84708495555220611", ".5283630",
				".106316369731288", "0.030461", "0.1129722761741958053", "0.23128233", "0.7899317209997",
				"0.00714171874", "0.931602137", "0.0", "0.072162182967880365", "0.761638163374972",
				"0.2513561", "0.349343280", ".970801090550015505", ".845801308", "0.04358962561878",
				".6471542", "0.81133561426725180", "0.92", "0.06", ".93355813",
				"0.0435681179437", ".5108916", ".912493284650747", ".64064460", ".7242359043895001",
				"0.531982450", ".992366", "0.90759365821523505814", ".2670", ".3941761126",
				".03551586229", ".686170122", ".7050", ".7195525264694", "0.0123996",
				"0.12859", ".78889989", "0.07973629280913", "0.6705", ".0262",
				"0.6987993986191176", "0.226266772719", "0.84388986491400", ".81741898843947", ".869106228069668",
				".16", "0.472538977349", ".43468798503957295", ".5898130631", ".938234144066042808",
				"0.91201072175340194131", "0.403704278888", "0.21921493513631675274", ".59910579122", "0.19039054",
				".102316742426520", "0.0785934332993", "0.722327839296", "0.8", ".72602793101236519",
				".9967915638", "0.2766650249694", "0.90360841", ".982", ".426868087",
				".639048583", "0.3298359", ".1", ".2694731602019", "0.519626870295944828",
				"0.19969285", "0.72698", ".9", "0.75669101599", ".431448499258",
				".2228306764953", "0.30066702", ".40", ".3433", ".7349109369866911793",
				"0.98206", "0.2840", "0.34102433", "0.75313953019", ".041",
				".51656193990", ".1834", ".8", ".211", ".7678198918",
				".9", "0.510078903464914854", "0.4", ".11876", "0.912",
				".4310912632", "0.6972961898", ".30681482928608", ".1362012331827", "0.5776469268679911",
				".2872293644", ".82538342959358715125", ".17385769598759725", ".1614100406197564623", ".0730247884269613",
				".85250", ".64789523352342482212", "0.26665572723079", ".37", ".662601958727",
				".333904", "0.31796513", "0.70234259154030", ".5170", ".23444270406750063",
				"0.3383", ".468570", "0.754", ".579", ".428453",
				".670100", ".8666143", ".059379276576080", "0.323899483363", "0.85397463",
				".5263569731780446", ".69666757746", "0.95365851", ".18226671084582883285", "0.65028892143250",
				".84", ".888271962602628736", ".2", ".949325980", ".2419644666198605439",
				".465977575640", "0.19059330331400", "0.70586", "0.26544637", "0.73232773900001821",
				"0.219462", "0.01251121628514944229", ".688324194", "0.6", "0.999184484619840",
				".42", "0.2863146492964", ".65115338", "0.5301872464322975", "0.763312508844381743",
				"0.58520545", "0.699997095630", ".1500923", ".3463628770944333937", "0.082391782337",
				"0.03945", ".071004508965", ".6816971940207", ".65121415", "0.8597",
				".33357006024184415", "0.121303", "0.9418", "0.39520633262407069", ".8387",
				".04", ".4", "0.6378006280363551", ".646040975296", ".0880336627157816389",
				".56", "0.46653831864", ".90468276925254553836", ".74917421811351684196", "0.7831",
				"0.5535571151", "0.52529", ".162995748", "0.286", "0.436426304366",
				".0091611473", ".68343124000998867581", "0.63100", ".51427429", ".564748564123",
				".4594668482150585338", "0.5243938017681924662", "0.84", "0.5883128499246564", ".39440396",
				"0.3766691", "0.58", ".8975393199218757", "0.606312079142", "0.87827508394085593329",
				".19193773620659248421", "0.38729089707405713", ".23351", ".803699405878298", "0.3259472",
				".309719906245", ".60153", "0.0", "0.93196058045056", ".725",
				"0.66968", ".855062041873", ".4172521730", "0.804373420043", ".7138308679167",
				"0.0456058903167381", ".189396091", ".11510", ".1", "0.95258789",
				".36432744", ".1535", ".20796994590", ".316000410", "0.1227899201678",
				"0.7841099", ".345079287841", "0.036787379", "0.23", "0.71431159",
				"0.636674615", ".136022897036404", "0.484217868", "0.6907186928167727", "0.665517",
				"0.31911973419088", "0.472825504", ".6293097", ".3749", "0.6392652678059862452",
				"0.86032710724", "0.9889530937485", ".156778941", "0.22139598029", ".99649246790380136",
				"0.8", "0.2747008", "0.3054610068834441", ".0246779720267", ".693977448171951598",
				"0.6", "0.38", "0.7450401", "0.2361", "0.5938296742765",
				"0.98", ".163780805", ".349526405765", ".81626172050092787762", ".1",
				"0.19415566135209766", "0.9615081291", "0.82678050373061558", "0.848", "0.846460321467777",
				".724376391", ".61674282218489034", ".8", ".764553054", "0.44083676",
				".450", ".0694823231", "0.953639", ".1962809634042593787", ".668666699375453",
				"0.425", "0.3550674331106314", "0.448651063141374303", ".308443087273077014", "0.4770057029487586",
				"0.5440669565311309937", "0.89269251517104987158", ".9628831353", ".72", "0.093",
				"0.66", "0.9655237623533176", "0.222701414680905", ".10542244070330", ".19580000055823729",
				".82744356835250149", "0.696162", "0.031853802", "0.5672731258566", "0.52416734162029",
				".0001663434689851064", ".07160133", "0.6848", ".16256", "0.6730400",
				".95580", "0.75491318642", ".9", "0.20596595055", "0.35990377928",
				".63572109484402881", ".749920091099733", "0.75510", "0.4503639109661", "0.64097651740269650677",
				".715", "0.883993", "0.18", "0.935952", ".3818325410895",
				"0.208063542815", "0.544285", ".245844", "0.916959957612367581", ".8864",
				"0.6732629436862812", ".5855", "0.3870", ".284033", ".178271221699249",
				".51570", ".053401", ".99872", ".5802160", "0.7554",
				"0.33413919", "0.0050274", "0.053622878", ".44", ".822893787921641748",
				"0.020293158070932", ".226", "0.4365831", ".1242", ".1904857472",
				"0.654096661460939515", "0.79856", "0.6005411314351570083", "0.97781057991324301", ".7231771929275",
				"0.213972822", ".03846257", ".1941335570536", ".74878154590", "0.94119285",
				"0.68579566421362", ".81934867", ".984", "0.65", "0.74344953",
				".89017", "0.6007323310814", "0.6763466015260407", "0.8120457", "0.199015615",
				"0.8947566021", "0.4227099976936", "0.295336874885733930", ".2484", ".04840143178831336396",
				"0.7934035182953179", "0.6096726289541764313", ".679", "0.124192223165442754", "0.2754177966288",
				".9174900529487428158", "0.3009236987717769", "0.178207433775", ".30585729987831348", "0.34811601769994822619",
				"0.4204756", "0.61146807450142", "0.173169923", ".5079337842956", "0.09508728959300529992",
				".02769", "0.525684576328584083", ".149300836245718524", "0.7447195800699474", "0.82488793974563785080",
			},
			// Here's how I got the expected answer to this:
			// I copied all of them then in a terminal, created a file with the numbers all filled out to 20 digits.
			// $ pbpaste | tr ',' '\n' | grep -v '^[[:space:]]*$' | \
			//     sed -E 's/[[:space:]"]+//g; s/^0//; s/$/00000000000000000000/; s/^(.{21}).*$/\1/;' > normalized-nums.txt
			// Then I added each digit individually (as an int):
			// $ tot=""; c=0; for d in $( seq 21 2 ); do \
			//     s="$( while read w; do printf '%d\n' "${w:(($d-1)):1}"; done < normalized-nums.txt | add )"; \
			//     s="$(( s + c ))"; o="$(( s % 10 ))"; c="$(( s / 10 ))"; tot="${o}${tot}"; \
			//     printf '%d = %d => %d %d   %q\n' "$d" "$s" "$c" "$o" "$tot"; \
			//   done
			// The last line is "2 = 4907 => 490 7   79009620742101359409" which says that the total is
			// 79009620742101359409 and the carry is 490. So the expected sum is 490.79009620742101359409.
			exp: "490.79009620742101359409",
		},
		{
			name: "200 floats between -1 and 1",
			args: []string{
				"-0.64611980", ".6107023618107903", "-0.09857224", ".0131629150", "-.8446951027310",
				"-0.41679284370984874", "-.7983375262067190", "-0.1", ".074308171", "-.543959179549083",
				"-.5540910634", ".1817688540537532", ".29423464417", ".341621", "0.3086748677873",
				"0.67", "-0.0731941035190695028", "-.374", "-0.56041886819708", "0.622258034384830070",
				"0.6450132", "-0.23228281130749", "-0.1073137071", ".938294596976177", "-0.2464971",
				"0.6772227229", ".03136662499859507398", "-.121306562884", "-0.41160", "-.717168197022719038",
				"-0.3907319", "-.94961875244268978901", "-0.402", "0.701229488444", "-.8095106258012769721",
				"-0.597326", ".4357463973066273334", "-.7910061", "-0.3648705726415163", ".1552356401336289",
				"0.599666991962", "-0.33", "0.14500", "0.878", "0.2521043251105216923",
				"-.049853805720898", "0.31095292941", ".618225639169", "-.922183", ".6367070",
				"0.27960213", "-.813458348035800", ".1950544373192704", "-0.81292577481897", "0.6967759",
				"0.27055073", "0.8094", "-0.97545124672273825", "0.468", "0.2400042177561253426",
				"-0.277", "0.32109375", "0.654368492522341546", ".552555455258", "0.20480",
				".75976906", "-.77402816461416", "0.8865170", ".080772", "0.50347519275501",
				".76502835", "-.12583622215544385616", ".479724563758", "0.07724775101050806531", "-0.700315306668285",
				"0.254046", "-.744", "-0.197870978837915", "0.52229592144482354", ".909678909835",
				"-0.02", ".04", "0.7044787191476740", ".5617637009527119490", "0.4245016082",
				"-.0172521566381", ".74847", "-0.899792866099153143", "-.0378766535530530", "0.646176013166557120",
				".7559501509", ".69064997", "0.03981", "-0.57491112301", "-0.11477186838784574434",
				"0.7218", "-0.91391", "-.22490948", "0.54626670146728400849", "-.71131",
				"0.9805960182569805951", "-.4064877941887949966", "0.897531047", "-.7130067250", "0.9042602778752809",
				".091198", "0.810246", "0.5897216667935230", ".7568416755208", "-.43953356515104757121",
				"-0.25243871637877", "0.40495236943", "-.22926408098583", "0.236682942249180", ".1647770735",
				"0.582325783660", "-0.99", "0.18034412983647492775", "-0.621296633", "-0.87886",
				"-.37520780", ".533444934833981", "0.52639", ".7546841", ".5665297131104143",
				"0.492393470845967", "0.0878276475880", "-.806", ".6357", "-.348274891",
				".78928", "0.7472", "0.97", ".457", ".568",
				"0.4511571952617029152", "-0.356950", "0.23733397489547215", "-0.13401", ".561",
				".844026195064423134", "0.9718", "-.2478704294219", "0.5618010125835", "-0.72281",
				"-0.92961341219", ".15392093539725631", "0.1741466", "0.26", "-.0158249848344",
				".41753746", "-.93000254449264", "-.549006920791", "-.15641134", "-0.0910158",
				".4163188547844", "0.0", "0.25613795408673641", ".8761409685486844159", ".525767635885400950",
				"0.4418408968675", "-0.99088174495", "-0.000987101", "-.96040042138672", "0.26519166",
				"-.5697255990", "-0.749553654119936855", ".53222848030", "-.9606940240", "0.8335205",
				"-.335980875513", ".3599", "0.7943", "-0.690", "-0.8048510414660964964",
				".464221771881734", "-0.567065645959641", "0.9", "0.871364674", "-0.155967413931",
				"0.536676477", ".818765782757", "-0.0", ".04863", "0.0544948711735",
				"-0.76409422885", "-0.56479026636639", ".240180119448999075", "-.9700", "-0.3913526",
				".06246179", "0.67834", "-.7", ".147640234164", ".4298457",
				"-.526293464496814", "-0.9511032", "0.49023357417786401121", ".5131989752128932636", ".11266177698922",
			},
			// Here's how I got the expected answer to this:
			// I copied all of them then in a terminal, created two files (+ and -) with the numbers all filled out to 20 digits.
			// $ pbpaste | tr ',' '\n' | sed -E 's/[[:space:]"]+//g;' | grep -v '^$' | grep '^-' | \
			//     sed -E 's/^0//; s/$/00000000000000000000/; s/^(.{21}).*$/\1/;' > normalized-neg.txt
			// $ pbpaste | tr ',' '\n' | sed -E 's/[[:space:]"]+//g;' | grep -v '^$' | grep '^-' | \
			//     sed -E 's/^-0/-/; s/$/00000000000000000000/; s/^(.{22}).*$/\1/;' > normalized-neg.txt
			// Then, for each, I added each digit individually (as an int). See previous case for the command.
			// The last line for positives is "2 = 559 => 55 9   98283605109141789884".
			// The last line for negatives is "3 = 432 => 43 2   23669697024883525462".
			// So we have 55.98283605109141789884 - 43.23669697024883525462 = 12.74613908084258264422
			exp:      "12.74613908084258264422",
			expInErr: nil,
		},
		{
			name: "a bunch of small negative numbers.",
			args: []string{
				"-0.64611980", "-0.09857224", "-.8446951027310", "-0.41679284370984874", "-.7983375262067190",
				"-0.1", "-.543959179549083", "-.5540910634", "-0.0731941035190695028", "-.374",
				"-0.56041886819708", "-0.23228281130749", "-0.1073137071", "-0.2464971", "-.121306562884",
				"-0.41160", "-.717168197022719038", "-0.3907319", "-.94961875244268978901", "-0.402",
				"-.8095106258012769721", "-0.597326", "-.7910061", "-0.3648705726415163", "-0.33",
				"-.049853805720898", "-.922183", "-.813458348035800", "-0.81292577481897", "-0.97545124672273825",
				"-0.277", "-.77402816461416", "-.12583622215544385616", "-0.700315306668285", "-.744",
				"-0.197870978837915", "-0.02", "-.0172521566381", "-0.899792866099153143", "-.0378766535530530",
				"-0.57491112301", "-0.11477186838784574434", "-0.91391", "-.22490948", "-.71131",
				"-.4064877941887949966", "-.7130067250", "-.43953356515104757121", "-0.25243871637877", "-.22926408098583",
				"-0.99", "-0.621296633", "-0.87886", "-.37520780", "-.806",
				"-.348274891", "-0.356950", "-0.13401", "-.2478704294219", "-0.72281",
				"-0.92961341219", "-.0158249848344", "-.93000254449264", "-.549006920791", "-.15641134",
				"-0.0910158", "-0.99088174495", "-0.000987101", "-.96040042138672", "-.5697255990",
				"-0.749553654119936855", "-.9606940240", "-.335980875513", "-0.690", "-0.8048510414660964964",
				"-0.567065645959641", "-0.155967413931", "-0.0", "-0.76409422885", "-0.56479026636639",
				"-.9700", "-0.3913526", "-.7", "-.526293464496814", "-0.9511032",
			},
			exp: "-43.23669697024883525462",
		},
		{
			name: "a bunch of small positive numbers",
			args: []string{
				".6107023618107903", ".0131629150", ".074308171", ".1817688540537532", ".29423464417",
				".341621", "0.3086748677873", "0.67", "0.622258034384830070", "0.6450132",
				".938294596976177", "0.6772227229", ".03136662499859507398", "0.701229488444", ".4357463973066273334",
				".1552356401336289", "0.599666991962", "0.14500", "0.878", "0.2521043251105216923",
				"0.31095292941", ".618225639169", ".6367070", "0.27960213", ".1950544373192704",
				"0.6967759", "0.27055073", "0.8094", "0.468", "0.2400042177561253426",
				"0.32109375", "0.654368492522341546", ".552555455258", "0.20480", ".75976906",
				"0.8865170", ".080772", "0.50347519275501", ".76502835", ".479724563758",
				"0.07724775101050806531", "0.254046", "0.52229592144482354", ".909678909835", ".04",
				"0.7044787191476740", ".5617637009527119490", "0.4245016082", ".74847", "0.646176013166557120",
				".7559501509", ".69064997", "0.03981", "0.7218", "0.54626670146728400849",
				"0.9805960182569805951", "0.897531047", "0.9042602778752809", ".091198", "0.810246",
				"0.5897216667935230", ".7568416755208", "0.40495236943", "0.236682942249180", ".1647770735",
				"0.582325783660", "0.18034412983647492775", ".533444934833981", "0.52639", ".7546841",
				".5665297131104143", "0.492393470845967", "0.0878276475880", ".6357", ".78928",
				"0.7472", "0.97", ".457", ".568", "0.4511571952617029152",
				"0.23733397489547215", ".561", ".844026195064423134", "0.9718", "0.5618010125835",
				".15392093539725631", "0.1741466", "0.26", ".41753746", ".4163188547844",
				"0.0", "0.25613795408673641", ".8761409685486844159", ".525767635885400950", "0.4418408968675",
				"0.26519166", ".53222848030", "0.8335205", ".3599", "0.7943",
				".464221771881734", "0.9", "0.871364674", "0.536676477", ".818765782757",
				".04863", "0.0544948711735", ".240180119448999075", ".06246179", "0.67834",
				".147640234164", ".4298457", "0.49023357417786401121", ".5131989752128932636", ".11266177698922",
			},
			exp:      "55.98283605109141789884",
			expInErr: nil,
		},
		{
			name: "six small floats that sum to 4.999",
			args: []string{
				"0.89", "0.5", "0.804", "0.911005", "0.993990", "0.900005",
			},
			exp:      "4.999000",
			expInErr: nil,
		},
		{
			name: "six small floats that sum to 5",
			args: []string{
				"0.89", "0.5", "0.805", "0.911005", "0.993990", "0.900005",
			},
			exp:      "5.000000",
			expInErr: nil,
		},
		{
			name: "six small floats that sum to 5.001",
			args: []string{
				"0.89", "0.5", "0.805", "0.911005", "0.993990", "0.901005",
			},
			exp:      "5.001000",
			expInErr: nil,
		},
		{
			name: "30 mixed numbers",
			args: []string{
				"67159084558770.2318", "1997812568451", "-908185273772020889", "-0.3451415014580", "460455737993476.57",
				"4164444", "-0.169404273651", "-.29189684", "-40391707941694704", "-14971374442985881937",
				"-.5629477", "3227006874.2117", "-854105602.16129", "-83603.086", ".1082688261379289",
				"-4575237", ".0", "-83886330743", "-116070.3245", "-0.198124388968",
				"0.80813992124068", "-3645801743247047075.446", "-.35768890139959323105", "-.963484574836", "8",
				"-486917167", "-36674207550.43", "-.45", ".2466157907657", "0.7576774085",
			},
			exp: "-19565223673986688555.85227623366828433105", // Yes I verified this manually.
		},
		{
			name: "floats only have zeros after decimal",
			args: []string{"5", "12.0", "74.00000", "-28.00"},
			exp:  "63.00000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origVerbose := Verbose
			defer func() {
				Verbose = origVerbose
			}()
			Verbose = tc.verbose

			var act string
			var err error
			testFunc := func() {
				act, err = sumFunc(tc.args)
			}
			require.NotPanics(t, testFunc, "%s(%q)", name, tc.args)
			if len(tc.expInErr) == 0 {
				assert.NoError(t, err, "%s(%q) error", name, tc.args)
			} else {
				assert.Error(t, err, "%s(%q) error, expecting %q", name, tc.args, tc.expInErr)
				if checkErrContents {
					for _, exp := range tc.expInErr {
						assert.ErrorContains(t, err, exp, "looking for %q in %s(%q) error", exp, name, tc.args)
					}
				}
			}
			assert.Equal(t, tc.exp, act, "%s(%q) result", name, tc.args)
		})
	}
}

func TestSum(t *testing.T) {
	RunSumFuncTests(t, "Sum", Sum, false)
}

func TestSum1(t *testing.T) {
	RunSumFuncTests(t, "Sum1", Sum1, true)
}

func TestSum2(t *testing.T) {
	RunSumFuncTests(t, "Sum2", Sum2, false)
}

func TestSum3(t *testing.T) {
	RunSumFuncTests(t, "Sum3", Sum3, true)
}

func TestSum4(t *testing.T) {
	RunSumFuncTests(t, "Sum4", Sum4, true)
}

func TestSum5(t *testing.T) {
	RunSumFuncTests(t, "Sum5", Sum5, true)
}

func TestMakeNumberPretty(t *testing.T) {
	tests := []struct {
		name string
		val  string
		exp  string
	}{
		{
			name: "empty string",
			val:  "",
			exp:  "",
		},
		{
			name: "already has a comma",
			val:  "123,4567890",
			exp:  "123,4567890",
		},
		{
			name: "one digit",
			val:  "1",
			exp:  "1",
		},
		{
			name: "two digits",
			val:  "12",
			exp:  "12",
		},
		{
			name: "three digits",
			val:  "123",
			exp:  "123",
		},
		{
			name: "four digits",
			val:  "4321",
			exp:  "4,321",
		},
		{
			name: "five digits",
			val:  "12345",
			exp:  "12,345",
		},
		{
			name: "six digits",
			val:  "444333",
			exp:  "444,333",
		},
		{
			name: "seven digits",
			val:  "4666999",
			exp:  "4,666,999",
		},
		{
			name: "eight digits",
			val:  "12543876",
			exp:  "12,543,876",
		},
		{
			name: "nine digits",
			val:  "789456123",
			exp:  "789,456,123",
		},
		{
			name: "20 digits",
			val:  "12345678901234567890",
			exp:  "12,345,678,901,234,567,890",
		},
		{
			name: "61 digits",
			val:  "1234567890098765432112345678900987654321123456789009876543210",
			exp:  "1,234,567,890,098,765,432,112,345,678,900,987,654,321,123,456,789,009,876,543,210",
		},
		{
			name: "one digit: negative",
			val:  "-1",
			exp:  "-1",
		},
		{
			name: "two digits: negative",
			val:  "-12",
			exp:  "-12",
		},
		{
			name: "three digits: negative",
			val:  "-123",
			exp:  "-123",
		},
		{
			name: "four digits: negative",
			val:  "-4321",
			exp:  "-4,321",
		},
		{
			name: "five digits: negative",
			val:  "-12345",
			exp:  "-12,345",
		},
		{
			name: "six digits: negative",
			val:  "-444333",
			exp:  "-444,333",
		},
		{
			name: "seven digits: negative",
			val:  "-4666999",
			exp:  "-4,666,999",
		},
		{
			name: "eight digits: negative",
			val:  "-12543876",
			exp:  "-12,543,876",
		},
		{
			name: "nine digits: negative",
			val:  "-789456123",
			exp:  "-789,456,123",
		},
		{
			name: "20 digits: negative",
			val:  "-12345678901234567890",
			exp:  "-12,345,678,901,234,567,890",
		},
		{
			name: "61 digits: negative",
			val:  "-1234567890098765432112345678900987654321123456789009876543210",
			exp:  "-1,234,567,890,098,765,432,112,345,678,900,987,654,321,123,456,789,009,876,543,210",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MakeNumberPretty(tc.val)
			}
			require.NotPanics(t, testFunc, "MakeNumberPretty(%q)", tc.val)
			assert.Equal(t, tc.exp, act, "MakeNumberPretty(%q)", tc.val)
		})
	}
}
