#!/bin/bash
# This file contains the git_checkout_tag function that uses fzf to let you select a tag to checkout as a new local branch.
# This file can be sourced to add the git_checkout_tag function to your environment.
# This file can also be executed to run the git_checkout_tag function without adding it to your environment.
#
# File contents:
#   git_checkout_tag  --> Select a tag and check it out as a new branch.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_checkout_tag [<tag>]
# If no <tag> is provided, you will be prompte to select one.
git_checkout_tag () {
    if ! in_git_folder; then
        printf 'git_checkout_tag: Not in a git repo.\n' >&2
        return 1
    fi
    local selection
    selection="$1"
    if [[ -z "$selection" ]]; then
        git fetch --tags
        selection="$( git tag | sort --version-sort | fzf +m --tac --cycle --header='Select the tag to checkout and press enter (or esc to cancel).' )"
    fi
    [[ -n "$selection" ]] && git checkout "$selection" -b "tag-$selection"
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
    git_checkout_tag "$@"
    exit $?
fi
unset sourced

return 0
