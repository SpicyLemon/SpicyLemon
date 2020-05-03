#!/bin/bash
# This file contains functions to help switch between versions of java.
# This file is meant to be sourced to add the functions to your environment.
#
# File contents:
#   java_8_activate  -----------------> Exports JAVA_HOME to point to Java 8.
#   java_8_deactivate  ---------------> Unsets JAVA_HOME.
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

# Usage: java_8_activate
java_8_activate () {
    export JAVA_HOME="$( /usr/libexec/java_home -v 1.8 )"
    echo -E "JAVA_HOME set to \"$JAVA_HOME\"."
}

# Usage: java_8_deactivate
java_8_deactivate () {
    unset JAVA_HOME
    echo -E "JAVA_HOME unset."
}

return 0