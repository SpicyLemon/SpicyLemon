#!/bin/bash
# This file houses functions for doing common text manipulation stuff.
# File contents:
#   split_x_per_line  -----------------> Break a long comma separated string into a number of entries per line.
#   ugly_json  ------------------------> Pulls some json from the clipboard and uses jq to make it compact, then puts it back into the clipboard.
#   flatten  --------------------------> Flattens input piped into it.
#   flatten_x  ------------------------> Flattens input piped into it and splits it, x per line.
#   flatten_clipboard  ----------------> Similar to flatten, but takes the input from the clipboard, and puts the result back in the clipboard.
#   flatten_x_clipboard  --------------> Similar to flatten_x, but takes the input from the clipboard, and puts the result back in the clipboard.
#   flatten_quote  --------------------> Similar to flatten, but single quotes each entry.
#   flatten_quote_x  ------------------> Similar to flatten_x, but single quotes each entry.
#   flatten_quote_clipboard  ----------> Similar to flatten_quote, but takes input from the clipbaord, and puts the result back in the clipboard.
#   flatten_quote_x_clipboard  --------> Similar to flatten_quote_x, but takes the input from the clipboard, and puts the result back in the clipboard.
#   flatten_double_quote  -------------> Similar to flatten, but double quotes each entry.
#   flatten_double_quote_x  -----------> Similar to flatten_x, but double quotes each entry.
#   flatten_double_quote_clipboard  ---> Similar to flatten_quote, but takes the input from the clipboard, and puts the result back in the clipboard.
#   flatten_double_quote_x_clipboard  -> Similar to flatten_quote_x, but takes the input from the clipboard, and puts the result back in the clipboard.
#   to_column_clipboard  --------------> Splits input by commas (with optional spaces) into a single column of values.
#   quote_clipboard  ------------------> Adds a single quote to the beginning and end of each line in the clipboard.
#   double_quote_clipboard  -----------> Adds double quotes to the beginning and end of each line in the clipboard.
#   unquote_clipboard  ----------------> For each line in the clipboard, removes a matching single or double quote at the beginning and end of a line.
#   strip_final_newline  --------------> Strips the final newline character from a string. Only the last line is changed.
#

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

# Splits a long series of comma separated values into lines containing a certain amount of entries.
# Usage: split_x_per_line "<count>" "<file>"
#  or    split_x_per_line "<count>" "<input>"
#  or    cat file | split_x_per_line "<count>"
split_x_per_line () {
    local count input
    count="$1"
    shift
    if [[ -z "$count" || "$count" -le 0 ]]; then
        >&2 echo "Usage: split_x_per_line <count> <file>"
        return 1
    fi
    if [[ $# -ge 1 && -f "$1" ]]; then
        input=$( cat "$1" )
    elif [[ $# -ge 1 ]]; then
        input="$*"
    else
        input=$( cat "-" )
    fi
    if [[ -n "$input" ]]; then
        echo -E "$input" | sed -E "s/($( string_repeat "[^,]+," "$count" ) )/\1~/g" | tr '~' '\n' | sed -E 's/ +$//'
    else
        >&2 echo "No input provided."
    fi
}

# Pulls json string from clipboard, makes it compact, puts it back into the clipboard.
# The -v option also outputs it to stdout
# Usage: ugly_json
ugly_json () {
    pbpaste | jq -c '.' | pbcopy
    if [[ -n "$1" && "$1" == "-v" ]]; then
        pbpaste | jq -c '.'     # re-doing it so we get the colors
        echo "(Copied to clipboard)"
    fi
}

# Usage: <do stuff> | flatten
flatten () {
    sed -E 's/^ +//; s/ *$/, /;' | tr -d '\n' | sed -E 's/, $//' | tr -d '\n'
}

# Usage: <do stuff> | flatten_x <number>
flatten_x () {
    flatten | split_x_per_line "$@"
}

# Takes the contents of the clipboard and combines it into a single line, with each previous line separated by commas.
# Usage: flatten_clipboard
flatten_clipboard () {
    pbpaste | flatten | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Takes the contents of the clipboard and combines it into lines with x entries per line, all entries comma separated.
# Usage: flatte_x_clipboard <count>
flatten_x_clipboard () {
    local count v
    if [[ -n "$1" && "$1" == '-v' ]]; then
        count="$2"
        v="$1"
    else
        count="$1"
        v="$2"
    fi
    pbpaste | flatten_x "$count" | pbcopy
    __output_clipboard_if_option_given "$v"
}

# Usage: <do stuff> | flatten_quote
flatten_quote () {
    sed -E "s/^/'/; s/ *$/', /" | tr -d '\n' | sed -E 's/, $//' | tr -d '\n'
}

# Usage: <do stuff> | flatten_quote_x <number>
flatten_quote_x () {
    flatten_quote | split_x_per_line "$@"
}

# Takes the contents of the clipboard and combines it into a single line, with each previous line single-quoted and saparated by commas.
# Usage: flatten_quote_clipboard
flatten_quote_clipboard () {
    pbpaste | flatten_quote | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Takes the contents of the clipboard and combines it into lines with x entries per line. All entries are single-quoted and comma separated.
# Usage: flatten_quote_x_clipboard <count>
flatten_quote_x_clipboard () {
    local count v
    if [[ -n "$1" && "$1" == '-v' ]]; then
        count="$2"
        v="$1"
    else
        count="$1"
        v="$2"
    fi
    pbpaste | flatten_quote_x "$count" | pbcopy
    __output_clipboard_if_option_given "$v"
}

# Usage: <do stuff> | flatten_double_quote
flatten_double_quote () {
    sed -E 's/^/"/; s/ *$/", /' | tr -d '\n' | sed -E 's/, $//' | tr -d '\n'
}

# Usage: <do stuff> | flatten_double_quote_x <number>
flatten_double_quote_x () {
    flatten_double_quote | split_x_per_line "$@"
}

# Takes the contents of the clipboard and combines it into a single line, with each previous line single-quoted and saparated by commas.
# Usage: flatten_quote_clipbaord
flatten_double_quote_clipboard () {
    pbpaste | flatten_double_quote | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Takes the contents of the clipboard and combines it into lines with x entries per line. All entries are single-quoted and comma separated.
# Usage: flatten_quote_x_clipboard <count>
flatten_double_quote_x_clipboard () {
    local count v
    if [[ -n "$1" && "$1" == '-v' ]]; then
        count="$2"
        v="$1"
    else
        count="$1"
        v="$2"
    fi
    pbpaste | flatten_double_quote_x "$count" | pbcopy
    __output_clipboard_if_option_given "$v"
}

# Takes the contets of the clipboard and turns it back into a column of data.
# This is sort of the compliment to the flatten functions.
# Usage: to_column_clipboard
to_column_clipboard () {
    pbpaste | sed 's/,[[:space:]]*/\'$'\n''/g' | awk NF | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Adds a single quote to the beginning and end of each line of input.
# Usage: quote_clipboard
quote_clipboard () {
    pbpaste | sed "s/^/'/; s/[[:space:]]*$/'/;" | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Adds double quotes to the beginning and end of each line of input.
# Usage: double_quote_clipboard
double_quote_clipboard () {
    pbpaste | sed 's/^/"/; s/[[:space:]]*$/"/;' | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Strips either single or double quotes from the beginning and end of each line of input.
# Usage unquote_clipboard
unquote_clipboard () {
    pbpaste | sed 's/^\(['\'\"']\)\(.*\)\1$/\2/' | pbcopy
    __output_clipboard_if_option_given "$1"
}

# Usage: <do stuff> | strip_final_newline
strip_final_newline () {
    if [[ -n "$1" ]]; then
        echo -E "$1" | strip_final_newline
        return 0
    fi
    awk ' { if(p) print(l); l=$0; p=1; } END { printf("%s", l); } '
}

# Internal function that outputs the clipboard if provided -v.
# Usage: __output_clipboard_if_option_given "$1"
__output_clipboard_if_option_given () {
    if [[ -n "$1" && "$1" == "-v" ]]; then
        pbpaste
        echo "(Copied to clipboard)"
    fi
}

return 0
