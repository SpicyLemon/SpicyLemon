#!/bin/bash
# This file contains the git_merge_diff function that is a handy way to see various diffs during a merge with conflicts.
# This file can be sourced to add the git_merge_diff function to your environment.
# This file can also be executed to run the git_merge_diff function without adding it to your environment.
#
# File contents:
#   git_merge_diff  --> Compares two mid-merge versions of a file
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_merge_diff () {
    local usage
    usage="$( cat << EOF
git_merge_diff - Show differences between merge stages of a file when trying to resolve a merge conflict.

Usage: git_merge_diff <stage from> <stage to> <file> [<additional files>]

    <stage from> and <stage to> must be one of the following:
        1 common
        2 ours
        3 theirs
    The first argument is the "from" stage.
    The second argument is the "to" stage.
    The third (and remaining) arguments are the files to diff.
EOF
)"
    local stage_from stage_to
    if [[ "$#" -lt '3' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    stage_from="$( __git_merge_diff_parse_stage "$1" )" || return $?
    shift
    stage_to="$( __git_merge_diff_parse_stage "$1" )" || return $?
    shift
    while [[ "$#" -gt '0' ]]; do
        __git_echo_do git diff ":${stage_from}:${1}" ":${stage_to}:${1}"
        shift
    done
}

__git_merge_diff_parse_stage () {
    case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        1|common|commo|comm|com|co|c|--common|-c)
            printf '1'
            ;;
        2|ours|our|ou|o|--ours|-o)
            printf '2'
            ;;
        3|theirs|their|thei|the|th|t|--theirs|-t)
            printf '3'
            ;;
        *)
            printf 'Unknown <stage> [%s].\n' "$1" >&2
            return 1
            ;;
    esac
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
    require_command '__git_echo_do' || exit $?
    git_diff_analysis "$@"
    exit $?
fi
unset sourced

return 0
