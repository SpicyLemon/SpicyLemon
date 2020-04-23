#!/bin/bash
# This file is meant to be sourced.
# This file defines the  capture_cmd  function in your environment.
# This function will execute a provided command while outputting the desired streams.
# Both stderr and stdout are captured in separate variables as well as in combination.
# See  capture_cmd --help  for more info.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within.
$( echo -e "\033[1;37msource $( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )\033[0m" )
EOF
    exit 1
fi

capture_cmd () {
    local usage
    usage="$( cat << EOF
capture_cmd - This function will execute a command and capture the output in some environment variables.

Usage: capture_cmd (--all|--stderr|--stdout|--none) [--show-command] <commmand>

    Exaclty one of --all, --stderr, --stdout, or --none must be provided.
    These tell capture_cmd what streams you wish to have displayed as the command is running.
    --all           Display both stderr and stdout during execution.
    --stderr        Display only stderr during execution.
    --stdout        Display only stdout during execution.
    --none          Display neither stderr nor stdout during execution.

    --show-command  This flag is optional.
                    If provided, it must be the 2nd parameter provided to this function.
                    When this flag is provided, the command being executed will be displayed
                    in bold white text prior to it being executed.
                    It will not be part of the captured stdout, though.

    <command>       This is the command to execute as well as any parameters that go with it.
                    In most cases, it does not need to be a single string.
                    It can be the command just as you would normall execute it.
                    One example of where it would need to be quoted and escaped is if
                    multiple commands are being executed.

Results:
    The results of the provided commmand, as well as the command and exit code,
    are captured and stored in some environment variables.
        CAPTURE_CMD_CMD_PARTS  ---> An array containing each piece of the command actually being executed.
        CAPTURE_CMD_CMD_STR  -----> A string version of the command being executed.
        CAPTURE_CMD_EXIT_CODE  ---> The exit code of the command. This function also returns this exit code.
        CAPTURE_CMD_STDERR  ------> The captured stderr output.
        CAPTURE_CMD_STDOUT  ------> The captured stdout output.
        CAPTURE_CMD_ALL  ---------> The combined stderr and stdout output.

Examples:
    capture_cmd --all curl -s -S --data-urlencode 'q=Pizza or Tacos?' 'https://www.flying-ferret.com/cgi-bin/api/v1/transform.cgi'
    capture_cmd --none --show-command foo="\\\$(( 3 + 7 ))"
    capture_cmd --stdout --show-command "echo \\"foo\\"; echo \\"bar\\""

Special exit codes:
    These exit codes can be returned by this command for the following reasons.
        122 - Returned if no parameters are given to this function.
        123 - Returned if the output type (first parameter) has an unexpected value.
        124 - Returned if no command is given.
    Otherwise, this function will return the same exit code that your command returned.

Known issues:
    Combining two parallel streams (stderr and stdout) while maintaing proper order between them is a difficult task.
    This function makes attempts to get it right, but ordering issues can still happen.
    This is most often seen when both stderr and stdout are receiving info rapidly and simultaneously.
    The stderr and stdout streams are being combined in two ways.
        1. For display while the command is running (if --all is given).
        2. As captured values (CAPTURE_CMD_ALL).
    Both instances can end up ordered differently from how the provided command issued them.
    The ordering of data in each stream is maintained, but when combined, its possible that a message
    from one stream will end up before a message from the other that was actually printed first.

EOF
    )"
    local to_show show_command cmd_pieces pieces_for_output cmd_piece tmp_stderr tmp_stdout tmp_stdall
    # Check for no parameters.
    if [[ "$#" -eq '0' ]]; then
        echo 'No parameters provided to capture_cmd.' >&2
        return 122
    fi
    # Get the desired live output.
    case "$1" in
    -h|--help)
        echo -e "$usage"
        return 0
        ;;
    -a|--all|--both)            to_show='ALL'       ;;
    -e|--err|--stderr)          to_show='STDERR'    ;;
    -o|--out|--stdout)          to_show='STDOUT'    ;;
    -x|-n|--none|-s|--silent)   to_show='NONE'      ;;
    *)
        echo "Unknown output type for capture_cmd: [$1]. Possible values: (--all|--stderr|--stdout|--none)." >&2
        return 123
        ;;
    esac
    shift
    # Check for whether or not to show the command before running it.
    if [[ "$1" =~ ^(-v|--verbose|--show-cmd|--show-command)$ ]]; then
        show_command='YES'
        shift
    fi
    # Make sure there's still arguments left to form the command.
    if [[ "$#" -eq '0' || "$@" =~ ^[[:space:]]*$ ]]; then
        echo 'No command provided to capture_cmd.' >&2
        return 124
    fi
    # Everything left represents the command to run.
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
        # We then need to slightly alter the pieces in order to properly output the command (if desired).
        cmd_pieces=( "$@" )
        pieces_for_output=()
        for cmd_piece in "$@"; do
            if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
                # If this piece has a space, a single, or double quote, then it needs to be escaped and wrapped.
                # Escape again all already escaped double quotes, then escape all double quotes.
                # And put the whole thing in double quotes.
                pieces_for_output+=( "\"$( echo -E "$cmd_piece" | sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' )\"" )
            else
                # Otherwise, no change is needed.
                pieces_for_output+=( "$cmd_piece" )
            fi
        done
    fi

    # Cleare out the any previously set values.
    unset CAPTURE_CMD_CMD_PARTS CAPTURE_CMD_CMD_STR CAPTURE_CMD_EXIT_CODE CAPTURE_CMD_STDERR CAPTURE_CMD_STDOUT CAPTURE_CMD_ALL

    # All the input has been received and prepared. It's time to get to work.
    # Create the two variables that contain the command parts and the command string.
    CAPTURE_CMD_CMD_PARTS=( "${cmd_pieces[@]}" )
    CAPTURE_CMD_CMD_STR="${pieces_for_output[@]}"

    # If desired, show the command string.
    if [[ -n "$show_command" ]]; then
        echo -en "\033[1;37m"
        echo -En "$CAPTURE_CMD_CMD_STR"
        echo -e "\033[0m"
    fi
    # Create some temporary files that will be used to hold the results.
    tmp_stderr="$( mktemp -t capture_cmd_stderr )"
    tmp_stdout="$( mktemp -t capture_cmd_stdout )"
    tmp_stdall="$( mktemp -t capture_cmd_stdall )"
    # Execute the command storing the exit code, writing the output to the temporary files, showing only what is desired.
    # How it works:
    #   The command is executed in a curly braces so that the current environment can be affected if needed.
    #       This allows for this function to be used to set environment variables.
    #       It also allows us to save the exit code without worrying about the side effects of anything in the redirection.
    #   Then, the stderr output is redirected to a process that sends the stream through tee.
    #       Tee writes the output to the stderr temp file and also sends the output to stdout.
    #       We pipe that to another command depending on what streams we want displayed while the command is running.
    #           If we want to see stderr, then we pipe to tee again.
    #               This time, tee appends to the stdall temp file and also does its normal output.
    #               The output is then sent to stderr instead of stdout.
    #           If we do not want to see stderr, then we just redirect the output to append to the stdall temp file.
    #   The same thing is done with stdout.
    #       Stdout is redirected to a process that sends it through tee to save to the stdout temp file.
    #       Then it's either sent to tee again for the stdall file and output (this time to stdout), or just the stdall file.
    case "$to_show" in
    ALL)
        { "${CAPTURE_CMD_CMD_PARTS[@]}"; CAPTURE_CMD_EXIT_CODE="$?"; } \
            2> >( tee "$tmp_stderr" | tee -a "$tmp_stdall" >&2 ) \
            1> >( tee "$tmp_stdout" | tee -a "$tmp_stdall" )
        ;;
    STDERR)
        { "${CAPTURE_CMD_CMD_PARTS[@]}"; CAPTURE_CMD_EXIT_CODE="$?"; } \
            2> >( tee "$tmp_stderr" | tee -a "$tmp_stdall" >&2 ) \
            1> >( tee "$tmp_stdout" >> "$tmp_stdall" )
        ;;
    STDOUT)
        { "${CAPTURE_CMD_CMD_PARTS[@]}"; CAPTURE_CMD_EXIT_CODE="$?"; } \
            2> >( tee "$tmp_stderr" >> "$tmp_stdall" ) \
            1> >( tee "$tmp_stdout" | tee -a "$tmp_stdall" )
        ;;
    *)
        # I'm not a fan of not having a default case.
        # If the to_show value is not NONE then we'll show an error message.
        # If it is NONE, then no error message is shown.
        # Either way, we execute the command without showing anything while it's running.
        if [[ "$to_show" != 'NONE' ]]; then
            # If you get this message, some to_show stuff isn't set up correctly either here or when pulling in the parameter.
            echo -e "\033[1;41mUnexpected to_show value: [\033[0m$to_show\033[1;41m].\033[0m" \
                  "\n\033[33mDefaulting to [\033[0mNONE\033[33m] behavior.\033[0m" \
                  "\n\033[31mThis really shouldn't have happened and indicates a bug in the code.\033[0m" \
                >&2
        fi
        { "${CAPTURE_CMD_CMD_PARTS[@]}"; CAPTURE_CMD_EXIT_CODE="$?"; } \
            2> >( tee "$tmp_stderr" >> "$tmp_stdall" ) \
            1> >( tee "$tmp_stdout" >> "$tmp_stdall" )
        ;;
    esac
    # Pull the contents of the temporary files into the appropriate variables.
    CAPTURE_CMD_STDERR="$( cat "$tmp_stderr" )"
    CAPTURE_CMD_STDOUT="$( cat "$tmp_stdout" )"
    CAPTURE_CMD_ALL="$( cat "$tmp_stdall" )"
    # Clean up the temporary files.
    rm "$tmp_stderr"
    rm "$tmp_stdout"
    rm "$tmp_stdall"
    # Return the exit code that the command produced.
    return "$CAPTURE_CMD_EXIT_CODE"
}

# Helpful for debugging.
# This diplays all the variables defined by capture_cmd
# Usage: __capture_cmd_print_vars
#    or: __capture_cmd_print_vars "$previous_exit_code"
#    or: __capture_cmd_print_vars 'yes'
#    or: __capture_cmd_print_vars "$previous_exit_code" 'yes'
# The "$previous_exit_code" is an optional exit code value that was possibly received from capture_cmd.
#   If provided, it will be output next to the CAPTURE_CMD_EXIT_CODE value for comparison.
#   If provided, it must be the first parameter.
# The 'yes' there indicates that you'd like the various values colored.
__capture_cmd_print_vars () {
    local retval with_colors part
    local cb cx cr co ce ca cc
    if [[ "$1" =~ ^[[:digit:]]+$ ]]; then
        retval="$1"
        with_colors="$2"
    else
        with_colors="$1"
    fi
    if [[ -n "$with_colors" ]]; then
        cb="\033[1m"    # color bold
        if [[ -n "$CAPTURE_CMD_EXIT_CODE" && "$CAPTURE_CMD_EXIT_CODE" -ne '0' ]]; then
            cx="\033[1;45m"   # color exit code: non-zero: bold with purple background.
        else
            cx="\033[42m"     # color exit code: zero: green background.
        fi
        if [[ -n "$retval" && -n "$CAPTURE_CMD_EXIT_CODE" && "$retval" -ne "$CAPTURE_CMD_EXIT_CODE" ]]; then
            cr="\033[1;41m"   # color retval: not equal to CAPTURE_CMD_EXIT_CODE: bold with red background.
        else
            cr="$cx"          # color retval: equal to CAPTURE_CMD_EXIT_CODE: same as the color for CAPTURE_CMD_EXIT_CODE.
        fi
        co="\033[32m"   #color stdout: green text.
        ce="\033[31m"   #color stderr: red text.
        ca="\033[36m"   #color stdall: cyan text.
        cc="\033[0m"    #color clear:  reset all.
    fi
    echo -En "CAPTURE_CMD_CMD_PARTS:"
    for part in "${CAPTURE_CMD_CMD_PARTS[@]}"; do
        echo -en " [$cb$part$cc]"
    done
    echo -E  ''
    echo -e  "  CAPTURE_CMD_CMD_STR: [$cb$CAPTURE_CMD_CMD_STR$cc]"
    echo -en "CAPTURE_CMD_EXIT_CODE: [$cx$CAPTURE_CMD_EXIT_CODE$cc]"
    [[ -n "$retval" ]] && echo -e " Received [$cr$retval$cc]" || echo -E ''
    echo -e  "   CAPTURE_CMD_STDOUT: [$co$CAPTURE_CMD_STDOUT$cc]"
    echo -e  "   CAPTURE_CMD_STDERR: [$ce$CAPTURE_CMD_STDERR$cc]"
    echo -e  "   CAPTURE_CMD_STDALL: [$ca$CAPTURE_CMD_ALL$cc]"
}

# Helper for debugging.
# It will call capture_cmd with the provided parameters.
# Then it will call __capture_cmd_print_vars.
# It will also return the same exit code that capture_cmd returned.
# Usage is exactly the same as capture_cmd.
__capture_cmd_debug () {
    local retval;
    capture_cmd "$@"
    retval="$?"
    echo -E '-------------------------------------------'
    __capture_cmd_print_vars "$retval"
    return "$retval"
}
