#!/bin/bash
# This file contains the bashcache function that can be used to facilitate caching of command-line data.
# This file can be sourced to add the bashcache function to your environment.
# This file can also be executed to run the bashcache function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

bashcache () {
    local exit_code_ok exit_code_bad_args exit_code_stale_data exit_code_no_data usage
    exit_code_ok=0
    exit_code_bad_args=1
    exit_code_stale_data=10
    exit_code_no_data=11
    usage="$( cat << EOF
bashcache: A bash command-line caching utility.

Usage: bashcache <command> <cache name> [<options>]

    The first argument is required and is the command you wish to run.
    The second argument is required (except for the list command), and represents the name of the cache.
        Cache names cannot start with a dash.

    Commands:
        write
            Stores the provided data as the given cache name.
            Data can be piped in, provided as a here doc, here string, or on the command line after the -- option.
            Examples:
                bashcache write foo -- This is my data.
                echo "This is my data." | bashcache write foo
                bashcache write foo <<< "This is my data."
                echo "This is my data." > foo-file.txt && bashcache write foo << foo-file.txt
        read
            Gets the data from the desired cache name.
            Example:
                foo="\$( bashcache read foo )"
            The exit codes for the check command are used for this command as well.
        check
            Checks that the desired data exists and is sufficiently fresh.
            Output for this command is done by means of the exit code.
            Example:
                if ! bashcache check foo; then ...
            Exit codes:
                $exit_code_ok - The requested cache data is available and up-to-date.
                $exit_code_bad_args - Invalid arguments provided to the bashcache command.
                $exit_code_stale_data - The requested cache data is available, but stale.
                $exit_code_no_data - The requested cache data is not available.
        list
            Outputs a list of cache names.
            When supplying this command, the cache name should be omitted.
            This command also has the --details option, which will cause it to output size and date information as well.
        delete
            Deletes the desired cache name.

    Options:
        -v --verbose                    Provide extra information while executing the desired command.
        -d --dir <directory>            Dictate the desired directory to use for cached information.
        -a --age --max-age <max age>    Dictate the desired max age value.
                                        The format is the same as used for the -atime option of the find command, without the +.
                                        Examples: '23h' '6d12h30m'

    Defaults:
        The default directory for cached information is /tmp/bashcache.
        To change this, set the BASHCACHE_DIR environment variable to the desired directory.
        The -d or --dir option takes precedence over the BASHCACHE_DIR value (or default if not set)

        The default max age is '24h'.
        To change this, set the BASHCACHE_MAX_AGE environment variable to the desired value.
        The -a or --age or --max-age option takes precedence over the BASHCACHE_MAX_AGE value (or default if not set)
EOF
    )"
    local cache_command cache_name verbose cache_dir max_age details input_from_args cache_file
    if [[ "$#" -eq '0' || "$1" == '-h' || "$1" == '--help' || "$2" == '-h' || "$2" == '--help' ]]; then
        printf '%s\n' "$usage"
        return $exit_code_ok
    fi
    cache_command="$( tr '[:upper:]' '[:lower:]' <<< "$1" )"
    shift
    if [[ "$#" -gt '0' && ( "$cache_command" != 'list' || ! "$1" =~ ^- ) ]]; then
        cache_name="$1"
        shift
    fi
    if [[ "$cache_name" =~ ^- ]]; then
        printf 'Cache name cannot start with a dash: [%s]\n' "$cache_name"
        return $exit_code_bad_args
    fi
    while [[ "$#" -gt '0' && -z "$input_from_args" ]]; do
        case "$1" in
        -h|--help)
            printf '%s\n' "$usage"
            return $exit_code_ok
            ;;
        -v|--verbose)
            verbose='YES'
            ;;
        -d|--dir)
            cache_dir="$2"
            shift
            ;;
        -a|--age|--max-age)
            max_age="$2"
            shift
            ;;
        --details)
            details='YES'
            ;;
        --)
            input_from_args='YES'
            ;;
        *)
            printf 'Unknown option: [%s].\n' "$1" >&2
            return $exit_code_bad_args
            ;;
        esac
        shift
    done

    if [[ -n "$input_from_args" ]]; then
        [[ -n "$verbose" ]] && printf 'Getting the data from the rest of the provided arguments.\n' "$cache_dir" >&2
        input_from_args="$*"
    fi

    if [[ -z "$cache_command" ]]; then
        printf 'No command provided.\n' >&2
        return $exit_code_bad_args
    fi
    if [[ ! "$cache_command" =~ ^(write|read|check|list|delete)$ ]]; then
        printf 'Unknown command: [%s].\n' "$cache_command" >&2
        return $exit_code_bad_args
    fi

    if [[ -n "$cache_name" ]]; then
        [[ -n "$verbose" ]] && printf 'Cache name: [%s].\n' "$cache_name" >&2
    elif [[ "$cache_command" != 'list' ]]; then
        printf 'No cache name provided.\n' >&2
        return $exit_code_bad_args
    fi


    if [[ -n "$cache_dir" ]]; then
        [[ -n "$verbose" ]] && printf 'Directory provided from argument: [%s].\n' "$cache_dir" >&2
    elif [[ -n "$BASHCACHE_DIR" ]]; then
        cache_dir="$BASHCACHE_DIR"
        [[ -n "$verbose" ]] && printf 'Directory provided from BASHCACHE_DIR environment variable: [%s].\n' "$cache_dir" >&2
    else
        cache_dir='/tmp/bashcache'
        [[ -n "$verbose" ]] && printf 'Directory set from default: [%s].\n' "$cache_dir" >&2
    fi
    if [[ "$cache_dir" =~ ^[^/] ]]; then
        printf 'Invalid directory: [%s]. It must start with a slash.\n' "$cache_dir" >&2
        return $exit_code_bad_args
    fi
    if [[ -n "$max_age" ]]; then
        [[ -n "$verbose" ]] && printf 'Max age provided from argument: [%s].\n' "$max_age" >&2
    elif [[ -n "$BASHCACHE_MAX_AGE" ]]; then
        max_age="$BASHCACHE_MAX_AGE"
        [[ -n "$verbose" ]] && printf 'Max age provided from BASHCACHE_MAX_AGE environment variable: [%s].\n' "$max_age" >&2
    else
        max_age='24h'
        [[ -n "$verbose" ]] && printf 'Max age set from default: [%s].\n' "$max_age" >&2
    fi
    if [[ ! "$max_age" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
        printf 'Invalid max age: [%s].\n' "$max_age" >&2
        return $exit_code_bad_args
    fi

    if [[ ! -d "$cache_dir" ]]; then
        [[ -n "$verbose" ]] && printf 'Cache directory does not yet exist. Attempting to create [%s].\n' "$cache_dir" >&2
        if ! mkdir -p "$cache_dir" 1>&2; then
            printf 'Unable to create cache directory: [%s].\n' "$cache_dir" >&2
            return $exit_code_bad_args
        fi
        [[ -n "$verbose" ]] && printf 'Cache directory created successfully.\n' >&2
    else
        [[ -n "$verbose" ]] && printf 'Cache directory exists.\n' >&2
    fi

    cache_file="${cache_dir}/${cache_name}"

    case "$cache_command" in
    write)
        if [[ -n "$input_from_args" ]]; then
            [[ -n "$verbose" ]] && printf 'Writing cache file from data provided as arguments: [%s].\n' "$cache_file" >&2
            printf '%s' "$input_from_args" > "$cache_file"
        else
            [[ -n "$verbose" ]] && printf 'Writing cache file from data provided using stdin: [%s].\n' "$cache_file" >&2
            cat - > "$cache_file"
        fi
        if [[ -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'Done caching [%s] of data in [%s].\n' "$( ls -lh "$cache_file" | awk -F " " '{print $5}' )" "$cache_file" >&2
            return $exit_code_ok
        else
            [[ -n "$verbose" ]] && printf 'Failed to cache data to [%s].\n' "$cache_file" >&2
            return $exit_code_bad_args
        fi
        ;;
    read)
        if [[ -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'Cache exists [%s], outputting data.\n' "$cache_file" >&2
            cat "$cache_file"
            [[ -n "$verbose" ]] && printf 'Done outputting data from [%s].\n' "$cache_file" >&2
            if [[ -n "$( find "$cache_file" -mtime "+$max_age" )" ]]; then
                [[ -n "$verbose" ]] && printf 'Cache is stale [%s] (older than [%s]).\n' "$cache_file" "$max_age" >&2
                return $exit_code_stale_data
            else
                [[ -n "$verbose" ]] && printf 'Cache is sufficiently fresh [%s] (not older than [%s]).\n' "$cache_file" "$max_age" >&2
                return $exit_code_ok
            fi
        else
            [[ -n "$verbose" ]] && printf 'Cache does not exist [%s].\n' "$cache_file" >&2
            return $exit_code_no_data
        fi
        ;;
    check)
        if [[ -f "$cache_file" ]]; then
            if [[ -n "$( find "$cache_file" -mtime "+$max_age" )" ]]; then
                [[ -n "$verbose" ]] && printf 'Cache is stale [%s] (older than [%s]).\n' "$cache_file" "$max_age" >&2
                return $exit_code_stale_data
            else
                [[ -n "$verbose" ]] && printf 'Cache is sufficiently fresh [%s] (not older than [%s]).\n' "$cache_file" "$max_age" >&2
                return $exit_code_ok
            fi
        else
            [[ -n "$verbose" ]] && printf 'Cache does not exist [%s].\n' "$cache_file" >&2
            return $exit_code_no_data
        fi
        ;;
    list)
        if [[ -d "$cache_dir" ]]; then
            if [[ -n "$details" ]]; then
                # list files using -l (long format) -h (human readible sizes) and -T (complete time info) and -t (sort by last modified)
                # Grep for just the file lines. Directories wills start with a 'd' and there's also a "total" line to ignore.
                # Get rid of the first 4 columns: permissions, number of links, owner name, group name.
                #   But maintain left-padding whitespace on the sizes so things still line up nicely.
                ls -lhTt "$cache_dir/" | grep '^-' | sed -E 's/^([^[:space:]]+[[:space:]]+){3}[^[:space:]]+[[:space:]]{2}//'
            else
                ls "$cache_dir/"
            fi
            [[ -n "$verbose" && ! "$( ls "$cache_dir/" )" =~ [^[:space:]] ]] && printf 'Cache directory is empty: [%s].\n' "$cache_dir" >&2
            return $exit_code_ok
        else
            [[ -n "$verbose" ]] && printf 'Cache directory does not exist: [%s].\n' "$cache_dir" >&2
            return $exit_code_no_data
        fi
        ;;
    delete)
        if [[ -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'Deleting cache file: [%s].\n' "$cache_file" >&2
            rm "$cache_file"
            return $exit_code_ok
        else
            [[ -n "$verbose" ]] && printf 'Cache does not exist to delete [%s].\n' "$cache_file" >&2
            return $exit_code_no_data
        fi
        ;;
    esac

    [[ -n "$verbose" ]] &&  printf 'Command [%s] did not specify its own exit code. This is a bug.\n' "$cache_command" >&2
    return $exit_code_bad_args
}

if [[ "$sourced" == 'YES' ]]; then
    if [[ -z "$BASHCACHE_DIR" ]]; then
        BASHCACHE_DIR='/tmp/bashcache'
    fi
    if [[ -z "$BASHCACHE_MAX_AGE" ]]; then
        BASHCACHE_MAX_AGE='24h'
    fi
else
    bashcache "$@"
    exit $?
fi
unset sourced

return 0
