#!/bin/bash
# This file contains the get_events function consicely displays tx events.
# This file can be sourced to add the get_events function to your environment.
# This file can also be executed to run the get_events function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

get_events () {
    local ec="$?"
    local filename decode path long opts
    opts=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h|help)
                cat << EOF
get_events - Concisely display tx events.

Usage: get_events <tx json file> [--path|-p <path>] [--decode|-d] [--long|-l]
   or: <stuff> | get_events [--path|-p <path>] [--decode|-d] [--long|-l]

The --path <path> option allows you to define the json path to the list of events.
    The default <path> is '.events'.
The --decode flag will cause the attribute keys and values to be base64 decoded.
The --long flag causes the full json path to each attribute to be displayed instead of a shorter form.
    Standard output format: [<event index>]<event type>[<attribute index>]: <key> = <value>
    Long output format:     <path to events>[<event index>].attributes[<attribute index>] (<event type>): <key> = <value>

EOF

                return 0
                ;;
            --path|-p)
                if [[ -z "$2" ]]; then
                    printf 'No argument provided after %s\n' "$1"
                    return 1
                fi
                path="$2"
                opts+=( "$1" "$2" )
                shift
                ;;
            --decode|-d)
                decode="$1"
                opts+=( "$1" )
                ;;
            --long|-l)
                long="$1"
                opts+=( "$1" )
                ;;
            *)
                if [[ -n "$filename" ]]; then
                    printf 'Unknown argument: %q\n' "$1"
                    return 1
                fi
                filename="$1"
                ;;
        esac
        shift
    done

    if [[ -n "$filename" ]]; then
        cat "$filename" | get_events "${opts[@]}"
        return $?
    fi
    if [[ "$ec" -ne '0' ]]; then
        return "$ec"
    fi
    if [[ -z "$path" ]]; then
        path='.events'
    fi
    jq -r "$path"' | to_entries | .[] | (.key|tostring) + " " + .value.type + " " + (.value.attributes|to_entries|.[]| (.key|tostring) + " " + .value.key + " " + .value.value)' \
        | while read ei et ai key value; do
            if [[ -n "$decode" ]]; then
                key="$( base64 -d <<< "$key" )"
                value="$( base64 -d <<< "$value" )"
            fi
            if [[ -z "$key" ]]; then
                key='<no key>'
            fi
            if [[ -z "$value" ]]; then
                value='<no value>'
            fi
            if [[ -n "$long" ]]; then
                printf '%s[%d].attributes[%d] (%s): %s = %s\n' "$path" "$ei" "$ai" "$et" "$key" "$value"
            else
                printf '[%d]%s[%d]: %s = %s\n' "$ei" "$et" "$ai" "$key" "$value"
            fi
        done
}

if [[ "$sourced" != 'YES' ]]; then
    get_events "$@"
    exit $?
fi
unset sourced

return 0
