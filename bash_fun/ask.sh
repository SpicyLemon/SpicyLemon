#!/bin/bash
# This file contains functions for asking questions of flying-ferret.
# File contents:
#   ask  ---------------------> Asks flying ferret a question.
#                                   If Perl is available on your system, and the flyingferret module is in @INC, that will be used.
#                                   Otherwise, an attempt to call the web api will be made.
#   ask_flying_ferret_perl  --> Uses your local perl library to ask flying ferret a question.
#   ask_flying_ferret_api  ---> Uses the web api to ask flying ferret a question.
#
# Usage:
#   To use these functions, put this file somewhere handy and in your .bash_profile, or .zshrc (or whatever) add a command to source it.
#   For example:    source "$HOME/ask.sh"
#   Then, any time you start up a terminal, you will have access to the ask function.
#   Example usage:
#       $ ask pizza or tacos?
#       $ ask roll 5d6
#       $ ask 'should I go to bed?'
#
# Usage Notes:
#   1. Some shells see the '?' and try to do stuff with it, so you might need to put the whole question in quotes.
#   2. In order for the perl module to be used, the flyingferret perl module must be in your standard perl path.
#   3. In order for the web api call to work, you might have to install the jq program. It's a json querying command-line utility.
#   4. The flyingferret perl module will be used if possible. Otherewise the api call will be attempted.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: ask should I stay or should I go?
ask () {
    if [[ "$#" -eq '0' || "$#" -eq '1' && (( "$1" == '-h' || "$1" == '--help' )) ]]; then
        printf 'Usage: ask <query>\n'
        return 1
    fi
    if [[ "$#" -eq '1' && (( "$1" == '-o' || "$1" == '--open' )) ]]; then
        open 'https://flying-ferret.com'
        return 0
    fi
    if command -v 'perl' > /dev/null 2>&1 && perl -e 'use flyingferret;' > /dev/null 2>&1; then
        ask_flying_ferret_perl "$@"
    else
        ask_flying_ferret_api "$@"
    fi
    return $?
}

ask_flying_ferret_perl () {
    local query
    query="$( sed -E -e 's/^[[:space:]]+//; s/[[:space:]]+$//;' -e "s/'/\\\'/g" <<< "$*" )"
    if [[ -z "$query" ]]; then
        printf 'Usage: ask <query>\n'
        return 1
    fi
    perl -Mflyingferret -e "print join(\"\\n\", @{flyingferret::transform('$query')}) . \"\\n\";"
    return $?
}

ask_flying_ferret_api () {
    local query api_url flying_ferret_says results
    query="$( sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//' <<< "$*" )"
    if [[ -z "$query" ]]; then
        printf 'Usage: ask <query>\n'
        return 1
    fi
    api_url='https://www.flying-ferret.com/cgi-bin/api/v1/transform.cgi'
    flying_ferret_says="$( curl -s --data-urlencode "q=$query" "$api_url" 2>&1 )"
    if [[ -z "$( jq ' . ' 2> /dev/null <<< "$flying_ferret_says" )" ]]; then
        printf 'Flying Ferret is confused. See %s for more info.\n' "$api_url?help="
        return 10
    fi
    results="$( jq -r ' .results | .[] ' <<< "$flying_ferret_says" )"
    if [[ -z "$results" ]]; then
        printf 'Flying Ferret returned without any results.\n'
        return 5
    fi
    printf '%s\n' "$results"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    ask "$@"
    exit $?
fi
unset sourced

return 0
