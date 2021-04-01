#!/bin/bash
# This file contains the java_sdk_switcher function which makes it easier to switch between installed java sdk versionss.
# This file is meant to be sourced to add the java_sdk_switcher function to your environment.
# This script uses the sdkman SDK manager: https://sdkman.io/
# Installation of SDKs is up to you. This just makes it easier to switch between installed ones.
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
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

# Usage: java_sdk_switcher
java_sdk_switcher () {
    local do_not_run
    for req_cmd in 'sdk' 'fzf'; do
        if ! command -v "$req_cmd" > /dev/null 2>&1; then
            do_not_run='yes'
            printf 'Missing required command: %s\n' "$req_cmd" >&2
            "$req_cmd"
        fi
    done
    if [[ -n "$do_not_run" ]]; then
        return 1
    fi
    if [[ "$1" == '-h' || "$1" == '--help' ]]; then
        printf 'To install SDKs, start with this command: sdk list java\n'
        printf 'See https://sdkman.io/ for more info.\n'
        printf 'Once some java SDKs are installed, use this function to easily switch between them: java_sdk_switcher\n'
        return 0
    fi
    local new_version sdk_list
    if [[ -n "$1" ]]; then
        new_version="$1"
    else
        sdk_list="$( sdk list java )"
        new_version="$( grep -E '(installed|local only)' <<< "$sdk_list" \
                        | awk '{split($0,a,"|"); print a[6]"~"a[2]"~"a[4]}' \
                        | sort -k1 -V \
                        | awk '{split($0,a,"~"); print a[2]"~"a[3]"~"a[1]}' \
                        | column -s '~' -t \
                        | fzf +m \
                        | sed -E 's/^.*[[:space:]]+([^[:space:]]+)[[:space:]]*$/\1/' )"
    fi
    if [[ -n "$new_version" ]]; then
        sdk use java "$new_version"
    fi
}

return 0
