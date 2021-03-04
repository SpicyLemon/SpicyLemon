# ticker.sh

> Real-time stock tickers from the command-line.

`ticker.sh` is a simple shell script using the Yahoo Finance API as a data source. It features colored output and is able to display pre- and post-market prices (denoted with `<` or `>` respectively).

It is based off a similar script by Patrick Stadler: https://github.com/pstadler/ticker.sh

![ticker.sh](/ticker/screenshot.png?raw=true Screenshot)

## Install

Copy the ticker.sh file to your computer and make it executable.

Requires [jq](https://stedolan.github.io/jq/), a versatile command-line JSON processor.

## Usage

```sh
# Single symbol:
$ ./ticker.sh AAPL

# Multiple symbols:
$ ./ticker.sh AAPL MSFT GOOG DOGE-USD

# Use different colors:
$ COLOR_BOLD="\e[38;5;248m" \
  COLOR_GREEN="\e[38;5;154m" \
  COLOR_RED="\e[38;5;202m" \
  ./ticker.sh AAPL

# Disable colors:
$ NO_COLOR=1 ./ticker.sh AAPL
```

Use [yahoo finance symbol lookup](https://finance.yahoo.com/lookup/) to look up valid symbols.
