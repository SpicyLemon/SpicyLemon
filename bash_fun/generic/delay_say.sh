#!/bin/bash
# This file contains the delay_say function that waits some minutes before saying something.
# This file can be sourced to add the delay_say function to your environment.
# This file can also be executed to run the delay_say function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: delay_say <minutes> <message>
delay_say () {
    local has_help arg
    for arg in "$@"; do
        if [[ "$arg" == '--help' || "$arg" == '-h' ]]; then
            has_help='YES'
            break
        fi
    done
    if [[ -n "$has_help" || "$#" -lt '2' ]]; then
        printf 'Usage: delay_say <minutes> <message>\n'
        return 0
    fi
    local minutes_in hours minutes seconds s_div message to_sleep
    minutes_in="$1"
    shift
    message="$*"

    if [[ "$minutes_in" =~ [hms] ]]; then
        [[ -n "$DEBUG" ]] && printf 'hms: "%s"\n' "$minutes_in"
        hours="$( sed -E 's/^.*([[:digit:]]+)h.*$/\1/' <<< "$minutes_in" )"
        minutes="$( sed -E 's/^.*([[:digit:]]+)m.*$/\1/' <<< "$minutes_in" )"
        seconds="$( sed -E 's/^.*([[:digit:]]+)s.*$/\1/' <<< "$minutes_in" )"
        if [[ -z "$hours" || "$hours" == "$minutes_in" ]]; then
            hours='0'
        fi
        if [[ -z "$minutes" || "$minutes" == "$minutes_in" ]]; then
            minutes='0'
        fi
        if [[ -z "$seconds" || "$seconds" == "$minutes_in" ]]; then
            seconds='0'
        fi
        [[ -n "$DEBUG" ]] && printf 'hms: [%s:%s:%s]\n' "$hours" "$minutes" "$seconds"
    elif [[ "$minutes_in" =~ \. ]]; then
        [[ -n "$DEBUG" ]] && printf 'dec: "%s"\n' "$minutes_in"
        hours='0'
        minutes="$( sed -E 's/^([[:digit:]]*)\..*$/\1/' <<< "$minutes_in" )"
        seconds="$( sed -E 's/^.*\.([[:digit:]]*)$/\1/' <<< "$minutes_in" )"
        if [[ -z "$minutes" ]]; then
            minutes='0'
        fi
        if [[ -n "$seconds" ]]; then
            # E.g. .5 => 5 * 60 / 10 = 30 seconds, .75 => 75 * 60 / 100 = 45 seconds.
            s_div="1$( sed -E 's/./0/g' <<< "$seconds" )"
            seconds="$(( seconds * 60 / s_div ))"
        else
            seconds='0'
        fi
    elif [[ -n "$minutes_in" ]]; then
        [[ -n "$DEBUG" ]] && printf 'raw: "%s"\n' "$minutes_in"
        hours='0'
        minutes="$minutes_in"
        seconds='0'
    else
        [[ -n "$DEBUG" ]] && printf 'default: "%s"\n' "$minutes_in"
        hours='0'
        minutes='1'
        seconds='0'
    fi
    to_sleep="$(( seconds + 60 * minutes + 3600 * hours ))"
    printf 'After %dh %dm %ds = %d seconds, saying "%s".\n' "$hours" "$minutes" "$seconds" "$to_sleep" "$message"
    sleep "$to_sleep" && say "$message" &
}

if [[ "$sourced" != 'YES' ]]; then
    delay_say "$@"
    exit $?
fi
unset sourced

return 0
