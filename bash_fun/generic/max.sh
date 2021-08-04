#!/bin/bash
# This file contains the max function that gets the maximum number from a list of numbers.
# This file can be sourced to add the max function to your environment.
# This file can also be executed to run the max function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | max
#    or: max <val1> [<val2> ...]
#    or: <stuff> | max - <val1> [<val2> ...]
max () {
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
    } | tr '[:space:]' '\n' | grep . | sort -n -r | head -n 1
}

if [[ "$sourced" != 'YES' ]]; then
    max "$@"
    exit $?
fi
unset sourced

return 0
