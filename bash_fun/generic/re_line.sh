#!/bin/bash
# This file contains the re_line function that reformats delimited items.
# This file can be sourced to add the re_line function to your environment.
# This file can also be executed to run the re_line function without adding it to your environment.
#
# File contents:
#   re_line  --> Reformats delimited and/or line-separated entries.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

re_line () {
    local usage
    usage="$( cat << EOF
re_line - Reformats delimited and/or line-separated entries.

Usage: re_line [-f <filename>|--file <filename>|-c|--from-clipboard|-|-p|--from-pipe|-- <input>]
               [-n <count>|--count <count>|--min-width <width>|--max-width <width>]
               [-d <string>|--delimiter <string>] [-b <string>|--break <string>]
               [-w <string>|--wrap <string>] [-l <string>|--left <string>] [-r <string>|--right <string>]

    -f or --filename defines the file to get the input from.
    -c or --clipboard dictates that the input should be pulled from the clipboard.
    - or -p or --from-pipe indicates that the input is being piped in.
        This can also be expressed with -f - or --filename -.
    -- indicates that all remaining parameters are to be considered input.
    Exactly one of these input options must be provided.

    -n or --count defines the number of entries per line the output should have.
        Cannot be combined with --min-width or --max-width.
        A count of 0 indicates that the output should not have any line-breaks.
    --min-width defines the minimum line width (in characters) the output should have.
        Once an item is added to a line that exceeds this amount, a newline is then started.
        Cannot be combined with -n, --count or --max-width.
    --max-width defines the maximum line width (in characters) the output should have.
        If adding the next item to the line would exceed this amount, a newline is started
        and that next item is the first item on it.
        Note: A line can still exceed this width in cases where a single item exceeds this width.
        Cannot be combined with -n, --count or --min-width.
    If none of -n, --count, --min-width, or --max-width are provided, the default is -n 10.
    If more than one of -n, --count, --min-width or --max-width are provided, the one provided last is used.

    -d or --delimiter defines the delimiter to use for the output.
        The default (if not supplied) is a comma followed by a space.
    -b or --break defines the delimiter to use on each line of input.
        The default (if not supplied) is a comma and any following spaces.
        These are not considered to be part of any item, and will not be in the output.
        To turn off the splitting of each line, use -b ''.
        The string is used as the LHS of a sed s/// statement.
    -w or --wrap defines a string that will be added to both the beginning and end of each item.
    -l or --left defines a string that will be added to the left of each item.
        This is added after applying any -w or --wrap string.
    -r or --right defines a string that will be added to the right of each item.
        This is added after applying any -w or --wrap string.
EOF
)"
    # Pre-process all options/arguments/parameters up to a -- if provided.
    local verbose from_filename filename from_clipboard from_pipe from_args
    local per_line min_width max_width
    local delimiter_out delimiter_in wrap left right
    delimiter_out=', '
    delimiter_in=',[[:space:]]*'
    while [[ "$#" -gt '0' && -z "$from_args" ]]; do
        case "$1" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        -f|--file|--filename)
            if [[ -z "$2" ]]; then
                printf 'No filename provided with the %s option.\n' "$2" >&2
                return 1
            fi
            from_filename="$1 $2"
            filename="$2"
            shift
            ;;
        -c|--clipboard)
            from_clipboard="$1"
            ;;
        -|-p|--pipe|--from-pipe)
            from_pipe="$1"
            ;;
        --)
            from_args="$1"
            ;;
        -n|--count)
            if [[ -z "$2" ]]; then
                printf 'No count provided with the %s option.\n' "$2" >&2
                return 1
            fi
            per_line="$2"
            min_width=
            max_width=
            shift
            ;;
        --min|--min-width)
            if [[ -z "$2" ]]; then
                printf 'No width provided with the %s option.\n' "$2" >&2
                return 1
            fi
            per_line=
            min_width="$2"
            max_width=
            shift
            ;;
        --max|--max-width)
            if [[ -z "$2" ]]; then
                printf 'No width provided with the %s option.\n' "$2" >&2
                return 1
            fi
            per_line=
            min_width=
            max_width="$2"
            shift
            ;;
        -d|--delimiter|--delimiter-out)
            if [[ "$#" -eq '1' ]]; then
                printf 'No string provided with the %s option.\n' "$2" >&2
                return 1
            fi
            delimiter_out="$2"
            shift
            ;;
        -b|--break|--delimiter-in)
            if [[ "$#" -eq '1' ]]; then
                printf 'No string provided with the %s option.\n' "$2" >&2
                return 1
            fi
            delimiter_in="$2"
            shift
            ;;
        -w|--wrap)
            if [[ "$#" -eq '1' ]]; then
                printf 'No string provided with the %s option.\n' "$2" >&2
                return 1
            fi
            wrap="$2"
            shift
            ;;
        -l|--left)
            if [[ "$#" -eq '1' ]]; then
                printf 'No string provided with the %s option.\n' "$2" >&2
                return 1
            fi
            left="$2"
            shift
            ;;
        -r|--right)
            if [[ "$#" -eq '1' ]]; then
                printf 'No string provided with the %s option.\n' "$2" >&2
                return 1
            fi
            right="$2"
            shift
            ;;
        *)
            printf 'Unknown option provided: [%s].\n' "$1" >&2
            return 1
            ;;
        esac
        shift
    done

    # Some final validation on provided options.
    local input_source input_count
    input_source="$( echo $from_filename $from_clipboard $from_pipe $from_args )"
    input_count=0
    [[ -n "$from_filename" ]] && input_count=$(( input_count + 1 ))
    [[ -n "$from_clipboard" ]] && input_count=$(( input_count + 1 ))
    [[ -n "$from_pipe" ]] && input_count=$(( input_count + 1 ))
    [[ -n "$from_args" ]] && input_count=$(( input_count + 1 ))
    [[ -n "$verbose" ]] && printf 'input source: [%s], input count: [%d].\n' "$input_source" "$input_count" >&2
    if [[ "$input_count" -eq '0' ]]; then
        printf 'No input defined.\n' >&2
        return 1
    elif [[ "$input_count" -ne '1' ]]; then
        printf 'Too many inputs defined: [%s].\n' "$input_source"
        return 1
    fi
    if [[ -n "$filename" && "$filename" != '-' && ! -f "$filename" ]]; then
        printf 'File not found: [%s].\n' "$filename" >&2
        return 1
    fi

    # Set any defaults that still need to be set.
    if [[ -z "$per_line" && -z "$min_width" && -z "$max_width" ]]; then
        per_line=10
        [[ -n "$verbose" ]] && printf 'Using default line limiter: per_line=[%d]\n' "$per_line" >&2
    fi
    if [[ -n "$from_pipe" ]]; then
        filename='-'
    fi

    # Get the input and split it into entries.
    local zwnj input orig_ifs entries line
    #zero-width non-joiner: used to separate entries during processing
    zwnj="$( printf "\xe2\x80\x8c" )"
    if [[ -n "$from_filename" || -n "$from_pipe" ]]; then
        [[ -n "$verbose" ]] && printf 'Getting input using: [cat %s].\n' "$filename" >&2
        input="$( cat "$filename" )" || return $?
    elif [[ -n "$from_clipboard" ]]; then
        [[ -n "$verbose" ]] && printf 'Getting input using: [pbpaste].\n' >&2
        input="$( pbpaste )" || return $?
    elif [[ -n "$from_args" ]]; then
        [[ -n "$verbose" ]] && printf 'Getting input using: [$*].\n' >&2
        input="$*"
    fi
    if [[ -z "$input" ]]; then
        printf 'The requested input source was empty: [%s].\n' "$input_source" >&2
        return 1
    fi
    if [[ -n "$delimiter_in" ]]; then
        [[ -n "$verbose" ]] && printf 'Splitting each input line using: /%s/.\n' "$delimiter_in" >&2
        input="$( sed -E "s/$delimiter_in/$zwnj/g; s/$zwnj\$//;" <<< "$input" )"
    fi
    [[ -n "$verbose" ]] && printf 'Splitting input into entries.\n' >&2
    orig_ifs="$IFS"
    IFS="$zwnj"
    entries=( $( tr '\n' "$zwnj" <<< "$input" ) )
    IFS="$orig_ifs"

    # Do the output!
    local line line_entry_count entry_index total_entries del del_no_end_sp
    local del_width del_width_no_end_sp entry entry_width line_width
    line=
    line_entry_count='0'
    entry_index='0'
    total_entries="${#entries[@]}"
    del=
    del_no_end_sp=
    if [[ -n "$delimiter_out" ]]; then
        del="$delimiter_out"
        del_no_end_sp="$( sed -E 's/[[:space:]]+$//' <<< "$delimiter_out" )"
    fi
    # TODO: Try to figure out a good way of getting piped input through here in real-time.
    # I might need to turn this whole thing on its head, and do stuff similar to the psql_runner.sh scripts.
    # That way, I can have other functions without polluting the environment with functions that shouldn't be used outside of here.
    # For now, wrap all output in a subshell so it doesn't actually happen until it's all put together.
    # This allows piping this back into a re_line (or something else) a bit better.
    (
        if [[ -n "$verbose" ]]; then
            {
                printf 'There are [%d] entries.\n' "$total_entries"
                if [[ -n "$delimiter_out" ]]; then
                    printf 'Using output delimiter: [%s] (%d chars), which starts with [%s] (%d chars) and ends with %d whitespace character(s).\n' \
                                "$del" "${#del}" "$del_no_end_sp" "${#del_no_end_sp}" "$(( ${#del} - ${#del_no_end_sp} ))"
                else
                    printf 'Not using any delimiter for output.\n'
                fi
            } >&2
        fi
        if [[ -n "$per_line" ]]; then
            if [[ "$per_line" -eq '0' ]]; then
                [[ -n "$verbose" ]] && printf 'Outputting all entries on a single line.\n' >&2
                for entry in "${entries[@]}"; do
                    entry_index=$(( entry_index + 1))
                    if [[ "$entry_index" -ge "$total_entries" ]]; then
                        del=
                        del_no_end_sp=
                    fi
                    printf '%s' "${left}${wrap}${entry}${wrap}${right}${del}"
                done
            else
                [[ -n "$verbose" ]] && printf 'Limiting lines to [%d] entries per line.\n' "$per_line" >&2
                for entry in "${entries[@]}"; do
                    line_entry_count=$(( line_entry_count + 1 ))
                    entry_index=$(( entry_index + 1))
                    entry="${left}${wrap}${entry}${wrap}${right}"
                    if [[ "$entry_index" -ge "$total_entries" ]]; then
                        del=
                        del_no_end_sp=
                    fi
                    if [[ "$line_entry_count" -ge "$per_line" ]]; then
                        printf '%s\n' "${line}${entry}${del_no_end_sp}"
                        line=
                        line_entry_count=0
                    else
                        line="${line}${entry}${del}"
                    fi
                done
            fi
        elif [[ -n "$min_width" ]]; then
            [[ -n "$verbose" ]] && printf 'Starting a new line when the current one contains [%d] or more characters.\n' "$min_width" >&2
            del_width_no_end_sp="${#del_no_end_sp}"
            for entry in "${entries[@]}"; do
                entry_index=$(( entry_index + 1))
                entry="${left}${wrap}${entry}${wrap}${right}"
                entry_width="${#entry}"
                line_width="${#line}"
                if [[ "$entry_index" -ge "$total_entries" ]]; then
                    del=
                    del_no_end_sp=
                    del_width_no_end_sp=0
                fi
                if [[ "$(( line_width + entry_width + del_width_no_end_sp ))" -ge "$min_width" ]]; then
                    printf '%s\n' "${line}${entry}${del_no_end_sp}"
                    line=
                else
                    line="${line}${entry}${del}"
                fi
            done
        elif [[ -n "$max_width" ]]; then
            [[ -n "$verbose" ]] && printf 'Limiting lines to [%d] characters per line.\n' "$max_width" >&2
            del_width="${#delimiter_out}"
            del_width_no_end_sp="${#del_no_end_sp}"
            for entry in "${entries[@]}"; do
                entry_index=$(( entry_index + 1))
                entry="${left}${wrap}${entry}${wrap}${right}"
                entry_width="${#entry}"
                line_width="${#line}"
                if [[ "$entry_index" -ge "$total_entries" ]]; then
                    del=
                    del_width=0
                    del_no_end_sp=
                    del_width_no_end_sp=0
                fi
                if [[ "$(( line_width + entry_width + del_width ))" -ge "$max_width" ]]; then
                    if [[ -z "$line" ]]; then
                        printf '%s\n' "${entry}${del_no_end_sp}"
                        line=
                    elif [[ "$(( line_width + entry_width + del_width_no_end_sp ))" -le "$max_width" ]]; then
                        printf '%s\n' "${line}${entry}${del_no_end_sp}"
                        line=
                    else
                        sed -E 's/[[:space:]]+$//;' <<< "$line"
                        line="${entry}${del}"
                    fi
                else
                    line="${line}${entry}${del}"
                fi
            done
        fi
        if [[ -n "$line" ]]; then
            printf '%s\n' "$line"
        fi
    )
    [[ -n "$verbose" ]] && printf 'Done.\n' >&2
}

if [[ "$sourced" != 'YES' ]]; then
    re_line "$@"
    exit $?
fi
unset sourced

return 0
