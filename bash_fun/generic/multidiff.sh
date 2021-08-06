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
    local cnums usage diff_cmd is_side_by_side pre_processor files files_p files_count \
        cf ccmd cn cre sed_cmd_fmt req_cmds req_cmd exit_code tempd f fp ec i j
    # 93 = Bright Yellow, 92 = Bright Green, 96 = Bright Cyan, 95 = Bright Purple, 91 = Bright Red, 97 = Bright White
    # 33 = Yellow, 32 = Green, 36 = Cyan, 35 = Purple, 31 = Red, 37 = White
    cnums=( 93 92 96 95 91 97 33 32 36 35 31 37 )
    usage="$( cat << EOF
Gets differences between sets of files.

Usage: multidiff [[<diff args>] [--pre-process <pre-processor>] --] <file1> <file2> [<file3>...]

    <file1> <file2> [<file3>...] are the files to diff. Up to ${#cnums[@]} can be supplied.
        Diffs are done between each possible pair of files.
        For example, with 3 files, there are 3 pairs: 1-2, 1-3, 2-3.
        With 4 files, you would end up with 6 pairs: 1-2, 1-3, 1-4, 2-3, 2-4, 3-4.

    If any arguments other than files are provided, the files must all follow a -- argument.

    <diff args> are any arguments that you want provided to each diff command.
    --pre-process <pre-processor> defines any pre-processing that should be done to each file before the diff.
        <pre-processor> values:
            none       This is the default. Do not do any pre-processing of the files.
            jq         Apply the command  jq --sort-keys '.' <file>  to each file and get the differences of the results.
            json_info  Apply the command  json_info -r -f <file>     to each file and get the differences of the results.
EOF
    )"
    if [[ "$#" -eq '0' ]]; then
        printf '%s\n\n' "$usage"
        return 0
    fi
    if command -v 'setopt' > /dev/null 2>&1; then
        setopt local_options KSH_ARRAYS
    fi
    diff_cmd=( diff )
    if [[ "$1" =~ ^- ]]; then
        while [[ "$#" -gt '0' ]]; do
            case "$1" in
            -h|--help)
                printf '%s\n\n' "$usage"
                return 0
                ;;
            --pre-process|--pre-processor|--pre-proc|--pre)
                if [[ "$#" -lt '2' || -z "$2" ]]; then
                    printf 'multidiff: No pre-processor provided after %s\n' "$1" >&2
                    return 1
                fi
                case "$2" in
                none|--none)
                    pre_processor=''
                    ;;
                jq|--jq)
                    pre_processor='jq'
                    ;;
                json_info|json-info|--json-info|--json_info)
                    pre_processor='json_info'
                    ;;
                *)
                    printf 'multidiff: Unknown pre-processor value: %s\n' "$2" >&2
                    return 1
                esac
                shift
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
    files=()
    while [[ "$#" -gt '0' ]]; do
        if [[ -n "$1" ]]; then
            if [[ -f "$1" ]]; then
                files+=( "$1" )
            else
                printf 'multidiff: File not found: %s\n' "$1" >&2
            fi
        fi
        shift
    done
    files_count="${#files[@]}"
    if [[ "$files_count" -eq '0' ]]; then
        printf 'multidiff: No files provided. Did you forget the -- before the files?\n' >&2
        return 1
    elif [[ "$files_count" -eq '1' ]]; then
        printf 'multidiff: Only one file provided.\n' >&2
        return 1
    elif [[ "$files_count" -gt "${#cnums[@]}" ]]; then
        printf 'multidiff: Too many files. Max: %d, Found: %d\n' "${#cnums[@]}" "$files_count" >&2
        return 1
    fi
    cf=()
    if [[ -t 1 ]]; then
        ccmd='\033[1;100m'  # Bold + dark gray background
        for cn in ${cnums[@]}; do
            cf+=( "\033[${cn}m" )
        done
        cre='\033[0m'       # Color reset
    else
        for cn in ${cnums[@]}; do
            cf+=( '' )
        done
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

    req_cmds=( 'seq' 'diff' 'basename' 'mktemp' )
    if [[ -n "$pre_processor" ]]; then
        req_cmds+=( "$pre_processor" )
    fi
    for req_cmd in "${req_cmds[@]}"; do
        if ! command -v "$req_cmd" > /dev/null 2>&1; then
            printf 'multidiff: Missing required command: \n' "$req_cmd" >&2
            command "$req_cmd" >&2
            return $?
        fi
    done

    exit_code=0
    if [[ -z "$pre_processor" ]]; then
        files_p=( "${files[@]}" )
    else
        tempd="$( mktemp -d -t multidiff )"
        exit_code=$?
        if [[ "$exit_code" -ne '0' ]]; then
            printf 'multidiff: Unable to create temp directory for pre-processed files.\n' >&2
        elif [[ "$pre_processor" == 'jq' ]]; then
            files_p=()
            for i in $( seq 1 $files_count ); do
                j=$(( i - 1 ))
                f="${files[$j]}"
                printf -v fp '%s/%2d_%s' "$tempd" "$i" "$( basename "$f" )"
                files_p+=( "$fp" )
                jq --sort-keys '.' "$f" > "$fp"
                ec=$?
                if [[ "$ec" -ne '0' ]]; then
                    printf 'multidiff: Invalid JSON in file %d: %s\n' "$i" "$f" >&2
                    exit_code=$ec
                fi
            done
        elif [[ "$pre_processor" == 'json_info' ]]; then
            files_p=()
            for i in $( seq 1 $files_count ); do
                j=$(( i - 1 ))
                f="${files[$j]}"
                printf -v fp '%s/%2d_%s.info' "$tempd" "$i" "$( basename "$f" )"
                files_p+=( "$fp" )
                json_info --max-string 0 -r -f "$f" > "$fp"
                ec=$?
                if [[ "$ec" -ne '0' ]]; then
                    printf 'multidiff: Invalid JSON in file %d: %s\n' "$i" "$f" >&2
                    exit_code=$ec
                fi
            done
        else
            printf 'multidiff: Unknown pre-processor type: %s\n' "$pre_processor" >&2
            exit_code=1
        fi
    fi

    if [[ "$exit_code" -eq '0' ]]; then
        for i in $( seq 0 $(( files_count - 2 )) ); do
            for j in $( seq $(( i + 1 )) $(( files_count - 1 )) ); do
                printf "$ccmd%s ${cf[$i]}%s ${cf[$j]}%s$cre\n" "${diff_cmd[*]}" "${files[$i]}" "${files[$j]}"
                "${diff_cmd[@]}" "${files_p[$i]}" "${files_p[$j]}" \
                    | sed -E "$( printf "$sed_cmd_fmt" "${cf[$i]}" "${cf[$j]}" )"
                ec="${PIPESTATUS[0]}${pipestatus[0]}"
                printf '\n'
                if [[ "$ec" -ne '0' ]]; then
                    exit_code=$ec
                fi
            done
        done
    fi
    [[ -n "$tempd" && -d "$tempd" ]] && rm -rf "$tempd" > /dev/null 2>&1
    return $exit_code
}

if [[ "$sourced" != 'YES' ]]; then
    multidiff "$@"
    exit $?
fi
unset sourced

return 0
