#!/bin/bash

if [[ -n "$1" ]]; then
    input="$1"
else
    input="example.input"
fi

printf 'input: [%s].\n' "$input"

{
    for group in $( sed 's/^[[:space:]]*$/~/' "$input" | tr -d '\n' | tr '~' '\n' ); do
        sed 's/\(.\)/\1\'$'\n/g' <<< "$group" | sort | uniq -c | grep [^[:space:][:digit:]] | sed 's/[^[:alpha:]]//g' | tr -d '\n';
    done
} | wc -c | sed 's/[^[:digit:]]//g'
