#!/bin/bash
# This file contains the gradle_do function which looks for a gradlew file in either your current location, or the root of the git repo you're in.
# It can be sourced to add the gradle_do function to your environment.
# It can also be executed for the same functionality.
# Aliasing this to gw is pretty handy too.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

gradle_do () {
    if [[ -x './gradlew' ]]; then
        ./gradlew "$@"
        return $?
    elif command -v 'git' > /dev/null 2>&1 && git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
        local git_root
        git_root="$( git rev-parse --show-toplevel )"
        if [[ -x "${git_root}/gradlew" ]]; then
            ( cd "$git_root" && './gradlew' "$@" )
            return $?
        fi
    fi
    printf 'No gradlew file found.\n' >&2
    return 1
}

# If this script was not sourced make it do things now.
if [[ "$sourced" != 'YES' ]]; then
    gradle_do "$@"
    exit $?
fi
unset sourced

return 0
