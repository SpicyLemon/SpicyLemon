#!/bin/bash
# This file contains the git_next_file function that parses git status output to find the next unstaged file.
# This file can be sourced to add the git_next_file function to your environment.
# This file can also be executed to run the git_next_file function without adding it to your environment.
#
# File contents:
#   git_next_file  --> Select a branch and switch to it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_next_file
git_next_file () {
    if ! in_git_folder; then
        printf 'git_next_file: Not in a git repo.\n' >&2
        return 1
    fi
    local result
    result="$( git status --porcelain | grep '^.[MU]' | head -n 1 | sed 's/...//' )"
    if [[ -z "$result" ]]; then
        return 1
    fi
    result="$( git rev-parse --show-toplevel )/$result"
    if command -v realpath > /dev/null 2>&1; then
        result="$( realpath --relative-to="$( pwd )" "$result" )"
    fi
    printf '%s' "$result"
    return 0
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
    require_command 'in_git_folder' || exit $?
    git_next_file "$@"
    exit $?
fi
unset sourced

return 0
