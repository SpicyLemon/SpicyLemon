#!/bin/bash
# This script will run the linter on the files in this dir.

while [[ "$#" -gt '0' ]]; do
    case "$1" in
    --help|-h)
        printf 'Usage: %s [--fix]\n' "$0"
        exit 0
        ;;
    --fix)
        fix='YES'
        ;;
    default)
        printf 'Usage: %s [--fix]\n' "$0"
        printf 'Unknown arg: %q\n' "$1"
        exit 1
        ;;
    esac
    shift
done

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

gofmtcmd=( gofmt -s )
lintcicmd=( golangci-lint run --config "$scriptDir/.golangci.yml" )
if [[ -n "$fix" ]]; then
    gofmtcmd+=( -w )
    lintcicmd+=( --fix )
else
    gofmtcmd+=( -d )
fi

for f in $( find . -type f -name '*.go' ); do
    printf '%s %s\n' "${gofmtcmd[*]}" "$f"
    "${gofmtcmd[@]}" "$f"
    printf '%s %s\n' "${lintcicmd[*]}" "$f"
    "${lintcicmd[@]}" "$f"
done

