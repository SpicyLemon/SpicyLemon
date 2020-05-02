#!/bin/bash
# This file contains the glclone function that makes it easier to clone one of our repos.
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

__glclone_options_display () {
    echo -E -n '[-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh] [-h|--help] [-p <project name>|--project <project name>] [-s|--select-project]'
}
__glclone_auto_options () {
    echo -E -n "$( __glclone_options_display | __gl_convert_display_options_to_auto_options )"
}
glclone () {
    __gl_require_token || return 1
    local usage
    usage="$( cat << EOF
glclone: GitLab Clone

This will look up all the projects you have access to in GitLab, and provide a way for you to select one or more to clone.

If you set the GITLAB_REPO_DIR environment variable to you root git directory,
new repos will automatically go into that directory regardless of where you are when running the command.
If that variable is not set, and no -b or --base-dir parameter is provided, the current directory is used.

Usage: glclone $( __glclone_options_display )

  The -b <dir> or --base-dir <dir> option will designate the directory to create your repo in.
        Providing this option overrides the default setting from the GITLAB_REPO_DIR.
  The -f or --force option will allow cloning into directories already under a git repo.
  The -r or --refresh option will cause your projects to be reloaded.
  The -p or --project option will allow you to supply the project name you are interested in.
    If the provided project name cannot be found, it will be used as an initial query,
    and you will be prompted to select the project.
    Multiple projects can be provided in the following ways:
        -p project1 -p project2
        -p 'project3 project4'
        -p project5 project6
    Additionallly, the -p or --project option can be omitted, and leftover parameters
    will be treated as the projects you are interested in.
        For example:
            glclone project7 project8
        Is the same as
            glclone -p project7 -p project8
    If no project name is provided after this option, it will be treated the same as -s or --select-projects
  The -s or --select-projects option forces glclone to prompt you to select projects.
      This is only needed if you are supplying projects to clone (with -p or --project),
      but also want to select others.

EOF
    )"
    local destination provided_repos option use_the_force refresh select_repo
    provided_repos=()
    while [[ "$#" -gt 0 ]]; do
        option="$( __gl_lowercase "$1" )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -b|--base-dir)
            __gl_require_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
            destination="$2"
            shift
            ;;
        -f|--force)
            use_the_force="YES"
            ;;
        -r|--refresh)
            refresh="YES"
            ;;
        -p|--project|--projects)
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                provided_repos+=( $2 )
                shift
            else
                select_repo="YES"
            fi
            ;;
        -s|--select-projects)
            select_repo="YES"
            ;;
        -*)
            >&2 echo -E "Unknown option [ $option ]."
            return 2
            ;;
        *)
            provided_repos+=( $1 )
            ;;
        esac
        shift
    done
    local orig_pwd projects selected_repo_count cloned_repo_count repo_url cmd cmd_output new_repo_dir
    if [[ -z "$destination" ]]; then
        if [[ -n "$GITLAB_REPO_DIR" ]]; then
            destination="$GITLAB_REPO_DIR"
        elif [[ -n "$GITLAB_BASE_DIR" ]]; then
            # The GITLAB_BASE_DIR variable is deprecated in favor of GITLAB_REPO_DIR.
            destination="$GITLAB_BASE_DIR"
        fi
    fi
    if [[ -n "$destination" ]]; then
        if [[ ! -d "$destination" ]]; then
            >&2 echo -E "Destination directory [$destination] does not exist."
            return 1
        fi
        orig_pwd="$( pwd )"
        cd "$destination"
    fi
    if [[ -z $use_the_force && $(git rev-parse --is-inside-work-tree 2>/dev/null) ]]; then
        if [[ -n "$destination" ]]; then
            >&2 echo -E "$destination is already inside a git repo. If you'd still like to clone a repo into this directory, use the --force option."
        else
            >&2 echo -E "You are already inside a git repo. If you'd still like to clone a repo into this directory, use the --force option."
        fi
        if [[ -n "$orig_pwd" ]]; then
            cd "$orig_pwd"
        fi
        return 1
    fi

    if [[ -n "$refresh" ]]; then
        GITLAB_USER_ID=''
        GITLAB_USERNAME=''
        __gl_projects_clear_cache
    fi

    __gl_ensure_user_info
    __gl_ensure_projects

    projects="$( __gl_project_subset "$select_repo" '' "$( echo -E "${provided_repos[@]}" )" )"
    selected_repo_count="$( echo -E "$projects" | jq ' length ' )"

    cloned_repo_count=0
    if [[ "$selected_repo_count" -gt '0' ]]; then
        echo -E ""
        for repo_url in $( echo -E "$projects" | jq -r ' .[] | .ssh_url_to_repo ' ); do
            cmd=( git clone --progress "$repo_url" )
            echo -e "\033[1;37m${cmd[@]}\033[0m"
            exec 3>&1
            cmd_output="$( "${cmd[@]}" 2>&1 | tee >( cat - >&3 ) )"
            exec 3>&-
            echo -E ""
            if [[ -z "$( echo "$cmd_output" | grep '^fatal:' )" ]]; then
                new_repo_dir="$( echo -E "$cmd_output" | grep '^Cloning into ' | sed -E "s/^Cloning into '(.+)'\.\.\.[[:space:]]*$/\1/;" )"
                cloned_repo_count=$(( cloned_repo_count + 1 ))
            fi
        done
    fi
    if [[ "$selected_repo_count" -eq '1' && "$cloned_repo_count" -eq '1' && -n "$new_repo_dir" && -d "$new_repo_dir" ]]; then
        cd "$new_repo_dir"
        echo -e "\033[1;37mcd $( pwd )\033[0m"
    elif [[ "$cloned_repo_count" -eq '0' && -n "$orig_pwd" ]]; then
        cd "$orig_pwd"
    else
        echo -e "\033[1;37mcd $( pwd )\033[0m"
    fi
}

return 0
