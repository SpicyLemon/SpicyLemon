#!/bin/bash
# This file is meant to be sourced.
# It will import all the go helper functions into your environment.
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source go-setup.sh  in your .bash_profile file.
# If you are running into problems, you can get more information on what's going on by using the -v flag.
#   For example,  source go-setup.sh -v
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi

unset sourced

# Putting all the setup stuff in a function so that cleanup is easier.
# Usage: __go_do_setup "<my source dir>"
__go_do_setup () {
    # input vars
    local where_i_am
    # Variables defining configuration.
    local title func_dir func_base_file_names extra_funcs_to_check required_external
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
    title='Go Helpers'

    # Where to look for the files to source.
    func_dir="${where_i_am}"

    # All of the filenames to source.
    # These will be looked for in $func_dir and '.sh' will be appended.
    # Handy command for generating this:
    #   ls *.sh | grep -v 'go-setup' | sed 's/\.sh$//' | re_line -p -n 5 -d '~' -w "'" | column -s '~' -t | sed 's/^/        /' | tee_pbcopy
    func_base_file_names=(
        'go_find_funcs_without_comments'    'go_get_func'    'go_mod_fix'
    )

    # These are extra functions defined in the files that will be checked (along with the primary functions).
    # Handy command for generating this:
    #   grep -E '^[[:alnum:]_]+[[:space:]]+\([[:space:]]*\)' * 2> /dev/null | sed 's/ .*$//' | grep -v -E -e '^(.*).sh:\1$' -e 'go-setup.sh' | sed 's/^.*\.sh://' | grep -v '^__' | sort | re_line -n 6 -d '~' -w "'" -p | column -s '~' -t | sed 's/^/        /' | tee_pbcopy
    extra_funcs_to_check=()

    # These are commands defined externally to check on before sourcing these files.
    # If any aren't available, then an error message will be output and nothing will be sourced.
    # To add or remove things and keep it nice and formatted, add an entry where you'd want it in the list.
    # Then copy all the lines into your clipboard and ...
    # pbpaste | re_line -p -n 5 -d '~' -b '[[:space:]]+' | column -s '~' -t | sed 's/^/        /' | tee_pbcopy
    required_external=(
        'go' 'awk' 'sed' 'grep'
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
    __go_if_verbose "$info" 0 "Loading $title functions."

    # Ensure all external commands are available.
    if [[ "${#required_external[@]}" -gt '0'  ]]; then
        __go_if_verbose "$info" 1 "Checking for required external commands."
        for cmd_to_check in "${required_external[@]}"; do
            if ! command -v "$cmd_to_check" > /dev/null 2>&1; then
                problems+=( "Command not found: [$cmd_to_check]." )
                __go_if_verbose "$error" 2 "The $cmd_to_check command was not found."
            else
                __go_if_verbose "$ok" 2 "The $cmd_to_check command is available."
            fi
        done
        __go_if_verbose "$info" 1 "Done checking for required external commands."
    fi

    # Quit now if there was already a problem.
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Could not set up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __go_if_verbose "$error" 1 "Quitting early due to problems."
        return 2
    fi

    # Find all the files to source.
    __go_if_verbose "$info" 1 "Finding files to source."
    if [[ "${#func_base_file_names[@]}" -eq '0' ]]; then
        __go_if_verbose "$error" 2 "The func_base_file_names setup variable does not have any entries."
    elif [[ ! -d "$func_dir" ]]; then
        __go_if_verbose "$error" 2 "The function directory was not found: [$func_dir]."
    else
        __go_if_verbose "$info" 2 "Locating specific source files in the directory [$func_dir]."
        for entry in "${func_base_file_names[@]}"; do
            func_file="${func_dir}/${entry}.sh"
            if [[ ! -f "$func_file" ]]; then
                __go_if_verbose "$error" 3 "Source file not found: [$func_file]."
            else
                files_to_source+=( "$func_file" )
                __go_if_verbose "$ok" 3 "The source file [$func_file] exists and will be sourced."
            fi
        done
        __go_if_verbose "$info" 2 "Done locating specific source files in the directory [$func_dir]."
    fi
    __go_if_verbose "$info" 1 "Done finding files to source."

    # Source all the files found.
    __go_if_verbose "$info" 1 "Sourcing files."
    for entry in "${files_to_source[@]}"; do
        __go_if_verbose "$info" 2 "Executing command: [source \"$entry\"]."
        source "$entry"
        exit_code="$?"
        if [[ "$exit_code" -ne '0' ]]; then
            problems+=( "The command [source \"$entry\"] exited code [$exit_code]." )
            __go_if_verbose "$error" 3 "The command [source \"$entry\"] exited code [$exit_code]."
        else
            __go_if_verbose "$ok" 3 "The command [source \"$entry\"] was successful."
        fi
    done
    __go_if_verbose "$info" 1 "Done sourcing files."

    # Check that the desired commands are now available.
    __go_if_verbose "$info" 1 "Checking that functions are available."
    for entry in "${func_base_file_names[@]}" "${extra_funcs_to_check[@]}"; do
        if ! command -v "$entry" > /dev/null 2>&1; then
            __go_if_verbose "$error" 2 "Command failed to load: [$entry]."
        else
            __go_if_verbose "$ok" 2 "The [$entry] command is loaded and ready."
        fi
    done
    __go_if_verbose "$info" 1 "Done checking that functions are available."

    # Final check for problems encountered.
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Error(s) encountered while setting up %s functions:\n' "$title" >&2
        printf '  %s\n' "${problems[@]}" >&2
        __go_if_verbose "$error" 1 "There were errors encountered during setup."
        return 3
    fi

    # And done!
    __go_if_verbose "$info" 0 "Setup of $title functions complete."
    return 0
}

GO_SETUP_VERBOSE=
# Usage: __go_if_verbose <level string> <indent-level> <message>
__go_if_verbose () {
    [[ -n "$GO_SETUP_VERBOSE" ]] && printf '%s %b: %s%s\n' "$( date '+%F %T %Z' )" "$1" "$( printf "%$(( $2 * 2 ))s" )" "$3"
    return 0
}

if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
    GO_SETUP_VERBOSE='YES'
fi

# Do what needs to be done.
__go_do_setup "$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
go_setup_exit_code=$?

# Now clean up after yourself.
unset -f __go_do_setup
unset -f __go_if_verbose
unset GO_SETUP_VERBOSE

# Trick to unset a variable but also return it. The string is created first, when the variable exists, then eval executes it.
eval "unset go_setup_exit_code; return $go_setup_exit_code"
