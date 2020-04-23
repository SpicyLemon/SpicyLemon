#!/bin/bash
# This file contains the glclean function that can be used to clean up some environment variables used by various gitlab functions.
# See ../gitlab-setup.sh for setup information.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained within.
$( echo -e "\033[1;37msource $( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )\033[0m" )
EOF
    exit 1
fi
unset sourced

__glclean_options_display () {
    echo -E -n '[-v|--verbose] [-l|--list] [-h|--help]'
}
__glclean_auto_options () {
    echo -E -n "$( __glclean_options_display | __convert_display_options_to_auto_options )"
}
glclean () {
    local vars_to_clean vars_str usage
    vars_to_clean=("GITLAB_USER_INFO" "GITLAB_USER_ID" "GITLAB_USERNAME" "GITLAB_PROJECTS" "GITLAB_MRS"
                   "GITLAB_MRS_TODO" "GITLAB_MRS_BY_ME" "GITLAB_TODOS" "GITLAB_JOBS" "GITLAB_MERGED_MRS"
                   "GITLAB_MERGED_MRS_REPO" "GITLAB_MRS_SEARCH_RESULTS" "GITLAB_MRS_DEEP_RESULTS")
    vars_str="$( echo -E "${vars_to_clean[*]}" | sed -E 's/ /~/g; s/([^~]+~[^~]+~[^~]+~[^~]+)~/\1\\n/g;' )"
    vars_str="$( echo -e "$vars_str" | column -s '~' -t | sed 's/^/    /' )"
    usage="$( cat << EOF
glclean: GitLab Clean

Cleans up all the persistant variables used by the functions in this file.
Use this when you want a fresh start with respects to the data these GitLab functions use.

This will NOT affect your GITLAB_PRIVATE_TOKEN variable.

The following variables will be removed:
$( echo -e "$vars_str" )

Usage: glclean $( __glclean_options_display )

  The -v or --verbose option will output the values of each variable before being deleted.
  The -l or --list option will just show the variable names without deleting them.
    Combined with the -v command, the contents of the variables will also be displayed.

EOF
    )"
    local option verbose just_show v
    while [[ "$#" -gt 0 ]]; do
        option="$( __to_lowercase "$1" )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -v|--verbose)
            verbose="YES"
            ;;
        -l|--list)
            just_show="YES"
            ;;
        *)
            >&2 echo -E "Unknown option: [ $option ]."
            >&2 echo -e "$usage"
            return 1
            ;;
        esac
        shift
    done
    if [[ -z "$just_show" ]]; then
        __delete_projects_file
        if [[ -n "$verbose" ]]; then
            echo "Deleted projects file."
        fi
    fi
    for v in ${vars_to_clean[@]}; do
        if [[ -n "$verbose" ]]; then
            if [[ -n "$( ps -o command= $$ | grep -E "zsh$" )" ]]; then
                echo -E "$v=${(P)v}"
            else
                echo -E "$v=${!v}"
            fi
        elif [[ -n "$just_show" ]]; then
            echo -E "$v"
        fi
        if [[ -z "$just_show" ]]; then
            unset $v
        fi
    done
    if [[ -z "$just_show" ]]; then
        echo -E "GitLab-associated variables cleaned."
    fi
}
