#!/bin/bash
# This file contains the timealert function that uses time to time a commmand and beeps twice when done.
# This file can be sourced to add the timealert function to your environment.
# This file can also be executed to run the timealert function without adding it to your environment.
#
# File contents:
#   timealert   --> Uses time to time a command and beeps twice when done.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

timealert  () {
    local ec
    time "$@"
    ec=$?
    printf '\a'
    sleep .3
    printf '\a'
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    timealert  "$@"
    exit $?
fi
unset sourced

return 0
