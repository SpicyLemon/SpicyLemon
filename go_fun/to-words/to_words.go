package to_words

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// IntToWords converts the provided number into English words.
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
// See also: UintToWords, StringToWords.
func IntToWords(num int) string {
	if num < 0 && num != math.MinInt {
		// Can't negate a min int because it doesn't fit back into an int.
		// For all others, we can do the normal thing and tack on "negative".
		return "negative " + IntToWords(-num)
	}

	// Handle the one digit cases and the two digit cases that use a single word.
	switch num {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	case 10:
		return "ten"
	case 11:
		return "eleven"
	case 12:
		return "twelve"
	case 13:
		return "thirteen"
	case 14:
		return "fourteen"
	case 15:
		return "fifteen"
	case 16:
		return "sixteen"
	case 17:
		return "seventeen"
	case 18:
		return "eighteen"
	case 19:
		return "nineteen"
	case 20:
		return "twenty"
	case 30:
		return "thirty"
	case 40:
		return "forty"
	case 50:
		return "fifty"
	case 60:
		return "sixty"
	case 70:
		return "seventy"
	case 80:
		return "eighty"
	case 90:
		return "ninety"
	}

	// Handle the rest of the 2 digit cases.
	if num > 0 && num < 100 {
		// We know the tens digit isn't 0 or 1 since that's handled in the switch above.
		// We also know the ones digit isn't zero for the same reason.
		ones := num % 10
		return IntToWords(num-ones) + "-" + IntToWords(ones)
	}

	// Handle all the three digit cases.
	if num > 0 && num < 1000 {
		// We know the hundreds digit isn't zero since that's handled in the above if block.
		lhs := IntToWords(num/100) + " hundred"
		rhsv := num % 100
		if rhsv == 0 {
			return lhs
		}
		return lhs + " " + IntToWords(rhsv)
	}

	// Handle anything over three digits by breaking it up into groups of three digits,
	// getting the words for those three and adding the quantifiers to each.
	// We know GroupsToWords won't return an error because there's no way an int has too many groups.
	rv, _ := GroupsToWords(IntToGroups(num))

	return rv
}

// UintToWords converts the provided number into English words.
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
// See also: IntToWords, StringToWords.
func UintToWords(num uint) string {
	if num <= math.MaxInt {
		return IntToWords(int(num))
	}
	return MustStringToWords(strconv.FormatUint(uint64(num), 10))
}

// StringToWords converts the provided number (in string form) into English words.
//
// Examples:
//   - "0" => "zero"
//   - "5" => "five"
//   - "12" => "twelve"
//   - "80" => "eighty"
//   - "43" => "forty-three"
//   - "111" => "one hundred eleven"
//   - "54321" => "fifty-four thousand three hundred twenty-one"
//   - "-1234" => "negative one thousand two hundred thirty-four"
//
// Returns an error if the provided string is not a number.
// See also: MustStringToWords, IntToWords, UintToWords.
func StringToWords(str string) (string, error) {
	groups, err := StringToGroups(str)
	if err != nil {
		return "", err
	}
	rv, err := GroupsToWords(groups)
	if err != nil {
		return "", fmt.Errorf("could not convert %q to words: %w", str, err)
	}
	return rv, nil
}

// MustStringToWords converts the provided number (in string form) into English words.
//
// Examples:
//   - "0" => "zero"
//   - "5" => "five"
//   - "12" => "twelve"
//   - "80" => "eighty"
//   - "43" => "forty-three"
//   - "111" => "one hundred eleven"
//   - "54321" => "fifty-four thousand three hundred twenty-one"
//   - "-1234" => "negative one thousand two hundred thirty-four"
//
// Panics if the provided string is not a number.
// See also: StringToWords.
func MustStringToWords(str string) string {
	rv, err := StringToWords(str)
	if err != nil {
		panic(err)
	}
	return rv
}

// GroupsToWords converts a slice of groups to words as if it were one whole number.
// e.g. [1, 2, 3] => "one million two thousand three".
// If the number is to be negative, only groups[0] should be negative. Making any other
// entries negative will cause the word "negative" to appear in weird places.
// Maximum number of groups is 16 (for quattuordecillion).
// Returns an error if there are zero groups or more groups than there are quantifiers.
// See also: MustGroupsToWords.
func GroupsToWords(groups []int16) (string, error) {
	quants, err := GetQuantifiers(len(groups))
	if err != nil {
		return "", err
	}
	groupWords := make([]string, 0, len(groups))
	for i, group := range groups {
		if group == 0 && i != 0 {
			continue
		}
		gw := IntToWords(int(group))
		if len(quants[i]) > 0 {
			gw += " " + quants[i]
		}
		groupWords = append(groupWords, gw)
	}
	return strings.Join(groupWords, " "), nil
}

// MustGroupsToWords converts a slice of groups to words as if it were one whole number.
// e.g. [1, 2, 3] => "one million two thousand three".
// If the number is to be negative, only groups[0] should be negative. Making any other
// entries negative will cause the word "negative" to appear in weird places.
// Maximum number of groups is 16 (for quattuordecillion).
// Panics if there are zero groups or more groups than there are quantifiers.
// See also: GroupsToWords.
func MustGroupsToWords(groups []int16) string {
	rv, err := GroupsToWords(groups)
	if err != nil {
		panic(err)
	}
	return rv
}

// IntToGroups divides up the provided num into groups of up to three digits.
// The first group might come from fewer than three digits, but the rest come from three.
// e.g. 12345 => [12, 345]
// If num is negative, the first group will be negative, but the rest will be positive (or zero).
func IntToGroups(num int) []int16 {
	if num == 0 {
		return []int16{0}
	}
	if num > -1000 && num < 1000 {
		return []int16{int16(num)}
	}

	isNeg := num < 0

	// Add groups from right to left since that math is easier.
	var rv []int16
	for num != 0 {
		group := num % 1000
		num = num / 1000
		if num < 0 {
			// This negation happens here because a negated min int won't fit into an int,
			// so we can't negate the num until after it's been reduced. Also note that
			// we don't use the isNeg bool here since we only want to do this once.
			group = -group
			num = -num
		}
		rv = append(rv, int16(group))
	}

	// And reverse them so they're back in the same order as num.
	slices.Reverse(rv)

	// Negate the first entry if the num is negative.
	if isNeg {
		rv[0] *= -1
	}

	return rv
}

// wholeNumRx matches a positive or negative whole number.
var wholeNumRx = regexp.MustCompile(`^-?[[:digit:]]+$`)

// StringToGroups divides up the provided number string into groups of up to three digits.
// The first group might come from fewer than three digits, but the rest come from three.
// e.g. "12345" => [12, 345]
// If the number is negative, the first group will be negative, but the rest will be positive (or zero).
// Returns an error if there's a problem parsing the string into numbers.
// See also: MustStringToGroups.
func StringToGroups(str string) ([]int16, error) {
	if !wholeNumRx.MatchString(str) {
		return nil, fmt.Errorf("cannot split %q into groups: not a number", str)
	}

	isNeg := strings.HasPrefix(str, "-")
	str = strings.TrimPrefix(str, "-")

	lhsLen := len(str) % 3
	if lhsLen == 0 {
		lhsLen = 3
	}
	lhs := str[:lhsLen]
	rhs := str[lhsLen:]
	groupStrs := make([]string, 0, 1+len(rhs)/3)
	groupStrs = append(groupStrs, lhs)
	for i := range len(rhs) / 3 {
		groupStrs = append(groupStrs, strings.TrimLeft(rhs[i*3:i*3+3], "0"))
	}

	rv := make([]int16, len(groupStrs))
	for i, group := range groupStrs {
		if len(group) == 0 {
			rv[i] = 0
			continue
		}
		// It shouldn't be possible to get these errors since we know str passes the wholeNumberRx, but just in case...
		val, err := strconv.Atoi(group)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q (from %q) into integer: %w", group, str, err)
		}
		if val < 0 || val > 1000 {
			return nil, fmt.Errorf("invalid value %d from group %q (from %q): must be between 0 and 999", val, group, str)
		}
		rv[i] = int16(val)
	}

	if isNeg {
		rv[0] *= -1
	}

	return rv, nil
}

// MustStringToGroups divides up the provided number string into groups of up to three digits.
// The first group might come from fewer than three digits, but the rest come from three.
// e.g. "12345" => [12, 345]
// If number is negative, the first group will be negative, but the rest will be positive (or zero).
// Panics if there's a problem parsing the string into numbers.
// See also: StringToGroups.
func MustStringToGroups(str string) []int16 {
	rv, err := StringToGroups(str)
	if err != nil {
		panic(err)
	}
	return rv
}

// Quantifiers are the words we add to groups of three digits to differentiate them.
var Quantifiers = []string{
	"quattuordecillion",
	"tredecillion",
	"duodecillion",
	"undecillion",
	"decillion",
	"nonillion",
	"octillion",
	"septillion",
	"sextillion",
	"quintillion",
	"quadrillion",
	"trillion",
	"billion",
	"million",
	"thousand",
	"",
}

// GetQuantifiers gets the quantifiers to use for the provided number of groups.
// They are in big-endian order, e.g. if groupCount = 3, this returns ["million", "thousand", ""].
// Returns an error if the groupCount is negative or more than the number of known quantifiers.
// See also: MustGetQuantifiers.
func GetQuantifiers(groupCount int) ([]string, error) {
	if groupCount <= 0 || groupCount > len(Quantifiers) {
		return nil, fmt.Errorf("cannot get quantifiers for %d groups: must be between 1 and %d", groupCount, len(Quantifiers))
	}
	return Quantifiers[len(Quantifiers)-groupCount:], nil
}

// MustGetQuantifiers gets the quantifiers to use for the provided number of groups.
// They are in big-endian order, e.g. if groupCount = 3, this returns ["million", "thousand", ""].
// Panics if the groupCount is negative or more than the number of known quantifiers.
// See also: GetQuantifiers.
func MustGetQuantifiers(groupCount int) []string {
	rv, err := GetQuantifiers(groupCount)
	if err != nil {
		panic(err)
	}
	return rv
}

// GetQuantifier gets the quantifier for the provided groupID.
// A groupID of 0 is the right-most set of 3 digits in a number, so the quantifier is "".
// A groupID of 1 returns "thousand", 2 returns "million" etc.
// Returns an error if the groupID is negative or more than the number of known quantifiers.
// See also: MustGetQuantifier.
func GetQuantifier(groupID int) (string, error) {
	if groupID < 0 || groupID >= len(Quantifiers) {
		return "", fmt.Errorf("cannot get quantifiers for group %d: must be between 0 and %d", groupID, len(Quantifiers)-1)
	}
	return Quantifiers[len(Quantifiers)-1-groupID], nil
}

// MustGetQuantifier gets the quantifier for the provided groupID.
// A groupID of 0 is the right-most set of 3 digits in a number, so the quantifier is "".
// A groupID of 1 returns "thousand", 2 returns "million" etc.
// Panics if the groupID is negative or more than the number of known quantifiers.
// See also: GetQuantifier.
func MustGetQuantifier(groupID int) string {
	rv, err := GetQuantifier(groupID)
	if err != nil {
		panic(err)
	}
	return rv
}
