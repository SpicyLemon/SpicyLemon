#!/bin/bash
# This file contains the pvarn function that prints variables with given names.
# This file can be sourced to add the pvarn function to your environment.
# This file can also be executed to run the pvarn function without adding it to your environment.
#
# File contents:
#   pvarn   --> Prints variables with the given names.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

pvarn  () {
    if [[ "$#" -eq '0' ]]; then
        printf 'Usage: pvarn <var name 1> [<var name 2> ...]\n'
        return 1
    fi
    local use_p val
    # With ${(P)1}, Bash will output -bash: ${(P)1}: bad substitution
    # But it's not part of command output, so you can't redirect it directly.
    # So I make a subshell to try it, so I can supress that error message.
    # If it fails, it will exit with code 1, resulting in use_p not being set.
    # In zsh, it'll work and exit with code 0, and use_p will be set.
    ( val="${(P)1}"; ) > /dev/null 2>&1 && use_p='YES'
    while [[ "$#" -gt '0' ]]; do
        if [[ -n "$use_p" ]]; then
            val="${(P)1}"
        else
            val="${!1}"
        fi
        printf '%s: [%s]\n' "$1" "$val"
        if [[ "$1" =~ PATH && "$val" =~ : ]]; then
            ws="$( printf '%s: ' "$1" | tr -C ' ' ' ' )"
            tr ':' '\n' <<< "$val" | sed "s/^/$ws/"
        fi
        shift
    done
}

if [[ "$sourced" != 'YES' ]]; then
    pvarn  "$@"
    exit $?
fi
unset sourced

return 0
