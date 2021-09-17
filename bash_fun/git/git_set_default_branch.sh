#!/bin/bash
# This file contains the git_set_default_branch function that sets the default branch in the local git config.
# This file can be sourced to add the git_set_default_branch function to your environment.
# This file can also be executed to run the git_set_default_branch function without adding it to your environment.
#
# File contents:
#   git_set_default_branch  --> Defines the default branch in the local git config.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_set_default_branch [<default branch>]
# If <default branch> is not provided, the current branch is used.
git_set_default_branch () {
    local default_branch cmd
    if [[ "$1" == '--help' || "$1" == '-h' ]]; then
        printf 'Usage: git_set_default_branch [<default branch>]\n'
        printf 'If <default branch> is not provided, the current branch is used.\n'
        return 0
    fi
    if ! in_git_folder; then
        printf 'git_set_default_branch: Not in a git repo: %s\n' "$( pwd )" >&2
        return 2
    fi
    [[ "$1" == '--' ]] && shift
    if [[ -n "$1" ]]; then
        default_branch="$1"
    else
        if ! command -v "git_branch_name" > /dev/null 2>&1; then
            printf 'git_set_default_branch: Missing required command: git_branch_name.\n' >&2
            return 2
        fi
        default_branch="$( git_branch_name )"
    fi
    git config --local --add spicylemon.defaultbranch "$default_branch"
    return $?
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
    require_command 'git_branch_name' || exit $?

    git_set_default_branch "$@"
    exit $?
fi
unset sourced

return 0
