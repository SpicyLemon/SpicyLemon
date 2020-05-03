#!/bin/bash
# This file contains the print_args function that outputs arguments that are passed in.
# This file can be sourced to add the print_args function to your environment.
# This file can also be executed to run the print_args function without adding it to your environment.
#
# File contents:
#   print_args  --> Outputs all parameters received.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

print_args () {
    local i
    if [[ "$#" -eq '0' ]]; then
        printf 'No arguments provided.\n' >&2
        return 1
    fi
    printf 'Arguments received:\n'
    i=0
    while [[ "$#" -gt '0' ]]; do
        i=$(( i + 1 ))
        printf '%2d: [%s]\n' "$i" "$1"
        shift
    done
}

if [[ "$sourced" != 'YES' ]]; then
    print_args "$@"
    exit $?
fi
unset sourced

return 0
