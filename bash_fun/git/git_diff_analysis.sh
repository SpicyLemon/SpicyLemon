#!/bin/bash
# This file contains the git_diff_analysis function that outputs a simple diff report of your branch.
# This file can be sourced to add the git_diff_analysis function to your environment.
# This file can also be executed to run the git_diff_analysis function without adding it to your environment.
#
# File contents:
#   git_diff_analysis  --> Compares two branches and gets some diff stats.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_diff_analysis () {
    local usage
    usage="$( cat << EOF
git_diff_analysis - Gets some stats on branch differences.

Usage: git_diff_analysis [<main branch> [<branch with changes>]]

    If no branches are supplied, the diff will be from main to your current branch.
    If only one branch is supplied, the diff will be from that branch to your current branch.
    If two brances are supplied, the diff will be from the first branch to the second.
EOF
)"
    local branches verbose
    branches=()
    while [[ "$#" -gt '0' ]]; do
        case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        -v|--verbose)
            verbose='--verbose'
            ;;
        *)
            branches+=( "$1" )
            ;;
        esac
        shift
    done
    if [[ "${#branches[@]}" -gt '2' ]]; then
        printf 'Only two branches can be supplied.\n' >&2
        return 1
    elif [[ "${#branches[@]}" -eq '1' ]]; then
        branches+=( "$( git_branch_name )" )
    elif [[ "${#branches[@]}" -eq '0' ]]; then
        branches=( 'main' "$( git_branch_name )" )
    fi
    if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
        printf 'This command must be run from a git folder.\n' >&2
        return 1
    fi

    local diff_numstats_cmd diff_numstats test_entries proto_entries auto_entries code_entries
    local total_lines_added total_lines_removed test_lines_added test_lines_removed
    local proto_lines_added proto_lines_removed auto_lines_added auto_lines_removed
    local code_lines_added code_lines_removed
    local diff_no_context_cmd diff_tests_filter diff_tests
    local tests_added tests_removed tests_delta

    [[ -n "$verbose" ]] && printf '\033[96mCounting line changes.\033[0m\n' >&2

    diff_numstats_cmd=( git diff ${branches[@]} --numstat )
    [[ -n "$verbose" ]] && printf '  \033[97m%s\033[0m\n' "${diff_numstats_cmd[*]}" >&2
    diff_numstats="$( "${diff_numstats_cmd[@]}" )"
    test_entries="$( grep "_test\.go$" <<< "$diff_numstats" )"
    proto_entries="$( grep "\.proto$" <<< "$diff_numstats" )"
    auto_entries="$( grep -e "\.pb\.go$" -e "\.pb\.gw\.go$" <<< "$diff_numstats" )"
    code_entries="$( grep -v -e "_test\.go$" -e ".proto$" -e "\.pb\.go$" -e "\.pb\.gw\.go$" <<< "$diff_numstats" )"

    total_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$diff_numstats" )"
    total_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$diff_numstats" )"
    test_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$test_entries" )"
    test_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$test_entries" )"
    proto_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$proto_entries" )"
    proto_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$proto_entries" )"
    auto_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$auto_entries" )"
    auto_lines_removed="$( awk '{sum-+$2} END { print sum }' <<< "$auto_entries" )"
    code_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$code_entries" )"
    code_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$code_entries" )"

    [[ -n "$verbose" ]] && printf '\033[96mDone.\033[0m\n' >&2
    (
        # Bash and ksh arrays are 0 based while most other shells are 1 based.
        # Luckily, in those other ones, asking for the 0th element will just return nothing.
        printf '      Repo Directory: %s\n' "$( basename "$( git rev-parse --show-toplevel )" )"
        printf '         From Branch: %s\n' "$( [[ -n "${branches[0]}" ]] && printf %s "${branches[0]}" || printf %s "${branches[1]}" )"
        printf '           To Branch: %s\n' "$( [[ -n "${branches[0]}" ]] && printf %s "${branches[1]}" || printf %s "${branches[2]}" )"
        printf '========================================\n'
        printf 'Line Changes -  Code: %+6d  %s\n' "$code_lines_added" "$( printf '%+6d' "$code_lines_removed" | sed 's/+/-/' )"
        printf 'Line Changes - Proto: %+6d  %s\n' "$proto_lines_added" "$( printf '%+6d' "$proto_lines_removed" | sed 's/+/-/' )"
        printf 'Line Changes -  Auto: %+6d  %s\n' "$auto_lines_added" "$( printf '%+6d' "$auto_lines_removed" | sed 's/+/-/' )"
        printf 'Line Changes - Tests: %+6d  %s\n' "$test_lines_added" "$( printf '%+6d' "$test_lines_removed" | sed 's/+/-/' )"
        printf 'Line Changes - Total: %+6d  %s\n' "$total_lines_added" "$( printf '%+6d' "$total_lines_removed" | sed 's/+/-/' )"
    )
}

if [[ "$sourced" != 'YES' ]]; then
    git_diff_analysis "$@"
    exit $?
fi
unset sourced

return 0
