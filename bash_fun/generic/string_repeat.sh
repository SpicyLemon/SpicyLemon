#!/bin/bash
# This file contains the string_repeat function that repeats a string a given number of times.
# This file can be sourced to add the string_repeat function to your environment.
# This file can also be executed to run the string_repeat function without adding it to your environment.
#
# File contents:
#   string_repeat  --> Repeat a string a number of times.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Repeat a string a number of times
# Usage: string_repeat <string> <count>
string_repeat () {
    local string count
    string="$1"
    count="$2"
    if [[ -z "$count" || "$count" =~ [^[:digit:]] ]]; then
        printf 'Usage: string_repeat <string> <count>\n' >&2
        return 1
    fi
    if [[ -n "$string" && "$count" -gt '0' ]]; then
        for i in $( seq "$count" ); do
            printf '%s' "$string"
        done
    fi
}

if [[ "$sourced" != 'YES' ]]; then
    string_repeat "$@"
    exit $?
fi
unset sourced

return 0
