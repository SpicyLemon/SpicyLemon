#!/bin/bash
# This file contains the h2b64 function that converts hex values to base64.
# This file can be sourced to add the h2b64 function to your environment.
# This file can also be executed to run the h2b64 function without adding it to your environment.
#
# File contents:
#   h2b64  --> Converts hex values to base64.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

h2b64 () {
    if [[ "$#" -eq '1' && ( "$1" == '-h' || "$1" == '--help' ) ]]; then
        cat << EOF
Converts hex values to base64.

Usage: h2b64 <val1> [<val2>...]
   or: <stuff> | h2b64

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
        xxd -r -p <<< "$v" | base64 || ec=$?
    done
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    h2b64 "$@"
    exit $?
fi
unset sourced

return 0
