#!/bin/bash
# This file contains the jqq function that slightly shortens running jq on an environment variable.
# This file can be sourced to add the jqq function to your environment.
# This file can also be executed to run the jqq function without adding it to your environment.
#
# File contents:
#   jqq  --> Shortcut for jq from an environment variable.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Just makes it easier to use jq on a variable.
# This is basically just a shortcut for  echo <json> | jq <options> <query>
# If the query is omitted '.' is used.
# Usage: jqq <json> [<query>] [<options>]
jqq () {
    local json query
    json="$1"
    shift
    if [[ "$json" == '-h' || "$json" == '--help' ]]; then
        cat << EOF
jqq - Quick jq command for dealing with json in environment variables.

Usage: jqq <json> [<query>] [<options>]

    The first argument is taken to be the json.
    The query is optional. The default is '.'.
    If the query is provided, all other arguments are passed in as options to jq.
    If the second argument starts with a - (dash) then it is treated as an option and the default query is used.

    Examples:
        jqq "\$foo"
        jqq "\$foo" -c
        jqq "\$foo" '.[]'
        jqq "\$foo" '.[3].name' -r

EOF
        return 0
    fi
    if [[ "$1" =~ ^- ]]; then
        query='.'
    else
        query="$1"
        shift
    fi
    echo "$json" | jq "$@" "$query"
}

if [[ "$sourced" != 'YES' ]]; then
    jqq "$@"
    exit $?
fi
unset sourced

return 0
