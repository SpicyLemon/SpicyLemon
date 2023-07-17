#!/bin/bash
# This file contains the git_recolor_diff function that outputs a colorized version of the provided git diff output.
# This file can be sourced to add the git_recolor_diff function to your environment.
# This file cannot be executed.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

# Usage: git_recolor_diff <filename>
# Usage: <do stuff> | git_recolor_diff
# Usage: git_recolor_diff <<< "<stuff to color>"
git_recolor_diff () {
    if [[ "$#" -gt '0' ]]; then
        cat "$1" | git_recolor_diff
        return "${PIPESTATUS[0]}${pipestatus[1]}"
    fi

    local meta context old new
    context=36
    meta=1
    old=31
    new=32

    # TODO: if in a git folder, check for diff color configuration and use those colors.
    # git config color.diff.[meta|context|old|new]

    GREP_COLOR=$meta grep --color=always -E '^diff.*$|^index.*$|^--- .*$|^\+\+\+ .*$|$' \
    | GREP_COLOR=$context grep --color=always -E '@@.*@@|$' \
    | GREP_COLOR=$old grep --color=always -E '^-[^-].*$|$' \
    | GREP_COLOR=$new grep --color=always -E '^\+[^+].*$|$'

    return 0
}

return 0
