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
# Usage: fp [-LP]
#   or   fp [-LP] <filename 1> [<filename 2> ...]
#   or   <stuff> | fp - [-LP]
# The -P and -L flags are provided to pwd if provided here, and the directory in question exists.
fp () {
    local filenames pwdargs selections filename fullpaths fullpath justdir
    filenames=()
    pwdargs=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -L|-P) pwdargs+=( "$1" );;
            -) filenames+=( $( cat - ) );;
            *) filenames+=( "$1" );;
        esac
        shift
    done
    if [[ "${#filenames[@]}" -eq '0' ]]; then
        if command -v 'fzf' > /dev/null 2>&1; then
            selections="$( ls -a | sort -f | fzf -m --tac --cycle )"
            if [[ -z "$selections" ]]; then
                printf 'No filenames selected.\n' >&2
                return 1
            fi
            while IFS= read line; do
                filenames+=( "$line" )
            done <<< "$selections"
        else
            printf 'No filenames provided.\n' >&2
            return 1
        fi
    fi
    fullpaths=()
    for filename in "${filenames[@]}"; do
        if [[ "$filename" == '/' ]]; then
            # The root / directory is a special case that doesn't need extra stuff.
            # And without this special handling, requires extra hoops in the rest of the stuff.
            fullpaths+=( '/' )
            continue
        elif [[ "$filename" =~ ^/ ]]; then
            fullpath="$filename"
        else
            fullpath="$( pwd )/$filename"
        fi
        # If the fullpath is a directory, it's a little easier.
        if [[ -d "$fullpath" ]]; then
            fullpaths+=( "$( cd "$fullpath"; pwd ${pwdargs[*]} )" )
        else
            # It's either a file, or doesn't exist, do some legwork.

            # Split it into the last part and the directory holding it.
            justfile="$( basename "$fullpath" )"
            justdir="$( dirname "$fullpath" )"
            # If the directory actually exists, simplify it.
            [[ -d "$justdir" ]] && justdir="$( cd "$justdir"; pwd ${pwdargs[*]} )"
            # Make sure it ends in a slash. It'll only not end in a slash if it's exactly '/'. E.g. when filename = /bin
            [[ ! "$justdir" =~ /$ ]] && justdir="$justdir/"
            # Put it back to gether and move on.
            fullpaths+=( "${justdir}${justfile}" )
        fi
    done
    if [[ "${#fullpaths[@]}" -eq '1' ]] && command -v "pbcopy" > /dev/null 2>&1; then
        printf '%s' "${fullpaths[@]}" | pbcopy
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
