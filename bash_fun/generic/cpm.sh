#!/bin/bash
# This file contains the cpm function that copies stuff to multiple places.
# This file can be sourced to add the cpm function to your environment.
# This file can also be executed to run the cpm function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

cpm () {
    local flags verbose sources targets entries ec cpec
    flags=()
    while [[ "$#" -gt '0' && "$1" =~ ^-[^-h] ]]; do
        if [[ "$1" == '-v' ]]; then
            verbose="YES"
        fi
        flags+=( "$1" )
        shift
    done
    sources=()
    entries=()
    targets=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                cat << EOF
cpm - Copies things to multiple places (cp multiple).

Usage: cpm [<flags for cp>] [--] source1 [source2 ... --] target1 [target2 ...]

The <flags for cp> must come first are are anything that start with a -.
If the first source file starts with a -, then put a -- before the source files.
If there are multiple sources, put a -- between the sources and targets.

You can also identify sources using --source <source> and targets using --target <target>.
Similarly, the --file <entry>, --dir <entry>, --entry <entry>, and --name <entry> flags all do the same thing:
  add the <entry> to either the sources or targets depending on it's position.

EOF
                return 0
                ;;
            --source)
                if [[ -z "$2" ]]; then
                    printf 'Nothing provided after [%s].\n' "$1" >&2
                    return 1
                fi
                sources+=( "$2" )
                shift
                ;;
            --target)
                if [[ -z "$2" ]]; then
                    printf 'Nothing provided after [%s].\n' "$1" >&2
                    return 1
                fi
                targets+=( "$2" )
                shift
                ;;
            --file|--dir|--entry|--name)
                if [[ -z "$2" ]]; then
                    printf 'Nothing provided after [%s].\n' "$1" >&2
                    return 1
                fi
                entries+=( "$2" )
                shift
                ;;
            --)
                sources+=( "${entries[@]}" )
                entries=()
                ;;
            *)
                if [[ "${#sources[@]}" -eq '0' ]]; then
                    sources+=( "$1" )
                else
                    entries+=( "$1" )
                fi
                ;;
        esac
        shift
    done
    targets+=( "${entries[@]}" )
    if [[ "${#sources[@]}" -eq '0' || "${#targets[@]}" -eq '0' ]]; then
        printf 'Usage: cpm [<flags for cp>] [--] source1 [source2 ... --] target1 [target2 ...]\n' "$short_usage" >&2
        return 64
    fi
    if [[ "${#flags[@]}" -gt '0' ]]; then
        flags+=( '--' )
    fi
    set "${targets[@]}"
    while [[ "$#" -gt '0' ]]; do
        [[ -n "$verbose" ]] && printf 'cp %s %s %q\n' "${flags[*]}" "$( printf ' %q ' "${sources[@]}" )" "$1"
        cp "${flags[@]}" "${sources[@]}" "$1"
        ec="$?"
        if [[ "$ec" -ne '0' ]]; then
            printf 'failed on command: cp %s %s %q\n' "${flags[*]}" "$( printf ' %q ' "${sources[@]}" )" "$1" >&2
            return "$ec"
        fi
        shift
    done
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    cpm "$@"
    exit $?
fi
unset sourced

return 0
