#!/bin/bash
# This file contains functions for helping do things with files.
# File contents:
#   flatten_file  ----------------------> Comma separates a file and removes line breaks.
#   flatten_quote_file  ----------------> Single-quotes each line, comma separates them and removes line breaks.
#   make_nice_files  -------------------> Does a flatten_file, flatten_quote_file and also creates files with 15 entries per line (using split_x_per_line).
#   multi_line_replace  ----------------> Replaces part of a file with multi-line replacement text.
#   get_all_system_logs  ---------------> Gets all the system logs.
#   check_system_log_timestamp_order  --> Checks that the lines of a system log file are in chronological order.
#
# Depends on:
#   add_to_filename - Function defined in generic/add_to_filename.sh
#   string_repeat - Function defined in text-helpers.sh
#   split_x_per_line - Function defined in text-helpers.sh

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

# Adds a comma space to the end of each line and gets rid of the line breaks
# Usage: flatten_file file.txt
flatten_file () {
    local filename filename_out
    filename="$1"
    if [[ -z "$filename" ]]; then
        echo "Usage: flatten_file <filename>"
    elif [[ -f "$filename" ]]; then
        filename_out="$( add_to_filename 'flat' "$filename" )"
        cat "$filename" | sed -E 's/ *$/, /' | tr -d '\n' | sed -E 's/, $//' > $filename_out
        echo "Created: $filename_out"
    else
        >&2 echo "File not found: $filename"
    fi
}

# Wraps each line in single quotes and adds a comma space and gets rid of line breaks
# Usage: flatten_quote_file file.txt
flatten_quote_file () {
    local filename filename_out
    filename="$1"
    if [[ -z "$filename" ]]; then
        echo "Usage: flatten_file <filename>"
    elif [[ -f "$filename" ]]; then
        filename_out="$( add_to_filename 'flat_quoted' "$filename" )"
        cat "$filename" | sed -E "s/^/'/; s/ *$/', /" | tr -d '\n' | sed -E 's/, $//' > $filename_out
        echo "Created: $filename_out"
    else
        >&2 echo "File not found: $filename"
    fi
}

# Makes up to 4 files based on another file that is assumed to have one entry per line.
#   File 1: The results of flatten_file
#   File 2: The results of flatten_quote_file (as long as the -q or --no_quoted option is not given).
#   File 3: Similar to flatten_file but with <count> entries per line (default <count> is 15 if not provided).
#   File 3: Similar to flatten_quote_file but with <count> entries per line (as long as -q or --no_quoted isn't given).
make_nice_files () {
    local usage option no_quoted count_in filename
    usage="Usage: make_nice_files [-q|--no-quoted] [-n <count>|--count <count>] <filename>"
    while [[ "$#" -gt "0" ]]; do
        option="$( printf %s "$1" | tr "[:upper:]" "[:lower:]" )"
        case "$option" in
        -h|--help)
            echo "$usage"
            return 0
            ;;
        -q|--no-quoted)
            no_quoted="YES"
            ;;
        -n|--count)
            if [[ "$2" =~ ^[[:digit:]]+$ ]]; then
                count_in="$2"
            else
                >&2 echo "Invalid count specified: '$2'"
                >&2 echo "$usage"
                return 1
            fi
            shift
            ;;
        *)
            filename="$1"
            if [[ ! -f "$filename" ]]; then
                >&2 echo "File not found: '$filename'"
                >&2 echo "$usage"
                return 1
            fi
            ;;
        esac
        shift
    done
    if [[ -z "$filename" ]]; then
        echo "$usage"
        return 0
    fi
    local count
    if [[ -n "$count_in" ]]; then
        count="$count_in"
    else
        count="15"
    fi
    local output flatten_file_out flattened_filename nice_filename
    output="Created: "
    flatten_file_out="$( flatten_file $filename )"
    flattened_filename="$( echo "$flatten_file_out" | sed -E 's/^[^ ]+ //' )"
    output="$output~$flattened_filename"
    nice_filename="$( add_to_filename 'nice' "$filename" )"
    split_x_per_line "$count" "$flattened_filename" > $nice_filename
    output="$output~$nice_filename"
    if [[ -z "$no_quoted" ]]; then
        local flatten_quote_file_out quoted_filename nice_quoted_filename
        output="$output\n ~"
        flatten_quote_file_out="$( flatten_quote_file $filename )"
        quoted_filename="$( echo "$flatten_quote_file_out" | sed -E 's/^[^ ]+ //' )"
        output="$output~$quoted_filename"
        nice_quoted_filename="$( add_to_filename 'quoted' "$nice_filename" )"
        split_x_per_line "$count" "$quoted_filename" > $nice_quoted_filename
        output="$output~$nice_quoted_filename"
    fi
    echo -e "$output" | column -s '~' -t
}

# Similar to sed 's/str_to_replace/replacement_text/' filename
# Except, each line that has the str_to_replace is replicated for each line in the multi-line replacement text.
# Usage: multi_line_replace <filename> <str_to_replace> <multi-line replacement text>
multi_line_replace () {
    if [[ "$#" -ne '3' ]]; then
        >&2 echo "Usage: multi_line_replace <filename> <str_to_replace> <multi-line replacement text>"
        return 1
    fi
    local filename to_replace replace_with loop_counter loop_max line_to_replace replacement_lines
    filename="$1"
    to_replace="$2"
    replace_with="$3"
    if [[ "$filename" != '-' && ! -f "$filename" ]]; then
        echo -E "File not found: [$filename]."
        return 2
    fi
    cat "$filename" | while IFS= read -r line; do
        if [[ "$line" =~ $to_replace ]]; then
            echo -E "$replace_with" | while read repl_line; do
                echo -E "$line" | sed "s/$to_replace/$repl_line/"
            done
        else
            echo -E "$line"
        fi
    done
}

# Usage: get_all_system_logs
# You'll probably want to pipe this to something or redirect it to a file though.
get_all_system_logs () {
    { cat /var/log/system.log; for l in $( ls /var/log/system.log.* ); do zcat < "$l"; done; } \
    | awk 'BEGIN { al = ""; }
        { if (/^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) /) {
            if (length(al)) { print al; }; al = $0; }
            else { al = al "~" $0; } }
        END { if (length(al)) { print al; } }' \
    | sort -s -k1bM -k2bn -k3.1b,3.2bn -k3.4b,3.5bn -k3.7b,3.8bn \
    | tr '~' '\n'
}

# Usage: check_system_log_timestamp_order <file>
check_system_log_timestamp_order () {
    local file
    file="$1"
    if [[ -z "$file" ]]; then
        echo "Usage: check_system_log_timestamp_order <file>"
        return 1
    fi
    if [[ ! -f "$file" ]]; then
        echo "File not found: $file"
        return 2
    fi
    cat "$file" \
        | awk 'BEGIN { pt = 0 ; pd = ""; }
            { if (/^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)/)
                { d = $1 " " $2 " " $3;
                    m = (index("JanFebMarAprMayJunJulAugSepOctNovDec",$1)+2)/3;
                    gsub(/(:)/, "", $3);
                    t = sprintf("%d%02d%06d", m, $2, $3);
                    if (pt > t) { print (NR-1) ": " pd " > " d " :" NR; }
                    pt = t; pd = d; } }'
}

return 0
