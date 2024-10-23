#!/bin/bash
# This file contains the set_ops function that applies a set operation to the lines of two files.
# This file can be sourced to add the set_ops function to your environment.
# This file can also be executed to run the set_ops function without adding it to your environment.
#
# File contents:
#   set_ops  --> Apply set operations to the lines of two files.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

set_ops () {
    local usage file1 file2 op swap tmp ec verbose
    usage="$( cat << EOF
Usage: set_ops <file1> <op> <file2> [--swap]

The <op> can be one of the following:
    union  u  +  or
    intersection  n  and  intersect
    compliment  -  c  not  subtract  sub  minus
    symetric-difference  sym-diff  symdiff  simdiff  sim-diff  diff  s  xor

The --swap flag will cause <file1> and <file2> to trade places for the operation.
    This is primarily for compliment, where the file order matters.
    These two commands are equivalent:
      $ set_ops file1 - file2 --swap
      $ set_ops file2 - file1

EOF
)"
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --help|-h)
                printf '%s\n' "$usage"
                return 0
                ;;
            --swap|-x|-s)
                swap="$1"
                ;;
            --verbose|-v)
                verbose="$1"
                ;;
            *)
                if [[ -z "$file1" ]]; then
                    file1="$1"
                elif [[ -z "$op" ]]; then
                    op="$1"
                elif [[ -z "$file2" ]]; then
                    file2="$1"
                else
                    printf 'Unknown arg: [%s].\n' "$1"
                    return 1
                fi
                ;;
        esac
        shift
    done

    if [[ -n "$verbose" ]]; then
        printf 'file1: %q\n' "$file1"
        printf '   op: %q\n' "$op"
        printf 'file2: %q\n' "$file2"
    fi

    if [[ -z "$file1" || -z "$op" || -z "$file2" ]]; then
        printf '%s\n' "$usage"
        return 0
    fi

    # Allow for the op to be in any arg position.
    if [[ -f "$op" ]]; then
        if [[ ! -f "$file1" ]]; then
            [[ -n "$verbose" ]] && printf 'Swapping: file1 <-> op : [%s] <-> [%s]\n' "$file1" "$op"
            tmp="$file1"
            file1="$op"
            op="$tmp"
        elif [[ ! -f "$file2" ]]; then
            [[ -n "$verbose" ]] && printf 'Swapping: op <-> file2 : [%s] <-> [%s]\n' "$op" "$file2"
            tmp="$file2"
            file2="$op"
            op="$tmp"
        fi
    fi

    if [[ -n "$swap" ]]; then
        [[ -n "$verbose" ]] && printf 'Swapping: file1 <-> file2 : [%s] <-> [%s]\n' "$file1" "$file2"
        tmp="$file1"
        file1="$file2"
        file2="$tmp"
    fi

    ec=0
    if [[ ! -f "$file1" ]]; then
        printf 'File not found: %s\n' "$file1"
        ec=1
    fi
    if [[ ! -f "$file2" ]]; then
        printf 'File not found: %s\n' "$file2"
        ec=1
    fi
    if [[ "$ec" -ne '0' ]]; then
        return "$ec"
    fi

    local cat_cmd uniq_cmd
    cat_cmd=( cat )
    uniq_cmd=( uniq )
    case "$op" in
        union|u|+|or)
            # cat A B | sort | uniq
            cat_cmd+=( "$file1" "$file2" )
            ;;
        intersection|n|and|intersect)
            # cat A B | sort | uniq -d
            cat_cmd+=( "$file1" "$file2" )
            uniq_cmd+=( -d )
            ;;
        compliment|-|c|not|subtract|sub|minus)
            # cat A B B | sort | uniq -u
            cat_cmd+=( "$file1" "$file2" "$file2" )
            uniq_cmd+=( -u )
            ;;
        symetric-difference|sym-diff|symdiff|simdiff|sim-diff|diff|s|xor)
            # cat A B | sort | uniq -u
            cat_cmd+=( "$file1" "$file2" )
            uniq_cmd+=( -u )
            ;;
    esac

    [[ -n "$verbose" ]] && printf '%s | sort | %s\n' "${cat_cmd[*]}" "${uniq_cmd[*]}"
    "${cat_cmd[@]}" | sort | "${uniq_cmd[@]}"
}

if [[ "$sourced" != 'YES' ]]; then
    set_ops "$@"
    exit $?
fi
unset sourced

return 0
