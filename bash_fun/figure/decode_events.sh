#!/bin/bash
# This file contains the decode_events function that base64 decodes the events in an JSON tx response.
# This file can be sourced to add the decode_events function to your environment.
# This file can also be executed to run the decode_events function without adding it to your environment.
#
# This decode_events function has been deprecated.
# It's replaced by: get_events --long --decode
# Although, you probably don't need the --decode flag now since the SDK fixed it so that the keys and values are no longer base64 encoded.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: decode_events <tx json file>
#    or: <stuff> | decode_events
decode_events () {
    local ec="$?"
    if [[ -n "$1" ]]; then
        if [[ "$1" == 'help' || "$1" == '--help' || "$1" == '-h' ]]; then
            printf 'Usage: decode_events <tx json file>   or   <stuff> | decode_events\n'
            return 0
        fi
        cat "$1" | decode_events
        return $?
    fi
    if [[ "$ec" -ne '0' ]]; then
        return "$ec"
    fi
    jq -r '.events|to_entries|.[]| (.key|tostring) + " " + .value.type + " " + (.value.attributes|to_entries|.[]| (.key|tostring) + " " + .value.key + " " + .value.value)' \
        | while read i1 i1t i2 key value; do
            printf 'events[%d].attributes[%d] (%s): "%s" = "%s"\n' "$i1" "$i2" "$i1t" "$( base64 -d <<< "$key" )" "$( base64 -d <<< "$value" )";
        done
}

if [[ "$sourced" != 'YES' ]]; then
    decode_events "$@"
    exit $?
fi
unset sourced

return 0
