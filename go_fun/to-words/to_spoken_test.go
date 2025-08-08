package to_words

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntToSpoken(t *testing.T) {
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
				act = IntToSpoken(tc.num)
			}
			require.NotPanics(t, testFunc, "IntToSpoken(%d)", tc.num)
			assert.Equal(t, tc.exp, act, "IntToSpoken(%d)", tc.num)
		})
	}
}

func TestUintToSpoken(t *testing.T) {
	tests := []struct {
		num uint
		exp string
	}{
		{num: 0, exp: "zero"},
		{num: 1, exp: "one"},
		{num: 2, exp: "two"},
		{num: 3, exp: "three"},
		{num: 4, exp: "four"},
		{num: 5, exp: "five"},
		{num: 6, exp: "six"},
		{num: 7, exp: "seven"},
		{num: 8, exp: "eight"},
		{num: 9, exp: "nine"},
		{num: 10, exp: "ten"},
		{num: 11, exp: "eleven"},
		{num: 12, exp: "twelve"},
		{num: 13, exp: "thirteen"},
		{num: 14, exp: "fourteen"},
		{num: 15, exp: "fifteen"},
		{num: 16, exp: "sixteen"},
		{num: 17, exp: "seventeen"},
		{num: 18, exp: "eighteen"},
		{num: 19, exp: "nineteen"},
		{num: 20, exp: "twenty"},
		{num: 21, exp: "twenty-one"},
		{num: 22, exp: "twenty-two"},
		{num: 23, exp: "twenty-three"},
		{num: 24, exp: "twenty-four"},
		{num: 25, exp: "twenty-five"},
		{num: 26, exp: "twenty-six"},
		{num: 27, exp: "twenty-seven"},
		{num: 28, exp: "twenty-eight"},
		{num: 29, exp: "twenty-nine"},
		{num: 30, exp: "thirty"},
		{num: 31, exp: "thirty-one"},
		{num: 32, exp: "thirty-two"},
		{num: 33, exp: "thirty-three"},
		{num: 34, exp: "thirty-four"},
		{num: 35, exp: "thirty-five"},
		{num: 36, exp: "thirty-six"},
		{num: 37, exp: "thirty-seven"},
		{num: 38, exp: "thirty-eight"},
		{num: 39, exp: "thirty-nine"},
		{num: 40, exp: "forty"},
		{num: 41, exp: "forty-one"},
		{num: 42, exp: "forty-two"},
		{num: 43, exp: "forty-three"},
		{num: 44, exp: "forty-four"},
		{num: 45, exp: "forty-five"},
		{num: 46, exp: "forty-six"},
		{num: 47, exp: "forty-seven"},
		{num: 48, exp: "forty-eight"},
		{num: 49, exp: "forty-nine"},
		{num: 50, exp: "fifty"},
		{num: 51, exp: "fifty-one"},
		{num: 52, exp: "fifty-two"},
		{num: 53, exp: "fifty-three"},
		{num: 54, exp: "fifty-four"},
		{num: 55, exp: "fifty-five"},
		{num: 56, exp: "fifty-six"},
		{num: 57, exp: "fifty-seven"},
		{num: 58, exp: "fifty-eight"},
		{num: 59, exp: "fifty-nine"},
		{num: 60, exp: "sixty"},
		{num: 61, exp: "sixty-one"},
		{num: 62, exp: "sixty-two"},
		{num: 63, exp: "sixty-three"},
		{num: 64, exp: "sixty-four"},
		{num: 65, exp: "sixty-five"},
		{num: 66, exp: "sixty-six"},
		{num: 67, exp: "sixty-seven"},
		{num: 68, exp: "sixty-eight"},
		{num: 69, exp: "sixty-nine"},
		{num: 70, exp: "seventy"},
		{num: 71, exp: "seventy-one"},
		{num: 72, exp: "seventy-two"},
		{num: 73, exp: "seventy-three"},
		{num: 74, exp: "seventy-four"},
		{num: 75, exp: "seventy-five"},
		{num: 76, exp: "seventy-six"},
		{num: 77, exp: "seventy-seven"},
		{num: 78, exp: "seventy-eight"},
		{num: 79, exp: "seventy-nine"},
		{num: 80, exp: "eighty"},
		{num: 81, exp: "eighty-one"},
		{num: 82, exp: "eighty-two"},
		{num: 83, exp: "eighty-three"},
		{num: 84, exp: "eighty-four"},
		{num: 85, exp: "eighty-five"},
		{num: 86, exp: "eighty-six"},
		{num: 87, exp: "eighty-seven"},
		{num: 88, exp: "eighty-eight"},
		{num: 89, exp: "eighty-nine"},
		{num: 90, exp: "ninety"},
		{num: 91, exp: "ninety-one"},
		{num: 92, exp: "ninety-two"},
		{num: 93, exp: "ninety-three"},
		{num: 94, exp: "ninety-four"},
		{num: 95, exp: "ninety-five"},
		{num: 96, exp: "ninety-six"},
		{num: 97, exp: "ninety-seven"},
		{num: 98, exp: "ninety-eight"},
		{num: 99, exp: "ninety-nine"},
		{num: 100, exp: "one hundred"},
		{num: 111, exp: "one hundred eleven"},
		{num: 54_321, exp: "fifty-four thousand three hundred twenty-one"},
		{num: 1_234, exp: "one thousand two hundred thirty-four"},
		{num: 1_000, exp: "one thousand"},
		{num: 1_001, exp: "one thousand one"},
		{num: 1_020, exp: "one thousand twenty"},
		{num: 1_300, exp: "one thousand three hundred"},
		{num: 1_045, exp: "one thousand forty-five"},
		{num: 1_670, exp: "one thousand six hundred seventy"},
		{num: 1_809, exp: "one thousand eight hundred nine"},
		{num: 2_000, exp: "two thousand"},
		{num: 2_001, exp: "two thousand one"},
		{num: 2_020, exp: "two thousand twenty"},
		{num: 2_300, exp: "two thousand three hundred"},
		{num: 2_045, exp: "two thousand forty-five"},
		{num: 2_670, exp: "two thousand six hundred seventy"},
		{num: 2_809, exp: "two thousand eight hundred nine"},
		{num: 3_000, exp: "three thousand"},
		{num: 3_001, exp: "three thousand one"},
		{num: 3_020, exp: "three thousand twenty"},
		{num: 3_300, exp: "three thousand three hundred"},
		{num: 3_045, exp: "three thousand forty-five"},
		{num: 3_670, exp: "three thousand six hundred seventy"},
		{num: 3_809, exp: "three thousand eight hundred nine"},
		{num: 4_000, exp: "four thousand"},
		{num: 4_001, exp: "four thousand one"},
		{num: 4_020, exp: "four thousand twenty"},
		{num: 4_300, exp: "four thousand three hundred"},
		{num: 4_045, exp: "four thousand forty-five"},
		{num: 4_670, exp: "four thousand six hundred seventy"},
		{num: 4_809, exp: "four thousand eight hundred nine"},
		{num: 5_000, exp: "five thousand"},
		{num: 5_001, exp: "five thousand one"},
		{num: 5_020, exp: "five thousand twenty"},
		{num: 5_300, exp: "five thousand three hundred"},
		{num: 5_045, exp: "five thousand forty-five"},
		{num: 5_670, exp: "five thousand six hundred seventy"},
		{num: 5_809, exp: "five thousand eight hundred nine"},
		{num: 6_000, exp: "six thousand"},
		{num: 6_001, exp: "six thousand one"},
		{num: 6_020, exp: "six thousand twenty"},
		{num: 6_300, exp: "six thousand three hundred"},
		{num: 6_045, exp: "six thousand forty-five"},
		{num: 6_670, exp: "six thousand six hundred seventy"},
		{num: 6_809, exp: "six thousand eight hundred nine"},
		{num: 7_000, exp: "seven thousand"},
		{num: 7_001, exp: "seven thousand one"},
		{num: 7_020, exp: "seven thousand twenty"},
		{num: 7_300, exp: "seven thousand three hundred"},
		{num: 7_045, exp: "seven thousand forty-five"},
		{num: 7_670, exp: "seven thousand six hundred seventy"},
		{num: 7_809, exp: "seven thousand eight hundred nine"},
		{num: 8_000, exp: "eight thousand"},
		{num: 8_001, exp: "eight thousand one"},
		{num: 8_020, exp: "eight thousand twenty"},
		{num: 8_300, exp: "eight thousand three hundred"},
		{num: 8_045, exp: "eight thousand forty-five"},
		{num: 8_670, exp: "eight thousand six hundred seventy"},
		{num: 8_809, exp: "eight thousand eight hundred nine"},
		{num: 9_000, exp: "nine thousand"},
		{num: 9_001, exp: "nine thousand one"},
		{num: 9_020, exp: "nine thousand twenty"},
		{num: 9_300, exp: "nine thousand three hundred"},
		{num: 9_045, exp: "nine thousand forty-five"},
		{num: 9_670, exp: "nine thousand six hundred seventy"},
		{num: 9_809, exp: "nine thousand eight hundred nine"},
		{num: 9_999, exp: "nine thousand nine hundred ninety-nine"},
		{num: 10_000, exp: "ten thousand"},
		{num: 24_745, exp: "twenty-four thousand seven hundred forty-five"},
		{num: 99_999, exp: "ninety-nine thousand nine hundred ninety-nine"},
		{num: 100_000, exp: "one hundred thousand"},
		{num: 552_887, exp: "five hundred fifty-two thousand eight hundred eighty-seven"},
		{num: 1_000_000, exp: "one million"},
		{num: 1_002_003, exp: "one million two thousand three"},
		{num: 5_485_065, exp: "five million four hundred eighty-five thousand sixty-five"},
		{num: 10_000_000, exp: "ten million"},
		{num: 82_212_496, exp: "eighty-two million two hundred twelve thousand four hundred ninety-six"},
		{num: 100_000_000, exp: "one hundred million"},
		{num: 100_200_300, exp: "one hundred million two hundred thousand three hundred"},
		{num: 126_490_799,
			exp: "one hundred twenty-six million four hundred ninety thousand seven hundred ninety-nine"},
		{num: 1_000_000_000, exp: "one billion"},
		{num: 9_007_912_442,
			exp: "nine billion seven million nine hundred twelve thousand four hundred forty-two"},
		{num: 10_000_000_000, exp: "ten billion"},
		{num: 10_000_000_030, exp: "ten billion thirty"},
		{num: 10_000_030_000, exp: "ten billion thirty thousand"},
		{num: 64_127_772_414,
			exp: "sixty-four billion one hundred twenty-seven million " +
				"seven hundred seventy-two thousand four hundred fourteen"},
		{num: 100_000_000_000, exp: "one hundred billion"},
		{num: 759_528_730_112,
			exp: "seven hundred fifty-nine billion five hundred twenty-eight million " +
				"seven hundred thirty thousand one hundred twelve"},
		{num: 1_000_000_000_000, exp: "one trillion"},
		{num: 9_515_965_217_456,
			exp: "nine trillion five hundred fifteen billion " +
				"nine hundred sixty-five million two hundred seventeen thousand four hundred fifty-six"},
		{num: 10_000_000_000_000, exp: "ten trillion"},
		{num: 50_558_442_088_500,
			exp: "fifty trillion five hundred fifty-eight billion " +
				"four hundred forty-two million eighty-eight thousand five hundred"},
		{num: 100_000_000_000_000, exp: "one hundred trillion"},
		{num: 875_545_170_963_847,
			exp: "eight hundred seventy-five trillion five hundred forty-five billion " +
				"one hundred seventy million nine hundred sixty-three thousand eight hundred forty-seven"},
		{num: 1_000_000_000_000_000, exp: "one quadrillion"},
		{num: 1_459_010_276_579_858,
			exp: "one quadrillion four hundred fifty-nine trillion " +
				"ten billion two hundred seventy-six million " +
				"five hundred seventy-nine thousand eight hundred fifty-eight"},
		{num: 10_000_000_000_000_000, exp: "ten quadrillion"},
		{num: 63_817_328_483_963_713,
			exp: "sixty-three quadrillion eight hundred seventeen trillion " +
				"three hundred twenty-eight billion four hundred eighty-three million " +
				"nine hundred sixty-three thousand seven hundred thirteen"},
		{num: 100_000_000_000_000_000, exp: "one hundred quadrillion"},
		{num: 503_030_044_673_410_914,
			exp: "five hundred three quadrillion thirty trillion " +
				"forty-four billion six hundred seventy-three million " +
				"four hundred ten thousand nine hundred fourteen"},
		{num: 1_000_000_000_000_000_000, exp: "one quintillion"},
		{num: 1_002_003_004_005_006_007,
			exp: "one quintillion two quadrillion three trillion " +
				"four billion five million six thousand seven"},
		{num: 6_095_934_577_086_450_739,
			exp: "six quintillion ninety-five quadrillion nine hundred thirty-four trillion " +
				"five hundred seventy-seven billion eighty-six million " +
				"four hundred fifty thousand seven hundred thirty-nine"},
		{num: 126, exp: "one hundred twenty-six"},
		{num: 127, exp: "one hundred twenty-seven"},
		{num: 128, exp: "one hundred twenty-eight"},
		{num: 254, exp: "two hundred fifty-four"},
		{num: 255, exp: "two hundred fifty-five"},
		{num: 256, exp: "two hundred fifty-six"},
		{num: 32_766, exp: "thirty-two thousand seven hundred sixty-six"},
		{num: 32_767, exp: "thirty-two thousand seven hundred sixty-seven"},
		{num: 32_768, exp: "thirty-two thousand seven hundred sixty-eight"},
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
		{num: 9_223_372_036_854_775_808,
			exp: "nine quintillion two hundred twenty-three quadrillion " +
				"three hundred seventy-two trillion thirty-six billion " +
				"eight hundred fifty-four million seven hundred seventy-five thousand eight hundred eight"},
		{num: 18_446_744_073_709_551_614,
			exp: "eighteen quintillion four hundred forty-six quadrillion " +
				"seven hundred forty-four trillion seventy-three billion " +
				"seven hundred nine million five hundred fifty-one thousand six hundred fourteen"},
		{num: 18_446_744_073_709_551_615,
			exp: "eighteen quintillion four hundred forty-six quadrillion " +
				"seven hundred forty-four trillion seventy-three billion " +
				"seven hundred nine million five hundred fifty-one thousand six hundred fifteen"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.num), func(t *testing.T) {
			var act string
			testFunc := func() {
				act = UintToSpoken(tc.num)
			}
			require.NotPanics(t, testFunc, "UintToSpoken(%d)", tc.num)
			assert.Equal(t, tc.exp, act, "UintToSpoken(%d)", tc.num)
		})
	}
}

func TestFloatRx(t *testing.T) {
	tests := []struct {
		str string
		exp []string
	}{
		{str: "", exp: nil},
		{str: "x", exp: nil},
		{str: "-", exp: nil},
		{str: ".", exp: nil},
		{str: "-.", exp: nil},
		{str: ".e", exp: nil},
		{str: "1.2.", exp: nil},
		// "<whole>.<fract>"
		{str: "1.2", exp: []string{"1.2", "1.2", "", "1", "2"}},
		{str: "-1.2", exp: []string{"-1.2", "-1.2", "", "-1", "2"}},
		{str: "1.y", exp: nil},
		{str: "-1.y", exp: nil},
		{str: "0.1713", exp: []string{"0.1713", "0.1713", "", "0", "1713"}},
		{str: "-0.1713", exp: []string{"-0.1713", "-0.1713", "", "-0", "1713"}},
		{str: "9876.543210", exp: []string{"9876.543210", "9876.543210", "", "9876", "543210"}},
		{str: "-9876.543210", exp: []string{"-9876.543210", "-9876.543210", "", "-9876", "543210"}},
		// "<whole>."
		{str: "1.", exp: []string{"1.", "1.", "1", "", ""}},
		{str: "-3.", exp: []string{"-3.", "-3.", "-3", "", ""}},
		{str: "x.", exp: nil},
		{str: "-x.", exp: nil},
		{str: "7531246890.", exp: []string{"7531246890.", "7531246890.", "7531246890", "", ""}},
		{str: "-7531246890.", exp: []string{"-7531246890.", "-7531246890.", "-7531246890", "", ""}},
		// ".<fract>"
		{str: ".5", exp: []string{".5", ".5", "", "", "5"}},
		{str: "-.5", exp: []string{"-.5", "-.5", "", "-", "5"}},
		{str: ".-8", exp: nil},
		{str: ".000111222", exp: []string{".000111222", ".000111222", "", "", "000111222"}},
		{str: "-.000111222", exp: []string{"-.000111222", "-.000111222", "", "-", "000111222"}},
		// "<whole>"
		{str: "1", exp: []string{"1", "1", "1", "", ""}},
		{str: "-3", exp: []string{"-3", "-3", "-3", "", ""}},
		{str: "x", exp: nil},
		{str: "-x", exp: nil},
		{str: "7531246890", exp: []string{"7531246890", "7531246890", "7531246890", "", ""}},
		{str: "-7531246890", exp: []string{"-7531246890", "-7531246890", "-7531246890", "", ""}},
	}

	for _, tc := range tests {
		name := tc.str
		if len(name) == 0 {
			name = "empty string"
		}
		t.Run(name, func(t *testing.T) {
			var matches [][]string
			testFunc := func() {
				matches = floatRx.FindAllStringSubmatch(tc.str, -1)
			}
			require.NotPanics(t, testFunc, "%s.FindAllStringSubmatch(%q)", floatRx, tc.str)
			if len(tc.exp) == 0 {
				assert.Len(t, matches, 0, "%s.FindAllStringSubmatch(%q) result", floatRx, tc.str)
			} else if assert.Len(t, matches, 1, "%s.FindAllStringSubmatch(%q) result", floatRx, tc.str) {
				act := matches[0]
				assert.Equal(t, tc.exp, act, "%s.FindAllStringSubmatch(%q) result", floatRx, tc.str)
			}
		})
	}
}

func TestFloatToSpoken(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		exp    string
		expErr string
	}{
		{
			name:   "empty string",
			str:    "",
			expErr: "not a float \"\"",
		},
		{
			name:   "not a float",
			str:    "not a float",
			expErr: "not a float \"not a float\"",
		},
		{
			name: "invalid whole number",
			// There are 16 quantifiers. 16x3 = 48 digits max, so this one is 49.
			str: "1234567890123456789012345678901234567890123456789",
			expErr: "could not convert \"1234567890123456789012345678901234567890123456789\" to words: " +
				"cannot get quantifiers for 17 groups: must be between 1 and 16",
		},
		{
			name: "invalid whole part",
			// There are 16 quantifiers. 16x3 = 48 digits max, so this one is 49.
			str: "1234567890123456789012345678901234567890123456789.123",
			expErr: "invalid float \"1234567890123456789012345678901234567890123456789.123\": " +
				"invalid whole part: " +
				"could not convert \"1234567890123456789012345678901234567890123456789\" to words: " +
				"cannot get quantifiers for 17 groups: must be between 1 and 16",
		},
		{
			name:   "just a decimal",
			str:    ".",
			expErr: "not a float \".\"",
		},
		{
			name: "whole number only",
			str:  "1234",
			exp:  "one thousand two hundred thirty-four",
		},
		{
			name: "negative whole number only",
			str:  "-5678",
			exp:  "negative five thousand six hundred seventy-eight",
		},
		{
			name: "decimal without fractional part",
			str:  "43.",
			exp:  "forty-three",
		},
		{
			name: "negative decimal without fractional part",
			str:  "-777.",
			exp:  "negative seven hundred seventy-seven",
		},
		{
			name: "decimal without whole part",
			str:  ".123",
			exp:  "point one two three",
		},
		{
			name: "negative decimal without whole part",
			str:  "-.123",
			exp:  "negative point one two three",
		},
	}

	for _, tc := range tests {
		t.Run("normal: "+tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = FloatToSpoken(tc.str)
			}
			require.NotPanics(t, testFunc, "FloatToSpoken(%q)", tc.str)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "FloatToSpoken(%q) error", tc.str)
			} else {
				assert.NoError(t, err, "FloatToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "FloatToSpoken(%q)", tc.str)
		})

		t.Run("must: "+tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MustFloatToSpoken(tc.str)
			}
			if len(tc.expErr) > 0 {
				require.PanicsWithError(t, tc.expErr, testFunc, "MustFloatToSpoken(%q)", tc.str)
			} else {
				require.NotPanics(t, testFunc, "MustFloatToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "MustFloatToSpoken(%q)", tc.str)
		})
	}
}

func TestScientificRx(t *testing.T) {
	tests := []struct {
		str string
		exp []string
	}{
		{str: "", exp: nil},
		{str: "123", exp: nil},
		{str: "123.456", exp: nil},
		{str: ".456", exp: nil},
		{str: "bexex", exp: []string{"bexex", "bex", "e", "", "x"}},
		// "<base>e<exponent>"
		{str: "bex", exp: []string{"bex", "b", "e", "", "x"}},
		{str: "-bex", exp: []string{"-bex", "-b", "e", "", "x"}},
		{str: "be-x", exp: []string{"be-x", "b", "e", "", "-x"}},
		{str: "-be-x", exp: []string{"-be-x", "-b", "e", "", "-x"}},
		{str: "be", exp: nil},
		{str: "ex", exp: nil},
		{str: "e", exp: nil},
		// "<base>E<exponent>"
		{str: "bEx", exp: []string{"bEx", "b", "E", "", "x"}},
		{str: "-bEx", exp: []string{"-bEx", "-b", "E", "", "x"}},
		{str: "bE-x", exp: []string{"bE-x", "b", "E", "", "-x"}},
		{str: "-bE-x", exp: []string{"-bE-x", "-b", "E", "", "-x"}},
		{str: "bE", exp: nil},
		{str: "Ex", exp: nil},
		{str: "E", exp: nil},
		// "<base>x10^<exponent>"
		{str: "bx10^x", exp: []string{"bx10^x", "b", "x10^", "^", "x"}},
		{str: "-bx10^x", exp: []string{"-bx10^x", "-b", "x10^", "^", "x"}},
		{str: "bx10^-x", exp: []string{"bx10^-x", "b", "x10^", "^", "-x"}},
		{str: "-bx10^-x", exp: []string{"-bx10^-x", "-b", "x10^", "^", "-x"}},
		{str: "bx10^", exp: nil},
		{str: "x10^x", exp: nil},
		{str: "x10^", exp: nil},
		// "<base>*10^<exponent>"
		{str: "b*10^x", exp: []string{"b*10^x", "b", "*10^", "^", "x"}},
		{str: "-b*10^x", exp: []string{"-b*10^x", "-b", "*10^", "^", "x"}},
		{str: "b*10^-x", exp: []string{"b*10^-x", "b", "*10^", "^", "-x"}},
		{str: "-b*10^-x", exp: []string{"-b*10^-x", "-b", "*10^", "^", "-x"}},
		{str: "b*10^", exp: nil},
		{str: "*10^x", exp: nil},
		{str: "*10^", exp: nil},
		// "<base>x10**<exponent>"
		{str: "bx10**x", exp: []string{"bx10**x", "b", "x10**", "**", "x"}},
		{str: "-bx10**x", exp: []string{"-bx10**x", "-b", "x10**", "**", "x"}},
		{str: "bx10**-x", exp: []string{"bx10**-x", "b", "x10**", "**", "-x"}},
		{str: "-bx10**-x", exp: []string{"-bx10**-x", "-b", "x10**", "**", "-x"}},
		{str: "bx10**", exp: nil},
		{str: "x10**x", exp: nil},
		{str: "x10**", exp: nil},
		// "<base>*10**<exponent>"
		{str: "b*10**x", exp: []string{"b*10**x", "b", "*10**", "**", "x"}},
		{str: "-b*10**x", exp: []string{"-b*10**x", "-b", "*10**", "**", "x"}},
		{str: "b*10**-x", exp: []string{"b*10**-x", "b", "*10**", "**", "-x"}},
		{str: "-b*10**-x", exp: []string{"-b*10**-x", "-b", "*10**", "**", "-x"}},
		{str: "b*10**", exp: nil},
		{str: "*10**x", exp: nil},
		{str: "*10**", exp: nil},
	}

	for _, tc := range tests {
		name := tc.str
		if len(name) == 0 {
			name = "empty string"
		}
		t.Run(name, func(t *testing.T) {
			var matches [][]string
			testFunc := func() {
				matches = scientificRx.FindAllStringSubmatch(tc.str, -1)
			}
			require.NotPanics(t, testFunc, "%s.FindAllStringSubmatch(%q)", scientificRx, tc.str)
			if len(tc.exp) == 0 {
				assert.Len(t, matches, 0, "%s.FindAllStringSubmatch(%q) result", scientificRx, tc.str)
			} else if assert.Len(t, matches, 1, "%s.FindAllStringSubmatch(%q) result", scientificRx, tc.str) {
				act := matches[0]
				assert.Equal(t, tc.exp, act, "%s.FindAllStringSubmatch(%q) result", scientificRx, tc.str)
			}
		})
	}
}

func TestScientificToSpoken(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		exp    string
		expErr string
	}{
		{
			name:   "empty string",
			str:    "",
			expErr: "not scientific notation \"\"",
		},
		{
			name:   "not scientific notation",
			str:    "no worky",
			expErr: "not scientific notation \"no worky\"",
		},
		{
			name:   "invalid base",
			str:    "ye3",
			expErr: "invalid base from \"ye3\": not a float \"y\"",
		},
		{
			name:   "invalid exponent",
			str:    "3ey",
			expErr: "invalid exponenet from \"3ey\": not a float \"y\"",
		},
		{
			name: "base has only whole part",
			str:  "3e4",
			exp:  "three times ten to the four",
		},
		{
			name: "base has only fractional part",
			str:  ".1e5",
			exp:  "point one times ten to the five",
		},
		{
			name: "base has whole and fractional part",
			str:  "12.66e78",
			exp:  "twelve point six six times ten to the seventy-eight",
		},
		{
			name: "negative base has only whole part",
			str:  "-3e4",
			exp:  "negative three times ten to the four",
		},
		{
			name: "negative base has only fractional part",
			str:  "-.1e5",
			exp:  "negative point one times ten to the five",
		},
		{
			name: "negative base has whole and fractional part",
			str:  "-12.66e78",
			exp:  "negative twelve point six six times ten to the seventy-eight",
		},
		{
			name: "exponent has only whole part",
			str:  "3e14",
			exp:  "three times ten to the fourteen",
		},
		{
			name: "exponent has only fractional part",
			str:  "3e.14",
			exp:  "three times ten to the point one four",
		},
		{
			name: "exponent has whole and fractional part",
			str:  "9e123.456",
			exp:  "nine times ten to the one hundred twenty-three point four five six",
		},
		{
			name: "negative exponent has only whole part",
			str:  "3e-14",
			exp:  "three times ten to the negative fourteen",
		},
		{
			name: "negative exponent has only fractional part",
			str:  "3e-.14",
			exp:  "three times ten to the negative point one four",
		},
		{
			name: "negative exponent has whole and fractional part",
			str:  "9e-123.456",
			exp:  "nine times ten to the negative one hundred twenty-three point four five six",
		},
		{
			name: "using notation: e",
			str:  "1e2",
			exp:  "one times ten to the two",
		},
		{
			name: "using notation: E",
			str:  "3E4",
			exp:  "three times ten to the four",
		},
		{
			name: "using notation: x10^",
			str:  "5x10^6",
			exp:  "five times ten to the six",
		},
		{
			name: "using notation: x10**",
			str:  "7x10**8",
			exp:  "seven times ten to the eight",
		},
		{
			name: "using notation: *10^",
			str:  "9*10^10",
			exp:  "nine times ten to the ten",
		},
		{
			name: "using notation: *10**",
			str:  "11*10**12",
			exp:  "eleven times ten to the twelve",
		},
	}

	for _, tc := range tests {
		t.Run("normal: "+tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = ScientificToSpoken(tc.str)
			}
			require.NotPanics(t, testFunc, "ScientificToSpoken(%q)", tc.str)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ScientificToSpoken(%q) error", tc.str)
			} else {
				assert.NoError(t, err, "ScientificToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "ScientificToSpoken(%q)", tc.str)
		})

		t.Run("must: "+tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MustScientificToSpoken(tc.str)
			}
			if len(tc.expErr) > 0 {
				require.PanicsWithError(t, tc.expErr, testFunc, "MustScientificToSpoken(%q)", tc.str)
			} else {
				require.NotPanics(t, testFunc, "MustScientificToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "MustScientificToSpoken(%q)", tc.str)
		})
	}
}

func TestStringToSpoken(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		exp    string
		expErr string
	}{
		{
			name:   "empty string",
			str:    "",
			expErr: "not a float \"\"",
		},
		{
			name:   "random string",
			str:    "random string",
			expErr: "not a float \"random string\"",
		},
		{
			name: "0",
			str:  "0",
			exp:  "zero",
		},
		{
			name: ".1",
			str:  ".1",
			exp:  "point one",
		},
		{
			name: "-2.3",
			str:  "-2.3",
			exp:  "negative two point three",
		},
		{
			name: "4e5",
			str:  "4e5",
			exp:  "four times ten to the five",
		},
	}

	for _, tc := range tests {
		t.Run("normal: "+tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = StringToSpoken(tc.str)
			}
			require.NotPanics(t, testFunc, "StringToSpoken(%q)", tc.str)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "StringToSpoken(%q) error", tc.str)
			} else {
				assert.NoError(t, err, "StringToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "StringToSpoken(%q)", tc.str)
		})

		t.Run("must: "+tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = MustStringToSpoken(tc.str)
			}
			if len(tc.expErr) > 0 {
				require.PanicsWithError(t, tc.expErr, testFunc, "MustStringToSpoken(%q)", tc.str)
			} else {
				require.NotPanics(t, testFunc, "MustStringToSpoken(%q)", tc.str)
			}
			assert.Equal(t, tc.exp, act, "MustStringToSpoken(%q)", tc.str)
		})
	}
}
