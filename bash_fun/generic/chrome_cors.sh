#!/bin/bash
# This file contains the chrome_cors function that opens a Chrome browser with CORS security disabled.
# This file can be sourced to add the chrome_cors function to your environment.
# This file can also be executed to run the chrome_cors function without adding it to your environment.
#
# File contents:
#   chrome_cors  --> Opens up a url in Chrome with CORS safety disabled.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Open up a chrome page with CORS safety disabled.
# Usage: chrome_cors <url>
chrome_cors () {
    open -n -a "Google Chrome" "$@" --args --user-data-dir="/tmp/chrome_dev_test" --disable-web-security --new-window
}

if [[ "$sourced" != 'YES' ]]; then
    chrome_cors "$@"
    exit $?
fi
unset sourced

return 0
