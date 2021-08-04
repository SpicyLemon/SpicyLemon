#!/bin/bash
# This file contains the json_diff function for getting a diff of JSON files.
# This file can be sourced to add the json_diff function to your environment.
# This file can also be executed to run the json_diff function without adding it to your environment.
#
# File contents:
#   json_diff  --> Function for diffing JSON files.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

json_diff () {
    local usage diff_cmd is_side_by_side use_json_info file1 file2 req_cmds req_cmd ccmd cf1 cf2 cre tempd file1p file2p sed_cmd_fmt exit_code
    usage="$( cat << EOF
Usage: json_diff [<diff args>] [--use-json-info|--use-jq] <file1> <file2>

    <file1> and <file2> are the filenames of the json to diff.
        These are always assumed to be the last two arguments provided.

   <diff args> are any arguments you want provided to the diff command.

    --use-json-info will use the json_info function to pre-process the json files.
        The resulting diff will then be on the output of json_info.
    --use-jq will use the jq program to pre-process the json files.
        The resulting diff will then be on the pretty-print version of the JSON.
        This is the default behavior.
    If both --use-json-info and --use-jq are provided, the last one provided is used.

EOF
    )"
    if [[ "$#" -eq '0' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    diff_cmd=( diff )
    while [[ "$#" -gt '2' ]]; do
        case "$1" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        --use-json-info)
            use_json_info='YES'
            ;;
        --use-jq)
            use_json_info=''
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
    file1="$1"
    file2="$2"
    shift
    shift
    if [[ "$file1" == '-h' || "$file1" == '--help' || "$file2" == '-h' || "$file2" == '--help' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi
    if [[ "$file1" == '' ]]; then
        printf 'json_diff: File 1 not provided.\n' >&2
        return 1
    elif [[ "$file2" == '' ]]; then
        printf 'json_diff: File 2 not provided.\n' >&2
        return 1
    fi
    if [[ "$#" -gt '0' ]]; then
        printf 'json_diff: Unknown arguments: %s\n' "$*" >&2
        return 1
    fi
    if [[ ! -f "$file1" ]]; then
        printf 'json_diff: File 1 not found: %s\n' >&2
        return 1
    elif [[ ! -f "$file2" ]]; then
        printf 'json_diff: File 2 not found: %s\n' >&2
        return 1
    fi
    req_cmds=( 'diff' 'mktemp' 'basename' )
    if [[ -n "$use_json_info" ]]; then
        req_cmds+=( 'json_info' )
    else
        req_cmds+=( 'jq' )
    fi
    for req_cmd in "${req_cmds[@]}"; do
        if ! command -v "$req_cmd" > /dev/null 2>&1; then
            printf 'json_diff: Missing required command: %s\n' "$req_cmd" >&2
            command "$req_cmd" >&2
            return $?
        fi
    done
    if [[ -t 1 ]]; then
        ccmd='\033[1;100m'  # Bold + dark gray background
        cf1='\033[93m'      # Bright yellow
        cf2='\033[92m'      # Bright green
        cre='\033[0m'       # Reset
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
        sed_cmd_fmt="s/^/$cf1/; s/(\t* +[|<>](\t|$)|\t+)/$cre\\\\1$cf2/; s/$/$cre/;"
    else
        sed_cmd_fmt="s/^(<.*)$/$cf1\\\\1$cre/; s/^(>.*)$/$cf2\\\\1$cre/;"
    fi

    # From this point on, need to delete tempd before returning.
    tempd="$( mktemp -d -t json_diff )" || return $?
    exit_code=$?
    if [[ "$exit_code" -ne '0' ]]; then
        printf 'json_diff: Unable to create temp directory for pre-processed JSON files.\n' >&2
    fi

    if [[ "$exit_code" -eq '0' ]]; then
        file1p="$tempd/1_$( basename "$file1" )"
        if [[ -n "$use_json_info" ]]; then
            json_info -r -f "$file1" > "$file1p"
            exit_code=$?
        else
            jq '.' "$file1" > "$file1p"
            exit_code=$?
        fi
        if [[ "$exit_code" -ne '0' ]]; then
            printf 'json_diff: Invalid JSON in file 1: %s\n' "$file1" >&2
        fi
    fi
    if [[ "$exit_code" -eq '0' ]]; then
        file2p="$tempd/2_$( basename "$file2" )"
        if [[ -n "$use_json_info" ]]; then
            json_info -r -f "$file2" > "$file2p"
            exit_code=$?
        else
            jq '.' "$file2" > "$file2p"
            exit_code=$?
        fi
        if [[ "$exit_code" -ne '0' ]]; then
            printf 'json_diff: Invalid JSON in file 2: %s\n' "$file2" >&2
        fi
    fi
    if [[ "$exit_code" -eq '0' ]]; then
        printf "$ccmd%s $cf1%s $cf2%s$cre\n" "${diff_cmd[*]}" "$file1" "$file2"
        "${diff_cmd[@]}" "$file1p" "$file2p" | sed -E "$( printf "$sed_cmd_fmt" )"
        exit_code="${PIPESTATUS[0]}${pipestatus[0]}"
    fi
    rm -rf "$tempd" > /dev/null 2>&1
    return $exit_code
}

if [[ "$sourced" != 'YES' ]]; then
    json_diff "$@"
    exit $?
fi
unset sourced

return 0
