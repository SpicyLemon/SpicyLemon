#!/bin/bash
# This file contains the check_system_log_timestamp_order function that checks the ordering of entries in the system log.
# This file can be sourced to add the check_system_log_timestamp_order function to your environment.
# This file can also be executed to run the check_system_log_timestamp_order function without adding it to your environment.
#
# File contents:
#   check_system_log_timestamp_order  --> Checks that the lines of a system log file are in chronological order.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: check_system_log_timestamp_order <file>
#   or   get_all_system_logs | check_system_log_timestamp_order -
check_system_log_timestamp_order () {
    local file
    file="$1"
    if [[ -z "$file" ]]; then
        echo "Usage: check_system_log_timestamp_order <file>"
        return 1
    fi
    if [[ "$file" != '-' && ! -f "$file" ]]; then
        echo "File not found: $file"
        return 2
    fi
    cat "$file" \
        | awk 'BEGIN { pt = 0 ; pd = ""; }
            { if (/^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)/)
                { d = $1 " " $2 " " $3;
                    m = (index("JanFebMarAprMayJunJulAugSepOctNovDec",$1)+2)/3;
                    gsub(/(:)/, "", $3);
                    t = sprintf("%d%02d%06d", m, $2, $3);
                    if (pt > t) { print (NR-1) ": " pd " > " d " :" NR; }
                    pt = t; pd = d; } }'
}

if [[ "$sourced" != 'YES' ]]; then
    check_system_log_timestamp_order "$@"
    exit $?
fi
unset sourced

return 0
