#!/bin/bash
# This file contains the go_get_func function that extracts an entire function from some files.
# This file can be sourced to add the go_get_func function to your environment.
# This file can also be executed to run the go_get_func function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_get_func () {
    local usage func files recurse file results
    usage='Usage: go_get_func <func name> <file> [<file 2> ...] [-r|--recursive]'
    files=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            -r|--recursive)
                recurse=1
                ;;
            -)
                files+=( $( cat - ) )
                ;;
            *)
                if [[ -z "$func" ]]; then
                    func="$1"
                else
                    files+=( "$1" )
                fi
                ;;
        esac
        shift
    done

    if [[ "${#files[@]}" -eq '0' && ! -t 0 ]]; then
        files+=( $( cat - ) )
    fi

    if [[ -z "$func" || "${#files[@]}" -eq '0' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi

    if [[ "${#files[@]}" -eq '1' && -d "${files[*]}" ]]; then
        recurse=1
    fi

    for file in "${files[@]}"; do
        if [[ ! -f "$file" ]]; then
            if [[ -d "$file" ]]; then
                if [[ "$recurse" ]]; then
                    find "$file" -type f -name '*.go' -not -path '*/vendor/*' | go_get_func "$func" -
                else
                    printf 'Skipping directory: %q\n' "$file"
                fi
            else
                printf 'File not found: %q\n' "$file"
            fi
        else
            results="$( awk -v funcre=" $func\(\\\(\|\\\[\)" \
                '{
                    if(in_func == 1) {
                        print $0;
                        if (/^\}/) { in_func = 0; };
                    };
                    if (/^func/ && $0 ~ funcre) {
                        if (length(comment) > 0) { print comment; }
                        print $0;
                        if ($0 !~ /\}[[:space:]]*$/) { in_func=1; };
                    };
                    if(in_func != 1) {
                        if (/^\/\//) {
                            if (length(comment) == 0) {
                                comment = $0;
                            } else {
                                comment = comment "\n" $0
                            }
                        } else {
                            comment = "";
                        };
                    };
                }' "$file" )"
            if [[ -n "$results" ]]; then
                printf '%s:\n%s\n\n' "$file" "$results"
            fi
        fi
    done

    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    go_get_func "$@"
    exit $?
fi
unset sourced

return 0
