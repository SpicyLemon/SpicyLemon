#!/bin/bash
# This file contains the java_sdk_switcher function which makes it easier to switch between installed java sdk versionss.
# This file is meant to be sourced to add the java_sdk_switcher function to your environment.
# This function uses the sdkman_fzf wrapper for the SDK manager: https://sdkman.io/
# The sdkman_fzf function can be found in sdkman_fzf.sh in the same directory as this file.
# This function primarily still exists for historical reasons.
#
# File contents:
#   java_sdk_switcher  --> Switch between java SDK versions.
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

# Usage: java_sdk_switcher
java_sdk_switcher () {
    sdkman_fzf use java _
}

return 0
