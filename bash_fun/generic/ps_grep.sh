#!/bin/bash
# This file contains the ps_grep function that helps find processes.
# This file can be sourced to add the ps_grep function to your environment.
# This file can also be executed to run the ps_grep function without adding it to your environment.
#
# File contents:
#   ps_grep  --> Greps ps with provided input.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: ps_grep <grep parameters>
ps_grep () {
    local ps_results header processes grep_results
    if [[ "$#" -eq '0' || "$1" == '-h' || "$1" == '--help' ]]; then
        printf 'Usage: ps_grep <grep parameters>\n'
        return 0
    fi
    ps_results="$( ps aux )"
    header="$( head -n 1 <<< "$ps_results" )"
    if [[ "$sourced" == 'NO' ]]; then
        # If being executed, hide the process it's executing in.
        # It will always match because the parameters provided to the script are
        # also listed in the ps results.
        # But this process is (hopefully) not the one you're actually looking for.
        processes="$( tail -n +2 <<< "$ps_results" | grep -v "$$" )"
    else
        processes="$( tail -n +2 <<< "$ps_results" )"
    fi
    grep_results="$( grep --color=always "$@" <<< "$processes" )"
    if [[ -z "$grep_results" ]]; then
        printf 'No matching processes were found.\n'
        return 1
    fi
    printf '%s\n%s\n' "$header" "$grep_results"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    ps_grep "$@"
    exit $?
fi
unset sourced

return 0
