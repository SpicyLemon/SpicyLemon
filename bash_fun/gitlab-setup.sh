#!/bin/bash
# This file is meant to be sourced.
# It will import all the functions needed to interact with GitLab in your terminal.
#
# In order to use any of these functions, you will first have to create a GitLab private token.
#   1) Log into GitLab.
#   2) Go to your personal settings page and to the "Access Tokens" page (e.g https://gitlab.com/profile/personal_access_tokens )
#   3) Create a token with the 'api' scope.
#   4) Set the GITLAB_PRIVATE_TOKEN environment variable to the value of that token.
#       For example, you could put   GITLAB_PRIVATE_TOKEN=123abcABC456-98ZzYy7  in your .bash_profile file
#       so that it's set every time you open a terminal (use your own actual token of course).
#   5) Optionally, the following optional environment variables can be defined.
#       GITLAB_REPO_DIR  ----------> The directory where your GitLab repositories are to be stored.
#                                    This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                    If not defined, functions that look for it will require it to be provided as input.
#       GITLAB_BASE_DIR  ----------> This variable has been deprecated in favor of GITLAB_REPO_DIR.
#                                    Please use that variable instead.
#       GITLAB_CONFIG_DIR  --------> The directory where you'd like to store some configuration information used in these functions.
#                                    This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                    If not defined, then, if HOME is defined, "$HOME/.config/gitlab" will be used.
#                                    If HOME is not defined, then, if GITLAB_REPO_DIR is defined, "$GITLAB_REPO_DIR/.gitlab_config" will be used.
#                                    If GITLAB_REPO_DIR is not defined either, then any functions that uses configuration information will be unavailable.
#                                    If a config dir can be determined, but it doesn't exist yet, it will be created automatically when needed.
#       GITLAB_TEMP_DIR  ----------> The temporary directory you'd like to use for some random file storage.
#                                    This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                    If not defined, "/tmp/gitlab" will be used.
#                                    If the directory does not exist, it will be created automatically when needed.
#       GITLAB_PROJECTS_MAX_AGE  --> The max age that the projects list can be before it's refreshed when needed.
#                                    Format is <number>[smhdw] where s -> seconds, m -> minutes, h -> hours, d -> days, w -> weeks.
#                                    see `man find` in the -atime section for more info.
#                                    Do not include a leading + or -.
#                                    If not defined, the default is '23h'.
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source gitlab-setup.sh  in your .bash_profile file.
# If you are running into problems, you can get more information on what's going on by using the -v flag.
#   For example,  source gitlab-setup.sh -v
#
# Lastly, these functions rely on the following programs (that you might not have installed yet):
#   * jq - Command-line JSON processor - https://github.com/stedolan/jq
#   * fzf - Command-line fuzzy finder - https://github.com/junegunn/fzf
#   * fzf_wrapper - A wrapper for fzf that adds a the --to-columns option. It's defined in the fzf_wrapper.sh file in this repo.
# And these, that you probably do have installed:
#   * awk - Pattern-Directed Scanning and Processing Language
#   * sed - Stream Editor
#   * curl - Transfer a URL
#   * grep - File Pattern Searcher
#   * git - The Stupid Content Tracker
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

# Putting all the setup stuff in a functions so that there aren't left-over variables and stuff.
# Usage: __gitlab_setup "<my source dir>" "<verbose>"
__gitlab_setup () {
    local where_i_am
    local gitlab_func_dir gitlab_funcs gitlab_func_file_names
    local files_to_source problems cmd_to_check
    local info ok warn error
    local fzf_wrapper_file entry exit_code can_auto can_complete can_compctl auto_opts_func
    where_i_am="$1"

    # Look for a gitlab/ directory in the same location that this source file is in.
    gitlab_func_dir="${where_i_am}/gitlab"

    # These are the available gitlab functions excluding the main  gitlab  one that pulls them all together.
    gitlab_funcs=(
        'gmr' 'gmrsearch' 'glmerged' 'gmrignore'
        'glclone' 'glopen' 'gtd' 'gljobs' 'glclean'
    )

    # These are the base file names containing the functions needed.
    gitlab_func_file_names=( 'gl-core' )
    gitlab_func_file_names+=( "${gitlab_funcs[@]}" )
    gitlab_func_file_names+=( 'gitlab' )

    # This will hold the full path to the files that need to be sourced.
    files_to_source=()

    # This will hold any error messages that should be displayed.
    problems=()

    # These are used for verbose output as line headers.
       ok="        \033[1;32m [ OK ] \033[0m"
     info="    \033[1;21m [ INFO ] \033[0m  "
     warn="  \033[1;33m [ WARN ] \033[0m    "
    error="\033[1;41m [ ERROR ] \033[0m     "

    # And, let's get started!
    __if_verbose "$info" "Checking on needed external programs and functions."

    # Make sure jq is available.
    if ! __i_can jq; then
        problems+=( "Command jq not found. See https://github.com/stedolan/jq for installation instructions." )
        __if_verbose "$error" "The jq program was not found."
    else
        __if_verbose "$ok" "The jq program is installed."
    fi

    # Make sure fzf is available.
    if ! __i_can fzf; then
        problems+=( "Command fzf not found. See https://github.com/junegunn/fzf for installation instructions." )
        __if_verbose "$error" "The fzf program was not found."
    else
        __if_verbose "$ok" "The fzf program is installed."
    fi

    # Make sure we can awk, sed, curl, and git
    for cmd_to_check in 'awk' 'sed' 'curl' 'grep' 'git'; do
        if ! __i_can "$cmd_to_check"; then
            problems+=( "Command $cmd_to_check not found." )
            __if_verbose "$error" "The $cmd_to_check command was not found."
        else
            __if_verbose "$ok" "The $cmd_to_check command is installed."
        fi
    done

    # Make sure fzf_wrapper is available.
    if ! __i_can fzf_wrapper; then
        __if_verbose "$warn" "The fzf_wrapper function was not found."
        # See if we can fix that on our own.
        fzf_wrapper_file="${where_i_am}/fzf_wrapper.sh"
        if [[ -f "$fzf_wrapper_file" ]]; then
            files_to_source+=( "$fzf_wrapper_file" )
            __if_verbose "$ok" "The file containing the fzf_wrapper function [$fzf_wrapper_file] exists and will be sourced."
        else
            problems+=( "Command fzf_wrapper not found." )
            __if_verbose "$error" "The file containing the fzf_wrapper function [$fzf_wrapper_file] was not found either."
        fi
    else
        __if_verbose "$ok" "The fzf_wrapper function has already been created."
    fi

    __if_verbose "$info" "Done checking on needed external programs and functions."
    __if_verbose "$info" "Checking on source files for GitLab functions."

    # Make sure the directory with the functions is there.
    if [[ ! -d "$gitlab_func_dir" ]]; then
        problems+=( "Directory not found: $gitlab_func_dir" )
        __if_verbose "$error" "The GitLab function directory [$gitlab_func_dir] was not found."
    else
        __if_verbose "$ok" "The GitLab function directory [$gitlab_func_dir] exists."
        # Make sure all the needed files are there too.
        for entry in "${gitlab_func_file_names[@]}"; do
            func_file="${gitlab_func_dir}/${entry}.sh"
            if [[ ! -f "$func_file" ]]; then
                problems+=( "File not found: $func_file" )
                __if_verbose "$error" "The GitLab function file [$func_file] was not found."
            else
                files_to_source+=( "$func_file" )
                __if_verbose "$ok" "The GitLab function file [$func_file] exists."
            fi
        done
    fi

    __if_verbose "$info" "Done checking on source files for GitLab functions."
    __if_verbose "$info" "Checking for problems with environment variables."

    if [[ -n "$GITLAB_PRIVATE_TOKEN" ]]; then
        __if_verbose "$info" "The GITLAB_PRIVATE_TOKEN environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_PRIVATE_TOKEN" =~ ^[[:space:]]*$ ]]; then
            problems+=( "The GITLAB_PRIVATE_TOKEN environment variable is blank." )
            __if_verbose "$error" "The GITLAB_PRIVATE_TOKEN environment variable is blank."
        else
            __if_verbose "$ok" "The GITLAB_PRIVATE_TOKEN environment variable is okay."
        fi
    elif [[ -z ${GITLAB_PRIVATE_TOKEN+x} ]]; then
        problems+=( "The GITLAB_PRIVATE_TOKEN environment variable is not defined." )
        __if_verbose "$error" "The GITLAB_PRIVATE_TOKEN environment variable is not defined."
    else
        problems+=( "The GITLAB_PRIVATE_TOKEN environment variable does not have a value." )
        __if_verbose "$error" "The GITLAB_PRIVATE_TOKEN environment variable is defined, but empty."
    fi

    if [[ -n "$GITLAB_REPO_DIR" ]]; then
        __if_verbose "$info" "The GITLAB_REPO_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_REPO_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] does not start with a /." )
            __if_verbose "$error" "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] does not start with a /."
        elif [[ "$GITLAB_REPO_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] must not end in a /." )
            __if_verbose "$error" "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] ends in a /."
        elif [[ ! -d "$GITLAB_REPO_DIR" ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] references a directory that does not exist." )
            __if_verbose "$error" "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] is a directory that does not exist."
        else
            __if_verbose "$ok" "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] is okay."
        fi
    elif [[ -n "$GITLAB_BASE_DIR" ]]; then
        __if_verbose "$info" "The GITLAB_REPO_DIR environment variable does not have a value, but the GITLAB_BASE_DIR environment variable does. Making sure it is okay."
        __if_verbose "$warn" "The GITLAB_BASE_DIR environment variable is deprecated. Please set GITLAB_REPO_DIR instead."
        if [[ ! "$GITLAB_BASE_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] does not start with a /." )
            __if_verbose "$error" "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] does not start with a /."
        elif [[ "$GITLAB_BASE_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] must not end in a /." )
            __if_verbose "$error" "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] ends in a /."
        elif [[ ! -d "$GITLAB_BASE_DIR" ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] references a directory that does not exist." )
            __if_verbose "$error" "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] is a directory that does not exist."
        else
            __if_verbose "$ok" "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] is okay."
        fi
    else
        __if_verbose "$warn" "The GITLAB_REPO_DIR environment variable is not set. Some functionality might not be available."
    fi

    if [[ -n "$GITLAB_CONFIG_DIR" ]]; then
        __if_verbose "$info" "The GITLAB_CONFIG_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_CONFIG_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] does not start with a /." )
            __if_verbose "$error" "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] does not start with a /."
        elif [[ "$GITLAB_CONFIG_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] must not end in a /." )
            __if_verbose "$error" "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] ends in a /."
        else
            __if_verbose "$ok" "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] is okay."
        fi
    else
        __if_verbose "$warn" "The GITLAB_CONFIG_DIR environment variable is not set. A default value will be used if possible, but some functionality might not be available."
    fi

    if [[ -n "$GITLAB_TEMP_DIR" ]]; then
        __if_verbose "$info" "The GITLAB_TEMP_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_TEMP_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] does not start with a /." )
            __if_verbose "$error" "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] does not start with a /."
        elif [[ "$GITLAB_TEMP_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] must not end in a /." )
            __if_verbose "$error" "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] ends in a /."
        else
            __if_verbose "$ok" "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] is okay."
        fi
    else
        __if_verbose "$warn" "The GITLAB_TEMP_DIR environment variable is not set. A default value will be used."
    fi

    if [[ -n "$GITLAB_PROJECTS_MAX_AGE" ]]; then
        __if_verbose "$info" "The GITLAB_PROJECTS_MAX_AGE environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_PROJECTS_MAX_AGE" =~ ^[+\-] ]]; then
            problems+=( "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] must not start with a + or -." )
            __if_verbose "$error" "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] starts with either a + or -."
        elif [[ "$GITLAB_PROJECTS_MAX_AGE" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
            problems+=( "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is not in the correct format. See the -atime section in 'man find'." )
            __if_verbose "$error" "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is not in the correct format."
        else
            __if_verbose "$ok" "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is okay."
        fi
    else
        __if_verbose "$warn" "The GITLAB_PROJECTS_MAX_AGE environment variable is not set. A default value will be used."
    fi

    __if_verbose "$info" "Done checking for problems with environment variables."
    __if_verbose "$info" "Checking for problems encountered so far."

    # If there were problems, yo, output them and quit.
    if [[ "${#problems[@]}" -gt '0' ]]; then
        >&2 echo -E "Could not set up GitLab cli functions:"
        for entry in "${problems[@]}"; do
            >&2 echo -E "  $entry"
        done
        __if_verbose "$error" "Quitting early due to problems."
        return 2
    fi

    __if_verbose "$ok" "No problems encountered so far."
    __if_verbose "$info" "Sourcing all needed files."

    # Okay, fine, source it all!
    for entry in "${files_to_source[@]}"; do
        __if_verbose "$info" "Executing command: source $entry"
        source "$entry"
        exit_code="$?"
        if [[ "$exit_code" -ne '0' ]]; then
            problems+=( "Failed to source the [$entry] file." )
            __if_verbose "$error" "The command to source [$entry] failed with an exit code of [$exit_code]."
        else
            __if_verbose "$ok" "The source command was successful for [$entry]."
        fi
    done

    __if_verbose "$info" "Done sourcing all needed files."
    __if_verbose "$info" "Checking that functions are available."

    for entry in "${gitlab_funcs[@]}" 'gitlab'; do
        if ! __i_can "$entry"; then
            problems+=( "The [$entry] command failed to load." )
            __if_verbose "$error" "The [$entry] command failed to load."
        else
            __if_verbose "$ok" "The [$entry] command is loaded and ready."
        fi
    done

    __if_verbose "$info" "Done checking that functions are available."
    __if_verbose "$info" "Setting up tab completion."

    if __i_can complete; then
        can_auto='YES'
        can_complete='YES'
        __if_verbose "$ok" "The [complete] tab completion program has been detected and will be used."
    elif __i_can compctl; then
        can_auto='YES'
        can_compctl='YES'
        __if_verbose "$ok" "The [compctl] tab completion program has been detected and will be used."
    else
        __if_verbose "$warn" "Unable to detect tab completion program. Tab complete will not be available for these GitLab functions."
    fi

    if [[ -n "$can_auto" ]]; then
        for entry in "${gitlab_funcs[@]}" 'gitlab'; do
            if __i_can "$entry"; then
                auto_opts_func="__${entry}_auto_options"
                if __i_can "$auto_opts_func"; then
                    exit_code=
                    if [[ -n "$can_complete" ]]; then
                        __if_verbose "$info" "Executing command: complete -W \"\$( $auto_opts_func )\" $entry"
                        complete -W "$( $auto_opts_func )" $entry
                        exit_code="$?"
                    elif [[ -n "$can_compctl" ]]; then
                        if [[ "$entry" == 'gitlab' ]]; then
                            __if_verbose "$info" "Executing command: compctl -x 'p[1]' -k \"( \$( $auto_opts_func ) )\" -- $entry"
                            compctl -x 'p[1]' -k "( $( $auto_opts_func ) )" -- $entry
                            exit_code="$?"
                        else
                            __if_verbose "$info" "Executing command: compctl -k \"( \$( $auto_opts_func ) )\" $entry"
                            compctl -k "( $( $auto_opts_func ) )" $entry
                            exit_code="$?"
                        fi
                    else
                        problems+=( "Unknown tab complete program. Cannot set up tab complete for [$entry]." )
                        __if_verbose "$error" "The tab completion program is not known. Tab completion unavailable for [$entry]."
                    fi
                    if [[ -n "$exit_code" ]]; then
                        if [[ "$exit_code" -ne '0' ]]; then
                            problems+=( "Tab completion setup failed for [$entry]." )
                            __if_verbose "$error" "The command to set up tab completion for [$entry] failed with an exit code of [$exit_code]."
                        else
                            __if_verbose "$ok" "Tab completion set up for [$entry]."
                        fi
                    fi
                else
                    __if_verbose "$warn" "The [$auto_opts_func] function was not found. Tab completion unavailable for [$entry]."
                fi
            fi
        done
    fi

    __if_verbose "$info" "Done setting up tab completion."
    __if_verbose "$info" "Doing final checking for problems encountered."

    if [[ "${#problems[@]}" -gt '0' ]]; then
        >&2 echo -E "Error(s) encountered while setting up GitLab cli functions:"
        for entry in "${problems[@]}"; do
            >&2 echo -E "  $entry"
        done
        __if_verbose "$error" "There were errors encountered during setup."
        return 3
    fi

    __if_verbose "$ok" "No errors encountered during setup."
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

GITLAB_SETUP_VERBOSE=
# Usage: __if_verbose <level string> <message>
__if_verbose () {
    [[ -n "$GITLAB_SETUP_VERBOSE" ]] && echo -e "$( date '+%F %T %Z' ) $1: $2"
}

if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
    GITLAB_SETUP_VERBOSE='YES'
fi

# Do what needs to be done.
__gitlab_setup "$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

# Now clean up after yourself.
unset -f __gitlab_setup
unset -f __i_can
unset -f __if_verbose
unset GITLAB_SETUP_VERBOSE

return 0
