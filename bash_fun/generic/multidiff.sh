#!/bin/bash
# This file contains the multidiff function that runs diffs on a set of files.
# This file can be sourced to add the multidiff function to your environment.
# This file can also be executed to run the multidiff function without adding it to your environment.
#
# File contents:
#   multidiff  --> Function for running diffs on a set o files.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

multidiff () {
    local usage cnums ccmd cn cf cre diff_cmd is_side_by_side files sed_cmd_fmt exit_code i j fi fj ec
    usage='Usage: multidiff [<diff flags> --] <file1> <file2> [<file3>...]\n'
    if [[ "$#" -eq '0' ]]; then
        printf '%s' "$usage"
        return 0
    fi
    if ! command -v 'seq' > /dev/null 2>&1; then
        printf 'Missing required command: seq\n' >&2
        command 'seq' >&2
        return $?
    fi
    if ! command -v 'diff' > /dev/null 2>&1; then
        printf 'Missing required command: diff\n' >&2
        command 'diff' >&2
        return $?
    fi
    cnums=( 93 92 96 95 91 97 33 32 36 35 31 37 )
    cf=()
    if [[ -t 1 ]]; then
        ccmd='\033[1;100m'  # Bold + dark gray background
        for cn in ${cnums[@]}; do
            cf+=( "\033[${cn}m" )
        done
        cre='\033[0m'
    else
        for cn in ${cnums[@]}; do
            cf+=( '' )
        done
    fi
    if command -v 'setopt' > /dev/null 2>&1; then
        setopt local_options KSH_ARRAYS
    fi
    diff_cmd=( diff )
    if [[ "$1" =~ ^- ]]; then
        while [[ "$#" -gt '0' ]]; do
            case "$1" in
            -h|--help)
                printf '%s' "$usage"
                return 0
                ;;
            --)
                shift
                break
                ;;
            -y|--side-by-side)
                is_side_by_side='YES'
                diff_cmd+=( "$1" )
                ;;
            *)
                diff_cmd+=( "$1" )
                ;;
            esac
            shift
        done
    fi
    files=( "$@" )
    if [[ "${#files[@]}" -eq '0' ]]; then
        printf 'No files provided. Did you forget the -- before the files?\n' >&2
        return 1
    elif [[ "${#files[@]}" -eq '1' ]]; then
        printf 'Only one file provided.\n' >&2
        return 1
    elif [[ "${#files[@]}" -gt "${#cnums[@]}" ]]; then
        printf 'Too many files. Max: %d, Found: %d\n' "${#files[@]}" "${#cnums[@]}" >&2
        return 1
    fi
    if [[ -n "$is_side_by_side" ]]; then
        # There's no super simple way to find the middle in the side-by-side diff.
        # I fiddled with trying to match/replace the whole line, but the greedy nature of sed makes it hard to get the middle.
        # Example problem diff line: '# \t\t- some notes\t\t\t# \t\t- some notes
        # The middle has the following pattern:
        #   Either
        #     any number of tabs
        #     followed by one or more spaces
        #     followed by | (changed), or < (added), or > (removed)
        #     followed by either a tab or the end of the line.
        #   Or
        #     One or more tabs
        # Example middles:
        #   Unchanged long line: '\t'
        #   Unchanged short line: '\t\t\t\t\t'
        #   Changed long line: '   |\t'
        #   Changed medium line: '\t   |\t'
        #   Changed short line: '\t\t\t\t   |\t'
        #   Added line: '\t\t\t\t\t\t\t   >\t'
        #   Removed long line: '   <'
        #   Removed medium line: '\t\t   <'
        #   Removed short line: '\t\t\t\t   <'
        # This messes up when the left file line has a tab, but it's about the best I'm gonna get.
        sed_cmd_fmt="s/^/%b/; s/(\t* +[|<>](\t|$)|\t+)/$cre\\\\1%b/; s/$/$cre/;"
    else
        sed_cmd_fmt="s/^(<.*)$/%b\\\\1$cre/; s/^(>.*)$/%b\\\\1$cre/;"
    fi
    exit_code=0
    for i in $( seq 0 $(( ${#files[@]} - 2 )) ); do
        for j in $( seq $(( $i + 1 )) $(( ${#files[@]} - 1 )) ); do
            fi="${files[$i]}"
            fj="${files[$j]}"
            printf "$ccmd%s ${cf[$i]}%s ${cf[$j]}%s$cre\n" "${diff_cmd[*]}" "$fi" "$fj"
            "${diff_cmd[@]}" "$fi" "$fj" \
                | sed -E "$( printf "$sed_cmd_fmt" "${cf[$i]}" "${cf[$j]}" )"
            ec="${PIPESTATUS[0]}${pipestatus[0]}"
            printf '\n'
            if [[ "$ec" -ne '0' ]]; then
                exit_code=$ec
            fi
        done
    done
    return $exit_code
}

if [[ "$sourced" != 'YES' ]]; then
    multidiff "$@"
    exit $?
fi
unset sourced

return 0
