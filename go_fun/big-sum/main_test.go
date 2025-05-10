package main

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertErrorValue(t *testing.T, expected string, theErr error, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(expected) == 0 {
		return assert.NoError(t, theErr, msgAndArgs...)
	}
	return assert.EqualError(t, theErr, expected, msgAndArgs...)
}

func TestSum(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		exp    string
		expErr string
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
			name:   "one arg: no decimal, not an int",
			args:   []string{"123nope4"},
			expErr: "could not parse \"123nope4\" as integer",
		},
		{
			name:   "one arg: with decimal, not a float",
			args:   []string{"87ba.d4"},
			expErr: "could not parse \"87ba.d4\" as float: expected end of string, found 'b'",
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
			name:   "two args: first is invalid int",
			args:   []string{"eleven", "12"},
			expErr: "could not parse \"eleven\" as integer",
		},
		{
			name:   "two args: first is invalid float",
			args:   []string{"eleven.4", "12"},
			expErr: "could not parse \"eleven.4\" as float: number has no digits",
		},
		{
			name:   "two args: second is invalid int",
			args:   []string{"11", "twelve"},
			expErr: "could not parse \"twelve\" as integer",
		},
		{
			name:   "two args: second is invalid float",
			args:   []string{"3.1", "6.twelve"},
			expErr: "could not parse \"6.twelve\" as float: expected end of string, found 't'",
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
			name:   "three args: first is invalid int",
			args:   []string{"seven", "2", "3"},
			expErr: "could not parse \"seven\" as integer",
		},
		{
			name:   "three args: first is invalid float",
			args:   []string{"3.1four1", "2", "3"},
			expErr: "could not parse \"3.1four1\" as float: expected end of string, found 'f'",
		},
		{
			name:   "three args: second is invalid int",
			args:   []string{"1", "seven", "3"},
			expErr: "could not parse \"seven\" as integer",
		},
		{
			name:   "three args: second is invalid float",
			args:   []string{"1", "3.1four1", "3"},
			expErr: "could not parse \"3.1four1\" as float: expected end of string, found 'f'",
		},
		{
			name:   "three args: third is invalid int",
			args:   []string{"1", "2", "seven"},
			expErr: "could not parse \"seven\" as integer",
		},
		{
			name:   "three args: third is invalid float",
			args:   []string{"1", "2", "3.1four1"},
			expErr: "could not parse \"3.1four1\" as float: expected end of string, found 'f'",
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			var err error
			testFunc := func() {
				act, err = Sum(tc.args)
			}
			require.NotPanics(t, testFunc, "Sum(%q)", tc.args)
			assertErrorValue(t, tc.expErr, err, "Sum(%q) error", tc.args)
			assert.Equal(t, tc.exp, act, "Sum(%q) result", tc.args)
		})
	}
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
