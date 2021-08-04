#!/bin/bash
# This file contains the b642h function that converts base64 values to hex.
# This file can be sourced to add the b642h function to your environment.
# This file can also be executed to run the b642h function without adding it to your environment.
#
# File contents:
#   b642h  --> Converts base64 values to hex.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

b642h () {
    if [[ "$#" -eq '1' && ( "$1" == '-h' || "$1" == '--help' ) ]]; then
        cat << EOF
Converts base64 values to hex.

Usage: b642h <val1> [<val2>...]
   or: <stuff> | b642h

EOF
        return 0
    fi
    local vs ec v
    if [[ "$#" -gt '0' ]]; then
        vs=( $( tr ' ' '\n' <<< $* ) )
    else
        vs=( $( cat - | tr ' ' '\n' ) )
    fi
    ec=0
    for v in "${vs[@]}"; do
        base64 -d <<< "$v" | xxd -p | tr -d '[:space:]' || ec=$?
        printf '\n'
    done
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    b642h "$@"
    exit $?
fi
unset sourced

return 0
