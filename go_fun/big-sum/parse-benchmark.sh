#!/bin/bash
# This script will parse the results of a benchmark test and output the key pieces of info.
# Example usage:
# $ make benchmark | tee benchmark-results.txt
# $ parse-benchmark.sh benchmark-results.txt

files=()
while [[ "$#" -gt '0' ]]; do
    case "$1" in
    --help|-h)
        printf 'Usage: parse-benchmark.sh <benchmark results file> [<second file> ...]\n'
        exit 0
        ;;
    --verbose|-v)
        verbose="$1"
        ;;
    *)
        files+=( "$1" )
        ;;
    esac
    shift
done

for f in "${files[@]}"; do
    if [[ ! -f "$f" ]]; then
        printf 'File does not exist: %s\n' "$f"
        exit 1
    fi

    if [[ "$verbose" != "" || "${#files[@]}" -ne '1' ]]; then
        printf '\033[4m%s\033[0m:\n' "$f"
    fi
    grep '^BenchmarkSumFuncs.*[[:space:]]' "$f" \
        | sed -E 's/^BenchmarkSumFuncs\///; s/_([[:digit:]]+)_/ \1 /; s/-[^[:space:]]*([[:space:]])/\1/;' \
        | awk -v verbose="$verbose" -f parse-benchmark.awk
done
