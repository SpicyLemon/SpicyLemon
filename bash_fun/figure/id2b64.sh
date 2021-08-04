#!/bin/bash
# This file contains the id2b64 function that converts a hex values into base64.
# This file can be sourced to add the id2b64 function to your environment.
# This file can also be executed to run the id2b64 function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: id2b64 <hex digits>
#    or: <stuff> | id2b64
# Example: id2b64 00 03611a53-afe4-43a7-a855-5b629b331cab
id2b64 () {
    if [[ "$#" -gt '0' ]]; then
        id2b64 <<< "$*"
        return $?
    fi
    sed -E 's/[^[:xdigit:]]//g;' | xxd -r -p | base64
}

if [[ "$sourced" != 'YES' ]]; then
    id2b64 "$@"
    exit $?
fi
unset sourced

return 0
