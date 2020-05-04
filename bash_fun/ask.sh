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
#   2. In order for the web api call to work, you might have to install the jq program. It's a json querying command-line utility.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: ask should I stay or should I go?
ask () {
    if [[ "$#" -eq '1' && (( "$1" == '-o' || "$1" == '--open' )) ]]; then
        open "https://flying-ferret.com"
        return 0
    fi
    local can_perl cant_perl_ff
    can_perl="$( type perl | grep -v 'not found' )"
    if [[ -n "$can_perl" ]]; then
        cant_perl_ff="$( perl -e "use flyingferret;" 2>&1 )"
    fi
    if [[ -n "$can_perl" && -z "$cant_perl_ff" ]]; then
        ask_flying_ferret_perl "$@"
    else
        ask_flying_ferret_api "$@"
    fi
}

ask_flying_ferret_perl () {
    local query
    query="$( echo -E "$*" | sed -e 's/^[[:space:]]+//; s/[[:space:]]+$//;' -e "s/'/\\\'/g" )"
    perl -Mflyingferret -e "print join(\"\\n\", @{flyingferret::transform('$query')}) . \"\\n\";"
}

ask_flying_ferret_api () {
    local query api_url flying_ferret_says results
    query="$( echo -E "$*" | sed -e 's/^[[:space:]]+//; s/[[:space:]]+$//' )"
    if [[ -z "$query" ]]; then
        echo "Usage: ask <query>"
        return 1
    fi
    api_url='https://www.flying-ferret.com/cgi-bin/api/v1/transform.cgi'
    flying_ferret_says="$( curl -s --data-urlencode "q=$query" "$api_url" 2>&1 )"
    if [[ -z "$( echo -E "$flying_ferret_says" | jq ' . ' 2> /dev/null )" ]]; then
        echo -E "Flying ferret is confused. See $api_url?help= for more info."
        return 10
    fi
    results="$( echo -E "$flying_ferret_says" | jq -r ' .results | .[] ' )"
    if [[ -z "$results" ]]; then
        echo -E 'Flying ferret returned without any results.'
        return 5
    fi
    echo -E "$results"
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    if [[ "$#" -gt '0' ]]; then
        ask "$@"
    else
        echo -E "Usage: ./$( basename "$0" ) <flying ferret query>"
    fi
fi
unset sourced
