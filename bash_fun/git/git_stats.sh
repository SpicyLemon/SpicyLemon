#!/bin/bash
# This file contains the git_stats function that returns 0 if currently in a git folder or 1 if not.
# This file can be sourced to add the git_stats function to your environment.
# This file can also be executed to run the git_stats function without adding it to your environment.
#
# File contents:
#   git_stats  --> Outputs some git change statistics.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: git_stats && echo "In a git folder!" || echo "Not in a git folder."
git_stats () {
    if !git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
        printf 'not in a git folder\n' >&2
        return 1
    fi
    local git_args author_given any_author
    git_args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h)
                cat << EOF
Usage: git_stats [--any-author] [<args for git log>]

If an --author is not provided, the current user's name will be used.
Provide the --any-author flag to prevent this automatic inclusion of a default --author value.

Commonly provided <args for git log>:
    --since='1 Oct, 2022'
    --before='1 Oct, 2022'
    --until='1 Oct, 2022' (same as --before)
    -<number>, -n <number>, --max-count=<number>
    --no-merges

EOF
                return 0
                ;;
            --any-author)
                any_author='YES'
                ;;
            --author*)
                author_given='YES'
                git_args+=( "$1" )
                ;;
            *)
                git_args+=( "$1" )
                ;;
        esac
        shift
    done
    if [[ -z "$author_given" && -z "$any_author" ]]; then
        git_args=( "--author=$( git config user.name )" "${git_args[@]}" )
    fi
    git_args=( --shortstat "${git_args[@]}" )
    [[ -n "$DEBUG" ]] && printf 'git log' && printf ' %q' "${git_args[@]}" && printf '\n'
    git log "${git_args[@]}" | grep -E 'fil(e|es) changed' | awk '{commits+=1; files+=$1; inserted+=$4; deleted-=$6} END {delta=inserted+deleted; if (inserted!=0) {ratio=-1*deleted/inserted} else {ratio=9.9999999}; printf "     Commits: %6d\nFile Changes: %6d\n Lines Added: %+6d\n     Deleted: %+6d\n       Delta: %+6d\n         +/-: %7.6f\n", commits, files, inserted, deleted, delta, ratio }' -
    return "${PIPESTATUS[0]}${pipestatus[1]}"
}

if [[ "$sourced" != 'YES' ]]; then
    git_stats "$@"
    exit $?
fi
unset sourced

return 0
