#!/bin/bash
# This file contains the go_find_funcs_without_comments function that searches a file for functions that don't have comments.
# This file can be sourced to add the go_find_funcs_without_comments function to your environment.
# This file can also be executed to run the go_find_funcs_without_comments function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_find_funcs_without_comments () {
    local usage files file results
    usage='Usage: go_find_funcs_without_comments <file> [<file 2> ...]'
    files=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            -)
                files+=( $( cat - ) )
                ;;
            *)
                files+=( "$1" )
                ;;
        esac
        shift
    done

    if [[ "${#files[@]}" -eq '0' && ! -t 0 ]]; then
        files+=( $( cat - ) )
    fi

    if [[ "${#files[@]}" -eq '0' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi

    for file in "${files[@]}"; do
        if [[ ! -f "$file" ]]; then
            printf 'File not found: %q\n' "$file"
        else
            results="$( awk '{if ($0 ~ /^func/ && ll !~ /^\/\//){ print "  " $0; }; ll = $0; }' "$file" \
                | sed -E 's/( [[:alnum:]]+)\(.*$/\1/' \
                | grep -Ev ' [a-z][[:alnum:]]+$' )"
            if [[ -n "$results" ]]; then
                printf '%s\n%s\n\n' "$file" "$results"
            fi
        fi
    done

    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    go_find_funcs_without_comments "$@"
    exit $?
fi
unset sourced

return 0
