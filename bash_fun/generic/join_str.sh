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

# Uses the provided delimiter to join all the provided arguments.
# Usage: join_str <delimiter> [<arg1> [<arg2>... ]]
join_str () {
    local d retval
    d="$1"
    shift
    printf %s "$1"
    shift
    while [[ "$#" -gt '0' ]]; do
        printf '%s%s' "$d" "$1"
        shift
    done
}

if [[ "$sourced" != 'YES' ]]; then
    join_str "$@"
    exit $?
fi
unset sourced

return 0
