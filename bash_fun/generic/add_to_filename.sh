#!/bin/bash
# This file contains the add_to_filename function that adds text to a filename just before the first extension.
# This file can be sourced to add the add_to_filename function to your environment.
# This file can also be executed to run the add_to_filename function without adding it to your environment.
#
# File contents:
#   add_to_filename  --> Adds some text to a filename.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Adds some text to the end of a filename just before any extension
# Usage: add_to_filename <text to add> <filename> [<another filename> ...]
#   or   <stuff> | add_to_filename <text to add> -
add_to_filename () {
    local addition show_orig filenames filename
    addition="$1"
    shift
    if [[ "$1" == '-o' || "$1" == '--orig' || "$1" == '--original' ]]; then
        show_orig='YES'
        shift
    fi
    if [[ "$1" == '-' ]]; then
        filenames=( $( cat - ) )
    else
        filenames=( "$@" )
    fi
    if [[ "${#filenames[@]}" -eq '0' ]]; then
        echo 'No filenames provided.' >&2
        return 1
    fi
    for filename in "${filenames[@]}"; do
        [[ -n "$show_orig" ]] && printf '%s  ' "$filename"
        sed -E "s/(\.)|$/_$addition\1/" <<< "$filename"
    done
}

if [[ "$sourced" != 'YES' ]]; then
    add_to_filename "$@"
    exit $?
fi
unset sourced

return 0
