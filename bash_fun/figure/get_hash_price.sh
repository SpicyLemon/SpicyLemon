#!/bin/bash
# This file contains several functions for getting and caching HASH token info from dlob.io.
# This file should be sourced to make the functions available in your environment.
#
# Primary Functions of Interest:
#   get_hash_price  ------------- Gets the current price of a HASH token, e.g. 0.100000000000000000.
#   get_hash_price_for_prompt  -- Same as get_hash_price with less digits, e.g. 0.1000, and no ending newline.
#
# Other Functions:
#   dlob_cache_refresh  ------------------ Gets and caches the daily price json and hash price value.
#   dlob_cache_check_required_commands  -- Checks that some required commands are available.
#
# Customizable Environment Variables:
#   DLOB_C_DIR  ------------ The directory bashcache uses for this stuff.
#   DLOB_C_MAX_AGE  -------- The maximum age of the cached data (before triggering a refresh).
#   DLOB_DAILY_PRICE_URL  -- The URL with the needed data.
#   DLOB_JQ_FILTER  -------- The filter to apply to the JSON result of the URL.
#   DLOB_DEFAULT_VALUE  ---- The Hash price to use if something is going wrong.
#   DLOB_PROMPT_FORMAT  ---- The format to use for the Hash price in a command line prompt.
#   See below for details and defaults.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced


######################################
# Customizable Environment Variables
#-----------------------------------

# The directory that bashcache uses in here.
# The path must be absolute.
# Default is '/tmp/dlob'.
DLOB_C_DIR="${DLOB_C_DIR:-/tmp/dlob}"

# The maximum cache age that bashcache uses in here.
# The format is the same as used for the -atime option of the find command, without the +. E.g. '10m' or '23h' or '6d12h30m'.
# When get_hash_price (or get_hash_price_for_prompt) is called, if the cache is older than this, a refresh is triggered.
# Default is '10m' (ten minutes).
DLOB_C_MAX_AGE="${DLOB_C_MAX_AGE:-10m}"

# The url to request.
# Defaults to 'https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price'.
DLOB_DAILY_PRICE_URL="${DLOB_DAILY_PRICE_URL:-https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price}"

# The filter given to jq in order to extract the desired value out of the result found at DLOB_DAILY_PRICE_URL.
# Default is '.latestDisplayPricePerDisplayUnit'.
DLOB_JQ_FILTER="${DLOB_JQ_FILTER:-.latestDisplayPricePerDisplayUnit}"

# A value to use when either there is an error or we don't have any data yet.
# Default is -69.42 (with a bunch of zeros to make it the same length as an expected value).
DLOB_DEFAULT_VALUE="${DLOB_DEFAULT_VALUE:--69.420000000000000000}"

# The format to use for the prompt.
# Default explained:
#   In a dark gray background (48;5;238), with bright white text (38;5;15),
#   Print a space
#   Print a # then an emoji that nudges it right and puts a rounded box around it.
#   Print two spaces because that # + emoji overlaps the next character and I want a space there.
#   Print the hash price rounded to 4 decimal places.
#   Print one last space for padding.
#   Turn off coloring and be done
DLOB_PROMPT_FORMAT="${DLOB_PROMPT_FORMAT:-\033[48;5;238;38;5;15m #\xE2\x83\xA3  %1.4f \033[0m}"


################################
# Static Environment Variables
#-----------------------------

# Define some bashcache names for storing various things.
DLOB_CN_HASH_PRICE='hash_price'                     # The Hash Price of interest.
DLOB_CN_DAILY_PRICE_JSON='daily_price_json'         # The full response (hopefully json) from the curl command.
DLOB_CN_DAILY_PRICE_HEADER='daily_price_header'     # The response header from the curl command.
DLOB_CN_JQ_ERROR='jq_error'                         # Any errors encountered using jq.


#################################
# Primary Functions of Interest
#------------------------------

# Usage: get_hash_price
# Outputs the value of a HASH token (in USD), e.g. 0.105000000000000000 with a newline at the end.
# Caching is done using bashcache so that multiple shells will have access to the same data.
# If the cache is fresh, the value is printed, and nothing more happens.
# If the cache is stale, the stale value is printed, and a background process is initiated to update the cache.
# If nothing is cached yet, the DLOB_DEFAULT_VALUE value is printed, and a background process is initiated to update the cache.
# If a required command is missing, The exit code will be 20.
# Otherwise, the exit code will be the same as the bashcache exit code.
get_hash_price () {
    if ! dlob_cache_check_required_commands > /dev/null 2>&1; then
        printf '%s\n' "$DLOB_DEFAULT_VALUE"
        return 20
    fi
    local cache_read_code
    # This will either output the cached hash price if we have it, or it won't output anything.
    bashcache read "$DLOB_CN_HASH_PRICE" -d "$DLOB_C_DIR" -a "$DLOB_C_MAX_AGE"
    cache_read_code=$?
    case "$cache_read_code" in
    0)
        # 0 - The requested cache data is available and up-to-date.
        # Do nothing.
        ;;
    1)
        # 1 - Invalid arguments provided to the bashcache command.
        printf 'Invalid arguments to bashcache.\n' >&2
        ;;
    10|11)
        # 10 - The requested cache data is available, but stale.
        # 11 - The requested cache data is not available.
        # Eiother way, we want to refresh the data.
        # If it's 11, output the error value since the read command above didn't output anything.
        [[ "$cache_read_code" -eq '11' ]] && printf '%s\n' "$DLOB_DEFAULT_VALUE"
        # Fire off a background process to update it for next time.
        # The () > /dev/null 2>&1 here is to supress the job/pid start and stop messages.
        ( dlob_cache_refresh & ) > /dev/null 2>&1
        ;;
    *)
        printf 'Unexpected bashcache exit code: [%d]\n' "$cache_read_code" >&2
        ;;
    esac
    return $cache_read_code
}

# Usage: get_hash_price_for_prompt
# This applies the DLOB_PROMPT_FORMAT format to the result of get_hash_price.
# This is intended to be used in a command prompt, e.g. PS1='$( get_hash_price_for_prompt ) $'.
# The exit code returned from this function will be the same as the previous exit code (from before this function is called).
get_hash_price_for_prompt () {
    local previous_exit=$?
    printf "$DLOB_PROMPT_FORMAT" "$( get_hash_price )"
    return $previous_exit
}


###################
# Other Functions
#----------------

# Usage: dlob_cache_check_required_commands
# This checks for some required commands that have at least a little chance of not being available.
# If a command is missing, some info will be printed to stderr and the exit code won't be 0.
# An exit code of zero means everything is available.
dlob_cache_check_required_commands () {
    local r c
    r=0
    for c in 'curl' 'jq' 'bashcache'; do
        if ! command -v "$c" > /dev/null 2>&1; then
            printf 'Missing required command: %s\n' "$c" >&2
            printf 'The functions from %s might not work correctly.\n' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" >&2
            command "$c" >&2
            r=$?
        fi
    done
    return $r
}

# Usage: dlob_cache_refresh [-v|-vv|-vvv]
# Gets the DLOB_DAILY_PRICE_URL, applies the DLOB_JQ_FILTER and caches it.
dlob_cache_refresh () {
    local v bcv val ec
    # Set the verbosity level
    v="$( sed 's/[^v]//g' <<< "$1" | awk '{print length}' )"
    # If the verbosity is high enough, set the bashcache verbosity flag
    [[ "$v" -ge '3' ]] && bcv='--verbose'

    # If -vv or more, output all the DLOB variables.
    [[ "$v" -ge '2' ]] && { set | grep '^DLOB_'; printf '\n'; } >&2

    # Check the required commands without printing anything (unless were runnin verbosely).
    if ! dlob_cache_check_required_commands > /dev/null 2>&1; then
        [[ "$v" -ge '1' ]] && printf 'Missing required command(s). Run dlob_cache_check_required_commands for more info.\n' >&2
        return 20
    fi

    # Curl the url storing both the header and output into the cache.
    [[ "$v" -ge '1' ]] && printf 'Curling url: %s ... ' "$DLOB_DAILY_PRICE_URL" >&2
    curl -s "$DLOB_DAILY_PRICE_URL" \
         --dump-header "$( bashcache file "$DLOB_CN_DAILY_PRICE_HEADER" -d "$DLOB_C_DIR" $bcv )" \
         --output "$( bashcache file "$DLOB_CN_DAILY_PRICE_JSON" -d "$DLOB_C_DIR" $bcv )" 2> /dev/null
    ec=$?
    [[ "$v" -ge '1' ]] && printf 'Done. Exit code: %d\n' "$ec" >&2
    if [[ "$ec" -ne '0' && "$v" -ge '1' || "$v" -ge '2' ]]; then
        printf 'Response header file: %s\n' "$( bashcache file "$DLOB_CN_DAILY_PRICE_HEADER" -d "$DLOB_C_DIR" $bcv )" >&2
        [[ "$v" -ge '2' ]] && bashcache read "$DLOB_CN_DAILY_PRICE_HEADER" -d "$DLOB_C_DIR" $bcv >&2
        printf 'Response content file: %s\n' "$( bashcache file "$DLOB_CN_DAILY_PRICE_JSON" -d "$DLOB_C_DIR" $bcv )" >&2
        [[ "$v" -ge '2' ]] && { bashcache read "$DLOB_CN_DAILY_PRICE_JSON" -d "$DLOB_C_DIR" $bcv; printf '\n\n'; } >&2
    fi

    if [[ "$ec" -eq '0' ]]; then
        # Apply the jq filter to the newly cached result to get the desired value.
        [[ "$v" -ge '1' ]] && printf 'Applying jq filter '"'"'%s'"'"' ... ' "$DLOB_JQ_FILTER" >&2
        val="$( jq -r "$DLOB_JQ_FILTER" "$( bashcache file "$DLOB_CN_DAILY_PRICE_JSON" -d "$DLOB_C_DIR" $bcv )" 2> "$( bashcache file "$DLOB_CN_JQ_ERROR" -d "$DLOB_C_DIR" $bcv )" )"
        ec=$?
        [[ "$v" -ge '1' ]] && printf 'Done. Exit code: %d\n' "$ec" >&2
        if [[ "$ec" -ne '0' && "$v" -ge '1' || "$v" -ge '2' ]]; then
            if [[ "$ec" -eq '0' ]]; then
                printf 'Result: %s\n' "$val" >&2
            else
                printf 'Error file: %s\n' "$( bashcache file "$DLOB_CN_JQ_ERROR" -d "$DLOB_C_DIR" $bcv )" >&2
                [[ "$v" -ge '2' ]] && { bashcache read "$DLOB_CN_JQ_ERROR" -d "$DLOB_C_DIR" $bcv; printf '\n'; } >&2
            fi
        fi
        [[ "$v" -ge '2' ]] && printf '\n' >&2
    fi

    # If there as a problem, use the default value.
    if [[ "$ec" -ne '0' || -z "$val" ]]; then
        [[ "$v" -ge '1' ]] && printf 'Using default value: %s\n' "$DLOB_DEFAULT_VALUE" >&2
        val="$DLOB_DEFAULT_VALUE"
        [[ "$ec" -eq '0' ]] && ec=21
    fi

    # Write the value to the cache.
    bashcache write "$DLOB_CN_HASH_PRICE" -d "$DLOB_C_DIR" $bcv -- "$val"
    [[ "$v" -ge '1' ]] && printf 'Value: %s\nCached in: %s\n' "$val" "$( bashcache file "$DLOB_CN_HASH_PRICE" -d "$DLOB_C_DIR" $bcv )" >&2

    return $ec
}

# Run the check now to print out any problems.
# Since this stuff is desiged for use in the command prompt, it'd be horribly annoying if it were printing out errors with every prompt.
# By running this check now, you find out about problems as the file's being sourced and can still investigate if get_hash_price_for_prompt is in your prompt.
dlob_cache_check_required_commands
