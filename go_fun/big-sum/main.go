package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
)

// PrintUsage outputs a multi-line string with info on how to run this program.
func PrintUsage(stdout io.Writer) {
	fmt.Fprintf(stdout, `big-sum: Add a bunch of numbers together with nearly infinite precision.

Usage: big-sum <number 1> [<number 2> ...] [--pipe|-] [--pretty|-p] [--verbose|-v]
  or : <stuff> | big-sum

The --pipe or - flag is implied if there are no arguments provided.
The --pretty or -p flag will add commas to the result.

Warning: In rare circumstances, floating point numbers may result in unwanted rounding.
`)
}

// Sum will parse each arg as a number and return a sum of all those numbers as a converted to a string.
// By default it uses the Sum4 function which is clearly better than the others in most categories.
// Sum2 and Sum3, in some of the benchmarks, performs better than Sum4, though, but only just barely.
var Sum = Sum3

// Sum1 will parse each arg as a number and return a sum of all those numbers as a converted to a string.
// This version parses each arg as a single number while trying to maintain enough precision for everything.
func Sum1(args []string) (string, error) {
	var totalInt *big.Int
	var totalFloat *big.Float
	wholeDigits := 0
	fractionalDigits := 0

	prec := calculatePrec(args)

	for _, arg := range args {
		orig := arg
		// Remove all commas so that people can provide numbers with the commas in them.
		arg = strings.ReplaceAll(arg, ",", "")

		if equalFoldOneOf(arg, "", "0", "0.0", ".0", "0.") {
			verbosef("Ignoring empty or zero arg: %q.", orig)
			continue
		}

		if !strings.Contains(arg, ".") {
			// Number doesn't have a ".", Parse it as an integer.
			// Note that by using 0 as the base, the arg can also handle underscore separators, e.g. 123_456!
			val, ok := new(big.Int).SetString(arg, 0)
			if !ok {
				return "", fmt.Errorf("could not parse %q as integer", orig)
			}
			if totalInt == nil {
				verbosef("Ints:   %40s  %q", val, arg)
				totalInt = val
			} else {
				verbosef("Ints: + %40s  %q", val, arg)
				totalInt.Add(totalInt, val)
				verbosef("Ints: = %40s", totalInt)
			}
		} else {
			// Number has a ".", parse it as a float.
			argWholeLen, argFractLen := countDigits(arg)
			if argWholeLen > wholeDigits {
				wholeDigits = argWholeLen
			}
			if argFractLen > fractionalDigits {
				fractionalDigits = argFractLen
			}

			val, _, err := big.ParseFloat(arg, 0, prec, big.ToNearestEven)
			if err != nil {
				return "", fmt.Errorf("could not parse %q as float: %w", orig, err)
			}
			if totalFloat == nil {
				verbosef("Floats:   %40s  from %q (prec=%d, acc=%s) (%d,%d)",
					val.Text('f', argFractLen)+strings.Repeat(" ", fractionalDigits-argFractLen), arg,
					val.Prec(), val.Acc(), argWholeLen, argFractLen)
				totalFloat = val
			} else {
				verbosef("Floats: + %40s  from %q (prec=%d, acc=%s) (%d,%d)",
					val.Text('f', argFractLen)+strings.Repeat(" ", fractionalDigits-argFractLen), arg,
					val.Prec(), val.Acc(), argWholeLen, argFractLen)
				totalFloat = new(big.Float).SetPrec(prec).Add(totalFloat, val)
				verbosef("Floats: = %40s (prec=%d, acc=%s) (%d,%d)",
					totalFloat.Text('f', fractionalDigits), totalFloat.Prec(),
					totalFloat.Acc(), wholeDigits, fractionalDigits)
			}
		}
	}

	if totalFloat != nil {
		if totalInt != nil {
			sumInts := totalInt.String()
			if len(sumInts) > wholeDigits {
				wholeDigits = len(sumInts)
			}
			verbosef("Sum Ints:     %40s", sumInts+strings.Repeat(" ", fractionalDigits+1))
			verbosef("Sum Floats: + %40s", totalFloat.Text('f', fractionalDigits))
			totalFloat = new(big.Float).SetPrec(precForLen(wholeDigits+fractionalDigits+1)).Add(totalFloat, new(big.Float).SetInt(totalInt))
			verbosef("Grand Sum:  = %40s", totalFloat.Text('f', fractionalDigits))
		}
		return totalFloat.Text('f', fractionalDigits), nil
	}
	if totalInt != nil {
		return totalInt.String(), nil
	}
	return "0", nil
}

// calculatePrec will calculate the precision needed to add all the floats together accurately.
func calculatePrec(args []string) uint {
	var wMax, fMax int
	for _, arg := range args {
		w, f := countDigits(arg)
		if w > wMax {
			wMax = w
		}
		if f > fMax {
			fMax = f
		}
	}
	// Add one to the length for extra growth room.
	// Add one to the length for each 3 args too for the same reason.
	return precForLen(wMax + fMax + 1 + len(args)/3)
}

// countDigits returns the number of whole and fractional digits in the provided number string.
func countDigits(arg string) (whole int, fractional int) {
	parts := strings.Split(arg, ".")
	if len(parts) > 0 {
		whole = len(parts[0])
		if strings.HasPrefix(parts[0], "-") {
			whole--
		}
	}
	if len(parts) > 1 {
		fractional = len(parts[1])
	}
	return
}

// Sum2 will parse each arg as a number and return a sum of all those numbers as a converted to a string.
// This version parses the whole and partial parts of each arg separately then combines them at the end.
// It first splits each arg then does all whole parts, then all fractional parts.
func Sum2(args []string) (string, error) {
	var wholes, fractionals []string
	fractionalDigits := 0
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}

		// Remove all commas so that people can provide numbers with the commas in them.
		arg = strings.ReplaceAll(arg, ",", "")

		if !strings.Contains(arg, ".") {
			wholes = append(wholes, arg)
			continue
		}

		parts := strings.Split(arg, ".")
		if len(parts[0]) > 0 && parts[0] != "-" {
			wholes = append(wholes, parts[0])
		}
		if len(parts[1]) > 0 {
			fp := "." + parts[1]
			if strings.HasPrefix(arg, "-") {
				fp = "-" + fp
			}
			fractionals = append(fractionals, fp)
			if len(parts[1]) > fractionalDigits {
				fractionalDigits = len(parts[1])
			}
		}
	}

	var totalInt *big.Int
	for _, whole := range wholes {
		val, ok := new(big.Int).SetString(whole, 0)
		if !ok {
			return "", fmt.Errorf("could not parse %q as integer", whole)
		}
		if totalInt == nil {
			verbosef("Ints:   %40s  %q", val, whole)
			totalInt = val
		} else {
			verbosef("Ints: + %40s  %q", val, whole)
			totalInt.Add(totalInt, val)
			verbosef("Ints: = %40s", totalInt)
		}
	}

	if totalInt == nil {
		totalInt = new(big.Int).SetInt64(0)
	}

	if len(fractionals) == 0 {
		return totalInt.String(), nil
	}

	var totalFloat *big.Float
	prec := precForLen(fractionalDigits)
	for _, fractional := range fractionals {
		val, _, err := big.ParseFloat(fractional, 0, prec, big.ToNearestEven)
		if err != nil {
			return "", fmt.Errorf("could not parse %q as float: %w", fractional, err)
		}
		fractLen := len(fractional) - 1
		if strings.HasPrefix(fractional, "-") {
			fractLen--
		}
		if totalFloat == nil {
			verbosef("Floats:   %40s  from %q (prec=%d, acc=%s) (%d,%d)",
				val.Text('f', fractLen)+strings.Repeat(" ", fractionalDigits-fractLen), fractional,
				val.Prec(), val.Acc(), 0, fractLen)
			totalFloat = val
		} else {
			verbosef("Floats: + %40s  from %q (prec=%d, acc=%s) (%d,%d)",
				val.Text('f', fractLen)+strings.Repeat(" ", fractionalDigits-fractLen), fractional,
				val.Prec(), val.Acc(), 0, fractLen)
			totalFloat = totalFloat.Add(totalFloat, val)
			verbosef("Floats: = %40s (prec=%d, acc=%s) (%d,%d)",
				totalFloat.Text('f', fractionalDigits), totalFloat.Prec(),
				totalFloat.Acc(), 0, fractionalDigits)
		}
	}

	wholeSign := totalInt.Sign()
	if wholeSign == 0 {
		// If there's no whole part, we can just return the fractional total.
		return totalFloat.Text('f', fractionalDigits), nil
	}

	// If the fractional parts total is at least one, move the whole part of it into
	// the whole total and make the float part just the fractional portion.
	if new(big.Float).Abs(totalFloat).Cmp(new(big.Float).SetInt64(1)) > 0 {
		totalFloatWhole, _ := totalFloat.Int(nil)
		totalInt = totalInt.Add(totalInt, totalFloatWhole)
		totalFloat = totalFloat.Sub(totalFloat, new(big.Float).SetPrec(prec).SetInt(totalFloatWhole))
	}

	floatSign := totalFloat.Sign()
	if floatSign != 0 && wholeSign != floatSign {
		// If there's a fractional part and its sign is different from the whole part, we need to adjust things.
		// The fixer is 1 with the same sign as the fractional part.
		// We take it out of the fractional part and put it into the whole part.
		// This makes the signs the same again so we can concatenate them for the total.
		fixer := new(big.Int).SetInt64(1)
		if floatSign < 0 {
			fixer = fixer.Neg(fixer)
		}
		totalFloat = totalFloat.Sub(totalFloat, new(big.Float).SetPrec(prec).SetInt(fixer))
		totalInt = totalInt.Add(totalInt, fixer)
	}

	wholePart := totalInt.String()
	fractionalPart := totalFloat.Text('f', fractionalDigits)
	fractionalPartParts := strings.Split(fractionalPart, ".")
	return wholePart + "." + fractionalPartParts[1], nil
}

// Sum3 will parse each arg as a number and return a sum of all those numbers as a converted to a string.
// This version parses both parts of each arg before going to the next arg.
func Sum3(args []string) (string, error) {
	var totalInt *big.Int
	var totalFloat *big.Float

	argInfos, fractionalDigits := prepArgs(args)
	prec := precForLen(fractionalDigits)

	for _, arg := range argInfos {
		whole, fractional, err := arg.Parse(prec)
		if err != nil {
			return "", err
		}
		logArg(arg, whole, fractional, fractionalDigits)

		switch {
		case totalInt == nil:
			totalInt = whole
		case whole != nil:
			totalInt.Add(totalInt, whole)
		}

		switch {
		case totalFloat == nil:
			totalFloat = fractional
		case fractional != nil:
			totalFloat.Add(totalFloat, fractional)
		}

		totalInt, totalFloat = normalizeAmounts3(totalInt, totalFloat)
		logSum(totalInt, totalFloat, fractionalDigits)
	}

	return getCombinedString(totalInt, totalFloat, fractionalDigits), nil
}

// normalizeAmounts3 takes the integer portion out of the float and adds it to the whole.
// The sum of the returned numbers will equal the sum of the provided numbers.
// If fractional == nil, nothing changes. If whole is nil, |fractional| must be at least five
// before it gets split. If there's already a whole amount, though, all integer portions of
// fractional are moved to the whole amount.
func normalizeAmounts3(whole *big.Int, fractional *big.Float) (*big.Int, *big.Float) {
	if fractional == nil {
		return whole, fractional
	}

	maxFract := getMaxFractCutoff(whole)
	if new(big.Float).Abs(fractional).Cmp(maxFract) < 0 {
		return whole, fractional
	}

	fractInt, _ := fractional.Int(nil)
	if whole == nil {
		whole = fractInt
	} else {
		whole.Add(whole, fractInt)
	}
	fractional.Sub(fractional, new(big.Float).SetPrec(fractional.Prec()).SetInt(fractInt))
	return whole, fractional
}

// getMaxFractCutoff returns the maximum amount that we'll let a float be.
// If whole isn't nil, returns 1, otherwise 5.
func getMaxFractCutoff(whole *big.Int) *big.Float {
	if whole != nil {
		if oneFloat == nil {
			oneFloat = new(big.Float).SetInt64(1)
		}
		return oneFloat
	}

	if fiveFloat == nil {
		fiveFloat = new(big.Float).SetInt64(5)
	}
	return fiveFloat
}

// argInfo holds an argument and how it gets split up.
type argInfo struct {
	Orig string

	WholeStr    string
	WholeDigits int

	FractionalStr    string
	FractionalDigits int
}

// prepArgs will create argInfo for each arg, and return the max number of fractional digits seen.
func prepArgs(args []string) ([]*argInfo, int) {
	var rv []*argInfo
	maxFractionalDigits := 0

	for _, arg := range args {
		info := &argInfo{Orig: arg}
		rv = append(rv, info)

		if len(arg) == 0 {
			continue
		}

		parts := strings.SplitN(arg, ".", 2)
		if len(parts) == 0 {
			continue
		}

		info.WholeStr = strings.ReplaceAll(parts[0], ",", "")
		info.WholeDigits = len(info.WholeStr)
		if strings.HasPrefix(info.WholeStr, "-") {
			info.WholeDigits--
		}

		if len(parts) == 1 || len(parts[1]) == 0 {
			continue
		}

		info.FractionalStr = "." + parts[1]
		info.FractionalDigits = len(parts[1])
		if strings.HasPrefix(info.WholeStr, "-") {
			info.FractionalStr = "-" + info.FractionalStr
		}
		if info.FractionalDigits > maxFractionalDigits {
			maxFractionalDigits = info.FractionalDigits
		}
	}

	return rv, maxFractionalDigits
}

// Parse will parse the given arg into its whole and fractional parts using the provided precision for the fractional part.
func (a *argInfo) Parse(precision uint) (*big.Int, *big.Float, error) {
	var whole *big.Int
	if a.WholeDigits > 0 {
		var ok bool
		whole, ok = new(big.Int).SetString(a.WholeStr, 0)
		if !ok {
			if a.FractionalDigits > 0 {
				return nil, nil, fmt.Errorf("could not parse %q as float: invalid integer part", a.Orig)
			}
			return nil, nil, fmt.Errorf("could not parse %q as integer", a.Orig)
		}
		// If it was just zero, we can just ignore it.
		if whole.Sign() == 0 {
			whole = nil
		}
	}

	var fractional *big.Float
	if a.FractionalDigits > 0 {
		var err error
		fractional, _, err = big.ParseFloat(a.FractionalStr, 0, precision, big.ToNearestEven)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse %q as float: invalid fractional part: %w", a.Orig, err)
		}
		// If it was just zero, we can just ignore it.
		if fractional.Sign() == 0 {
			fractional = nil
		}
	}

	return whole, fractional, nil
}

// logArg will output the provided arg info to stderr iff verbose is enabled.
func logArg(arg *argInfo, whole *big.Int, fractional *big.Float, fractionalDigits int) {
	if !Verbose {
		return
	}
	wholeStr, fractStr, prec := "", "", uint(0)
	if whole != nil {
		wholeStr = whole.String()
	}
	if fractional != nil {
		fractStr = fractional.Text('f', arg.FractionalDigits) + strings.Repeat(" ", fractionalDigits-arg.FractionalDigits)
		prec = fractional.Prec()
	}
	stderrPrintf("+ %25s %25s (%2d,prec=%d) %q", wholeStr, fractStr, arg.FractionalDigits, prec, arg.Orig)
}

// logSum will output the provided sum to stderr iff verbose is enabled.
func logSum(whole *big.Int, fractional *big.Float, fractionalDigits int) {
	if !Verbose {
		return
	}
	wholeStr, fractStr, prec := "", "", uint(0)
	if whole != nil {
		wholeStr = whole.String()
	}
	if fractional != nil {
		fractStr = fractional.Text('f', fractionalDigits)
		prec = fractional.Prec()
	}
	stderrPrintf("= %25s %25s (%2d,prec=%d)", wholeStr, fractStr, fractionalDigits, prec)
}

// Sum4 will parse each arg as a number and return a sum of all those numbers as a converted to a string.
// This version adds all the ints, then all the int parts of the floats and lastly the fractional part of the floats.
func Sum4(args []string) (string, error) {
	// First, add any integers and tuck any floats away for later.
	var totalInt *big.Int
	var floats []string
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}
		if strings.Contains(arg, ".") {
			floats = append(floats, arg)
			continue
		}
		num, ok := new(big.Int).SetString(strings.ReplaceAll(arg, ",", ""), 0)
		if !ok {
			return "", fmt.Errorf("could not parse %q as integer", arg)
		}
		verbosef("+ %25s %q", num, arg)
		if totalInt == nil {
			totalInt = num
		} else {
			totalInt.Add(totalInt, num)
		}
		verbosef("= %25s", totalInt)
	}

	if len(floats) == 0 {
		if totalInt != nil {
			return totalInt.String(), nil
		}
		return "0", nil
	}

	// Add the integer portion of all the floats and identify the fractional parts for later.
	fractionals := make([]string, len(floats))
	maxDigits := 0
	haveFractional := false
	for i, arg := range floats {
		parts := strings.SplitN(arg, ".", 2)
		if len(parts[1]) > maxDigits {
			maxDigits = len(parts[1])
		}
		if strings.TrimRight(parts[1], "0") != "" {
			haveFractional = true
			if strings.HasPrefix(arg, "-") {
				fractionals[i] = "-." + parts[1]
			} else {
				fractionals[i] = "." + parts[1]
			}
		}

		if len(parts[0]) == 0 || parts[0] == "-" || parts[0] == "0" || parts[0] == "-0" {
			continue
		}

		num, ok := new(big.Int).SetString(strings.ReplaceAll(parts[0], ",", ""), 0)
		if !ok {
			return "", fmt.Errorf("could not parse %q as float: invalid integer part", arg)
		}
		verbosef("+ %25s from %q", num, arg)

		if totalInt == nil {
			totalInt = num
		} else {
			totalInt.Add(totalInt, num)
		}
		verbosef("= %25s", totalInt)
	}

	if !haveFractional {
		// If we're here but didn't actually have any fractional portions, It means at least
		// one number was provided with one or more zeros after the decimal. We want those
		// included in the result, but we don't have to actually do all the math.
		return totalInt.String() + "." + strings.Repeat("0", maxDigits), nil
	}

	// Because we're not including the whole portions, we don't need as much precision per digit as the others.
	// The tests all pass at 3.75 too, but a few fail at 3.66, so I went with 4 just to be a bit safe.
	prec := uint(maxDigits * 4) // precForLen(maxDigits)
	var totalFloat *big.Float
	for i, fract := range fractionals {
		if len(fract) == 0 {
			continue
		}

		fractional, _, err := big.ParseFloat(fract, 0, prec, big.ToNearestEven)
		if err != nil {
			return "", fmt.Errorf("could not parse %q as float: invalid fractional part: %w", floats[i], err)
		}
		if Verbose {
			digits := len(fract) - 1 // Subtract one for the ".".
			if fract[0] == '-' {
				digits--
			}
			fractStr := fractional.Text('f', digits) + strings.Repeat(" ", maxDigits-digits)
			stderrPrintf("+ %25s %25s (%2d,prec=%d) from %q", "", fractStr, digits, prec, floats[i])
		}

		if totalFloat == nil {
			totalFloat = fractional
		} else {
			totalFloat.Add(totalFloat, fractional)
		}
		totalInt, totalFloat = normalizeAmounts4(totalInt, totalFloat)
		logSum(totalInt, totalFloat, maxDigits)
	}

	return getCombinedString(totalInt, totalFloat, maxDigits), nil
}

// precForLen returns a safe precision that can be used to represent the provided number of digits.
func precForLen(digits int) uint {
	// Through trial and error, it seems like precision should go up 7 for each digit provided.
	// Once I got to 7, all my unit tests finally passed. Before that, there were rounding errors
	// affecting up to 5 digits. Also, I'm erring on the side of too much precision since I'm
	// pretty sure more precision means it's more likely to get the correct answer.
	return uint(digits * 7)
}

// normalizeAmounts4 takes the integer portion out of the float and adds it to the whole.
// The sum of the returned numbers will equal the sum of the provided numbers.
// If fractional == nil, nothing changes. If whole is nil, |fractional| must be at least five
// before it gets split. If there's already a whole amount, though, all integer portions of
// fractional are moved to the whole amount.
func normalizeAmounts4(whole *big.Int, fractional *big.Float) (*big.Int, *big.Float) {
	if fractional == nil {
		return whole, fractional
	}

	fractSign := fractional.Sign()
	cutoff := getFractCutoff(whole, fractional.Prec(), fractSign)
	if fractSign == 0 || (fractSign > 0 && fractional.Cmp(cutoff) < 0) || (fractSign < 0 && fractional.Cmp(cutoff) > 0) {
		return whole, fractional
	}

	fractInt, _ := fractional.Int(nil)
	if whole == nil {
		whole = fractInt
	} else {
		whole.Add(whole, fractInt)
	}
	fractional.Sub(fractional, new(big.Float).SetPrec(fractional.Prec()).SetInt(fractInt))
	return whole, fractional
}

// getFractCutoff returns the cutoff for moving integer portions from the float to int.
// If whole isn't nil, returns 1 or -1, otherwise 5 or -5 depending on the provided sign.
func getFractCutoff(whole *big.Int, prec uint, sign int) *big.Float {
	if whole != nil {
		return getOneFloat(prec, sign)
	}
	return getFiveFloat(prec, sign)
}

var oneFloat, oneFloatNeg *big.Float

// getOneFloat returns oneFloat if sign >= 0, or oneFloatNeg otherwise, creating them first if needed.
func getOneFloat(prec uint, sign int) *big.Float {
	if sign < 0 {
		if oneFloatNeg == nil {
			oneFloatNeg = new(big.Float).SetPrec(prec).SetInt64(-1)
		}
		return oneFloatNeg
	}
	if oneFloat == nil {
		oneFloat = new(big.Float).SetPrec(prec).SetInt64(1)
	}
	return oneFloat
}

var fiveFloat, fiveFloatNeg *big.Float

// getFiveFloat returns fiveFloat if sign >= 0, or fiveFloatNeg otherwise, creating them first if needed.
func getFiveFloat(prec uint, sign int) *big.Float {
	if sign < 0 {
		if fiveFloatNeg == nil {
			fiveFloatNeg = new(big.Float).SetPrec(prec).SetInt64(-5)
		}
		return fiveFloatNeg
	}
	if fiveFloat == nil {
		fiveFloat = new(big.Float).SetPrec(prec).SetInt64(5)
	}
	return fiveFloat
}

// getCombinedString will return a string that is the result of totalInt + totalFloat.
func getCombinedString(totalInt *big.Int, totalFloat *big.Float, fractDigits int) string {
	totalInt, totalFloat = normalizeAmounts4(totalInt, totalFloat)

	// If the whole part is zero, act like we don't have one so that we just use the fractional total.
	// We don't do this for the fractional total, though, because if there were any fractional parts
	// involved, we want the decimals in the result (even if they're all zeros).
	if totalInt != nil && totalInt.Sign() == 0 {
		totalInt = nil
	}

	// Handle the simple cases where we have only one (or zero) parts.
	switch {
	case totalInt == nil && totalFloat == nil:
		return "0"
	case totalFloat == nil:
		if fractDigits > 0 {
			return totalInt.String() + "." + strings.Repeat("0", fractDigits)
		}
		return totalInt.String()
	case totalInt == nil:
		return totalFloat.Text('f', fractDigits)
	}

	// Okay, we've got both a whole and fractional part.
	// We're going to use concatenation to put them together, so the fractional total
	// needs to not have a whole portion, and they need to both have the same sign.

	// If they've got different signs, we need to change the sign of the fractional
	// portion by removing one from it and giving it to the whole total.
	floatSign := totalFloat.Sign()
	if floatSign != 0 && totalInt.Sign() != floatSign {
		// E.g. If fractional = 0.3, fixer = 1 so that fractional becomes 0.3 - 1 = -0.7.
		// E.g. If fractional = -0.3, fixer = -1 so that fractional becomes -0.3 - -1 = -0.3 + 1 = 0.7.
		// Then, whatever was subtracted from the partial needs to be added to the whole to keep the sum the same.
		fixer := new(big.Int).SetInt64(1)
		if floatSign < 0 {
			fixer.Neg(fixer)
		}
		totalFloat.Sub(totalFloat, new(big.Float).SetPrec(totalFloat.Prec()).SetInt(fixer))
		totalInt.Add(totalInt, fixer)
	}

	// Now we can get the strings and concatenate them.
	wholePart := totalInt.String()
	fractionalPart := totalFloat.Text('f', fractDigits)
	verbosef("combining %q and %q for solution.", wholePart, fractionalPart)
	fractionalPartParts := strings.Split(fractionalPart, ".")
	return wholePart + "." + fractionalPartParts[1]
}

// mainE is the actual runner of this program, possibly returning an error.
func mainE(argsIn []string, stdout io.Writer, stdin io.Reader) error {
	args, stopNow, err := processFlags(argsIn, stdout, stdin)
	if stopNow || err != nil {
		return err
	}

	answer, err := Sum(args.Values)
	if err != nil {
		return err
	}
	if args.Pretty {
		answer = MakeNumberPretty(answer)
	}
	fmt.Fprintln(stdout, answer)
	return nil
}

// sumParams are the parameters defined by command-line arguments on how to behave and execute.
type sumParams struct {
	Values []string
	Pretty bool
}

// processFlags will handle all the flags in the provided args. It will also read stdin if called for.
func processFlags(argsIn []string, stdout io.Writer, stdin io.Reader) (*sumParams, bool, error) {
	rv := &sumParams{}
	verbosef("args provided (%d):", len(argsIn))
	for i := 0; i < len(argsIn); i++ {
		rawArg := argsIn[i]
		arg := strings.TrimSpace(rawArg)
		switch {
		case equalFoldOneOf(arg, "--help", "-h", "help"):
			verbosef("[%d]: help arg identified, %q", i, rawArg)
			PrintUsage(stdout)
			return nil, true, nil
		case equalFoldOneOf(arg, "--pretty", "-p"):
			verbosef("[%d]: pretty arg identified, %q", i, rawArg)
			rv.Pretty = true
		case equalFoldOneOf(arg, "--verbose", "-v"):
			Verbose = true
			verbosef("[%d]: verbose flag identified, %q", i, rawArg)
		case equalFoldOneOf(arg, "--pipe", "-p", "-"):
			verbosef("[%d]: pipe flag identified, %q", i, rawArg)
			newArgs, err := readStdin(stdin)
			if err != nil {
				return nil, true, err
			}
			stdin = nil
			rv.Values = append(rv.Values, newArgs...)
		default:
			verbosef("[%d]: number identified, %q", i, rawArg)
			rv.Values = append(rv.Values, strings.Fields(arg)...)
		}
	}

	if len(rv.Values) == 0 {
		if stdin != nil {
			// If we have stdin, and no other args were provided, we get everything from the pipe.
			verbosef("no args provided, using pipe.")
			newArgs, err := readStdin(stdin)
			if err != nil {
				return nil, true, err
			}
			rv.Values = append(rv.Values, newArgs...)
		} else {
			// If we don't have stdin, and no args were provided, print help.
			verbosef("no args provided, and no pipe either.")
			PrintUsage(stdout)
			return nil, true, nil
		}
	}

	return rv, false, nil
}

// readStdin reads all possible info from sdtdin.
func readStdin(stdin io.Reader) ([]string, error) {
	if stdin == nil {
		return nil, errors.New("no stdin available")
	}

	var rv []string
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		line := scanner.Text()
		rv = append(rv, strings.Fields(line)...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from stdin: %w", err)
	}

	return rv, nil
}

// equalFoldOneOf returns true of one of the provided options is equal to the arg (ignoring case).
func equalFoldOneOf(arg string, options ...string) bool {
	for _, opt := range options {
		if strings.EqualFold(arg, opt) {
			return true
		}
	}
	return false
}

// MakeNumberPretty takes in a number string and adds commas to the whole part.
// Examples: "1234567" -> "1,234,567", "12345.678901" -> "12,345.678901"
// If the string already has commas, or has more than one period, the provided value is returned unchanged.
func MakeNumberPretty(val string) string {
	if len(val) <= 3 || strings.Contains(val, ",") {
		return val
	}
	parts := strings.Split(val, ".")
	if len(parts) == 0 || len(parts) > 2 {
		return val
	}

	wholePart := parts[0]
	hasNeg := len(wholePart) > 0 && wholePart[0] == '-'
	if hasNeg {
		wholePart = wholePart[1:]
	}

	if len(wholePart) > 3 {
		lenLhs := len(wholePart)
		lhs := make([]rune, 0, lenLhs+(lenLhs-1)/3+1)
		if hasNeg {
			lhs = append(lhs, '-')
		}
		for i, digit := range wholePart {
			if i > 0 && (lenLhs-i)%3 == 0 {
				lhs = append(lhs, ',')
			}
			lhs = append(lhs, digit)
		}
		parts[0] = string(lhs)
	}

	return strings.Join(parts, ".")
}

// Verbose keeps track of whether verbose output is enabled.
var Verbose bool

// verbosef prints the provided message to stderr if verbose output is enabled. If not enabled, this is a no-op.
func verbosef(format string, args ...interface{}) {
	if Verbose {
		stderrPrintf(format, args...)
	}
}

// stderrPrintf prints the provided stuff to stderr.
func stderrPrintf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// isCharDev returns true if the provided file is a character device.
// This essentially returns true if there's stuff being piped in.
func isCharDev(stdin *os.File) bool {
	stat, err := stdin.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) == 0
}

func main() {
	if val, ok := os.LookupEnv("VERBOSE"); ok {
		Verbose, _ = strconv.ParseBool(val)
		verbosef("verbose environment variable detected")
	}
	var stdin io.Reader
	if isCharDev(os.Stdin) {
		stdin = os.Stdin
	}
	if err := mainE(os.Args[1:], os.Stdout, stdin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
