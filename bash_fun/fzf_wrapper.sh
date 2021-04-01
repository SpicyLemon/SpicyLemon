#!/bin/bash
# This file contains a fzf wrapper function that adds some functionality to fzf.
# It can be sourced to add the fzf_wrapper function to your environment.
# It can also be executed in place of the fzf command.
# If you want to set up an alias for either, I suggest sourcing the file and aliasing the function.
#   It's a little bit faster that way.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Information on additional functionality added by this wrapper.
__fzf_wrapper_additions () {
    cat << EOF
  Custom wrapper additions
    --to-columns          Transform input into columns for display.
                          With this option, the delimiter turns into
                          a raw string rather than a regex.

EOF
}

# The main wrapper command that adds the extra stuff.
fzf_wrapper () {
    if ! command -v 'fzf' > /dev/null 2>&1; then
        printf 'Missing required command: fzf\n' >&2
        fzf
        return $?
    fi
    local fzf_cmd do_columns delimiter_flag delimiter exit_code
    fzf_cmd=( fzf )
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
        -h|--help)
            fzf "$1"
            __fzf_wrapper_additions
            return 0
            ;;
        --to-columns)
            do_columns='YES'
            ;;
        -d|--delimiter)
            delimiter_flag="$1"
            if [[ "$2" == '-' || ! "$2" =~ ^- ]]; then
                delimiter="$2"
                shift
            fi
            ;;
        --delimiter=*)
            delimiter_flag="--delimiter="
            delimiter="$( echo -E "$1" | sed 's/^--delimiter=//;' )"
            ;;
        -d*)
            delimiter_flag="-d"
            delimiter="$( echo -E "$1" | sed 's/^-d//;' )"
            ;;
        *)
            fzf_cmd+=( "$1" )
            ;;
        esac
        shift
    done
    if [[ -z "$do_columns" ]]; then
        if [[ -n "$delimiter_flag" ]]; then
            fzf_cmd+=( "--delimiter=$delimiter" )
        fi
        cat - | "${fzf_cmd[@]}"
        exit_code="${PIPESTATUS[1]}${pipestatus[2]}"
    else
        if [[ -z "$delimiter" ]]; then
            >&2 echo "No delimiter provided."
            return 8
        fi
        local delimiter_keyword delimiter_replace
        local add_marks to_columns undo_columns
        delimiter_keyword="$( echo -E "$delimiter" | sed 's/[]\/$*.^[]/\\&/g' )"
        delimiter_replace="$( echo -E "$delimiter" | sed 's/[\/&]/\\&/g' )"
        fzf_cmd+=( "--delimiter=$( __fzf_zwnj )" )
        add_marks=( sed -E "s/$delimiter_keyword/$( __fzf_zwsp )&$( __fzf_zwnj )/g" )
        to_columns=( column -s "$delimiter" -t )
        undo_columns=( sed -E "s/$( __fzf_zwsp )[ ]+$( __fzf_zwnj)/$delimiter_replace/g" )
        cat - | "${add_marks[@]}" | "${to_columns[@]}" | "${fzf_cmd[@]}" > >( "${undo_columns[@]}" )
        exit_code="${PIPESTATUS[3]}${pipestatus[4]}"
    fi
    return "$exit_code"
}

# Ouptuts a zero-width space
# These are used to indicate the start of the column spacing
# Usage: __fzf_zwsp
__fzf_zwsp () {
    printf "\xe2\x80\x8b"
}

# Outputs a zero-width non-joiner
# These are used to indicate the end of the column spacing
# Also used as the fzf delimiter when showing columns
# Usage: __fzf_zwnj
__fzf_zwnj () {
    printf "\xe2\x80\x8c"
}

# If this script was not sourced make it do things now.
if [[ "$sourced" != 'YES' ]]; then
    cat - | fzf_wrapper "$@"
    exit $?
fi
unset sourced

return 0
