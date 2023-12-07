#!/bin/bash
# This file contains the go_imports function that extracts all the imports in go files.
# This file can be sourced to add the go_imports function to your environment.
# This file can also be executed to run the go_imports function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_imports () {
    local usage no_filenames files file results
    usage='Usage: go_imports <file> [<file 2> ...] [--no-filenames]'
    files=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            --no-filenames|--no-filename|--no-file)
                no_filenames=1
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

    if [[ "${#files[@]}" -eq '1' ]]; then
        no_filenames=1
    fi

    for file in "${files[@]}"; do
        if [[ ! -f "$file" ]]; then
            printf 'File not found: %q\n' "$file"
        else
            results="$( awk \
                '{
                    if (in_imp == 1) {
                        if (/^\)/) { in_imp = 0; }
                        else if (/[^[:space:]]/) {
                            gsub(/^[[:space:]]+/,"",$0);
                            gsub(/ ?\/\/.*$/,"",$0);
                            print $0;
                        };
                    }
                    else if (/^import[[:space:]]*\(/) { in_imp = 1; }
                    else if (/^import /) {
                        gsub(/ ?\/\/.*$/,"",$0);
                        for (i=2; i<NF; i++) printf $i " "; print $NF;
                    };
                }' "$file" )"
            [[ "$no_filenames" ]] || printf '%s:\n' "$file"
            printf '%s\n' "$results"
            [[ "$no_filenames" ]] || printf '\n'
        fi
    done

    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    go_imports "$@"
    exit $?
fi
unset sourced

return 0
