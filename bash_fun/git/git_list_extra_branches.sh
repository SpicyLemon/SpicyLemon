#!/bin/bash
# This file contains the git_list_extra_branches function that lists all repos and branches that have more than the default branch.
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
    local ignore_branches show_all cwd repo extra_branches branches default_branch branch
    ignore_branches=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf 'Usage: git_list_extra_branches [-a|--all] [<ignore branch 1> [<ignore branch 2> ...]]\n'
                return 0
                ;;
            -a|--all)
                show_all='YES'
                ;;
            *)
                ignore_branches+=( "$1" )
                ;;
        esac
        shift
    done
    cwd="$( pwd )"
    while IFS= read -r repo; do
        cd "$repo"
        extra_branches=''
        branches="$( git branch )"
        if [[ -z "$show_all" ]]; then
            default_branch="$( git_get_default_branch )"
            [[ -z "$default_branch" ]] && default_branch='master'
            extra_branches="$( sed 's/^[* ][[:space:]]*//; s/[[:space:]]*$//;' <<< "$branches" )"
            for branch in "$default_branch" "${ignore_branches[@]}"; do
                extra_branches="$( grep -vFx "$branch" <<< "$extra_branches" )"
            done
        fi
        if [[ -n "$show_all" || -n "$extra_branches" ]]; then
            printf '\033[1;37m%s\033[0m\n' "$repo"
            sed -E 's/^[*] (.+)$/* '$'\033[32m''\1'$'\033[0m''/; s/^/  /;' <<< "$branches"
            printf '\n'
        fi
    done <<< "$( __git_get_all_repos )"
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
    require_command 'git_get_default_branch' || exit $?
    require_command '__git_get_all_repos' || exit $?
    git_list_extra_branches "$@"
    exit $?
fi
unset sourced

return 0
