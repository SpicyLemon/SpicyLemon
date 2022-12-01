#!/bin/bash
# This file contains the find_dead_links function that finds symlinks to files that don't exist.
# This file can be sourced to add the find_dead_links function to your environment.
# This file can also be executed to run the find_dead_links function without adding it to your environment.
#
# File contents:
#   find_dead_links   --> Find symlinks to files that don't exist.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

find_dead_links  () {
    local d v ec ls l
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf 'Usage: find_dead_links [<dir>]\n'
                return 0
                ;;
            -v|--verbose)
                v=1
                ;;
            *)
                if [[ -n "$d" ]]; then
                    printf 'Unkown arg: [%s].\n' "$1"
                    return 1
                fi
                d="$1"
                ;;
        esac
        shift
    done
    if [[ -z "$d" ]]; then
        d='.'
    fi
    [[ -n "$v" ]] && printf 'Finding dead symlinks under: %s\n' "$d"
    ec=0
    links="$( find "$d" -type l -not -exec test -e {} \; -print )" || ec=$?
    [[ "$ec" -ne '0' ]] && return "$ec"
    while IFS= read -r l; do
        if [[ -n "$l" ]]; then
            ec=1
            printf '%s -> %s\n' "$l" "$( readlink $l )"
        fi
    done <<< "$links"
    [[ -n "$v" && "$ec" -eq '0' ]] && printf 'No dead symlinks found.\n'
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    find_dead_links  "$@"
    exit $?
fi
unset sourced

return 0
