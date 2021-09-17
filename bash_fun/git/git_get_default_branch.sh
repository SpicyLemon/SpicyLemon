#!/bin/bash
# This file contains the git_get_default_branch function that outputs the default branch for a repo.
# This file can be sourced to add the git_get_default_branch function to your environment.
# This file can also be executed to run the git_get_default_branch function without adding it to your environment.
#
# File contents:
#   git_get_default_branch  --> Prints the default branch for a repo.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_get_default_branch [<repo dir>]
# If <repo dir> is not provided, the current directory is used.
git_get_default_branch () {
    local cwd default_branch branch_order branches branch remote
    if [[ "$1" == '--help' || "$1" == '-h' ]]; then
        printf 'Usage: git_get_default_branch [<repo dir>]\n'
        printf 'If <repo dir> is not provided, the current directory is used.'
        return 0
    fi
    [[ "$1" == '--' ]] && shift
    if [[ -n "$*" ]]; then
        if [[ -d "$*" ]]; then
            cwd="$( pwd )"
            cd "$*"
        else
            printf 'git_get_default_branch: directory not found: %s\n' "$1" >&2
            return 2
        fi
    fi
    if ! in_git_folder; then
        printf 'Not a git repo: %s\n' "$( pwd )" >&2
        if [[ -n "$cwd" ]]; then
            cd "$cwd"
        fi
        return 2
    fi
    # First, check the local git config for a defaultbranch value.
    default_branch="$( git config --local spicylemon.defaultbranch 2> /dev/null )"
    # If there's nothing there, do a git branch and look for the branches in this order: main master
    if [[ -z "$default_branch" ]]; then
        branch_order=( main master )
        branches="$( git branch | sed 's/^[* ] //' )"
        for branch in "${branch_order[@]}"; do
            if grep -qFx "$branch" <<< "$branches"; then
                default_branch="$branch"
                break
            fi
        done
    fi
    # If there's STILL nothing there, check for those branches on each remote.
    if [[ -z "$default_branch" ]]; then
        for branch in "${branch_order[@]}"; do
            for remote in "$( git remote )"; do
                if git ls-remote --exit-code --heads "$remote" "$branch" > /dev/null 2>&1; then
                    default_branch="$branch"
                    break 2
                fi
            done
        done
    fi
    if [[ -n "$cwd" ]]; then
        cd "$cwd"
    fi
    if [[ -z "$default_branch" ]]; then
        return 1
    fi
    printf '%s\n' "$default_branch"
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
    git_get_default_branch "$@"
    exit $?
fi
unset sourced

return 0
