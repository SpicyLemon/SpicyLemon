#!/bin/bash
# This file contains the git_commit_diff function that gets the diff of a commit.
# This file can be sourced to add the git_commit_diff function to your environment.
# This file can also be executed to run the git_commit_diff function without adding it to your environment.
#
# File contents:
#   git_commit_diff  --> Provides the diff for a commit.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_commit_diff () {
    local usage
    usage="$( cat << EOF
git_commit_diff - Shows the commit message and diff for a single commit.

Usage: git_commit_diff <commit hash> [git diff args]

    The first argument must either be the commit hash or --select.
    If --select is supplied, you will be prompted to select the commit to show (using fzf).
    All other arguments are provided to the git diff command.

EOF
)"
    if [[ "$1" == '--help' || "$1" == '-h' || "$1" == 'help' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    local commit_hash diff_args zwnj
    commit_hash="$1"
    shift
    diff_args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf '%s\n' "$usage"
                return 0
                ;;
            *)
                diff_args+=( "$1" )
                ;;
        esac
        shift
    done
    if [[ -z "$commit_hash" || "$commit_hash" == '--select' ]]; then
        if ! command -v fzf > /dev/null 2>&1; then
            printf 'fzf not available for commit selection.\n' >&2
            fzf >&2
            return $?
        fi
        zwnj="$( printf '\xe2\x80\x8b' )"
        commit_hash="$( git log --date=format:'%F %T %A' --format=format:"%H${zwnj}%h  %<(30)%ad %an${zwnj}%<(60,trunc)%s" \
                | tac \
                | fzf +m --cycle --tac --with-nth=2,3 --delimiter="$zwnj" --layout=reverse-list \
                    --header=' hash     commit date                    author          subject' \
                | sed -E 's/([[:xdigit:]]+).*$/\1/;'
        )"
    fi
    if [[ -z "$commit_hash" ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    printf '> git --no-pager log "%s" -n 1\n' "$commit_hash" \
        && git --no-pager log "$commit_hash" -n 1 \
        && printf '\n> git --no-pager diff "%s~" "%s" %s\n' "$commit_hash" "$commit_hash" "${diff_args[*]}" \
        && git --no-pager diff "$commit_hash~" "$commit_hash" "${diff_args[@]}"
    return $?
}

if [[ "$sourced" != 'YES' ]]; then
    git_commit_diff "$@"
    exit $?
fi
unset sourced

return 0
