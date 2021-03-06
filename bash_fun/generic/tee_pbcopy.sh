#!/bin/bash
# This file contains the tee_pbcopy function that sends output to both stdout and the clipboard.
# This file can be sourced to add the tee_pbcopy function to your environment.
# This file can also be executed to run the tee_pbcopy function without adding it to your environment.
#
# File contents:
#   tee_pbcopy  --> Outputs to stdout as well as copy it to the clipboard.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <do stuff> | tee_pbcopy
tee_pbcopy () {
    local do_not_run
    for req_cmd in 'strip_final_newline' 'pbcopy'; do
        if ! command -v "$req_cmd" > /dev/null 2>&1; then
            do_not_run='yes'
            printf 'Missing required command: %s\n' "$req_cmd" >&2
            "$req_cmd"
        fi
    done
    if [[ -n "$do_not_run" ]]; then
        return 1
    fi
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | tee_pbcopy
        exit $?
    fi
    tee >( strip_final_newline | pbcopy )
}

if [[ "$sourced" != 'YES' ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    require_command () {
        local cmd cmd_fn
        cmd="$1"
        if ! command -v "$cmd" > /dev/null 2>&1; then
            cmd_fn="$where_i_am/$cmd.sh"
            if [[ -f "$cmd_fn" ]]; then
                source "$cmd_fn"
                if [[ "$?" -ne '0' ]] || ! command -v "$cmd" > /dev/null 2>&1; then
                    ( printf 'This script relies on the [%s] function.\n' "$cmd"
                      printf 'The file [%s] was found and sourced, but there was a problem loading the [%s] function.\n' "$cmd_fn" "$cmd" ) >&2
                    return 1
                fi
            else
                ( printf 'This script relies on the [%s] function.\n' "$cmd"
                  printf 'The file [%s] was looked for, but not found.\n' "$cmd_fn" ) >&2
                return 1
            fi
        fi
    }
    require_command 'strip_final_newline' || exit $?
    tee_pbcopy "$@"
    exit $?
fi
unset sourced

return 0
