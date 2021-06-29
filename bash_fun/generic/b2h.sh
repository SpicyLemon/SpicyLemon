#!/bin/bash
# This file contains the b2h function that converts byte values to human readable values.
# This file can be sourced to add the b2h function to your environment.
# This file can also be executed to run the b2h function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

b2h () {
    local args i arg vals show_usage use_base_ten stdin_used verbose
    if [[ "$#" -eq '0' ]]; then
        show_usage='YES'
    fi
    # Pre-process args to allow for multiple short-versions together, e.g. -vs is the same as -v -s
    # Also split single args with multiple numbers into multiple args.
    args=()
    while [[ "$#" -gt '0' ]]; do
        if [[ "$1" =~ ^-- ]]; then
            args+=( "$1" )
        elif [[ "$1" == '-' ]]; then
            args+=( '--stdin' )
        elif [[ "$1" =~ ^- ]]; then
            for i in $( seq 1 "$(( ${#1} - 1 ))" ); do
                arg="${1:$i:1}"
                case "$arg" in
                    h) args+=( '--help' );;
                    t) args+=( '--base-ten' );;
                    b|2) args+=( '--base-two' );;
                    v) args+=( '--verbose' );;
                    s) args+=( '--stdin' );;
                    *) args+=( "-$arg" );;
                esac
            done
        else
            # args+=( $1 ) doesn't work the same between bash and zsh.
            # In bash, '2 3' would become two entries in, while in zsh, it would stay as one.
            # We want it split, though, so we gotta get a bit fancier.
            args+=( $( printf '%s' "$1" ) )
        fi
        shift
    done
    # Actually handle the args.
    vals=()
    for arg in "${args[@]}"; do
        case "$arg" in
            --help) show_usage='YES';;
            --base-ten|--base-10|--ten|--10) use_base_ten='YES';;
            --base-two|--base-2|--two|--binary) use_base_ten='';;
            --verbose) verbose='YES';;
            --stdin)
                if [[ -z "$stdin_used" ]]; then
                    vals+=( $( cat - ) )
                    stdin_used='YES'
                fi
                ;;
            -*)
                printf 'Unknown flag: %s\n' "$flag" >&2
                return 2
                ;;
            *)
                vals+=( $arg )
                ;;
        esac
    done
    if [[ -n "$show_usage" ]]; then
        cat << EOF
Converts byte values to human readable values.

Usage: b2h [flags] <value1> [<value2> ...]
   or: <stuff that outputs values> | b2h [flags] [<values>]

    Values:
        Values must be positive numbers.
        Any fractional portions (a period followed by digits) will be ignored.
        Commas are okay.
        One or more spaces will separate numbers (even if provided as the same argument).
        Values have an upper limit based on your system.
            e.g. a 32-bit system has a max of 2,147,483,647 -> 1.99 GiB or 2.14 GB
                 a 64-bit system has a max of 9,223,372,036,854,775,807 -> 7.99 EiB or 9.22 EB
        Values will be processed in the order they are provided.

    Flags:
        --help -h
            Display help (and ignore everything else).
        --base-ten --base-10 --ten --10 -t
            Calculate using 1000 as the divisor instead of 1024.
        --base-two --base-2 --two --binary -2 -b
            Calculate using 1024 as the divisor.
            This is the default behavior.
            If both --base-ten and --base-two (or their aliases) are provided, whichever is last will be used.
        --verbose -v
            Include the original bytes value in the output.
        --stdin -s -
            Get values from stdin.
            If values are also provided as arguments, the ordering depends on where in the arguments this flag is first given.
            I.e. If this flag is before any provided values, the piped in values will be processed first.
                 If this flag is after all provided values, the piped in values will be processed last.
                 If this flag is between provided values, the values before it will be processed,
                    then the piped in values, followed by the values provided after.
            E.g. All of these commands will process the numbers 1, 2, 3, 4, and 5 in the same order.
                printf '1 2' | b2h - 3 4 5
                printf '4 5' | b2h 1 2 3 --stdin
                printf '2 3 4' | b2h 1 --pipe 5
                printf '2 3' | b2h 1 - 4 -s 5
                    (the second instance of the flag is just ignored)

EOF
        return 0
    fi
    # Add stdin values if requested.
    if [[ -n "$use_stdin" ]]; then
        #vals+=( $( printf '%s' "$( cat - )" ) )
        vals+=( $( cat - ) )
    fi

    local U d o mu e v w f u nw
    # Define the units (in order).
    # Max 64-bit signed int (as bytes) is in EiB, but who knows...
    # Having zetta and yotta here doesn't really hurt anything anyway.
    U=( "Bytes" "KiB" "MiB" "GiB" "TiB" "PiB" "EiB" "ZiB" "YiB" )
    # Set the divisor.
    d=1024
    if [[ -n "$use_base_ten" ]]; then
        U=( "Bytes" "KB" "MB" "GB" "TB" "PB" "EB" "ZB" "YB" )
        d=1000
    fi
    # Some shells have zero based arrays, some have 1 based arrays.
    # Set o (the offset) to the lowest array index.
    [[ -n "${U[0]}" ]] && o=0 || o=1
    # pre-calc the max U index.
    mu=$(( ${#U[@]} - 1 + o ))
    # Set the initial exit code to return.
    e=0
    for v in "${vals[@]}"; do
        # Get rid of any fractional parts of numbers.
        # Get rid of commas from inside numbers.
        # w is then initially the number of bytes, and will end up being the whole number portion in whatever units we end up in.
        w="$( sed -E 's/([[:digit:]])\.[[:digit:]]+/\1/g; s/([[:digit:]]),([[:digit:]])/\1\2/g;' <<< "$v" )"
        if [[ ! "$w" =~ ^[[:digit:]]+$ ]]; then
            # w is either empty or contains something other than a digit.
            [[ -n "$verbose" ]] && printf "[%s] Bytes = " "$v" >&2
            printf 'Invalid number: %s\n' "$w" >&2
            e=1
        elif [[ "$( { printf '%s' "$(( w - 1 + 1 ))"; } 2> /dev/null )" != "$w" ]]; then
            # Overflow detected, the number's too big.
            [[ -n "$verbose" ]] && printf "[%s] Bytes = " "$v" >&2
            printf 'Value too large: %s\n' "$w" >&2
            e=1
        else
            # Initialize the fractional part. Will also contain the decimal (separator).
            f=''
            # Initialize the index of the units array.
            u=$o
            # Specifically using 1000 here instead of d in this test so that the whole number portion is at most 3 digits.
            while (( w >= 1000 && u < mu )); do
                f="$( printf '.%02d' "$((w % d * 100 / d))" 2> /dev/null )"
                w="$(( w / d ))"
                u=$(( u + 1 ))
            done
            [[ -n "$verbose" ]] && printf "[%s] Bytes = " "$v"
            printf '%d%s %s\n' "$w" "$f" "${U[$u]}"
        fi
    done
    return $e
}

if [[ "$sourced" != 'YES' ]]; then
    b2h "$@"
    exit $?
fi
unset sourced

return 0
