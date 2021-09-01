#!/bin/bash
# This file contains the pretty_json function that uses jq to make json look pretty.
# This file can be sourced to add the pretty_json function to your environment.
# This file can also be executed to run the pretty_json function without adding it to your environment.
#
# File contents:
#   pretty_json  --> Takes in some json and makes it pretty.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

pretty_json () {
    if ! command -v "jq" > /dev/null 2>&1; then
        printf 'Missing required command: jq\n' >&2
        jq
        return $?
    fi
    local usage
    usage="$( cat << EOF
pretty_json - Makes Json Pretty.

Usage: pretty_json [-q|--quiet] [-c|--clipboard] [-s <file>|--save <file>] [-f <file>|--file <file>|-|-- <json>]

    If none of -f, --file, -, or -- are provided, pbpaste will be used if available.
    At most, only one of -f, --file, -, or -- can be provided.

    -q or --quiet will supress normal stdout output.
    -c or --clipboard will cause the output to be placed in the clipboard.
    -s or --save will cause the output to be written to the provided file.
        If also using the -f or --file option, the provided file can be ommitted, and
        output will go to a file with a name based on the one provided with -f or --file.
        If the input filename has .ugly -ugly or _ugly in it, then that will be changed
        to .pretty -pretty or _pretty for the output filename.
        Otherwise, -pretty will be added to the input filename just before the first period,
        or at the end of the filename if there is no period.
    -f or --file will read the json from the provided file.
    - indicates that the json is being piped in.
    -- indicates the end of parameters, and anything following is treated as json.

EOF
)"
    local keep_quiet to_clipboard to_file output_filename from_file input_filename from_pipe from_args json_in
    local last_was_save this_can_be_output_filename last_was_file this_can_be_input_filename
    while [[ "$#" -gt '0' ]]; do
        if [[ -n "$last_was_save" ]]; then
            last_was_save=
            this_can_be_output_filename='YES'
        fi
        if [[ -n "$last_was_file" ]]; then
            last_was_file=
            this_can_be_input_filename='YES'
        fi
        case "$1" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        -q|--quiet)
            keep_quiet="$1"
            ;;
        -c|--clipboard)
            if command -v 'pbcopy' > /dev/null 2>&1; then
                to_clipboard="$1"
            else
                printf 'Ignoring option [%s] because the command [pbcopy] is not available.\n' "$1" 2>&1
            fi
            ;;
        -s|--save)
            to_file="$1"
            last_was_save="$1"
            ;;
        -f|--file)
            from_file="$1"
            last_was_file="$1"
            ;;
        -)
            from_pipe='YES'
            ;;
        --)
            from_args='YES'
            ;;
        *)
            if [[ -n "$this_can_be_output_filename" ]]; then
                output_filename="$1"
            elif [[ -n "$this_can_be_input_filename" ]]; then
                input_filename="$1"
            else
                printf 'Unknown option: [%s].\n' "$1" >&2
                return 1
            fi
            ;;
        esac
        shift
        this_can_be_output_filename=
        this_can_be_input_filename=
        [[ -n "$from_args" ]] && break
    done
    if [[ -n "$to_file" && -z "$output_filename" ]]; then
        if [[ -z "$input_filename" ]]; then
            printf 'No output filename provided with the %s option.\n' "$to_file" >&2
            return 1
        fi
        if [[ "$input_filename" =~ [-._]ugly ]]; then
            output_filename="$( sed -E 's/([-._])ugly/\1pretty/' <<< "$input_filename" )"
        else
            output_filename="$( sed -E 's/(\.)|$/-pretty\1/' <<< "$input_filename" )"
        fi
    fi
    if [[ -n "$from_file" ]]; then
        if [[ -z "$input_filename" ]]; then
            printf 'No input filename provided with the %s option.\n' "$from_file" >&2
            return 1
        elif [[ ! -f "$input_filename" ]]; then
            printf 'File not found: [%s].' "$input_filename" >&2
            return 1
        fi
        json_in="$( cat "$input_filename" )"
    elif [[ -n "$from_pipe" ]]; then
        json_in="$( cat - )"
    elif [[ -n "$from_args" ]]; then
        json_in="$*"
    elif command -v 'pbpaste' > /dev/null 2>&1; then
        json_in="$( pbpaste )"
    else
        printf 'No input provided.\n' >&2
        return 1
    fi
    local jq_exit color_output normal_output
    if [[ -z "$keep_quiet" ]]; then
        color_output="$( jq --sort-keys -C '.' <<< "$json_in" )"
        jq_exit="$?"
        if [[ -n "$to_file" || -n "$to_clipboard" ]]; then
            normal_output="$( sed -E "s/$( echo -e "\033" )\[[[:digit:]]+(;[[:digit:]]+)*m//g" <<< "$color_output" )"
        fi
    else
        normal_output="$( jq --sort-keys '.' <<< "$json_in" )"
        jq_exit="$?"
    fi
    if [[ -z "$keep_quiet" ]]; then
        printf '%b\n' "$color_output"
    fi
    if [[ -n "$to_file" ]]; then
        printf '%s' "$normal_output" > "$output_filename"
        printf 'Pretty json saved in file [%s].\n' "$output_filename" >&2
    fi
    if [[ -n "$to_clipboard" ]]; then
        pbcopy <<< "$normal_output"
        printf 'Pretty json copied to clipboard.\n' >&2
    fi
    return "$jq_exit"
}

if [[ "$sourced" != 'YES' ]]; then
    pretty_json "$@"
    exit $?
fi
unset sourced

return 0
