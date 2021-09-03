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
    local cnums usage diff_cmd is_side_by_side pre_processor replstr files files_p files_count \
        is_pd_pp cf ccmd cn cre sed_cmd_fmt req_cmds req_cmd exit_code tempd f fp pp_cmd ec i j
    # 93 = Bright Yellow, 92 = Bright Green, 96 = Bright Cyan, 95 = Bright Purple, 91 = Bright Red, 97 = Bright White
    # 33 = Yellow, 32 = Green, 36 = Cyan, 35 = Purple, 31 = Red, 37 = White
    cnums=( 93 92 96 95 91 97 33 32 36 35 31 37 )
    usage="$( cat << EOF
Gets differences between sets of files.

Usage: multidiff [[<diff args>] [--pre-process <pre-processor-cmd> [--replstr <replstr>]] --] <file1> <file2> [<file3>...]

    If any arguments other than files are provided, the files must all follow a -- argument.

    <file1> <file2> [<file3>...] are the files to diff. Up to ${#cnums[@]} can be supplied.
        Diffs are done between each possible pair of files.
        For example, with 3 files, there are 3 pairs: 1-2, 1-3, 2-3.
        With 4 files, you would end up with 6 pairs: 1-2, 1-3, 1-4, 2-3, 2-4, 3-4.

    <diff args> are any arguments that you want provided to each diff command.
    --pre-process <pre-processor-cmd> defines any pre-processing that should be done to each file before the diff.
        The <pre-processor-cmd> will be run for each provided file.
        The result will be stored in a temp file which will then be used for the diffs.
        By default, the filename will be added to the end of the <pre-processor-cmd>.
        If --replstr <replstr> is provided, the <pre-processor-cmd> should contain <replstr>,
            and the first instance of it will be replaced with the filename.
            If <replstr> is an empty string, file placement goes back to default behavior.
            Suggested <replstr> values: '{}', 'FFFF'
        There are some pre-defined <pre-processor-cmd>s that can be provided:
            'jq': same as 'jq --sort "."'
            'json-info': same as 'json_info --max-string 0 -r -f'
        Example of args for pre-defined pre-processor: --pre-process jq
        An empty string for <pre-processor-cmd> will deactivate pre-processing.
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
            --pre|--pre-proc|--pre-process|--pre-processor)
                if [[ "$#" -lt '2' || "$2" == '--' ]]; then
                    printf 'multidiff: No pre-processor provided after %s\n' "$1" >&2
                    return 1
                fi
                pre_processor="$2"
                shift
                ;;
            --replstr)
                if [[ "$#" -lt '2' || "$2" == '--' ]]; then
                    printf 'multidiff: No replstr provided after %s\n' "$1" >&2
                    return 1
                fi
                replstr="$2"
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

    is_pd_pp='YES'
    case "$pre_processor" in
        jq|--jq) pre_processor="jq --sort-keys '.'";;
        json|json_info|json-info|--json|--json_info|--json-info) pre_processor='json_info --max-string 0 -r -f';;
        *) is_pd_pp='';;
    esac
    if [[ -n "$replstr" ]]; then
        if [[ -z "$pre_processor" ]]; then
            printf 'multidiff: A replstr cannot be provided without a pre-processor-cmd.\n' >&2
            return 1
        elif [[ -n "$is_pd_pp" ]]; then
            printf 'multidiff: A replstr cannot be provided with a pre-defined pre-processor.\n' >&2
            return 1
        elif ! grep -q "$replstr" <<< "$pre_processor" > /dev/null 2>&1; then
            printf 'multidiff: The replstr %s was not found in the pre-processor command: %s\n' "$replstr" "$pre_processor" >&2
            return 1
        fi
    fi

    req_cmds=( 'seq' 'diff' 'basename' 'mktemp' )
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
        else
            files_p=()
            for i in $( seq 1 $files_count ); do
                j=$(( i - 1 ))
                f="${files[$j]}"
                printf -v fp '%s/%02d_%s.pp' "$tempd" "$i" "$( basename "$f" )"
                files_p+=( "$fp" )
                if [[ -z "$replstr" ]]; then
                    pp_cmd="$pre_processor '$f'"
                else
                    pp_cmd="$( sed "s/$replstr/'$f'/" <<< "$pre_processor" )"
                fi
                eval $pp_cmd > "$fp"
                ec=$?
                if [[ "$ec" -ne '0' ]]; then
                    printf 'multidiff: Pre-processing command failed for file %d: %s\n' "$i" "$pp_cmd" >&2
                    exit_code=$ec
                fi
            done
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
