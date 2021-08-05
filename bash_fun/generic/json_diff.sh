#!/bin/bash
# This file contains the json_diff function for getting a diff of JSON files.
# This file can be sourced to add the json_diff function to your environment.
# This file can also be executed to run the json_diff function without adding it to your environment.
#
# File contents:
#   json_diff  --> Function for diffing JSON files.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

json_diff () {
    local usage diff_args use_json_info file1 file2
    usage="$( cat << EOF
Usage: json_diff [<diff args>] [--use-json-info|--use-jq] <file1> <file2>

    <file1> and <file2> are the filenames of the json to diff.
        These are always assumed to be the last two arguments provided.

   <diff args> are any arguments you want provided to the diff command.

    --use-json-info will use the json_info function to pre-process the json files.
        The resulting diff will then be on the output of json_info.
    --use-jq will use the jq program to pre-process the json files.
        The resulting diff will then be on the pretty-print version of the JSON.
        This is the default behavior.
    If both --use-json-info and --use-jq are provided, the last one provided is used.
EOF
    )"
    if [[ "$#" -eq '0' ]]; then
        printf '%s\n\n' "$usage"
        return 0
    fi
    diff_args=()
    while [[ "$#" -gt '2' ]]; do
        case "$1" in
        -h|--help)
            printf '%s\n\n' "$usage"
            return 0
            ;;
        --use-json-info)
            use_json_info='YES'
            ;;
        --use-jq)
            use_json_info=''
            ;;
        *)
            diff_args+=( "$1" )
            ;;
        esac
        shift
    done
    file1="$1"
    file2="$2"
    shift
    shift
    if [[ "$file1" == '-h' || "$file1" == '--help' || "$file2" == '-h' || "$file2" == '--help' ]]; then
        printf '%s\n\n' "$usage"
        return 0
    fi
    if [[ "$file1" == '' ]]; then
        printf 'json_diff: File 1 not provided.\n' >&2
        return 1
    elif [[ "$file2" == '' ]]; then
        printf 'json_diff: File 2 not provided.\n' >&2
        return 1
    fi
    if [[ "$#" -gt '0' ]]; then
        printf 'json_diff: Unknown arguments: %s\n' "$*" >&2
        return 1
    fi
    if [[ -n "$use_json_info" ]]; then
        diff_args+=( '--pre-processor' 'json_info' )
    else
        diff_args+=( '--pre-processor' 'jq' )
    fi
    if ! command -v 'multidiff' > /dev/null 2>&1; then
        printf 'json_diff: Missing required command: multidiff\n' >&2
        command 'multidiff' >&2
        return $?
    fi

    multidiff "${diff_args[@]}" -- "$file1" "$file2"
}

if [[ "$sourced" != 'YES' ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    require_command () {
        local cmd cmd_fn
        cmd="$1"
        if ! command -v "$cmd" > /dev/null 2>&1; then
            cmd_fn="$where_i_am/$cmd.sh"
            if [[ -f "$cmd_fn" ]]; then
                source "$cmd_fn"
                if [[ "$?" -ne '0' ]] || ! command -v "$cmd" > /dev/null 2>&1; then
                    ( printf 'This script relies on the [%s] function.\n' "$cmd"
                      printf 'The file [%s] was found and sourced, but there was a problem loading the [%s] function.\n' "$cmd_fn" "$cmd" ) >&2
                    return 1
                fi
            else
                ( printf 'This script relies on the [%s] function.\n' "$cmd"
                  printf 'The file [%s] was looked for, but not found.\n' "$cmd_fn" ) >&2
                return 1
            fi
        fi
    }
    require_command 'multidiff' || exit $?
    json_diff "$@"
    exit $?
fi
unset sourced

return 0
