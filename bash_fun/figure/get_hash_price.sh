#!/bin/bash
# This file contains several functions for getting and caching HASH token info from dlob.io.
# This file should be sourced to make the functions available in your environment.
#
# Primary functions of interest:
#   get_hash_price            -- Gets the current price of a HASH token, e.g. 0.100000000000000000.
#   get_hash_price_for_prompt -- Same as get_hash_price with less digits, e.g. 0.1000, and no ending newline.
#
# Other functions:
#   dlob_cache_get_and_store_daily_price -- Gets and caches the daily price json and hash price value.
#

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

# Set the DLOB cache directory (if not already set). Defaults to /tmp/dlob.
DLOB_C_DIR="${DLOB_C_DIR-/tmp/dlob}"
# Set the DLOB cache max age (if not already set). Defaults to 10m (ten minutes).
DLOB_C_MAX_AGE="${DLOB_C_MAX_AGE-10m}"

# Set the names of the different things to cache.
DLOB_CN_DAILY_PRICE_JSON="DAILY_PRICE_JSON"     # Full json result of daily-price query
DLOB_CN_HASH_PRICE="HASH_PRICE"                 # Just the latest hash price, e.g. 0.105000000000000000

# Usage: dlob_cache_get_and_store_daily_price
# Gets and caches the daily price json and hash price value.
dlob_cache_get_and_store_daily_price () {
    local json
    json="$( curl -s https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price 2> /dev/null)"
    bashcache write "$DLOB_CN_DAILY_PRICE_JSON" -d "$DLOB_C_DIR" <<< "$json"
    jq -r '.latestDisplayPricePerDisplayUnit' <<< "$json" | bashcache write "$DLOB_CN_HASH_PRICE" -d "$DLOB_C_DIR"
    return 0
}

# Usage: get_hash_price
# Outputs the value of a HASH token (in USD), e.g. 0.105000000000000000 with a newline at the end.
# Caching is done using files so that multiple shells will have access to the same data.
# By default, the max cache max age is 10m. This can be altered by setting the DLOB_C_MAX_AGE value in your shell.
#   The format is the same as used for the -atime option of the find command, without the +. E.g. '23h' '6d12h30m'
# If the cache is fresh, the value is printed, and nothing more happens.
# If the cache is stale, the stale value is printed, and a background process is initiated to update the cache.
# If nothing is cached yet, it'll be looked up and cached first, then printed.
# The exit code returned from this function will be the same as the bashcache return code.
get_hash_price () {
    local cache_read_code
    # This will either output the hash price if we have it, or it won't output anything.
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
    10)
        # 10 - The requested cache data is available, but stale.
        # Fire off a background process to update it for next time.
        # The () > /dev/null 2>&1 here is to supress the job/pid start and stop messages.
        ( dlob_cache_get_and_store_daily_price & ) > /dev/null 2>&1
        ;;
    11)
        # 11 - The requested cache data is not available.
        # Get it then try to read it again.
        dlob_cache_get_and_store_daily_price
        bashcache read "$DLOB_CN_HASH_PRICE" -d "$DLOB_C_DIR"
        cache_read_code=$?
        ;;
    *)
        printf 'Unexpected bashcache exit code: [%d]\n' "$cache_read_code" >&2
        ;;
    esac
    return $cache_read_code
}

# Usage: get_hash_price_for_prompt
# It is intended to be used in a command prompt, e.g. PS1='$( get_hash_price_for_prompt ) $'.
# In a dark gray background with bright green font, output a # in a rounded box (emoji \x23\xE2\x83\xA3)
# followed by the result of get_hash_price_for_prompt rounded to 4 decimal places (half-up).
# This is all padded by a space on each side. There is also no newline at the end of this output.
# The exit code returned from this function will be the same as the previous exit code (from before this function is called).
get_hash_price_for_prompt () {
    local previous_exit=$?
    printf '\033[92;100m \x23\xE2\x83\xA3  %1.4f \033[0m' "$( get_hash_price )"
    return $previous_exit
}

# Check that some needed commands are available.
retval=0
for c in 'curl' 'jq' 'bashcache'; do
    if ! command -v "$c" > /dev/null 2>&1; then
        printf 'Missing required command: %s\n' "$c" >&2
        printf 'The functions from %s might not work correctly.\n' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )"
        "$c"
        retval=$?
    fi
done
unset c

# Trick: String expension creates the proper return command, then eval unsets the retval variable before returning as desired.
eval "unset retval; return $retval"
