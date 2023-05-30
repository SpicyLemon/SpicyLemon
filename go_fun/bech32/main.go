package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// CmdConfig contains all the flags and info about the command being run.
type CmdConfig struct {
	// ToHRPs are the strings provided with the --hrp flag(s).
	ToHRPs []string
	// ToHex indicates the presence of the --hex flag.
	ToHex bool
	// ToBase64 indicates the presence of the --base64 flag.
	ToBase64 bool
	// ToRaw indicates the presence of the --raw flag.
	ToRaw bool

	// Quiet indicates the presence of the --quiet flag.
	Quiet bool

	// From is the string provided with the --from arg.
	From string
	// FromVal is the normalized version of the From field.
	FromVal FromVal

	// Writer is what's used for printing the output.
	Writer io.Writer
	// Count is a count of the args provided.
	Count int
	// Inputs is all the input strings to convert
	Inputs []string
}

// Prep sets up the final stuff needed in the CmdConfig before trying to do stuff.
func (c *CmdConfig) Prep(cmd *cobra.Command, args []string) error {
	c.Writer = cmd.OutOrStdout()

	c.Inputs = args
	c.Count = len(c.Inputs)

	var err error
	c.FromVal, err = ToFromVal(c.From)
	if err != nil {
		return err
	}

	if len(c.Inputs) == 0 {
		// Read from stdin if either:
		// a) it's being piped too (e.g. <cmd1> | bech32)
		// b) it's not a *File (unit tests).
		canReadStdin := true
		stdin := cmd.InOrStdin()
		if stdinFile, isFile := stdin.(*os.File); isFile {
			stdinStat, _ := stdinFile.Stat()
			// If the mode is character device, it's the terminal (not a pipe).
			canReadStdin = (stdinStat.Mode() & os.ModeCharDevice) == 0
		}

		if canReadStdin {
			var bzIn []byte
			bzIn, err = io.ReadAll(stdin)
			if err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
			if len(bzIn) > 0 {
				if c.FromVal == FromValRaw {
					c.Inputs = []string{string(bzIn)}
				} else {
					c.Inputs = strings.Fields(string(bzIn))
				}
				c.Count = len(c.Inputs)
			}
		}
	}

	if c.Count == 0 {
		return errors.New("no input addresses provided")
	}

	return nil
}

// OutputTypeDefined returns true if any output types have been dictated.
func (c *CmdConfig) OutputTypeDefined() bool {
	return len(c.ToHRPs) > 0 || c.ToHex || c.ToBase64 || c.ToRaw
}

// A FromVal is a valid string to provide with the --from flag.
type FromVal string

// String returns this FromVal as a string.
func (v FromVal) String() string {
	return string(v)
}

const (
	// FromValDetect attempts to detect the input type.
	FromValDetect FromVal = "detect"
	// FromValBech32 decodes the input as a bech32 string.
	FromValBech32 FromVal = "bech32"
	// FromValBase64 decodes the input as a base64 encoded string.
	FromValBase64 FromVal = "base64"
	// FromValHex decodes the input as a hex string.
	FromValHex FromVal = "hex"
	// FromValRaw indicates that the input is raw and should not be decoded.
	FromValRaw FromVal = "raw"
)

// FromValOptionsStr is a string indicating all the valid --from options.
var FromValOptionsStr = `"` + strings.Join([]string{
	FromValDetect.String(), FromValBech32.String(), FromValBase64.String(), FromValHex.String(), FromValRaw.String(),
}, `" "`) + `"`

// ToFromVal converts the provided string into a FromVal or returns an error.
func ToFromVal(str string) (FromVal, error) {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case string(FromValDetect), "d", "det", "any", "a", "":
		return FromValDetect, nil
	case string(FromValBech32), "b32", "32":
		return FromValBech32, nil
	case string(FromValBase64), "b64", "64":
		return FromValBase64, nil
	case string(FromValHex), "h", "x":
		return FromValHex, nil
	case string(FromValRaw), "r":
		return FromValRaw, nil
	}
	return FromValDetect, fmt.Errorf("invalid --from value %q, must be one of %s", str, FromValOptionsStr)
}

func NewRootCmd() *cobra.Command {
	cmdConfig := &CmdConfig{}
	cmd := &cobra.Command{
		Use:   "bech32 <addr> [<addr2> ...]",
		Short: "Convert bech32 strings",
		Long: `Convert bech32 strings to hex, base64, or new HRPs.

If none of --hrp --base64 --hex or --raw are provided, --hex is used.
The --hrp flag can be provided multiple times.
Multiple HRPs can be provided after --hrp by separating each with commas.

When multiple output types are requested, they will be in this order:
  1. Bech32(s) in the order the HRPs were provided
  2. Base64
  3. Hex
  4. Raw

Example Usage:
$ bech32 xyz1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9fzxqpn --hrp abc
$ bech32 0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b --hrp abc,def --from hex
$ bech32 5c5c5c5c5c5c --hrp abc --hrp def --hex --from base64

`,
		PreRunE: cmdConfig.Prep,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ConvertAndPrintAll(cmdConfig)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringSliceVar(&cmdConfig.ToHRPs, "hrp", cmdConfig.ToHRPs, "Output address(es) as bech32 with provided HRPs")
	cmd.Flags().BoolVarP(&cmdConfig.ToBase64, "base64", "b", cmdConfig.ToHex, "Output address(es) as base64")
	cmd.Flags().BoolVarP(&cmdConfig.ToHex, "hex", "x", cmdConfig.ToHex, "Output address(es) as hex")
	cmd.Flags().BoolVarP(&cmdConfig.ToRaw, "raw", "r", cmdConfig.ToRaw, "Output raw address(es) bytes")
	cmd.Flags().BoolVarP(&cmdConfig.Quiet, "quiet", "q", cmdConfig.Quiet, "Only print the converted output")
	cmd.Flags().StringVar(&cmdConfig.From, "from", string(FromValDetect),
		"The type of strings being provided, options: "+FromValOptionsStr,
	)

	return cmd
}

// ConvertAndPrintAll converts and prints all the provided args.
func ConvertAndPrintAll(cfg *CmdConfig) error {
	for i, arg := range cfg.Inputs {
		err := ConvertAndPrint(cfg, arg, i+1)
		if err != nil {
			return err
		}
	}

	return nil
}

// ConvertAndPrint converts the provided argument and prints results to the provided writer.
func ConvertAndPrint(cfg *CmdConfig, arg string, i int) error {
	addr, err := GetAddrBytes(cfg, arg)
	if err != nil {
		return err
	}

	outputs, err := EncodeAddr(cfg, addr)
	if err != nil {
		return fmt.Errorf("error encoding %q %v: %w", arg, addr, err)
	}

	lead := ""
	if cfg.Count > 1 && !cfg.Quiet {
		lead = fmt.Sprintf("[%d/%d] %s => ", i, cfg.Count, arg)
	}

	for _, output := range outputs {
		_, err = fmt.Fprintf(cfg.Writer, "%s%s\n", lead, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAddrBytes decodes the provided address string.
func GetAddrBytes(cfg *CmdConfig, input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	if cfg.FromVal == FromValRaw {
		return []byte(input), nil
	}

	isDetect := cfg.FromVal == FromValDetect || len(cfg.FromVal)+len(cfg.From) == 0

	var addr, addrTmp []byte
	var err error
	okTypes := make([]string, 0, 3)

	if cfg.FromVal == FromValBech32 || isDetect {
		_, addrTmp, err = bech32.DecodeAndConvert(input)
		if err == nil {
			okTypes = append(okTypes, FromValBech32.String())
			addr = addrTmp
		}
	}

	if cfg.FromVal == FromValBase64 || isDetect {
		addrTmp, err = base64.StdEncoding.DecodeString(input)
		if err == nil {
			okTypes = append(okTypes, FromValBase64.String())
			addr = addrTmp
		}
	}

	if cfg.FromVal == FromValHex || isDetect {
		addrTmp, err = hex.DecodeString(input)
		if err == nil {
			okTypes = append(okTypes, FromValHex.String())
			addr = addrTmp
		}
	}

	if !isDetect {
		if err != nil {
			return nil, fmt.Errorf("could not decode %q as %s: %w", input, cfg.FromVal, err)
		}
		return addr, nil
	}

	switch len(okTypes) {
	case 0:
		return nil, fmt.Errorf("could not decode %q as bech32, hex, or base64", input)
	case 1:
		return addr, nil
	default:
		return nil, fmt.Errorf(`could not detect %q type between "%s"`, input, strings.Join(okTypes, `" "`))
	}
}

// EncodeAddr encodes the provided address as desired.
func EncodeAddr(cfg *CmdConfig, addr []byte) ([]string, error) {
	var err error
	rv := make([]string, len(cfg.ToHRPs), len(cfg.ToHRPs)+3)
	for i, hrp := range cfg.ToHRPs {
		rv[i], err = bech32.ConvertAndEncode(hrp, addr)
		if err != nil {
			return nil, err
		}
	}
	if cfg.ToBase64 {
		rv = append(rv, base64.StdEncoding.EncodeToString(addr))
	}
	if cfg.ToHex || !cfg.OutputTypeDefined() {
		rv = append(rv, hex.EncodeToString(addr))
	}
	if cfg.ToRaw {
		rv = append(rv, string(addr))
	}
	return rv, nil
}

func main() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
