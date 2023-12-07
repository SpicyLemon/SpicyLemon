#!/bin/bash
# This file contains the go_count_imports function that outputs count info on go imports in/under a directory.
# This file can be sourced to add the go_count_imports function to your environment.
# This file can also be executed to run the go_count_imports function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_count_imports () {
    local usage ds
    usage='Usage: go_count_imports [<dir> ...]'
    ds=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            *)
                ds+=( "$1" )
                ;;
        esac
        shift
    done

    if [[ "${#ds[@]}" -eq '0' ]]; then
        ds+=( '.' )
    fi

    # Find all the go files we care about in the given directories.
    # Get all the imports in each of those files.
    # Put the alias (if there is one) behind the library).
    # Sort it all, keeping each unique entry only once.
    # Group the lines by library and output a <count> <library> <aliases>.
    # Sort the lines numerically to put the higher counts at the bottom.
    find "${ds[@]}" -type f -name '*.go' -not -path './vendor/*' -not -name '*.pb.go' -not -name '*.pb.gw.go' \
        | go_imports - --no-file \
        | sed -E 's/^([^[:space:]]+) (.+)$/\2 \1/' \
        | sort -u \
        | awk \
            '{
                lib = $1;
                alias = $2; if (alias == "") { alias = "[none]"; };
                if (prev == lib) {
                    count = count + 1;
                    aliases = aliases " " alias;
                } else {
                    if (prev != "") { print count " " prev " " aliases; };
                    prev = lib;
                    aliases = alias;
                    count = 1;
                };
            }
            END {
                if (prev != "") { print count " " prev " " aliases; };
            }' \
        | sort -n

    return 0
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
    require_command 'go_imports' || exit $?
    go_count_imports "$@"
    exit $?
fi
unset sourced

return 0
