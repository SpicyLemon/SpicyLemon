#!/bin/bash
# This file contains the pbsort function that sorts the clipboard contents.
# This file can be sourced to add the pbsort function to your environment.
# This file can also be executed to run the pbsort function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

pbsort () {
    local args verbose ec
    args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf 'Usage: pbsort [<sort flags>] [-v|--verbose]\n'
                return 0
                ;;
            -v|--verbose)
                verbose='YES'
                ;;
            -vv)
                verbose='VERY'
                ;;
            *)
                args+=( "$1" )
                ;;
        esac
        shift
    done
    if [[ "$verbose" == 'VERY' ]]; then
        printf 'Before:\n'
        pbpaste
        printf '\n\n'
        printf 'Command: pbpaste | sort '
        printf '%s ' "${args[@]}"
        printf '| pbcopy\n\n'
    fi
    pbpaste | sort "${args[@]}" | pbcopy
    ec="${PIPESTATUS[1]}${pipestatus[2]}"
    if [[ "$verbose" == 'VERY' ]]; then
        printf 'After:\n'
    fi
    [[ -n "$verbose" ]] && pbpaste
    return "$ec"

}

if [[ "$sourced" != 'YES' ]]; then
    pbsort "$@"
    exit $?
fi
unset sourced

return 0
