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
git_commit_diff - Shows the diff for one or more commits.

Usage: git_commit_diff [git diff args] [--select] [--commit <commit hash>] [--hash <commit hash>] [-- <commit hash> ...]

    If --select is supplied, you will be prompted to select the commit(s) to show.
    Multiple commit hashes can be provided using multiple --commit or --hash options.
    Anything provided after -- will also be used as a commit hash.
    All other arguments are provided to the git diff command.

EOF
)"
    local do_select diff_args commit_hashes zwnj commit_hash return_code exit_code
    diff_args=()
    commit_hashes=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf '%s\n' "$usage"
                return 0
                ;;
            --select)
                do_select=YES
                ;;
            --commit|--hash)
                if [[ -z "$2" ]]; then
                    printf 'No hash provided after %s\n' "$1" >&2
                fi
                commit_hashes+=( "$2" )
                shift
                ;;
            --)
                shift
                commit_hashes+=( "$@" )
                set --
                break
                ;;
            *)
                diff_args+=( "$1" )
                ;;
        esac
        shift
    done
    if [[ -n "$do_select" ]]; then
        zwnj="$( printf '\xe2\x80\x8b' )"
        commit_hashes+=(
            $( git log --date=format:'%F %T %A' --format=format:"%H${zwnj}%h  %<(30)%ad %an${zwnj}%<(60,trunc)%s" \
                | fzf -m --cycle --tac --with-nth=2,3 --delimiter="$zwnj" --header=' hash     commit date                    author          subject' \
                | sed -E 's/([[:xdigit:]]+).*$/\1/;' )
        )
    fi
    if [[ "${#commit_hashes[@]}" -eq '0' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    return_code=0
    for commit_hash in "${commit_hashes[@]}"; do
        printf '> git log "%s" -n 1\n' "$commit_hash" \
            && git log "$commit_hash" -n 1 \
            && printf '\n> git --no-pager diff %s "%s~" "%s"\n' "${diff_args[*]}" "$commit_hash" "$commit_hash" \
            && git --no-pager diff "${diff_args[@]}" "$commit_hash~" "$commit_hash"
        exit_code=$?
        if [[ $exit_code -ne 0 ]]; then
            return_code=$exit_code
        fi
    done
    return $return_code
}

if [[ "$sourced" != 'YES' ]]; then
    git_commit_diff "$@"
    exit $?
fi
unset sourced

return 0
