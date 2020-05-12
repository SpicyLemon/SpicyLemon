#!/bin/bash
# This file contains the escape_escapes function that escapes escape characters.
# This file can be sourced to add the escape_escapes function to your environment.
# This file can also be executed to run the escape_escapes function without adding it to your environment.
#
# File contents:
#   escape_escapes  --> Escapes any escape characters in a stream.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | escape_escapes
escape_escapes () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | escape_escapes
        return $?
    fi
    sed -E "s/$( echo -e "\033" )/\\\033/g"
}

if [[ "$sourced" != 'YES' ]]; then
    escape_escapes "$@"
    exit $?
fi
unset sourced

return 0
