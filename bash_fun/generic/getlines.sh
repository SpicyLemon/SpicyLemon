#!/bin/bash
# This file contains the getlines function that gets requested lines from a file.
# This file can be sourced to add the getlines function to your environment.
# This file can also be executed to run the getlines function without adding it to your environment.
#
# File contents:
#   getlines  --> Function for getting specific lines or line ranges from a file.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

getlines () {
    if ! command -v 'join_str' > /dev/null 2>&1; then
        printf 'Missing required command: join_str\n' >&2
        join_str
        return $?
    fi
    local usage
    usage='Usage: getlines <file> [<line number>|<line1>-<line2>]'
    local filename other_params verbose pieces piece error errors awk_clauses awk_test
    if command -v 'setopt' > /dev/null 2>&1; then
        setopt local_options BASH_REMATCH KSH_ARRAYS
    fi
    other_params=()
    if [[ "$#" -eq '0' ]]; then
        printf '%s' "$usage"
        return 0
    fi
    while [[ "$#" -gt '0' ]]; do
        case "$( tr '[:upper:]' '[:lower:]' <<< "$1" )" in
        -h|--help)
            printf '%s' "$usage"
            return 0
            ;;
        -v|--verbose)
            verbose='--verbose'
            ;;
        *)
            if [[ -z "$filename" && -f "$1" ]]; then
                filename="$1"
            else
                other_params+=( "$1" )
            fi
            ;;
        esac
        shift
    done
    errors=()
    awk_clauses=()
    pieces=( $( join_str ',' "${other_params[@]}" | sed -E 's/[[:space:]]+//g; s/,/ /g' ) )
    [[ -n "$verbose" ]] && printf 'Input:' && printf ' [%s]' "${pieces[@]}" && printf '.\n'
    for piece in "${pieces[@]}"; do
        if [[ "$piece" =~ ^([[:digit:]]+)-([[:digit:]]+)$ ]]; then
            awk_clauses+=( "(NR>=${BASH_REMATCH[1]}&&NR<=${BASH_REMATCH[2]})" )
        elif [[ "$piece" =~ ^([[:digit:]]+)$ ]]; then
            awk_clauses+=( "NR==${BASH_REMATCH[1]}" )
        else
            errors+=( "Unknown parameter: [$piece]." )
        fi
    done
    if [[ "${#awk_clauses[@]}" -eq '0' ]]; then
        errors+=( 'No lines requested.' )
    fi
    if [[ "${#errors[@]}" -gt '0' ]]; then
        printf '%s\n' "${errors[@]}" >&2
        return 1
    fi
    [[ -n "$verbose" ]] && printf 'Awk Clauses:' && printf ' [%s]' "${awk_clauses[@]}" && printf '.\n'
    awk_test="$( join_str '||' "${awk_clauses[@]}" )"
    [[ -n "$verbose" ]] && printf 'Awk Test: [%s].\n' "$awk_test"
    if [[ -n "$filename" ]]; then
        [[ -n "$verbose" ]] && printf 'Reading from filename: [%s].\n' "$filename" \
                            && printf '  \033[1mawk "%s" "%s"\033[0m\n' "$awk_test" "$filename"
        awk "$awk_test" "$filename"
    else
        [[ -n "$verbose" ]] && printf 'Reading from piped input.\n' "$filename" \
                            && printf '  \033[1mcat "-" | awk "%s"\033[0m\n' "$awk_test"
        cat "-" | awk "$awk_test"
    fi
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
    require_command 'join_str' || exit $?
    getlines "$@"
    exit $?
fi
unset sourced

return 0
