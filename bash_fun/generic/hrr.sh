#!/bin/bash
# This file contains the hrr function that outputs a colorful banner helpful for identifying key points in your terminal backlog.
# This file can be sourced to add the hrr, hhr, hr, and pick_a_palette functions to your environment.
# This file can also be executed to run the hrr function without adding it to your environment.
#
# File contents:
#   hr  --------------> Creates a single-line horizontal rule in the terminal with a message in it.
#   hrr  -------------> Creates a 3-line horizontal rule in the terminal with a message in it.
#   hhr  -------------> So you don't have to remember if it's hrr or hhr.
#   pick_a_palette  --> Sets the PALETTE environment variable if not already set.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Creates a single-line horizontal rule with a message in it.
# Usage: hr <message>
hr () {
    if ! command -v "tput" > /dev/null 2>&1; then
        printf 'Missing required command: tput\n' >&2
        tput
        return $?
    fi
    local message char termwidth available sixths leftover block empty section padding left_wing right_wing unset_palette
    message="$*"
    if [[ -n "$message" ]]; then
        message=" $message "
    fi
    char='#'
    termwidth=$( tput cols )
    available=$(( $termwidth - ${#message} - 2 ))
    sixths=$(( $available / 12 ))
    leftover=$(( $(( $available - $sixths * 12 )) / 2 ))
    block="$( printf '%0.1s' "$char"{1..500} )"
    empty="$( echo -E "$block" | sed "s/$char/ /g" )"
    section="${block:0:$sixths}"
    padding="${empty:0:$leftover}"
    left_wing=""
    right_wing=""
    pick_a_palette && unset_palette="Yup"
    for i in ${PALETTE[*]}; do
        new_piece="$( echo -E "\033[38;5;${i}m${section}\033[0m" )"
        left_wing="$left_wing$new_piece"
        right_wing="$new_piece$right_wing"
    done
    echo -e "$padding$left_wing\033[38;5;15m$message\033[0m$right_wing$padding"
    [[ -n "$unset_palette" ]] && unset PALETTE
}

# Creates a 3-line horizontal rule with a message in it.
# Usage: hrr <message>
hrr () {
    local message blank unset_palette
    message="$*"
    pick_a_palette && unset_palette="Yup"
    blank="$( hr )"
    echo -e "$blank"
    hr "$message"
    echo -e "$blank"
    [[ -n "$unset_palette" ]] && unset PALETTE
}

# So you don't have to remember if it's hrr or hhr.
hhr () {
    hrr "$@"
}

# Sets the PALETTE environment variable if it's not already set.
# Usage: pick_a_palette && echo "PALETTE set to ${PALETTE[*]}"
pick_a_palette () {
    if [[ -z "${PALETTE+x}" ]]; then
        local choice
        choice=$[RANDOM%8]
        case "$choice" in
            0) PALETTE=(232 236 240 244 248 252);;   #white
            1) PALETTE=(16 17 18 19 20 21);;         #blue
            2) PALETTE=(16 22 28 34 40 46);;         #green
            3) PALETTE=(16 64 106 148 184 226);;     #yellow
            4) PALETTE=(16 94 130 166 202 208);;     #orange
            5) PALETTE=(16 52 88 124 160 196);;      #red
            6) PALETTE=(16 54 92 129 165 206);;      #purple
            7) PALETTE=(16 $[RANDOM%256] $[RANDOM%256] $[RANDOM%256] $[RANDOM%256] $[RANDOM%256]);;
        esac
        return 0
    fi
    return 1
}

if [[ "$sourced" != 'YES' ]]; then
    hrr "$@"
    exit $?
fi
unset sourced

return 0
