#!/bin/bash
# This file contains the __git_get_all_repos function that outputs all known git repo folders on the system.
# This file can be sourced to add the __git_get_all_repos function to your environment.
# This file can also be executed to run the __git_get_all_repos function without adding it to your environment.
#
# File contents:
#   __git_get_all_repos  --> Gets a list of all the known git repo folders.
#
# In order to find the desired folders, this function will look at the following environment variables:
#   GIT_REPO_DIR, GITLAB_REPO_DIR, GITHUB_REPO_DIR
# For each of those variables that are defined, and contain the path to a directory, that dirctory will be looked in.
# Each direct sub-directory will be examined to see if it as git repo. If so, it will be part of the output.
# In order to use this file as an executable, those environment variables must also be exported.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

__git_get_all_repos () {
    local base_dirs repos cwd base_dir repo
    base_dirs=()
    if [[ -n "$GIT_REPO_DIR" && -d "$GIT_REPO_DIR" ]]; then
        base_dirs+=( "$GIT_REPO_DIR" )
    fi
    if [[ -n "$GITLAB_REPO_DIR" && -d "$GITLAB_REPO_DIR" ]] && ! printf '%s\n' "${base_dirs[@]}" | grep -qFx "$GITLAB_REPO_DIR"; then
        base_dirs+=( "$GITLAB_REPO_DIR" )
    fi
    if [[ -n "$GITHUB_REPO_DIR" && -d "$GITHUB_REPO_DIR" ]] && ! printf '%s\n' "${base_dirs[@]}" | grep -qFx "$GITHUB_REPO_DIR"; then
        base_dirs+=( "$GITHUB_REPO_DIR" )
    fi
    repos=()
    if [[ "${#base_dirs[@]}" -gt '0' ]]; then
        cwd="$( pwd )"
        for base_dir in "${base_dirs[@]}"; do
            for repo in $( ls -d $base_dir/*/ ); do
                if [[ -d "$repo" ]]; then
                    cd "$repo"
                    if [[ "$?" -eq '0' ]] && in_git_folder; then
                        repos+=( "$repo" )
                    fi
                fi
            done
        done
        cd "$cwd"
    fi
    if [[ "${#repos[@]}" -gt '0' ]]; then
        printf '%s ' "${repos[@]}"
        return 0
    fi
    return 1
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
    __git_get_all_repos "$@"
    exit $?
fi
unset sourced

return 0
