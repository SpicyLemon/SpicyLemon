package to_words

import (
	"fmt"
	"regexp"
)

// IntToSpoken converts the provided number into English words as we'd speak them.
// This is the same thing as IntToWords.
//
// Examples:
//   - 0 => "zero"
//   - 5 => "five"
//   - 12 => "twelve"
//   - 80 => "eighty"
//   - 43 => "forty-three"
//   - 111 => "one hundred eleven"
//   - 54,321 => "fifty-four thousand three hundred twenty-one"
//   - -1234 => "negative one thousand two hundred thirty-four"
//
// See also: UintToSpoken, FloatToSpoken, ScientificToSpoken, StringToSpoken.
func IntToSpoken(num int) string {
	return IntToWords(num)
}

// UintToSpoken converts the provided number into English words as we'd speak them.
// This is the same thing as UintToWords.
//
// Examples:
//   - 0 => "zero"
//   - 5 => "five"
//   - 12 => "twelve"
//   - 80 => "eighty"
//   - 43 => "forty-three"
//   - 111 => "one hundred eleven"
//   - 54,321 => "fifty-four thousand three hundred twenty-one"
//
// See also: IntToSpoken, FloatToSpoken, ScientificToSpoken, StringToSpoken.
func UintToSpoken(num uint) string {
	return UintToWords(num)
}

// floatRx matches a floating point number.
// Matches: "<whole>.<fract>"  ".<fract>"  "<whole>." "<whole>"
//
// Match groups:
//
// 1. The provided string without any leading negative indicator.
// 2. The whole number if there is only a whole number portion.
// 3. The whole number if there is also a fractional number portion.
// 4. The fractional number.
var floatRx = regexp.MustCompile(`^((-?[[:digit:]]+)\.?|(-?[[:digit:]]*)\.([[:digit:]]+))$`)

// FloatToSpoken converts the provided floating point number into English words as we'd speak them.
// The whole portion is fully expanded as words; the fractional portion is converted to just the digit words.
//
// Examples:
//   - "0" => "zero"
//   - "5." => "five"
//   - ".5" => "point five"
//   - "1.5" => "one point five"
//   - "-10.713" => "negative ten point seven one three"
//   - "54321.987" => "fifty-four thousand three hundred twenty-one point nine eight seven"
//
// Returns an error if the provided string is not a number.
// See also: MustFloatToSpoken, IntToSpoken, UintToSpoken, ScientificToSpoken, StringToSpoken.
func FloatToSpoken(str string) (string, error) {
	matches := floatRx.FindAllStringSubmatch(str, 1)
	if len(matches) == 0 {
		return "", fmt.Errorf("not a float %q", str)
	}

	// If the 2nd group has a match, there's only got a whole number portion.
	if len(matches[0][2]) > 0 {
		return StringToWords(matches[0][2])
	}

	// No match on 2nd group, there must be a fractional portion.
	wholeNum := matches[0][3]
	fractNum := matches[0][4]

	whole := ""
	switch {
	case wholeNum == "-":
		whole = "negative"
	case len(wholeNum) > 0:
		var err error
		whole, err = StringToWords(wholeNum)
		if err != nil {
			return "", fmt.Errorf("invalid float %q: invalid whole part: %w", str, err)
		}
	}

	fract := StringToDigits("." + fractNum)

	if len(whole) == 0 || len(fract) == 0 {
		return whole + fract, nil
	}
	return whole + " " + fract, nil
}

// MustFloatToSpoken converts the provided floating point number into English words as we'd speak them.
// The whole portion is fully expanded as words; the fractional portion is converted to just the digit words.
//
// Examples:
//   - "0" => "zero"
//   - "5." => "five"
//   - ".5" => "point five"
//   - "1.5" => "one point five"
//   - "10.713" => "ten point seven one three"
//   - "54321.987" => "fifty-four thousand three hundred twenty-one point nine eight seven"
//
// Panics if the provided string is not a number.
// See also: FloatToSpoken, IntToSpoken, UintToSpoken, ScientificToSpoken, StringToSpoken.
func MustFloatToSpoken(str string) string {
	rv, err := FloatToSpoken(str)
	if err != nil {
		panic(err)
	}
	return rv
}

// scientificRx matches a string that ends with a scientific notation as we'd speak them.
// Matches: "<base>e<exponent>"  "<base>E<exponent>"  "<base>x10^<exponent>"
// "<base>*10^<exponent>"  "<base>x10**<exponent>"  "<base>*10**<exponent>"
//
// Match groups:
//
//  1. The base number (might not be a number).
//  2. The scientific notation indicator (i.e. "e", "E", "x10^", "x10**", "*10^", "*10**").
//  3. The exponent indicator (i.e. "^" "**") (also included in group 2).
//  4. The exponent number (might not be a number).
var scientificRx = regexp.MustCompile(`^(.+)(e|E|[*x]10(\^|\*\*))(.+)$`)

// ScientificToSpoken converts the provided number (in scientific notation) into English words as we'd speak them.
// Both the base and exponent are converted to fully written English words.
//
// Examples:
//
//   - "1e5" => "one times ten to the five"
//   - "12E10 => "twelve times ten to the ten"
//   - "-3*10^4 => "negative three times ten to the four"
//   - "0.454x10^15 => "zero point four five four times ten to the fifteen"
//   - "-1.2*10**3.4 => "negative one point two times ten to the three point four"
//   - "210.567x10**-7.123 => "two hundred ten point five six seven times to the negative seven point one two three"
//
// Returns an error if the provided string is not a number.
// See also: MustScientificToSpoken, IntToSpoken, UintToSpoken, FloatToSpoken, StringToSpoken.
func ScientificToSpoken(str string) (string, error) {
	matches := scientificRx.FindAllStringSubmatch(str, 1)
	if len(matches) != 1 {
		return "", fmt.Errorf("not scientific notation %q", str)
	}

	base := matches[0][1]
	exponent := matches[0][4]

	baseWords, err := FloatToSpoken(base)
	if err != nil {
		return "", fmt.Errorf("invalid base from %q: %w", str, err)
	}

	exponentWords, err := FloatToSpoken(exponent)
	if err != nil {
		return "", fmt.Errorf("invalid exponenet from %q: %w", str, err)
	}

	return baseWords + " times ten to the " + exponentWords, nil
}

// MustScientificToSpoken converts the provided number (in scientific notation) into English words as we'd speak them.
// Both the base and exponent are converted to fully written English words.
//
// Examples:
//
//   - "1e5" => "one times ten to the five"
//   - "12E10 => "twelve times ten to the ten"
//   - "-3*10^4 => "negative three times ten to the four"
//   - "0.454x10^15 => "zero point four five four times ten to the fifteen"
//   - "-1.2*10**3.4 => "negative one point two times ten to the three point four"
//   - "210.567x10**-7.123 => "two hundred ten point five six seven times to the negative seven point one two three"
//
// Panics if the provided string is not a number.
// See also: ScientificToSpoken, IntToSpoken, UintToSpoken, FloatToSpoken, StringToSpoken.
func MustScientificToSpoken(str string) string {
	rv, err := ScientificToSpoken(str)
	if err != nil {
		panic(err)
	}
	return rv
}

// StringToSpoken converts the provided number string into English words as we'd speak them.
// This can be either a whole number, floating point number, or number in scientific notation.
//
// Examples:
//   - "0" => "zero"
//   - ".1" => "point one"
//   - "-2.3" => "negative two point three"
//   - "4e5" => "four times ten to the five"
//
// Returns an error if the provided string is not a convertable number.
// See also: MustStringToSpoken, IntToSpoken, UintToSpoken, FloatToSpoken, ScientificToSpoken.
func StringToSpoken(str string) (string, error) {
	if scientificRx.MatchString(str) {
		return ScientificToSpoken(str)
	}

	return FloatToSpoken(str)
}

// MustStringToSpoken converts the provided number string into English words as we'd speak them.
// This can be either a whole number, floating point number, or number in scientific notation.
//
// Examples:
//   - "0" => "zero"
//   - ".one" => "point one"
//   - "-2.3" => "negative two point three"
//   - "4e5" => "four times ten to the five"
//
// Panics if the provided string is not a convertable number.
// See also: StringToSpoken, IntToSpoken, UintToSpoken, FloatToSpoken, ScientificToSpoken.
func MustStringToSpoken(str string) string {
	rv, err := StringToSpoken(str)
	if err != nil {
		panic(err)
	}
	return rv
}
