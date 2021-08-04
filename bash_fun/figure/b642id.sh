#!/bin/bash
# This file contains the b642id function that converts a base64 string into a Metadata Address ID parts.
# This file can be sourced to add the b642id function to your environment.
# This file can also be executed to run the b642id function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: b642id <base64> [<base64 2> ...]
#    or: <stuff> | b642id
# Example: b642id "CANhGlOv5EOnqFVbYpszHKsDYRpTr/RDp6hVW2KbMxyv"
b642id () {
    local b64s b64
    if [[ "$#" -eq '0' ]]; then
        b64s=( $( cat - | tr ' ' '\n' ) )
    else
        b64s=( $( tr ' ' '\n' <<< $* ) )
    fi
    for b64 in ${b64s[@]}; do
        if [[ "$b64" =~ ^[[:alnum:]+/]+=*$ ]]; then
            printf '%s => ' "$b64"
            base64 -d <<< "$b64" \
                | od -t x1 -An \
                | tr -d '[:space:]' \
                | sed -E 's/^([[:xdigit:]]{2})/\1 /;
                          s/([[:xdigit:]]{8})([[:xdigit:]]{4})([[:xdigit:]]{4})([[:xdigit:]]{4})([[:xdigit:]]{12})/ \1-\2-\3-\4-\5/g;' \
                | awk '{ if ($1 == "00") $1 = $1" (scope)";
                    else if ($1 == "01") $1 = $1" (session)";
                    else if ($1 == "02") $1 = $1" (record)";
                    else if ($1 == "03") $1 = $1" (contract spec)";
                    else if ($1 == "04") $1 = $1" (scope spec)";
                    else if ($1 == "05") $1 = $1" (record spec)";
                    else                 $1 = $1" (UNKNOWN)";
                        print; }'
        fi
    done
}

if [[ "$sourced" != 'YES' ]]; then
    b642id "$@"
    exit $?
fi
unset sourced

return 0
