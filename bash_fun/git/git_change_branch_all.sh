#!/bin/bash
# This file contains the git_change_branch_all function that uses fzf to let you select a branch (including remote branches) to switch to.
# This file can be sourced to add the git_change_branch_all function to your environment.
# This file can also be executed to run the git_change_branch_all function without adding it to your environment.
#
# File contents:
#   git_change_branch_all  --> Gets a list of all branches (local and remote) and lets you pick one to checkout.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_change_branch_all
git_change_branch_all () {
    if ! in_git_folder; then
        printf 'git_change_branch_all: Not in a git repo.\n' >&2
        return 1
    fi
    command -v 'setopt' > /dev/null 2>&1 && setopt local_options BASH_REMATCH KSH_ARRAYS
    local options remote remote_branches known_branches new_options selected_entry remote_and_branch_rx just_branch_rx branch
    options="$( git branch | sed -E 's#^([* ]) #\1 ~ ~#' )"
    for remote in $( git remote ); do
        git fetch -q "$remote"
        remote_branches="$( git ls-remote --heads "$remote" | sed -E 's#^.*refs/heads/##' )"
        if [[ -n "$remote_branches" ]]; then
            known_branches="$( sed -E 's#^[^~]*~[^~]*~##' <<< "$options" )"
            new_options="$( printf '%s\n%s\n%s\n' "$known_branches" "$known_branches" "$remote_branches" | sort | uniq -u | sed -E "s#^#  ~$remote~#" )"
            if [[ -n "$new_options" ]]; then
                printf -v options '%s\n%s' "$options" "$new_options"
            fi
        fi
    done
    selected_entry="$( sort -t '~' -k 3 -k 2 <<< "$options" | column -s '~' -t | fzf +m --cycle  --header='Select the branch to change to and press enter (or esc to cancel).' )"
    if [[ -z "$selected_entry" ]]; then
        return 0
    fi
    remote_and_branch_rx='^[* ] +([^ ]+) +(.+)$'
    just_branch_rx='^[* ] +(.+)$'
    if [[ "$selected_entry" =~ $remote_and_branch_rx ]]; then
        remote="${BASH_REMATCH[1]}"
        branch="${BASH_REMATCH[2]}"
        __git_echo_do git checkout --track "$remote/$branch"
    elif [[ "$selected_entry" =~ $just_branch_rx ]]; then
        branch="${BASH_REMATCH[1]}"
        __git_echo_do git checkout "$branch"
    else
        printf 'Unknown selection: [%s]\n' "$selected_entry" >&2
        return 5
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
    git_change_branch_all "$@"
    exit $?
fi
unset sourced

return 0
