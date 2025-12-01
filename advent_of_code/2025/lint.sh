#!/bin/bash
# This script will run the linter on the files in this dir.

files=()
while [[ "$#" -gt '0' ]]; do
    case "$1" in
    --help|-h)
        printf 'Usage: %s [--fix] [<file> [...]]\n' "$0"
        exit 0
        ;;
    --fix)
        fix='YES'
        ;;
    *)
        if [[ -f "$1" ]]; then
            files+=( "$1" )
        elif [[ -d "$1" ]]; then
            files+=( $( find "$1" -type f -name '*.go' | sort ) )
        else
            printf 'Unknown argument: [%s]\n' "$1"
            exit 1
        fi
        ;;
    esac
    shift
done

if [[ "${#files[@]}" -eq '0' ]]; then
    files=( $( find . -type f -name '*.go' | sort ) )
fi

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

gofmtcmd=( gofmt -s )
lintcicmd=( golangci-lint run --config "$scriptDir/.golangci.yml" )
if [[ -n "$fix" ]]; then
    gofmtcmd+=( -w )
    lintcicmd+=( --fix )
else
    gofmtcmd+=( -d )
fi

for f in "${files[@]}"; do
    printf '%s gofmt ... ' "$f"
    "${gofmtcmd[@]}" "$f"
    printf 'golangci-lint ... '
    "${lintcicmd[@]}" "$f"
    printf 'Done\n'
done

