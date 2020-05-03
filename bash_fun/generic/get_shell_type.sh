#!/bin/bash
# This file contains the get_shell_type function that outputs either 'bash' or 'zsh' or the process running your shell.
# This file is meant to be sourced to add the get_shell_type function to your environment.
#
# File contents:
#   get_shell_type  --> Gets the type of shell you're in, either "zsh" "bash" or else the process running your shell
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within.
$( echo -e "\033[1;37msource $( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )\033[0m" )
EOF
    exit 1
fi
unset sourced

get_shell_type () {
    local shell_command
    shell_command=$( ps -o command= $$ | awk '{ print $1; }' )
    if [[ -n $( echo "$shell_command" | grep -E "zsh$" ) ]]; then
        printf 'zsh'
    elif [[ -n $( echo "$shell_command" | grep -E "bash$" ) ]]; then
        printf 'bash'
    else
        printf '%s' "$shell_command"
    fi
}

return 0
