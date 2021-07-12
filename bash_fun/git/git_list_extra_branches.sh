#!/bin/bash
# This file contains the git_list_extra_branches function that creates a fresh branch from master.
# This file can be sourced to add the git_list_extra_branches function to your environment.
# This file can also be executed to run the git_list_extra_branches function without adding it to your environment.
#
# File contents:
#   git_list_extra_branches  --> List all the local extra branches in all your local repos.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_list_extra_branches () {
    local show_all cwd repos repo branches
    if [[ "$1" == '-a' || "$1" == '--all' ]]; then
        show_all='YES'
    fi
    cwd="$( pwd )"
    repos=( $( __git_get_all_repos ) )
    for repo in "${repos[@]}"; do
        cd "$repo"
        branches="$( git branch )"
        if [[ -n "$show_all" || -n "$( grep -E -v '^[* ] (master|main|develop)$' <<< "$branches" )" ]]; then
            printf '\033[1;37m%s\033[0m\n' "$repo"
            sed -E 's/^[*] (.+)$/* '$'\033[32m''\1'$'\033[0m''/; s/^/  /;' <<< "$branches"
            printf '\n'
        fi
    done
    cd "$cwd"
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
    require_command '__git_get_all_repos' || exit $?
    git_list_extra_branches "$@"
    exit $?
fi
unset sourced

return 0
