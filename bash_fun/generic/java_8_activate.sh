#!/bin/bash
# This file contains the java_8_activate function which exports JAVA_HOME so that java 8 is active.
# This file is meant to be sourced to add the java_8_activate function to your environment.
#
# File contents:
#   java_8_activate  --> Exports JAVA_HOME to point to Java 8.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

# Usage: java_8_activate
java_8_activate () {
    export JAVA_HOME="$( /usr/libexec/java_home -v 1.8 )"
    echo -E "JAVA_HOME set to \"$JAVA_HOME\"."
}

return 0
