#!/bin/bash
# This file contains the git_master_pull_all function that pulls the master branch on all of your repos.
# This file can be sourced to add the git_master_pull_all function to your environment.
# This file can also be executed to run the git_master_pull_all function without adding it to your environment.
#
# File contents:
#   git_master_pull_all  --> Goes through all your repos and pulls master.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_master_pull_all
git_master_pull_all () {
    local repos repo repo_count cwd repo_index successful_repos failed_repos repo_failed default_branch cur_branch
    repos=()
    while IFS= read -r repo; do
        repos+=( "$repo" )
    done <<< "$( __git_get_all_repos )"
    repo_count="${#repos[@]}"
    if [[ "$repo_count" -le '0' ]]; then
        printf 'No repos found.\n' >&2
        return 1
    fi
    cwd="$( pwd )"
    repo_index=0
    successful_repos=()
    failed_repos=()
    for repo in "${repos[@]}"; do
        repo_index=$(( repo_index + 1 ))
        repo_failed=''
        cur_branch=''
        printf '\033[1;36m%d of %d\033[0m - \033[1;33m%s\033[0m\n' "$repo_index" "$repo_count" "$repo"
        __git_echo_do cd "$repo" || repo_failed='YES'
        if [[ -z "$repo_failed" ]]; then
            default_branch="$( git_get_default_branch )"
            [[ -z "$default_branch" ]] && default_branch='master'
            cur_branch="$( git rev-parse --abbrev-ref HEAD )"
            if [[ "$cur_branch" != "$default_branch" ]]; then
                __git_echo_do git checkout "$default_branch" || repo_failed='YES'
            fi
        fi
        if [[ -z "$repo_failed" ]]; then
            __git_echo_do git pull || repo_failed='YES'
        fi
        if [[ -z "$repo_failed" && "$cur_branch" != "$( git rev-parse --abbrev-ref HEAD )" ]]; then
            __git_echo_do git checkout "$cur_branch" || repo_failed='YES'
        fi
        if [[ -n "$repo_failed" ]]; then
            printf '\033[1;97;41m An error occurred \033[0m - \033[1;31mSee above\033[0m\n'
            failed_repos+=( "$repo" )
        else
            successful_repos+=( "$repo" )
        fi
    done
    if [[ "$cwd" != "$( pwd )" ]]; then
        __git_echo_do cd "$cwd"
    fi
    if [[ "${#successful_repos[@]}" -gt '0' ]]; then
        printf '\033[1;32m%d repo%s successfully updated:\033[0m\n' "${#successful_repos[@]}" "$( [[ "${#successful_repos[@]}" -ne '1' ]] && printf 's' )"
        printf '    \033[32m%s\033[0m\n' "${successful_repos[@]}"
        printf '\n'
    fi
    if [[ "${#failed_repos[@]}" -gt '0' ]]; then
        printf '\033[1;31m%d repo%s had problems:\033[0m\n' "${#failed_repos[@]}" "$( [[ "${#failed_repos[@]}" -ne '1' ]] && printf 's' )"
        printf '    \033[31m%s\033[0m\n' "${failed_repos[@]}"
        printf '\n'
    fi
    return "${#failed_repos[@]}"
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
    require_command '__git_echo_do' || exit $?
    require_command 'git_get_default_branch' || exit $?
    git_master_pull_all "$@"
    exit $?
fi
unset sourced

return 0
