# SpicyLemon / go_fun / bech32

The `bech32` program is a command-line utility for converting bech32, hex, and base64 strings.

## Usage

Input strings can either be provided as arguments or piped in (but not both).

By default, it will attempt to identify what format the input strings are (bech32, hex, base64).
If the input is valid for multiple formats, an error is returned.
You can tell it what format the input is using the `--from {bech32|hex|base64|raw}` flag.

The output format(s) are controlled using the `--hrp <string>`, `--hex`, `--base64`, and `--raw` flags.
Multiple of these can be provided to get the output in multiple forms.

Example:
```shell
$  bech32 xyz1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9fzxqpn --hrp abc
abc1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg90pd5a4
```

When multiple output types are requested, they will be in this order:
  1. Bech32(s) in the order the HRPs were provided
  2. Base64
  3. Hex
  4. Raw

When multiple input strings are provided, each line of output will contain an index/count, the input string, then an output string.
In such cases, if you only want the output strings, use the `--quiet` flag.

Example:
```shell
$ bech32 0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b 5c5c5c5c5c5c --hrp abc,def --from hex
[1/2] 0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b => abc1pv9skzctpv9skzctpv9skzctpv9skzctpf7pnd
[1/2] 0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b => def1pv9skzctpv9skzctpv9skzctpv9skzcty63esx
[2/2] 5c5c5c5c5c5c => abc1t3w9chzutshm29wd
[2/2] 5c5c5c5c5c5c => def1t3w9chzutssw9h0q
```

With --quiet:
```shell
bech32 0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b 5c5c5c5c5c5c --hrp abc,def --from hex --quiet
abc1pv9skzctpv9skzctpv9skzctpv9skzctpf7pnd
def1pv9skzctpv9skzctpv9skzctpv9skzcty63esx
abc1t3w9chzutshm29wd
def1t3w9chzutssw9h0q
```

## Installation

Using make:
```shell
$ make install
```

Without make:
```shell
$ go install -mod=readonly -ldflags '-w -s' -trimpath
```

Or, to build it without installing it:
```shell
$ make build
```

Build without make:
```shell
$ go build -o build/ -mod=readonly -ldflags '-w -s' -trimpath
```

