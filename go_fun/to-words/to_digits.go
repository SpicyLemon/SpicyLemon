package to_words

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// IntToDigits converts the provided number into English digits.
//
// Examples:
//   - 0 => "zero"
//   - 5 => "five"
//   - 12 => "one two"
//   - 80 => "eight zero"
//   - 43 => "four three"
//   - 111 => "one one one"
//   - 54,321 => "five four three two one"
//   - -1234 => "negative one two three four"
//
// See also: UintToDigits, StringToDigits.
func IntToDigits(num int) string {
	return StringToDigits(strconv.Itoa(num))
}

// UintToDigits converts the provided number into English digits.
//
// Examples:
//   - 0 => "zero"
//   - 5 => "five"
//   - 12 => "one two"
//   - 80 => "eight zero"
//   - 43 => "four three"
//   - 111 => "one one one"
//   - 54,321 => "five four three two one"
//
// See also: IntToDigits, StringToDigits.
func UintToDigits(num uint) string {
	return StringToDigits(strconv.FormatUint(uint64(num), 10))
}

// StringToDigits converts the provided number (in string form) into English digits.
//
// Examples:
//   - "0" => "zero"
//   - "5" => "five"
//   - "12" => "one two"
//   - "80" => "eight zero"
//   - "43" => "four three"
//   - "111" => "one one one"
//   - "54321" => "five four three two one"
//   - "-1234" => "negative one two three four"
//   - "987.456" => "nine eight seven point four five six"
//
// Any characters other than digits, '-', and '.', are preserved as provided.
// See also: IntToDigits, UintToDigits.
func StringToDigits(num string) string {
	words := make([]string, utf8.RuneCountInString(num))
	for i, r := range num {
		switch r {
		case '-':
			words[i] = "negative "
		case '0':
			words[i] = "zero "
		case '1':
			words[i] = "one "
		case '2':
			words[i] = "two "
		case '3':
			words[i] = "three "
		case '4':
			words[i] = "four "
		case '5':
			words[i] = "five "
		case '6':
			words[i] = "six "
		case '7':
			words[i] = "seven "
		case '8':
			words[i] = "eight "
		case '9':
			words[i] = "nine "
		case '.':
			words[i] = "point "
		default:
			words[i] = string(r)
			// Scenario: num = "99 balloons".
			// The "99" gets turned into two instances of "nine ". The space then gets added raw.
			// But now there's two spaces in a row. So let's remove the space at the end of the
			// previous entry. But if num = "99  balloons", we still want both spaces, so we only
			// trim the space from the previous entry if the previous entry is a space itself.
			// This way:
			//  - "99  balloons" becomes "nine nine  balloons"
			//  - "99 balloons" becomes "nine nine ballooons"
			//  - "99balloons" becomes "nine nine balloons"
			if i > 0 && r == ' ' && words[i-1] != " " {
				words[i-1] = strings.TrimSuffix(words[i-1], " ")
			}
		}
	}
	return strings.TrimSpace(strings.Join(words, ""))
}
