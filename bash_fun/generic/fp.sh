#!/bin/bash
# This file contains the fp function that outputs the full path to one or more files.
# This file can be sourced to add the fp function to your environment.
# This file can also be executed to run the fp function without adding it to your environment.
#
# File contents:
#   fp  --> File Path - Get the full path to one or more files either passed in as arguments or selected.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Get the full path to a file
# Usage: fp
#   or   fp <filename 1> [<filename 2> ...]
#   or   <stuff> | fp -
fp () {
    local filenames filename fullpath fullpaths
    if [[ "$#" -eq '1' && "$1" == '-' ]]; then
        filenames=( $( cat - ) )
    elif [[ "$#" -gt '0' ]]; then
        filenames=( "$@" )
    elif command -v "fzf" > /dev/null 2>&1; then
        filenames=( $( ls -a | grep -v "^\.\.$" | sort -f | fzf -m --tac --cycle ) )
    else
        filenames=()
    fi
    if [[ "${#filenames[@]}" -eq '0' ]]; then
        printf 'No filenames provided or selected.\n'
        return 1
    fi
    if command -v "setopt" > /dev/null 2>&1; then
        setopt local_options BASH_REMATCH KSH_ARRAYS
    fi
    fullpaths=()
    for filename in "${filenames[@]}"; do
        fullpath="$PWD/$filename"
        # Convert /./ to just /
        while [[ "$fullpath" =~ (/\./) ]]; do
            fullpath="${fullpath/${BASH_REMATCH[1]}//}"
        done
        # Remove sections that go backwards. e.g. /foo/bar/baz/../myfile.txt becomes /foo/bar/myfile.txt
        while [[ "$fullpath" =~ ([^/]+/\.\.(/|$)) ]]; do
            fullpath="${fullpath/${BASH_REMATCH[1]}/}"
        done
        # If there is still a /. at the end, remove it.
        fullpath="${fullpath%/.}"
        fullpaths+=( "$fullpath" )
    done
    if [[ "${#fullpaths[@]}" -eq '1' ]] && command -v "pbcopy" > /dev/null 2>&1; then
        printf %s "${fullpaths[@]}" | pbcopy
        printf '%s - copied to clipboard.\n' "${fullpaths[@]}"
        return 0
    fi
    printf '%s\n' "${fullpaths[@]}"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    fp "$@"
    exit $?
fi
unset sourced

return 0
