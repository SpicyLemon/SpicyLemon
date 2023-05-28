package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func AssertErrorContents(t *testing.T, theErr error, errString string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(errString) > 0 {
		return assert.EqualError(t, theErr, errString, msgAndArgs...)
	}
	return assert.NoError(t, theErr, msgAndArgs...)
}

func TestCmdConfig_Prep(t *testing.T) {
	dummyCmd := &cobra.Command{
		Use: "dummy",
		Run: func(cmd *cobra.Command, args []string) {
			panic("dummy command should never be run")
		},
	}

	tests := []struct {
		name       string
		c          *CmdConfig
		args       []string
		expErr     string
		expFromVal FromVal
		expCount   int
	}{
		{
			name:       "invalid from",
			c:          &CmdConfig{From: "what"},
			args:       []string{"0a"},
			expErr:     `invalid --from value "what", must be one of "detect" "bech32" "base64" "hex"`,
			expFromVal: FromValDetect,
			expCount:   1,
		},
		{
			name:       "valid from",
			c:          &CmdConfig{From: "bech32"},
			args:       []string{"0a", "0b", "0c"},
			expFromVal: FromValBech32,
			expCount:   3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testString := fmt.Sprintf("This is a %q tesst string", tc.name)
			var buffer bytes.Buffer
			dummyCmd.SetOut(&buffer)
			err := tc.c.Prep(dummyCmd, tc.args)
			AssertErrorContents(t, err, tc.expErr, "Prep error")
			assert.Equal(t, tc.expCount, tc.c.Count, "Count")
			assert.Equal(t, tc.expFromVal, tc.c.FromVal, "FromVal")

			// Make sure the writer got set to the buffer we gave to the command.
			_, err = fmt.Fprintf(tc.c.Writer, "%s", testString)
			if assert.NoError(t, err, "Fprintf to the Writer") {
				outStr := buffer.String()
				assert.Equal(t, testString, outStr, "string from the writer")
			}
		})
	}
}

func TestToFromVal(t *testing.T) {
	tests := []struct {
		str    string
		exp    FromVal
		expErr string
	}{
		{str: "", exp: FromValDetect, expErr: `invalid --from value "", must be one of ` + FromValOptionsStr},

		{str: "detect", exp: FromValDetect},
		{str: "DETECT", exp: FromValDetect},
		{str: "DeteCt", exp: FromValDetect},
		{str: "d", exp: FromValDetect},
		{str: "det", exp: FromValDetect},
		{str: "any", exp: FromValDetect},
		{str: "a", exp: FromValDetect},
		{str: " detect", exp: FromValDetect},
		{str: "detect ", exp: FromValDetect},
		{str: " detect ", exp: FromValDetect},

		{str: "bech32", exp: FromValBech32},
		{str: "BECH32", exp: FromValBech32},
		{str: "bEcH32", exp: FromValBech32},
		{str: "b32", exp: FromValBech32},
		{str: "32", exp: FromValBech32},

		{str: "base64", exp: FromValBase64},
		{str: "BASE64", exp: FromValBase64},
		{str: "BasE64", exp: FromValBase64},
		{str: "b64", exp: FromValBase64},
		{str: "64", exp: FromValBase64},

		{str: "hex", exp: FromValHex},
		{str: "HEX", exp: FromValHex},
		{str: "Hex", exp: FromValHex},
		{str: "h", exp: FromValHex},
		{str: "x", exp: FromValHex},
	}

	for _, tc := range tests {
		name := tc.str
		if len(tc.str) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			fromVal, err := ToFromVal(tc.str)
			AssertErrorContents(t, err, tc.expErr, "ToFromVal(%q)", tc.str)
			assert.Equal(t, tc.exp, fromVal, "ToFromVal(%q)", tc.str)
		})
	}
}

func TestConvertAndPrintAll(t *testing.T) {
	mustConvertAndEncodeBech32 := func(hrp string, bz []byte) string {
		rv, err := bech32.ConvertAndEncode(hrp, bz)
		if err != nil {
			panic(err)
		}
		return rv
	}
	multiFmt := "[%d/%d] %s => %s"

	tests := []struct {
		name   string
		cfg    *CmdConfig
		args   []string
		expErr string
		expOut []string
	}{
		{
			name: "nil args",
			cfg: &CmdConfig{
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Quiet:    false,
				FromVal:  FromValHex,
				Count:    0,
			},
			args: nil,
		},
		{
			name: "empty args",
			cfg: &CmdConfig{
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Quiet:    false,
				FromVal:  FromValHex,
				Count:    0,
			},
			args: []string{},
		},
		{
			name: "1 arg",
			cfg: &CmdConfig{
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Quiet:    false,
				FromVal:  FromValHex,
				Count:    1,
			},
			args: []string{"0a0b0c"},
			expOut: []string{
				"newhrp1pg9sczch4yg",
				"CgsM",
				"0a0b0c",
			},
		},
		{
			name: "1 invalid arg",
			cfg: &CmdConfig{
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Quiet:    false,
				FromVal:  FromValHex,
				Count:    1,
			},
			args:   []string{"x"},
			expErr: `could not decode "x" as hex: encoding/hex: invalid byte: U+0078 'x'`,
		},
		{
			name: "2 good args",
			cfg:  &CmdConfig{HRPs: []string{"myhrp", "yourhrp"}, FromVal: FromValHex, Count: 2},
			args: []string{"0a", "0b"},
			expOut: []string{
				fmt.Sprintf(multiFmt, 1, 2, "0a", mustConvertAndEncodeBech32("myhrp", []byte{0x0a})),
				fmt.Sprintf(multiFmt, 1, 2, "0a", mustConvertAndEncodeBech32("yourhrp", []byte{0x0a})),
				fmt.Sprintf(multiFmt, 2, 2, "0b", mustConvertAndEncodeBech32("myhrp", []byte{0x0b})),
				fmt.Sprintf(multiFmt, 2, 2, "0b", mustConvertAndEncodeBech32("yourhrp", []byte{0x0b})),
			},
		},
		{
			name:   "2 args first bad",
			cfg:    &CmdConfig{HRPs: []string{"hhrrpp"}, FromVal: FromValHex, Count: 2},
			args:   []string{"x", "0a"},
			expErr: `could not decode "x" as hex: encoding/hex: invalid byte: U+0078 'x'`,
		},
		{
			name:   "2 args second bad",
			cfg:    &CmdConfig{HRPs: []string{"hhrrpp"}, FromVal: FromValHex, Count: 2},
			args:   []string{"0a", "x"},
			expErr: `could not decode "x" as hex: encoding/hex: invalid byte: U+0078 'x'`,
			expOut: []string{fmt.Sprintf(multiFmt, 1, 2, "0a", mustConvertAndEncodeBech32("hhrrpp", []byte{0x0a}))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expOut == nil {
				tc.expOut = []string{}
			}
			var buffer bytes.Buffer
			tc.cfg.Writer = &buffer

			err := ConvertAndPrintAll(tc.cfg, tc.args)
			AssertErrorContents(t, err, tc.expErr, "ConvertAndPrintAll error")
			outStr := buffer.String()
			outLines := strings.Split(outStr, "\n")
			if len(outLines[len(outLines)-1]) == 0 {
				outLines = outLines[:len(outLines)-1]
			}
			assert.Equal(t, tc.expOut, outLines, "ConvertAndPrintAll printed output")
		})
	}
}

func TestConvertAndPrint(t *testing.T) {
	mustConvertAndEncodeBech32 := func(hrp string, bz []byte) string {
		rv, err := bech32.ConvertAndEncode(hrp, bz)
		if err != nil {
			panic(err)
		}
		return rv
	}
	multiFmt := "[%d/%d] %s => %s"

	tests := []struct {
		name   string
		cfg    *CmdConfig
		arg    string
		i      int
		expErr string
		expOut []string
	}{
		{
			name:   "invalid input",
			cfg:    &CmdConfig{FromVal: FromValHex},
			arg:    "x",
			expErr: `could not decode "x" as hex: encoding/hex: invalid byte: U+0078 'x'`,
		},
		// Not sure how to make EncodeAddr return an error.
		{
			name: "1 count to bech32 base64 and hex",
			cfg: &CmdConfig{
				FromVal:  FromValBech32,
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Count:    1,
			},
			arg: mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 20)),
			i:   1,
			expOut: []string{
				mustConvertAndEncodeBech32("newhrp", bytes.Repeat([]byte{0x8a}, 20)),
				"ioqKioqKioqKioqKioqKioqKioo=",
				strings.Repeat("8a", 20),
			},
		},
		{
			name: "2 count to bech32 base64 and hex",
			cfg: &CmdConfig{
				FromVal:  FromValBech32,
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Count:    2,
			},
			arg: mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 5)),
			i:   1,
			expOut: []string{
				fmt.Sprintf(multiFmt, 1, 2,
					mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 5)),
					mustConvertAndEncodeBech32("newhrp", bytes.Repeat([]byte{0x8a}, 5))),
				fmt.Sprintf(multiFmt, 1, 2,
					mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 5)),
					"ioqKioo="),
				fmt.Sprintf(multiFmt, 1, 2,
					mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 5)),
					strings.Repeat("8a", 5)),
			},
		},
		{
			name: "2 count quiet to bech32 base64 and hex",
			cfg: &CmdConfig{
				FromVal:  FromValBech32,
				HRPs:     []string{"newhrp"},
				ToHex:    true,
				ToBase64: true,
				Count:    2,
				Quiet:    true,
			},
			arg: mustConvertAndEncodeBech32("oldhrp", bytes.Repeat([]byte{0x8a}, 5)),
			i:   1,
			expOut: []string{
				mustConvertAndEncodeBech32("newhrp", bytes.Repeat([]byte{0x8a}, 5)),
				"ioqKioo=",
				strings.Repeat("8a", 5),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expOut == nil {
				tc.expOut = []string{}
			}
			var buffer bytes.Buffer
			tc.cfg.Writer = &buffer

			err := ConvertAndPrint(tc.cfg, tc.arg, tc.i)
			AssertErrorContents(t, err, tc.expErr, "ConvertAndPrint error")
			outStr := buffer.String()
			outLines := strings.Split(outStr, "\n")
			if len(outLines[len(outLines)-1]) == 0 {
				outLines = outLines[:len(outLines)-1]
			}
			assert.Equal(t, tc.expOut, outLines, "ConvertAndPrint printed output")
		})
	}
}

func TestGetAddrBytes(t *testing.T) {
	mustConvertAndEncodeBech32 := func(hrp string, bz []byte) string {
		rv, err := bech32.ConvertAndEncode(hrp, bz)
		if err != nil {
			panic(err)
		}
		return rv
	}
	tests := []struct {
		name   string
		cfg    *CmdConfig
		input  string
		exp    []byte
		expErr string
	}{
		{
			name:  "empty string",
			input: "",
			exp:   []byte{},
		},
		{
			name:  "detect empty bech32",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: mustConvertAndEncodeBech32("emptything", []byte{}),
			exp:   []byte{},
		},
		{
			name:  "detect short bech32",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: mustConvertAndEncodeBech32("shortthing", bytes.Repeat([]byte{0}, 5)),
			exp:   bytes.Repeat([]byte{0}, 5),
		},
		{
			name:  "detect normal bech32",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: mustConvertAndEncodeBech32("normalthing", bytes.Repeat([]byte{1}, 20)),
			exp:   bytes.Repeat([]byte{1}, 20),
		},
		{
			name:   "detect long bech32",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  "longthing1qgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqd8e8n8",
			expErr: `could not detect "longthing1qgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqd8e8n8" type between "bech32" "base64"`,
		},
		{
			name:  "detect single byte hex lower",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "0f",
			exp:   []byte{0x0f},
		},
		{
			name:  "detect short hex lower",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: strings.Repeat("0a", 5),
			exp:   bytes.Repeat([]byte{0x0a}, 5),
		},
		{
			name:   "detect normal hex lower",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  strings.Repeat("0b", 20),
			expErr: `could not detect "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b" type between "base64" "hex"`,
		},
		{
			name:   "detect long hex lower",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  strings.Repeat("0c", 32),
			expErr: `could not detect "0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c0c" type between "base64" "hex"`,
		},
		{
			name:  "detect single byte hex upper",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "0F",
			exp:   []byte{0x0f},
		},
		{
			name:  "detect short hex upper",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: strings.Repeat("0A", 5),
			exp:   bytes.Repeat([]byte{0x0a}, 5),
		},
		{
			name:   "detect normal hex upper",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  strings.Repeat("0B", 20),
			expErr: `could not detect "0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B0B" type between "base64" "hex"`,
		},
		{
			name:   "detect long hex upper",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  strings.Repeat("0C", 32),
			expErr: `could not detect "0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C0C" type between "base64" "hex"`,
		},
		{
			name:  "detect single byte base64",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "Bg==",
			exp:   []byte{6},
		},
		{
			name:  "detect short base64",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "DQ0NDQ0=",
			exp:   bytes.Repeat([]byte{0x0d}, 5),
		},
		{
			name:  "detect normal base64",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "AgICAgICAgICAgICAgICAgICAgI=",
			exp:   bytes.Repeat([]byte{2}, 20),
		},
		{
			name:  "detect long base64",
			cfg:   &CmdConfig{FromVal: FromValDetect},
			input: "BwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwc=",
			exp:   bytes.Repeat([]byte{7}, 32),
		},
		{
			name:   "detect not bech32 hex or base64",
			cfg:    &CmdConfig{FromVal: FromValDetect},
			input:  "x",
			expErr: `could not decode "x" as bech32, hex, or base64`,
		},
		{
			name:  "bech32 empty bytes",
			cfg:   &CmdConfig{FromVal: FromValBech32},
			input: mustConvertAndEncodeBech32("one", []byte{9}),
			exp:   []byte{9},
		},
		{
			name:  "bech32 1 byte",
			cfg:   &CmdConfig{FromVal: FromValBech32},
			input: mustConvertAndEncodeBech32("one", []byte{9}),
			exp:   []byte{9},
		},
		{
			name:  "bech32 5 bytes",
			cfg:   &CmdConfig{FromVal: FromValBech32},
			input: mustConvertAndEncodeBech32("five", bytes.Repeat([]byte{5}, 5)),
			exp:   bytes.Repeat([]byte{5}, 5),
		},
		{
			name:  "bech32 20 bytes",
			cfg:   &CmdConfig{FromVal: FromValBech32},
			input: mustConvertAndEncodeBech32("twenty", bytes.Repeat([]byte{20}, 20)),
			exp:   bytes.Repeat([]byte{20}, 20),
		},
		{
			name:  "bech32 32 bytes",
			cfg:   &CmdConfig{FromVal: FromValBech32},
			input: mustConvertAndEncodeBech32("twenty", bytes.Repeat([]byte{32}, 32)),
			exp:   bytes.Repeat([]byte{32}, 32),
		},
		{
			name:   "bech32 invalid",
			cfg:    &CmdConfig{FromVal: FromValBech32},
			input:  "notabech32",
			expErr: `could not decode "notabech32" as bech32: decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:  "hex 1 byte lower",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: "a6",
			exp:   []byte{0xa6},
		},
		{
			name:  "hex 1 byte upper",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: "A6",
			exp:   []byte{0xa6},
		},
		{
			name:  "hex 5 bytes lower",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: "dead00beef",
			exp:   []byte{0xde, 0xad, 0x00, 0xbe, 0xef},
		},
		{
			name:  "hex 5 bytes upper",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: "DEAD00BEEF",
			exp:   []byte{0xde, 0xad, 0x00, 0xbe, 0xef},
		},
		{
			name:  "hex 20 bytes lower",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: strings.Repeat("0b", 20),
			exp:   bytes.Repeat([]byte{0x0b}, 20),
		},
		{
			name:  "hex 20 bytes upper",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: strings.Repeat("0B", 20),
			exp:   bytes.Repeat([]byte{0x0b}, 20),
		},
		{
			name:  "hex 20 bytes mixed",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: strings.Repeat("eF", 20),
			exp:   bytes.Repeat([]byte{0xef}, 20),
		},
		{
			name:  "hex 32 bytes lower",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: strings.Repeat("ac", 32),
			exp:   bytes.Repeat([]byte{0xac}, 32),
		},
		{
			name:  "hex 32 bytes upper",
			cfg:   &CmdConfig{FromVal: FromValHex},
			input: strings.Repeat("AC", 32),
			exp:   bytes.Repeat([]byte{0xac}, 32),
		},
		{
			name:   "hex invalid",
			cfg:    &CmdConfig{FromVal: FromValHex},
			input:  "nothex",
			expErr: `could not decode "nothex" as hex: encoding/hex: invalid byte: U+006E 'n'`,
		},
		{
			name:  "base64 1 byte",
			cfg:   &CmdConfig{FromVal: FromValBase64},
			input: "Bg==",
			exp:   []byte{6},
		},
		{
			name:  "base64 5 bytes",
			cfg:   &CmdConfig{FromVal: FromValBase64},
			input: "DQ0NDQ0=",
			exp:   bytes.Repeat([]byte{0x0d}, 5),
		},
		{
			name:  "base64 20 bytes",
			cfg:   &CmdConfig{FromVal: FromValBase64},
			input: "AgICAgICAgICAgICAgICAgICAgI=",
			exp:   bytes.Repeat([]byte{2}, 20),
		},
		{
			name:  "base64 32 bytes",
			cfg:   &CmdConfig{FromVal: FromValBase64},
			input: "BwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwc=",
			exp:   bytes.Repeat([]byte{7}, 32),
		},
		{
			name:   "base64 invalid",
			cfg:    &CmdConfig{FromVal: FromValBase64},
			input:  "invalidbase64",
			expErr: `could not decode "invalidbase64" as base64: illegal base64 data at input byte 12`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := GetAddrBytes(tc.cfg, tc.input)
			AssertErrorContents(t, err, tc.expErr, "GetAddrBytes error")
			assert.Equal(t, tc.exp, addr, "GetAddrBytes bytes")
		})
	}
}

func TestEncodeAddr(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *CmdConfig
		addr   []byte
		exp    []string
		expErr string
	}{
		{
			name: "one hrp",
			cfg:  &CmdConfig{HRPs: []string{"abc"}},
			addr: bytes.Repeat([]byte{0}, 20),
			exp:  []string{"abc1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqql9d4m7"},
		},
		{
			name: "two hrps",
			cfg:  &CmdConfig{HRPs: []string{"abc", "def"}},
			addr: bytes.Repeat([]byte{1}, 20),
			exp: []string{
				"abc1qyqszqgpqyqszqgpqyqszqgpqyqszqgp74v53l",
				"def1qyqszqgpqyqszqgpqyqszqgpqyqszqgpmxrvj5",
			},
		},
		{
			name: "to base64",
			cfg:  &CmdConfig{ToBase64: true},
			addr: bytes.Repeat([]byte{2}, 20),
			exp:  []string{"AgICAgICAgICAgICAgICAgICAgI="},
		},
		{
			name: "to hex",
			cfg:  &CmdConfig{ToHex: true},
			addr: bytes.Repeat([]byte{3}, 20),
			exp:  []string{"0303030303030303030303030303030303030303"},
		},
		{
			name: "no flags",
			cfg:  &CmdConfig{},
			addr: bytes.Repeat([]byte{4}, 20),
			exp:  []string{"0404040404040404040404040404040404040404"},
		},
		{
			name: "three hrps to base64 and to hex",
			cfg:  &CmdConfig{HRPs: []string{"abc", "def", "xyz"}, ToBase64: true, ToHex: true},
			addr: bytes.Repeat([]byte{5}, 20),
			exp: []string{
				"abc1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg90pd5a4",
				"def1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg92jzv77",
				"xyz1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9fzxqpn",
				"BQUFBQUFBQUFBQUFBQUFBQUFBQU=",
				"0505050505050505050505050505050505050505",
			},
		},
		{
			name: "single byte",
			cfg:  &CmdConfig{HRPs: []string{"hrp"}, ToBase64: true, ToHex: true},
			addr: []byte{6},
			exp: []string{
				"hrp1qcfv3tsc",
				"Bg==",
				"06",
			},
		},
		{
			name: "five bytes",
			cfg:  &CmdConfig{HRPs: []string{"hrp"}, ToBase64: true, ToHex: true},
			addr: bytes.Repeat([]byte{0x0d}, 5),
			exp: []string{
				"hrp1p5xs6rgd5ryvhl",
				"DQ0NDQ0=",
				"0d0d0d0d0d",
			},
		},
		{
			name: "32 bytes",
			cfg:  &CmdConfig{HRPs: []string{"hrp"}, ToBase64: true, ToHex: true},
			addr: bytes.Repeat([]byte{7}, 32),
			exp: []string{
				"hrp1qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qursax9spp",
				"BwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBwc=",
				"0707070707070707070707070707070707070707070707070707070707070707",
			},
		},
		{
			name: "empty bytes",
			cfg:  &CmdConfig{HRPs: []string{"ipq"}, ToBase64: true, ToHex: true},
			addr: []byte{},
			exp:  []string{"ipq1dsr0xv", "", ""},
		},
		// Not sure if there's a way to make bech32.ConvertAndEncode return an error.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outputs, err := EncodeAddr(tc.cfg, tc.addr)
			AssertErrorContents(t, err, tc.expErr, "EncodeAddr error")
			assert.Equal(t, tc.exp, outputs, "EncodeAddr outputs")
		})
	}
}
