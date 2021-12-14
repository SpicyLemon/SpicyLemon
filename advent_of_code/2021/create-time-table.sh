#!/bin/bash

# Usage: create-time-table.sh [--no-rebuild]

# Make sure we're in the right directory (the one containing this script).
curDir="$( pwd )"
scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
if [[ "$curDir" != "$scriptDir" ]]; then
    cd "$scriptDir"
fi

source ../../bash_fun/generic/re_line.sh

if [[ -e build ]]; then
    mkdir build
elif [[ ! -d build ]]; then
    printf 'build is not a directory.\n' 2>&1
    exit 1
fi

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help|hep) printf 'Usage: create-time-table.sh [--no-rebuild]\n';;
        --no-rebuild) no_rebuild='YES';;
        *)
            printf 'Unknown argument: [%s]' "$1"
            exit 1
            ;;
    esac
    shift
done

set -v
# Get a list of all the days
all_days=( $( find . -type d -maxdepth 1 -mindepth 1 -name 'day-*' | sed 's/..//' | sort ) )

printf 'no_rebuild = [%s]\n' "$no_rebuild"
if [[ -z "$no_rebuild" ]]; then
    # Build all the days.
    for d in "${all_days[@]}"; do go build -o build "$d/$d.go"; done
fi

# Run timing an all days using go run and then pre-compiled.
for d in "${all_days[@]}"; do
    printf '\ntime %s\n' "$d"
    time go run "$d/$d.go" "$d/actual.input" 2>&1
done > build/all-go-run.txt 2>&1
for d in "${all_days[@]}"; do
    printf '\ntime %s\n' "$d"
    time "build/$d" "$d/actual.input" 2>&1
done > build/all-compiled-run.txt 2>&1

# Extract timings for each
grep -E '^(time|real|user|sys)' build/all-go-run.txt \
    | sed -E 's/^time.*(day-[[:digit:]][[:digit:]][ab]).*$/\1/; s/^(real|user|sys)[[:space:]]*//;' \
    | { set +v; re_line -n 4 -d ' ' -; set -v; } \
    > build/all-go-run-times.txt
grep -E '^(time|real|user|sys)' build/all-compiled-run.txt \
    | sed -E 's/^time.*(day-[[:digit:]][[:digit:]][ab]).*$/\1/; s/^(real|user|sys)[[:space:]]*//;' \
    | { set +v; re_line -n 4 -d ' ' -; set -v; } \
    > build/all-compiled-run-times.txt
{
    printf ' ~compiled~ ~ ~go run~ ~\n'
    printf 'day~real~user~sys~real~user~sys\n'
    for d in "${all_days[@]}"; do
        set +v
        printf '%s %s\n' "$( grep -F "$d" build/all-compiled-run-times.txt )" "$( grep -F "$d" build/all-go-run-times.txt | cut -f 2- -d ' ' )" \
            | tr ' ' '~'
        set -v
    done
} | column -s '~' -t > build/all-times.txt
