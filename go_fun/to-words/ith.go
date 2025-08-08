package to_words

import (
	"strconv"
	"strings"
)

// Ith returns a string of the provide number with the sequence suffix added.
// E.g. 1 -> "1st", 2 -> "2nd", 3 -> "3rd"
func Ith(n int) string {
	nStr := strconv.Itoa(n)
	sfx := "th"
	switch {
	case strings.HasSuffix(nStr, "1") && !strings.HasSuffix(nStr, "11"):
		sfx = "st"
	case strings.HasSuffix(nStr, "2") && !strings.HasSuffix(nStr, "12"):
		sfx = "nd"
	case strings.HasSuffix(nStr, "3") && !strings.HasSuffix(nStr, "13"):
		sfx = "rd"
	}
	return nStr + sfx
}

// IthWords converts the provided number to words and makes it a sequence number.
// E.g. 1 -> "first", 32 -> "thirty-second".
func IthWords(n int) string {
	return WordsToIth(IntToWords(n))
}

// seqWords is a map of number string (e.g. "one") to sequence number (e.g. "first").
var seqWords = map[string]string{
	"one":     "first",
	"two":     "second",
	"three":   "third",
	"five":    "fifth",
	"eight":   "eighth",
	"nine":    "ninth",
	"twelve":  "twelfth",
	"twenty":  "twentieth",
	"thirty":  "thirtieth",
	"forty":   "fortieth",
	"fifty":   "fiftieth",
	"sixty":   "sixtieth",
	"seventy": "seventieth",
	"eighty":  "eightieth",
	"ninety":  "ninetieth",
}

// WordsToIth converts the output of one of the *ToWords functions to a sequence number.
// E.g. "one" -> "first", "thirty-two" -> "thirty-second".
func WordsToIth(val string) string {
	if len(val) == 0 {
		return ""
	}
	for num, ith := range seqWords {
		if strings.HasSuffix(val, num) {
			return strings.TrimSuffix(val, num) + ith
		}
	}
	// All other possible ending words just add "th".
	return val + "th"
}
