#!/bin/bash
# This file contains the add function that adds a collection of numbers together.
# This file can be sourced to add the add function to your environment.
# This file can also be executed to run the add function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | add
add () {
    if [[ "$#" -eq '0' ]]; then
        add $( cat - )
        return $?
    fi
    local retval
    retval=0
    while [[ "$#" -gt '0' ]]; do
        retval=$(( retval + $1 ))
        shift
    done
    printf '%d' "$retval"
}

if [[ "$sourced" != 'YES' ]]; then
    strip_colors "$@"
    exit $?
fi
unset sourced

return 0