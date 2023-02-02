#!/bin/bash
# This file contains the go_list_funcs function that lists all funcs in one or more files.
# This file can be sourced to add the go_list_funcs function to your environment.
# This file can also be executed to run the go_list_funcs function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_list_funcs () {
    local usage files color col_file col_func col_rcvr col_name
    usage="$( cat << EOF
Usage: go_list_funcs <files>

Any number of files can be provided.

Coloring can be controlled with the following env vars:
    GLF_NO_COLOR   - Set to anything (other than an empty string) to disable output coloring.
    GLF_FILE_COLOR - The color to use for the filename. The default is 36 (cyan).
    GLF_FUNC_COLOR - The color to use for the text "func". The default is 90 (dark gray).
    GLF_RCVR_COLOR - The color to use for the function receiver. The default is 95 (light-magenta).
    GLF_NAME_COLOR - The color to use for the function name. The default is 1 (bold).
    GLF_COLORS     - Four comma separated color values for (in order):
                        the filename, "func", the receiver, the function name.
                     Specific color env vars (e.g. GLF_NAME_COLOR) take
                     precidence over an entry in GLF_COLORS.
                     The default is '36,90,95,1'
EOF
)"
    files=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            -)
                files+=( $( cat - ) )
                ;;
            *)
                files+=( "$1" )
                ;;
        esac
        shift
    done

    if [[ "${#files[@]}" -eq '0' && ! -t 0 ]]; then
        files+=( $( cat - ) )
    fi

    if [[ "${#files[@]}" -eq '0' ]]; then
        printf '%s\n' "$usage"
        return 0
    fi

    color='always'
    if [[ -n "$GLF_NO_COLOR" ]]; then
        color='never'
    fi
    if [[ -n "$GLF_COLORS" ]]; then
        if [[ ! "$GLF_COLORS" =~ ^[[:digit:]\;]*,[[:digit:]\;]*,[[:digit:]\;]*,[[:digit:]\;]*$ ]]; then
            printf 'Invalid GLF_COLORS value "%s". Must be four comma delimited numbers. Ignoring it.\n' "$GLF_COLORS" >&2
        else
            col_file="$( printf '%s' "$GLF_COLORS" | cut -f 1 -d ',' )"
            col_func="$( printf '%s' "$GLF_COLORS" | cut -f 2 -d ',' )"
            col_rcvr="$( printf '%s' "$GLF_COLORS" | cut -f 3 -d ',' )"
            col_name="$( printf '%s' "$GLF_COLORS" | cut -f 4 -d ',' )"
        fi
    fi
    col_file="${GLF_FILE_COLOR:-$col_file}"
    col_func="${GLF_FUNC_COLOR:-$col_func}"
    col_rcvr="${GLF_RCVR_COLOR:-$col_rcvr}"
    col_name="${GLF_NAME_COLOR:-$col_name}"

    # Notes:
    # -o = --only-matching => Only output the text that matches.
    # -E = --extended-regexp => More complex matching than by default.
    # -s = --no-messages => Don't output an error message about files being dirs or not existing.
    # -H = --with-filename => Always include the filename in the output.
    # In the colorings, the |$ part of the pattern causes a line to be included even if it doesn't match what trying to coloring.

    grep -oEsH '^func( \([^)]+\))? [^[:space:](]+' "${files[@]}" \
        | GREP_COLOR="${col_file:-36}" grep --color=$color -E '^[^:]+|$' \
        | GREP_COLOR="${col_func:-90}" grep --color=$color -E 'func|$' \
        | GREP_COLOR="${col_rcvr:-95}" grep --color=$color -E '\([^)]*\)|$' \
        | GREP_COLOR="${col_name:-1}"  grep --color=$color -E '[^[:space:]]*$'

    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    go_list_funcs "$@"
    exit $?
fi
unset sourced

return 0
