#!/bin/bash
# This file is meant to be sourced.
# It will import all the generic functions into your environment.
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source generic-setup.sh  in your .bash_profile file.
# If you are running into problems, you can get more information on what's going on by using the -v flag.
#   For example,  source generic-setup.sh -v
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

# Putting all the setup stuff in a function so that cleanup is easier.
# Usage: __do_setup "<my source dir>"
__do_setup () {
    # input vars
    local where_i_am
    # Variables defining configuration.
    local title func_dir func_base_file_names funcs_to_double_check required_external desired_external
    # Variables that define strings used in verbose output.
    local info ok warn error
    # Variables that will hold output.
    local problems
    # Variables used in processing.
    local files_to_source cmd_to_check func_file entry exit_code

    # Hopefully contains the full directory path to this file.
    where_i_am="$1"

    # A title describing the functions being added here.
    # Used in output if needed.
    title='Generic'

    # Where to look for the files to source.
    func_dir="${where_i_am}/generic"

    # Dependency notes:
    # echo_color - join_str, escape_escapes (from stream-helpers).

    # All of the filenames to source.
    # These will be looked for in $func_dir and '.sh' will be appended.
    # They will be sourced in this order too.
    func_base_file_names=(
        'join_str'        'stream-helpers'  'fp'          'add_to_filename'
        'change_word'     'chrome_cors'     'echo_color'  'echo_do'
        'get_shell_type'  'hrr'             'i_can'       'java-switchers'
        'jqq'             'print_args'      'ps_grep'
        'to_date'         'to_epoch'
    )

    # These are functions that will be double checked after sourcing to make sure they got added to the environment.
    # If a problem is found, then a message will be output at the end.
    funcs_to_double_check=(
        'change_word'   'chrome_cors'     'echo_color'        'show_colors'        'to_date'       'to_epoch'    'fp'
        'echo_red'      'echo_green'      'echo_yellow'       'echo_blue'          'echo_cyan'     'echo_bold'   'echo_underline'
        'echo_debug'    'echo_info'       'echo_warn'         'echo_error'         'echo_success'  'echo_good'   'echo_bad'
        'echo_do'       'get_shell_type'  'hr'  'hrr'  'hhr'  'pick_a_palette'     'ps_grep'       'tee_pbcopy'  'add_to_filename'
        'i_can'         'can_i'           'java_8_activate'   'java_8_deactivate'  'join_str'      'jqq'         'print_args'
        'strip_colors'  'escape_escapes'  'to_stdout_and_strip_colors_log'         'to_stderr_and_strip_colors_log'
    )

    # These are programs/functions defined externally to check on before sourcing these files.
    # If any aren't available, then an error message will be output and nothing will be sourced.
    required_external=(
        'cat'   'echo'  'printf'  'head'  'tail'
        'sed'   'tee'   'awk'     'grep'  'tr'
        'open'  'ps'    'tput'    'date'
    )

    # These are programs/functions defined externally that might cause some of the new functions to not work properly.
    # Any that aren't available will be included in a message after sourcing all the files.
    desired_external=(
        'dirname'  'basename'  'pbcopy'  '/usr/libexec/java_home'  'jq'
    )

    # These are used for verbose output as line headers.
       ok="        \033[1;32m [ OK ] \033[0m"
     info="    \033[1;21m [ INFO ] \033[0m  "
     warn="  \033[1;33m [ WARN ] \033[0m    "
    error="\033[1;41m [ ERROR ] \033[0m     "

    # This will hold the full path to the files that need to be sourced.
    files_to_source=()

    # This will hold any error messages that should be displayed.
    problems=()

    # And, let's get started!
    __if_verbose "$info" 0 "Loading $title functions."

    if [[ "${#required_external[@]}" -gt '0'  ]]; then
        __if_verbose "$info" 1 "Checking for needed external programs and functions."
        for cmd_to_check in "${required_external[@]}"; do
            if ! __i_can "$cmd_to_check"; then
                problems+=( "Command not found: [$cmd_to_check]." )
                __if_verbose "$error" 2 "The $cmd_to_check command was not found."
            else
                __if_verbose "$ok" 2 "The $cmd_to_check command is available."
            fi
        done
        __if_verbose "$info" 1 "Done checking for needed external programs and functions."
    fi

    __if_verbose "$info" 1 "Checking for source files."
    if [[ "${#func_base_file_names[@]}" -eq '0' ]]; then
        problems+=( "No function files defined." )
        __if_verbose "$error" 2 "The func_base_file_names setup variable does not have any entries."
    elif [[ ! -d "$func_dir" ]]; then
        problems+=( "Function directory not found: [$func_dir]." )
        __if_verbose "$error" 2 "The function directory was not found: [$func_dir]."
    else
        __if_verbose "$ok" 2 "The function directory [$func_dir] exists."
        for entry in "${func_base_file_names[@]}"; do
            func_file="${func_dir}/${entry}.sh"
            if [[ ! -f "$func_file" ]]; then
                problems+=( "File not found: [$func_file]." )
                __if_verbose "$error" 2 "Function file not found: [$func_file]."
            else
                files_to_source+=( "$func_file" )
                __if_verbose "$ok" 2 "The function file [$func_file] exists."
            fi
        done
    fi
    __if_verbose "$info" 1 "Done checking for source files."

    __if_verbose "$info" 1 "Checking for problems encountered so far."
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Could not set up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __if_verbose "$error" 2 "Quitting early due to problems."
        return 2
    fi
    __if_verbose "$ok" 1 "No problems encountered so far."

    __if_verbose "$info" 1 "Sourcing the files."
    for entry in "${files_to_source[@]}"; do
        __if_verbose "$info" 2 "Executing command: [source \"$entry\"]."
        source "$entry"
        exit_code="$?"
        if [[ "$exit_code" -ne '0' ]]; then
            problems+=( "Failed to source the [$entry] file." )
            __if_verbose "$error" 3 "The command to source [$entry] failed with an exit code of [$exit_code]."
        else
            __if_verbose "$ok" 3 "The source command was successful for [$entry]."
        fi
    done
    __if_verbose "$info" 1 "Done sourcing the files."

    if [[ "${#funcs_to_double_check[@]}" -gt '0'  ]]; then
        __if_verbose "$info" 1 "Checking that functions are available."
        for entry in "${funcs_to_double_check[@]}"; do
            if ! __i_can "$entry"; then
                problems+=( "The [$entry] command failed to load." )
                __if_verbose "$error" 2 "Command failed to load: [$entry]."
            else
                __if_verbose "$ok" 2 "The [$entry] command is loaded and ready."
            fi
        done
        __if_verbose "$info" 1 "Done checking that functions are available."
    fi

    __if_verbose "$info" 1 "Doing final check for problems encountered."
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Error(s) encountered while setting up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __if_verbose "$error" 2 "There were errors encountered during setup."
        return 3
    fi
    __if_verbose "$ok" 1 "No problems encountered during setup."

    if [[ "${#desired_external[@]}" -gt '0'  ]]; then
        __if_verbose "$info" 1 "Checking for desired external programs."
        for cmd_to_check in "${desired_external[@]}"; do
            if ! __i_can "$cmd_to_check"; then
                problems+=( "Command $cmd_to_check not found." )
                __if_verbose "$error" 2 "The $cmd_to_check command was not found."
            else
                __if_verbose "$ok" 2 "The $cmd_to_check command is available."
            fi
        done
        if [[ "${#problems[@]}" -gt '0' ]]; then
            printf 'One or more commands used by %s functions are not available:\n' "$title" >&2
            printf '  %s\n' "${problems[@]}" >&2
            printf 'Some newly added functions might not behave as expected.\n' >&2
            __if_verbose "$warn" 2 "Some desired functions were not found."
        fi
        __if_verbose "$info" 1 "Done checking for desired external programs."
    fi

    __if_verbose "$info" 0 "Setup of $title functions complete."
    return 0
}

# Tests if a command is available.
# Usage: if __i_can "foo"; then echo "I can totally foo"; else echo "There's no way I can foo."; fi
__i_can () {
    if [[ "$#" -eq '0' ]]; then
        return 1
    fi
    command -v "$@" > /dev/null 2>&1
}

GENERIC_SETUP_VERBOSE=
# Usage: __if_verbose <level string> <indent-level> <message>
__if_verbose () {
    [[ -n "$GENERIC_SETUP_VERBOSE" ]] && printf '%s %b: %s%s\n' "$( date '+%F %T %Z' )" "$1" "$( printf "%$(( $2 * 2 ))s" )" "$3"
}

if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
    GENERIC_SETUP_VERBOSE='YES'
fi

# Do what needs to be done.
__do_setup "$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

# Now clean up after yourself.
unset -f __do_setup
unset -f __i_can
unset -f __if_verbose
unset GENERIC_SETUP_VERBOSE

return 0
