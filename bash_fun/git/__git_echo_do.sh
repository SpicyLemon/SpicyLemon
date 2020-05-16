#!/bin/bash
# This file contains the __git_echo_do function that outputs a command before executing it.
# This file can be sourced to add the __git_echo_do function to your environment.
# This file can also be executed to run the __git_echo_do function without adding it to your environment.
#
# File contents:
#   __git_echo_do  --> Outputs a command in bright white, then executes it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

__git_echo_do () {
    # If echo_do is available, just use that.
    if command -v echo_do > /dev/null 2>&1; then
        echo_do "$@"
        return $?
    fi
    # Otherwise, do something probably similar.
    local cmd_pieces pieces_for_output cmd_piece retval
    if [[ "$#" -eq '0' || "$@" =~ ^[[:space:]]*$ ]]; then
        printf '__git_echo_do: No command provided.\n' >&2
        return 124
    fi
    if [[ "$#" -eq '1' && ( "$@" =~ [[:space:]\(=] || -z "$( command -v "$@" )" ) ]]; then
        cmd_pieces=( 'eval' "$@" )
        pieces_for_output=( "$@" )
    else
        cmd_pieces=( "$@" )
        pieces_for_output=()
        for cmd_piece in "$@"; do
            if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
                pieces_for_output+=( "\"$( sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' <<< "$cmd_piece" )\"" )
            else
                pieces_for_output+=( "$cmd_piece" )
            fi
        done
    fi

    printf '\033[1;37m%s\033[0m\n' "${pieces_for_output[*]}"
    "${cmd_pieces[@]}"
    retval=$?
    printf '\n'
    return $?
}

if [[ "$sourced" != 'YES' ]]; then
    __git_echo_do "$@"
    exit $?
fi
unset sourced

return 0
