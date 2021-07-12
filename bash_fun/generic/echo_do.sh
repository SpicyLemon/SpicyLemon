#!/bin/bash
# This file contains the echo_do function that outputs a command before executing it.
# This file can be sourced to add the echo_do function to your environment.
# This file can also be executed to run the echo_do function without adding it to your environment.
#
# File contents:
#   echo_do  --> Outputs a command in bright white, then executes it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Output a command, then execute it.
# Usage: echo_do <command> [<arg1> [<arg2> ...]]
#   or   echo_do "command string"
# Examples:
#   echo_do say -vVictoria -r200 "Buu Whoa"
#   echo_do "say -vVictoria -r200 \"YEAH BUDDY\""
# If no command is provided, this will return with exit code 124.
echo_do () {
    local cmd_pieces pieces_for_output cmd_piece retval
    # Check for no parameters.
    # Make sure there's still arguments left to form the command.
    if [[ "$#" -eq '0' || "$@" =~ ^[[:space:]]*$ ]]; then
        printf 'echo_do: No command provided.\n' >&2
        return 124
    fi
    # Do a little processing on the provided arguments.
    if [[ "$#" -eq '1' && ( "$@" =~ [[:space:]\(=] || -z "$( command -v "$@" )" ) ]]; then
        # If there's only 1 argument and
        #   it contains a space, open parenthesis, or an equals
        #   or it is not an actual command
        # then we need to run it using eval.
        # This primarily allows for setting environment variables using this function.
        cmd_pieces=( 'eval' "$@" )
        pieces_for_output=( "$@" )
    else
        # Otherwise, we can just throw everything into the command pieces as it is.
        cmd_pieces=( "$@" )
        # We then need to slightly alter the pieces in order to properly output the command.
        pieces_for_output=()
        for cmd_piece in "$@"; do
            if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
                # If this piece has a space, a single, or double quote, then it needs to be escaped and wrapped.
                # Escape again all already escaped double quotes, then escape all double quotes.
                # And put the whole thing in double quotes.
                pieces_for_output+=( "\"$( printf '%s' "$cmd_piece" | sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' )\"" )
            else
                # Otherwise, no change is needed.
                pieces_for_output+=( "$cmd_piece" )
            fi
        done
    fi

    # Show the command string in bold white.
    printf '\033[1;37m%s\033[0m\n' "${pieces_for_output[*]}"
    # Execute the command.
    "${cmd_pieces[@]}"
    retval=$?
    printf '\n'
    return $retval
}

if [[ "$sourced" != 'YES' ]]; then
    echo_do "$@"
    exit $?
fi
unset sourced

return 0
