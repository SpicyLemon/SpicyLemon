#!/bin/bash
# This file contains the indiff function that diffs two portions of a file.
# This file can be sourced to add the indiff function to your environment.
# This file can also be executed to run the indiff function without adding it to your environment.
#
# File contents:
#   indiff  --> Function for getting a diff of two portions of a file.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

indiff () {
    if ! command -v 'getlines' > /dev/null 2>&1; then
        printf 'Missing required command: getlines\n' >&2
        getlines
        return $?
    fi
    # If there weren't any args provided, pretend the --help flags was provided.
    if [[ "$#" -eq '0' ]]; then
        set -- --help
    fi
    local usage
    usage="$( cat << EOF
Usage: indiff <start1>-<end1> <start2>-<end2> <file>
   or: indiff <start1> <end1> <start2> <end2> <file>
   or: indiff (-l|--left) <start1>-<end1> (-r|--right) <start2>-<end2> (-f|--file) <file>
   or: indiff <file> --select

The <file> can be provided in any position in the args.
EOF
)"
    local range_rx verbose do_select l_start l_end r_start r_end filename line_max
    range_rx='^[[:digit:]]+[- ][[:digit:]]+$'
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                cat <<< "$usage"
                return 0
                ;;
            -v|--verbose)
                verbose="$1"
                ;;
            -f|--file|--filename)
                if [[ -n "$2" ]]; then
                    printf 'No argument provided after the %s flag.\n' "$1"
                    return 1
                fi
                filename="$2"
                shift
                ;;
            -s|--select)
                do_select="$1"
                ;;
            -l|--left)
                if [[ -n "$2" ]]; then
                    printf 'No argument provided after the %s flag.\n' "$1"
                    return 1
                fi
                if ! [[ "$2" =~ $range_rx ]]; then
                    printf 'Invalid %s option format [%s]. Expect <start1>-<end1>.\n' "$1" "$2"
                    return 1
                fi
                l_start="$( sed 's/[-[:blank:]].*$//' <<< "$2" )"
                l_end="$( sed 's/^.*[-[:blank:]]//' <<< "$2" )"
                shift
                ;;
            -r|--right)
                if [[ -n "$2" ]]; then
                    printf 'No argument provided after the %s flag.\n' "$1"
                    return 1
                fi
                if ! [[ "$2" =~ $range_rx ]]; then
                    printf 'Invalid %s option format [%s]. Expect <start2>-<end2>.\n' "$1" "$2"
                    return 1
                fi
                r_start="$( sed 's/[-[:blank:]].*$//' <<< "$2" )"
                r_end="$( sed 's/^.*[-[:blank:]]//' <<< "$2" )"
                shift
                ;;
            *)
                if [[ "$1" =~ $range_rx ]]; then
                    if [[ -z "$l_start" && -z "$l_end" ]]; then
                        l_start="$( sed 's/[-[:blank:]].*$//' <<< "$1" )"
                        l_end="$( sed 's/^.*[-[:blank:]]//' <<< "$1" )"
                    elif [[ -n "$l_start" && -n "$l_end" && -z "$r_start" && -z "$r_end" ]]; then
                        r_start="$( sed 's/[-[:blank:]].*$//' <<< "$1" )"
                        r_end="$( sed 's/^.*[-[:blank:]]//' <<< "$1" )"
                    else
                        printf 'Unknown argument: [%s].\n' "$1"
                        return 1
                    fi
                elif [[ "$1" =~ ^[[:digit:]]+$ ]]; then
                    if [[ -z "$l_start" ]]; then
                        l_start="$1"
                    elif [[ -z "$l_end" ]]; then
                        l_end="$1"
                    elif [[ -z "$r_start" ]]; then
                        r_start="$1"
                    elif [[ -z "$r_end" ]]; then
                        r_end="$1"
                    else
                        printf 'Unknown argument: [%s].\n' "$1"
                        return 1
                    fi
                elif [[ -z "$filename" ]]; then
                    filename="$1"
                else
                    printf 'Unknown argument: [%s].\n' "$1"
                    return 1
                fi
                ;;
        esac
        shift
    done

    if [[ -z "$filename" ]]; then
        printf 'No <file> provided.\n'
        cat <<< "$usage"
        return 1
    elif [[ ! -f "$filename" ]]; then
        printf 'File not found: %s\n' "$filename"
        return 1
    fi
    line_max="$( grep -c '^' "$filename" )"

    if [[ -n "$verbose" ]]; then
        printf 'filename="%s"\n' "$filename"
        printf 'line_max="%s"\n' "$line_max"
    fi

    if [[ -n "$do_select" ]]; then
        if [[ -n "$l_start" ]]; then
            printf 'Cannot provide both %s and <start1>: [%s]\n' "$do_select" "$l_start"
            return 1
        elif [[ -n "$l_end" ]]; then
            printf 'Cannot provide both %s and <end1>: [%s]\n' "$do_select" "$l_end"
            return 1
        elif [[ -n "$r_start" ]]; then
            printf 'Cannot provide both %s and <start2>: [%s]\n' "$do_select" "$r_start"
            return 1
        elif [[ -n "$r_end" ]]; then
            printf 'Cannot provide both %s and <end2>: [%s]\n' "$do_select" "$r_end"
            return 1
        fi

        local selected s_count
        [[ -n "$verbose" ]] && printf 'Selecting lines using fzf on file: [%s].\n' "$filename"
        selected="$( nl -b a "$filename" \
            | fzf --layout=reverse-list --multi --header 'Select exactly 4 lines for <start1> <end1> <start2> <end2>' \
            | sed -E 's/^[[:blank:]]*([[:digit:]]+)[[:blank:]].*$/\1/'
        )"
        s_count="$( printf '%s' "$selected" | grep -c '^' )"
        if [[ -n "$verbose" ]]; then
            printf 's_count="%s"\n' "$s_count"
            printf 'selected=%q\n' "$selected"
        fi
        if [[ "$s_count" -ne '4' ]]; then
            printf 'You selected %d lines, but must select exactly 4.\n' "$s_count"
            return 1
        fi
        selected="$( sed -E 's/^[[:blank:]]*([[:digit:]]+)[[:blank:]].*$/\1/' <<< "$selected" )"
        l_start="$( head -n 1 <<< "$selected" )"
        l_end="$( head -n 2 <<< "$selected" | tail -n 1 )"
        r_start="$( head -n 3 <<< "$selected" | tail -n 1 )"
        r_end="$( tail -n 1 <<< "$selected" )"
    fi

    if [[ -z "$l_start" ]]; then
        printf 'No <start1> provided.\n'
        cat <<< "$usage"
        return 1
    elif ! [[ "$l_start" =~ ^[[:digit:]]+$ ]]; then
        printf 'Invalid <start1>: [%s]. Must only be digits.\n' "$l_start"
        return 1
    elif [[ "$l_start" -gt "$line_max" ]]; then
        printf 'Excessive <start1>: %s. There are only %d lines in file: %s\n' "$l_start" "$line_max" "$filename"
        return 1
    fi

    if [[ -z "$l_end" ]]; then
        printf 'No <end1> provided.\n'
        cat <<< "$usage"
        return 1
    elif ! [[ "$l_end" =~ ^[[:digit:]]+$ ]]; then
        printf 'Invalid <end1>: [%s]. Must only be digits.\n' "$l_end"
        return 1
    elif [[ "$l_end" -gt "$line_max" ]]; then
        printf 'Excessive <end1>: %s. There are only %d lines in file: %s\n' "$l_end" "$line_max" "$filename"
        return 1
    elif [[ "$l_end" -lt "$l_start" ]]; then
        printf 'Inferior <end1>: %s. Cannot be less than <start1>: %s.\n' "$l_end" "$l_start"
        return 1
    fi

    if [[ -z "$r_start" ]]; then
        printf 'No <start2> provided.\n'
        cat <<< "$usage"
        return 1
    elif ! [[ "$r_start" =~ ^[[:digit:]]+$ ]]; then
        printf 'Invalid <start2>: [%s]. Must only be digits.\n' "$r_start"
        return 1
    elif [[ "$r_start" -gt "$line_max" ]]; then
        printf 'Excessive <start2>: %s. There are only %d lines in file: %s\n' "$r_start" "$line_max" "$filename"
        return 1
    fi

    if [[ -z "$r_end" ]]; then
        printf 'No <end2> provided.\n'
        cat <<< "$usage"
        return 1
    elif ! [[ "$r_end" =~ ^[[:digit:]]+$ ]]; then
        printf 'Invalid <end2>: [%s]. Must only be digits.\n' "$r_end"
        return 1
    elif [[ "$r_end" -gt "$line_max" ]]; then
        printf 'Excessive <end2>: %s. There are only %d lines in file: %s\n' "$r_end" "$line_max" "$filename"
        return 1
    elif [[ "$r_end" -lt "$r_start" ]]; then
        printf 'Inferior <end2>: %s. Cannot be less than <start2>: %s.\n' "$r_end" "$r_start"
        return 1
    fi

    if [[ -n "$verbose" ]]; then
        printf ' l_start="%s"\n' "$l_start"
        printf '   l_end="%s"\n' "$l_end"
        printf ' r_start="%s"\n' "$r_start"
        printf '   r_end="%s"\n' "$r_end"
    fi

    local t_dir l_range r_range basename l_file r_file l_full r_full c_dir ec d_cmd
    if ! t_dir="$( mktemp -d -t indiff.XXXX )"; then
        printf 'Could not create temp directory.\n'
        return 2
    fi
    [[ -n "$verbose" ]] && printf '   t_dir="%s"\n' "$t_dir"

    # From here on out, we need to make sure that the temp dir is deleted when we're done.

    l_range="${l_start}-${l_end}"
    r_range="${r_start}-${r_end}"

    basename="$( sed 's|^.*/||' <<< "$filename" )"
    l_file="${basename}_${l_range}"
    r_file="${basename}_${r_range}"
    l_file="lines ${l_range} of ${basename}"
    r_file="lines ${r_range} of ${basename}"
    l_full="${t_dir}/${l_file}"
    r_full="${t_dir}/${r_file}"
    c_dir="$( pwd )"

    if [[ -n "$verbose" ]]; then
        printf ' l_range="%s"\n' "$l_range"
        printf ' r_range="%s"\n' "$r_range"
        printf 'basename="%s"\n' "$basename"
        printf '  l_file="%s"\n' "$l_file"
        printf '  r_file="%s"\n' "$r_file"
        printf '  l_full="%s"\n' "$l_full"
        printf '  r_full="%s"\n' "$r_full"
        printf '   c_dir="%s"\n' "$c_dir"
    fi

    ec=0
    if [[ "$ec" -eq '0' ]]; then
        [[ -n "$verbose" ]] && printf 'Creating left file: getlines %q %q > %q\n' "$l_range" "$filename" "$l_full"
        if ! getlines "$l_range" "$filename" > "$l_full"; then
            printf 'Could not create left file: %s\n' "$l_full"
            ec=2
        fi
    fi
    if [[ "$ec" -eq '0' ]]; then
        [[ -n "$verbose" ]] && printf 'Creating right file: getlines %q %q > %q\n' "$r_range" "$filename" "$r_full"
        if ! getlines "$r_range" "$filename" > "$r_full"; then
            printf 'Could not create right file: %s\n' "$r_full"
            ec=2
        fi
    fi
    if [[ "$ec" -eq '0' ]]; then
        [[ -n "$verbose" ]] && printf 'Changing to temp dir: cd %q\n' "$t_dir"
        if ! cd "$t_dir"; then
            printf 'Could not change to temp dir: %s\n' "$t_dir"
            ec=2
        fi
    fi

    if [[ "$ec" -eq '0' ]]; then
        d_cmd='diff'
        if command -v colordiff > /dev/null 2>&1; then
            d_cmd='colordiff'
        fi
        [[ -n "$verbose" ]] && printf 'Doing diff: %q -U %q %q %q\n' "$d_cmd" "$line_max" "$l_file" "$r_file"
        printf '\033[1mindiff: %s\033[0m\n' "$filename"
        if "$d_cmd" -U "$line_max" "$l_file" "$r_file"; then
            printf '\033[93m(no differences)\033[0m\n'
        fi
    fi

    [[ -n "$verbose" ]] && printf 'Chang back to original dir: cd %q\n' "$c_dir"
    if ! cd "$c_dir"; then
        printf 'Could not change back to original dir: %s\n' "$c_dir"
        ec=5
    fi
    [[ -n "$verbose" ]] && printf 'Deleting temp dir: rm -rf %q\n' "$t_dir"
    if !  rm -rf "$t_dir"; then
        printf 'Could not delete temp dir: %s\n' "$t_dir"
        ec=5
    fi

    [[ -n "$verbose" ]] && printf 'Done. Returning %d\n' "$ec"
    return "$ec"
}

if [[ "$sourced" != 'YES' ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    require_command () {
        local cmd cmd_fn
        cmd="$1"
        if ! command -v "$cmd" > /dev/null 2>&1; then
            cmd_fn="$where_i_am/$cmd.sh"
            if [[ -f "$cmd_fn" ]]; then
                source "$cmd_fn"
                if [[ "$?" -ne '0' ]] || ! command -v "$cmd" > /dev/null 2>&1; then
                    ( printf 'This script relies on the [%s] function.\n' "$cmd"
                      printf 'The file [%s] was found and sourced, but there was a problem loading the [%s] function.\n' "$cmd_fn" "$cmd" ) >&2
                    return 1
                fi
            else
                ( printf 'This script relies on the [%s] function.\n' "$cmd"
                  printf 'The file [%s] was looked for, but not found.\n' "$cmd_fn" ) >&2
                return 1
            fi
        fi
    }
    require_command 'join_str' || exit $? # Needed by getlines.
    require_command 'getlines' || exit $?
    indiff "$@"
    exit $?
fi
unset sourced

return 0
