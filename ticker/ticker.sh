#!/bin/bash

if [[ "$#" -eq '0' ]]; then
    echo "Usage: ./ticker.sh AAPL MSFT GOOG DOGE-USD"
    exit
fi

SYMBOLS=($@)

# Set this to anything non-empty to cause some debug statements to be printed along the way.
DEBUG=

if ! $(type jq > /dev/null 2>&1); then
    echo "'jq' is not in the PATH. (See: https://stedolan.github.io/jq/)"
    exit 1
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

url_symbols=$( IFS=,; echo "${SYMBOLS[*]}" )
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

for symbol in ${SYMBOLS[*]}; do
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

    printf "%-10s $COLOR_BOLD%13.6f$COLOR_RESET $color%+13.6f %+7.2f%%$COLOR_RESET %1s [%s]\n" \
        "$symbol" "$price" "$diff" "$percent" "$marketStateIndicator" "$shortName"
done
