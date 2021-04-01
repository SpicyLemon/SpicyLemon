#!/bin/bash
# This file contains the get_all_system_logs function that pulls all the system logs and combines multi-line entries.
# This file can be sourced to add the get_all_system_logs function to your environment.
# This file can also be executed to run the get_all_system_logs function without adding it to your environment.
#
# File contents:
#   get_all_system_logs  --> Gets all the system logs.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: get_all_system_logs
# You'll probably want to pipe this to something or redirect it to a file though.
get_all_system_logs () {
    { cat /var/log/system.log; for l in $( ls /var/log/system.log.* ); do zcat < "$l"; done; } \
    | awk 'BEGIN { al = ""; }
        { if (/^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) /) {
            if (length(al)) { print al; }; al = $0; }
            else { al = al "~" $0; } }
        END { if (length(al)) { print al; } }' \
    | sort -s -k1bM -k2bn -k3.1b,3.2bn -k3.4b,3.5bn -k3.7b,3.8bn \
    | tr '~' '\n'
}

if [[ "$sourced" != 'YES' ]]; then
    get_all_system_logs "$@"
    exit $?
fi
unset sourced

return 0
