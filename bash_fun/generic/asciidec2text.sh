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

    # Zero Width Non-Joiner = \xe2\x80\x8c
    # It's printed before printing any non-printable, converted value.
    ec=0
    for arg in "${vals[@]}"; do
        if [[ -z "$arg" || "$arg" =~ [^[:digit:]] || "$arg" -gt 255 ]]; then
            # Invalid value.
            printf '\xe2\x80\x8c[!%s]' "$arg"
            ec=1
        else
            # Strip leading zeros since that makes them behave like octal.
            # E.g. [[ '033' -eq '27' ]] is true.
            # But since this is specific to decimal, assume everything coming in is meant to be decimal.
            arg="$( sed -E 's/^0+//' <<< "$arg" )"
            if [[ "$arg" -ge '32' && "$arg" -le '126' ]]; then
                # Printable ascii value
                printf '%b' "$( printf '\\x%x' "$arg" )"
            elif [[ "$arg" -eq '9' ]]; then
                printf '\xe2\x80\x8c\\t' # Tab: [HT] was confusing me and \t is more familiar.
            elif [[ "$arg" -eq '10' ]]; then
                printf '\xe2\x80\x8c\\n' # Newline: Most commonly seen as \n, even though [NL] is easy to guess at.
            elif [[ "$arg" -eq '13' ]]; then
                printf '\xe2\x80\x8c\\r' # Carriage Return: Most commonly seen as \r, even though [CR] is easy to guess at.
            elif [[ "$arg" -le '127' ]]; then
                # Non-printable ascii value (not specifically covered earlier).
                # Use od to convert it to it's named character version.
                # E.g. 0 will become NUL and 7 will become BELL
                printf '\xe2\x80\x8c[%s]' "$( printf '%b' "$( printf '\\x%x' "$arg" )" | od -a -An | sed 's/[[:space:]]//g' | tr -d '\n' | tr '[:lower:]' '[:upper:]' )"
            else
                printf '\xe2\x80\x8c[%s]' "$arg"
            fi
        fi
    done

    return "$ec"
}

if [[ "$sourced" != 'YES' ]]; then
    asciidec2text "$@"
    exit $?
fi
unset sourced

return 0
