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
    local n118 v118 n120 v120 n121 v121 n123 v123
    n118='1.18'
    v118='../Cellar/go@1.18/1.18.10/bin/go'
    n120='1.20'
    v120='/usr/local/go/bin/go' # 1.20.1
    n121='1.21'
    v121='../Cellar/go/1.21.4/bin/go'
    n123='1.23'
    v123='../Cellar/go/1.23.0/bin/go'

    local verbose listing which_go desired_link cur_link rv
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf 'Usage: go_use {%s|list} [-v|--verbose]\n' "$n118|$n120|$n121|$n123"
                return 0
                ;;
            -v|--verbose)
                verbose=1
                ;;
            "$n118"|"v$n118")
                desired_link="$v118"
                ;;
            "$n120"|"v$n120")
                desired_link="$v120"
                ;;
            "$n121"|"v$n121")
                desired_link="$v121"
                ;;
            "$n123"|"v$n123")
                desired_link="$v123"
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
        local opts
        opts=()
        opts+=( "$( n="$n118"; v="$v118"; if [[ "$cur_link" == "$v" ]]; then printf '  \033[1m%s\033[0m: %s  \033[1m(current)\033[0m\n' "$n" "$v"; else printf '  %s: %s' "$n" "$v"; fi )" )
        opts+=( "$( n="$n120"; v="$v120"; if [[ "$cur_link" == "$v" ]]; then printf '  \033[1m%s\033[0m: %s  \033[1m(current)\033[0m\n' "$n" "$v"; else printf '  %s: %s' "$n" "$v"; fi )" )
        opts+=( "$( n="$n121"; v="$v121"; if [[ "$cur_link" == "$v" ]]; then printf '  \033[1m%s\033[0m: %s  \033[1m(current)\033[0m\n' "$n" "$v"; else printf '  %s: %s' "$n" "$v"; fi )" )
        opts+=( "$( n="$n123"; v="$v123"; if [[ "$cur_link" == "$v" ]]; then printf '  \033[1m%s\033[0m: %s  \033[1m(current)\033[0m\n' "$n" "$v"; else printf '  %s: %s' "$n" "$v"; fi )" )

        printf 'available versions:\n'
        printf '%b\n' "${opts[@]}"
        if ! grep -qF '(current)' <<< "${opts[*]}" > /dev/null 2>&1; then
            printf '\033[1mCurrent\033[0m: %s\n' "$cur_link"
        fi
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
