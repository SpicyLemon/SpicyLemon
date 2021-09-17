#!/bin/bash
# This file contains the in_git_folder function that returns 0 if currently in a git folder or 1 if not.
# This file can be sourced to add the in_git_folder function to your environment.
# This file can also be executed to run the in_git_folder function without adding it to your environment.
#
# File contents:
#   in_git_folder  --> Returns 0 (true) if currently in a git folder, or 1 (false) if not.
#
# Deprecated.
#    Being replaced with alias in_git_repo='git rev-parse --is-inside-work-tree > /dev/null 2>&1'
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: in_git_folder && echo "In a git folder!" || echo "Not in a git folder."
in_git_folder () {
    git rev-parse --is-inside-work-tree > /dev/null 2>&1 && return 0
    return 1
}

if [[ "$sourced" != 'YES' ]]; then
    in_git_folder "$@"
    exit $?
fi
unset sourced

return 0
