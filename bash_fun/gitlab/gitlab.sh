#!/bin/bash
# This file contains the main gitlab function that ties all the other functions together.
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

__gitlab_options_display () {
    echo -E -n '(help|merge-requests|mr-search|merged-mrs|ignore-list|clone|open|todo|jobs|clean)'
}
__gitlab_auto_options () {
    echo -E -n "$( __gitlab_options_display | __gl_convert_display_options_to_auto_options )"
}
gitlab () {
    local usage
    usage="$( cat << EOF
gitlab - This is a gateway to all GitLab functions.

Usage:
    gitlab $( __gitlab_options_display ) [command options]

    gitlab help
        Display this message.
        All commands also accept a -h or --help options.

    gitlab merge-requests $( __gmr_options_display_1 )
                          $( __gmr_options_display_2 )
        Get information about merge requests.
        Same as the $( __gl_bold_white "gmr" ) function.

    gitlab mr-search <options>
        Do a search for merge requests with given criteria.
        Same as the $( __gl_bold_white "gmrsearch" ) function.

    gitlab merged-mrs $( __glmerged_options_display )
        Lists MRs that have been merged.
        Same as the $( __gl_bold_white "glmerged" ) function.

    gitlab ignore-list $( __gmrignore_options_display )
        Manage a project ignore list that gmr -d will pay attention to.
        Same as the $( __gl_bold_white "gmrignore" ) function.

    gitlab clone $( __glclone_options_display )
        Easily clone repos from GitLab.
        Same as the $( __gl_bold_white "glclone" ) function.

    gitlab open $( __glopen_options_display_1 )
                $( __glopen_options_display_2 )
        Open various webpages of a GitLab repo.
        Same as the $( __gl_bold_white "glopen" ) function.

    gitlab todo $( __gtd_options_display )
        Get and manage your GitLab todo list.
        Same as the $( __gl_bold_white "gtd" ) function.

    gitlab jobs $( __gljobs_options_display_1 )
                $( __gljobs_options_display_2 )
        Get information about jobs in GitLab.
        Same as the $( __gl_bold_white "gljobs" ) function.

    gitlab clean $( __glclean_options_display )
        Cleans up environment variables storing GitLab information.
        Same as the $( __gl_bold_white "glclean" ) function.

EOF
    )"
    local option cmd
    if [[ -z "$1" ]]; then
        echo -e "$usage"
        return 0
    fi
    option="$( __gl_lowercase "$1" )"
    case "$option" in
    -h|--help|help)
        echo -e "$usage"
        return 0
        ;;
    mrs|merge-requests|merge|prs|pull-requests|pull)
        if [[ "$option" == 'pull' && "$( __gl_lowercase "$2" )" == 'requests' ]]; then
            shift
        fi
        cmd='gmr'
        ;;
    clone)
        cmd='glclone'
        ;;
    todo|todos)
        cmd='gtd'
        ;;
    jobs)
        cmd='gljobs'
        ;;
    clean)
        cmd='glclean'
        ;;
    merged|merged-mrs)
        if [[ "$option" == 'merged' && "$( __gl_lowercase "$2" )" == 'mrs' ]]; then
            shift
        fi
        cmd='glmerged'
        ;;
    open|repo)
        cmd='glopen'
        ;;
    ignore|ignore-list)
        # Allow `gitlab ignore list <command>` to work while also allowing `gitlab ignore list [<state>]` to be the same as `gitlab ignore-list list [<state>]`
        if [[ "$option" == "ignore" && "$( __gl_lowercase "$2" )" == 'list' && -n "$3" ]] && __gmrignore_auto_options | __gl_lowercase | grep -q -w "$( __gl_lowercase "$3" )"; then
            shift
        fi
        cmd='gmrignore'
        ;;
    *)
        >&2 echo -E "Unknown command: [ $1 ]. Use -h for help."
        return 1
        ;;
    esac
    shift
    $cmd "$@"
}
