#!/bin/bash
# This file contains the strip_final_newline function that removes any final newline from input.
# This file can be sourced to add the strip_final_newline function to your environment.
# This file can also be executed to run the strip_final_newline function without adding it to your environment.
#
# File contents:
#   strip_final_newline  --> Strips the final newline character from input. Only the last line is changed.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <do stuff> | strip_final_newline
#   or   strip_final_newline <input>
strip_final_newline () {
    if [[ -n "$1" ]]; then
        echo -E "$1" | strip_final_newline
        return $?
    fi
    awk ' { if(p) print(l); l=$0; p=1; } END { printf("%s", l); } '
}

if [[ "$sourced" != 'YES' ]]; then
    strip_final_newline "$@"
    exit $?
fi
unset sourced

return 0
