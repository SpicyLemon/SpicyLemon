#!/bin/bash
# This script uses the Yahoo Finance API to get ticker quotes, then formats and outputs the results.
# This script is designed to be executed (as opposed to being sourced).
# To install it, place it in one of your PATH directories.
#
# To get debug information, set the DEBUG environment variable to 1 (or anything other than an empty string, really).

if ! command -v 'jq' > /dev/null 2>&1; then
    printf 'Missing required command: jq\n' >&2
    jq >&2  # Possibly provides shell/system specific information about the missing command.
    exit $?
fi

print_usage () {
    cat << EOF
ticker.sh - Get ticker info from the Yahoo Finance API and format it nicely.

Usage: ./ticker.sh <symbol> [<symbol 2> ...]

Special "symbols":
    These special symbols should be wrapped in [], (), {} or <> (whichever is easiest in your environment).
    Depending on your environment, you may also need to quote them as arguments.
        br - Adds an empty line to the output. E.g. '[br]'
        hr - Adds a line of dashes to the output. E.g. '(hr)'

Example:
> ./ticker AAPL MSFT GOOG '[hr]' DOGE-USD

The following environment variables can control things:
    DEBUG    - Set to a non-empty string to output debug information.
    NO_COLOR - Set to a non-empty string to not use coloring.
    COLOR_SYMBOL    - Color code for the symbol.                Default: ''
    COLOR_AMOUNT    - Color code for the amount.                Default: '1;37' (bold white)
    COLOR_UP        - Color code for a positive diff.           Default: '32' (green)
    COLOR_EVEN      - Color code for a zero diff.               Default: ''
    COLOR_DOWN      - Color code for a negative diff.           Default: '31' (red)
    COLOR_INDICATOR - Color code for the market indicator.      Default: ''
    COLOR_NAME      - Color code for the name.                  Default: ''
    COLOR_RESET     - Color code for ending a color section.    Default: '0' (reset all color aspects)
EOF

}

symbols=()
symbols_to_get=()

hr='<hr>'
br='<br>'

while [[ "$#" -gt '0' ]]; do
    # Check for special cases <br> and <hr> allowing for multiple bracket styles and casing.
    # Techinically this will catch stuff like <bR], but this reads better than trying to match stuff,
    # and catching extra stuff like that isn't going to hurt anything.
    if [[ "$1" == '--help' || "$1" == '-h' ]]; then
        help_requested=1
    elif [[ "$1" =~ ^[\[\(\{\<][bB][rR][\]\)\}\>]$ ]]; then
        symbols+=( "$br" )
    elif [[ "$1" =~ ^[\[\(\{\<][hH][rR][\]\)\}\>]$ ]]; then
        symbols+=( "$hr" )
    else
        symbols+=( $1 )
        symbols_to_get+=( $1 )
    fi
    shift
done

if [[ "${#symbols_to_get[@]}" -eq '0' ]]; then
    print_usage
    exit 0
fi
if [[ -n "$help_requested" ]]; then
    print_usage
fi

to_escape_code () {
    local val
    val="$( sed -E 's/^[[:space:]]+//; s/[[:space:]]$//;' <<< "$1" )"
    if [[ "${val:=$2}" =~ ^[[:digit:]\;]+$ ]]; then
        val="\e[${val}m"
    fi
    printf '%b' "$val"
}

if [[ -z "$NO_COLOR" ]]; then
    c_sym="$( to_escape_code "$COLOR_SYMBOL" '' )"
    c_amt="$( to_escape_code "$COLOR_AMOUNT" '1;37' )"
    c_up="$( to_escape_code "$COLOR_UP" '32' )"
    c_even="$( to_escape_code "$COLOR_EVEN" '' )"
    c_down="$( to_escape_code "$COLOR_DOWN" '31' )"
    c_ind="$( to_escape_code "$COLOR_INDICATOR" '' )"
    c_name="$( to_escape_code "$COLOR_NAME" '' )"
    c_rst="$( to_escape_code "$COLOR_RESET" '0' )"
fi

api_endpoint="https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com"
fields=(symbol shortName marketState regularMarketPrice regularMarketChange regularMarketChangePercent \
  preMarketPrice preMarketChange preMarketChangePercent postMarketPrice postMarketChange postMarketChangePercent)

url_symbols=$( IFS=,; echo "${symbols_to_get[*]}" )
url_fields=$( IFS=,; echo "${fields[*]}" )

api_url="$api_endpoint&fields=$url_fields&symbols=$url_symbols"

[[ -n "$DEBUG" ]] && printf 'api_url: [%s]\n' "$api_url"

fullResults=$( curl --silent "$api_url" )
[[ -n "$DEBUG" ]] && printf 'fullResults:\n%s\n' "$fullResults"

parsedResults="$(
    jq -r 'def hasNonZero($k): has($k) and .[$k] != null and .[$k] != 0;
           def safeStr($k): if .[$k] == null then "0" else (.[$k]|tostring) end;
        .quoteResponse.result[] |
        .symbol + " " + (
            if .marketState == "PRE" and hasNonZero("preMarketChange") then
                safeStr("preMarketPrice") + " " + safeStr("preMarketChange") + " " + safeStr("preMarketChangePercent") + " <"
            elif .marketState != "REGULAR" and hasNonZero("postMarketChange") then
                safeStr("postMarketPrice") + " " + safeStr("postMarketChange") + " " + safeStr("postMarketChangePercent") + " >"
            else
                safeStr("regularMarketPrice") + " " + safeStr("regularMarketChange") + " " + safeStr("regularMarketChangePercent") + " ="
            end
        ) + " " + .shortName' <<< "$fullResults"
)"
[[ -n "$DEBUG" ]] && printf 'parsedResults:\n%s\n' "$parsedResults"

width='85'  # 50 for the symbols, numbers, and indicator, leaving 35 for the name.
# If tput is available, use that to get the window width
if command -v 'tput' > /dev/null 2>&1; then
    width="$( tput cols )"
fi

for symbol in ${symbols[*]}; do
    if [[ "$symbol" == "$br" ]]; then
        printf '\n'
        continue
    fi
    if [[ "$symbol" == "$hr" ]]; then
        printf "%${width}s\n" '' | tr ' ' '-'
        continue
    fi

    symbolLine="$( grep -i "^$symbol " <<< "$parsedResults" )"
    [[ -n "$DEBUG" ]] && printf 'symbolLine: [%s]\n' "$symbolLine"

    if [[ -z "$symbolLine" ]]; then
        printf 'No results for symbol "%s"\n' $symbol
        continue
    fi

    read symbol price diff percent marketStateIndicator shortName <<< "$symbolLine"

    if [[ "$diff" =~ ^0(\.0*)?$ ]]; then
        c_diff="$c_even"
    elif [[ "$diff" =~ ^- ]]; then
        c_diff="$c_down"
    else
        c_diff="$c_up"
    fi

    if [[ "$marketStateIndicator" == '=' ]]; then
        marketStateIndicator=' '
    fi

    # Some short names have info like " - Class 2" or " Class A Common Stock" at the end.
    # This truncation first removes those "class" parts.
    # Then it removes the "Inc" part (e.g. ", Inc.", " Inc.", ", Inc", " Inc")
    # Then it limits what's left to 28 characters.
    # And lastly, removes any trailing whitespace.
    shortNameT="$( sed -E 's/( -)? [cC][lL][aA][sS][sS] [[:alnum:]].*$//; s/,? [iI][nN][cC]\.? *$//;' <<< "$shortName" | head -c 28 | sed 's/ *$//' )"

    # Line Breakdown:
    #   index  format    len  description
    # symbol coloring on
    #   1-10   '%-10s'   10   Ticker symbol, left justified.
    # color reset
    #   11     ' '       1    Space.
    # amount coloring on
    #   12-24  '%13.6f'  13   Value: 6 whole digits, 1 decimal, 6 fractional digits.
    #                         This will be more than 13 characters for values over 999,999.99999.
    # color reset
    #   25     ' '       1    Space.
    # up/even/down coloring on
    #   26-38  '%+13.6f' 13   value change: 1 +/- sign, 5 whole digits, 1 decimal, 6 fractional digits.
    #                         This will be more than 13 characters for values over 99,999.999999.
    #   39     ' '       1    Space.
    #   40-46  '%+7.2f'  7    Percent change: 1 +/- sign, 3 whole digits, 1 decimal, 2 fractional digits.
    #                         This will be more than 7 characters for absolute values over 999.99.
    #   47     '%%'      1    Percent symbol.
    # color reset
    #   48     ' '       1    Space.
    #   49     '%1s'     1    Market state indicator character: ' ', '<', or '>'.
    #   50     ' '       1    Space.
    # name coloring on
    #   51-80  '[%s]'    30   Short name: 1 open bracket, up to 28 for shortname, 1 close bracket.
    #                         This might not take up the whole 30 characters allotted for it.
    # color reset
    printf "$c_sym%-10s$c_rst $c_amt%13.6f$c_rst $c_diff%+13.6f %+7.2f%%$c_rst $c_ind%1s$c_rst $c_name[%s]$c_rst\n" \
        "$symbol" "$price" "$diff" "$percent" "$marketStateIndicator" "$shortNameT"
done
