#!/bin/bash
# This file contains various functions for doing some handy things with streams.
# This file is meant to be sourced to add the functions to your environment.
#
# File contents:
#   strip_colors  --------------------> Strips the color stuff from a stream.
#   escape_escapes  ------------------> Escapes any escape characters in a stream.
#   to_stdout_and_strip_colors_log  --> Outputs to stdout and logs to a file with color stuff stripped out.
#   to_stdout_and_strip_colors_log  --> Outputs to stderr and logs to a file with color stuff stripped out.
#   tee_pbcopy  ----------------------> Outputs to stdout as well as copy it to the clipboard.
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

# Removes color escape codes from the provided stream.
# Usage: <stuff> | strip_colors
strip_colors () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | strip_colors
        return 0
    fi
    sed -E "s/$( echo -e "\033" )\[[[:digit:]]+(;[[:digit:]]+)*m//g"
}

# Escapes all escape characters in a stream so that they appear as "\033" in the output.
# Usage: <stuff> | escape_escapes
escape_escapes () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | escape_escapes
        return 0
    fi
    sed -E "s/$( echo -e "\033" )/\\\033/g"
}

# Takes a colored input stream, outputs the stream to stdout unchanged.
# It also strips the color info out of the stream and appends that to the provided logfile.
# Usage: <stuff> | to_stdout_and_strip_colors_log "logfile"
to_stdout_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stdout_and_strip_colors_log <filename>"
    fi
    cat - > >( tee >( strip_colors >> "$1" ) )
}

# Takes a colored input stream, outputs the stream to stderr unchanged.
# It also strips the color info out of the stream and appends that to the provided logfile.
# Usage: <stuff> | to_stderr_and_strip_colors_log "logfile"
to_stderr_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stderr_and_strip_colors_log <filename>"
    fi
    cat - > >( >&2 tee >( strip_colors >> "$1" ) )
}

# Takes in a stream and outputs it to stdout while also putting it into pbcopy (with trailing newline removed).
# Usage: <do stuff> | tee_pbcopy
tee_pbcopy () {
    tee >( awk '{if(p) print(l);l=$0;p=1;} END{printf("%s",l);}' | pbcopy )
}

return 0
