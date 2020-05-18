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

    If no branches are supplied, the diff will be from master to your current branch.
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
        branches=( 'master' "$( git_branch_name )" )
    fi
    if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
        printf 'This command must be run from a git folder.\n' >&2
        return 1
    fi

    local diff_numstats_cmd diff_numstats test_entries code_entries
    local total_lines_added total_lines_removed test_lines_added test_lines_removed code_lines_added code_lines_removed
    local diff_no_context_cmd diff_tests_filter diff_tests
    local tests_added tests_removed tests_delta

    [[ -n "$verbose" ]] && printf '\033[96mCounting line changes.\033[0m\n' >&2

    diff_numstats_cmd=( git diff ${branches[@]} --numstat )
    [[ -n "$verbose" ]] && printf '  \033[97m%s\033[0m\n' "${diff_numstats_cmd[*]}" >&2
    diff_numstats="$( "${diff_numstats_cmd[@]}" )"
    test_entries="$( grep 'src/test' <<< "$diff_numstats" )"
    code_entries="$( grep -v 'src/test' <<< "$diff_numstats" )"

    total_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$diff_numstats" )"
    total_lines_removed="$( awk '{sum-=$2} END { print sum }' <<< "$diff_numstats" )"
    test_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$test_entries" )"
    test_lines_removed="$( awk '{sum-=$2} END { print sum }' <<< "$test_entries" )"
    code_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$code_entries" )"
    code_lines_removed="$( awk '{sum-=$2} END { print sum }' <<< "$code_entries" )"

    [[ -n "$verbose" ]] && printf '\033[96mCounting unit test changes.\033[0m\n' >&2

    diff_no_context_cmd=( git diff ${branches[@]} -U0 )
    diff_tests_filter=( grep '@Test' )
    [[ -n "$verbose" ]] && printf '  \033[97m%s | %s\033[0m\n' "${diff_no_context_cmd[*]}" "${diff_tests_filter[*]}" >&2
    diff_tests="$( "${diff_no_context_cmd[@]}" | "${diff_tests_filter[@]}" )"

    tests_added="$( awk '{ if ($0 ~ /^\+/) sum+=1; } END { print sum; }' <<< "$diff_tests" )"
    tests_removed="$( awk '{ if ($0 ~ /^\-/) sum-=1; } END { print sum; }' <<< "$diff_tests" )"
    tests_delta=$(( tests_added + tests_removed ))

    [[ -n "$verbose" ]] && printf '\033[96mDone.\033[0m\n' >&2
    (
        # Bash and ksh arrays are 0 based while most other shells are 1 based.
        # Luckily, in those other ones, asking for the 0th element will just return nothing.
        printf '      Repo Directory: %s\n' "$( basename "$( git rev-parse --show-toplevel )" )"
        printf '         From Branch: %s\n' "$( [[ -n "${branches[0]}" ]] && printf %s "${branches[0]}" || printf %s "${branches[1]}" )"
        printf '           To Branch: %s\n' "$( [[ -n "${branches[0]}" ]] && printf %s "${branches[1]}" || printf %s "${branches[2]}" )"
        printf '========================================\n'
        printf 'Line Changes -  Code: %+6d  %+6d\n' "$code_lines_added" "$code_lines_removed"
        printf 'Line Changes - Tests: %+6d  %+6d\n' "$test_lines_added" "$test_lines_removed"
        printf 'Line Changes - Total: %+6d  %+6d\n' "$total_lines_added" "$total_lines_removed"
        printf 'Unit Tests -   Added: %+6d\n' "$tests_added"
        printf 'Unit Tests - Removed: %+6d\n' "$tests_removed"
        printf 'Unit Tests -   Delta: %+6d\n' "$tests_delta"
    )
}

if [[ "$sourced" != 'YES' ]]; then
    git_diff_analysis "$@"
    exit $?
fi
unset sourced

return 0
