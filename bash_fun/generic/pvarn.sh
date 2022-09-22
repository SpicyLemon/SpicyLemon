#!/bin/bash
# This file contains the pvarn function that prints variables with given names.
# This file can be sourced to add the pvarn function to your environment.
# This file can also be executed to run the pvarn function without adding it to your environment.
#
# File contents:
#   pvarn   --> Prints variables with the given names.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

pvarn  () {
    if [[ "$#" -eq '0' ]]; then
        printf 'Usage: pvarn <var name 1> [<var name 2> ...]\n'
        return 1
    fi
    while [[ "$#" -gt '0' ]]; do
        printf '%s: [%s]\n' "$1" "${!1}"
        if [[ "$1" =~ PATH && "${!1}" =~ : ]]; then
            ws="$( printf '%s: ' "$1" | tr -C ' ' ' ' )"
            tr ':' '\n' <<< "${!1}" | sed "s/^/$ws/"
        fi
        shift
    done
}

if [[ "$sourced" != 'YES' ]]; then
    pvarn  "$@"
    exit $?
fi
unset sourced

return 0
