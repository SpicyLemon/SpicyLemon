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
    local exit_code_ok exit_code_bad_args exit_code_stale_data exit_code_no_data default_cache_dir default_max_age usage
    exit_code_ok=0
    exit_code_bad_args=1
    exit_code_stale_data=10
    exit_code_no_data=11
    default_cache_dir='/tmp/bashcache'
    default_max_age='24h'
    usage="$( cat << EOF
bashcache: A bash command-line caching utility.

Usage: bashcache <command> <cache name> [<options>]

    The <command> is required and must be one of: write, read, check, file, list, delete.
    The <cache name> is required (except for with the list command) and must not be any of the commands.
        It represents the name of the cached item to act on.
        Cache names cannot start with a dash.

    Options:
        -v --verbose                    Provide extra information while executing the desired command.
        -d --dir <directory>            Dictate the desired directory to use for cached information.
                                        This must be an absolute path starting with a slash.
                                        See also: BASHCACHE_DIR
        -a --age --max-age <max age>    Dictate the desired max age value.
                                        The format is the same as used for the -atime option of the find command, without the +.
                                        Examples: '10m' '23h' '6d12h30m'
                                        See also: BASHCACHE_MAX_AGE

    Commands:
        write
            Stores the provided data as the given cache name.
            Data can be piped in, provided as a here doc, here string, or on the command line after the -- option.
            Examples:
                echo "This is my data." | bashcache write foo
                echo "This is my data." > foo-file.txt && bashcache write foo << foo-file.txt
                bashcache write foo <<< "This is my data."
                bashcache write foo -- This is my data.
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
        file
            This is the same as the check command except it also outputs the filename being used.
            This is handy when you need to provide a filename to some other program, but still want it
            managed by bashcache.
        list
            Outputs a list of cache names.
            When supplying this command, the cache name should be omitted.
            This command also has a --details option, which will cause it to also output size and date information.
        delete
            Deletes the desired cache name.

    Defaults:
        The default directory for cached information is '$default_cache_dir'.
        To change this, set the BASHCACHE_DIR environment variable to the desired directory.
        The -d or --dir option takes precedence over the BASHCACHE_DIR value (or default if not set)

        The default max age is '$default_max_age'.
        To change this, set the BASHCACHE_MAX_AGE environment variable to the desired value.
        The -a or --age or --max-age option takes precedence over the BASHCACHE_MAX_AGE value (or default if not set)
EOF
    )"
    local cache_command cache_name verbose cache_dir max_age details input_from_args cache_file
    [[ "$#" -eq '0' ]] && set -- -h
    while [[ "$#" -gt '0' ]]; do
        case "$( tr '[:upper:]' '[:lower:]' <<< $1 )" in
        -h|--help)
            printf '%s\n\n' "$usage"
            return $exit_code_ok
            ;;
        -v|--verbose)
            verbose="$1"
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
            details="$1"
            ;;
        write|read|check|file|list|delete)
            if [[ -n "$cache_command" ]]; then
                printf 'bashcache: Only one cache command can be provided. Found %s then %s.\n' "$cache_command" "$1" >&2
                return $exit_code_bad_args
            fi
            cache_command="$1"
            ;;
        --)
            shift
            [[ -n "$verbose" ]] && printf 'bashcache: Found --. Getting input from remaining arguments.\n' >&2
            input_from_args="$*"
            set -- --
            ;;
        -*)
            printf 'bashcache: Unknown option: [%s].\n' "$1" >&2
            return $exit_code_bad_args
            ;;
        *)
            if [[ -n "$cache_name" ]]; then
                printf 'bashcache: Unknown option: [%s].\n' "$1" >&2
                return $exit_code_bad_args
            fi
            cache_name="$1"
            ;;
        esac
        shift
    done

    if [[ -z "$cache_command" ]]; then
        printf 'bashcache: No command provided.\n' >&2
        return $exit_code_bad_args
    fi
    cache_command="$( tr '[:upper:]' '[:lower:]' <<< "$cache_command" )"
    [[ -n "$verbose" ]] && printf 'bashcache: Cache command: [%s].\n' "$cache_command" >&2

    if [[ "$cache_command" != 'list' ]]; then
        if [[ -z "$cache_name" ]]; then
            printf 'bashcache: No cache name provided.\n' >&2
            return $exit_code_bad_args
        fi
        [[ -n "$verbose" ]] && printf 'bashcache: Cache name: [%s].\n' "$cache_name" >&2
        if [[ -n "$details" ]]; then
            printf 'bashcache: Unknown option: [%s].\n' "$details" >&2
            return $exit_code_bad_args
        fi
    elif [[ -n "$cache_name" ]]; then
        printf 'bashcache: Unknown option: [%s].\n' "$cache_name" >&2
        return $exit_code_bad_args
    fi

    if [[ -n "$cache_dir" ]]; then
        [[ -n "$verbose" ]] && printf 'bashcache: Directory provided from argument: [%s].\n' "$cache_dir" >&2
    elif [[ -n "$BASHCACHE_DIR" ]]; then
        cache_dir="$BASHCACHE_DIR"
        [[ -n "$verbose" ]] && printf 'bashcache: Directory provided from BASHCACHE_DIR environment variable: [%s].\n' "$cache_dir" >&2
    else
        cache_dir="$default_cache_dir"
        [[ -n "$verbose" ]] && printf 'bashcache: Directory set from default: [%s].\n' "$cache_dir" >&2
    fi
    if [[ "$cache_dir" =~ ^[^/] ]]; then
        printf 'bashcache: Invalid directory: [%s]. It must start with a slash.\n' "$cache_dir" >&2
        return $exit_code_bad_args
    fi

    if [[ -n "$max_age" ]]; then
        [[ -n "$verbose" ]] && printf 'bashcache: Max age provided from argument: [%s].\n' "$max_age" >&2
    elif [[ -n "$BASHCACHE_MAX_AGE" ]]; then
        max_age="$BASHCACHE_MAX_AGE"
        [[ -n "$verbose" ]] && printf 'bashcache: Max age provided from BASHCACHE_MAX_AGE environment variable: [%s].\n' "$max_age" >&2
    else
        max_age="$default_max_age"
        [[ -n "$verbose" ]] && printf 'bashcache: Max age set from default: [%s].\n' "$max_age" >&2
    fi
    if [[ ! "$max_age" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
        printf 'bashcache: Invalid max age: [%s].\n' "$max_age" >&2
        return $exit_code_bad_args
    fi

    if [[ ! -d "$cache_dir" ]]; then
        [[ -n "$verbose" ]] && printf 'bashcache: Creating cache directory: [%s].\n' "$cache_dir" >&2
        if ! mkdir -p "$cache_dir" 1>&2; then
            printf 'bashcache: Unable to create cache directory: [%s].\n' "$cache_dir" >&2
            return $exit_code_bad_args
        fi
        [[ -n "$verbose" ]] && printf 'bashcache: Cache directory created successfully: [%s].\n' "$cache_dir" >&2
    else
        [[ -n "$verbose" ]] && printf 'bashcache: Cache directory exists: [%s]\n' "$cache_dir" >&2
    fi

    cache_file="${cache_dir}/${cache_name}"

    case "$cache_command" in
    write)
        if [[ -n "$input_from_args" ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Writing cache file from data provided as arguments: [%s].\n' "$cache_file" >&2
            printf '%s' "$input_from_args" > "$cache_file"
        else
            [[ -n "$verbose" ]] && printf 'bashcache: Writing cache file from data provided using stdin: [%s].\n' "$cache_file" >&2
            cat - > "$cache_file"
        fi
        if [[ -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Done caching %s of data in [%s].\n' "$( ls -lh "$cache_file" | awk -F " " '{print $5}' )" "$cache_file" >&2
            return $exit_code_ok
        else
            [[ -n "$verbose" ]] && printf 'bashcache: Failed to cache data to [%s].\n' "$cache_file" >&2
            return $exit_code_bad_args
        fi
        ;;
    read|check|file)
        if [[ "$cache_command" == 'file' ]]; then
            printf '%s' "$cache_file"
            [[ -n "$verbose" && -t 1 ]] && printf '\n' >&2
        fi
        if [[ ! -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Cache does not exist [%s].\n' "$cache_file" >&2
            return $exit_code_no_data
        fi
        if [[ "$cache_command" == 'read' ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Cache exists [%s], outputting data.\n' "$cache_file" >&2
            [[ -n "$verbose" && -t 1 ]] && printf '[' >&2
            cat "$cache_file"
            [[ -n "$verbose" && -t 1 ]] && printf ']\n' >&2
        fi
        if [[ -n "$( find "$cache_file" -mtime "+$max_age" )" ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Cache is stale [%s] (older than [%s]).\n' "$cache_file" "$max_age" >&2
            return $exit_code_stale_data
        fi
        [[ -n "$verbose" ]] && printf 'bashcache: Cache is sufficiently fresh [%s] (not older than [%s]).\n' "$cache_file" "$max_age" >&2
        return $exit_code_ok
        ;;
    list)
        if [[ -n "$details" ]]; then
            # list files using -l (long format) -h (human readible sizes) and -T (complete time info) and -t (sort by last modified)
            # Grep for just the file lines. Directories will start with a 'd' and there's also a "total" line to ignore.
            # Get rid of the first 4 columns: permissions, number of links, owner name, group name.
            #   But maintain left-padding whitespace on the sizes so things still line up nicely (use {2} instead of + or *).
            ls -lhTt "$cache_dir/" | grep '^-' | sed -E 's/^([^[:space:]]+[[:space:]]+){3}[^[:space:]]+[[:space:]]{2}//'
        else
            ls "$cache_dir/"
        fi
        [[ -n "$verbose" && ! "$( ls "$cache_dir/" )" =~ [^[:space:]] ]] && printf 'bashcache: Cache directory is empty: [%s].\n' "$cache_dir" >&2
        return $exit_code_ok
        ;;
    delete)
        if [[ -f "$cache_file" ]]; then
            [[ -n "$verbose" ]] && printf 'bashcache: Deleting cache file: [%s].\n' "$cache_file" >&2
            rm "$cache_file"
            if [[ -f "$cache_file" ]]; then
                [[ -n "$verbose" ]] && printf 'bashcache: Could not delete cache file: [%s]\n' "$cache_file" >&2
                return $exit_code_bad_args
            fi
            [[ -n "$verbose" ]] && printf 'bashcache: Cache file deleted: [%s]\n' "$cache_file" >&2
            return $exit_code_ok
        fi
        [[ -n "$verbose" ]] && printf 'bashcache: Cache does not exist to delete: [%s].\n' "$cache_file" >&2
        return $exit_code_no_data
        ;;
    *)
        printf 'bashcache: Unhandled command: [%s].\n' "$cache_command" >&2
        return $exit_code_bad_args
        ;;
    esac

    [[ -n "$verbose" ]] &&  printf 'bashcache: Command [%s] did not specify its own exit code. This is a bug.\n' "$cache_command" >&2
    return $exit_code_bad_args
}

if [[ "$sourced" != 'YES' ]]; then
    bashcache "$@"
    exit $?
fi
unset sourced

return 0
