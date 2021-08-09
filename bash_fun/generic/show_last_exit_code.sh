#!/bin/bash
# This file contains the show_last_exit_code function that outputs the last exit code received.
# This file is meant to be sourced to add the show_last_exit_code function to your environment.
#
# File contents:
#   show_last_exit_code  --> Displays the last exit code received.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

show_last_exit_code () {
    local exit_code=$?
    if [[ "$exit_code" -ne '0' ]]; then
        # White text with red background, skull.
        printf '\033[97;41m \xF0\x9F\x92\x80 %3d \033[0m' "$exit_code"
    else
        # White text with green background, gold star.
        printf '\033[97;42m \xE2\xAD\x90 %3d \033[0m' "$exit_code"
    fi
    return $exit_code
}

return 0
