#!/bin/bash
# This file contains the tee_strip_colors function that sends output to both stdout and strips colors before saving to a file.
# This file can be sourced to add the tee_strip_colors function to your environment.
# This file can also be executed to run the tee_strip_colors function without adding it to your environment.
#
# File contents:
#   tee_strip_colors  --> Outputs to stdout and also strips colors and saves to the provided file.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: <stuff> | tee_strip_colors "logfile"
tee_strip_colors () {
    local filename
    filename="$1"
    shift
    if [[ -z "$filename" ]]; then
        printf 'Usage: tee_strip_colors <filename>\n' >&2
        return 1
    fi
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | tee_strip_colors "$filename"
        return $?
    fi
    cat - > >( tee >( strip_colors >> "$filename" ) )
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
    require_command 'strip_colors' || exit $?
    tee_strip_colors "$@"
    exit $?
fi
unset sourced

return 0
