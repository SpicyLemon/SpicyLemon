#!/bin/bash
# This file contains the asciidec2text function that converts ascii decimal values to characters.
# This file can be sourced to add the asciidec2text function to your environment.
# This file can also be executed to run the asciidec2text function without adding it to your environment.
#
# File contents:
#   asciidec2text  --> Converts ascii decimal values to characters
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

asciidec2text () {
    local vals arg show_help ec
    vals=()
    for arg in "$@"; do
        case "$arg" in
            -h|--help|help)
                show_help='YES'
                ;;
            -)
                vals+=( $( cat - ) )
                ;;
            *)
                vals+=( $arg )
                ;;
        esac
    done
    if [[ "$#" -eq '0' && ! -t 0 ]]; then
        vals+=( $( cat - ) )
    fi
    if [[ "${#vals[@]}" -eq '0' ]]; then
        show_help='YES'
    fi
    if [[ -n "$show_help" ]]; then
        cat << EOF
Converts decimal ascii values into text.

Usage: asciidec2text <values>
   or: <stuff> | asciidec2text

EOF
        return 0
    fi

    ec=0
    for arg in "${vals[@]}"; do
        if [[ -z "$arg" || ! "$arg" =~ ^[[:digit:]]*$ || "$arg" -gt '127' ]]; then
            printf 'Invalid decimal ascii value: [%s]\n' "$arg" >&2
        else
            case "$arg" in
                9) printf '[\\t]';;
                10) printf '[\\n]';;
                13) printf '[\\r]';;
                27) printf '[\\e]';;
                *)
                    if [[ "$arg" -ge '32' && "$arg" -le '126' ]]; then
                        printf '%b' "$( printf '\\x%x' "$arg" )"
                    else
                        printf '[%b]' "$( printf '\\x%x' "$arg" )" | od -a -An | sed 's/[[:space:]]//g' | tr -d '\n' | tr '[:lower:]' '[:upper:]'
                    fi
                    ;;
            esac
        fi
    done
}

if [[ "$sourced" != 'YES' ]]; then
    asciidec2text "$@"
    exit $?
fi
unset sourced

return 0
