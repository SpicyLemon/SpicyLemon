#!/bin/bash
# This file contains the git_change_branch function that uses fzf to let you select a local branch to switch to.
# This file can be sourced to add the git_change_branch function to your environment.
# This file can also be executed to run the git_change_branch function without adding it to your environment.
#
# File contents:
#   git_change_branch  --> Select a branch and switch to it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_change_branch
git_change_branch () {
    if ! in_git_folder; then
        printf 'git_change_branch: Not in a git repo.\n' >&2
        return 1
    fi
    local selection query
    if [[ -n "$1" ]]; then
        query=( --query "$1" --select-1 )
    else
        query=()
    fi
    selection="$( git branch | fzf +m --cycle --header='Select the branch to change to and press enter (or esc to cancel).' "${query[@]}" | sed -E 's/^[* ]+//' )"
    [[ -n "$selection" ]] && git checkout "$selection"
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
    git_change_branch "$@"
    exit $?
fi
unset sourced

return 0
