package to_words

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToWords(t *testing.T) {
	tests := []struct {
		num int
		exp string
	}{
		{num: 0, exp: "zero"},
		{num: 1, exp: "one"},
		{num: -1, exp: "negative one"},
		{num: 2, exp: "two"},
		{num: -2, exp: "negative two"},
		{num: 3, exp: "three"},
		{num: -3, exp: "negative three"},
		{num: 4, exp: "four"},
		{num: -4, exp: "negative four"},
		{num: 5, exp: "five"},
		{num: -5, exp: "negative five"},
		{num: 6, exp: "six"},
		{num: -6, exp: "negative six"},
		{num: 7, exp: "seven"},
		{num: -7, exp: "negative seven"},
		{num: 8, exp: "eight"},
		{num: -8, exp: "negative eight"},
		{num: 9, exp: "nine"},
		{num: -9, exp: "negative nine"},
		{num: 10, exp: "ten"},
		{num: -10, exp: "negative ten"},
		{num: 11, exp: "eleven"},
		{num: -11, exp: "negative eleven"},
		{num: 12, exp: "twelve"},
		{num: -12, exp: "negative twelve"},
		{num: 13, exp: "thirteen"},
		{num: -13, exp: "negative thirteen"},
		{num: 14, exp: "fourteen"},
		{num: -14, exp: "negative fourteen"},
		{num: 15, exp: "fifteen"},
		{num: -15, exp: "negative fifteen"},
		{num: 16, exp: "sixteen"},
		{num: -16, exp: "negative sixteen"},
		{num: 17, exp: "seventeen"},
		{num: -17, exp: "negative seventeen"},
		{num: 18, exp: "eighteen"},
		{num: -18, exp: "negative eighteen"},
		{num: 19, exp: "nineteen"},
		{num: -19, exp: "negative nineteen"},
		{num: 20, exp: "twenty"},
		{num: -20, exp: "negative twenty"},
		{num: 21, exp: "twenty-one"},
		{num: -21, exp: "negative twenty-one"},
		{num: 22, exp: "twenty-two"},
		{num: -22, exp: "negative twenty-two"},
		{num: 23, exp: "twenty-three"},
		{num: -23, exp: "negative twenty-three"},
		{num: 24, exp: "twenty-four"},
		{num: -24, exp: "negative twenty-four"},
		{num: 25, exp: "twenty-five"},
		{num: -25, exp: "negative twenty-five"},
		{num: 26, exp: "twenty-six"},
		{num: -26, exp: "negative twenty-six"},
		{num: 27, exp: "twenty-seven"},
		{num: -27, exp: "negative twenty-seven"},
		{num: 28, exp: "twenty-eight"},
		{num: -28, exp: "negative twenty-eight"},
		{num: 29, exp: "twenty-nine"},
		{num: -29, exp: "negative twenty-nine"},
		{num: 30, exp: "thirty"},
		{num: -30, exp: "negative thirty"},
		{num: 31, exp: "thirty-one"},
		{num: -31, exp: "negative thirty-one"},
		{num: 32, exp: "thirty-two"},
		{num: -32, exp: "negative thirty-two"},
		{num: 33, exp: "thirty-three"},
		{num: -33, exp: "negative thirty-three"},
		{num: 34, exp: "thirty-four"},
		{num: -34, exp: "negative thirty-four"},
		{num: 35, exp: "thirty-five"},
		{num: -35, exp: "negative thirty-five"},
		{num: 36, exp: "thirty-six"},
		{num: -36, exp: "negative thirty-six"},
		{num: 37, exp: "thirty-seven"},
		{num: -37, exp: "negative thirty-seven"},
		{num: 38, exp: "thirty-eight"},
		{num: -38, exp: "negative thirty-eight"},
		{num: 39, exp: "thirty-nine"},
		{num: -39, exp: "negative thirty-nine"},
		{num: 40, exp: "forty"},
		{num: -40, exp: "negative forty"},
		{num: 41, exp: "forty-one"},
		{num: -41, exp: "negative forty-one"},
		{num: 42, exp: "forty-two"},
		{num: -42, exp: "negative forty-two"},
		{num: 43, exp: "forty-three"},
		{num: -43, exp: "negative forty-three"},
		{num: 44, exp: "forty-four"},
		{num: -44, exp: "negative forty-four"},
		{num: 45, exp: "forty-five"},
		{num: -45, exp: "negative forty-five"},
		{num: 46, exp: "forty-six"},
		{num: -46, exp: "negative forty-six"},
		{num: 47, exp: "forty-seven"},
		{num: -47, exp: "negative forty-seven"},
		{num: 48, exp: "forty-eight"},
		{num: -48, exp: "negative forty-eight"},
		{num: 49, exp: "forty-nine"},
		{num: -49, exp: "negative forty-nine"},
		{num: 50, exp: "fifty"},
		{num: -50, exp: "negative fifty"},
		{num: 51, exp: "fifty-one"},
		{num: -51, exp: "negative fifty-one"},
		{num: 52, exp: "fifty-two"},
		{num: -52, exp: "negative fifty-two"},
		{num: 53, exp: "fifty-three"},
		{num: -53, exp: "negative fifty-three"},
		{num: 54, exp: "fifty-four"},
		{num: -54, exp: "negative fifty-four"},
		{num: 55, exp: "fifty-five"},
		{num: -55, exp: "negative fifty-five"},
		{num: 56, exp: "fifty-six"},
		{num: -56, exp: "negative fifty-six"},
		{num: 57, exp: "fifty-seven"},
		{num: -57, exp: "negative fifty-seven"},
		{num: 58, exp: "fifty-eight"},
		{num: -58, exp: "negative fifty-eight"},
		{num: 59, exp: "fifty-nine"},
		{num: -59, exp: "negative fifty-nine"},
		{num: 60, exp: "sixty"},
		{num: -60, exp: "negative sixty"},
		{num: 61, exp: "sixty-one"},
		{num: -61, exp: "negative sixty-one"},
		{num: 62, exp: "sixty-two"},
		{num: -62, exp: "negative sixty-two"},
		{num: 63, exp: "sixty-three"},
		{num: -63, exp: "negative sixty-three"},
		{num: 64, exp: "sixty-four"},
		{num: -64, exp: "negative sixty-four"},
		{num: 65, exp: "sixty-five"},
		{num: -65, exp: "negative sixty-five"},
		{num: 66, exp: "sixty-six"},
		{num: -66, exp: "negative sixty-six"},
		{num: 67, exp: "sixty-seven"},
		{num: -67, exp: "negative sixty-seven"},
		{num: 68, exp: "sixty-eight"},
		{num: -68, exp: "negative sixty-eight"},
		{num: 69, exp: "sixty-nine"},
		{num: -69, exp: "negative sixty-nine"},
		{num: 70, exp: "seventy"},
		{num: -70, exp: "negative seventy"},
		{num: 71, exp: "seventy-one"},
		{num: -71, exp: "negative seventy-one"},
		{num: 72, exp: "seventy-two"},
		{num: -72, exp: "negative seventy-two"},
		{num: 73, exp: "seventy-three"},
		{num: -73, exp: "negative seventy-three"},
		{num: 74, exp: "seventy-four"},
		{num: -74, exp: "negative seventy-four"},
		{num: 75, exp: "seventy-five"},
		{num: -75, exp: "negative seventy-five"},
		{num: 76, exp: "seventy-six"},
		{num: -76, exp: "negative seventy-six"},
		{num: 77, exp: "seventy-seven"},
		{num: -77, exp: "negative seventy-seven"},
		{num: 78, exp: "seventy-eight"},
		{num: -78, exp: "negative seventy-eight"},
		{num: 79, exp: "seventy-nine"},
		{num: -79, exp: "negative seventy-nine"},
		{num: 80, exp: "eighty"},
		{num: -80, exp: "negative eighty"},
		{num: 81, exp: "eighty-one"},
		{num: -81, exp: "negative eighty-one"},
		{num: 82, exp: "eighty-two"},
		{num: -82, exp: "negative eighty-two"},
		{num: 83, exp: "eighty-three"},
		{num: -83, exp: "negative eighty-three"},
		{num: 84, exp: "eighty-four"},
		{num: -84, exp: "negative eighty-four"},
		{num: 85, exp: "eighty-five"},
		{num: -85, exp: "negative eighty-five"},
		{num: 86, exp: "eighty-six"},
		{num: -86, exp: "negative eighty-six"},
		{num: 87, exp: "eighty-seven"},
		{num: -87, exp: "negative eighty-seven"},
		{num: 88, exp: "eighty-eight"},
		{num: -88, exp: "negative eighty-eight"},
		{num: 89, exp: "eighty-nine"},
		{num: -89, exp: "negative eighty-nine"},
		{num: 90, exp: "ninety"},
		{num: -90, exp: "negative ninety"},
		{num: 91, exp: "ninety-one"},
		{num: -91, exp: "negative ninety-one"},
		{num: 92, exp: "ninety-two"},
		{num: -92, exp: "negative ninety-two"},
		{num: 93, exp: "ninety-three"},
		{num: -93, exp: "negative ninety-three"},
		{num: 94, exp: "ninety-four"},
		{num: -94, exp: "negative ninety-four"},
		{num: 95, exp: "ninety-five"},
		{num: -95, exp: "negative ninety-five"},
		{num: 96, exp: "ninety-six"},
		{num: -96, exp: "negative ninety-six"},
		{num: 97, exp: "ninety-seven"},
		{num: -97, exp: "negative ninety-seven"},
		{num: 98, exp: "ninety-eight"},
		{num: -98, exp: "negative ninety-eight"},
		{num: 99, exp: "ninety-nine"},
		{num: -99, exp: "negative ninety-nine"},
		{num: 100, exp: "one hundred"},
		{num: -100, exp: "negative one hundred"},
		{num: 111, exp: "one hundred eleven"},
		{num: -111, exp: "negative one hundred eleven"},
		{num: 54_321, exp: "fifty-four thousand three hundred twenty-one"},
		{num: -54_321, exp: "negative fifty-four thousand three hundred twenty-one"},
		{num: 1_234, exp: "one thousand two hundred thirty-four"},
		{num: -1_234, exp: "negative one thousand two hundred thirty-four"},
		{num: 1_000, exp: "one thousand"},
		{num: -1_000, exp: "negative one thousand"},
		{num: 1_001, exp: "one thousand one"},
		{num: -1_001, exp: "negative one thousand one"},
		{num: 1_020, exp: "one thousand twenty"},
		{num: -1_020, exp: "negative one thousand twenty"},
		{num: 1_300, exp: "one thousand three hundred"},
		{num: -1_300, exp: "negative one thousand three hundred"},
		{num: 1_045, exp: "one thousand forty-five"},
		{num: -1_045, exp: "negative one thousand forty-five"},
		{num: 1_670, exp: "one thousand six hundred seventy"},
		{num: -1_670, exp: "negative one thousand six hundred seventy"},
		{num: 1_809, exp: "one thousand eight hundred nine"},
		{num: -1_809, exp: "negative one thousand eight hundred nine"},
		{num: 2_000, exp: "two thousand"},
		{num: -2_000, exp: "negative two thousand"},
		{num: 2_001, exp: "two thousand one"},
		{num: -2_001, exp: "negative two thousand one"},
		{num: 2_020, exp: "two thousand twenty"},
		{num: -2_020, exp: "negative two thousand twenty"},
		{num: 2_300, exp: "two thousand three hundred"},
		{num: -2_300, exp: "negative two thousand three hundred"},
		{num: 2_045, exp: "two thousand forty-five"},
		{num: -2_045, exp: "negative two thousand forty-five"},
		{num: 2_670, exp: "two thousand six hundred seventy"},
		{num: -2_670, exp: "negative two thousand six hundred seventy"},
		{num: 2_809, exp: "two thousand eight hundred nine"},
		{num: -2_809, exp: "negative two thousand eight hundred nine"},
		{num: 3_000, exp: "three thousand"},
		{num: -3_000, exp: "negative three thousand"},
		{num: 3_001, exp: "three thousand one"},
		{num: -3_001, exp: "negative three thousand one"},
		{num: 3_020, exp: "three thousand twenty"},
		{num: -3_020, exp: "negative three thousand twenty"},
		{num: 3_300, exp: "three thousand three hundred"},
		{num: -3_300, exp: "negative three thousand three hundred"},
		{num: 3_045, exp: "three thousand forty-five"},
		{num: -3_045, exp: "negative three thousand forty-five"},
		{num: 3_670, exp: "three thousand six hundred seventy"},
		{num: -3_670, exp: "negative three thousand six hundred seventy"},
		{num: 3_809, exp: "three thousand eight hundred nine"},
		{num: -3_809, exp: "negative three thousand eight hundred nine"},
		{num: 4_000, exp: "four thousand"},
		{num: -4_000, exp: "negative four thousand"},
		{num: 4_001, exp: "four thousand one"},
		{num: -4_001, exp: "negative four thousand one"},
		{num: 4_020, exp: "four thousand twenty"},
		{num: -4_020, exp: "negative four thousand twenty"},
		{num: 4_300, exp: "four thousand three hundred"},
		{num: -4_300, exp: "negative four thousand three hundred"},
		{num: 4_045, exp: "four thousand forty-five"},
		{num: -4_045, exp: "negative four thousand forty-five"},
		{num: 4_670, exp: "four thousand six hundred seventy"},
		{num: -4_670, exp: "negative four thousand six hundred seventy"},
		{num: 4_809, exp: "four thousand eight hundred nine"},
		{num: -4_809, exp: "negative four thousand eight hundred nine"},
		{num: 5_000, exp: "five thousand"},
		{num: -5_000, exp: "negative five thousand"},
		{num: 5_001, exp: "five thousand one"},
		{num: -5_001, exp: "negative five thousand one"},
		{num: 5_020, exp: "five thousand twenty"},
		{num: -5_020, exp: "negative five thousand twenty"},
		{num: 5_300, exp: "five thousand three hundred"},
		{num: -5_300, exp: "negative five thousand three hundred"},
		{num: 5_045, exp: "five thousand forty-five"},
		{num: -5_045, exp: "negative five thousand forty-five"},
		{num: 5_670, exp: "five thousand six hundred seventy"},
		{num: -5_670, exp: "negative five thousand six hundred seventy"},
		{num: 5_809, exp: "five thousand eight hundred nine"},
		{num: -5_809, exp: "negative five thousand eight hundred nine"},
		{num: 6_000, exp: "six thousand"},
		{num: -6_000, exp: "negative six thousand"},
		{num: 6_001, exp: "six thousand one"},
		{num: -6_001, exp: "negative six thousand one"},
		{num: 6_020, exp: "six thousand twenty"},
		{num: -6_020, exp: "negative six thousand twenty"},
		{num: 6_300, exp: "six thousand three hundred"},
		{num: -6_300, exp: "negative six thousand three hundred"},
		{num: 6_045, exp: "six thousand forty-five"},
		{num: -6_045, exp: "negative six thousand forty-five"},
		{num: 6_670, exp: "six thousand six hundred seventy"},
		{num: -6_670, exp: "negative six thousand six hundred seventy"},
		{num: 6_809, exp: "six thousand eight hundred nine"},
		{num: -6_809, exp: "negative six thousand eight hundred nine"},
		{num: 7_000, exp: "seven thousand"},
		{num: -7_000, exp: "negative seven thousand"},
		{num: 7_001, exp: "seven thousand one"},
		{num: -7_001, exp: "negative seven thousand one"},
		{num: 7_020, exp: "seven thousand twenty"},
		{num: -7_020, exp: "negative seven thousand twenty"},
		{num: 7_300, exp: "seven thousand three hundred"},
		{num: -7_300, exp: "negative seven thousand three hundred"},
		{num: 7_045, exp: "seven thousand forty-five"},
		{num: -7_045, exp: "negative seven thousand forty-five"},
		{num: 7_670, exp: "seven thousand six hundred seventy"},
		{num: -7_670, exp: "negative seven thousand six hundred seventy"},
		{num: 7_809, exp: "seven thousand eight hundred nine"},
		{num: -7_809, exp: "negative seven thousand eight hundred nine"},
		{num: 8_000, exp: "eight thousand"},
		{num: -8_000, exp: "negative eight thousand"},
		{num: 8_001, exp: "eight thousand one"},
		{num: -8_001, exp: "negative eight thousand one"},
		{num: 8_020, exp: "eight thousand twenty"},
		{num: -8_020, exp: "negative eight thousand twenty"},
		{num: 8_300, exp: "eight thousand three hundred"},
		{num: -8_300, exp: "negative eight thousand three hundred"},
		{num: 8_045, exp: "eight thousand forty-five"},
		{num: -8_045, exp: "negative eight thousand forty-five"},
		{num: 8_670, exp: "eight thousand six hundred seventy"},
		{num: -8_670, exp: "negative eight thousand six hundred seventy"},
		{num: 8_809, exp: "eight thousand eight hundred nine"},
		{num: -8_809, exp: "negative eight thousand eight hundred nine"},
		{num: 9_000, exp: "nine thousand"},
		{num: -9_000, exp: "negative nine thousand"},
		{num: 9_001, exp: "nine thousand one"},
		{num: -9_001, exp: "negative nine thousand one"},
		{num: 9_020, exp: "nine thousand twenty"},
		{num: -9_020, exp: "negative nine thousand twenty"},
		{num: 9_300, exp: "nine thousand three hundred"},
		{num: -9_300, exp: "negative nine thousand three hundred"},
		{num: 9_045, exp: "nine thousand forty-five"},
		{num: -9_045, exp: "negative nine thousand forty-five"},
		{num: 9_670, exp: "nine thousand six hundred seventy"},
		{num: -9_670, exp: "negative nine thousand six hundred seventy"},
		{num: 9_809, exp: "nine thousand eight hundred nine"},
		{num: -9_809, exp: "negative nine thousand eight hundred nine"},
		{num: 9_999, exp: "nine thousand nine hundred ninety-nine"},
		{num: -9_999, exp: "negative nine thousand nine hundred ninety-nine"},
		{num: 10_000, exp: "ten thousand"},
		{num: -10_000, exp: "negative ten thousand"},
		{num: 24_745, exp: "twenty-four thousand seven hundred forty-five"},
		{num: -24_745, exp: "negative twenty-four thousand seven hundred forty-five"},
		{num: 99_999, exp: "ninety-nine thousand nine hundred ninety-nine"},
		{num: -99_999, exp: "negative ninety-nine thousand nine hundred ninety-nine"},
		{num: 100_000, exp: "one hundred thousand"},
		{num: -100_000, exp: "negative one hundred thousand"},
		{num: 552_887, exp: "five hundred fifty-two thousand eight hundred eighty-seven"},
		{num: -552_887, exp: "negative five hundred fifty-two thousand eight hundred eighty-seven"},
		{num: 1_000_000, exp: "one million"},
		{num: -1_000_000, exp: "negative one million"},
		{num: 1_002_003, exp: "one million two thousand three"},
		{num: -1_002_003, exp: "negative one million two thousand three"},
		{num: 5_485_065, exp: "five million four hundred eighty-five thousand sixty-five"},
		{num: -5_485_065, exp: "negative five million four hundred eighty-five thousand sixty-five"},
		{num: 10_000_000, exp: "ten million"},
		{num: -10_000_000, exp: "negative ten million"},
		{num: 82_212_496, exp: "eighty-two million two hundred twelve thousand four hundred ninety-six"},
		{num: -82_212_496, exp: "negative eighty-two million two hundred twelve thousand four hundred ninety-six"},
		{num: 100_000_000, exp: "one hundred million"},
		{num: -100_000_000, exp: "negative one hundred million"},
		{num: 100_200_300, exp: "one hundred million two hundred thousand three hundred"},
		{num: -100_200_300, exp: "negative one hundred million two hundred thousand three hundred"},
		{num: 126_490_799,
			exp: "one hundred twenty-six million four hundred ninety thousand seven hundred ninety-nine"},
		{num: -126_490_799,
			exp: "negative one hundred twenty-six million four hundred ninety thousand seven hundred ninety-nine"},
		{num: 1_000_000_000, exp: "one billion"},
		{num: -1_000_000_000, exp: "negative one billion"},
		{num: 9_007_912_442,
			exp: "nine billion seven million nine hundred twelve thousand four hundred forty-two"},
		{num: -9_007_912_442,
			exp: "negative nine billion seven million nine hundred twelve thousand four hundred forty-two"},
		{num: 10_000_000_000, exp: "ten billion"},
		{num: -10_000_000_000, exp: "negative ten billion"},
		{num: 10_000_000_030, exp: "ten billion thirty"},
		{num: -10_000_000_030, exp: "negative ten billion thirty"},
		{num: 10_000_030_000, exp: "ten billion thirty thousand"},
		{num: -10_000_030_000, exp: "negative ten billion thirty thousand"},
		{num: 64_127_772_414,
			exp: "sixty-four billion one hundred twenty-seven million " +
				"seven hundred seventy-two thousand four hundred fourteen"},
		{num: -64_127_772_414,
			exp: "negative sixty-four billion one hundred twenty-seven million " +
				"seven hundred seventy-two thousand four hundred fourteen"},
		{num: 100_000_000_000, exp: "one hundred billion"},
		{num: -100_000_000_000, exp: "negative one hundred billion"},
		{num: 759_528_730_112,
			exp: "seven hundred fifty-nine billion five hundred twenty-eight million " +
				"seven hundred thirty thousand one hundred twelve"},
		{num: -759_528_730_112,
			exp: "negative seven hundred fifty-nine billion five hundred twenty-eight million " +
				"seven hundred thirty thousand one hundred twelve"},
		{num: 1_000_000_000_000, exp: "one trillion"},
		{num: -1_000_000_000_000, exp: "negative one trillion"},
		{num: 9_515_965_217_456,
			exp: "nine trillion five hundred fifteen billion " +
				"nine hundred sixty-five million two hundred seventeen thousand four hundred fifty-six"},
		{num: -9_515_965_217_456,
			exp: "negative nine trillion five hundred fifteen billion " +
				"nine hundred sixty-five million two hundred seventeen thousand four hundred fifty-six"},
		{num: 10_000_000_000_000, exp: "ten trillion"},
		{num: -10_000_000_000_000, exp: "negative ten trillion"},
		{num: 50_558_442_088_500,
			exp: "fifty trillion five hundred fifty-eight billion " +
				"four hundred forty-two million eighty-eight thousand five hundred"},
		{num: -50_558_442_088_500,
			exp: "negative fifty trillion five hundred fifty-eight billion " +
				"four hundred forty-two million eighty-eight thousand five hundred"},
		{num: 100_000_000_000_000, exp: "one hundred trillion"},
		{num: -100_000_000_000_000, exp: "negative one hundred trillion"},
		{num: 875_545_170_963_847,
			exp: "eight hundred seventy-five trillion five hundred forty-five billion " +
				"one hundred seventy million nine hundred sixty-three thousand eight hundred forty-seven"},
		{num: -875_545_170_963_847,
			exp: "negative eight hundred seventy-five trillion five hundred forty-five billion " +
				"one hundred seventy million nine hundred sixty-three thousand eight hundred forty-seven"},
		{num: 1_000_000_000_000_000, exp: "one quadrillion"},
		{num: -1_000_000_000_000_000, exp: "negative one quadrillion"},
		{num: 1_459_010_276_579_858,
			exp: "one quadrillion four hundred fifty-nine trillion " +
				"ten billion two hundred seventy-six million " +
				"five hundred seventy-nine thousand eight hundred fifty-eight"},
		{num: -1_459_010_276_579_858,
			exp: "negative one quadrillion four hundred fifty-nine trillion " +
				"ten billion two hundred seventy-six million " +
				"five hundred seventy-nine thousand eight hundred fifty-eight"},
		{num: 10_000_000_000_000_000, exp: "ten quadrillion"},
		{num: -10_000_000_000_000_000, exp: "negative ten quadrillion"},
		{num: 63_817_328_483_963_713,
			exp: "sixty-three quadrillion eight hundred seventeen trillion " +
				"three hundred twenty-eight billion four hundred eighty-three million " +
				"nine hundred sixty-three thousand seven hundred thirteen"},
		{num: -63_817_328_483_963_713,
			exp: "negative sixty-three quadrillion eight hundred seventeen trillion " +
				"three hundred twenty-eight billion four hundred eighty-three million " +
				"nine hundred sixty-three thousand seven hundred thirteen"},
		{num: 100_000_000_000_000_000, exp: "one hundred quadrillion"},
		{num: -100_000_000_000_000_000, exp: "negative one hundred quadrillion"},
		{num: 503_030_044_673_410_914,
			exp: "five hundred three quadrillion thirty trillion " +
				"forty-four billion six hundred seventy-three million " +
				"four hundred ten thousand nine hundred fourteen"},
		{num: -503_030_044_673_410_914,
			exp: "negative five hundred three quadrillion thirty trillion " +
				"forty-four billion six hundred seventy-three million " +
				"four hundred ten thousand nine hundred fourteen"},
		{num: 1_000_000_000_000_000_000, exp: "one quintillion"},
		{num: -1_000_000_000_000_000_000, exp: "negative one quintillion"},
		{num: 1_002_003_004_005_006_007,
			exp: "one quintillion two quadrillion three trillion " +
				"four billion five million six thousand seven"},
		{num: -1_002_003_004_005_006_007,
			exp: "negative one quintillion two quadrillion three trillion " +
				"four billion five million six thousand seven"},
		{num: 6_095_934_577_086_450_739,
			exp: "six quintillion ninety-five quadrillion nine hundred thirty-four trillion " +
				"five hundred seventy-seven billion eighty-six million " +
				"four hundred fifty thousand seven hundred thirty-nine"},
		{num: -6_095_934_577_086_450_739,
			exp: "negative six quintillion ninety-five quadrillion nine hundred thirty-four trillion " +
				"five hundred seventy-seven billion eighty-six million " +
				"four hundred fifty thousand seven hundred thirty-nine"},
		{num: 126, exp: "one hundred twenty-six"},
		{num: 127, exp: "one hundred twenty-seven"},
		{num: 128, exp: "one hundred twenty-eight"},
		{num: -127, exp: "negative one hundred twenty-seven"},
		{num: -128, exp: "negative one hundred twenty-eight"},
		{num: -129, exp: "negative one hundred twenty-nine"},
		{num: 254, exp: "two hundred fifty-four"},
		{num: 255, exp: "two hundred fifty-five"},
		{num: 256, exp: "two hundred fifty-six"},
		{num: 32_766, exp: "thirty-two thousand seven hundred sixty-six"},
		{num: 32_767, exp: "thirty-two thousand seven hundred sixty-seven"},
		{num: 32_768, exp: "thirty-two thousand seven hundred sixty-eight"},
		{num: -32_767, exp: "negative thirty-two thousand seven hundred sixty-seven"},
		{num: -32_768, exp: "negative thirty-two thousand seven hundred sixty-eight"},
		{num: -32_769, exp: "negative thirty-two thousand seven hundred sixty-nine"},
		{num: 65_534, exp: "sixty-five thousand five hundred thirty-four"},
		{num: 65_535, exp: "sixty-five thousand five hundred thirty-five"},
		{num: 65_536, exp: "sixty-five thousand five hundred thirty-six"},
		{num: 2_147_483_646,
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-six"},
		{num: 2_147_483_647,
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-seven"},
		{num: 2_147_483_648,
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-eight"},
		{num: -2_147_483_647,
			exp: "negative two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-seven"},
		{num: -2_147_483_648,
			exp: "negative two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-eight"},
		{num: -2_147_483_649,
			exp: "negative two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-nine"},
		{num: 4_294_967_294,
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-four"},
		{num: 4_294_967_295,
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-five"},
		{num: 4_294_967_296,
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-six"},
		{num: 9_223_372_036_854_775_806,
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred six"},
		{num: 9_223_372_036_854_775_807,
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred seven"},
		{num: -9_223_372_036_854_775_807,
			exp: "negative nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred seven"},
		{num: -9_223_372_036_854_775_808,
			exp: "negative nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred eight"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.num), func(t *testing.T) {
			var act string
			testFunc := func() {
				act = ToWords(tc.num)
			}
			require.NotPanics(t, testFunc, "ToWords(%d)", tc.num)
			assert.Equal(t, tc.exp, act, "ToWords(%d)", tc.num)
		})
	}
}

func TestGroupsToWords(t *testing.T) {
	tests := []struct {
		name   string
		groups []uint16
		exp    string
		expErr bool
	}{
		{name: "0", groups: []uint16{0}, exp: "zero"},
		{name: "1", groups: []uint16{1}, exp: "one"},
		{name: "2", groups: []uint16{2}, exp: "two"},
		{name: "3", groups: []uint16{3}, exp: "three"},
		{name: "4", groups: []uint16{4}, exp: "four"},
		{name: "5", groups: []uint16{5}, exp: "five"},
		{name: "6", groups: []uint16{6}, exp: "six"},
		{name: "7", groups: []uint16{7}, exp: "seven"},
		{name: "8", groups: []uint16{8}, exp: "eight"},
		{name: "9", groups: []uint16{9}, exp: "nine"},
		{name: "10", groups: []uint16{10}, exp: "ten"},
		{name: "11", groups: []uint16{11}, exp: "eleven"},
		{name: "12", groups: []uint16{12}, exp: "twelve"},
		{name: "13", groups: []uint16{13}, exp: "thirteen"},
		{name: "14", groups: []uint16{14}, exp: "fourteen"},
		{name: "15", groups: []uint16{15}, exp: "fifteen"},
		{name: "16", groups: []uint16{16}, exp: "sixteen"},
		{name: "17", groups: []uint16{17}, exp: "seventeen"},
		{name: "18", groups: []uint16{18}, exp: "eighteen"},
		{name: "19", groups: []uint16{19}, exp: "nineteen"},
		{name: "20", groups: []uint16{20}, exp: "twenty"},
		{name: "21", groups: []uint16{21}, exp: "twenty-one"},
		{name: "22", groups: []uint16{22}, exp: "twenty-two"},
		{name: "23", groups: []uint16{23}, exp: "twenty-three"},
		{name: "24", groups: []uint16{24}, exp: "twenty-four"},
		{name: "25", groups: []uint16{25}, exp: "twenty-five"},
		{name: "26", groups: []uint16{26}, exp: "twenty-six"},
		{name: "27", groups: []uint16{27}, exp: "twenty-seven"},
		{name: "28", groups: []uint16{28}, exp: "twenty-eight"},
		{name: "29", groups: []uint16{29}, exp: "twenty-nine"},
		{name: "30", groups: []uint16{30}, exp: "thirty"},
		{name: "31", groups: []uint16{31}, exp: "thirty-one"},
		{name: "32", groups: []uint16{32}, exp: "thirty-two"},
		{name: "33", groups: []uint16{33}, exp: "thirty-three"},
		{name: "34", groups: []uint16{34}, exp: "thirty-four"},
		{name: "35", groups: []uint16{35}, exp: "thirty-five"},
		{name: "36", groups: []uint16{36}, exp: "thirty-six"},
		{name: "37", groups: []uint16{37}, exp: "thirty-seven"},
		{name: "38", groups: []uint16{38}, exp: "thirty-eight"},
		{name: "39", groups: []uint16{39}, exp: "thirty-nine"},
		{name: "40", groups: []uint16{40}, exp: "forty"},
		{name: "41", groups: []uint16{41}, exp: "forty-one"},
		{name: "42", groups: []uint16{42}, exp: "forty-two"},
		{name: "43", groups: []uint16{43}, exp: "forty-three"},
		{name: "44", groups: []uint16{44}, exp: "forty-four"},
		{name: "45", groups: []uint16{45}, exp: "forty-five"},
		{name: "46", groups: []uint16{46}, exp: "forty-six"},
		{name: "47", groups: []uint16{47}, exp: "forty-seven"},
		{name: "48", groups: []uint16{48}, exp: "forty-eight"},
		{name: "49", groups: []uint16{49}, exp: "forty-nine"},
		{name: "50", groups: []uint16{50}, exp: "fifty"},
		{name: "51", groups: []uint16{51}, exp: "fifty-one"},
		{name: "52", groups: []uint16{52}, exp: "fifty-two"},
		{name: "53", groups: []uint16{53}, exp: "fifty-three"},
		{name: "54", groups: []uint16{54}, exp: "fifty-four"},
		{name: "55", groups: []uint16{55}, exp: "fifty-five"},
		{name: "56", groups: []uint16{56}, exp: "fifty-six"},
		{name: "57", groups: []uint16{57}, exp: "fifty-seven"},
		{name: "58", groups: []uint16{58}, exp: "fifty-eight"},
		{name: "59", groups: []uint16{59}, exp: "fifty-nine"},
		{name: "60", groups: []uint16{60}, exp: "sixty"},
		{name: "61", groups: []uint16{61}, exp: "sixty-one"},
		{name: "62", groups: []uint16{62}, exp: "sixty-two"},
		{name: "63", groups: []uint16{63}, exp: "sixty-three"},
		{name: "64", groups: []uint16{64}, exp: "sixty-four"},
		{name: "65", groups: []uint16{65}, exp: "sixty-five"},
		{name: "66", groups: []uint16{66}, exp: "sixty-six"},
		{name: "67", groups: []uint16{67}, exp: "sixty-seven"},
		{name: "68", groups: []uint16{68}, exp: "sixty-eight"},
		{name: "69", groups: []uint16{69}, exp: "sixty-nine"},
		{name: "70", groups: []uint16{70}, exp: "seventy"},
		{name: "71", groups: []uint16{71}, exp: "seventy-one"},
		{name: "72", groups: []uint16{72}, exp: "seventy-two"},
		{name: "73", groups: []uint16{73}, exp: "seventy-three"},
		{name: "74", groups: []uint16{74}, exp: "seventy-four"},
		{name: "75", groups: []uint16{75}, exp: "seventy-five"},
		{name: "76", groups: []uint16{76}, exp: "seventy-six"},
		{name: "77", groups: []uint16{77}, exp: "seventy-seven"},
		{name: "78", groups: []uint16{78}, exp: "seventy-eight"},
		{name: "79", groups: []uint16{79}, exp: "seventy-nine"},
		{name: "80", groups: []uint16{80}, exp: "eighty"},
		{name: "81", groups: []uint16{81}, exp: "eighty-one"},
		{name: "82", groups: []uint16{82}, exp: "eighty-two"},
		{name: "83", groups: []uint16{83}, exp: "eighty-three"},
		{name: "84", groups: []uint16{84}, exp: "eighty-four"},
		{name: "85", groups: []uint16{85}, exp: "eighty-five"},
		{name: "86", groups: []uint16{86}, exp: "eighty-six"},
		{name: "87", groups: []uint16{87}, exp: "eighty-seven"},
		{name: "88", groups: []uint16{88}, exp: "eighty-eight"},
		{name: "89", groups: []uint16{89}, exp: "eighty-nine"},
		{name: "90", groups: []uint16{90}, exp: "ninety"},
		{name: "91", groups: []uint16{91}, exp: "ninety-one"},
		{name: "92", groups: []uint16{92}, exp: "ninety-two"},
		{name: "93", groups: []uint16{93}, exp: "ninety-three"},
		{name: "94", groups: []uint16{94}, exp: "ninety-four"},
		{name: "95", groups: []uint16{95}, exp: "ninety-five"},
		{name: "96", groups: []uint16{96}, exp: "ninety-six"},
		{name: "97", groups: []uint16{97}, exp: "ninety-seven"},
		{name: "98", groups: []uint16{98}, exp: "ninety-eight"},
		{name: "99", groups: []uint16{99}, exp: "ninety-nine"},
		{name: "100", groups: []uint16{100}, exp: "one hundred"},
		{name: "111", groups: []uint16{111}, exp: "one hundred eleven"},
		{name: "54,321", groups: []uint16{54, 321}, exp: "fifty-four thousand three hundred twenty-one"},
		{name: "1,234", groups: []uint16{1, 234}, exp: "one thousand two hundred thirty-four"},
		{name: "1,000", groups: []uint16{1, 0}, exp: "one thousand"},
		{name: "1,001", groups: []uint16{1, 1}, exp: "one thousand one"},
		{name: "1,020", groups: []uint16{1, 20}, exp: "one thousand twenty"},
		{name: "1,300", groups: []uint16{1, 300}, exp: "one thousand three hundred"},
		{name: "1,045", groups: []uint16{1, 45}, exp: "one thousand forty-five"},
		{name: "1,670", groups: []uint16{1, 670}, exp: "one thousand six hundred seventy"},
		{name: "1,809", groups: []uint16{1, 809}, exp: "one thousand eight hundred nine"},
		{name: "2,000", groups: []uint16{2, 0}, exp: "two thousand"},
		{name: "2,001", groups: []uint16{2, 1}, exp: "two thousand one"},
		{name: "2,020", groups: []uint16{2, 20}, exp: "two thousand twenty"},
		{name: "2,300", groups: []uint16{2, 300}, exp: "two thousand three hundred"},
		{name: "2,045", groups: []uint16{2, 45}, exp: "two thousand forty-five"},
		{name: "2,670", groups: []uint16{2, 670}, exp: "two thousand six hundred seventy"},
		{name: "2,809", groups: []uint16{2, 809}, exp: "two thousand eight hundred nine"},
		{name: "3,000", groups: []uint16{3, 0}, exp: "three thousand"},
		{name: "3,001", groups: []uint16{3, 1}, exp: "three thousand one"},
		{name: "3,020", groups: []uint16{3, 20}, exp: "three thousand twenty"},
		{name: "3,300", groups: []uint16{3, 300}, exp: "three thousand three hundred"},
		{name: "3,045", groups: []uint16{3, 45}, exp: "three thousand forty-five"},
		{name: "3,670", groups: []uint16{3, 670}, exp: "three thousand six hundred seventy"},
		{name: "3,809", groups: []uint16{3, 809}, exp: "three thousand eight hundred nine"},
		{name: "4,000", groups: []uint16{4, 0}, exp: "four thousand"},
		{name: "4,001", groups: []uint16{4, 1}, exp: "four thousand one"},
		{name: "4,020", groups: []uint16{4, 20}, exp: "four thousand twenty"},
		{name: "4,300", groups: []uint16{4, 300}, exp: "four thousand three hundred"},
		{name: "4,045", groups: []uint16{4, 45}, exp: "four thousand forty-five"},
		{name: "4,670", groups: []uint16{4, 670}, exp: "four thousand six hundred seventy"},
		{name: "4,809", groups: []uint16{4, 809}, exp: "four thousand eight hundred nine"},
		{name: "5,000", groups: []uint16{5, 0}, exp: "five thousand"},
		{name: "5,001", groups: []uint16{5, 1}, exp: "five thousand one"},
		{name: "5,020", groups: []uint16{5, 20}, exp: "five thousand twenty"},
		{name: "5,300", groups: []uint16{5, 300}, exp: "five thousand three hundred"},
		{name: "5,045", groups: []uint16{5, 45}, exp: "five thousand forty-five"},
		{name: "5,670", groups: []uint16{5, 670}, exp: "five thousand six hundred seventy"},
		{name: "5,809", groups: []uint16{5, 809}, exp: "five thousand eight hundred nine"},
		{name: "6,000", groups: []uint16{6, 0}, exp: "six thousand"},
		{name: "6,001", groups: []uint16{6, 1}, exp: "six thousand one"},
		{name: "6,020", groups: []uint16{6, 20}, exp: "six thousand twenty"},
		{name: "6,300", groups: []uint16{6, 300}, exp: "six thousand three hundred"},
		{name: "6,045", groups: []uint16{6, 45}, exp: "six thousand forty-five"},
		{name: "6,670", groups: []uint16{6, 670}, exp: "six thousand six hundred seventy"},
		{name: "6,809", groups: []uint16{6, 809}, exp: "six thousand eight hundred nine"},
		{name: "7,000", groups: []uint16{7, 0}, exp: "seven thousand"},
		{name: "7,001", groups: []uint16{7, 1}, exp: "seven thousand one"},
		{name: "7,020", groups: []uint16{7, 20}, exp: "seven thousand twenty"},
		{name: "7,300", groups: []uint16{7, 300}, exp: "seven thousand three hundred"},
		{name: "7,045", groups: []uint16{7, 45}, exp: "seven thousand forty-five"},
		{name: "7,670", groups: []uint16{7, 670}, exp: "seven thousand six hundred seventy"},
		{name: "7,809", groups: []uint16{7, 809}, exp: "seven thousand eight hundred nine"},
		{name: "8,000", groups: []uint16{8, 0}, exp: "eight thousand"},
		{name: "8,001", groups: []uint16{8, 1}, exp: "eight thousand one"},
		{name: "8,020", groups: []uint16{8, 20}, exp: "eight thousand twenty"},
		{name: "8,300", groups: []uint16{8, 300}, exp: "eight thousand three hundred"},
		{name: "8,045", groups: []uint16{8, 45}, exp: "eight thousand forty-five"},
		{name: "8,670", groups: []uint16{8, 670}, exp: "eight thousand six hundred seventy"},
		{name: "8,809", groups: []uint16{8, 809}, exp: "eight thousand eight hundred nine"},
		{name: "9,000", groups: []uint16{9, 0}, exp: "nine thousand"},
		{name: "9,001", groups: []uint16{9, 1}, exp: "nine thousand one"},
		{name: "9,020", groups: []uint16{9, 20}, exp: "nine thousand twenty"},
		{name: "9,300", groups: []uint16{9, 300}, exp: "nine thousand three hundred"},
		{name: "9,045", groups: []uint16{9, 45}, exp: "nine thousand forty-five"},
		{name: "9,670", groups: []uint16{9, 670}, exp: "nine thousand six hundred seventy"},
		{name: "9,809", groups: []uint16{9, 809}, exp: "nine thousand eight hundred nine"},
		{name: "9,999", groups: []uint16{9, 999}, exp: "nine thousand nine hundred ninety-nine"},
		{name: "10,000", groups: []uint16{10, 0}, exp: "ten thousand"},
		{name: "24,745", groups: []uint16{24, 745}, exp: "twenty-four thousand seven hundred forty-five"},
		{name: "99,999", groups: []uint16{99, 999}, exp: "ninety-nine thousand nine hundred ninety-nine"},
		{name: "100,000", groups: []uint16{100, 0}, exp: "one hundred thousand"},
		{name: "552,887", groups: []uint16{552, 887},
			exp: "five hundred fifty-two thousand eight hundred eighty-seven"},
		{name: "1,000,000", groups: []uint16{1, 0, 0}, exp: "one million"},
		{name: "1,002,003", groups: []uint16{1, 2, 3}, exp: "one million two thousand three"},
		{name: "5,485,065", groups: []uint16{5, 485, 65},
			exp: "five million four hundred eighty-five thousand sixty-five"},
		{name: "10,000,000", groups: []uint16{10, 0, 0}, exp: "ten million"},
		{name: "82,212,496", groups: []uint16{82, 212, 496},
			exp: "eighty-two million two hundred twelve thousand four hundred ninety-six"},
		{name: "100,000,000", groups: []uint16{100, 0, 0}, exp: "one hundred million"},
		{name: "100,200,300", groups: []uint16{100, 200, 300},
			exp: "one hundred million two hundred thousand three hundred"},
		{name: "126,490,799", groups: []uint16{126, 490, 799},
			exp: "one hundred twenty-six million four hundred ninety thousand seven hundred ninety-nine"},
		{name: "1,000,000,000", groups: []uint16{1, 0, 0, 0}, exp: "one billion"},
		{name: "9,007,912,442", groups: []uint16{9, 7, 912, 442},
			exp: "nine billion seven million nine hundred twelve thousand four hundred forty-two"},
		{name: "10,000,000,000", groups: []uint16{10, 0, 0, 0}, exp: "ten billion"},
		{name: "10,000,000,030", groups: []uint16{10, 0, 0, 30}, exp: "ten billion thirty"},
		{name: "10,000,030,000", groups: []uint16{10, 0, 30, 0}, exp: "ten billion thirty thousand"},
		{name: "64,127,772,414", groups: []uint16{64, 127, 772, 414},
			exp: "sixty-four billion one hundred twenty-seven million " +
				"seven hundred seventy-two thousand four hundred fourteen"},
		{name: "100,000,000,000", groups: []uint16{100, 0, 0, 0}, exp: "one hundred billion"},
		{name: "759,528,730,112", groups: []uint16{759, 528, 730, 112},
			exp: "seven hundred fifty-nine billion five hundred twenty-eight million " +
				"seven hundred thirty thousand one hundred twelve"},
		{name: "1,000,000,000,000", groups: []uint16{1, 0, 0, 0, 0}, exp: "one trillion"},
		{name: "9,515,965,217,456", groups: []uint16{9, 515, 965, 217, 456},
			exp: "nine trillion five hundred fifteen billion " +
				"nine hundred sixty-five million two hundred seventeen thousand four hundred fifty-six"},
		{name: "10,000,000,000,000", groups: []uint16{10, 0, 0, 0, 0}, exp: "ten trillion"},
		{name: "50,558,442,088,500", groups: []uint16{50, 558, 442, 88, 500},
			exp: "fifty trillion five hundred fifty-eight billion " +
				"four hundred forty-two million eighty-eight thousand five hundred"},
		{name: "100,000,000,000,000", groups: []uint16{100, 0, 0, 0, 0}, exp: "one hundred trillion"},
		{name: "875,545,170,963,847", groups: []uint16{875, 545, 170, 963, 847},
			exp: "eight hundred seventy-five trillion five hundred forty-five billion " +
				"one hundred seventy million nine hundred sixty-three thousand eight hundred forty-seven"},
		{name: "1,000,000,000,000,000", groups: []uint16{1, 0, 0, 0, 0, 0}, exp: "one quadrillion"},
		{name: "1,459,010,276,579,858", groups: []uint16{1, 459, 10, 276, 579, 858},
			exp: "one quadrillion four hundred fifty-nine trillion " +
				"ten billion two hundred seventy-six million " +
				"five hundred seventy-nine thousand eight hundred fifty-eight"},
		{name: "10,000,000,000,000,000", groups: []uint16{10, 0, 0, 0, 0, 0}, exp: "ten quadrillion"},
		{name: "63,817,328,483,963,713", groups: []uint16{63, 817, 328, 483, 963, 713},
			exp: "sixty-three quadrillion eight hundred seventeen trillion " +
				"three hundred twenty-eight billion four hundred eighty-three million " +
				"nine hundred sixty-three thousand seven hundred thirteen"},
		{name: "100,000,000,000,000,000", groups: []uint16{100, 0, 0, 0, 0, 0}, exp: "one hundred quadrillion"},
		{name: "503,030,044,673,410,914", groups: []uint16{503, 30, 44, 673, 410, 914},
			exp: "five hundred three quadrillion thirty trillion " +
				"forty-four billion six hundred seventy-three million " +
				"four hundred ten thousand nine hundred fourteen"},
		{name: "1,000,000,000,000,000,000", groups: []uint16{1, 0, 0, 0, 0, 0, 0},
			exp: "one quintillion"},
		{name: "1,002,003,004,005,006,007", groups: []uint16{1, 2, 3, 4, 5, 6, 7},
			exp: "one quintillion two quadrillion three trillion four billion five million six thousand seven"},
		{name: "6,095,934,577,086,450,739", groups: []uint16{6, 95, 934, 577, 86, 450, 739},
			exp: "six quintillion ninety-five quadrillion " +
				"nine hundred thirty-four trillion five hundred seventy-seven billion " +
				"eighty-six million four hundred fifty thousand seven hundred thirty-nine"},
		{name: "126", groups: []uint16{126}, exp: "one hundred twenty-six"},
		{name: "127", groups: []uint16{127}, exp: "one hundred twenty-seven"},
		{name: "128", groups: []uint16{128}, exp: "one hundred twenty-eight"},
		{name: "129", groups: []uint16{129}, exp: "one hundred twenty-nine"},
		{name: "254", groups: []uint16{254}, exp: "two hundred fifty-four"},
		{name: "255", groups: []uint16{255}, exp: "two hundred fifty-five"},
		{name: "256", groups: []uint16{256}, exp: "two hundred fifty-six"},
		{name: "32,766", groups: []uint16{32, 766}, exp: "thirty-two thousand seven hundred sixty-six"},
		{name: "32,767", groups: []uint16{32, 767}, exp: "thirty-two thousand seven hundred sixty-seven"},
		{name: "32,768", groups: []uint16{32, 768}, exp: "thirty-two thousand seven hundred sixty-eight"},
		{name: "32,769", groups: []uint16{32, 769}, exp: "thirty-two thousand seven hundred sixty-nine"},
		{name: "65,534", groups: []uint16{65, 534}, exp: "sixty-five thousand five hundred thirty-four"},
		{name: "65,535", groups: []uint16{65, 535}, exp: "sixty-five thousand five hundred thirty-five"},
		{name: "65,536", groups: []uint16{65, 536}, exp: "sixty-five thousand five hundred thirty-six"},
		{name: "2,147,483,646", groups: []uint16{2, 147, 483, 646},
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-six"},
		{name: "2,147,483,647", groups: []uint16{2, 147, 483, 647},
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-seven"},
		{name: "2,147,483,648", groups: []uint16{2, 147, 483, 648},
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-eight"},
		{name: "2,147,483,649", groups: []uint16{2, 147, 483, 649},
			exp: "two billion one hundred forty-seven million " +
				"four hundred eighty-three thousand six hundred forty-nine"},
		{name: "4,294,967,294", groups: []uint16{4, 294, 967, 294},
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-four"},
		{name: "4,294,967,295", groups: []uint16{4, 294, 967, 295},
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-five"},
		{name: "4,294,967,296", groups: []uint16{4, 294, 967, 296},
			exp: "four billion two hundred ninety-four million " +
				"nine hundred sixty-seven thousand two hundred ninety-six"},
		{name: "9,223,372,036,854,775,806", groups: []uint16{9, 223, 372, 36, 854, 775, 806},
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred six"},
		{name: "9,223,372,036,854,775,807", groups: []uint16{9, 223, 372, 36, 854, 775, 807},
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred seven"},
		{name: "9,223,372,036,854,775,808", groups: []uint16{9, 223, 372, 36, 854, 775, 808},
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred eight"},
		{
			name:   "123,234,345,456,567,678,789,890,901,012,135,246,357,468,579,680",
			groups: []uint16{123, 234, 345, 456, 567, 678, 789, 890, 901, 12, 135, 246, 357, 468, 579, 680},
			exp: "one hundred twenty-three quattuordecillion two hundred thirty-four tredecillion " +
				"three hundred forty-five duodecillion four hundred fifty-six undecillion " +
				"five hundred sixty-seven decillion six hundred seventy-eight nonillion " +
				"seven hundred eighty-nine octillion eight hundred ninety septillion " +
				"nine hundred one sextillion twelve quintillion " +
				"one hundred thirty-five quadrillion two hundred forty-six trillion " +
				"three hundred fifty-seven billion four hundred sixty-eight million " +
				"five hundred seventy-nine thousand six hundred eighty",
		},
		{
			name:   "nil groups",
			groups: nil,
			expErr: true,
		},
		{
			name:   "empty groups",
			groups: []uint16{},
			expErr: true,
		},
		{
			name:   "too many groups",
			groups: []uint16{123, 234, 345, 456, 567, 678, 789, 890, 901, 12, 135, 246, 357, 468, 579, 680, 791},
			expErr: true,
		},
	}

	for _, tc := range tests {
		var expErr string
		if tc.expErr {
			expErr = fmt.Sprintf("cannot get quantifiers for %d groups: must be between 1 and 16", len(tc.groups))
		}

		t.Run("normal: "+tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = GroupsToWords(tc.groups)
			}
			require.NotPanics(t, testFunc, "GroupsToWords(%d)", tc.groups)
			if len(expErr) > 0 {
				assert.EqualError(t, err, expErr, "GroupsToWords(%d) error", tc.groups)
			} else {
				assert.NoError(t, err, "GroupsToWords(%d) error", tc.groups)
			}
			assert.Equal(t, tc.exp, act, "GroupsToWords(%d) result", tc.groups)
		})

		t.Run("must: "+tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MustGroupsToWords(tc.groups)
			}
			if len(expErr) > 0 {
				require.PanicsWithError(t, expErr, testFunc, "MustGroupsToWords(%d)", tc.groups)
			} else {
				require.NotPanics(t, testFunc, "MustGroupsToWords(%d)", tc.groups)
			}
			assert.Equal(t, tc.exp, act, "MustGroupsToWords(%d)", tc.groups)
		})
	}
}

func TestToGroups(t *testing.T) {
	tests := []struct {
		name string
		num  int
		exp  []uint16
	}{
		{name: "zero", num: 0, exp: []uint16{0}},
		{name: "one digit", num: 1, exp: []uint16{1}},
		{name: "neg: one digit", num: -1, exp: []uint16{1}},
		{name: "two digits", num: 43, exp: []uint16{43}},
		{name: "neg: two digits", num: -43, exp: []uint16{43}},
		{name: "three digits", num: 921, exp: []uint16{921}},
		{name: "neg: three digits", num: -921, exp: []uint16{921}},
		{name: "four digits: last three are zero", num: 4_000, exp: []uint16{4, 0}},
		{name: "neg: four digits: last three are zero", num: -4_000, exp: []uint16{4, 0}},
		{name: "four digits: zero hundreds and tens", num: 6_005, exp: []uint16{6, 5}},
		{name: "neg: four digits: zero hundreds and tens", num: -6_005, exp: []uint16{6, 5}},
		{name: "four digits: zero hundreds", num: 8_012, exp: []uint16{8, 12}},
		{name: "neg: four digits: zero hundreds", num: -8_012, exp: []uint16{8, 12}},
		{name: "four digits: none zero", num: 4_118, exp: []uint16{4, 118}},
		{name: "neg: four digits: none zero", num: -4_118, exp: []uint16{4, 118}},
		{name: "five digits", num: 54_123, exp: []uint16{54, 123}},
		{name: "neg: five digits", num: -54_123, exp: []uint16{54, 123}},
		{name: "six digits", num: 100_000, exp: []uint16{100, 0}},
		{name: "neg: six digits", num: -100_000, exp: []uint16{100, 0}},
		{name: "seven digits", num: 4_717_010, exp: []uint16{4, 717, 10}},
		{name: "neg: seven digits", num: -4_717_010, exp: []uint16{4, 717, 10}},
		{name: "eight digits", num: 12_345_678, exp: []uint16{12, 345, 678}},
		{name: "neg: eight digits", num: -12_345_678, exp: []uint16{12, 345, 678}},
		{name: "nine digits", num: 987_654_321, exp: []uint16{987, 654, 321}},
		{name: "neg: nine digits", num: -987_654_321, exp: []uint16{987, 654, 321}},
		{name: "ten digits", num: 1_366_150_224, exp: []uint16{1, 366, 150, 224}},
		{name: "neg: ten digits", num: -1_366_150_224, exp: []uint16{1, 366, 150, 224}},
		{name: "eleven digits", num: 55_292_409_676, exp: []uint16{55, 292, 409, 676}},
		{name: "neg: eleven digits", num: -55_292_409_676, exp: []uint16{55, 292, 409, 676}},
		{name: "twelve digits", num: 482_992_041_424, exp: []uint16{482, 992, 41, 424}},
		{name: "neg: twelve digits", num: -482_992_041_424, exp: []uint16{482, 992, 41, 424}},
		{name: "thirteen digits", num: 6_099_094_908_519, exp: []uint16{6, 99, 94, 908, 519}},
		{name: "neg: thirteen digits", num: -6_099_094_908_519, exp: []uint16{6, 99, 94, 908, 519}},
		{name: "fourteen digits", num: 62_276_354_917_434, exp: []uint16{62, 276, 354, 917, 434}},
		{name: "neg: fourteen digits", num: -62_276_354_917_434, exp: []uint16{62, 276, 354, 917, 434}},
		{name: "fifteen digits", num: 647_480_380_208_808, exp: []uint16{647, 480, 380, 208, 808}},
		{name: "neg: fifteen digits", num: -647_480_380_208_808, exp: []uint16{647, 480, 380, 208, 808}},
		{name: "sixteen digits", num: 6_743_766_849_744_459, exp: []uint16{6, 743, 766, 849, 744, 459}},
		{name: "neg: sixteen digits", num: -6_743_766_849_744_459, exp: []uint16{6, 743, 766, 849, 744, 459}},
		{name: "seventeen digits", num: 14_714_454_048_183_145, exp: []uint16{14, 714, 454, 48, 183, 145}},
		{name: "neg: seventeen digits", num: -14_714_454_048_183_145, exp: []uint16{14, 714, 454, 48, 183, 145}},
		{name: "eighteen digits", num: 836_535_708_029_426_971, exp: []uint16{836, 535, 708, 29, 426, 971}},
		{name: "neg: eighteen digits", num: -836_535_708_029_426_971, exp: []uint16{836, 535, 708, 29, 426, 971}},
		{name: "max int", num: 9_223_372_036_854_775_807, exp: []uint16{9, 223, 372, 36, 854, 775, 807}},
		{name: "min int", num: -9_223_372_036_854_775_808, exp: []uint16{9, 223, 372, 36, 854, 775, 808}},
	}

	toStrs := func(nums []uint16) []string {
		if nums == nil {
			return nil
		}
		rv := make([]string, len(nums))
		for i, num := range nums {
			rv[i] = strconv.Itoa(int(num))
		}
		return rv
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act []uint16
			testFunc := func() {
				act = ToGroups(tc.num)
			}
			require.NotPanics(t, testFunc, "ToGroups(%d)", tc.num)

			expStrs := toStrs(tc.exp)
			actStrs := toStrs(act)
			// The assertions output unsigned numbers as hex, so first do the comparison as strings so that the failure
			// messages will show the numbers as defined (and be easier to troubleshoot).
			if assert.Equal(t, expStrs, actStrs, "ToGroups(%d) (as strings)", tc.num) {
				// And just to make sure the string conversion isn't masking something.
				assert.Equal(t, tc.exp, act, "ToGroups(%d) (as numbers)", tc.num)
			}
		})
	}
}

func TestGetQuantifiers(t *testing.T) {
	tests := []struct {
		name       string
		groupCount int
		exp        []string
		expErr     bool
	}{
		{name: "negative one", groupCount: -1, expErr: true},
		{name: "zero", groupCount: 0, expErr: true},
		{name: "one", groupCount: 1, exp: []string{""}},
		{name: "two", groupCount: 2, exp: []string{"thousand", ""}},
		{name: "three", groupCount: 3, exp: []string{"million", "thousand", ""}},
		{name: "four", groupCount: 4, exp: []string{"billion", "million", "thousand", ""}},
		{name: "five", groupCount: 5, exp: []string{"trillion", "billion", "million", "thousand", ""}},
		{name: "six", groupCount: 6, exp: []string{"quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "seven", groupCount: 7,
			exp: []string{"quintillion", "quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "eight", groupCount: 8,
			exp: []string{"sextillion", "quintillion", "quadrillion",
				"trillion", "billion", "million", "thousand", ""}},
		{name: "nine", groupCount: 9,
			exp: []string{"septillion", "sextillion", "quintillion", "quadrillion",
				"trillion", "billion", "million", "thousand", ""}},
		{name: "ten", groupCount: 10,
			exp: []string{"octillion", "septillion", "sextillion", "quintillion",
				"quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "eleven", groupCount: 11,
			exp: []string{"nonillion", "octillion", "septillion", "sextillion", "quintillion",
				"quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "twelve", groupCount: 12,
			exp: []string{"decillion", "nonillion", "octillion", "septillion", "sextillion",
				"quintillion", "quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "thirteen", groupCount: 13,
			exp: []string{"undecillion", "decillion", "nonillion", "octillion", "septillion", "sextillion",
				"quintillion", "quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "fourteen", groupCount: 14,
			exp: []string{"duodecillion", "undecillion", "decillion", "nonillion", "octillion", "septillion",
				"sextillion", "quintillion", "quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "fifteen", groupCount: 15,
			exp: []string{"tredecillion", "duodecillion", "undecillion", "decillion",
				"nonillion", "octillion", "septillion", "sextillion", "quintillion",
				"quadrillion", "trillion", "billion", "million", "thousand", ""}},
		{name: "sixteen", groupCount: 16,
			exp: []string{"quattuordecillion", "tredecillion", "duodecillion", "undecillion", "decillion",
				"nonillion", "octillion", "septillion", "sextillion", "quintillion", "quadrillion",
				"trillion", "billion", "million", "thousand", ""}},
		{name: "seventeen", groupCount: 17, expErr: true},
	}

	for _, tc := range tests {
		var expErr string
		if tc.expErr {
			expErr = fmt.Sprintf("cannot get quantifiers for %d groups: must be between 1 and 16", tc.groupCount)
		}

		t.Run("normal: "+tc.name, func(t *testing.T) {
			var act []string
			var err error
			testFunc := func() {
				act, err = GetQuantifiers(tc.groupCount)
			}
			require.NotPanics(t, testFunc, "GetQuantifiers(%d)", tc.groupCount)
			if len(expErr) > 0 {
				assert.EqualError(t, err, expErr, "GetQuantifiers(%d) error", tc.groupCount)
			} else {
				assert.NoError(t, err, "GetQuantifiers(%d) error", tc.groupCount)
			}
			assert.Equal(t, tc.exp, act, "GetQuantifiers(%d) result", tc.groupCount)
		})

		t.Run("must: "+tc.name, func(t *testing.T) {
			var act []string
			testFunc := func() {
				act = MustGetQuantifiers(tc.groupCount)
			}
			if len(expErr) > 0 {
				require.PanicsWithError(t, expErr, testFunc, "MustGetQuantifiers(%d)", tc.groupCount)
			} else {
				require.NotPanics(t, testFunc, "MustGetQuantifiers(%d)", tc.groupCount)
			}
			assert.Equal(t, tc.exp, act, "MustGetQuantifiers(%d)", tc.groupCount)
		})
	}
}

func TestGetQuantifier(t *testing.T) {
	tests := []struct {
		name     string
		groupID  int
		exp      string
		expPanic bool
	}{
		{name: "negative one", groupID: -1, expPanic: true},
		{name: "zero", groupID: 0, exp: ""},
		{name: "one", groupID: 1, exp: "thousand"},
		{name: "two", groupID: 2, exp: "million"},
		{name: "three", groupID: 3, exp: "billion"},
		{name: "four", groupID: 4, exp: "trillion"},
		{name: "five", groupID: 5, exp: "quadrillion"},
		{name: "six", groupID: 6, exp: "quintillion"},
		{name: "seven", groupID: 7, exp: "sextillion"},
		{name: "eight", groupID: 8, exp: "septillion"},
		{name: "nine", groupID: 9, exp: "octillion"},
		{name: "ten", groupID: 10, exp: "nonillion"},
		{name: "eleven", groupID: 11, exp: "decillion"},
		{name: "twelve", groupID: 12, exp: "undecillion"},
		{name: "thirteen", groupID: 13, exp: "duodecillion"},
		{name: "fourteen", groupID: 14, exp: "tredecillion"},
		{name: "fifteen", groupID: 15, exp: "quattuordecillion"},
		{name: "sixteen", groupID: 16, expPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MustGetQuantifier(tc.groupID)
			}
			if tc.expPanic {
				expPanic := fmt.Sprintf("cannot get quantifiers for group %d: must be between 0 and 15", tc.groupID)
				require.PanicsWithError(t, expPanic, testFunc, "MustGetQuantifier(%d)", tc.groupID)
			} else {
				require.NotPanics(t, testFunc, "MustGetQuantifier(%d)", tc.groupID)
			}
			assert.Equal(t, tc.exp, act, "MustGetQuantifier(%d)", tc.groupID)
		})
	}
}
