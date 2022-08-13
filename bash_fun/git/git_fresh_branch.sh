#!/bin/bash
# This file contains the git_fresh_branch function that creates a fresh branch from the default branch.
# This file can be sourced to add the git_fresh_branch function to your environment.
# This file can also be executed to run the git_fresh_branch function without adding it to your environment.
#
# File contents:
#   git_fresh_branch  --> Pulls the default branch and creates a fresh branch from it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_fresh_branch <branch name>
git_fresh_branch () {
    local branch default_branch
    branch="$1"
    if ! in_git_folder; then
        printf 'git_fresh_branch: Not in a git repo.\n' >&2
        return 1
    fi
    if [[ -z "$branch" ]]; then
        printf 'Usage: git_fresh_branch <branch name>\n' >&2
        return 2
    fi
    default_branch="$( git_get_default_branch )"
    if [[ -z "$default_branch" ]]; then
        printf 'git_fresh_branch: No default branch found.\n' >&2
        return 3
    fi
    if [[ "$( git rev-parse --abbrev-ref HEAD )" != "$default_branch" ]]; then
        __git_echo_do git checkout "$default_branch" || return $?
    fi
    __git_echo_do git pull
    __git_echo_do git checkout -b "$branch"
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
    require_command 'git_get_default_branch' || exit $?
    git_fresh_branch "$@"
    exit $?
fi
unset sourced

return 0
