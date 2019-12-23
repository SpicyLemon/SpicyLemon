#!/bin/bash
# Runs a Spring boot gradle app with the supplied command-line arguments.
# To use this script, place it in the same directory as the ./gradlew script you want to use.
# Make sure the script is executable, and then call it.
# Examples:
#   command: ./run.sh
#   same as: ./gradlew bootRun
#
#   command: ./run.sh -h
#   same as: ./graldew bootRun -Pargs=-h
#
#   command: ./run.sh arg1 --foo fooParam --bar barParam 'an arg with spaces'
#   same as: ./gradlew bootRun -Pargs=arg1,--foo,fooParam,--bar,barParam,'an arg with spaces'
#

# This is the file or command that will be executed or called.
GRADLE_CMD='./gradlew'
# This is the task that will be run.
GRADLE_TASK='bootRun'

SCRIPT_IN_DIR="$( dirname "$0" )"
SCRIPT_NAME="$( basename "$0" )"
SCRIPT="$SCRIPT_IN_DIR/$SCRIPT_NAME"

# Joins all provided parameters using the provided delimiter.
# Usage: string_join <delimiter> [<arg1> [<arg2>... ]]
string_join () {
    local d retval
    d="$1"
    shift
    retval="$1"
    shift
    while [[ "$#" -gt '0' ]]; do
        retval="${retval}${d}$1"
        shift
    done
    echo -E -n "$retval"
}

# Escapes (for bash output) all provided parameters and joins them using the provided delimiter.
# Usage: escape_and_join <delimiter> [<arg1> [<arg2>... ]]
escape_and_join () {
    local d retval
    d="$1"
    shift
    retval="$1"
    shift
    while [[ "$#" -gt '0' ]]; do
        retval="${retval}${d}$( cli_escape "$1" )"
        shift
    done
    echo -E -n "$retval"
}

# Escapes slashes and quotes.
# Usage: cli_escape <text>
cli_escape () {
    local str_in do_wrap str_out
    str_in="$@"
    # If the string has a space, single quote, or double quote, we need to wrap it in quotes when we're done.
    if [[ "$str_in" =~ [[:space:]\'\"] ]]; then
        do_wrap="YES"
    fi
    # Replace all space-like characters with actual spaces, change \ to \\, then change " to \".
    str_out="$( echo -E "$str_in" | sed -E 's/[[:space:]]/ /g; s/\\/\\\\/g; s/"/\\"/g;' )"
    # And wrap it in quotes if we need to.
    if [[ -n "$do_wrap" ]]; then
        str_out="\"$str_out\""
    fi
    echo -E -n "$str_out"
}

# Checks to make sure that the provided command is valid, and that there is a task defined.
# If something is wrong, a message will be sent to stderr and the return value will not be 0.
# If everything's hunky dory, nothing will be output, and the return value will be 0.
# Return code meanings:
#   0   --> Everything is good.
#   10  --> Part of the setup is undefined.
#   11  --> Could not find thing to execute.
# Usage: check_setup || exit $?
check_setup () {
    # Make sure we can do what we want to do, and give a nice message if we can't.
    if [[ "$GRADLE_CMD" =~ ^[[:space:]]*$ ]]; then
        >&2 echo -E "No command defined to execute. Check setup of $SCRIPT with respect to GRADLE_CMD."
        return 10
    fi
    if [[ "$GRADLE_TASK" =~ ^[[:space:]]*$ ]]; then
        >&2 echo -E "No gradle task defined to run. Check setup of $SCRIPT with respect to GRADLE_TASK."
        return 10
    fi

    if [[ "$GRADLE_CMD" =~ / ]]; then
        # If there's a slash in it, it'll be treated like an executable file.
        # e.g. ./gradlew    or    sub-project/gradlew
        # Create a full path to the executable in case we need to print an error message.
        if [[ "$GRADLE_CMD" =~ ^./ ]]; then
            cmd_file="$( pwd )/${GRADLE_CMD:2}"
        elif [[ "$GRADLE_CMD" =~ ^/ ]]; then
            cmd_file="$GRADLE_CMD"
        else
            cmd_file="$( pwd )/$GRADLE_CMD"
        fi
        # Since GRADLE_CMD is what will actually be executed, check that, but output the cmd file from above if something is wrong.
        if [[ ! -f "$GRADLE_CMD" ]]; then
            >&2 echo -E "File not found: $cmd_file"
            return 11
        elif [[ -d "$GRADLE_CMD" ]]; then
            >&2 echo -E "Directory found when file expected: $cmd_file"
            return 11
        elif [[ ! -x "$GRADLE_CMD" ]]; then
            >&2 echo -E "File not executable: $cmd_file"
            return 11
        fi
    else
        # No slash means it'll be treated like a command (and looked for in PATH).
        # e.g. gradle
        cmd_parts=( $GRADLE_CMD )
        if [[ -z "$( command -v "${cmd_parts[0]}" 2> /dev/null )" ]]; then
            >&2 echo -E "Command not found: $GRADLE_CMD"
            return 11
        fi
    fi
}

# Runs the desired command with the desired task and optional additional arguments.
# This function will return with the same code that the provided command returns with.
# Usage: run_gradle_task_with_params <command> <task> [<arg1> [<arg2> ...]]
run_gradle_task_with_params () {
    local base_cmd task args cmd_pieces cmd_for_output
    base_cmd="$1"
    shift
    task="$1"
    shift
    args=( "$@" )
    cmd_pieces=( "$base_cmd" "$task" )
    cmd_for_output="${cmd_pieces[@]}"
    if [[ "$#" -gt '0' && -n "$@" ]]; then
        args=( "$@" )
        # Because I'm putting the pieces of the command in an array to execute them, they don't need to be escaped.
        # But in order to ouput the command that's being run, that ouput needs to be escaped.
        # So I have to build two slightly different things. :grumpyface:
        cmd_pieces+=( "-Pargs=$( string_join "," "${args[@]}" )" )
        cmd_for_output="$cmd_for_output -Pargs=$( escape_and_join "," "${args[@]}" )"
    fi
    # Output what we're about to do.
    echo -e -n "\033[1;37m"
    echo -E -n "$cmd_for_output"
    echo -e "\033[0m"
    # Do it!
    "${cmd_pieces[@]}"
    return $?
}

cd "$SCRIPT_IN_DIR"
check_setup || exit $?
run_gradle_task_with_params "$GRADLE_CMD" "$GRADLE_TASK" "$@"
exit $?
