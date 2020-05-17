#!/bin/bash
# This file contains the git_pull_master function that switches to the master branch, pulls it, switches back and merges it into your branch.
# This file can be sourced to add the git_pull_master function to your environment.
# This file can also be executed to run the git_pull_master function without adding it to your environment.
#
# File contents:
#   git_pull_master  --> Pull master and merge it into your branch.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_pull_master
git_pull_master () {
    if ! in_git_folder; then
        printf 'git_pull_master: Not in a git repo.\n' >&2
        return 1
    fi
    local cur_branch exit_code
    cur_branch="$( git_branch_name )"
    if [[ "$cur_branch" == 'master' ]]; then
        __git_echo_do git pull
        exit_code=$?
    else
        __git_echo_do git checkout master \
            && __git_echo_do git pull \
            && __git_echo_do git checkout "$cur_branch" \
            && __git_echo_do git merge master
        exit_code=$?
        __git_echo_do git status
    fi
    return $exit_code
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
    git_pull_master "$@"
    exit $?
fi
unset sourced

return 0
