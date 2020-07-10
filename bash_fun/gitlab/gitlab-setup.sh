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
#   5) Optionally, the following environment variables can be defined.
#       GITLAB_REPO_DIR  ---------------> The directory where your GitLab repositories are to be stored.
#                                         This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                         If not defined, functions that look for it will require it to be provided as input.
#       GITLAB_BASE_DIR  ---------------> This variable has been deprecated in favor of GITLAB_REPO_DIR.
#                                         Please use that variable instead.
#       GITLAB_CONFIG_DIR  -------------> The directory where you'd like to store some configuration information used in these functions.
#                                         This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                         If not defined, then, if HOME is defined, "$HOME/.config/gitlab" will be used.
#                                         If HOME is not defined, then, if GITLAB_REPO_DIR is defined, "$GITLAB_REPO_DIR/.gitlab_config" will be used.
#                                         If GITLAB_REPO_DIR is not defined either, then any functions that uses configuration information will be unavailable.
#                                         If a config dir can be determined, but it doesn't exist yet, it will be created automatically when needed.
#       GITLAB_TEMP_DIR  ---------------> The temporary directory you'd like to use for some random file storage.
#                                         This should be absolute, (starting with a '/'), but it should not end with a '/'.
#                                         If not defined, "/tmp/gitlab" will be used.
#                                         If the directory does not exist, it will be created automatically when needed.
#       GITLAB_CACHE_DEFAULT_MAX_AGE  --> The default max age for cached data.
#                                         Format is <number>[smhdw] where s -> seconds, m -> minutes, h -> hours, d -> days, w -> weeks.
#                                         See `man find` in the -atime section for more info.
#                                         Do not include a leading + or -.
#                                         If not defined, the default is '23h'.
#       GITLAB_PROJECTS_MAX_AGE  -------> The max age that the projects list can be before it's refreshed when needed.
#                                         If not set, GITLAB_CACHE_DEFAULT_MAX_AGE or its default will be used.
#       GITLAB_GROUPS_MAX_AGE  ---------> The max age that the groups list can be before it's refreshed when needed.
#                                         If not set, GITLAB_CACHE_DEFAULT_MAX_AGE or its default will be used.
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source gitlab-setup.sh  in your .bash_profile file.
# If you are running into problems, you can get more information on what's going on by using the -v flag.
#   For example,  source gitlab-setup.sh -v
#
# These programs are required, and don't usually come pre-installed:
#   * jq - Json processor. See https://github.com/stedolan/jq
#   * fzf - Fuzzy finder. See https://github.com/junegunn/fzf
#   * git - The stupid content tracker. https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
#
# These functions are looked for and if not found, the file containing them is the looked for and sourced if possible:
#   * fzf_wrapper - Adds column support to fzf. See https://github.com/SpicyLemon/SpicyLemon/blob/master/bash_fun/fzf_wrapper.sh
#   * curl_link_header - Adds link header processing to curl. See https://github.com/SpicyLemon/SpicyLemon/blob/master/bash_fun/curl_link_header.sh
#
# These programs are required, and are almost always already available:
#   * awk - Pattern scanning and processing.
#   * sed - Stream editor.
#   * grep - Pattern search.
#   * curl - Url transfer utility.
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
# Usage: __gitlab_do_setup "<my source dir>" "<verbose>"
__gitlab_do_setup () {
    local where_i_am
    local gitlab_func_dir gitlab_funcs gitlab_func_file_names
    local files_to_source problems cmd_to_check
    local info ok warn error
    local fzf_wrapper_file_1 fzf_wrapper_file_2 curl_link_header_1 curl_link_header_2 entry exit_code can_auto can_complete can_compctl auto_opts_func
    where_i_am="$1"

    # Look for a gitlab/ directory in the same location that this source file is in.
    gitlab_func_dir="${where_i_am}"

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
    __gitlab_if_verbose "$info" 0 "Loading GitLab functions."

    __gitlab_if_verbose "$info" 1 "Checking on needed external programs and functions."
    if ! __gitlab_i_can jq; then
        problems+=( "Command jq not found. See https://github.com/stedolan/jq for installation instructions." )
        __gitlab_if_verbose "$error" 2 "The jq program was not found."
    else
        __gitlab_if_verbose "$ok" 2 "The jq program is installed."
    fi
    if ! __gitlab_i_can fzf; then
        problems+=( "Command fzf not found. See https://github.com/junegunn/fzf for installation instructions." )
        __gitlab_if_verbose "$error" 2 "The fzf program was not found."
    else
        __gitlab_if_verbose "$ok" 2 "The fzf program is installed."
    fi
    for cmd_to_check in 'awk' 'sed' 'curl' 'grep' 'git'; do
        if ! __gitlab_i_can "$cmd_to_check"; then
            problems+=( "Command $cmd_to_check not found." )
            __gitlab_if_verbose "$error" 2 "The $cmd_to_check command was not found."
        else
            __gitlab_if_verbose "$ok" 2 "The $cmd_to_check command is installed."
        fi
    done
    if ! __gitlab_i_can fzf_wrapper; then
        __gitlab_if_verbose "$warn" 2 "The fzf_wrapper function was not found."
        # See if we can fix that on our own.
        fzf_wrapper_file_1="${where_i_am}/fzf_wrapper.sh"
        fzf_wrapper_file_2="${where_i_am}/../fzf_wrapper.sh"
        if [[ -f "$fzf_wrapper_file_1" ]]; then
            files_to_source+=( "$fzf_wrapper_file_1" )
            __gitlab_if_verbose "$ok" 3 "The file containing the fzf_wrapper function [$fzf_wrapper_file_1] exists and will be sourced."
        elif [[ -f "$fzf_wrapper_file_2" ]]; then
            files_to_source+=( "$fzf_wrapper_file_2" )
            __gitlab_if_verbose "$ok" 3 "The file containing the fzf_wrapper function [$fzf_wrapper_file_2] exists and will be sourced."
        else
            problems+=( "Command fzf_wrapper not found." )
            __gitlab_if_verbose "$error" 3 "The file containing the fzf_wrapper function was not found at either [$fzf_wrapper_file_1] or [$fzf_wrapper_file_2]."
        fi
    else
        __gitlab_if_verbose "$ok" 2 "The fzf_wrapper function is installed."
    fi
    if ! __gitlab_i_can curl_link_header; then
        __gitlab_if_verbose "$warn" 2 "The curl_link_header function was not found."
        # See if we can fix that on our own.
        curl_link_header_1="${where_i_am}/curl_link_header.sh"
        curl_link_header_2="${where_i_am}/../curl_link_header.sh"
        if [[ -f "$curl_link_header_1" ]]; then
            files_to_source+=( "$curl_link_header_1" )
            __gitlab_if_verbose "$ok" 3 "The file containing the curl_link_header function [$curl_link_header_1] exists and will be sourced."
        elif [[ -f "$curl_link_header_2" ]]; then
            files_to_source+=( "$curl_link_header_2" )
            __gitlab_if_verbose "$ok" 3 "The file containing the curl_link_header function [$curl_link_header_2] exists and will be sourced."
        else
            problems+=( "Command fzf_wrapper not found." )
            __gitlab_if_verbose "$error" 3 "The file containing the curl_link_header function was not found at either [$curl_link_header_1] or [$curl_link_header_2]."
        fi
    else
        __gitlab_if_verbose "$ok" 2 "The curl_link_header function is installed."
    fi
    __gitlab_if_verbose "$info" 1 "Done checking on needed external programs and functions."

    __gitlab_if_verbose "$info" 1 "Checking on source files for GitLab functions."
    if [[ ! -d "$gitlab_func_dir" ]]; then
        problems+=( "Directory not found: $gitlab_func_dir" )
        __gitlab_if_verbose "$error" 2 "The GitLab function directory [$gitlab_func_dir] was not found."
    else
        __gitlab_if_verbose "$ok" 2 "The GitLab function directory [$gitlab_func_dir] exists."
        # Make sure all the needed files are there too.
        for entry in "${gitlab_func_file_names[@]}"; do
            func_file="${gitlab_func_dir}/${entry}.sh"
            if [[ ! -f "$func_file" ]]; then
                problems+=( "File not found: $func_file" )
                __gitlab_if_verbose "$error" 2 "The GitLab function file [$func_file] was not found."
            else
                files_to_source+=( "$func_file" )
                __gitlab_if_verbose "$ok" 2 "The GitLab function file [$func_file] exists."
            fi
        done
    fi
    __gitlab_if_verbose "$info" 1 "Done checking on source files for GitLab functions."

    __gitlab_if_verbose "$info" 1 "Checking for problems with environment variables."
    if [[ -n "$GITLAB_PRIVATE_TOKEN" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_PRIVATE_TOKEN environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_PRIVATE_TOKEN" =~ ^[[:space:]]*$ ]]; then
            problems+=( "The GITLAB_PRIVATE_TOKEN environment variable is blank." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_PRIVATE_TOKEN environment variable is blank."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_PRIVATE_TOKEN environment variable is okay."
        fi
    elif [[ -z ${GITLAB_PRIVATE_TOKEN+x} ]]; then
        problems+=( "The GITLAB_PRIVATE_TOKEN environment variable is not defined." )
        __gitlab_if_verbose "$error" 2 "The GITLAB_PRIVATE_TOKEN environment variable is not defined."
    else
        problems+=( "The GITLAB_PRIVATE_TOKEN environment variable does not have a value." )
        __gitlab_if_verbose "$error" 2 "The GITLAB_PRIVATE_TOKEN environment variable is defined, but empty."
    fi
    if [[ -n "$GITLAB_REPO_DIR" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_REPO_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_REPO_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] does not start with a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] does not start with a /."
        elif [[ "$GITLAB_REPO_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] must not end in a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] ends in a /."
        elif [[ ! -d "$GITLAB_REPO_DIR" ]]; then
            problems+=( "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] references a directory that does not exist." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] is a directory that does not exist."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_REPO_DIR environment variable [$GITLAB_REPO_DIR] is okay."
        fi
    elif [[ -n "$GITLAB_BASE_DIR" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_REPO_DIR environment variable does not have a value, but the GITLAB_BASE_DIR environment variable does. Making sure it is okay."
        __gitlab_if_verbose "$warn" 2 "The GITLAB_BASE_DIR environment variable is deprecated. Please set GITLAB_REPO_DIR instead."
        if [[ ! "$GITLAB_BASE_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] does not start with a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] does not start with a /."
        elif [[ "$GITLAB_BASE_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] must not end in a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] ends in a /."
        elif [[ ! -d "$GITLAB_BASE_DIR" ]]; then
            problems+=( "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] references a directory that does not exist." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] is a directory that does not exist."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_BASE_DIR environment variable [$GITLAB_BASE_DIR] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_REPO_DIR environment variable is not set. Some functionality might not be available."
    fi
    if [[ -n "$GITLAB_CONFIG_DIR" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_CONFIG_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_CONFIG_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] does not start with a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] does not start with a /."
        elif [[ "$GITLAB_CONFIG_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] must not end in a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] ends in a /."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_CONFIG_DIR environment variable [$GITLAB_CONFIG_DIR] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_CONFIG_DIR environment variable is not set. A default value will be used if possible, but some functionality might not be available."
    fi
    if [[ -n "$GITLAB_TEMP_DIR" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_TEMP_DIR environment variable has a value. Making sure it is okay."
        if [[ ! "$GITLAB_TEMP_DIR" =~ ^/ ]]; then
            problems+=( "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] does not start with a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] does not start with a /."
        elif [[ "$GITLAB_TEMP_DIR" =~ /$ ]]; then
            problems+=( "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] must not end in a /." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] ends in a /."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_TEMP_DIR environment variable [$GITLAB_TEMP_DIR] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_TEMP_DIR environment variable is not set. A default value will be used."
    fi
    if [[ -n "$GITLAB_CACHE_DEFAULT_MAX_AGE" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_CACHE_DEFAULT_MAX_AGE" =~ ^[+\-] ]]; then
            problems+=( "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable [$GITLAB_CACHE_DEFAULT_MAX_AGE] must not start with a + or -." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable [$GITLAB_CACHE_DEFAULT_MAX_AGE] starts with either a + or -."
        elif [[ "$GITLAB_CACHE_DEFAULT_MAX_AGE" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
            problems+=( "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable [$GITLAB_CACHE_DEFAULT_MAX_AGE] is not in the correct format. See the -atime section in 'man find'." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable [$GITLAB_CACHE_DEFAULT_MAX_AGE] is not in the correct format."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable [$GITLAB_CACHE_DEFAULT_MAX_AGE] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_CACHE_DEFAULT_MAX_AGE environment variable is not set. A default value will be used."
    fi
    if [[ -n "$GITLAB_PROJECTS_MAX_AGE" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_PROJECTS_MAX_AGE environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_PROJECTS_MAX_AGE" =~ ^[+\-] ]]; then
            problems+=( "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] must not start with a + or -." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] starts with either a + or -."
        elif [[ "$GITLAB_PROJECTS_MAX_AGE" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
            problems+=( "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is not in the correct format. See the -atime section in 'man find'." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is not in the correct format."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_PROJECTS_MAX_AGE environment variable [$GITLAB_PROJECTS_MAX_AGE] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_PROJECTS_MAX_AGE environment variable is not set. A default value will be used."
    fi
    if [[ -n "$GITLAB_GROUPS_MAX_AGE" ]]; then
        __gitlab_if_verbose "$info" 2 "The GITLAB_GROUPS_MAX_AGE environment variable has a value. Making sure it is okay."
        if [[ "$GITLAB_GROUPS_MAX_AGE" =~ ^[+\-] ]]; then
            problems+=( "The GITLAB_GROUPS_MAX_AGE environment variable [$GITLAB_GROUPS_MAX_AGE] must not start with a + or -." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_GROUPS_MAX_AGE environment variable [$GITLAB_GROUPS_MAX_AGE] starts with either a + or -."
        elif [[ "$GITLAB_GROUPS_MAX_AGE" =~ ^([[:digit:]]+[smhdw])+$ ]]; then
            problems+=( "The GITLAB_GROUPS_MAX_AGE environment variable [$GITLAB_GROUPS_MAX_AGE] is not in the correct format. See the -atime section in 'man find'." )
            __gitlab_if_verbose "$error" 3 "The GITLAB_GROUPS_MAX_AGE environment variable [$GITLAB_GROUPS_MAX_AGE] is not in the correct format."
        else
            __gitlab_if_verbose "$ok" 3 "The GITLAB_GROUPS_MAX_AGE environment variable [$GITLAB_GROUPS_MAX_AGE] is okay."
        fi
    else
        __gitlab_if_verbose "$warn" 2 "The GITLAB_GROUPS_MAX_AGE environment variable is not set. A default value will be used."
    fi
    __gitlab_if_verbose "$info" 1 "Done checking for problems with environment variables."

    __gitlab_if_verbose "$info" 1 "Checking for problems encountered so far."
    # If there were problems, yo, output them and quit.
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Could not set up GitLab cli functions:\n' >&2
        printf '  %s\n' "${problems[@]}" >&2
        __gitlab_if_verbose "$error" 0 "Quitting early due to problems."
        return 2
    fi
    __gitlab_if_verbose "$ok" 1 "No problems encountered so far."

    __gitlab_if_verbose "$info" 1 "Sourcing all needed files."
    for entry in "${files_to_source[@]}"; do
        __gitlab_if_verbose "$info" 2 "Executing command: source $entry"
        source "$entry"
        exit_code="$?"
        if [[ "$exit_code" -ne '0' ]]; then
            problems+=( "Failed to source the [$entry] file." )
            __gitlab_if_verbose "$error" 3 "The command to source [$entry] failed with an exit code of [$exit_code]."
        else
            __gitlab_if_verbose "$ok" 3 "The source command was successful for [$entry]."
        fi
    done
    __gitlab_if_verbose "$info" 1 "Done sourcing all needed files."

    __gitlab_if_verbose "$info" 1 "Checking that functions are available."
    for entry in "${gitlab_funcs[@]}" 'gitlab'; do
        if ! __gitlab_i_can "$entry"; then
            problems+=( "The [$entry] command failed to load." )
            __gitlab_if_verbose "$error" 2 "The [$entry] command failed to load."
        else
            __gitlab_if_verbose "$ok" 2 "The [$entry] command is loaded and ready."
        fi
    done
    __gitlab_if_verbose "$info" 1 "Done checking that functions are available."

    __gitlab_if_verbose "$info" 1 "Setting up tab completion."
    if __gitlab_i_can complete; then
        can_auto='YES'
        can_complete='YES'
        __gitlab_if_verbose "$ok" 2 "The [complete] tab completion program has been detected and will be used."
    elif __gitlab_i_can compctl; then
        can_auto='YES'
        can_compctl='YES'
        __gitlab_if_verbose "$ok" 2 "The [compctl] tab completion program has been detected and will be used."
    else
        __gitlab_if_verbose "$warn" 2 "Unable to detect tab completion program. Tab complete will not be available for these GitLab functions."
    fi

    if [[ -n "$can_auto" ]]; then
        for entry in "${gitlab_funcs[@]}" 'gitlab'; do
            if __gitlab_i_can "$entry"; then
                auto_opts_func="__${entry}_auto_options"
                if __gitlab_i_can "$auto_opts_func"; then
                    exit_code=
                    if [[ -n "$can_complete" ]]; then
                        __gitlab_if_verbose "$info" 2 "Executing command: complete -W \"\$( $auto_opts_func )\" $entry"
                        complete -W "$( $auto_opts_func )" $entry
                        exit_code="$?"
                    elif [[ -n "$can_compctl" ]]; then
                        if [[ "$entry" == 'gitlab' ]]; then
                            __gitlab_if_verbose "$info" 2 "Executing command: compctl -x 'p[1]' -k \"( \$( $auto_opts_func ) )\" -- $entry"
                            compctl -x 'p[1]' -k "( $( $auto_opts_func ) )" -- $entry
                            exit_code="$?"
                        else
                            __gitlab_if_verbose "$info" 2 "Executing command: compctl -k \"( \$( $auto_opts_func ) )\" $entry"
                            compctl -k "( $( $auto_opts_func ) )" $entry
                            exit_code="$?"
                        fi
                    else
                        problems+=( "Unknown tab complete program. Cannot set up tab complete for [$entry]." )
                        __gitlab_if_verbose "$error" 2 "The tab completion program is not known. Tab completion unavailable for [$entry]."
                    fi
                    if [[ -n "$exit_code" ]]; then
                        if [[ "$exit_code" -ne '0' ]]; then
                            problems+=( "Tab completion setup failed for [$entry]." )
                            __gitlab_if_verbose "$error" 3 "The command to set up tab completion for [$entry] failed with an exit code of [$exit_code]."
                        else
                            __gitlab_if_verbose "$ok" 3 "Tab completion set up for [$entry]."
                        fi
                    fi
                else
                    __gitlab_if_verbose "$warn" 2 "The [$auto_opts_func] function was not found. Tab completion unavailable for [$entry]."
                fi
            fi
        done
    fi
    __gitlab_if_verbose "$info" 1 "Done setting up tab completion."

    __gitlab_if_verbose "$info" 1 "Doing final checking for problems encountered."
    if [[ "${#problems[@]}" -gt '0' ]]; then
        printf 'Error(s) encountered while setting up GitLab cli functions:' >&2
        printf '  %s\n' "${problems[@]}" >&2
        __gitlab_if_verbose "$error" 0 "There were errors encountered during setup."
        return 3
    fi
    __gitlab_if_verbose "$ok" 1 "No errors encountered during setup."

    __gitlab_if_verbose "$info" 0 "Setup of GitLab functions complete."
    return 0
}

# Tests if a command is available.
# Usage: if __gitlab_i_can "foo"; then echo "I can totally foo"; else echo "There's no way I can foo."; fi
__gitlab_i_can () {
    if [[ "$#" -eq '0' ]]; then
        return 1
    fi
    command -v "$@" > /dev/null 2>&1
}

GITLAB_SETUP_VERBOSE=
# Usage: __gitlab_if_verbose <level string> <indent-level> <message>
__gitlab_if_verbose () {
    [[ -n "$GITLAB_SETUP_VERBOSE" ]] && printf '%s %b: %s%s\n' "$( date '+%F %T %Z' )" "$1" "$( printf "%$(( $2 * 2 ))s" )" "$3"
}

if [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
    GITLAB_SETUP_VERBOSE='YES'
fi

# Do what needs to be done.
__gitlab_do_setup "$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"

# Now clean up after yourself.
unset -f __gitlab_do_setup
unset -f __gitlab_i_can
unset -f __gitlab_if_verbose
unset GITLAB_SETUP_VERBOSE

return 0
