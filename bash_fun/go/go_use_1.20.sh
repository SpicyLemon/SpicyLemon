#!/bin/bash
# This file contains the go_use_1.20.sh function that switches the go bin symlink to version 1.20.
# This file can be sourced to add the go_use_1.20 function to your environment.
# This file can also be executed to run the go_use_1.20 function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_use_1.20 () {
    local verbose which_go desired_link cur_link rv
    if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
        verbose=1
    fi
    [[ "$verbose" ]] && printf 'which go: '
    which_go="$( which go )"
    [[ "$verbose" ]] && printf '%q\n' "$which_go"
    if [[ ! -L "$which_go" ]]; then
        printf 'Not a symlink: ' >&2
        ls -al "$which_go" >&2
        return 1
    fi
    desired_link='/usr/local/go/bin/go'
    [[ "$verbose" ]] && printf 'readlink %q: ' "$which_go"
    cur_link="$( readlink "$which_go" )"
    [[ "$verbose" ]] && printf '%q\n' "$cur_link"
    if [[ "$cur_link" == "$desired_link" ]]; then
        printf 'Already: '
        ls -al "$which_go"
        return 0
    fi
    rv=0
    printf '    Was: '
    ls -al "$which_go"
    ln_flags='-sf'
    [[ "$verbose" ]] && ln_flags="${ln_flags}v"
    [[ "$verbose" ]] && printf 'ln %s %q %q\n' "$ln_flags" "$desired_link" "$which_go"
    ln $ln_flags "$desired_link" "$which_go" || rv=$?
    printf ' Is Now: '
    ls -al "$which_go"
    return $rv
}

if [[ "$sourced" != 'YES' ]]; then
    go_use_1.19 "$@"
    exit $?
fi
unset sourced

return 0
