#!/bin/bash
# This file contains the strip_colors function that removes color escape squences from input.
# This file can be sourced to add the strip_colors function to your environment.
# This file can also be executed to run the strip_colors function without adding it to your environment.
#
# File contents:
#   strip_colors  --> Strips the color stuff from a stream.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | strip_colors
strip_colors () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | strip_colors
        return $?
    fi
    sed -E "s/$( echo -e "\033" )\[[[:digit:]]+(;[[:digit:]]+)*m//g"
}

if [[ "$sourced" != 'YES' ]]; then
    strip_colors "$@"
    exit $?
fi
unset sourced

return 0
