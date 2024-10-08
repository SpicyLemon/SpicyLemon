#!/bin/bash
# This file contains the go_resolve_mod function that outputs which go.mod file applies to a file.
# This file can be sourced to add the go_resolve_mod function to your environment.
# This file can also be executed to run the go_resolve_mod function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: go_resolve_mod [<file>]
# If no <file> is provided, . is used.
go_resolve_mod () {
    local f h
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                h=1
                break
                ;;
            *)
                if [[ -n "$f" ]]; then
                    printf 'Unknown args: %s\n' "$*"
                    h=1
                    break
                fi
                f="$1"
                ;;
        esac
        shift
    done
    if [[ -n "$h" ]]; then
        printf 'Usage: go_resolve_mod [<file>]\n'
        return 0
    fi
    if [[ -n "$f" && ! -d "$f" ]]; then
        f="$( dirname "$f" )"
    fi
    while [[ -n "$f" && "$f" != '.' ]]; do
        if [[ -f "$f/go.mod" ]]; then
            printf '%s/go.mod\n' "$f"
            return 0
        fi
        f="$( dirname "$f" )"
    done
    if [[ -f 'go.mod' ]]; then
        printf 'go.mod\n'
        return 0
    fi
    return 1
}

if [[ "$sourced" != 'YES' ]]; then
    go_resolve_mod "$@"
    exit $?
fi
unset sourced

return 0
