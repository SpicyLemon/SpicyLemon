#!/bin/bash
# This file contains the join_str function that joins multiple values using a delimiter.
# This file can be sourced to add the join_str function to your environment.
# This file can also be executed to run the join_str function without adding it to your environment.
#
# File contents:
#   join_str  --> Joins a list of parameters using a delimiter.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

join_str () {
    local d
    if [[ "$#" -eq '0' || "$1" == '-h' || "$1" == '--help' ]]; then
        printf 'Usage: join_str <delimiter> [-|[<val1>] [<val2>... ]]\n'
        return 0
    fi
    d="$1"
    shift

    # If multiple values are provided, or if only one is provided that isn't '-', join all the args provided.
    if [[ "$#" -ge '2' || ( "$#" -eq '1' && "$1" != '-' ) ]]; then
        printf %s "$1"
        shift
        while [[ "$#" -gt '0' ]]; do
            printf '%s%s' "$d" "$1"
            shift
        done
        return 0
    fi

    # If stdin isn't interactive (i.e. stuff is being piped in), join each line being piped in.
    if [[ ! -t 0 ]]; then
        local line inc_d
        while IFS= read -r line; do
            if [[ -n "$inc_d" ]]; then
                printf '%s' "$d"
            else
                inc_d='YES'
            fi
            printf '%s' "$line"
        done
        return 0
    fi

    # Not sure what to do.
    return 1
}

if [[ "$sourced" != 'YES' ]]; then
    join_str "$@"
    exit $?
fi
unset sourced

return 0
