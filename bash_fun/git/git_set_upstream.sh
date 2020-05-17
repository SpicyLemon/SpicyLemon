#!/bin/bash
# This file contains the git_set_upstream function that sets the upstream for the repo and branch you're in.
# This file can be sourced to add the git_set_upstream function to your environment.
# This file can also be executed to run the git_set_upstream function without adding it to your environment.
#
# File contents:
#   git_set_upstream  --> Sets the upstream appropriately for the repo and branch you're in.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_set_upstream
git_set_upstream () {
    if ! in_git_folder; then
        printf 'git_set_upstream: Not in a git repo.\n' >&2
        return 1
    fi
    local cur_branch
    cur_branch="$( git_branch_name )"
    __git_echo_do git branch "--set-upstream-to=origin/$cur_branch" "$cur_branch" \
        && __git_echo_do git pull
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
    require_command '__git_echo_do' || exit $?
    require_command 'git_branch_name' || exit $?
    git_set_upstream "$@"
    exit $?
fi
unset sourced

return 0
