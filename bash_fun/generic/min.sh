#!/bin/bash
# This file contains the min function that gets the minimum number from a list of numbers.
# This file can be sourced to add the min function to your environment.
# This file can also be executed to run the min function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | min
#    or: min <val1> [<val2> ...]
#    or: <stuff> | min - <val1> [<val2> ...]
min () {
    {
        if [[ "$#" -eq '0' ]]; then
            cat -
        else
            while [[ "$#" -gt '0' ]]; do
                if [[ "$1" == '-' ]]; then
                    cat -
                else
                    printf ' %s ' "$1"
                fi
                shift
            done
        fi
    } | tr '[:space:]' '\n' | grep . | sort -n | head -n 1
}

if [[ "$sourced" != 'YES' ]]; then
    min "$@"
    exit $?
fi
unset sourced

return 0
