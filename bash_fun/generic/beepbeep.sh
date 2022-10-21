#!/bin/bash
# This file contains the beepbeep function that outputs two bell chars .3 seconds apart.
# This file can be sourced to add the beepbeep function to your environment.
# This file can also be executed to run the beepbeep function without adding it to your environment.
#
# File contents:
#   beepbeep   --> beeps twice and returns with previous exit code.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

beepbeep  () {
    # Presever last exit code
    local ec=$?
    printf '\a'
    sleep .3
    printf '\a'
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    beepbeep  "$@"
    exit $?
fi
unset sourced

return 0
