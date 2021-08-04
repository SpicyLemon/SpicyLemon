#!/bin/bash
# This file contains the github_config_as_personal function that sets local config options to values defined in the GITHUB_PERSONAL_CONFIG array.
# This file can be sourced to add the github_config_as_personal function to your environment.
# This file can also be executed to run the github_config_as_personal function without adding it to your environment.
#
# In your .bash_profile (or .zshrc, or whatever), export the GITHUB_PERSONAL_CONFIG array similar to this (using your own desired values):
# export GITHUB_PERSONAL_CONFIG=(
#   'user.email=github@wedul.com'
#   'user.name=Daniel Wedul'
# )
#
# Then, after cloning a personal repo (or whenever really), run the github_config_as_personal function (or this file).
#
# File contents:
#   github_config_as_personal  --> Sets local git config options to my personal github account options.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_pull_merge <branch>
github_config_as_personal () {
    if ! in_git_folder; then
        printf 'Not in a git repository.\n' >&2
        return 1
    fi
    if [[ -z "${GITHUB_PERSONAL_CONFIG+x}" ]]; then
        printf 'Array GITHUB_PERSONAL_CONFIG not defined in the environment.\n' >&2
        return 1
    fi
    local retval setting name val cur_val origin
    retval=0
    for setting in "${GITHUB_PERSONAL_CONFIG[@]}"; do
        # Split on the 1st equals sign
        name="$( sed 's/=.*$//' <<< "$setting" )"
        val="$( sed 's/[^=]*=//' <<< "$setting" )"
        printf 'Checking that setting [%s] is [%s] ... ' "$name" "$val"
        if cur_val="$( git config --local --get-all "$name" )"; then
            if [[ "$cur_val" != "$val" ]]; then
                printf 'Is currently [%s]; updating now ... ' "$cur_val"
                git config --local --replace-all "$name" "$val" || retval=$?
                printf 'Done.\n'
            else
                printf 'Already correct. Done.\n'
            fi
        else
            printf 'Not set yet; setting now ... '
            git config --local --add "$name" "$val" || retval=$?
            printf 'Done.\n'
        fi
    done
    if [[ -n "$GITHUB_PERSONAL_URL" ]]; then
        origin="$( git remote get-url origin )"
        printf 'Checking remote origin ... '
        if [[ ! "$origin" =~ github\.com-personal && "$origin" =~ github\.com ]]; then
            printf 'Is currently [%s]; updating now ... ' "$origin"
            git remote set-url origin "$( sed 's/github\.com/github\.com-personal/' <<< "$origin" )"
            printf 'Done.\n'
        else
            printf 'Already correct. Done.\n'
        fi
    else
        printf 'Variable GITHUB_PERSONAL_URL not defined. Skipping remote origin check.\n'
    fi
    printf '\n'
    __git_echo_do git remote get-url origin
    __git_echo_do git config --local --list
    return $retval
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
    github_config_as_personal "$@"
    exit $?
fi
unset sourced

return 0
