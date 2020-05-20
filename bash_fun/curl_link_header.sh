#!/bin/bash
# This file contains a function that uses curl to follow the rel="next" url in a link header to get multiple pages of results.
# It can be sourced to add the curl_link_header function to your environment.
# It can also be executed for the same functionality.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Information on using this function
__curl_link_header_usage () {
    cat << EOF
curl_link_header - Uses the link header in a result to get multiple pages of results.

This function uses curl to do all the work.

Usage: curl_link_header [--max-calls NUM] [--delimiter <text>] [--rel <text>] [-i|--interactive] <curl options>
    The --max-calls option will limit the number of curl calls made.
        The minimum NUM is 1. The maximum NUM is 65535.
        A NUM of 1 would be the same as just using the curl commmand with the provided options.
        If not provided, the default is 100.
    The --delimiter option defines some text that will be output before every request except the first.
    The --rel option defines the rel of the link-value entry to follow.
        If not provided, the default "next" is used.
    The -i or --interactive option allows you to select the desired next link to get after each result.
        If this is provided, any --rel option will be ignored.

See curl --help for information on curl options.

Limitations on curl options:
    Only one -o or --output option can be provided.
    The -O, --remote-name, -J, and --remote-header-name options cannot be used.
    Exactly one url must be supplied, and it must start with either http:// or https://.
    The provided url cannot include "part sets" or sequences, e.g. {one,two,three} or [1-10].
    Similarly, if an output file is provided, it cannot have a # in it (used for variable replacement with "part sets" and sequences).

EOF
}

# The main wrapper command that adds the extra stuff.
curl_link_header () {
    local curl_args max_calls delimiter rel_value egrep_escaped_rel_value interactive clean_up_header_file next_link exit_code
    local initial_url url_count header_file output_file output_file_count create_dirs dir_to_create verbose
    rel_value='next'
    curl_args=()
    url_count=0
    output_file_count=0
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
        -h|--help)
            __curl_link_header_usage
            return 0
            ;;
        --max-links|--max-calls)
            max_calls="$2"
            shift
            ;;
        --max-links=*|--max-calls=*)
            max_calls="$( sed -E 's/^--max-(links|calls)=//;' <<< "$1" )"
            ;;
        --delimiter)
            delimiter="$2"
            shift
            ;;
        --delimiter=*)
            delimiter="$( sed 's/^--delimiter=//;' <<< "$1" )"
            ;;
        --rel)
            rel_value="$2"
            shift
            ;;
        --rel=*)
            rel_value="$( sed 's/^--rel=//;' <<< "$1" )"
            ;;
        -i|--interactive)
            if ! command -v 'fzf' > /dev/null 2>&1; then
                printf 'The fzf program is required for interactive mode. See https://github.com/junegunn/fzf\n' >&2
                return 1
            fi
            interactive="$1"
            ;;
        -D|--dump-header)
            header_file="$2"
            curl_args+=( "$1" "$2" )
            shift
            ;;
        -D=*|--dump-header=*)
            header_file="$( sed -E 's/^(-D|--dump-header)=//;' <<< "$1" )"
            curl_args+=( "$1" )
            ;;
        -o|--output)
            output_file="$2"
            output_file_count=$(( output_file_count + 1 ))
            ;;
        -o=*|--output=*)
            output_file="$( sed -E 's/^(-o|--output)=//;' <<< "$1" )"
            output_file_count=$(( output_file_count + 1 ))
            ;;
        --create-dirs)
            create_dirs='YES'
            ;;
        -v|--verbose)
            verbose='YES'
            curl_args+=( "$1" )
            ;;
        -O|--remote-name|-J|--remote-header-name)
            printf 'The [%s] option is not supported by the curl_link_header function.\n' "$1" >&2
            return 1
            ;;
        *)
            if [[ -n "$( grep -iE '^https?://' <<< "$1" )" ]]; then
                initial_url="$1"
                url_count=$(( url_count + 1 ))
            else
                curl_args+=( "$1" )
            fi
            ;;
        esac
        shift
    done
    # Make sure we've got an initial url, and that we were given only one.
    if [[ -z "$initial_url" ]]; then
        printf 'No initial url was found in the provided arguments.\n' >&2
        return 1
    elif [[ -n "$( grep -E '[^\\][]{}[]' <<< "$initial_url" )" ]]; then
        printf 'Variable replacement inside urls is not supported by the curl_link_header function.\n' >&2
        return 1
    elif [[ "$url_count" -gt '1' ]]; then
        printf 'This curl_link_header function only supports a single url.\n' >&2
        return 1
    fi
    # Make sure the --max-links number is valid.
    if [[ -z "$max_calls" ]]; then
        max_calls="100"
    elif [[ "$max_calls" =~ [^[:digit:]] ]]; then
        printf 'The --max-links must be a positive number, but [%s] was provided.\n' "$max_calls" >&2
        return 1
    elif [[ "$max_calls" -lt '1' ]]; then
        printf 'The --max-links must be at least 1, but [%s] was provided.\n' "$max_calls" >&2
        return 1
    elif [[ "$max_calls" -gt '65535' ]]; then
        printf 'The --max-links must be at most 65535, but [%s] was provided.\n' "$max_calls" >&2
        return 1
    fi
    # Make sure we have a --rel value.
    if [[ -z "$rel_value" ]]; then
        printf 'No rel value defined.\n' >&2
        return 1
    fi
    egrep_escaped_rel_value="$( sed 's/[][\.|$(){}?+*^]/\\&/g' <<< "$rel_value" )"
    # Make sure the output file is valid.
    if [[ "$output_file_count" -ge '2' ]]; then
        printf 'This curl_link_header function only supports a single -o or --output option.\n' >&2
        return 1
    elif [[ "$output_file" =~ \# ]]; then
        printf 'This curl_link_header function does not support # in filenames: [%s].\n' "$output_file" >&2
        return 1
    elif [[ "$output_file" == '-' ]]; then
        output_file=
        output_file_count=0
    fi
    # Remove the output file if it already exists.
    if [[ -n "$output_file" && -f "$output_file" ]]; then
        [[ -n "$verbose" ]] && printf 'Removing existing output file: [%s].\n' "$output_file" >&2
        rm "$output_file" || return $?
    fi
    # Create the directories if needed
    if [[ -n "$output_file" && -n "$create_dirs" ]]; then
        dir_to_create="$( dirname "$output_file" )"
        if [[ ! -d "$dir_to_create" ]]; then
            [[ -n "$verbose" ]] && printf 'Creating directories for output file: [%s].\n' "$dir_to_create" >&2
            mkdir -p "$dir_to_create" || return $?
        fi
    fi
    # Make sure we've got a header file to use.
    # Beyond this point, we cannot quit the function without taking care of the header file if it's a temp one.
    if [[ -z "$header_file" ]]; then
        clean_up_header_file='YES'
        header_file="$( mktemp -t curl_link_header )"
        curl_args+=( '--dump-header' "$header_file" )
        [[ -n "$verbose" ]] && printf 'Using temporary file for response header: [%s].\n' "$header_file" >&2
    fi

    if [[ -n "$verbose" ]]; then
        {
            printf '         Initial url: [%s]\n' "$initial_url"
            printf '                 rel: [%s]%s\n' "$rel_value" \
                        "$( [[ "$rel_value" != "$egrep_escaped_rel_value" ]] && printf ' escaped: [%s]' "$egrep_escaped_rel_value" )"
            printf '         Interactive: [%s]\n' "$( [[ -n "$interactive" ]] && printf 'YES' || printf 'no' )"
            printf '           Max calls: [%s]\n' "$max_calls"
            printf '           Delimiter: [%s]\n' "$delimiter"
            printf 'Response header file: [%s]%s\n' "$header_file" "$( [[ -n "$clean_up_header_file" ]] && printf ' (temporary)' )"
            printf '           Output to: %s\n' "$( [[ -n "$output_file" ]] && printf '[%s]' "$output_file" || printf '<stdout>' )"
            printf ' Parameters for curl:'; printf ' [%s]' "${curl_args[@]}"; printf '\n'
        } >&2
    fi

    # Starting up a sub-shell in order to make a function private that also has access to the variable so far.
    (
        # Create some functions in a subshell to prevent them from being used outside the curl_link_header function.
        # It also keeps the outside environment a bit cleaner.
        # Additionally, since they're inside the curl_link_header function, they have access to all the variables from there.

        # Figure out what the next link is.
        # Usage: __curl_link_header_set_next_link
        __curl_link_header_set_next_link () {
            local full_link_header link_value_entries rel_next_link_value
            if [[ ! -f "$header_file" ]]; then
                printf 'Header file not found: [%s].\n' "$header_file" >&2
                return 1
            elif [[ ! -r "$header_file" ]]; then
                printf 'Cannot read from header file: [%s].\n' "$header_file" >&2
                return 2
            fi

            full_link_header="$( grep -i -E '^links?:' "$header_file" )"
            if [[ -z "$full_link_header" ]]; then
                if [[ -n "$verbose" ]]; then
                    {
                        printf 'No Link header found in response header'
                        if [[ -n "$clean_up_header_file" ]]; then
                            printf '\n'
                            cat "$header_file" | sed 's/^/  /;'
                        else
                            printf ' [%s].\n' "$header_file"
                        fi
                    } >&2
                fi
                # Unfortunately, there's no real way to tell from here if the lack of a link header
                # is a bad thing (e.g. you're not using the url you think you are),
                # or a good thing (you've gotten all of the results).
                # So at least, for now, just return link it's all good.
                return 0
            fi

            link_value_entries="$( sed -E 's/^[Ll][Ii][Nn][Kk][sS]?:[[:space:]]*//; s/(<[^>]*>[^,]*)(,|$)[[:space:]]*/\1\'$'\n/g;' <<< "$full_link_header" )"
            [[ -n "$verbose" ]] && printf 'Link-value entries in header:\n%s\n' "$link_value_entries" >&2

            if [[ -n "$interactive" ]]; then
                if [[ "$calls_made" -ge "$max_calls" ]]; then
                    [[ -n "$verbose" ]] && printf 'Max number of calls reached.\n' >&2
                    return 0
                fi
                rel_next_link_value="$( fzf --tac +m --cycle --header='Select the desired link-value.'<<< "$link_value_entries" )"
                if [[ -z "$rel_next_link_value" ]]; then
                    [[ -n "$verbose" ]] && printf 'No link-value selected.\n' >&2
                    # Nothing selected. Stop going.
                    return 0
                fi
                [[ -n "$verbose" ]] && printf 'Link-value selected:\n%s\n' "$rel_next_link_value" >&2
            else
                rel_next_link_value="$( grep -E ';[[:space:]]*rel="'"$egrep_escaped_rel_value"'"[[:space:]]*(;|$)' <<< "$link_value_entries" )"
                if [[ -z "$rel_next_link_value" ]]; then
                    [[ -n "$verbose" ]] && printf 'No link-value with the [rel="%s"] link-param found in the link header.\n' "$rel_value" >&2
                    # No normal output here because this is an expected thing on the last result.
                    return 0
                fi
                [[ -n "$verbose" ]] && printf 'Link-value found for [rel="%s"]:\n%s\n' "$rel_value" "$rel_next_link_value" >&2
            fi

            next_link="$( sed -E 's/^.*<([^>]*)>.*$/\1/;' <<< "$rel_next_link_value" )"
            if [[ -z "$next_link" ]]; then
                if [[ -n "$interactive" ]]; then
                    printf 'The URI-Reference in the selected link-value, is empty.\n' >&2
                else
                    printf 'The URI-Reference in the link-value with the [rel="%s"] link-param, is empty.\n' "$rel_value" >&2
                fi
                return 12
            fi
            [[ -n "$verbose" ]] && printf 'Next link: [%s]\n' "$next_link" >&2
            return 0
        }

        # This function is for actually making the curl call, and getting the next link.
        # Usage: __curl_link_header_do_curl "$url"
        __curl_link_header_do_curl () {
            local url curl_cmd
            url="$1"
            curl_cmd=( curl "${curl_args[@]}" "$url" )
            calls_made=$(( calls_made + 1 ))
            [[ -n "$verbose" ]] && printf 'Call # [% 5d]\n' "$calls_made" >&2
            # Make the curl call and set the exit code.
            if [[ -n "$output_file" ]]; then
                [[ -n "$verbose" ]] && printf 'Executing command: %s >> %s\n' "${curl_cmd[*]}" "$output_file" >&2
                "${curl_cmd[@]}" >> "$output_file"
                exit_code=$?
            else
                [[ -n "$verbose" ]] && printf 'Executing command: %s\n' "${curl_cmd[*]}" >&2
                "${curl_cmd[@]}"
                exit_code=$?
            fi

            # Get the next link unless there was an issue.
            next_link=
            if [[ "$exit_code" -eq '0' ]]; then
                __curl_link_header_set_next_link
            fi
        }

        # Let's get it started (in here)!
        calls_made=0
        __curl_link_header_do_curl "$initial_url"

        # Keep going until the well runs dry!
        while [[ "$exit_code" -eq '0' && -n "$next_link" && "$calls_made" -lt "$max_calls" ]]; do
            if [[ -n "$delimiter" ]]; then
                if [[ -n "$output_file" ]]; then
                    printf '%s' "$delimiter" >> "$output_file"
                else
                    printf '%s' "$delimiter"
                fi
            fi
            __curl_link_header_do_curl "$next_link"
        done

        [[ -n "$verbose" ]] && printf 'Made [%s] curl calls.\n' "$calls_made" >&2
        [[ -n "$next_link" ]] && printf 'The final result still had a [rel="%s"] link: [%s].\n' "$rel_value" "$next_link" >&2

        # Unfortunately, the subshell makes all variables set inside it go back to what they used to be when it ends.
        # Fortunatly though, all I want out of here is the exit code! Easy peasy!
        exit "$exit_code"
    )
    exit_code="$?"

    # Clean up
    if [[ -n "$clean_up_header_file" ]]; then
        [[ -n "$verbose" ]] && printf 'Deleting temporary response header file: [%s].\n' "$header_file" >&2
        rm "$header_file"
    fi

    [[ -n "$verbose" ]] && printf 'Returning with exit code [%s].\n' "$exit_code" >&2

    return "$exit_code"
}

# If this script was not sourced make it do things now.
if [[ "$sourced" != 'YES' ]]; then
    curl_link_header "$@"
    exit $?
fi
unset sourced

return 0
