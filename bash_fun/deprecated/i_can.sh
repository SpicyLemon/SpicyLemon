#!/bin/bash
# This file contains the i_can and can_i functions that test if a function/program/command is available for use.
# This file is meant to be sourced to add the i_can and can_i functions to your environment.
#
# File contents:
#   i_can  --> Tests if a command is available.
#   can_i  --> Outputs results of i_can.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

# Usage: if i_can "foo"; then echo "I can totally foo"; else echo "There's no way I can foo."; fi
i_can () {
    if [[ "$#" -eq '0' ]]; then
        return 1
    fi
    command -v "$@" > /dev/null 2>&1
}

# Usage: can_i "foo"
can_i () {
    local c e
    c="$@"
    if [[ -z "$c" ]]; then
        printf 'Usage: can_i <command>\n'
        return 2
    fi
    if i_can "$c"; then
        printf 'Yes. You can [%s].\n' "$c"
        return 0
    else
        e="$?"
        printf 'No. You cannot [%s].\n' "$c"
        return $e
    fi
}

return 0
