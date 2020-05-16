#!/bin/bash
# This file contains the git_clean_repo function that tries to clean up the current repo, just like new.
# This file can be sourced to add the git_clean_repo function to your environment.
# This file can also be executed to run the git_clean_repo function without adding it to your environment.
#
# File contents:
#   git_clean_repo  --> Takes several actions to help you clean up a git repo.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_clean_repo
git_clean_repo () {
    if ! in_git_folder; then
        printf 'git_clean_repo: Not in a git repo.\n' >&2
        return 1
    fi
    # Check out master
    __git_echo_do git checkout master
    # Delete any branches if desired.
    __git_echo_do git_delete_branches
    # Do git clean: -f -> force to delete untracked files,
    #               -d -> recurse into untracked directories,
    #               -x -> ignore standard ignore rules
    #               -e .idea -> but leave the .idea directory alone
    __git_echo_do git clean -fdx -e .idea
    # Remove any stale remote tracking branches
    printf "\033[1;37mgit branch -r | grep -v 'HEAD' | xargs -L 1 git branch -rD\033[0m\n"
    git branch -r | grep -v 'HEAD' | xargs -L 1 git branch -rD
    # And get the most recent info
    __git_echo_do git fetch
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
    require_command 'git_delete_branches' || exit $?
    git_clean_repo "$@"
    exit $?
fi
unset sourced

return 0
