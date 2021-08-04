#!/bin/bash
# This file contains the set_title function that changes the title of an iTerm tab.
# This file can be sourced to add the set_title function to your environment.
# This file can also be executed to run the fp function without adding it to your environment.
#
# File contents:
#   set_title  --> Set Title - Sets the title of the current iTerm tab.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Set the title of the current tab.
# Usage: set_title <title>
#   If no title is provided, then a default will be used.
#   If you are somewhere inside a git repo, the default is the directory name of the root of the repo.
#   If you are not in a git repo, the default is the top directory that you're in.
set_title () {
    if [[ "$TERM_PROGRAM" != 'iTerm.app' ]]; then
        return 1
    fi
    local title
    if [[ "$#" -gt '0' ]]; then
        title="$*"
    elif command -v git > /dev/null 2>&1 && git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
        title="$( basename "$( git rev-parse --show-toplevel )" )"
    else
        title="${PWD##*/}"
    fi
    printf '\033]0;%s\007' "$title"
}

if [[ "$sourced" != 'YES' ]]; then
    set_title "$@"
    exit $?
fi
unset sourced

return 0
