#!/bin/bash
# This script will query for blocks and output timing info for them.
# A cache is used for the block info, so only unknown blocks will be re-queried.
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

show_usage () {
    cat << EOF >&2
Get block times for a range of blocks.

Usage: ./get-block-times.sh <height1> [<height2>] [--cache <dir>] [--prov <provenanced>]

If <height2> is provided, <height1> and <height2> define an inclusive height range.

--cache <dir>
    The directory to hold the query block results.
    If the CACHE_DIR env var is set, that is the default.
    Otherwise, if the ./blocks or ./archive/blocks dir exists, that is the default.
    If CACHE_DIR is not set, and neither of those dirs exist, you will either need
    to make one of those directories, or else provide this --cache <dir> option.

--prov <provenanced>
    Use a specific command for the provenanced commands.
    If the PROV env var is set, that is the default.
    Otherwise, if the file ./prov or ./provenanced exists and is executable, that is the default.
    Otherwise, provenanced is the default and must be in your PATH env var.
    The provenanced commmand is not used if the cache already has all needed blocks.

EOF

}

cache="$CACHE_DIR"
prov_cmd="$PROV"

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        --cache)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s.\n' "$1" >&2
                exit 1
            fi
            cache="$2"
            shift
            ;;
        --prov|--prov-cmd)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after %s.\n' "$1" >&2
                exit 1
            fi
            prov_cmd="$2"
            shift
            ;;
        *)
            if [[ -z "$h1" ]]; then
                h1="$1"
            elif [[ -z "$h2" ]]; then
                h2="$1"
            else
                printf 'Unknown argument: [%s]\n' "$1" >&2
                exit 1
            fi
            ;;
    esac
    shift
done

# If no h1 was given, assume no args were provided and output usage info; otherwise, make sure it's only digits.
if [[ -z "$h1" ]]; then
    show_usage
    exit 0
elif [[ "$h1" =~ [^[:digit:]] ]]; then
    printf 'Invalid height 1: [%s]. Can only contain digits.\n' "$h1" >&2
    exit 1
fi
# If no h2 as given, use h1 for it; otherwise, make sure it's only digits.
if [[ -z "$h2" ]]; then
    h2="$h1"
elif [[ "$h2" =~ [^[:digit:]] ]]; then
    printf 'Invalid height 2: [%s]. Can only contain digits.\n' "$h2" >&2
    exit 1
fi
[[ -n "$verbose" ]] && printf 'desired heights: %d - %d\n' "$h1" "$h2"  >&2

# If no cache has been defined, look for the defaults.
if [[ -z "$cache" ]]; then
    if [[ -d './blocks' ]]; then
        cache='./blocks'
    elif [[ -d './archive/blocks' ]]; then
        cache='./archive/blocks'
    fi
fi
# Make sure we've got a dir to use and that it exists as a directory.
if [[ -z "$cache" ]]; then
    printf 'Could not identify cache directory. Please provide it with the --cache <dir> option.\n' >&2
    exit 1;
fi
if [[ ! -d "$cache" ]]; then
    printf 'Cache dir does not exist: %s\n' "$cache" >&2
    exit 1;
fi
[[ -n "$verbose" ]] && printf '      cache dir: %s\n' "$cache" >&2

# Identify the first (lowest) and last (highest) heights to get.
# Get an extra block at the start so we can know how long the first block took.
# Get an extra block at the end so we can include round and vote info for the last block.
if [[ "$h1" -le "$h2" ]]; then
    h_start="$(( h1 - 1 ))"
    h_stop="$(( h2 + 1 ))"
else
    h_start="$(( h2 - 1 ))"
    h_stop="$(( h1 + 1 ))"
fi
[[ -n "$verbose" ]] && printf ' needed heights: %d - %d (%d)\n' "$h_start" "$h_stop" "$(( 1 + h_stop - h_start ))" >&2

get_block_filename () {
    printf '%s/block-%s.json' "$cache" "$1"
}

# First, get a count of how many blocks we will need to query for.
# This way, I can have some special output if we need to query for any blocks, and include the progress.
[[ -n "$verbose" ]] && printf 'Counting the blocks that need to be queried in range: [%s, %s].\n' "$h_start" "$h_stop" >&2
to_query_count=0
h="$h_start"
while [[ "$h" -le "$h_stop" ]]; do
    bf="$( get_block_filename "$h" )"
    if [[ -e "$bf" ]]; then
        if [[ ! -f "$bf" ]]; then
            printf '%d: Block file is not a file: %s\n' "$h" "$bf" >&2
            exit 1
        fi
        [[ -n "$verbose" ]] && printf '%d: Using existing block file: %s\n' "$h" "$bf" >&2
    else
        [[ -n "$verbose" ]] && printf '%d: Query needed. File does not exist: %s\n' "$h" "$bf" >&2
        to_query_count=$(( to_query_count + 1 ))
    fi
    h=$(( h + 1 ))
done

# Do any querying that's needed.
if [[ "$to_query_count" -ne '0' ]]; then
    # If the prov_cmd hasn't yet been determined, pick a default.
    if [[ -z "$prov_cmd" ]]; then
        if [[ -x './prov' ]]; then
            prov_cmd='./prov'
        elif [[ -x './provenanced' ]]; then
            prov_cmd='./provenanced'
        else
            prov_cmd='provenanced'
        fi
    fi
    [[ -n "$verbose" ]] && printf 'provenanced command: %s\n' "$prov_cmd" >&2

    # Make sure the command exists.
    if ! command -v "$prov_cmd" > /dev/null 2>&1; then
        # output a standard command-not-found message
        "$prov_cmd" 1>&2
        exit 1;
    fi

    printf 'Querying for %d blocks.\n' "$to_query_count" >&2
    i=0
    h="$h_start"
    while [[ "$h" -le "$h_stop" ]]; do
        bf="$( get_block_filename "$h" )"
        if [[ ! -e "$bf" ]]; then
            i=$(( i + 1 ))
            if [[ -n "$verbose" || "$to_query_count" -lt '10' || "$i" -eq "$to_query_count" || "$(( i % 10 ))" -eq '1' ]]; then
                printf '[%d/%d]: Querying for block %s and storing result: %s\n' "$i" "$to_query_count" "$h" "$bf" >&2
            fi
            if ! "$prov_cmd" query block --type height "$h" > "$bf"; then
                printf '[%d/%d]: Query failed for block %s.\n' "$i" "$to_query_count" "$h" >&2
                cat "$bf" >&2 2> /dev/null
                rm "$bf" > /dev/null 2>&1
                exit 1
            fi
        fi
        h=$(( h + 1 ))
    done
else
    [[ -n "$verbose" ]] && printf 'Already have all needed blocks.\n' >&2
fi

# First, print out column headers, then the data.
# To get the data, first, print out all the applicable filenames in the right order, and pipe that to an xargs jq command.
#   The jq command will output stuff about the last commit, then a newline, then info about the block itself.
#     Each jq result line will look like this: 2024-08-12T17:21:07.184069629Z 18312791 6 | 18312791 1 61/61
#     Fields: <stamp> <height> <tx count> | <last_commit height> <round> <voted count>/<validator count>
#     The stuff before the | comes from one file, and the stuff after comes from the next.
#     The first line will be the last_commit info from a block we don't care about.
#     And the last line will not have any last_commit info (and be for a block we don't care about).
#   Then, pipe that to an awk command that will calculate the block time and reformat each line a bit.
#     Each awk result line will look like this: 2024-08-12T17:21:07.184069629Z~18312791~2.727s~6~1~61/61
#     Fields: <stamp> <height> <time> <tx count> <round> <voted count><validator count>
#     The first line is last_commit info for a block we don't care about, and is ignored.
#     The second line has a timestamp we need to calculate the first block time. We extract that timestamp, but omit this line too.
#     The last line is for a block we don't care about, but was included so that the last block we care about has last_commit info.
# Then, pipe it all (both header and data) to the column command to make it look nice.
# And Lastly, use sed to right-justify the <time> and <tx count> columns.
#   This is a little tricky, but basically just identifies applicable whitespace and shifts it.
{
    printf 'Stamp~Height~Time~Tx~R~Votes\n'
    { h="$h_start"; while [[ "$h" -le "$h_stop" ]]; do get_block_filename "$h"; printf '\n'; h=$(( h + 1 )); done; } \
        | xargs jq -rj '.last_commit.height + " " +
                        (.last_commit.round|tostring) + " " +
                        ([.last_commit.signatures[]|select(.block_id_flag=="BLOCK_ID_FLAG_COMMIT")]|length|tostring) + "/" +
                        (.last_commit.signatures|length|tostring) +
                        "\n" +
                        .header.time + " " +
                        .header.height + " " +
                        (.data.txs|length|tostring) + " | "' \
        | awk '{
                    if ($1~/^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}T[[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}\.[[:digit:]]+Z$/) {
                        iso=$1;
                        sub(/Z$/,"",iso);
                        date = substr(iso,  1, 10);
                        h = substr(iso, 12, 2) + 0;
                        m = substr(iso, 15, 2) + 0;
                        s = substr(iso, 18, 2) + 0;
                        ms = substr(iso "000", 21, 3) + 0;
                        d="--";
                        if (last_line!="") {
                            cur=s+60*m+3600*h;
                            if (date!=last_date) { cur=cur + 86400; };
                            last=last_s+60*last_m+3600*last_h;
                            sd=cur-last;
                            if (ms >= last_ms) {
                                dmsv=ms-last_ms;
                            } else {
                                sd=sd-1;
                                dmsv=1000+ms-last_ms;
                            };
                            dms=substr(dmsv "000", 1, 3);
                            d=sd "." dms "s";
                        };
                        bl=($2==$5?$2:$2 "!=" $5);
                        r=$6
                        v=$7
                        if (show && $5!="") {
                            print $1 "~" bl "~" d "~" $3 "~" r "~" v;
                        } else {
                            show=1;
                        };
                        last_line=$0;
                        last_date=date;
                        last_h=h;
                        last_m=m;
                        last_s=s;
                        last_ms=ms;
                    };
                }'
} \
    | column -s '~' -t \
    | sed -E 's/^([^ ]+ +[^ ]+  )([^ ]+)( *)  ([^ ]+)( *)  (.*)$/\1\3\2  \5\4  \6/'

