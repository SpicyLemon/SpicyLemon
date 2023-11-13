#!/usr/bin/env bash
# This script will either take in or look up the current block height, then estimate when it will reach the requested height.
# This script is meant to be executed, not sourced.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" == 'YES' ]]; then
    unset sourced
    printf 'This file is to be executed, not sourced.\n'
    return 1
fi

# Just the name of this file with ./ in front of it.
fn="./$( basename "$0" )"

# The number of blocks to go back when estimating milliseconds per block
blocks_back='10000'

req_date_format='YYYY-MM-DD{T| }hh:mm[:ss[.ssssss]][Z]'

usage="$( cat << EOF
$fn - Estimates chain height for a desired date and time, or the date and time the chain will reach a certain height.

Usage: $fn [<options>] <desired date time>
   or: $fn [<options>] <desired height>

    Exactly one of <desired date time> and <desired height> must be provided.

    <desired date time> is the date and time you want an estimate of what the height will be.
        It must have the format: $req_date_format
        Where YYYY = 4 digit year, MM = 2 digit month, and DD = 2 digit day of month.
        {T| } is either the letter T or a space (note that if a space is used, the argument must be quoted).
        Then hh = 2 digit hour in 24-hour format, mm = minutes, ss = seconds, and .ssssss is any number of fractional seconds.
        And [Z] is a timezone offset as either the letter Z for UTC (Zulu) or {+|-}hhmm (e.g. -0600).
        The seconds are optional and default to 0.
        The fractional seconds are optional and default to 0. Seconds must be provided in order to provide fractional seconds.
        The time zone is optional and defaults to the time zone of your system.
        Examples: '2021-09-15T12:56', '2021-09-15T12:56:00', '2021-09-15T12:56:00-0600', '2021-09-15T12:56:00.000000-0600',
                  '2021-09-15 12:56', '2021-09-15 12:56:00', '2021-09-15 12:56:00-0600', '2021-09-15 12:56:00.000000-0600'

    <desired height> is the height you want an estimate of what the date and time will be.

    The   provenanced query block   command is used to get block and height information.
    To use a specific provenanced binary, set and export the PROVD environment variable as the path to the binary.
        e.g. export PROVD="\$HOME/git/provenance/build/provenanced"
    Flags can be provided to the command using exported environment variables as usual.
        e.g. export PIO_HOME="\$HOME/git/provenance/build/run/provenanced"

    Options:
        Ordering of the options and desired date time or height does not matter.
        If an option is provided more than once, the last instance is used.

    --current-height <height> is the height to use as the current height.
        The <height> must be only digits, e.g '10555'.
        If not provided, the current height and date time will be looked up using provenanced.
        If provided, that lookup will not happen. the provided height will be used as the current height,
        and the system date time is used as the current date time (overwritable with --current-date-time).

    --current-date-time <date time> is the date and time to use as the current date time.
        It must have the format: $req_date_format
        See <desired date time> for details.

    --ms-per-block <milliseconds> is the number of milliseconds it takes to create an average block.
        The <milliseconds> must be digits only, e.g. '5000' (for 5 seconds).
        If not supplied, a previous block will be looked up using provenanced,
        and the average from there to the current block is used.

    --blocks-back <number> sets the number of blocks back to go to estimate ms-per-block.
        If both --blocks-back and --from-height are provided, the last one is used.

    --from-height <height> is the past height to look up in order to estimate ms-per-block.
        If both --blocks-back and --from-height are provided, the last one is used.

    The default --blocks-back or --from-height depends on the type of input.
        If a date/time was provided, then the default is --blocks-back $blocks_back.
        If a height was provided, then the default is --blocks-back (<desired height> - <current height>), minimum 100.

EOF
)"

# The date commmand is used in here a bunch, but there's two versions, Linux/GNU and Unix/Mac.
# The Unix/Mac version doesn't like the --help flag and treats it like an illegal option, exiting with code 1.
# But the Linux/GNU version outputs a bunch of stuff and exits with code 0.
if date --help > /dev/null 2>&1; then
    use_gnu_date='YES'
fi

# Usage: to_epoch_ms <date time and zone>
# <date time and zone> should be in the format YYYY-MM-DD{T| }hh:mm[:ss[.ssssss]][Z]
# T is the letter T.
# Z can either be the letter Z or an offset, e.g. '-0600'.
to_epoch_ms () {
    local pieces the_date the_time the_zone s_fractions ms epoch_s epoch_ms
    # Swap out the T for a space, put a space before the zone offset, and change an ending Z into an offset of +0000.
    pieces=( $( sed -E 's/T/ /; s/([-+][[:digit:]]+)$/ \1/; s/Z$/ +0000/;' <<< "$*" ) )
    the_date="${pieces[0]}"
    the_time="${pieces[1]}"
    the_zone="${pieces[2]}"
    if [[ "$the_time" =~ ^([[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2})[.,]([[:digit:]]+)$ ]]; then
        the_time="${BASH_REMATCH[1]}"
        s_fractions="$( sed -E 's/0+$//' <<< "${BASH_REMATCH[2]}" )"
    elif [[ "$the_time" =~ ^[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        the_time="${the_time}:00"
    fi
    ms="$( head -c 3 <<< "${s_fractions}000" )"
    if [[ -z "$the_zone" ]]; then
        # This one is the same for both date commands.
        the_zone="$( date '+%z' )"
    else
        the_zone="$( tr -d ':' <<< "${the_zone}0000" | head -c 5 )"
    fi
    if [[ -n "$use_gnu_date" ]]; then
        epoch_s="$( date -d "$the_date $the_time $the_zone" '+%s' )" || return $?
    else
        epoch_s="$( date -j -f '%F %T %z' "$the_date $the_time $the_zone" '+%s' )" || return $?
    fi
    epoch_ms="$( sed -E 's/^0+//' <<< "${epoch_s}${ms}" )"
    if [[ -z "$epoch_ms" ]]; then
        epoch_ms='0'
    fi
    printf '%s' "$epoch_ms"
    return 0
}

# Usage: epoch_ms_to_date_time <milliseconds>
# Converts epoch as milliseconds into a date.
epoch_ms_to_date_time () {
    local ms s rv
    ms="$1"
    s="$(( ms / 1000 ))" || return $?
    if [[ -n "$use_gnu_date" ]]; then
        rv="$( date -d "@$s" +'%F %T %z' )" || return $?
    else
        rv="$( date -j -f '%s' "$s" +'%F %T %z' )" || return $?
    fi
    printf '%s' "$rv"
    return 0
}

ensure_provd () {
    if [[ -z "$PROVD" ]]; then
        if ! command -v provenanced > /dev/null 2>&1; then
            printf 'Command not found: provenanced\n' >&2
            return 1
        fi
        PROVD='provenanced'
    fi
    return 0
}


################
# Parse args.

if [[ "$#" -eq '0' ]]; then
    printf '%s\n' "$usage"
    exit 0
fi

while [[ "$#" -gt '0' ]]; do
    case "$( tr '[:upper:]' '[:lower:]' <<< "$1" )" in
        -h|--help)
            printf '%s\n' "$usage"
            exit 0
            ;;
        -v|-vv|--verbose)
            verbose="$1"
            ;;
        --current-height|--current-block)
            if [[ -z "$2" ]]; then
                printf 'No <height> provided after %s flag.\n' "$1" >&2
                exit 1
            fi
            arg_current_height="$2"
            shift
            ;;
        --current-date-time|--current-date|--current-time)
            if [[ -z "$2" ]]; then
                printf 'No <date time> provided after %s flag.\n' "$1" >&2
                exit 1
            fi
            arg_current_time="$2"
            shift
            ;;
        --ms-per-block|--milliseconds-per-block)
            if [[ -z "$2" ]]; then
                printf 'No <milliseconds> provided after %s flag.\n' "$1" >&2
                exit 1
            fi
            arg_ms_per_block="$2"
            shift
            ;;
        --blocks_back|--blocks-back)
            if [[ -z "$2" ]]; then
                printf 'No <number> provided after %s flag.\n' "$1" >&2
                exit 1
            fi
            if [[ ! "$2" =~ ^[[:digit:]]+$ ]]; then
                printf 'Invalid <number> provided after %s flag: [%s].\n' "$1" "$2" >&2
                exit 1
            fi
            if [[ "$2" -lt '1' ]]; then
                printf 'The <number> provided after %s flag must be at least 1, have: [%s].\n' "$1" "$2" >&2
                exit 1
            fi
            arg_blocks_back="$2"
            arg_from_height=''
            shift
            ;;
        --from-height|--from_height|--from-block|--from_block)
            if [[ -z "$2" ]]; then
                printf 'No <height> provided after %s flag.\n' "$1" >&2
                exit 1
            fi
            if [[ ! "$2" =~ ^[[:digit:]]+$ ]]; then
                printf 'Invalid <height> provided after %s flag: [%s].\n' "$1" "$2" >&2
                exit 1
            fi
            if [[ "$2" -lt '1' ]]; then
                printf 'The <height> provided after %s flag must be at least 1, have: [%s].\n' "$1" "$2" >&2
                exit 1
            fi
            arg_from_height="$2"
            args_blocks_back=''
            shift
            ;;
        *)
            if [[ -n "$arg_desired" ]]; then
                printf 'Unknown argument: [%s].\n' "$1" >&2
                exit 1
            fi
            if [[ -n "$2" && "$2" =~ ^[[:digit:]][[:digit:]]: ]]; then
                arg_desired="$1 $2"
                shift
            else
                arg_desired="$1"
            fi
            ;;
    esac
    shift
done

if [[ -n "$verbose" ]]; then
    if [[ -n "$use_gnu_date" ]]; then
        printf 'Using Linux/GNU version of date commands.\n' >&2
    else
        printf 'Using Unix/Mac version of date commands.\n' >&2
    fi
fi

###################
# Validate args.

if [[ -z "$arg_desired" ]]; then
    printf 'No <desired date time> or <desired height> provided.\n' >&2
    exit 1
elif [[ "$arg_desired" =~ ^[[:digit:]]+$ ]]; then
    desired_height="$arg_desired"
    [[ -n "$verbose" ]] && printf 'Desired height provided: [%s].\n' "$desired_height" >&2
else
    desired_ms="$( to_epoch_ms "$arg_desired" )"
    ec=$?
    if [[ "$ec" -ne '0' ]]; then
        printf 'Invalid <desired date time>: [%s]. Must be %s\n' "$arg_desired" "$req_date_format" >&2
        exit $ec
    fi
    desired_time_disp="$( epoch_ms_to_date_time "$desired_ms" )" || exit $?
    if [[ -n "$verbose" ]]; then
        printf 'Desired date time provided: [%s].\n' "$arg_desired" >&2
        printf 'Desired date time epoch ms: [%s].\n' "$desired_ms" >&2
    fi
fi

if [[ -n "$arg_current_height" ]]; then
    if [[ "$arg_current_height" =~ [^[:digit:]] ]]; then
        printf 'Invalid current <height>: [%s]. Must only contain digits.\n' "$arg_current_height" >&2
        exit 1
    fi
    current_height="$arg_current_height"
    [[ -n "$verbose" ]] && printf 'Current height provided: [%s].\n' "$current_height" >&2
fi

if [[ -n "$arg_current_time" ]]; then
    current_ms="$( to_epoch_ms "$arg_current_time" )"
    ec=$?
    if [[ "$ec" -ne '0' ]]; then
        printf 'Invalid current <date time>: [%s]. Must be %s\n' "$arg_current_time" "$req_date_format" >&2
        exit $ec
    fi
    current_time="$arg_current_time"
    if [[ -n "$verbose" ]]; then
        printf 'Current date time provided: [%s].\n' "$current_time" >&2
        printf 'Current date time epoch ms: [%s].\n' "$current_ms" >&2
    fi
fi

if [[ -n "$arg_ms_per_block" ]]; then
    if [[ "$arg_ms_per_block" =~ [^[:digit:]] ]]; then
        printf 'Invalid <milliseconds>: [%s]. Must only contain digits.\n' "$arg_ms_per_block" >&2
        exit 1
    fi
    ms_per_block="$arg_ms_per_block"
    [[ -n "$verbose" ]] && printf 'Milliseconds per block provided: [%s].\n' "$ms_per_block" >&2
fi

if [[ -n "$arg_blocks_back" ]]; then
    blocks_back="$arg_blocks_back"
    [[ -n "$verbose" ]] && printf 'Blocks back provided: [%s].\n' "$arg_blocks_back" >&2
elif [[ -n "$arg_from_height" ]]; then
    old_height="$arg_from_height"
    [[ -n "$verbose" ]] && printf 'From height provided: [%s].\n' "$arg_from_height" >&2
else
    use_default_blocks_back='YES'
    [[ -n "$verbose" ]] && printf 'Using default blocks back: [%s].\n' "$blocks_back" >&2
fi


#############################################
# Lookup information not provided by args.

if [[ -z "$current_height" ]]; then
    ensure_provd || exit $?
    [[ -n "$verbose" ]] && printf 'Executing command: %s query block  ... ' "$PROVD" >&2
    current_block="$( "$PROVD" query block )" || exit $?
    [[ -n "$verbose" ]] && printf 'Done\n' >&2
    provd_used='YES'
    [[ "$verbose" =~ vv ]] && jq '.' <<< "$current_block" >&2
    current_height="$( jq -r '.block.header.height' <<< "$current_block" )"
    [[ -n "$verbose" ]] && printf 'Current height looked up: [%s].\n' "$current_height" >&2
fi

if [[ -z "$current_time" ]]; then
    if [[ -n "$current_block" ]]; then
        current_time="$( jq -r '.block.header.time' <<< "$current_block" )"
        [[ -n "$verbose" ]] && printf 'Current date time from block: [%s].\n' "$current_time" >&2
    elif [[ -n "$use_gnu_date" ]]; then
        current_time="$( date '+%FT%T.%N%z' )"
        [[ -n "$verbose" ]] && printf 'Current date time from system: [%s].\n' "$current_time" >&2
    else
        current_time="$( date +'%FT%T%z' )"
        [[ -n "$verbose" ]] && printf 'Current date time from system: [%s].\n' "$current_time" >&2
    fi
fi

if [[ -z "$current_ms" ]]; then
    current_ms="$( to_epoch_ms "$current_time" )" || exit $?
    [[ -n "$verbose" ]] && printf 'Current date time epoch ms: [%s].\n' "$current_ms" >&2
fi
current_time_disp="$( epoch_ms_to_date_time "$current_ms" )" || exit $?

if [[ -z "$ms_per_block" ]]; then
    ensure_provd || exit $?
    if [[ -z "$old_height" ]]; then
        if [[ -n "$desired_height" && -n "$use_default_blocks_back" && "$desired_height" -gt "$current_height" ]]; then
            blocks_back="$(( desired_height - current_height ))"
            if [[ "$blocks_back" -lt "100" ]]; then
                blocks_back='100'
            fi
            [[ -n "$verbose" ]] && printf 'Changing blocks back to [%s].\n' "$blocks_back" >&2
        fi
        old_height="$(( current_height - blocks_back ))" || exit $?
    fi
    if [[ "$old_height" -lt '1' ]]; then
        old_height=0
    fi
    [[ -n "$verbose" ]] && printf 'Executing command: %s query block %s  ... ' "$PROVD" "$old_height" >&2
    old_block="$( "$PROVD" query block "$old_height" )" || exit $?
    [[ -n "$verbose" ]] && printf 'Done\n' >&2
    provd_used='YES'
    [[ "$verbose" =~ vv ]] && jq '.' <<< "$old_block" >&2
    old_time="$( jq -r '.block.header.time' <<< "$old_block" )"
    [[ -n "$verbose" ]] && printf 'Old block date time: [%s].\n' "$old_time" >&2
    old_ms="$( to_epoch_ms "$old_time" )" || exit $?
    [[ -n "$verbose" ]] && printf 'Old block date time epoch ms: [%s].\n' "$old_ms" >&2
    ms_per_block="$(( ( current_ms - old_ms ) / ( current_height - old_height ) ))" || exit $?
    [[ -n "$verbose" ]] && printf 'Milliseconds per block calculated: [%s].\n' "$ms_per_block" >&2
fi

if [[ -n "$provd_used" ]]; then
    [[ -n "$verbose" ]] && printf 'Executing command: %s config get all ... ' "$PROVD" >&2
    config_all="$( "$PROVD" config get all 2>&1 )"
    ec=$?
    if [[ "$ec" -eq '0' ]]; then
        [[ -n "$verbose" ]] && printf 'Done\n' >&2
        moniker="$( grep '^moniker=' <<< "$config_all" | sed 's/^moniker="//; s/"$//;' )"
        chain_id="$( grep '^chain-id=' <<< "$config_all" | sed 's/^chain-id="//; s/"$//;' )"
    else
        [[ -n "$verbose" ]] && printf 'Failed with code %s.\n%s' "$ec" "$config_all" >&2
    fi
    if [[ -n "$PIO_HOME" ]]; then
        pio_home="$PIO_HOME"
    fi
fi


################################
# Do the desired calculation.

if [[ -z "$desired_height" ]]; then
    if [[ -n "$verbose" ]]; then
        printf 'Calculating desired height.\n' >&2
        printf 'Givens:\n' >&2
        printf '  Desired epoch milliseconds: [%s].\n' "$desired_ms" >&2
        printf '  Current epoch milliseconds: [%s].\n' "$current_ms" >&2
        printf '  Milliseconds per block: [%s].\n' "$ms_per_block" >&2
        printf '  Current height: [%s].\n' "$current_height" >&2
        printf 'Results:\n' >&2
    fi
    ms_diff="$(( desired_ms - current_ms ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Elapsed milliseconds: [%s].\n' "$ms_diff" >&2
    if [[ "$ms_diff" -lt '0' ]]; then
        printf 'Desired date time [%s] must be after the current date time [%s].\n' "$desired_time_disp" "$current_time_disp" >&2
        exit 1
    fi
    block_diff="$(( ms_diff / ms_per_block ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Elapsed blocks: [%s].\n' "$block_diff" >&2
    desired_height="$(( current_height + block_diff ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Desired height: [%s].\n' "$desired_height" >&2
elif [[ -z "$desired_ms" ]]; then
    if [[ -n "$verbose" ]]; then
        printf 'Calculating desired date time.\n' >&2
        printf 'Givens:\n' >&2
        printf '  Desired height: [%s].\n' "$desired_height" >&2
        printf '  Current height: [%s].\n' "$current_height" >&2
        printf '  Milliseconds per block: [%s].\n' "$ms_per_block" >&2
        printf '  Current epoch milliseconds: [%s].\n' "$current_ms" >&2
        printf 'Results:\n' >&2
    fi
    if [[ "$desired_height" -le "$current_height" ]]; then
        printf 'Desired height [%s] has already occurred. Currently at height [%s].\n' "$desired_height" "$current_height" >&2
        exit 1
    fi
    block_diff="$(( desired_height - current_height ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Elapsed blocks: [%s].\n' "$block_diff" >&2
    ms_diff="$(( block_diff * ms_per_block ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Elapsed milliseconds: [%s].\n' "$ms_diff" >&2
    desired_ms="$(( current_ms + ms_diff ))" || exit $?
    [[ -n "$verbose" ]] && printf '  Desired epoch milliseconds: [%s].\n' "$desired_ms" >&2
    desired_time_disp="$( epoch_ms_to_date_time "$desired_ms" )" || exit $?
    [[ -n "$verbose" ]] && printf '  Desired date time: [%s].\n' "$desired_time_disp" >&2
else
    printf 'Bug: Nothing to calculate. Should not have gotten here.\n' >&2
    exit 2
fi


#########################
# Do the output dance!

[[ -n "$pio_home" ]] && printf 'PIO Home: %s\n' "$pio_home"
[[ -n "$chain_id" ]] && printf 'Chain-Id: %s\n' "$chain_id"
[[ -n "$moniker" ]] && printf 'Moniker: %s\n' "$moniker"
printf 'Current: %s Height: %s\n' "$current_time_disp" "$current_height"
printf 'Desired: %s Height: %s\n' "$desired_time_disp" "$desired_height"
printf 'Elapsed milliseconds: %s = %s blocks (at %s milliseconds per block from last %s blocks).\n' "$ms_diff" "$block_diff" "$ms_per_block" "$blocks_back"


exit 0
