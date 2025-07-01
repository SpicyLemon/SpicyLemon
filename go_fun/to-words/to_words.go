package to_words

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// ToWords converts the provided number into English words.
// Examples:
//   - 0 => "zero"
//   - 5 => "five"
//   - 12 => "twelve"
//   - 80 => "eighty"
//   - 43 => "forty-three"
//   - 111 => "one hundred eleven"
//   - 54,321 => "fifty-four thousand three hundred twenty-one"
//   - -1234 => "negative one thousand two hundred thirty-four"
func ToWords(num int) string {
	if num < 0 && num != math.MinInt {
		// Can't negate a min int because it doesn't fit back into an int.
		// For all others, we can do the normal thing and tack on "negative".
		return "negative " + ToWords(-num)
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
		return ToWords(num-ones) + "-" + ToWords(ones)
	}

	// Handle all the three digit cases.
	if num > 0 && num < 1000 {
		// We know the hundreds digit isn't zero since that's handled in the above if block.
		lhs := ToWords(num/100) + " hundred"
		rhsv := num % 100
		if rhsv == 0 {
			return lhs
		}
		return lhs + " " + ToWords(rhsv)
	}

	// Handle anything over three digits by breaking it up into groups of three digits,
	// getting the words for those three and adding the quantifiers to each.
	// We know GroupsToWords won't return an error because there's no way an int has too many groups.
	rv, _ := GroupsToWords(ToGroups(num))

	// When num is min-int, we still need to add "negative" to the result.
	if num < 0 {
		return "negative " + rv
	}
	return rv
}

// GroupsToWords converts a slice of groups to words as if it were one whole number.
// Maximum number of groups is 16 (for quattuordecillion).
// e.g. [1, 2, 3] => "one million two thousand three".
// Returns an error if there are zero groups or more groups than there are quantifiers.
// See also: MustGroupsToWords.
func GroupsToWords(groups []uint16) (string, error) {
	quants, err := GetQuantifiers(len(groups))
	if err != nil {
		return "", err
	}
	groupWords := make([]string, 0, len(groups))
	for i, group := range groups {
		if group == 0 && i != 0 {
			continue
		}
		gw := ToWords(int(group))
		if len(quants[i]) > 0 {
			gw += " " + quants[i]
		}
		groupWords = append(groupWords, gw)
	}
	return strings.Join(groupWords, " "), nil
}

// MustGroupsToWords converts a slice of groups to words as if it were one whole number.
// Maximum number of groups is 16 (for quattuordecillion).
// e.g. [1, 2, 3] => "one million two thousand three".
// Panics if there are zero groups or more groups than there are quantifiers.
// See also: GroupsToWords.
func MustGroupsToWords(groups []uint16) string {
	rv, err := GroupsToWords(groups)
	if err != nil {
		panic(err)
	}
	return rv
}

// ToGroups divides up the provided num into groups of three digits.
// The only entry in the returned value that might not have three digits is the 0th, e.g. 12345 => [12, 345].
// All returned values will be positive or zero (no negatives even if num is negative), e.g. -9001 => [9, 1].
func ToGroups(num int) []uint16 {
	if num == 0 {
		return []uint16{0}
	}
	// Add groups from right to left.
	var rv []uint16
	for num != 0 {
		group := num % 1000
		num = num / 1000
		if group < 0 {
			// We negate things here (instead of at the start) to handle when num is min-int.
			// A negated min-int won't fit in an int. Here we know it'll now fit into an int though.
			group = -group
			num = -num
		}
		rv = append(rv, uint16(group))
	}
	// And reverse them so they're back in the same order as num.
	slices.Reverse(rv)
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
