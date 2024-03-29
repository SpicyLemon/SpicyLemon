#!/bin/bash
# This file contains the to_epoch function that converts a date and time into milliseconds since the epoch.
# This file can be sourced to add the to_epoch function to your environment.
# This file can also be executed to run the to_epoch function without adding it to your environment.
#
# File contents:
#   to_epoch  --> Converts a date in YYYY-mm-dd HH:MM:SS format (using local time zone) to an epoch as milliseconds.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Convert a date and time into an epoch as milliseconds.
# Usage: to_epoch yyyy-MM-dd [HH:mm[:ss[.ddd]]] [(+|-)HHmm]
#  or    to_epoch now
to_epoch () {
    local pieces the_date the_time the_time_zone s_fractions ms_fractions ms epoch_s epoch_ms
    if [[ -z "$1" || "$1" == "-h" || "$1" == "--help" ]]; then
        printf 'Usage: to_epoch yyyy-MM-dd [HH:mm[:ss[.ddd]]] [(+|-)HHmm]\n'
        return 0
    fi
    if [[ "$1" == "now" ]]; then
        date '+%s000'
        return 0
    fi
    # Allow for the input to be in ISO 8601 format where the date and time are combined with a T.
    # Also allow for ISO 8601 time zone Z and/or no space between the time and offset.
    pieces=( $( sed -E 's/T/ /; s/([-+][[:digit:]]+)$/ \1/; s/Z$/ +0000/;' <<< "$*" ) )
    # zsh is 1 indexed, bash is 0.
    if [[ -n "${pieces[0]}" ]]; then
        the_date="${pieces[0]}"
        the_time="${pieces[1]}"
        the_time_zone="${pieces[2]}"
    else
        the_date="${pieces[1]}"
        the_time="${pieces[2]}"
        the_time_zone="${pieces[3]}"
    fi
    # Since $the_time is optional, if it starts with a + or -,
    # it's actually the time zone piece.
    if [[ "$the_time" =~ ^[+-] ]]; then
        the_time_zone="$the_time"
        the_time=
    fi
    # Try to make $the_date into yyyy-MM-dd format.
    # Allow for input to be in the formats yyyy, yyyyMM, yyyy-MM, yyyyMMdd, yyyyMM-dd, yyyy-MMdd, yyyy-MM-dd,
    # or MM-dd-yyyy
    # or have different delimiters.
    # Note: Can't use a herestring here because they add a final \n, which would then get changed to a - by tr.
    the_date="$( printf '%s' "$the_date" | tr -c "[:digit:]" "-" )"
    if [[ "$the_date" =~ ^[[:digit:]]{4}(-?[[:digit:]]{2}){0,2}$ ]]; then
        the_date="$( tr -d '-' <<< "$the_date" | sed 's/$/0101/' | head -c 8 | sed -E 's/^(....)(..)(..)$/\1-\2-\3/' )"
    elif [[ "$the_date" =~ ^[[:digit:]]{2}-[[:digit:]]{2}-[[:digit:]]{4}$ ]]; then
        pieces=( $( tr '-' ' ' <<< "$the_date" ) )
        if [[ -n "${pieces[0]}" ]]; then
            the_date="${pieces[2]}-${pieces[0]}-${pieces[1]}"
        else
            the_date="${pieces[3]}-${pieces[1]}-${pieces[2]}"
        fi
    fi
    if [[ ! "$the_date" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
        printf 'Invalid date format [%s]. Use yyyy-MM-dd.\n' "$the_date" >&2
        return 1
    fi
    # Try to make $the_time into HH:mm:ss format and handle any extra precision.
    # Allow for no time input,
    # or formats of HH, HHmm, HH:mm, HHmmss, HHmm:ss, HH:mmss, HH:mm:ss
    # or formats of HHmmss.d+, HHmm:ss.d+, HH:mmss.d+, HH:mm:ss.d+
    s_fractions=
    ms_fractions=
    if [[ -z "$the_time" ]]; then
        the_time='00:00:00'
    elif [[ "$the_time" =~ ^[[:digit:]]{2}(:?[[:digit:]]{2}){0,2}$ ]]; then
        the_time="$( tr -d ':' <<< "$the_time" | sed 's/$/0000/' | head -c 6 | sed -E 's/^(..)(..)(..)$/\1:\2:\3/' )"
    elif [[ "$the_time" =~ ^[[:digit:]]{2}:?[[:digit:]]{2}:?[[:digit:]]{2}\.[[:digit:]]+$ ]]; then
        pieces=( $( tr '.' ' ' <<< "$the_time" ) )
        if [[ -n "${pieces[0]}" ]]; then
            the_time="${pieces[0]}"
            s_fractions="${pieces[1]}"
        else
            the_time="${pieces[1]}"
            s_fractions="${pieces[2]}"
        fi
        the_time="$( tr -d ':' <<< "$the_time" | sed -E 's/^(..)(..)(..)$/\1:\2:\3/' )"
        s_fractions="$( sed -E 's/0+$//' <<< "$s_fractions" )"
        if [[ "${#s_fractions}" -gt '3' ]]; then
            ms_fractions=".$( sed -E 's/^...//' <<< "$s_fractions" )"
        fi
    fi
    if [[ ! "$the_time" =~ ^[[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        printf 'Invalid time format [%s]. Use HH:mm[:ss[.ddd]].\n' "$the_time" >&2
        return 1
    fi
    # Make sure the milliseconds have exactly three decials by padding the right with zeros if needed.
    ms="$( head -c 3 <<< "${s_fractions}000" )"
    # Try to make $the_time_zone into (+|-)HHmm format.
    # Allow for no time zone, (+|-)HH, (+|-)HHmm (+|-)HH:mm
    if [[ -z "$the_time_zone" ]]; then
        the_time_zone="$( date '+%z' )"
    elif [[ "$the_time_zone" =~ ^[+-][[:digit:]]{2}$ ]]; then
        the_time_zone="${the_time_zone}00"
    elif [[ "$the_time_zone" =~ ^[+-][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        the_time_zone="$( tr -d ':' <<< "$the_time_zone" )"
    fi
    if [[ ! "$the_time_zone" =~ ^[+-][[:digit:]]{4}$ ]]; then
        printf 'Invalid timezone format [%s]. Use (+|-)HHmm.\n' "$the_time_zone" >&2
        return 1
    fi
    # Get the epoch as seconds
    epoch_s="$( date -j -f '%F %T %z' "$the_date $the_time $the_time_zone" '+%s' )" || return $?
    # Append the milliseconds and remove any leading zeros.
    epoch_ms="$( sed -E 's/^0+//;' <<< "${epoch_s}${ms}" )"
    # But make sure there's still at least one digit.
    if [[ -z "$epoch_ms" ]]; then
        epoch_ms="0"
    fi
    printf '%s%s\n' "$epoch_ms" "$ms_fractions"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    to_epoch "$@"
    exit $?
fi
unset sourced

return 0
