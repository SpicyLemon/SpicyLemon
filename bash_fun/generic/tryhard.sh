#!/bin/bash
# This file contains the tryhard function that runs a command until it exists 0, then it beeps twice
# This file can be sourced to add the tryhard function to your environment.
# This file can also be executed to run the tryhard function without adding it to your environment.
#
# File contents:
#   tryhard   --> runs the provided command until it exits 0, then it beeps twice
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

tryhard  () {
    local c i
    c='92' # light-green
    i=1
    while printf '\033[%sm%s |%2d\033[0m:\033[1m' "$c" "$( date '+%H:%M:%S' )" "$i" && printf ' %q' "$@" && printf '\033[0m\n' && ! "$@"; do
        i=$(( i + 1 ))
        case "$i" in
            2) c='96';; # light-cyan
            4) c='93';; # light-yellow
            7) c='95';; # light-magenta
            10) c='91';; # light-red
            15) c='41;1';; #red background + bold
        esac
        sleep .5
    done
    printf '\a'
    sleep .3
    printf '\a'
}

if [[ "$sourced" != 'YES' ]]; then
    tryhard  "$@"
    exit $?
fi
unset sourced

return 0
