#!/bin/bash
# This file contains several functions that output a colorful banner helpful for identifying key points in your terminal backlog.
# This file can be sourced to add the functions to your environment.
# This file can also be executed to run the hrr function without adding it to your environment.
#
# File contents:
#   hr  ---------------------> Creates a single-line horizontal rule in the terminal with a message in it.
#   hr1  --------------------> Similar to hr except padding is automatically added to any message.
#   hr3  --------------------> Creates a 3-line horizontal rule in the terminal with a message in it.
#   hrr  --------------------> Same as hr3. Only here for historical reasons. Deprecated
#   hhr  --------------------> Same as hr3. Only here for historical reasons. Deprecated
#   hr5  --------------------> Creates a 5-line horizontal rule in the terminal with a message in it.
#   hr7  --------------------> Creates a 7-line horizontal rule in the terminal with a message in it.
#   hr9  --------------------> Creates a 9-line horizontal rule in the terminal with a message in it.
#   hr11  -------------------> Creates an 11-line horizontal rule in the terminal with a message in it.
#   pick_a_palette  ---------> Sets the PALETTE environment variable if not already set.
#   what_palette_was_that  --> Prints out the last palette that was used.
#   show_all_palettes  ------> Uses hr1 to output all the different palettes available.
#   test_palette  -----------> Outputs hr1 through hr11 using a supplied palette and optional message.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Creates a single-line horizontal rule with a message in it.
# Usage: hr <message>
hr () {
    local message termwidth unset_palette available piece_len leftover char block section left_wing right_wing c
    message="$*"
    termwidth=80
    if [[ -n "$HR_WIDTH" ]]; then
        termwidth="$HR_WIDTH"
    elif command -v "tput" > /dev/null 2>&1; then
        termwidth=$( tput cols )
    fi
    pick_a_palette && unset_palette="Yup"
    available=$(( termwidth - ${#message} ))
    piece_len=$(( available / ${#PALETTE[@]} / 2 ))
    leftover=$(( available - piece_len * ${#PALETTE[@]} * 2 ))
    char='#'
    block="$( printf '%0.1s' "$char"{1..500} )"
    section="${block:0:$piece_len}"
    left_wing=''
    right_wing=''
    for c in ${PALETTE[@]}; do
        [[ "$leftover" -le '0' ]] && char=''
        right_wing="${right_wing}\033[38;5;${c}m${char}${section}\033[0m"
        leftover=$(( leftover - 1 ))
        [[ "$leftover" -le '0' ]] && char=''
        left_wing="\033[38;5;${c}m${char}${section}\033[0m${left_wing}"
        leftover=$(( leftover - 1 ))
    done
    printf '%b\033[38;5;15m%s\033[0m%b\n' "$left_wing" "$message" "$right_wing"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Similar to hr except if a message is provided, it's padded with some space.
# Usage: hr1 <message>
hr1 () {
    if [[ -n "$*" ]]; then
        hr " $* "
    else
        hr
    fi
    return 0
}

# Usage: hr3 <message>
hr3 () {
    local m unset_palette hrb hrm
    m="$*"
    pick_a_palette && unset_palette="Yup"
    hrb="$( hr )"
    if [[ -n "$m" ]]; then
        hrm="$( hr "  $m  " )"
    else
        hrm="$hrb"
    fi
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrm"
    printf '%b\n' "$hrb"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Usage: hrr <message>
hrr () {
    hr3 "$@"
}

# Usage: hhr <message>
hhr () {
    hr3 "$@"
}

# Usage: hr5 <message>
hr5 () {
    local m unset_palette hrb hrm
    m="$*"
    pick_a_palette && unset_palette="Yup"
    hrb="$( hr )"
    if [[ -n "$m" ]]; then
        hrm="$( hr "   $m   " )"
    else
        hrm="$hrb"
    fi
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrm"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Usage: hr7 <message>
hr7 () {
    local m unset_palette hrb hrm hrbs
    m="$*"
    pick_a_palette && unset_palette="Yup"
    hrb="$( hr  )"
    if [[ -n "$m" ]]; then
        hrm="$( hr "   $m   " )"
        hrbs="$( hr " $( sed 's/./ /g' <<< "$m" ) " )"
    else
        hrm="$hrb"
        hrbs="$hrb"
    fi
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrm"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Usage: hr9 <message>
hr9 () {
    local m unset_palette hrb hrm hrbs
    m="$*"
    pick_a_palette && unset_palette="Yup"
    hrb="$( hr )"
    if [[ -n "$m" ]]; then
        hrm="$( hr "    $m    " )"
        hrbs="$( hr "  $( sed 's/./ /g' <<< "$m" )  " )"
    else
        hrm="$hrb"
        hrbs="$hrb"
    fi
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrm"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Usage: hr11 <message>
hr11 () {
    local m unset_palette hrb hrm hrbs hrbss
    m="$*"
    pick_a_palette && unset_palette="Yup"
    hrb="$( hr )"
    if [[ -n "$m" ]]; then
        hrm="$( hr "     $m     " )"
        hrbs="$( hr "   $( sed 's/./ /g' <<< "$m" )   " )"
        hrbss="$( hr "$( sed 's/./ /g' <<< "${m:0:(( (${#m}-1)/3 + (${#m}-1)%3 + 1 ))}" )" )"
    else
        hrm="$hrb"
        hrbs="$hrb"
        hrbss="$hrb"
    fi
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrbss"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrm"
    printf '%b\n' "$hrbs"
    printf '%b\n' "$hrbss"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    printf '%b\n' "$hrb"
    [[ -n "$unset_palette" ]] && unset PALETTE
    return 0
}

# Usage: pick_a_palette [<choice>]
# Sets the PALETTE environment variable if it's not already set.
# An exit code of 0 means it has not been set yet, and you are in charge of unsetting it later.
# An exit code of 1 means that it's already set, so nothing is happening.
# The <choice> is optional and should be a number.
# If you have the palette_vector_no_wrap function available, <choice> can be a number from 0 to 295 (inclusive).
# Otherwise, some pre-generated ones will be used and it can be a number from 0 to 17 (inclusive).
# If palette_vector_generate is not available and a choice is provided out of the preset range,
#   then random numbers will be picked for the palette.
# If not provided, a <choice> will be chosen randomly.
pick_a_palette () {
    if [[ -n "$1" || -z "${PALETTE+x}" ]]; then
        local choice
        if command -v palette_vector_no_wrap > /dev/null 2>&1; then
            [[ -n "$1" ]] && choice="$1" || choice=$(( RANDOM%296 ))
            PALETTE=( $( palette_vector_no_wrap $choice ) )
        else
            [[ -n "$1" ]] && choice="$1" || choice=$(( RANDOM%18 ))
            case "$choice" in
                0) PALETTE=(232 236 240 244 248 252);;   # white --> black
                1) PALETTE=(252 248 244 240 236 232);;   # white <-- black
                2) PALETTE=(16 17 18 19 20 21);;         # blue --> black
                3) PALETTE=(21 20 19 18 17 16);;         # blue <-- black
                4) PALETTE=(16 22 28 34 40 46);;         # green --> black
                5) PALETTE=(46 40 34 28 22 16);;         # green <-- black
                6) PALETTE=(16 64 106 148 184 226);;     # yellow --> black
                7) PALETTE=(226 184 148 106 64 16);;     # yellow <-- black
                8) PALETTE=(16 94 130 166 202 208);;     # orange --> black
                9) PALETTE=(208 202 166 130 94 16);;     # orange <-- black
                10) PALETTE=(16 52 88 124 160 196);;     # red --> black
                12) PALETTE=(196 160 124 88 52 16);;     # red <-- black
                11) PALETTE=(16 54 92 129 165 206);;     # purple --> black
                13) PALETTE=(206 165 129 92 54 16);;     # purple <-- black
                14) PALETTE=(201 206 211 216 221 226);;  # purple --> yellow
                15) PALETTE=(226 221 216 211 206 201);;  # purple <-- yellow
                16) PALETTE=(51 80 109 138 167 196);;    # cyan --> red
                17) PALETTE=(196 167 138 109 80 51);;    # cyan <-- red
                *) PALETTE=($(( 16 + RANDOM%216 )) $(( 16 + RANDOM%216 )) $(( 16 + RANDOM%216 )) $(( 16 + RANDOM%216 )) $(( 16 + RANDOM%216 )) $(( 16 + RANDOM%216 )));;
            esac
        fi
        LAST_CHOICE="$choice"
        LAST_PALETTE=( ${PALETTE[@]} )
        return 0
    fi
    return 1
}

# Usage: what_palette_was_that
what_palette_was_that () {
    if [[ -n "${LAST_PALETTE+x}" ]]; then
        [[ -n "$LAST_CHOICE" ]] && printf '%3d: ' "$LAST_CHOICE"
        printf '%s\n' "${LAST_PALETTE[*]}"
    fi
}

# Usage: show_all_palettes
show_all_palettes () {
    local p maxp
    command -v palette_vector_no_wrap > /dev/null 2>&1 && maxp=295 || maxp=17
    for p in $( seq 0 $maxp ); do
        pick_a_palette "$p"
        hr1 "$( printf '%3d:' "$p"; printf ' %3d' "${PALETTE[@]}"; )"
    done
    unset PALETTE
    return 0
}

# Usage: test_palette <col1> <col2> <col3> <col4> <col5> <col6> [<message>]\n
# The first 6 arguments must be the palette numbers (0 to 255).
# If no <message> is provided, the palette numbers will be used as the message.
test_palette () {
    if [[ "$#" -lt '6' ]]; then
        printf 'Usage: test_palette <col1> <col2> <col3> <col4> <col5> <col6> [<message>]\n' >&2
        return 1
    fi
    PALETTE=( $1 $2 $3 $4 $5 $6 )
    if [[ "$#" -gt '6' ]]; then
        shift; shift; shift; shift; shift; shift;
    fi
    hr1 "$*"
    printf '\n'
    hr3 "$*"
    printf '\n'
    hr5 "$*"
    printf '\n'
    hr7 "$*"
    printf '\n'
    hr9 "$*"
    printf '\n'
    hr11 "$*"
    unset PALETTE
    return 0
}


if [[ "$sourced" != 'YES' ]]; then
    hrr "$@"
    exit $?
fi
unset sourced

return 0
