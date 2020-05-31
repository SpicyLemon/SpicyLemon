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
# Usage: __generic_do_setup "<my source dir>"
__generic_do_setup () {
    # input vars
    local where_i_am
    # Variables defining configuration.
    local title func_dir func_base_file_names extra_funcs_to_check required_external desired_external
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
    func_dir="${where_i_am}"

    # All of the filenames to source.
    # These will be looked for in $func_dir and '.sh' will be appended.
    # Handy command for generating this:
    #   ls *.sh | grep -v 'generic-setup' | sed 's/\.sh$//' | re_line -p -n 5 -d '~' -w "'" | column -s '~' -t | tee_pbcopy
    func_base_file_names=(
        'add_to_filename'      'change_word'     'check_system_log_timestamp_order'  'chrome_cors'              'echo_color'
        'echo_do'              'escape_escapes'  'fp'                                'get_all_system_logs'      'get_shell_type'
        'getlines'             'hrr'             'i_can'                             'java_8_activate'          'java_8_deactivate'
        'join_str'             'jqq'             'multi_line_replace'                'pretty_json'              'print_args'
        'ps_grep'              're_line'         'show_last_exit_code'               'string_repeat'            'strip_colors'
        'strip_final_newline'  'tee_pbcopy'      'tee_strip_colors'                  'tee_strip_colors_pbcopy'  'to_date'
        'to_epoch'             'ugly_json'
    )

    # These are extra functions defined in the files that will be checked (along with the primary functions).
    extra_funcs_to_check=(
        'echo_red'    'echo_green'  'echo_yellow'  'echo_blue'   'echo_cyan'     'echo_bold'  'echo_underline'
        'echo_debug'  'echo_info'   'echo_warn'    'echo_error'  'echo_success'  'echo_good'  'echo_bad'
        'hr'  'hhr'  'pick_a_palette'
        'can_i'
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
        'dirname'  'basename'  'pbcopy'  'pbpaste'  '/usr/libexec/java_home'  'jq'  'fzf'
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
    __generic_if_verbose "$info" 0 "Loading $title functions."

    if [[ "${#required_external[@]}" -gt '0'  ]]; then
        __generic_if_verbose "$info" 1 "Checking for needed external programs and functions."
        for cmd_to_check in "${required_external[@]}"; do
            if ! __generic_i_can "$cmd_to_check"; then
                problems+=( "Command not found: [$cmd_to_check]." )
                __generic_if_verbose "$error" 2 "The $cmd_to_check command was not found."
            else
                __generic_if_verbose "$ok" 2 "The $cmd_to_check command is available."
            fi
        done
        __generic_if_verbose "$info" 1 "Done checking for needed external programs and functions."
    fi

    __generic_if_verbose "$info" 1 "Checking for source files."
    if [[ "${#func_base_file_names[@]}" -eq '0' ]]; then
        problems+=( "No function files defined." )
        __generic_if_verbose "$error" 2 "The func_base_file_names setup variable does not have any entries."
    elif [[ ! -d "$func_dir" ]]; then
        problems+=( "Function directory not found: [$func_dir]." )
        __generic_if_verbose "$error" 2 "The function directory was not found: [$func_dir]."
    else
        __generic_if_verbose "$ok" 2 "The function directory [$func_dir] exists."
        for entry in "${func_base_file_names[@]}"; do
            func_file="${func_dir}/${entry}.sh"
            if [[ ! -f "$func_file" ]]; then
                problems+=( "File not found: [$func_file]." )
                __generic_if_verbose "$error" 2 "Function file not found: [$func_file]."
            else
                files_to_source+=( "$func_file" )
                __generic_if_verbose "$ok" 2 "The function file [$func_file] exists."
            fi
        done
    fi
    __generic_if_verbose "$info" 1 "Done checking for source files."

    __generic_if_verbose "$info" 1 "Checking for problems encountered so far."
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Could not set up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __generic_if_verbose "$error" 2 "Quitting early due to problems."
        return 2
    fi
    __generic_if_verbose "$ok" 1 "No problems encountered so far."

    __generic_if_verbose "$info" 1 "Sourcing the files."
    for entry in "${files_to_source[@]}"; do
        __generic_if_verbose "$info" 2 "Executing command: [source \"$entry\"]."
        source "$entry"
        exit_code="$?"
        if [[ "$exit_code" -ne '0' ]]; then
            problems+=( "Failed to source the [$entry] file." )
            __generic_if_verbose "$error" 3 "The command to source [$entry] failed with an exit code of [$exit_code]."
        else
            __generic_if_verbose "$ok" 3 "The source command was successful for [$entry]."
        fi
    done
    __generic_if_verbose "$info" 1 "Done sourcing the files."

    __generic_if_verbose "$info" 1 "Checking that functions are available."
    for entry in "${func_base_file_names[@]}" "${extra_funcs_to_check[@]}"; do
        if ! __generic_i_can "$entry"; then
            problems+=( "The [$entry] command failed to load." )
            __generic_if_verbose "$error" 2 "Command failed to load: [$entry]."
        else
            __generic_if_verbose "$ok" 2 "The [$entry] command is loaded and ready."
        fi
    done
    __generic_if_verbose "$info" 1 "Done checking that functions are available."

    __generic_if_verbose "$info" 1 "Doing final check for problems encountered."
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Error(s) encountered while setting up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __generic_if_verbose "$error" 2 "There were errors encountered during setup."
        return 3
    fi
    __generic_if_verbose "$ok" 1 "No problems encountered during setup."

    if [[ "${#desired_external[@]}" -gt '0'  ]]; then
        __generic_if_verbose "$info" 1 "Checking for desired external programs."
        for cmd_to_check in "${desired_external[@]}"; do
            if ! __generic_i_can "$cmd_to_check"; then
                problems+=( "Command $cmd_to_check not found." )
                __generic_if_verbose "$error" 2 "The $cmd_to_check command was not found."
            else
                __generic_if_verbose "$ok" 2 "The $cmd_to_check command is available."
            fi
        done
        if [[ "${#problems[@]}" -gt '0' ]]; then
            printf 'One or more commands used by %s functions are not available:\n' "$title" >&2
            printf '  %s\n' "${problems[@]}" >&2
            printf 'Some newly added functions might not behave as expected.\n' >&2
            __generic_if_verbose "$warn" 2 "Some desired functions were not found."
        fi
        __generic_if_verbose "$info" 1 "Done checking for desired external programs."
    fi

    __generic_if_verbose "$info" 0 "Setup of $title functions complete."
    return 0
}

# Tests if a command is available.
# Usage: if __generic_i_can "foo"; then echo "I can totally foo"; else echo "There's no way I can foo."; fi
__generic_i_can () {
    if [[ "$#" -eq '0' ]]; then
        return 1
    fi
    command -v "$@" > /dev/null 2>&1
}

GENERIC_SETUP_VERBOSE=
# Usage: __generic_if_verbose <level string> <indent-level> <message>
__generic_if_verbose () {
    [[ -n "$GENERIC_SETUP_VERBOSE" ]] && printf '%s %b: %s%s\n' "$( date '+%F %T %Z' )" "$1" "$( printf "%$(( $2 * 2 ))s" )" "$3"
}

if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
    GENERIC_SETUP_VERBOSE='YES'
fi

# Do what needs to be done.
__generic_do_setup "$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

# Now clean up after yourself.
unset -f __generic_do_setup
unset -f __generic_i_can
unset -f __generic_if_verbose
unset GENERIC_SETUP_VERBOSE

return 0
