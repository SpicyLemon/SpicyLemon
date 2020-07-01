#!/bin/bash
# This file contains the git_delete_branches function that uses fzf to let you select branches to delete, then deletes them.
# This file can be sourced to add the git_delete_branches function to your environment.
# This file can also be executed to run the git_delete_branches function without adding it to your environment.
#
# File contents:
#   git_delete_branches  --> Select branches that you want to delete, and then deletes them.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_delete_branches
git_delete_branches () {
    if ! in_git_folder; then
        printf 'git_delete_branches: Not in a git repo.\n' >&2
        return 1
    fi
    local local_branches branches
    local_branches="$( git branch | grep -v -e '^\*' -e ' master[[:space:]]*$' | sed -E 's/^ +| +$//g' | sort )"
    if [[ -n "$local_branches" ]]; then
        branches="$( fzf --tac -m --cycle --header="Select branches to delete using tab. Press enter when ready (or esc to cancel)." <<< "$local_branches" )"
        if [[ -n "$branches" ]]; then
            for branch in $( sed -l '' <<< "$branches" ); do
                __git_echo_do git branch -D "$branch"
            done
        else
            printf 'No branches selected for deletion.\n' >&2
            return 2
        fi
    else
        printf 'No branches to delete.\n' >&2
        return 3
    fi
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
    require_command '__git_echo_do' || exit $?
    git_delete_branches "$@"
    exit $?
fi
unset sourced

return 0
