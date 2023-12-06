#!/bin/bash
# This file contains the git_add_next_file function that runs git add on the next unstaged file.
# This file can be sourced to add the git_add_next_file function to your environment.
# This file can also be executed to run the git_add_next_file function without adding it to your environment.
#
# File contents:
#   git_add_next_file  --> Run git add on the next unstaged file.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_add_next_file
git_add_next_file () {
    if ! in_git_folder; then
        printf 'git_add_next_file: Not in a git repo.\n' >&2
        return 1
    fi
    local nf args
    nf="$( git_next_file )"
    if [[ "$?" -ne '0' || -z "$nf" ]]; then
        git status
        printf 'git_add_next_file: Cannot find next file to diff.\n'
        return 1
    fi
    args=( "$@" "$nf" )
    # Print the command and put it in the history before running it.
    printf 'git add %s\n' "${args[*]}"
    history -s git add "${args[@]}"
    git add "${args[@]}"
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
    require_command 'git_next_file' || exit $?
    git_add_next_file "$@"
    exit $?
fi
unset sourced

return 0
