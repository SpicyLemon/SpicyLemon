#!/bin/bash
# This file contains the to_date function that converts epochs in milliseconds to a date string.
# This file can be sourced to add the to_date function to your environment.
# This file can also be executed to run the to_date function without adding it to your environment.
#
# File contents:
#   to_date  --> Converts an epoch as milliseconds into a date.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Convert an epoch as milliseconds into a date and time.
# Usage: to_date <epoch in milliseconds>
#  or    to_date now
to_date () {
    local input pieces epoch_ms ms_fractions ms s_fractions epoch_s
    input="$1"
    if [[ -z "$input" || "$input" == '-h' || "$input" == '--help' ]]; then
        printf 'Usage: to_date <epoch in milliseconds>\n'
        return 0
    fi
    if [[ "$input" == 'now' ]]; then
        date '+%F %T %z (%Z) %A'
        return 0
    fi
    # Split out the input into milliseconds and fractional milliseconds
    if [[ "$input" =~ ^[[:digit:]]+(\.[[:digit:]]+)?$ ]]; then
        pieces=( $( tr '.' ' ' <<< "$input" ) )
        if [[ -n "${pieces[0]}" ]]; then
            epoch_ms="${pieces[0]}"
            ms_fractions="${pieces[1]}"
        else
            epoch_ms="${pieces[1]}"
            ms_fractions="${pieces[2]}"
        fi
    else
        printf "Invalid input: [%s].\n" "$input" >&2
        return 1
    fi
    ms="$( printf '%s' "$epoch_ms" | tail -c 3 )"
    s_fractions="$( sed -E 's/0+$//' <<< "${ms}${ms_fractions}")"
    if [[ -n "$s_fractions" ]]; then
        s_fractions=".$s_fractions"
    fi
    epoch_s="$( sed -E 's/...$//' <<< "$epoch_ms" )"
    date -r "$epoch_s" "+%F %T${s_fractions} %z (%Z) %A"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    to_date "$@"
    exit $?
fi
unset sourced

return 0
