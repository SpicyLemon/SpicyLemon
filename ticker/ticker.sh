#!/bin/bash

if [[ "$#" -eq '0' ]]; then
    echo "Usage: ./ticker.sh AAPL MSFT GOOG DOGE-USD"
    exit
fi

SYMBOLS=()
SYMBOLS_TO_GET=()

hr='<hr>'
br='<br>'

while [[ "$#" -gt '0' ]]; do
    # Check for special cases <br> and <hr> allowing for multiple bracket styles and casing.
    # Techinically this will catch stuff like <bR], but this reads better than trying to match stuff,
    # and catching extra stuff like that isn't going to hurt anything.
    if [[ "$1" =~ ^[\[\(\{\<][bB][rR][\]\)\}\>]$ ]]; then
        SYMBOLS+=( "$br" )
    elif [[ "$1" =~ ^[\[\(\{\<][hH][rR][\]\)\}\>]$ ]]; then
        SYMBOLS+=( "$hr" )
    else
        SYMBOLS+=( $1 )
        SYMBOLS_TO_GET+=( $1 )
    fi
    shift
done


# Set this to anything non-empty to cause some debug statements to be printed along the way.
DEBUG=

if ! command -v 'jq' > /dev/null 2>&1; then
    printf 'Missing required command: jq\n' >&2
    jq >&2  # Possibly provides shell/system specific information about the missing command.
    exit $?
fi

if [[ -z "$NO_COLOR" ]]; then
    : "${COLOR_BOLD:=\e[1;37m}"
    : "${COLOR_GREEN:=\e[32m}"
    : "${COLOR_RED:=\e[31m}"
    : "${COLOR_RESET:=\e[0m}"
fi

API_ENDPOINT="https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com"
FIELDS=(symbol shortName marketState regularMarketPrice regularMarketChange regularMarketChangePercent \
  preMarketPrice preMarketChange preMarketChangePercent postMarketPrice postMarketChange postMarketChangePercent)

url_symbols=$( IFS=,; echo "${SYMBOLS_TO_GET[*]}" )
url_fields=$( IFS=,; echo "${FIELDS[*]}" )

api_url="$API_ENDPOINT&fields=$url_fields&symbols=$url_symbols"

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

for symbol in ${SYMBOLS[*]}; do
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
        color=
    elif [[ "$diff" =~ ^- ]]; then
        color=$COLOR_RED
    else
        color=$COLOR_GREEN
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
    #   1-10   '%-10s'   10   Ticker symbol, left justified.
    #   11     ' '       1    Space.
    # bold coloring on
    #   12-24  '%13.6f'  13   Value: 6 whole digits, 1 decimal, 6 fractional digits.
    #                         This will be more than 13 characters for values over 999,999.99999.
    # color reset
    #   25     ' '       1    Space.
    # up(green)/down(red) coloring on
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
    #   51-80  '[%s]'    30   Short name: 1 open bracket, up to 28 for shortname, 1 close bracket.
    #                         This might not take up the whole 30 characters allotted for it.
    printf "%-10s $COLOR_BOLD%13.6f$COLOR_RESET $color%+13.6f %+7.2f%%$COLOR_RESET %1s [%s]\n" \
        "$symbol" "$price" "$diff" "$percent" "$marketStateIndicator" "$shortNameT"
done
