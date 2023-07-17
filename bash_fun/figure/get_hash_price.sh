#!/bin/bash
# This file contains several functions for getting and caching HASH token info.
# This file should be sourced to make the functions available in your environment.
#
# Primary Functions of Interest:
#   get_hash_price  ------------- Gets the current price of a HASH token, e.g. 0.100000000000000000.
#   get_hash_price_for_prompt  -- Formats the output of get_hash_price.
#
# Other Functions:
#   hashcache  -------------------------- A wrapper over bashcache that applies standard HASH args.
#   hashcache_refresh  ------------------ Gets and caches the hash price json and hash price value.
#   hashcache_check_required_commands  -- Checks that some required commands are available.
#   hashcache_check_required_env_vars  -- Checks that the required enviroment variables are defined.
#
# Customizable Setup Environment Variables:
#   These are only used when this file is being sourced.
#   HASH_PRICE_SOURCE  ----- The source to use for getting the hash price.
# Customizable Environment Variables:
#   HASH_C_DIR  ------------ The directory bashcache uses for this stuff.
#   HASH_C_MAX_AGE  -------- The maximum age of the cached data (before triggering a refresh).
#   HASH_PRICE_URL  -------- The URL that will return the needed data.
#   HASH_JQ_FILTER  -------- The filter to apply to the JSON result of the URL.
#   HASH_DEFAULT_VALUE  ---- The Hash price to use if something is going wrong.
#   HASH_PROMPT_FORMAT  ---- The format to use for the Hash price in a command line prompt.
#   See below for details and defaults.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced


############################################
# Customizable Setup Environment Variables
#-----------------------------------------

# These environment variables are only used when this file is sourced.
# That is, changing them after sourcing this file won't affect the behavior of the get_hash_price function.

# The source to use to look up the hash price.
# Options: 'dlob'  'yahoo' 'custom'
# The default is yahoo.
# If HASH_PRICE_URL or HASH_JQ_FILTER are already set, they won't be changed.
# If set to 'custom', HASH_PRICE_URL and HASH_JQ_FILTER aren't set and warnings about missing env vars is suppressed.
# That means re-sourcing this file will not change them unless they're unset first,
# e.g. HASH_PRICE_SOURCE=dlob && unset HASH_JQ_FILTER HASH_PRICE_URL && source get_hash_price.sh
case "${HASH_PRICE_SOURCE:-yahoo}" in
    dlob)
        HASH_PRICE_URL="${HASH_PRICE_URL:-https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price}"
        HASH_JQ_FILTER="${HASH_JQ_FILTER:-.latestDisplayPricePerDisplayUnit}"
        ;;
    yahoo)
        HASH_PRICE_URL="${HASH_PRICE_URL:-https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=symbol,shortName,regularMarketPrice&symbols=HASH1-USD}"
        HASH_JQ_FILTER="${HASH_JQ_FILTER:-.quoteResponse.result[0].regularMarketPrice}"
        ;;
    custom)
        # This option exists as a way to suppress the complaints about a wrong HASH_PRICE_SOURCE or missing env vars.
        # There's nothing to actually set for it though.
        ;;
    *)
        printf 'Warning: Unknown HASH_PRICE_SOURCE: [%s]\nMust be "", "dlob", "yahoo", or "custom".\n' "$HASH_PRICE_SOURCE" >&2
        ;;
esac

######################################
# Customizable Environment Variables
#-----------------------------------

# These environment variables are used whenever the functions in this file are executed.
# That is, changing them after this file is sourced WILL affect the behavior of get_hash_price.

# The directory that bashcache uses in here.
# The path must be absolute.
# Default is '/tmp/hash'.
HASH_C_DIR="${HASH_C_DIR:-/tmp/hash}"

# The maximum cache age that bashcache uses in here.
# The format is the same as used for the -atime option of the find command, without the +. E.g. '10m' or '23h' or '6d12h30m'.
# When get_hash_price (or get_hash_price_for_prompt) is called, if the cache is older than this, a refresh is triggered.
# Default is '10m' (ten minutes).
HASH_C_MAX_AGE="${HASH_C_MAX_AGE:-10m}"

# The url to request.
# Yahoo Finance: https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=symbol,regularMarketPrice&symbols=HASH1-USD
# DLOB: https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price
# Default is Yahoo Finance (see HASH_PRICE_SOURCE above).
HASH_PRICE_URL="$HASH_PRICE_URL"

# The filter given to jq in order to extract the desired value out of the result found at HASH_PRICE_URL.
# Yahoo Finance: .quoteResponse.result[0].regularMarketPrice
# DLOB: .latestDisplayPricePerDisplayUnit
# Default is Yahoo Finance (see HASH_PRICE_SOURCE above).
HASH_JQ_FILTER="$HASH_JQ_FILTER"

# A value to use when either there is an error or we don't have any data yet.
# Default is -69.42 (with a bunch of zeros to make it the same length as an expected value).
HASH_DEFAULT_VALUE="${HASH_DEFAULT_VALUE:--69.420000000000000000}"

# The format to use for the prompt.
# Default explained:
#   In a dark gray background (48;5;238), with bright white text (38;5;15),
#   Print a space
#   Print a # then an emoji that nudges it right and puts a rounded box around it.
#   Print two spaces because that # + emoji overlaps the next character and I want a space there.
#   Print the hash price rounded to 4 decimal places.
#   Print one last space for padding.
#   Turn off coloring and be done
HASH_PROMPT_FORMAT="${HASH_PROMPT_FORMAT:-\033[48;5;238;38;5;15m #\xE2\x83\xA3  %1.4f \033[0m}"


################################
# Static Environment Variables
#-----------------------------

# Define some bashcache names for storing various things.
HASH_CN_HASH_PRICE='hash_price'         # The Hash Price of interest.
HASH_CN_PRICE_JSON='price_json'         # The full response (hopefully json) from the curl command.
HASH_CN_PRICE_HEADER='price_header'     # The response header from the curl command.
HASH_CN_JQ_ERROR='jq_error'             # Any errors encountered using jq.


#################################
# Primary Functions of Interest
#------------------------------

# Usage: get_hash_price [--refresh] [--no-wait]
# Outputs the value of a HASH token (in USD), e.g. 0.105000000000000000 with a newline at the end.
# Caching with bashcache is used to prevent spamming of the api.
# If the cache is fresh, the value is printed, and nothing special happens.
# If the cache is empty, it will be updated then printed, which can take some time.
# If the cache is stale (more than 10 minutes old), the known value will be printed,
#   and a background process will be initiated to update the cache.
# The --refresh flag can be provided to force the cache to be refreshed.
#   By default, this happens in the foreground, and you'll have to wait for it.
# The --no-wait flag indicates that all refreshes should happen in the background, and you want a value immediately.
#   If the cache doesn't exist, a default is printed. Otherwise, whatever is cached is printed.
#   If a refresh is requested or needed, it is initiated in the background.
# Exit codes:
#   0: The data exists and is fresh (from bashcache).
#   1: The arguments provided to bashcache were incorrect (from bashcache).
#   2: A required command is missing, and get_hash_price cannot work.
#   3: Illegal arguments provided to get_hash_price.
#   4: A required environment variable is missing, and get_hash_price cannot work.
#   10: The cached data was available but stale (from bashcache).
#   11: The cached data was not available (from bashcache).
get_hash_price () {
    if ! hashcache_check_required_commands > /dev/null 2>&1; then
        printf '%s\n' "$HASH_DEFAULT_VALUE"
        return 2
    fi
    if ! hashcache_check_required_env_vars > /dev/null 2>&1; then
        printf '%s\n' "$HASH_DEFAULT_VALUE"
        return 4
    fi
    local no_wait force_refresh hash_price cache_read_code
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --no-wait) no_wait="$1";;
            --refresh) force_refresh="$1";;
            *)
                printf 'Unknown argument: [%s].\n' "$1"
                return 3
                ;;
        esac
        shift
    done

    # If we're forcing a refresh, and don't mind waiting, do the refresh now.
    [[ -n "$force_refresh" && -z "$no_wait" ]] && hashcache_refresh

    # This will either get the cached hash price if we have it, or it'll be an empty string.
    hash_price="$( hashcache read "$HASH_CN_HASH_PRICE" )"
    cache_read_code=$?

    # If we didn't want to wait, but wanted to force a refresh, kick that off now.
    [[ -n "$force_refresh" && -n "$no_wait" ]] && hashcache_refresh --background

    # No matter what, if there were invalid bashcache arguments, we want to know, so handle that now.
    if [[ "$cache_read_code" -eq '1' ]]; then
        # 1 - Invalid arguments provided to the bashcache command.
        printf 'Invalid arguments to bashcache.\n' >&2
    fi

    # If we aren't forcing a refresh, check the exit code to see if one is needed.
    if [[ -z "$force_refresh" ]]; then
        case "$cache_read_code" in
        10)
            # 10 - The requested cache data is available, but stale.
            # Update the cache in the background.
            hashcache_refresh --background
            ;;
        11)
            # 11 - The requested cache data is not available.
            if [[ -n "$no_wait" ]]; then
                # We don't want to wait, kick off an update in the background and move on.
                hashcache_refresh --background
            else
                # We don't mind waiting, refresh it now, wait for it, then read it.
                hashcache_refresh
                hash_price="$( hashcache read "$HASH_CN_HASH_PRICE" )"
                cache_read_code=$?
            fi
            ;;
        esac
    fi
    [[ -z "$hash_price" ]] && hash_price="$HASH_DEFAULT_VALUE"
    printf '%s\n' "$hash_price"
    return $cache_read_code
}

# Usage: get_hash_price_for_prompt
# This applies the HASH_PROMPT_FORMAT format to the result of get_hash_price.
# This is intended to be used in a command prompt, e.g. PS1='$( get_hash_price_for_prompt ) $'.
# The exit code returned from this function will be the same as the previous exit code (from before this function is called).
get_hash_price_for_prompt () {
    local previous_exit=$?
    printf "$HASH_PROMPT_FORMAT" "$( get_hash_price --no-wait )"
    return $previous_exit
}


###################
# Other Functions
#----------------

# Usage: hashcache <command> <cache name> [options]
# This is just a wrapper over bashcache to provide the standard directory and age arguments.
hashcache () {
    bashcache -d "$HASH_C_DIR" -a "$HASH_C_MAX_AGE" "$@"
}

# Usage: hashcache_refresh [-v|-vv|-vvv|--background]
# Gets the HASH_PRICE_URL, applies the HASH_JQ_FILTER and caches it.
# The -v -vv and -vvv flags are different levels of verbosity.
# The --background flag will cause this to do this in the background, and return from the call immediately.
hashcache_refresh () {
    if [[ "$1" == '--background' ]]; then
        # Fire off a background process to update the cache.
        # The () > /dev/null 2>&1 here is to supress the job/pid start and stop messages.
        ( hashcache_refresh & ) > /dev/null 2>&1
        return 0
    fi
    local v bcv val ec
    # Set the verbosity level
    v="$( sed 's/[^v]//g' <<< "$1" | awk '{print length}' )"
    # If the verbosity is high enough, set the bashcache verbosity flag
    [[ "$v" -ge '3' ]] && bcv='--verbose'

    # If -vv or more, output all the HASH variables.
    [[ "$v" -ge '2' ]] && { set | grep '^HASH_'; printf '\n'; } >&2

    # Check the required commands without printing anything (unless were runnin verbosely).
    if ! hashcache_check_required_commands > /dev/null 2>&1; then
        [[ "$v" -ge '1' ]] && printf 'Missing required command(s). Run hashcache_check_required_commands for more info.\n' >&2
        return 20
    fi

    # Curl the url storing both the header and output into the cache.
    [[ "$v" -ge '1' ]] && printf 'Curling url: %s ... ' "$HASH_PRICE_URL" >&2
    [[ -n "$bcv" ]] && printf '\n' >&2
    curl -s "$HASH_PRICE_URL" \
         --dump-header "$( hashcache file "$HASH_CN_PRICE_HEADER" $bcv )" \
         --output "$( hashcache file "$HASH_CN_PRICE_JSON" $bcv )" 2> /dev/null
    ec=$?
    [[ "$v" -ge '1' ]] && printf 'Done. Exit code: %d\n' "$ec" >&2
    if [[ "$ec" -ne '0' && "$v" -ge '1' || "$v" -ge '2' ]]; then
        printf 'Response header file: %s\n' "$( hashcache file "$HASH_CN_PRICE_HEADER" )" >&2
        [[ "$v" -ge '2' ]] && hashcache read "$HASH_CN_PRICE_HEADER" >&2
        printf 'Response content file: %s\n' "$( hashcache file "$HASH_CN_PRICE_JSON" )" >&2
        [[ "$v" -ge '2' ]] && { hashcache read "$HASH_CN_PRICE_JSON"; printf '\n\n'; } >&2
    fi

    if [[ "$ec" -eq '0' ]]; then
        # Apply the jq filter to the newly cached result to get the desired value.
        [[ "$v" -ge '1' ]] && printf 'Applying jq filter '"'"'%s'"'"' ... ' "$HASH_JQ_FILTER" >&2
        [[ -n "$bcv" ]] && printf '\n' >&2
        val="$( jq -r "$HASH_JQ_FILTER" "$( hashcache file "$HASH_CN_PRICE_JSON" $bcv )" 2> "$( hashcache file "$HASH_CN_JQ_ERROR" $bcv )" )"
        ec=$?
        [[ "$v" -ge '1' ]] && printf 'Done. Exit code: %d\n' "$ec" >&2
        if [[ "$ec" -ne '0' && "$v" -ge '1' || "$v" -ge '2' ]]; then
            if [[ "$ec" -eq '0' ]]; then
                printf 'Result: %s\n' "$val" >&2
            else
                printf 'Error file: %s\n' "$( hashcache file "$HASH_CN_JQ_ERROR" )" >&2
                [[ "$v" -ge '2' ]] && { hashcache read "$HASH_CN_JQ_ERROR"; printf '\n'; } >&2
            fi
        fi
        [[ "$v" -ge '2' ]] && printf '\n' >&2
    fi

    # If there as a problem, use the default value.
    if [[ "$ec" -ne '0' || -z "$val" ]]; then
        [[ "$v" -ge '1' ]] && printf 'Using default value: %s\n' "$HASH_DEFAULT_VALUE" >&2
        val="$HASH_DEFAULT_VALUE"
        [[ "$ec" -eq '0' ]] && ec=21
    fi

    # Write the value to the cache.
    hashcache write "$HASH_CN_HASH_PRICE" $bcv -- "$val"
    [[ "$v" -ge '1' ]] && printf 'Value: %s\nCached in: %s\n' "$val" "$( hashcache file "$HASH_CN_HASH_PRICE" )" >&2

    return $ec
}

# Usage: hashcache_check_required_commands
# This checks for some required commands that have at least a little chance of not being available.
# If a command is missing, some info will be printed to stderr and the exit code won't be 0.
# An exit code of zero means everything is available.
hashcache_check_required_commands () {
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


# Usage: hashcache_check_required_env_vars
# This checks that the required environment variables have values.
# If one or more missing, an error is printed to stderr and the exit code will be 1.
# An exit code of zero means everything is available.
hashcache_check_required_env_vars () {
    local rv use_p var val
    rv=0
    # With ${(P)rv}, Bash will output -bash: ${(P)rv}: bad substitution
    # But it's not part of command output, so you can't redirect it.
    # This is trickier than pvarn because this gets called as the shell is loading.
    # Doing it in a ( ... ) subshell causes a freeze that you have to ctrl+c and type "exit" to get out of.
    # Doing it in a { ... } subshell makes it just halt this function's execution at that point.
    # After trying all sorts of combos of exist and returns and subshells and spacing, this is what I found that actually works
    # It runs just fine as the shell is first loading, it suppresses the "bad substitution" message,
    # and properly identifies which variable expansion syntax to use.
    { use_p="$( foo="${(P)bar}" && printf 'YES' )"; } > /dev/null 2>&1
    for var in 'HASH_C_DIR' 'HASH_C_MAX_AGE' 'HASH_PRICE_URL' 'HASH_JQ_FILTER' 'HASH_DEFAULT_VALUE' 'HASH_PROMPT_FORMAT' 'HASH_CN_HASH_PRICE' 'HASH_CN_PRICE_JSON' 'HASH_CN_PRICE_HEADER' 'HASH_CN_JQ_ERROR'; do
        if [[ -n "$use_p" ]]; then
            val="${(P)var}"
        else
            val="${!var}"
        fi
        if [[ -z "$val" ]]; then
            [[ "$HASH_PRICE_SOURCE" == 'custom' ]] || printf 'Environment variable %s not defined.\n' "$var" >&2
            rv=1
        fi
    done
    return "$rv"
}

# Run the checks now to print out any problems.
# Always run both, but only run each one time. If either fails, this script should return with a non-zero exit code.
# Since this stuff is desiged for use in the command prompt, it'd be horribly annoying if it were printing out errors with every prompt.
# By running this check now, you find out about problems as the file's being sourced and can still investigate if get_hash_price_for_prompt is in your prompt.
if hashcache_check_required_commands; then
    hashcache_check_required_env_vars
else
    hashcache_check_required_env_vars
    return 1
fi
