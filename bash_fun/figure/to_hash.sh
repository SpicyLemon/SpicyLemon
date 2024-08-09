#!/bin/bash
# This file contains the to_hash function that converts an amount of nhash into hash.
# This file can be sourced to add the to_hash function to your environment.
# This file can also be executed to run the to_hash function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: to_hash <amount>
#    or: to_hash <amount>nhash
#    or: <stuff> | to_hash -
to_hash () {
    if [[ "$#" -eq '0' ]]; then
        printf 'Usage: to_hash <amount> [<amount 2> ...]\n'
        return 0
    fi
    while [[ "$#" -gt '0' ]]; do
        if [[ "$1" == '-' ]]; then
            sed -E 's/nhash//; s/^/000000000/; s/(.{9})$/.\1/; s/^0+//; s/^\./0./; s/$/ hash/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/;'
        else
            sed -E 's/nhash//; s/^/000000000/; s/(.{9})$/.\1/; s/^0+//; s/^\./0./; s/$/ hash/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/; s/([[:digit:]])([[:digit:]]{3})([,.])/\1,\2\3/;' <<< "$1"
        fi
        shift
    done
}

if [[ "$sourced" != 'YES' ]]; then
    to_hash "$@"
    exit $?
fi
unset sourced

return 0
