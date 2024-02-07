#!/bin/bash
# This file contains the git_commit_number function that runs git commit -m "[<number>]: <args>".
# This file can be sourced to add the git_commit_number function to your environment.
# This file can also be executed to run the git_commit_number function without adding it to your environment.
#
# File contents:
#   git_commit_number  --> If the current branch has the format <user>/<number>-<stuff>, then do
#                            git commit -m "[<number>]: <args>". Otherwise, do git commit -m "<args>".
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_commit_number <message>
git_commit_number () {
    if [[ "$#" -eq 0 ]]; then
        printf 'Usage: git_commit_number <message>\n'
        return 0
    fi
    local branch num msg
    # Get the branch. If we can't, it should print the problem to stderr, and there's nothing more we can do.
    branch="$( git symbolic-ref --short HEAD )" || return $?
    # Remove the leading '<user>/', then the trailing '-<stuff>'. Then, only keep what's left if it's only digits.
    num="$( sed -E 's/^[[:alnum:]]+\///; s/-.*$//;' <<< "$branch" | grep '^[[:digit:]]*$' )"
    # Create the msg.
    if [[ -n "$num" ]]; then
        printf -v msg '[%s]: %s' "$num" "$*"
    else
        msg="$*"
    fi
    # Print the command and put it in the history before running it.
    # For output and history, wrap the msg in quotes so that it looks like (and is) a single arg.
    printf 'git commit -m "%s"\n' "$msg"
    history -s git commit -m '"'"$msg"'"'
    # No need to add extra quotes here; the outside quotes make it all one arg.
    git commit -m "$msg"
}

if [[ "$sourced" != 'YES' ]]; then
    git_commit_number "$@"
    exit $?
fi
unset sourced

return 0
