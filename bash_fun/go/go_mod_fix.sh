#!/bin/bash
# This file contains the go_mod_fix function that updates go module stuff in a directory structure.
# This file can be sourced to add the go_mod_fix function to your environment.
# This file can also be executed to run the go_mod_fix function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_mod_fix () {
    local d
    d="${1:-.}"
    find "$d" -name 'go.mod' -not -path '*/vendor/*' \
        | while IFS= read -r modfile; do
            (
                printf '%s: ' "$modfile"
                cd "$( dirname "$modfile" )"
                printf 'go mod tidy ... '
                go mod tidy
                printf 'go mod vendor ... '
                go mod vendor
                printf 'go mod verify ... '
                go mod verify
            )
        done
}

if [[ "$sourced" != 'YES' ]]; then
    go_mod_fix "$@"
    exit $?
fi
unset sourced

return 0
