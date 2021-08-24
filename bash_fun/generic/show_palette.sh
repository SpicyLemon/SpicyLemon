#!/bin/bash
# This file contains the show_palette function that display color palettes.
# This file can be sourced to add the show_palette function to your environment.
# This file can also be executed to run the show_palette function without adding it to your environment.
#
# File contents:
#   show_palette  --> Displays color palettes.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: show_palette [[-f] <fg col1> [<fg col2> ...]] [-b <bg col1> [<bg col2> ...]] [-a|--all] [-t <sample text>]
#    or: show_palette [<pair1> [<pair2> ...]] [-a|--all] [-t <sample text>]
show_palette () {
    local usage fgs bgs do_all is_bg text had_pair had_nonpair c fgc bgc maxc f b fcol bcol w hw lp rp
    usage="$( cat << EOF
Displays color palettes in your terminal.

Usage: show_palette [[-f] <fg col1> [<fg col2> ...]] [-b <bg col1> [<bg col2> ...]] [-a|--all] [-t <sample text>]
   or: show_palette [<pair1> [<pair2> ...]] [-a|--all] [-t <sample text>]

    The colors must be numbers between 0 and 255 inclusive.
        0 to 15 are some standard colors.
        16 to 231 are gradiented colors.
        231 to 255 are a black to white gradient.

    Desired colors can be provided in one of two ways:
        1: Single color entries.
            Any numbers provided first or after a -f flag are foreground colors.
            Any numbers provided after a -b flag are background colors.
            The -f and -b flags can be provided as many times as needed.
            If no foreground colors are provided, 7 is used.
            If no background colors are provided, 0 is used.
           E.g. show_palette 16 54 92 124 162 200 -b 27
           E.g. show_palette -b 200 -f 16 -b 162 -f 54 -b 124 -f 92
        2: Pairs of fg,bg entries.
            Each pair should be provided in a single argument.
            The foreground color should be first, then a comma, then the background color.
           E.g. show_palette '16,27' '54,27' '92,27' '124,27' '162,27' '200,27'
           E.g. show_palette '16,200' '54,162' '92,124'

    By default a number of lines are printed with the fg,bg color pair first (in the terminal default),
        followed by some sample text in that color combination.
        The 1st foreground color is paired with the 1st background color and printed.
        Then the 2nd foreground color is paired with the 2nd background color and printed.
        And so on.
        If an unequal number of foregrounds and backgrounds are provided,
        the smaller list cycles until the larger list has been completely shown.
        The sample text can be changed using the -t option.

    If the -a or --all flag is provided, a grid of all combinations of foreground and background colors is printed.
        Column and row headers are printed in the terminal default.
        The columns are the background colors.
        The rows are the foreground colors.
        The default sample text in this mode is the fg,bg pair.
        This can be changed to static supplied text using the -t option.

EOF
    )"
    if [[ "$#" -eq '0' ]]; then
        printf '%s\n\n' "$usage"
        return 0
    fi
    if command -v 'setopt' > /dev/null 2>&1; then
        setopt local_options BASH_REMATCH KSH_ARRAYS
    fi
    fgs=()
    bgs=()
    text=''
    while [[ "$#" -gt '0' ]]; do
        case "$( tr '[:upper:]' '[:lower:]' <<< "$1" )" in
            -h|--help)
                printf '%s\n\n' "$usage"
                return 0
                ;;
            -a|-all)
                do_all="$1"
                ;;
            -b)
                is_bg="$1"
                ;;
            -f)
                is_bg=''
                ;;
            -t|--text|--sample|--sample-text)
                if [[ -z "$2" ]]; then
                    printf 'No text argument supplied after [%s] flag.' "$1" >&2
                    return 0
                fi
                text="$2"
                shift
                ;;
            --)
                shift
                text="$*"
                break
                ;;
            *)
                if [[ "$1" =~ [,.[:space:]] ]]; then
                    had_pair='YES'
                    fgs+=( "$( sed 's/[,._[:space:]].*$//' <<< "$1" )" )
                    bgs+=( "$( sed 's/^[^,._[:space:]]*[,.[:space:]]//' <<< "$1" )" )
                elif [[ -z "$is_bg" ]]; then
                    had_nonpair='YES'
                    fgs+=( "$1" )
                else
                    had_nonpair='YES'
                    bgs+=( "$1" )
                fi
                ;;
        esac
        shift
    done
    if [[ -z "$had_pair" && -z "$had_nonpair" ]]; then
        printf 'No colors provided.\n' >&2
        return 1
    fi
    if [[ -n "$had_pair" && -n "$had_nonpair" ]]; then
        printf 'Cannot provide mix of fg,bg pairs and singular fb or bg values.\n' >&2
        return 1
    fi
    for c in "${fgs[@]}" "${bgs[@]}"; do
        if [[ ! "$c" =~ ^[[:digit:]]+$ || "$c" -gt 255 ]]; then
            printf 'Invalid color: [%s]. Must be a number between 0 and 255 inclusive.\n' "$c" >&2
            return 1
        fi
    done
    if [[ "${#fgs[@]}" -eq '0' ]]; then
        fgs+=( '7' )
    fi
    if [[ "${#bgs[@]}" -eq '0' ]]; then
        bgs+=( '0' )
    fi

    if [[ -n "$do_all" ]]; then
        if [[ -n "$text" ]]; then
            if [[ "${#text}" -lt 5 ]]; then
                hw="${#text}"
                printf -v text "%$(( (5 - hw) / 2 + (5 - hw) % 2 ))s%s%$(( (5 - hw) / 2 ))s" '' "$text" ''
            fi
            w="${#text}"
        fi
        printf '     '
        for bcol in "${bgs[@]}"; do
            if [[ -z "$text" ]]; then
                printf '     %-3d ' "$bcol"
            else
                hw="${#bcol}"
                lp="$(( (w - hw) / 2 + (w - hw) % 2 ))"
                rp="$(( (w - hw) / 2 ))"
                printf "%${lp}s%s%${rp}s" '' "$bcol" ''
            fi
        done
        printf '\n'
        for fcol in "${fgs[@]}"; do
            printf '%3d: ' "$fcol"
            for bcol in "${bgs[@]}"; do
                printf '\033[38;5;%d;48;5;%dm' "$fcol" "$bcol"
                if [[ -n "$text" ]]; then
                    printf '%s' "$text"
                else
                    printf ' %3d,%-3d ' "$fcol" "$bcol"
                fi
                printf '\033[0m'
            done
            printf '\n'
        done
        return 0
    fi

    # Done showing all if that's what we were doing.
    if [[ -z "$text" ]]; then
        text=' The quick brown fox jumps over the lazy dog. '
    fi
    fgc="${#fgs[@]}"
    bgc="${#bgs[@]}"
    if [[ "$fgc" -gt "$bgc" ]]; then
        maxi="$(( fgc - 1 ))"
    else
        maxi="$(( bgc - 1 ))"
    fi
    for c in $( seq 0 "$maxi" ); do
        f="$(( c % fgc ))"
        b="$(( c % bgc ))"
        fcol="${fgs[$f]}"
        bcol="${bgs[$b]}"
        printf '%3d,%-3d : \033[38;5;%d;48;5;%dm%s\033[0m\n' "$fcol" "$bcol" "$fcol" "$bcol" "$text"
    done
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    show_palette "$@"
    exit $?
fi
unset sourced

return 0
