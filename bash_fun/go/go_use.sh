#!/bin/bash
# This file contains the go_use.sh function that switches the go bin symlink to a provided version.
# This file can be sourced to add the go_use function to your environment.
# This file can also be executed to run the go_use function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

go_use () {
    local v118 v119 v120 v121
    v118='../Cellar/go@1.18/1.18.10/bin/go'
    v119='../Cellar/go/1.19.6/bin/go'
    v120='/usr/local/go/bin/go' # 1.20.1
    v121='../Cellar/go/1.21.4/bin/go'

    local verbose listing which_go desired_link cur_link rv
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf 'Usage: go_use {1.18|1.19|1.20|1.21|list} [-v|--verbose]\n'
                return 0
                ;;
            -v|--verbose)
                verbose=1
                ;;
            1.18|v1.18)
                desired_link="$v118"
                ;;
            1.19|v1.19)
                desired_link="$v119"
                ;;
            1.20|v1.20)
                desired_link="$v120"
                ;;
            1.21|v1.21)
                desired_link="$v121"
                ;;
            -l|--list|l|list)
                listing=1
                ;;
            *)
                printf 'Unknown argument: %q\n' "$1"
                return 1
                ;;
        esac
        shift
    done

    if [[ -z "$desired_link" ]]; then
        listing=1
    fi

    [[ "$verbose" ]] && printf 'which go: '
    which_go="$( which go )"
    [[ "$verbose" ]] && printf '%q\n' "$which_go"
    if [[ ! -L "$which_go" ]]; then
        printf 'Not a symlink: ' >&2
        ls -al "$which_go" >&2
        return 1
    fi

    [[ "$verbose" ]] && printf 'readlink %q: ' "$which_go"
    cur_link="$( readlink "$which_go" )"
    [[ "$verbose" ]] && printf '%q\n' "$cur_link"

    if [[ "$listing" ]]; then
        printf 'available versions:\n'
        printf '  %s: %s\n' '1.18' "$v118" '1.19' "$v119" '1.20' "$v120" '1.21' "$v121"
        printf 'Current: %s\n' "$cur_link"
        return 0
    fi

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
    go_use "$@"
    exit $?
fi
unset sourced

return 0
