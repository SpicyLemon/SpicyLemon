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

Usage: curl_link_header [--max-calls NUM] [--delimiter <text>] [--rel <text>] <curl options>
    The --max-calls option will limit the number of curl calls made.
        The minimum NUM is 1. The maximum NUM is 65535.
        A NUM of 1 would be the same as just using the curl commmand with the provided options.
        If not provided, the default is 100.
    The --delimiter option defines some text that will be output before every request except the first.
    The --rel option defines which rel link to follow.
        If not provided, the default "next" is used.

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
    local curl_args curl_arg max_calls delimiter clean_up_header_file next_link exit_code
    local initial_url url_count header_file output_file output_file_count create_dirs dir_to_create verbose
    curl_args=()
    url_count=0
    output_file_count=0
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
        -h|--help)
            __curl_link_header_usage
            return 0
            ;;
        --max-links)
            max_calls="$2"
            shift
            ;;
        --max-links=*)
            max_calls="$( printf %s "$1" | sed 's/^--max-links=//;' )"
            ;;
        --delimiter)
            delimiter="$2"
            shift
            ;;
        --delimiter=*)
            delimiter="$( printf %s "$1" | sed 's/^--delimiter=//;' )"
            ;;
        -D|--dump-header)
            header_file="$2"
            curl_args+=( "$1" "$2" )
            shift
            ;;
        -D=*|--dump-header=*)
            header_file="$( printf %s "$1" | sed -E 's/^(-D|--dump-header)=//;' )"
            curl_args+=( "$1" )
            ;;
        -o|--output)
            output_file="$2"
            output_file_count=$(( output_file_count + 1 ))
            ;;
        -o=*|--output=*)
            output_file="$( printf %s "$1" | sed -E 's/^(-o|--output)=//;' )"
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
            echo -e "The $1 option is not supported by the curl_link_header function." >&2
            return 1
            ;;
        *)
            if [[ -n "$( printf %s "$1" | grep -iE '^https?://' )" ]]; then
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
        echo -e "No initial url was found in the provided arguments." >&2
        return 1
    elif [[ -n "$( printf %s "$initial_url" | grep -E '[^\\][]{}[]' )" ]]; then
        echo -e "Variable replacement inside urls is not supported by the curl_link_header function." >&2
        return 1
    elif [[ "$url_count" -gt '1' ]]; then
        echo -e "This curl_link_header function only supports a single url." >&2
        return 1
    fi
    # Make sure the --max-links number is valid.
    if [[ -z "$max_calls" ]]; then
        max_calls="100"
    elif [[ "$max_calls" =~ [^[:digit:]] ]]; then
        echo -e "The --max-links must be a positive number, but [$max_calls] was provided." >&2
        return 1
    elif [[ "$max_calls" -lt '1' ]]; then
        echo -e "The --max-links must be at least 1, but [$max_calls] was provided." >&2
        return 1
    elif [[ "$max_calls" -gt '65535' ]]; then
        echo -e "The --max-links must be at most 65535, but [$max_calls] was provided." >&2
        return 1
    fi
    # Make sure the output file is valid.
    if [[ "$output_file_count" -ge '2' ]]; then
        echo -e "This curl_link_header function only supports a single -o or --output option." >&2
        return 1
    elif [[ "$output_file" =~ \# ]]; then
        echo -e "This curl_link_header function does not support # in filenames: [$output_file]." >&2
        return 1
    elif [[ "$output_file" == '-' ]]; then
        output_file=
        output_file_count=0
    fi
    # Remove the output file if it already exists.
    if [[ -n "$output_file" && -f "$output_file" ]]; then
        [[ -n "$verbose" ]] && echo -e "Removing existing output file: [$output_file]." >&2
        rm "$output_file" || return $?
    fi
    # Create the directories if needed
    if [[ -n "$output_file" && -n "$create_dirs" ]]; then
        dir_to_create="$( dirname "$output_file" )"
        if [[ ! -d "$dir_to_create" ]]; then
            [[ -n "$verbose" ]] && echo "Creating directories for output file: [$dir_to_create]." >&2
            mkdir -p "$dir_to_create" || return $?
        fi
    fi
    # Make sure we've got a header file to use.
    # Beyond this point, we cannot quit the function without taking care of the header file if it's a temp one.
    if [[ -z "$header_file" ]]; then
        clean_up_header_file='YES'
        header_file="$( mktemp -t curl_link_header )"
        curl_args+=( '--dump-header' "$header_file" )
        [[ -n "$verbose" ]] && echo "Using temporary file for response header: [$header_file]." >&2
    fi

    if [[ -n "$verbose" ]]; then
        echo -e    "         Initial url: [$initial_url]" >&2
        echo -e    "           Max calls: [$max_calls]" >&2
        echo -e    "           Delimiter: [$delimiter]" >&2
        echo -e    "Response header file: [$header_file]$( [[ -n "$clean_up_header_file" ]] && echo -e -n ' (temporary)' )" >&2
        echo -e    "           Output to: [$( [[ -n "$output_file" ]] && printf %s "$output_file" || echo -e -n '<stdout>' )]" >&2
        echo -e -n " Parameters for curl:" >&2
        for curl_arg in "${curl_args[@]}"; do
            echo -e -n " [$curl_arg]" >&2
        done
        echo -e    '' >&2
    fi

    # Starting up a sub-shell in order to make a function private that also has access to the variable so far.
    (
        # Create some functions in a subshell to prevent them from being used outside the curl_link_header function.
        # It also keeps the outside environment a bit cleaner.
        # Additionally, since they're inside the curl_link_header function, they have access to all the variables from there.

        # Figure out what the next link is.
        # Usage: __curl_link_header_get_next_link
        __curl_link_header_get_next_link () {
            local full_link_header link_header_values rel_next_link_value
            next_link=
            if [[ ! -f "$header_file" ]]; then
                echo -e "Header file not found: [$header_file]." >&2
                return 1
            elif [[ ! -r "$header_file" ]]; then
                echo -e "Cannot read from header file: [$header_file]." >&2
                return 2
            fi

            full_link_header="$( grep -i '^link:' "$header_file" )"
            if [[ -z "$full_link_header" ]]; then
                echo -e "No link header found in response header." >&2
                return 10
            fi

            link_header_values="$( echo -e "$full_link_header" | sed -E 's/^[Ll][Ii][Nn][Kk]:[[:space:]]*//; s/(<[^>]*>[^,]*)(,|$)[[:space:]]*/\1\'$'\n/g;' )"
            [[ -n "$verbose" ]] && echo -e "Link-value entries in header:\n$link_header_values" >&2

            rel_next_link_value="$( echo -e "$link_header_values" | grep -E ';[[:space:]]*rel="next"[[:space:]]*(;|$)' )"
            if [[ -z "$rel_next_link_value" ]]; then
                [[ -n "$verbose" ]] && echo -e "No link-value with the [rel=\"next\"] link-param found in the link header." >&2
                # No normal output here because this is an expected thing on the last result.
                return 0
            fi
            [[ -n "$verbose" ]] && echo -e "Link-value found for [rel=\"next\"]:\n$rel_next_link_value" >&2

            next_link="$( echo -e "$rel_next_link_value" | sed -E 's/^.*<([^>]*)>.*$/\1/;' )"
            if [[ -z "$next_link" ]]; then
                echo -e "The URI-Reference in the link-value with the [rel=\"next\"] link-param, is empty." >&2
                return 12
            fi
            [[ -n "$verbose" ]] && echo -e "Next link: $next_link" >&2
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
                [[ -n "$verbose" ]] && echo -e "Executing command: ${curl_cmd[@]} >> $output_file" >&2
                "${curl_cmd[@]}" >> "$output_file"
                exit_code=$?
            else
                [[ -n "$verbose" ]] && echo -e "Executing command: ${curl_cmd[@]}" >&2
                "${curl_cmd[@]}"
                exit_code=$?
            fi

            # Get the next link unless there was an issue.
            if [[ "$exit_code" -eq '0' ]]; then
                __curl_link_header_get_next_link
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

        [[ -n "$verbose" ]] && echo "Made [$calls_made] curl calls." >&2
        [[ -n "$next_link" ]] && echo "The final call still had a next link: [$next_link]." >&2

        # Unfortunately, the subshell makes all variables set inside it go back to what they used to be when it ends.
        # Fortunatly though, all I want out of here is the exit code! Easy peasy!
        exit "$exit_code"
    )
    exit_code="$?"

    # Clean up
    if [[ -n "$clean_up_header_file" ]]; then
        [[ -n "$verbose" ]] && echo "Deleting temporary response header file: [$header_file]." >&2
        rm "$header_file"
    fi

    [[ -n "$verbose" ]] && echo -e "Returning with exit code [$exit_code]." >&2

    return "$exit_code"
}

# If this script was not sourced make it do things now.
if [[ "$sourced" != 'YES' ]]; then
    curl_link_header "$@"
    exit $?
fi
unset sourced

return 0
