#!/bin/bash
# This file contains the modulo function that gets calculates integer division.
# This file can be sourced to add the modulo function to your environment.
# This file can also be executed to run the modulo function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: modulo <numerator> <denominator>
modulo () {
    printf '%d / %d = %d r %d\n' "$1" "$2" "$(( $1 / $2 ))" "$(( $1 % $2 ))"
}

if [[ "$sourced" != 'YES' ]]; then
    modulo "$@"
    exit $?
fi
unset sourced

return 0
