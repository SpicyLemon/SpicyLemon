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
    local default_branch cur_branch local_branches head_inst head_def head_cur header height branches
    # Don't allow selection of the current branch because it can't be deleted without switching branches.
    # Also don't allow selection of the default branch since that's almost never what someone wants to do.
    # The fzf header should list the unselectable branches because completely omitting them causes confusion.
    # If we're in a detatched state:
    #   `git branch --show-current` returns nothing.
    #   `git rev-parse --abbrev-ref HEAD` returns "HEAD".
    #   `git branch` puts a "* " before an entry that looks like "(HEAD detatched at <hash/tag>)" (or something similar).
    # I use `git branch` here to identify the current branch so that we get that extra context when detatched.
    # I've no idea if that will help anything, but I'm pretty sure it won't hurt anything.
    default_branch="$( git_get_default_branch )"
    cur_branch="$( git branch | grep '^*' | sed 's/^\*[[:space:]]*//; s/[[:space:]]*$//;' )"
    local_branches="$( git branch | grep -v '^\*' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//g;' | grep -vFx "$default_branch" | sort )"
    if [[ -n "$local_branches" ]]; then
        head_inst='Select branches to delete using tab. Press enter when ready (or esc to cancel).'
        head_def="Default: $default_branch (not listed)."
        head_cur="Current: $cur_branch (not listed)."
        printf -v header '%s\n%s  %s\n' "$head_inst" "$head_cur" "$head_def"
        height="$(( $( wc -l <<< "$local_branches" ) + 4 ))" # +2 for the header and +2 for the fzf info line and prompt.
        branches="$( fzf --layout reverse-list -m --cycle --header-lines 2 --height "$height" <<< "${header}${local_branches}" )"
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
    require_command 'git_get_default_branch' || exit $?
    require_command '__git_echo_do' || exit $?
    git_delete_branches "$@"
    exit $?
fi
unset sourced

return 0
