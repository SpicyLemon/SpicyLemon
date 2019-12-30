#!/bin/bash
# This script creates some functions that are handing for interacting with GitLab.
# File contents:
#   gitlab  ----> Wrapper function for more easily accessing all the other functionality in here.
#   gmr  -------> GitLab Merge Requests - For finding merge requests that merit your attention.
#   glclone  ---> GitLab Clone - Easily find and clone a repo right from your terminal.
#   gtd  -------> GitLab ToDo - Get and open your ToD list and items on it.
#   gljobs  ----> GitLab Jobs - Get info about jobs that have run (or are running).
#   glclean  ---> GitLab Clean - Cleans up all the environment variables used by these functions.
#   glmerged  --> GitLab Merged - Gets a list of merged MRs in merge order for a repo.
#   glopen  ----> GitLab Open - Open a repo main page.
#
# In order to use any of these functions, you will first have to create a GitLab private token.
#   1) Log into GitLab.
#   2) Go to your personal settings page and to the "Access Tokens" page (e.g https://gitlab.com/profile/personal_access_tokens )
#   3) Create a token with the 'api' scope.
#   4) Set the GITLAB_PRIVATE_TOKEN environment variable to the value of that token.
#       For example, you could put   GITLAB_PRIVATE_TOKEN=123abcABC456-98ZzYy7  in your .bash_profile file
#       so that it's set every time you open a terminal (use your own actual token of course).
#   5) Optionally, you can also set the GITLAB_BASE_DIR to your base git directory to help facilitate cloning.
#       Example:  GITLAB_BASE_DIR="$HOME/git"
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source gitlab.sh  in your .bash_profile file.
#
# NOTE: The functions in here rely on the following programs (that you might not have installed yet):
#   * fzf - Command-line fuzzy finder - https://github.com/junegunn/fzf
#   * jq - Command-line JSON processor - https://github.com/stedolan/jq
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Define the location where some files are used to store info.
# This way, some stuff can be shared between terminals/environments.
GITLAB_TEMP_DIR='/tmp/gitlab'
# Define the max age of the projects.
# If the projects have not been looked up recently, they will be looked up again.
# <number>[smhdw] where s -> seconds, m -> minutes, h -> hours, d -> days, w -> weeks
# see `man find` in the -atime section for more info.
GITLAB_PROJECTS_MAX_AGE='1d'

__gitlab_options_display () {
    echo -E -n '(help|merge-requests|clone|todo|jobs|clean|merged-mrs|open)'
}
__gitlab_auto_options () {
    echo -E -n "$( __gitlab_options_display | __convert_display_options_to_auto_options )"
}
gitlab () {
    local usage
    usage="$( cat << EOF
gitlab - This is a gateway to all GitLab functions.

Usage:
    gitlab $( __gitlab_options_display ) [command options]

    gitlab help
        Display this message.

    gitlab merge-requests $( __gmr_options_display )
        Get information about merge requests.
        Same as the $( __highlight "gmr" ) function.

    gitlab clone $( __glclone_options_display )
        Easily clone repos from GitLab.
        Same as the $( __highlight "glclone" ) function.

    gitlab todo $( __gtd_options_display )
        Get and manage your GitLab todo list.
        Same as the $( __highlight "gtd" ) function.

    gitlab jobs $( __gljobs_options_display_1 )
                $( __gljobs_options_display_2 )
        Get information about jobs in GitLab.
        Same as the $( __highlight "gljobs" ) function.

    gitlab clean $( __glclean_options_display )
        Cleans up environment variables storing GitLab information.
        Same as the $( __highlight "glclean" ) function.

    gitlab merged-mrs $( __glmerged_options_display )
        Lists MRs that have been merged.
        Same as the $( __highlight "glmerged" ) function.

    gitlab open $( __glopen_options_display_1 )
                $( __glopen_options_display_2 )
        Open the hompeage of a GitLab repo.
        Same as the $( __highlight "glopen" ) function.

EOF
    )"
    local cmd
    if [[ -z "$1" ]]; then
        echo -e "$usage"
        return 0
    fi
    case "$1" in
    -h|--help|help)
        echo -e "$usage"
        return 0
        ;;
    mrs|merge-requests|merge|prs|pull-requests|pull)
        if [[ "$2" == 'requests' ]]; then
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
        if [[ "$2" == 'mrs' ]]; then
            shift
        fi
        cmd='glmerged'
        ;;
    open|repo)
        cmd='glopen'
        ;;
    *)
        >&2 echo -E "Unknown command: [ $1 ]. Use -h for help."
        return 1
        ;;
    esac
    shift
    $cmd "$@"
}

__gmr_options_display () {
    echo -E -n '[-s|--select] [-r|--refresh] [-d|--deep] [-u|--update] [-q|--quiet] [-m|--mine] [-o|--open-all] [-h|--help]'
}
__gmr_auto_options () {
    echo -E -n "$( __gmr_options_display | __convert_display_options_to_auto_options )"
}
gmr () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
gmr: GitLab Merge Requests

Gets information about merge requests you are involved in.

Usage: gmr $( __gmr_options_display )

  With no options, if there is no previous results, new results are looked up.
  With no options, if there ARE previous results, those old results are displayed.
  The -s or --select option makes gmr prompt you to select entries that will be opened in your browser.
        You can select multiple entries using the tab key. All selected entries will be opened in your browser.
  The -r or --refresh option causes gmr to reach out to GitLab to get a current list of your MRs (the easy, but incomplete way).
  The -d or --deep option causes gmr to go through each project you can see to check for merge requests that request your approval.
        This will take longer, but might uncover some MRs that do not show up with the simple (-r) lookup.
        If supplied with the -r option, the -r option is ignored.
  The -u or --update option causes gmr to go through the known lists of MRs and remove them if you have approved them.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -m or --mine option lists MRs that you created.
  The -o or --open-all option causes all MRs to be opened in your browser.

Basically, the first time you run the  gmr  command, you will get a list of MRs (eventually).
After that (in the same terminal) running  gmr  again will display the previous results.
In order to update the list again, do a  gmr --refresh

EOF
    )"
    local option do_refresh do_update do_deep do_mine do_selector keep_quiet open_all refresh_type filter_type discussion_type mrs todo_count
    while [[ "$#" -gt 0 ]]; do
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -r|--refresh)
            do_refresh="YES"
            ;;
        -u|--update)
            do_update="YES"
            ;;
        -d|--deep)
            do_deep="YES"
            ;;
        -m|--mine)
            do_mine="YES"
            ;;
        -s|--select)
            do_selector="YES"
            ;;
        -q|--quiet)
            keep_quiet="YES"
            ;;
        -o|--open-all)
            open_all="YES"
            ;;
        *)
            >&2 echo -E "Unknown option [ $option ]."
            >&2 echo -e "$usage"
            return 2
            ;;
        esac
        shift
    done
    if [[ -n "$do_deep" && -n "$do_mine" ]]; then
        >&2 echo -E "--deep and --mine are mutually exclusive options. Please only supply one of them."
        >&2 echo -e "$usage"
        return 0
    fi
    if [[ -n "$do_refresh" && "$do_update" ]]; then
        >&2 echo -E "The --refresh option overrides the --update option; --update is being ignored."
    fi
    if [[ -n "$do_update" && "$do_mine" ]]; then
        >&2 echo -E "The --update option is not applicable to the --mine option. You probably want --refresh instead."
    fi

    if [[ -n "$do_deep" ]]; then
        refresh_type="DEEP"
        filter_type="STANDARD"
        discussion_type="STANDARD"
    elif [[ -n "$do_mine" ]]; then
        if [[ -n "$do_refresh" || -z "$GITLAB_MRS_BY_ME" ]]; then
            refresh_type="MINE"
            discussion_type="MINE"
        fi
    elif [[ -n "$do_refresh" || -z "$GITLAB_MRS" ]]; then
        refresh_type="STANDARD"
        filter_type="STANDARD"
        discussion_type="STANDARD"
    fi

    if [[ -n "$do_update" && -z "$do_mine" && -z "$do_refresh" ]]; then
        filter_type="SHORT"
        discussion_type="STANDARD"
    fi

    __ensure_gitlab_user_info "$keep_quiet"
    __ensure_gitlab_projects "$keep_quiet"

    if [[ -n "$refresh_type" ]]; then
        case "$refresh_type" in
        "DEEP")
            __get_my_gitlab_mrs_deep "$keep_quiet"
            ;;
        "MINE")
            __get_gitlab_mrs_i_created "$keep_quiet"
            ;;
        *)
            __get_my_gitlab_mrs "$keep_quiet"
            ;;
        esac
    fi

    if [[ -n $filter_type ]]; then
        __filter_gitlab_mrs "$keep_quiet" "$filter_type"
    fi

    if [[ -n "$discussion_type" ]]; then
        __add_discussion_info_to_mrs "$keep_quiet" "$discussion_type"
    fi

    mrs="$( if [[ -n "$do_mine" ]]; then echo -E "$GITLAB_MRS_BY_ME"; else echo -E "$GITLAB_MRS_TODO"; fi )"
    todo_count=$( echo -E "$mrs" | jq ' length ' )
    if [[ $todo_count -eq 0 ]]; then
        [[ -n "$keep_quiet" ]] || echo -E "You have no MRs$( if [[ -z "$do_mine" ]]; then echo -E " to review!!"; else echo -E "."; fi )"
    else
        if [[ -z "$keep_quiet" ]]; then
            echo -E "You have $todo_count MRs$( [[ -z "$do_mine" ]] && echo -E " to review (oldest on top)" )."
            ( echo -E '┌───▪ Repo~┌───▪ Author~┌───▪ Discussions~┌───▪ Title~┌───▪ Url' \
                && echo -E "$mrs" \
                    | jq -r --arg box_checked '☑' --arg box_empty '☐' \
                        ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          .[] | .col_head = "├─" + ( if .approved == true then $box_checked else $box_empty end) + " "
                              |         .col_head + ( .project_name | clean )
                                + "~" + .col_head + ( .author.name | cleanname | .[0:20] )
                                + "~" + .col_head + .discussion_stats
                                + "~" + .col_head + ( .title | .[0:35] | clean )
                                + "~" + .col_head + .web_url ' ) \
                | sed '$s/├/└/g' \
                | column -s '~' -t
        fi
        if [[ -n "$open_all" ]]; then
            for web_url in $( echo -E "$mrs" | jq -r ' .[] | .web_url ' ); do
                open "$web_url"
            done
        fi
        if [[ -n $do_selector ]]; then
            local selected_lines selected_line web_url
            selected_lines="$( ( echo -E " ~ Repo~ Author~ Discussions~ Title$( [[ -z "$do_mine" ]] && echo -E " (oldest on top)" )" \
                && echo -E "$mrs" \
                    | jq -r --arg box_checked '☑' --arg box_empty '☐' \
                        ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          .[] |          ( if .approved == true then $box_checked else $box_empty end)
                                + " ~" + ( .project_name | clean )
                                + " ~" + ( .author.name | cleanname )
                                + " ~" + .discussion_stats
                                + " ~" + ( .title | .[0:80] | clean )
                                + " ~TARGET_URL>" + .web_url + "<" ' ) \
                | column -s '~' -t \
                | fzf --tac --header-lines=1 --cycle --with-nth=1,2,3,4,5 --delimiter="   +" -m )"
            echo -E "$selected_lines" | while read selected_line; do
                web_url=$( echo -E "$selected_line" | grep -E -o 'TARGET_URL>[^[:space:]]+<' | sed -E 's/^TARGET_URL>|<$//g' )
                if [[ -n $web_url ]]; then
                    open "$web_url"
                fi
            done
        fi
    fi
}

__glclone_options_display () {
    echo -E -n '[-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh] [-h|--help] [-p <project name>|--project <project name>] [-s|--select-project]'
}
__glclone_auto_options () {
    echo -E -n "$( __glclone_options_display | __convert_display_options_to_auto_options )"
}
glclone () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
glclone: GitLab Clone

This will look up all the projects you have access to in GitLab, and provide a way for you to select one or more to clone.

If you set the GITLAB_BASE_DIR environment variable to you root git directory,
new repos will automatically go into that directory regardless of where you are when running the command.
If that variable is not set, and no -b or --base-dir parameter is provided, the current directory is used.

Usage: glclone $( __glclone_options_display )

  The -b <dir> or --base-dir <dir> option will designate the directory to create your repo in.
        Providing this option overrides the default setting from the GITLAB_BASE_DIR.
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
    destination="$GITLAB_BASE_DIR"
    provided_repos=()
    while [[ "$#" -gt 0 ]]; do
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -b|--base-dir)
            __ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
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
        __delete_projects_file
    fi

    __ensure_gitlab_user_info
    __ensure_gitlab_projects

    projects="$( __filter_projects "$select_repo" '' "$( echo -E "${provided_repos[@]}" )" )"
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

__gtd_options_display () {
    echo -E -n '[-s|--select] [-o|--open] [-m|--mark-as-done] [--mark-all-as-done] [-q|--quiet] [-h|--help]'
}
__gtd_auto_options () {
    echo -E -n "$( __gtd_options_display | __convert_display_options_to_auto_options )"
}
gtd () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
gtd: GitLab ToDo List

Gets your GitLab TODO list.
You must create an api token from your profile in GitLab first. See: $( __get_gitlab_base_url )/profile/personal_access_tokens
Then, you must set the token value as the GITLAB_PRIVATE_TOKEN environment variable in your terminal (e.g. in .bash_profile)

Usage: gtd $( __gtd_options_display )

  With no options, your todo list is looked up and displayed.
  The -s or --select option will prompt you to choose entries.
        Selected entries are opened in your browser.
        This does not mark the entry as done, but some actions on the page might.
        You can select multiple entries using the tab key.
  The -o or --open option will cause your main todo page to be opened in your browser.
  The -m or --mark-as-done option will prompt you to choose entries (like with the -s option) and they will be marked as done.
        You can select multiple entries using the tab key.
        All selected entries will be marked as done.
  The -q or --quiet option will prevent unneeded output to stdout.
  The --mark-all-as-done option will mark all entries as done for you.

EOF
    )"
    local do_selector todo_count todo_list keep_quiet do_mark_as_done do_mark_all_as_done
    while [[ "$#" -gt 0 ]]; do
        local option
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -s|--select)
            do_selector="YES"
            ;;
        -o|--open)
            open "$( __get_gitlab_base_url )/dashboard/todos"
            ;;
        -m|--mark-as-done)
            do_mark_as_done="YES"
            do_selector="YES"
            ;;
        --mark-all-as-done)
            do_mark_all_as_done="YES"
            ;;
        -q|--quiet)
            keep_quiet="YES"
            ;;
        *)
            >&2 echo -E "Unknown option [ $option ]."
            return 2
            ;;
        esac
        shift
    done
    if [[ -n "$do_mark_all_as_done" ]]; then
        __mark_gitlab_todo_all_as_done "$keep_quiet"
    fi
    __get_gitlab_todos "$keep_quiet"
    todo_count=$( echo -E "$GITLAB_TODOS" | jq ' length ' )
    if [[ $todo_count -eq 0 ]]; then
        [[ -n "$keep_quiet" ]] || echo -E "You have nothing on your ToDo list."
    else
        if [[ $todo_count -eq 1 ]]; then
            [[ -n "$keep_quiet" ]] || echo -E "You have 1 thing to do."
        else
            [[ -n "$keep_quiet" ]] || echo -E "You have $todo_count things to do (oldest at top)."
        fi
        if [[ -z "$keep_quiet" ]]; then
            ( echo -E '┌───▪ Repo~┌───▪ Type~┌───▪ Title~┌───▪ Author~┌───▪ Url' \
                && echo -E "$GITLAB_TODOS" \
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                              def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                              .[] |    "├─▪ " + ( .project.name | clean )
                                    + "~├─▪ " + ( .target_type | clean )
                                    + "~├─▪ " + ( .body | .[0:25] | clean )
                                    + "~├─▪ " + ( .author.name | cleanname | .[0:20] )
                                    + "~├─▪ " + .target_url ' ) \
                | sed '$s/├/└/g' \
                | column -s '~' -t
        fi
        if [[ -n "$do_selector" ]]; then
            local fzf_todo_list selected_lines selected_line todo_id web_url
            fzf_todo_list=$( echo -E "$GITLAB_TODOS" | jq ' .[] | "ID>" + (.id|tostring) + "< ~" + .project.name + "~" + .target_type + "~" + (.body|.[0:80]|sub("\n";" ")) + "~" + .author.name + "~TARGET_URL>" + .target_url + "<" ' | sed -E 's/^"|"$//g' )
            selected_lines="$( ( echo -E ' ID~ Repo~ Type~ Title (oldest at top)~ Author' \
                && echo -E "$GITLAB_TODOS" \
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                              def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                              .[] |    "ID>" + ( .id | tostring ) + "< "
                                    + "~" + ( .project.name | clean )
                                    + "~" + ( .target_type | clean )
                                    + "~" + ( .body | .[0:80] | clean )
                                    + "~" + ( .author.name | cleanname )
                                    + "~TARGET_URL>" + .target_url  + "<" ' ) \
                | column -s '~' -t \
                | fzf --tac --header-lines=1 --cycle --with-nth=2,3,4,5 --delimiter="  +" -m )"
            echo -E "$selected_lines" | while read selected_line; do
                if [[ -n "$do_mark_as_done" ]]; then
                    todo_id=$( echo -E "$selected_line" | grep -E -o 'ID>[[:digit:]]+<' | sed -E 's/^ID>|<$//g' )
                    __mark_gitlab_todo_as_done "$keep_quiet" "$todo_id"
                else
                    web_url=$( echo -E "$selected_line" | grep -E -o 'TARGET_URL>[^[:space:]]+<' | sed -E 's/^TARGET_URL>|<$//g' )
                    if [[ -n $web_url ]]; then
                        open "$web_url"
                    fi
                fi
            done
        fi
    fi
}

__gljobs_options_display_1 () {
    echo -E -n '[-r <repo>|--repo <repo>] [-b <branch>|--branch <branch>|-a|--all-branches] [-q|--quiet] [-s|--select] [-o|--open]'
}
__gljobs_options_display_2 () {
    echo -E -n '[-p <page count>|--page-count <page count>|-d|--deep] [-x|--no-refresh] [-t <type>|--type <type>|--all-types] [-h|--help]'
}
__gljobs_auto_options () {
    echo -E -n "$( echo -E "$( __gljobs_options_display_1 ) $( __gljobs_options_display_2 )" | __convert_display_options_to_auto_options )"
}
gljobs () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
gljobs: GitLab Jobs

Get info about jobs in GitLab.

Usage: gljobs $( __gljobs_options_display_1 )
              $( __gljobs_options_display_2 )

  By default, if you are in a git repo, that will be used as the repo, and your current branch will be used as the branch.
  Also, by default, only the first page (100 jobs) of most recent jobs are retrieved for the repo.

  The -r <repo> or --repo <repo> option allows you to provide the repo instead of using the default.
  The -b <branch> or --branch <branch> option allows you to provide the branch instead of using the default.
        Cannot be used with -a or --all-branches.
  The -a or --all-branches option will display all branches.
        Cannot be used with a -b or --branch option.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option prompts you to select entries that will then be opened in your browser.
        Select multiple using the tab key.
  The -o or --open option will cause the first result to automatically be opened in your browser.
  The -p <page count> or --page-count <page count> option generally defines how far back in time to look.
        By default, only the first page of results (100) is retrieved across all branches for your repo.
        This option gives you a way to retrieve more jobs before filtering for your branch (or not filtering if you used -a).
        Cannot be used with -d or --deep.
  The -d or --deep option will retrieve all jobs for the repo.
        Cannot be used with the -p or --page-count option.
  The -x or --no-refresh option prevents a new lookup from happening and just displays the last results retrieved.
        Can only be combined with the -s and/or -q flags.
  The -t or --type option allows you to filter on build type.
        The list of jobs will be filtered to only include the supplied type.
        Common types are "build" "client" and "sdlc"
        If the provided type starts with a ~ then filtering will be to remove the supplied type.
        By default, there is a filter type of "~sdlc".
  The --all-types option disables the type filter.
        This is the same as -t "".

EOF
    )"
    local option provided_repo provided_branch do_all_branches keep_quiet do_selector open_first provided_page_count do_all_pages no_refresh filter_type all_types
    local repo branch page_count filter_type_with filter_type_base filter_type_msg filtered_list_count header selected_lines selected_line web_url
    while [[ "$#" -gt 0 ]]; do
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -r|--repo)
            __ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
            provided_repo="$2"
            shift
            ;;
        -b|--branch)
            __ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
            provided_branch="$2"
            shift
            ;;
        -a|--all-branches)
            do_all_branches="YES"
            ;;
        -q|--quiet)
            keep_quiet="YES"
            ;;
        -s|--select)
            do_selector="YES"
            ;;
        -o|--open)
            open_first="YES"
            ;;
        -p|--page-count)
            __ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
            provided_page_count="$2"
            shift
            ;;
        -d|--deep)
            do_all_pages="YES"
            ;;
        -x|--no-refresh)
            no_refresh="YES"
            ;;
        -t|--type)
            if [[ -n "${2+z}" && -z "$2" ]]; then
                all_types="YES"
            else
                __ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
                filter_type="$2"
            fi
            shift
            ;;
        --all-types)
            all_types="YES"
            ;;
        *)
            >&2 echo -E "Unknown option [ $option ]."
            >&2 echo -e "$usage"
            return 1
            ;;
        esac
        shift
    done
    if [[ -n "$provided_branch" && -n "$do_all_branches" ]]; then
        >&2 echo -E "Incompatible options: -b|--branch and -a|--all_branches."
        return 1
    fi
    if [[ -n "$page_count" && -n "$do_all_pages" ]]; then
        >&2 echo -E "Incompatible options: -p|--page-count and -d|--deep."
        return 1
    fi
    if [[ -n "$provided_repo" ]]; then
        repo="$provided_repo"
    elif [[ -n "$( git rev-parse --is-inside-work-tree 2>/dev/null )" ]]; then
        repo="$( basename $( git rev-parse --show-toplevel ) )"
    fi
    if [[ -z "$repo" ]]; then
        >&2 echo -E "Repo could not be determined."
        return 1
    fi
    if [[ -n "$provided_branch" ]]; then
        branch="$provided_branch"
    elif [[ -z "$do_all_branches" && -n "$( git rev-parse --is-inside-work-tree 2>/dev/null )" ]]; then
        branch="$( git branch | grep '^\*' | sed 's/^\* //' )"
    fi
    if [[ -z "$branch" && -z "$do_all_branches" ]]; then
        >&2 echo -E "Branch could not be determined."
        return 1
    fi
    if [[ -n "$do_all_pages" ]]; then
        page_count=
    elif [[ -z "$provided_page_count" ]]; then
        page_count=1
    elif [[ "$provided_page_count" =~ [^[:digit:]] || "$provided_page_count" -eq "0" ]]; then
        >&2 echo -E "Invalid page count [$provided_page_count]"
        return 1
    else
        page_count="$provided_page_count"
    fi
    if [[ -n "$no_refresh" ]]; then
        local bad_option
        if [[ -n "$provided_repo" ]]; then
            bad_option="-r|--repo"
        elif [[ -n "$provided_branch" ]]; then
            bad_option="-b|--branch"
        elif [[ -n "$do_all_branches" ]]; then
            bad_option="-a|--all-branches"
        elif [[ -n "$provided_page_count" ]]; then
            bad_option="-n"
        elif [[ -n "$do_all_pages" ]]; then
            bad_option="-d|--deep"
        fi
        if [[ -n "$bad_option" ]]; then
            >&2 echo -E "Incompatible options: -x|--no-refresh and $bad_option"
            return 1
        fi
    fi
    if [[ -n "$all_types" && -n "$filter_type" ]]; then
        >&2 echo -E "Incompatible options: -t|--type and --all-types"
        return 1
    fi
    if [[ -z "$filter_type"  && -z "$all_types" ]]; then
        filter_type="~sdlc"
    fi

    if [[ -z "$no_refresh" ]]; then
        __ensure_gitlab_projects "$keep_quiet"
        __get_jobs_for_project "$keep_quiet" "$repo" "$page_count" || return 2
        if [[ -n "$branch" ]]; then
            __filter_jobs_by_branch "$branch"
        fi
        if [[ -n "$filter_type" ]]; then
            __filter_jobs_by_type "$filter_type"
        fi
    fi
    filtered_list_count="$( echo -E "$GITLAB_JOBS" | jq ' length ' )"
    if [[ -n "$filter_type" ]]; then
        if [[ "${filter_type:0:1}" == "~" ]]; then
            filter_type_with="without"
            filter_type_base="${filter_type:1}"
        else
            filter_type_with="with"
            filter_type_base="$filter_type"
        fi
        filter_type_msg="$filter_type_with type $( __yellow "$filter_type_base" )"
    fi
    if [[ "$filtered_list_count" -eq "0" ]]; then
        if [[ -z "$keep_quiet" ]]; then
            header="No $( __yellow "$repo" )"
            header="$header jobs found"
            [[ -n "$branch" ]] && header="$header for $( __yellow "$branch" )"
            [[ -n "$filter_type_msg" ]] && header="$header $filter_type_msg"
            echo -e "$header"
        fi
    else
        GITLAB_JOBS="$( echo -E "$GITLAB_JOBS" | jq -c ' sort_by(-.commit_time_int, .status_sort, .short_type_sort, -.display_time_int) ' )"
        if [[ -n "$open_first" ]]; then
            open "$( echo -E "$GITLAB_JOBS" | jq -r ' .[0] | .web_url ' )"
        fi
        if [[ -z "$keep_quiet" ]]; then
            header="$filtered_list_count $( __yellow "$repo" )"
            header="$header job"
            [[ "$filtered_list_count" -ne 1 ]] && header="${header}s"
            header="$header found"
            [[ -n "$branch" ]] && header="$header for $( __yellow "$branch" )"
            [[ -n "$filter_type_msg" ]] && header="$header $filter_type_msg"
            header="$header (newest at top): "
            echo -e "$header"
            ( echo -E '┌───▪ Time~┌───▪ Status~┌───▪ Type~┌───▪ title~┌───▪ Url' \
                && echo -E "$GITLAB_JOBS" \
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                              .[] |    "├─▪ " + .display_time
                                    + "~├─▪ " + .status
                                    + "~├─▪ " + .short_type
                                    + "~├─▪ " + ( .commit.title | .[0:40] | clean )
                                    + "~├─▪ " + .web_url ' ) \
                | sed '$s/├/└/g' \
                | column -s '~' -t
        fi
        if [[ -n "$do_selector" ]]; then
            selected_lines="$( ( echo -E ' Time~ Status~ Type~ Title (newest at top)' \
                && echo -E "$GITLAB_JOBS" \
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                              .[] |         .display_time
                                    + "~" + .status
                                    + "~" + .short_type
                                    + "~" + ( .commit.title | .[0:80] | clean )
                                    + "~TARGET_URL>" + .web_url + "<" ' ) \
                | column -s '~' -t \
                | fzf --tac --header-lines=1 --cycle --with-nth=1,2,3,4 --delimiter="  +" -m )"
            if [[ -n "$selected_lines" ]]; then
                echo -E "$selected_lines" | while read selected_line; do
                    web_url=$( echo -E "$selected_line" | grep -E -o 'TARGET_URL>[^[:space:]]+<' | sed -E 's/^TARGET_URL>|<$//g' )
                    if [[ -n $web_url ]]; then
                        open "$web_url"
                    else
                        echo -E "Could not find TARGET_URL in '$selected_line'."
                    fi
                done
            fi
        fi
    fi
}

__glclean_options_display () {
    echo -E -n '[-v|--verbose] [-l|--list] [-h|--help]'
}
__glclean_auto_options () {
    echo -E -n "$( __glclean_options_display | __convert_display_options_to_auto_options )"
}
glclean () {
    local vars_to_clean vars_str usage
    vars_to_clean=("GITLAB_USER_ID" "GITLAB_USERNAME" "GITLAB_PROJECTS" "GITLAB_MRS"
                   "GITLAB_MRS_TODO" "GITLAB_MRS_BY_ME" "GITLAB_TODOS" "GITLAB_JOBS"
                   "GITLAB_MERGED_MRS" "GITLAB_MERGED_MRS_REPO")
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
        option="$( printf %s "$1" | __to_lowercase )"
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

__glmerged_options_display () {
    echo -E -n '<project> [-n <count>|--count <count>|--all] [-s|--select] [-q|--quiet]'
}
__glmerged_auto_options () {
    echo -E -n "$( __glmerged_options_display | __convert_display_options_to_auto_options )"
}
glmerged () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
glmerged: Looks up merged MRs for a GitLab repo.

glmerged $( __glmerged_options_display )

  If the provided project uniquely identifies a GitLab repository you have access to, it will be used.
  If there are multiple matches, you will be prompted to select the one you want.

  The -n or --count option indicates that you want to get a specific number of most recent merge requests.
        Default is 20.
        Cannot be combined with --all.
  The --all option indicates you want all the merge requests.
        Cannot be combined with -n or --count.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option prompts you to select entries that will then be opened in your browser.
        Select multiple using the tab key.

EOF
    )"
    if [[ "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ]]; then
        echo -e "$usage"
        return 0
    fi
    local search option res_count do_all keep_quiet do_select page_max per_page
    local initial_search search_count selected_repo repo_id branch repo_name mrs_url mrs
    local lines header selected_lines selected_line web_url
    if [[ -n "$1" && "${1:0:1}" != '-' ]]; then
        search="$1"
        shift
    fi
    while [[ "$#" -gt "0" ]]; do
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -n|--count)
            __ensure_option "$2" "$option" || return 1
            if [[ "$2" =~ [^[:digit:]] ]]; then
                >&2 echo -E "Invalid count [ $count ]."
                return 1
            fi
            res_count="$2"
            shift
            ;;
        -q|--quiet)
            keep_quiet="YES"
            ;;
        -s|--select)
            do_select="YES"
            ;;
        --all)
            do_all="YES"
            ;;
        *)
            >&2 echo -E "Unknown option: [ $option ]."
            >&2 echo -e "$usage"
            return 1
            ;;
        esac
        shift
    done
    if [[ -n "$res_count" && -n "$do_all" ]]; then
        >&2 echo "Conflicting options: [ --count $res_count ] and [ --all ]."
        return 1
    fi
    if [[ -n "$do_all" ]]; then
        res_count='999900'
        page_max='9999'
        per_page='100'
    elif [[ -n "$res_count" ]]; then
        if [[ "$res_count" -gt "100" ]]; then
            page_max="$( echo "$res_count" | sed 's/..$//;' )"
            per_page='100'
        else
            page_max='1'
            per_page="$res_count"
        fi
    else
        res_count='20'
        page_max='1'
        per_page='20'
    fi

    __ensure_gitlab_projects

    if [[ -n "$search" ]]; then
        initial_search="$( echo -E "$GITLAB_PROJECTS" \
                           | jq -r --arg search "$search" ' .[]
                                            | select( ( .path_with_namespace | contains($search) ) or ( .name_with_namespace | contains($search) ) )
                                            | .name_with_namespace + "   ID>" + (.id|tostring) + "<   BRANCH>" + .default_branch + "<   NAME>" + .name + "<" ' )"
        search_count="$( echo "$initial_search" | wc -l | sed 's/[^[:digit:]]//g' )"
        if [[ "$search_count" -eq "0" ]]; then
            >&2 echo "No repos found matching [ $search ]."
            return 1
        elif [[ "$search_count" -eq "1" ]]; then
            selected_repo="$initial_search"
        fi
    fi
    if [[ -z "$selected_repo" ]]; then
        selected_repo="$( echo -E "$GITLAB_PROJECTS" \
                          | jq -r ' .[] | .name_with_namespace + "   ID>" + (.id|tostring) + "<   BRANCH>" + .default_branch + "<   NAME>" + .name + "<" ' \
                          | fzf --tac --cycle --with-nth=1 --delimiter="   +" +m --query="$search" )"
    fi
    if [[ -z "$selected_repo" ]]; then
        return 0
    fi

    repo_id="$( echo -E "$selected_repo" | grep -E -o ' ID>[[:digit:]]+<' | sed -E 's/[^[:digit:]]//g' )"
    branch="$( echo -E "$selected_repo" | grep -E -o ' BRANCH>[^[:space:]]+<' | sed -E 's/^ BRANCH>|<$//g' )"
    repo_name="$( echo -E "$selected_repo" | grep -E -o ' NAME>[^[:space:]]+<' | sed -E 's/^ NAME>|<$//g' )"

    [[ -n "$keep_quiet" ]] || echo -e -n "Getting merged MRs for $( __yellow "$repo_name" )... "
    mrs_url="$( __get_gitlab_url_project_mrs "$repo_id" )?state=merged&target_branch=$branch&"
    mrs="$( __get_pages_of_url "$mrs_url" "$page_max" "$per_page" )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."

    GITLAB_MERGED_MRS="$( echo -E "$mrs" | jq -c --arg res_count "$res_count" ' sort_by(.merged_at) | reverse | .[0:( $res_count | tonumber )] ' )"
    GITLAB_MERGED_MRS_REPO="$repo_name"

    if [[ -z "$keep_quiet" ]]; then
        ( echo -E '┌────▪ ~┌───▪ Merged~┌───▪ Author~┌───▪ Title  (newest at top)~┌───▪ Url' \
            && echo -E "$GITLAB_MERGED_MRS" \
                | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          def cleandate: sub("T"; " ") | sub("\\.\\d\\d\\dZ"; "");
                          [ foreach .[] as $entry (0; .+1; . as $idx | $entry | .index = $idx ) ] | .[]
                            |    "├─▪ " + ( .index | tostring )
                              + "~├─▪ " + ( .merged_at | cleandate )
                              + "~├─▪ " + ( .author.name | cleanname | .[0:20] )
                              + "~├─▪ " + ( .title | .[0:40] | clean )
                              + "~├─▪ " + .web_url ' ) \
        | sed '$s/├/└/g' \
        | column -s '~' -t
    fi
    if [[ -n "$do_select" ]]; then
        selected_lines="$( ( echo -E ' ~ Merged~ Author~ Title (newest at top)' \
            && echo -E "$GITLAB_MERGED_MRS" \
                | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          def cleandate: sub("T"; " ") | sub("\\.\\d\\d\\dZ"; "");
                          [ foreach .[] as $entry (0; .+1; . as $idx | $entry | .index = $idx ) ] | .[]
                            |         ( .index | tostring )
                              + "~" + ( .merged_at | cleandate )
                              + "~" + ( .author.name | cleanname )
                              + "~" + ( .title | .[0:80] | clean )
                              + "~" + " TARGET_URL>" + .web_url + "<" ' ) \
            | column -s '~' -t \
            | fzf --tac --header-lines=1 --cycle --with-nth=1,2,3,4 --delimiter="  +" -m )"
        if [[ -n "$selected_lines" ]]; then
            echo -E "$selected_lines" | while read selected_line; do
                web_url="$( echo -E "$selected_line" | grep -E -o 'TARGET_URL>[^[:space:]]+<' | sed -E 's/^TARGET_URL>|<$//g' )"
                if [[ -n "$web_url" ]]; then
                    open "$web_url"
                else
                    echo -E "Could not find TARGET_URL in '$selected_line'."
                fi
            done
        fi
    fi
}

__glopen_options_display_1 () {
    echo -E -n '[-r [<repo>]|--repo [<repo>]|--select-repo] [-b [<branch]|--branch [<branch>]|--select-branch]'
}
__glopen_options_display_2 () {
    echo -E -n '[-d [<branch>]|--diff [<branch>]|--select-diff-branch] [-q|--quiet] [-x|--do-not-open]'
}
__glopen_auto_options () {
    echo -E -n "$( echo -E -n "$( __glopen_options_display_1 ) $( __glopen_options_display_2 )" | __convert_display_options_to_auto_options )"
}
glopen () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
glopen: GitLab Open

Opens up the webpage of a repo.

glopen $( __glopen_options_display_1 )
       $( __glopen_options_display_2 )

  The -r <repo> or --repo <repo> option is a way to provide the desired repo.
    If the desired repo cannot be found, you will be prompted to select one.
    If one of these are supplied without a desired repo,
      then these options are just like --select-repo.
    If this option is not supplied, and you are in a git repo, that repo will be used.
    If this option is not supplied, and you are not in a git repo, you will be prompted to select one.
    Additionally, if the GitLab project cannot be located by either the repo you are in, or the
      provided repo, you will be prompted to select one.
  The --select-repo option causes glopen to prompt you to select a repo.
  The -b <branch> or --branch <branch> option indicates that you want to open a specific branch.
    If a branch name is supplied with this option, that branch name will be used.
    If a branch name is not supplied, and -b or --branch are supplied,
      and you are in a git repo, your current branch will be used.
    If a branch is not supplied, and you are not in a git repo, you will be prompted to select a branch.
    Note: If you are in repo x, but provide or select repo y (using -r --repo or --select-repo), and
      -b or --branch is provided without a branch name, the branch that you are currently in (in repo x)
      will be used in conjuction with the base url for repo y.
      This makes it easier to open your current branch in multiple repos.
  The --select-branch option cause glopen to prompt you to select a specific branch.
  The -d <branch> or --diff <branch> option indicates that you want to open a diff page (instead of a specific branch page).
    This option defines the "from" branch for the diff.  The "to" branch is defined by the -b or --branch option behavior.
    If -d or --diff is provided without a branch, the "from" branch will default to master.
        The "to" branch is whatever is defined by a -b, --branch, or --select-branch option.
        If none of those are supplied, the "to" branch will be determined as if the -b option were provided.
  The --select-diff-branch option causes glopen to prompt you to select a specific "from" branch.
    It also indicates that you want to open the diff page (instead of the main page).
    The "to" branch will be determined the same way it would be if the -d or --diff option is supplied.
  The -q or --quiet option suppresses normal terminal output.
  The -x or --do-not-open option will prevent the pages from being opened and only output the info.
    Technically, you can provide both -q and -x, but then nothing will really happen.

EOF
    )"
    local provided_repos provided_branches option select_repo random_repo use_branch select_branch do_diff diff_branch select_diff_branch keep_quiet do_not_open
    provided_repos=()
    provided_branches=()
    while [[ "$#" -gt '0' ]]; do
        option="$( printf %s "$1" | __to_lowercase )"
        case "$option" in
        -h|--help)
            echo "$usage"
            return 0
            ;;
        -r|--repo)
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                provided_repos+=( $2 )
                shift
            else
                select_repo="YES"
            fi
            ;;
        --select-repo)
            select_repo="YES"
            ;;
        --random-repo|--random)
            if [[ -n "$2" && "$2" =~ ^[[:digit:]]+$ && "$2" -gt '1' ]]; then
                random_repo="$2"
                shift
            else
                random_repo='1'
            fi
            ;;
        -b|--branch)
            use_branch="YES"
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                provided_branches+=( $2 )
                shift
            fi
            ;;
        --select-branch)
            select_branch="YES"
            ;;
        -d|--diff)
            do_diff="YES"
            use_branch="YES"
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                diff_branch="$2"
                shift
            fi
            ;;
        --select-diff-branch)
            do_diff="YES"
            use_branch="YES"
            select_diff_branch="YES"
            ;;
        -q|--quiet)
            keep_quiet="YES"
            ;;
        -x|--do-not-open)
            do_not_open="YES"
            ;;
        *)
            >&2 echo "Unkown option: [$option]."
            return 1
        esac
        shift
    done
    local in_repo in_branch projects urls messages project_id urls_to_add project project_url \
          project_name project_ssh_url repo_branches fzf_header branch url message
    __ensure_gitlab_projects "$keep_quiet"
    if [[ -n "$random_repo" ]]; then
        for project_url in $( echo -E "$GITLAB_PROJECTS" | jq -r ' .[] | .web_url ' | sort -R | head -n "$random_repo" ); do
            open "$project_url"
        done
        return 0
    fi
    if [[ -n "$( git rev-parse --is-inside-work-tree 2>/dev/null )" ]]; then
        in_repo="$( basename $( git rev-parse --show-toplevel ) )"
        in_branch="$( git branch | grep '^\*' | sed 's/^\* //' )"
    fi
    projects="$( __filter_projects "$select_repo" "$in_repo" "$( echo -E "${provided_repos[@]}" )" )"
    if [[ "$( echo -E "$projects" | jq ' length ' )" -eq '0' ]]; then
        >&2 echo -E "GitLab project could not be determined."
        return 1
    fi
    if [[ -n "$do_diff" && -z "$select_diff_branch" && -z "$diff_branch" ]]; then
        diff_branch="master"
    fi
    urls=()
    messages=()
    for project_id in $( echo -E "$projects" | jq -r ' .[] | .id ' ); do
        urls_to_add=()
        project="$( echo -E "$projects" | jq -c --arg project_id "$project_id" ' .[] | select ( .id == ( $project_id | tonumber ) ) ' )"
        project_url="$( echo -E "$project" | jq -r ' .web_url ' )"
        project_name="$( echo -E "$project" | jq -r ' .name ' )"
        project_ssh_url="$( echo -E "$project" | jq -r ' .ssh_url_to_repo ' )"
        repo_branches=''
        if [[ -n "$do_diff" && -n "$select_diff_branch" ]]; then
            repo_branches="$( __get_branches_of_repo "$project_ssh_url" )"
            diff_branch="$( echo -E "$repo_branches" | fzf --tac --cycle +m --header="$project_name (from)" )"
        fi
        if [[ -n "$select_branch" || (( -n "$use_branch" && "${#provided_branches[@]}" -eq '0' && -z "$in_branch" )) ]]; then
            fzf_header="$project_name"
            if [[ -n "$do_diff" ]]; then
                fzf_header="$fzf_header (to)"
            fi
            if [[ -z "$repo_branches" ]]; then
                repo_branches="$( __get_branches_of_repo "$project_ssh_url" )"
            fi
            for branch in $( echo -E "$repo_branches" | fzf --tac --cycle -m --header="$fzf_header" ); do
                url="$( __get_glopen_url "$project_url" "$branch" "$diff_branch" )"
                urls_to_add+=( "$url" )
                messages+=( "$( __get_glopen_message "$project_name" "$url" "$branch" "$diff_branch" )" )
            done
        elif [[ "${#provided_branches[@]}" -gt '0' ]]; then
            for branch in "${provided_branches[@]}"; do
                url="$( __get_glopen_url "$project_url" "$branch" "$diff_branch" )"
                urls_to_add+=( "$url" )
                messages+=( "$( __get_glopen_message "$project_name" "$url" "$branch" "$diff_branch" )" )
            done
        elif [[ -n "$use_branch" && -n "$in_branch" ]]; then
            branch="$in_branch"
            url="$( __get_glopen_url "$project_url" "$branch" "$diff_branch" )"
            urls_to_add+=( "$url" )
            messages+=( "$( __get_glopen_message "$project_name" "$url" "$branch" "$diff_branch" )" )
        fi
        if [[ "${#urls_to_add[@]}" -eq 0 ]]; then
            urls_to_add+=( "$project_url" )
            messages+=( "$( __get_glopen_message "$project_name" "$project_url" "" "" )" )
        fi
        urls+=( "${urls_to_add[@]}" )
    done
    [[ -n "$keep_quiet" ]] || for message in "${messages[@]}"; do echo "$message"; done | column -s '~' -t
    if [[ -z "$do_not_open" ]]; then
        for url in "${urls[@]}"; do
            open "$url"
        done
    fi
}

# Makes sure that a GitLab token is set.
# This must be set outside of this file, and is kind of a secret thing.
# Usage: __ensure_gitlab_token
__ensure_gitlab_token () {
    if [[ -z "$GITLAB_PRIVATE_TOKEN" ]]; then
        >&2 cat << EOF
No GITLAB_PRIVATE_TOKEN has been set.
To create one, go to $( __get_gitlab_base_url )/profile/personal_access_tokens and create one with the "api" scope.
Then you can set it using
GITLAB_PRIVATE_TOKEN=whatever-your-token-is
It is probably best to put that line somewhere so that it will get executed whenever you start your terminal (e.g. .bash_profile)

EOF
        return 1
    fi
}

# Usage: <do stuff> | __convert_display_options_to_auto_options
__convert_display_options_to_auto_options () {
    if [[ -n "$1" ]]; then
        echo -E "$1" | __convert_display_options_to_auto_options
        return 0
    fi
    sed -E 's/<[^>]+>//g; s/\[|\]|\(|\)//g; s/\|/ /g; s/[[:space:]][[:space:]]+/ /g; s/^[[:space:]]//; s/[[:space:]]$//;'
}

# Usage: __highlight <text>
__highlight () {
    echo -e -n "\033[1;37m$1\033[0m"
}

# Usage: __yellow <text>
__yellow () {
    echo -e -n "\033[1;33m$1\033[0m"
}

# Joins all provided parameters using the provided delimiter.
# Usage: string_join <delimiter> [<arg1> [<arg2>... ]]
__gl_join () {
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

# Makes sure that an option was provided with a flag.
# Usage: __ensure_option "$2" "option name" || echo "bad option."
__ensure_option () {
    local value option
    value="$1"
    option="$2"
    if [[ -z "$value" || ${value:0:1} == "-" ]]; then
        >&2 echo -E "A parameter must be supplied with the $option option."
        return 1
    fi
    return 0
}

# Makes sure that your GitLab user info has been loaded.
# If not, it is looked up.
# Usage: __ensure_gitlab_user_info <keep quiet>
__ensure_gitlab_user_info () {
    local keep_quiet
    keep_quiet="$1"
    if [[ -z "$GITLAB_USER_ID" || -z "$GITLAB_USERNAME" ]]; then
        __get_gitlab_user_info "$keep_quiet"
    fi
}

# Looks up your GitLab user info. Results are stored in $GITLAB_USER_ID and $GITLAB_USERNAME.
# Usage: __get_gitlab_user_info <keep quiet>
__get_gitlab_user_info () {
    local keep_quiet user_info
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting your GitLab user id... "
    user_info="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __get_gitlab_url_user )" )"
    GITLAB_USER_ID="$( echo -E "$user_info" | jq '.id' )"
    GITLAB_USERNAME="$( echo -E "$user_info" | jq '.username' | sed -E 's/^"|"$//g' )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Makes sure that the $GITLAB_PROJECTS variable has a value.
# A temp file is used to store the project info too.
# If the file doesn't exist, or is older than a day, or is empty,
#   the projects info will be refreshed and stored in the file.
# Otherwise, it's contents will be loaded into the $GITLAB_PROJECTS variable.
# Usage: __ensure_gitlab_projects <keep quiet> <verbose>
__ensure_gitlab_projects () {
    local keep_quiet verbose projects_file
    keep_quiet="$1"
    verbose="$2"
    __ensure_temp_dir
    projects_file="$( __get_projects_filename )"
    if [[ ! -f "$projects_file" \
            || $( find "$projects_file" -mtime "+$GITLAB_PROJECTS_MAX_AGE" ) ]] \
            || ! $( grep -q '[^[:space:]]' "$projects_file" ); then
        __get_gitlab_projects "$keep_quiet" "$verbose"
        echo -E "$GITLAB_PROJECTS" > "$projects_file"
    else
        GITLAB_PROJECTS="$( cat "$projects_file" )"
    fi
}

__delete_projects_file () {
    projects_file="$( __get_projects_filename )"
    if [[ -f "$projects_file" ]]; then
        rm "$projects_file"
    fi
}

# Gets the full path and name of the file to store projects info.
# Usage: __get_projects_filename
__get_projects_filename () {
    echo -E -n "$GITLAB_TEMP_DIR/projects.json"
}

# Look up info on all available projects. Results are stored in $GITLAB_PROJECTS.
# Usage: __get_gitlab_projects <keep quiet> <verbose>
__get_gitlab_projects () {
    local keep_quiet verbose projects_url page per_page previous_count projects
    keep_quiet="$1"
    verbose="$2"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your GitLab projects... "
    projects_url="$( __get_gitlab_url_projects )?simple=true&membership=true&"
    projects="$( __get_pages_of_url "$projects_url" '' '' "$verbose" )"
    GITLAB_PROJECTS="$projects"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# This can be used to get a subset of $GITLAB_PROJECTS based on some searches, or usage of fzf.
# If <force select> has value:
#   * fzf will be used to prompt the user to select projects
# If <force select> does not have value:
#   * The <provided repos> are looked for, matching by name in the $GITLAB_PROJECTS data.
#   * If no <provided repos> are provided, the <current repo> is looked for.
#   * If any <provided repos> that aren't found exactly, fzf will prompt the user to select projects.
#       The projects that weren't found are used as an initial query for fzf.
# Usage: __filter_projects <force select> <current repo> <provided repos>
__filter_projects () {
    local force_select current_repo provided_repos projects search project fzf_search project_ids project_id
    force_select="$1"
    current_repo="$2"
    if [[ -n "$3" && "$3" =~ [^[:space:]] ]]; then
        provided_repos=( $3 )
    elif [[ -z "$force_select" && -n "$current_repo" ]]; then
        provided_repos=( $current_repo )
    else
        provided_repos=()
    fi

    projects='[]'
    if [[ "${#provided_repos[@]}" -gt '0' ]]; then
        for search in "${provided_repos[@]}"; do
            project="$( echo -E "$GITLAB_PROJECTS" | jq -c --arg search "$search" ' .[] | select( ( .name | ascii_downcase ) == ( $search | ascii_downcase ) ) ' )"
            if [[ -n "$project" ]]; then
                projects="$( echo -E "[$projects,[$project]]" | jq -c ' add ' )"
            else
                if [[ -z "$fzf_search" ]]; then
                    fzf_search="$search"
                else
                    fzf_search="$fzf_search | $search"
                fi
            fi
        done
    fi

    if [[ -n "$force_select" || -n "$fzf_search" || "$( echo -E "$projects" | jq ' length ' )" -eq '0' ]]; then
        project_ids="$( echo -E "$GITLAB_PROJECTS" \
            | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "");
                      sort_by(.name_with_namespace) | .[]
                        |                    ( .name_with_namespace | clean )
                          + "~PROJECT_ID>" + ( .id | tostring ) + "<" ' \
            | column -s "~" -t \
            | fzf --tac --cycle --with-nth=1 --delimiter="  +" -m -i --query="$fzf_search" \
            | grep -E -o 'PROJECT_ID>[^[:space:]]+<' \
            | sed -E 's/^PROJECT_ID>|<$//g' )"
        if [[ -n "$project_ids" ]]; then
            for project_id in $( echo -E "$project_ids" ); do
                project="$( echo -E "$GITLAB_PROJECTS" | jq -c --arg project_id "$project_id" ' .[] | select( .id == ( $project_id | tonumber ) ) ' )"
                if [[ -n "$project" ]]; then
                    projects="$( echo -E "[$projects,[$project]]" | jq -c ' add ' )"
                fi
            done
        fi
    fi

    if [[ "$( echo -E "$projects" | jq ' length ' )" -eq '0' ]]; then
        return 1
    fi
    echo -E -n "$( echo -E "$projects" | jq -c ' unique_by( .id ) | sort_by( .name ) ' )"
    return 0
}

# This is primarily used for research and testing.
# It's an easy way to get a project entry from $GITLAB_PROJECTS.
# Usage: __get_project <repo>
__get_project () {
    __filter_projects '' '' "$*"
}

# Gets all the branches of a repo.
# Usage: __get_branches_of_repo <repo ssh url>
__get_branches_of_repo () {
    git ls-remote "$1" 'refs/heads/*' | sed -E 's#^.*refs/heads/(.+)$#\1#;' | sort --ignore-case
}

# Creates the desired url for glopen to use.
# Usage: __get_glopen_url <base url> <branch> <diff_branch>
__get_glopen_url () {
    local base_url branch diff_branch
    base_url="$1"
    branch="$2"
    diff_branch="$3"
    echo -E -n "$base_url"
    if [[ -n "$branch" ]]; then
        if [[ -n "$diff_branch" && "$branch" != "$diff_branch" ]]; then
            echo -E -n "/compare/$diff_branch...$branch"
        else
            echo -E -n "/tree/$branch"
        fi
    fi
}

# Usage: <project name> <url> <branch> <diff branch>
__get_glopen_message () {
    local project_name url branch diff_branch cols
    project_name="$1"
    url="$2"
    branch="$3"
    diff_branch="$4"
    cols=()
    if [[ -n "$branch" && -n "$diff_branch" && "$branch" != "$diff_branch" ]]; then
        cols+=( "$diff_branch to $branch" "in" )
    elif [[ -n "$branch" ]]; then
        cols+=( "$branch" "in" )
    else
        cols+=( "main page" "of" )
    fi
    cols+=( "$project_name:" )
    cols+=( "$url" )
    __gl_join "~" "${cols[@]}"
}

# Look up a project name from its id.
# Usage: __get_project_name <project id>
__get_project_name () {
    local project_id project_name
    project_id="$1"
    project_name=$( echo -E "$GITLAB_PROJECTS" | jq " .[] | select(.id==$project_id) | .name " )
    echo -E -n "$project_name"
}

# Gets all the open MRS that you've created. Results are stored in $GITLAB_MRS_BY_ME.
# Usage: __get_gitlab_mrs_i_created <keep quiet>
__get_gitlab_mrs_i_created () {
    local keep_quiet mrs_url mrs
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all open MRs you created... "
    mrs_url="$( __get_gitlab_url_mrs )?scope=created_by_me&state=opened&"
    mrs="$( __get_pages_of_url "$mrs_url" )"
    GITLAB_MRS_BY_ME="$( echo -E "$mrs" | jq -c ' sort_by(.source_branch, .project_id) ' )"
    __add_project_names_to_mrs_i_created
    [[ -n "$keep_quiet" ]] || echo -E "Done."
    __add_approved_status_to_mrs_i_created
}

# Adds the .project_name parameter to the entries in $GITLAB_MRS_BY_ME.
# Usage: __add_project_names_to_mrs_i_created
__add_project_names_to_mrs_i_created () {
    local mr_project_ids mr_project_id project_name
    mr_project_ids="$( echo -E "$GITLAB_MRS_BY_ME" | jq ' [ .[] | .project_id ] | unique | .[] ' )"
    for mr_project_id in $( echo -E "$mr_project_ids" | sed -l '' ); do
        project_name="$( __get_project_name "$mr_project_id" )"
        GITLAB_MRS_BY_ME="$( echo -E "$GITLAB_MRS_BY_ME" | jq -c " [ .[] | if (.project_id == $mr_project_id) then (.project_name = $project_name) else . end ] " )"
    done
}

# Adds an .approved boolean parameter to each entry in $GITLAB_MRS_BY_ME.
# Usage: __add_approved_status_to_mrs_i_created <keep quiet>
__add_approved_status_to_mrs_i_created () {
    local keep_quiet mr_count mr_ids mr_index my_mrs mr_id mr mr_iid mr_project_id mr_project_name mr_approvals mr_approved
    keep_quiet="$1"
    mr_count="$( echo -E "$GITLAB_MRS_BY_ME" | jq ' length ' )"
    mr_ids="$( echo -E "$GITLAB_MRS_BY_ME" | jq ' .[] | .id ' )"
    mr_index=1
    my_mrs="[]"
    for mr_id in $( echo -E "$mr_ids" | sed -l '' ); do
        mr="$( echo -E "$GITLAB_MRS_BY_ME" | jq -c --arg mr_id "$mr_id" ' .[] | select(.id==($mr_id|tonumber)) ' )"
        mr_iid="$( echo -E "$mr" | jq ' .iid ' )"
        mr_project_id="$( echo -E "$mr" | jq ' .project_id ' )"
        mr_project_name="$( __get_project_name "$mr_project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rGetting approval Status of MRs: $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_approvals="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __get_gitlab_url_project_mr_approvals "$mr_project_id" "$mr_iid" )" )"
        mr_approved="$( echo -E "$mr_approvals" | jq ' .approved ' )"
        mr="$( echo -E "$mr" | jq -c " .approved = $mr_approved " )"
        my_mrs="$( echo -E "[$my_mrs,[$mr]]" | jq -c ' add ' )"
        mr_index=$(( mr_index + 1 ))
    done
    GITLAB_MRS_BY_ME="$( echo -E "$my_mrs" | jq -c ' sort_by(.source_branch, .project_id) ' )"
    [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r"
}

# Do a superficial search for MRs. Results are put in $GITLAB_MRS.
# This is a quicker search than __get_my_gitlab_mrs_deep, but often leaves MRs out of the list because of a bug in GitLab.
# Usage: __get_my_gitlab_mrs <keep quiet>
__get_my_gitlab_mrs () {
    local keep_quiet mrs_url mrs
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your open MRs... "
    mrs_url="$( __get_gitlab_url_mrs )?scope=all&state=opened&approver_ids\[\]=$GITLAB_USER_ID&"
    mrs="$( __get_pages_of_url "$mrs_url" )"
    GITLAB_MRS="$mrs"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Do a deep scan to get a full list of all MRs that are available for me to view. Results are put in $GITLAB_MRS.
# This usually takes a while because it will go through each project and get all MRs for that project (at least one call per project).
# It will often find more MRs than __get_my_gitlab_mrs though because of a bug in GitLab.
# Usage: __get_my_gitlab_mrs_deep <keep quiet>
__get_my_gitlab_mrs_deep () {
    local keep_quiet mrs mr_count project_ids project_count project_index project_id project_name mrs_url project_mrs project_mr_count
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E "Getting all your open MRs from all of your available projects... "
    mrs="[]"
    mr_count=0
    project_ids="$( echo -E "$GITLAB_PROJECTS" | jq ' .[] | .id ' )"
    project_count="$( echo -E "$GITLAB_PROJECTS" | jq ' length ' )"
    project_index=1
    for project_id in $project_ids; do
        project_name="$( __get_project_name "$project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r($mr_count) $project_index/$project_count - $project_id: $project_name "
        mrs_url="$( __get_gitlab_url_project_mrs "$project_id" )?state=opened&"
        project_mrs="$( __get_pages_of_url "$mrs_url" )"
        project_mr_count="$( echo -E "$project_mrs" | jq ' length ' )"
        if [[ "$project_mr_count" -gt "0" ]]; then
            mrs="$( echo -E "[$mrs,$project_mrs]" | jq -c ' add ')"
            mr_count=$(( mr_count + project_mr_count ))
        fi
        project_index=$(( project_index + 1 ))
    done
    GITLAB_MRS="$mrs"
    [[ -n "$keep_quiet" ]] || echo -e "\033[1K\rDone."
}

# Filter either $GITLAB_MRS_TODO or $GITLAB_MRS for only MRs where you are a suggested approver.
# The results are placed in $GIBLAB_MRS_TODO.
# This basically weeds out any MRs that either you don't need to care about, or you've already approved of.
# Usage: __filter_gitlab_mrs <keep quiet> <filter type>
# If filter type is "SHORT" then $GITLAB_MRS_TODO is filtered. Otherwise $GITLAB_MRS is filtered.
__filter_gitlab_mrs () {
    local keep_quiet filter_type mrs_to_filter mr_count mr_ids mr_index mr_todo_count my_mrs mr_id mr mr_iid mr_project_id mr_project_name mr_approvals keep_mr mr_approved
    keep_quiet="$1"
    filter_type="$2"
    mrs_to_filter="$( if [[ "$filter_type" == "SHORT" ]]; then echo -E "$GITLAB_MRS_TODO"; else echo -E "$GITLAB_MRS"; fi )"
    mr_count="$( echo -E "$mrs_to_filter" | jq ' length ' )"
    mr_ids="$( echo -E "$mrs_to_filter" | jq ' .[] | .id ' )"
    mr_index=1
    mr_todo_count=0
    my_mrs="[]"
    for mr_id in $( echo -E "$mr_ids" | sed -l '' ); do
        mr="$( echo -E "$mrs_to_filter" | jq -c --arg mr_id "$mr_id" ' .[] | select(.id==($mr_id|tonumber)) ' )"
        mr_iid="$( echo -E "$mr" | jq ' .iid ' )"
        mr_project_id="$( echo -E "$mr" | jq ' .project_id ' )"
        mr_project_name="$( __get_project_name "$mr_project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rFiltering MRs: ($mr_todo_count) $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_approvals="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __get_gitlab_url_project_mr_approvals "$mr_project_id" "$mr_iid" )" )"
        mr_state="$( echo -E "$mr_approvals" | jq -r ' .state ' )"
        keep_mr="$( echo -E "$mr_approvals" | jq --arg GITLAB_USER_ID "$GITLAB_USER_ID" ' .suggested_approvers[] | select(.id==($GITLAB_USER_ID|tonumber)) | "KEEP" ' )"
        if [[ -n "$keep_mr" && "$mr_state" == "opened" ]]; then
            mr_approved="$( echo -E "$mr_approvals" | jq ' .approved ' )"
            mr="$( echo -E "$mr" | jq -c " .project_name = $mr_project_name | .approved = $mr_approved " )"
            my_mrs="$( echo -E "[$my_mrs,[$mr]]" | jq -c ' add ' )"
            mr_todo_count=$(( mr_todo_count + 1 ))
        fi
        mr_index=$(( mr_index + 1 ))
    done
    GITLAB_MRS_TODO="$( echo -E "$my_mrs" | jq -c ' sort_by(.created_at) ' )"
    [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r"
}

# Adds discussion information to either $GITLAB_MRS_BY_ME or $GITLAB_MRS_TODO.
# Usage: __add_discussion_info_to_mrs <keep quiet> <mr list type>
# If mr list type is "MINE" then $GITLAB_MRS_BY_ME is processed, otherwise $GITLAB_MRS_TODO is.
__add_discussion_info_to_mrs () {
    local keep_quiet mr_list_type mrs_to_do mr_count mr_ids mr_index mr_todo_count my_mrs mr_id mr mr_iid mr_project_id mr_project_name
    local mr_discussions_url mr_discussions mr_discussions_of_interest mr_discussions_resolved mr_discussions_total mr_discussions_notes mr_discussions_stats
    keep_quiet="$1"
    mr_list_type="$2"
    mrs_to_do="$( if [[ "$mr_list_type" == "MINE" ]]; then echo -E "$GITLAB_MRS_BY_ME"; else echo -E "$GITLAB_MRS_TODO"; fi )"
    mr_count="$( echo -E "$mrs_to_do" | jq ' length ')"
    mr_ids="$( echo -E "$mrs_to_do" | jq ' .[] | .id ' )"
    mr_index=1
    mr_todo_count=0
    my_mrs="[]"
    for mr_id in $( echo -E "$mr_ids" | sed -l '' ); do
        mr=$( echo -E "$mrs_to_do" | jq -c --arg mr_id "$mr_id" ' .[] | select(.id==($mr_id|tonumber)) ' )
        mr_iid=$( echo -E "$mr" | jq ' .iid ' )
        mr_project_id=$( echo -E "$mr" | jq ' .project_id ' )
        mr_project_name=$( __get_project_name "$mr_project_id" )
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rGetting MR Discussion Info: $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_discussions_url="$( __get_gitlab_url_project_mr_discussions "$mr_project_id" "$mr_iid" )?"
        mr_discussions="$( __get_pages_of_url "$mr_discussions_url" )"
        mr_discussions_of_interest="$( echo -E "$mr_discussions" | jq -c '
                [ .[]
                    | .resolvable_notes = [ .notes[] | select(.resolvable) ]
                    | select((.resolvable_notes | length) != 0)
                    | .resolved = (reduce .resolvable_notes[] as $note (true; . and $note.resolved))
                ] ' )"
        mr_discussions_resolved="$( echo -E "$mr_discussions_of_interest" | jq -r ' [ .[] | select(.resolved) ] | length ' )"
        mr_discussions_total="$( echo -E "$mr_discussions_of_interest" | jq -r ' length ' )"
        mr_discussions_notes="$( echo -E "$mr_discussions_of_interest" | jq -r ' reduce .[] as $dis (0; . + ($dis.resolvable_notes | length)) ' )"
        mr_discussions_stats="$( printf "%3d/%3d (%3d)" "$mr_discussions_resolved" "$mr_discussions_total" "$mr_discussions_notes" )"
        mr="$( echo -E "$mr" | jq --arg mr_discussions_stats "$mr_discussions_stats" ' .discussion_stats = $mr_discussions_stats ')"
        my_mrs="$( echo -E "[$my_mrs,[$mr]]" | jq -c ' add ' )"
        mr_index=$(( mr_index + 1 ))
    done
    if [[ "$mr_list_type" == "MINE" ]]; then
        GITLAB_MRS_BY_ME="$( echo -E "$my_mrs" | jq -c ' sort_by(.created_at) ' )"
    else
        GITLAB_MRS_TODO="$( echo -E "$my_mrs" | jq -c ' sort_by(.created_at) ' )"
    fi
    [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r"
}

# Gets all the entries in your todo list.
# Usage: __get_gitlab_todos <keep quiet>
__get_gitlab_todos () {
    local keep_quiet todos todos_url
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting your GitLab ToDo List... "
    todos_url="$( __get_gitlab_url_todos )?"
    todos="$( __get_pages_of_url "$todos_url" )"
    GITLAB_TODOS="$( echo -E "$todos" | jq -c ' sort_by(.created_at) | reverse ' )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Marks a give todo item as done.
# Usage: __mark_gitlab_todo_as_done <keep quiet> <todo id>
__mark_gitlab_todo_as_done () {
    local keep_quiet todo_id marked_todo
    keep_quiet="$1"
    todo_id="$2"
    [[ -n "$keep_quiet" ]] || echo -E -n "Marking off ToDo... "
    marked_todo="$( curl -s --request POST --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __get_gitlab_url_todo_mark_as_done "$todo_id" )" )"
    if [[ -z "$keep_quiet" ]]; then
        local project_name todo_title
        project_name="$( echo -E "$marked_todo" | jq ' .project.name ' | sed -E 's/^"|"$//g' )"
        todo_title="$( echo -E "$marked_todo" | jq ' .body|.[0:80] ' )"
        echo -E "$project_name: $todo_title is marked as done."
    fi
}

# Marks all todo items as done.
# Usage: __mark_gitlab_todo_all_as_done <keep quiet>
__mark_gitlab_todo_all_as_done () {
    local keep_quiet all_done
    keep_quiet="$1"
    all_done="$( curl -s --request POST --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __get_gitlab_url_todo_mark_all_as_done )" )"
    [[ -n "$keep_quiet" ]] || echo -E "All TODO items marked as done."
}

# Gets all the jobs for the given project.
# Usage: __get_jobs_for_project <keep quiet> <project name> <page count max>
__get_jobs_for_project () {
    local keep_quiet project_name page_count_max project_id jobs_url gl_jobs short_types statuses
    keep_quiet="$1"
    project_name="$2"
    page_count_max="$3"
    project_id="$( echo -E "$GITLAB_PROJECTS" | jq " .[] | select(.name == \"$project_name\") | .id " )"
    if [[ -z "$project_id" ]]; then
        >&2 echo -E "Unkown project name: '$project_name'"
        return 1
    fi
    [[ -z "$keep_quiet" ]] && echo -E -n "Finding jobs for ($project_id) $project_name... "
    jobs_url="$( __get_gitlab_url_project_jobs "$project_id" )?"
    gl_jobs="$( __get_pages_of_url "$jobs_url" "$page_count_max" )"
    short_types='[{"k":"client","v":2},{"k":"build","v":1},{"k":"sdlc","v":3},{"k":"migrate","v":4}]'
    statuses='["manual","created","pending","started","running","success","canceled","failed","skipped"]'
    GITLAB_JOBS="$( echo -E "$gl_jobs" | jq -c --argjson short_types "$short_types" --argjson statuses "$statuses" '
        def cleandate: sub("T"; " ") | sub("\\.\\d\\d\\dZ"; "");
        def tovalue: gsub("\\D"; "") | tonumber;
        def indexordefault($l; $d): . as $v | $l | index($v) | . // $d;
        def keyvalue($k; $v; $l): . as $val | $l
            | ( label $out | foreach .[] as $e (-1; .+1; if ($e|.[$k]) == $val then ($e|.[$v]), break $out else empty end) // null );
        [ .[]
            | .short_type = ( (.name | ascii_downcase) as $type | $short_types
                              | reduce .[] as $st (null; if . == null and ($type | contains($st|.k)) then ($st|.k) else . end)
                              | if . != null then . else ($type | .[0:10]) end )
            | .short_type_sort = ( .short_type | keyvalue("k"; "v"; $short_types) // 99 )
            | .display_time = ( .finished_at // .created_at | cleandate )
            | .display_time_int = ( .display_time | tovalue )
            | .commit_time = ( .commit.created_at // .created_at | cleandate )
            | .commit_time_int = ( .commit_time | tovalue )
            | .status_sort = ( .status | indexordefault($statuses; 99) )
        ] ' )"
    [[ -z "$keep_quiet" ]] && echo -E "Done."
}

# Filters the entries in $GITLAB_JOBS to get just the ones applicable to the provided branch name.
# Usage: __filter_jobs_by_branch <branch name>
__filter_jobs_by_branch () {
    local branch_name
    branch_name="$1"
    if [[ -n "$branch_name" ]]; then
        GITLAB_JOBS="$( echo -E "$GITLAB_JOBS" | jq -c " [ .[] | select(.ref==\"$branch_name\") ] " )"
    fi
}

# Filters the entries in $GITLAB_JOBS based on a provided filter on the short_type parameter (added after getting all the jobs).
# The filter can be the short type to keep (e.g. "build") or the short_type to ignore (e.g. "~sdlc").
# Usage: __filter_jobs_by_type <filter type>
__filter_jobs_by_type () {
    local filter_type relation
    filter_type="$1"
    relation="=="
    if [[ -n "$filter_type" ]]; then
        if [[ "${filter_type:0:1}" == "~" ]]; then
            filter_type="${filter_type:1}"
            relation="!="
        fi
        GITLAB_JOBS="$( echo -E "$GITLAB_JOBS" | jq -c " [ .[] | select(.short_type $relation \"$filter_type\") ] " )"
    fi
}

# Gets all the pages for a given endpoint.
# The url is required, and must end in either a ? or a &.
# page count max is optional. Default is 9999. It is forced to be between 1 and 9999 (inclusive).
# per page is optional. Default is 100. It is forced to be between 1 and 100 (inclusive).
# Usage: __get_pages_of_url <url> [<page count max>] [<per page>] [<verbose>]
__get_pages_of_url () {
    local url page_count_max per_page verbose results page previous_count full_url page_data
    url="$1"
    page_count_max="$( __clamp "$2" "1" "9999" "9999" )"
    per_page="$( __clamp "$3" "1" "100" "100" )"
    verbose="$4"
    if [[ "$url" =~ [^?\&]$ ]]; then
        >&2 echo -E "__get_pages_of_url [$url] must end in either a ? or a & so that the per_page and page parameters can be added."
        return 1
    fi
    [[ -z "$verbose" ]] || >&2 echo -E "Page Count Max: [$page_count_max], Per Page: [$per_page]"
    results="[]"
    page=1
    previous_count=$per_page
    while [[ "$page" -le "$page_count_max" && "$previous_count" -eq "$per_page" ]]; do
        full_url="${url}per_page=${per_page}&page=${page}"
        [[ -z "$verbose" ]] || >&2 echo -E -n "Requesting $full_url ... "
        page_data="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$full_url" )"
        if [[ -n "$( echo -E "$page_data" | jq -r ' if type=="array" then "okay" else "" end ' )" ]]; then
            results="$( echo -E "[$results,$page_data]" | jq -c ' add ' )"
            previous_count="$( echo -E "$page_data" | jq ' length ' )"
            [[ -z "$verbose" ]] || >&2 echo -E "Done. Received $previous_count entries."
        else
            [[ -z "$verbose" ]] || >&2 echo -e "\033[1;38;5;231;48;5;196m ERROR \033[0m"
            if [[ -n "$verbose" || -n "$( echo -E "$page_data" | jq -r ' if type != "object" or .message != "403 Forbidden" then "showit" else "" end ' )" ]]; then
                >&2 echo -E "$full_url -> $page_data"
            fi
            previous_count=0
        fi
        page=$(( page + 1 ))
    done
    [[ -z "$verbose" ]] || >&2 echo -E "Final result count: $( echo -E "$results" | jq 'length' )."
    echo -E "$results"
}

# Makes sure that the gitlab temp directory exists.
# Usage: __ensure_temp_dir
__ensure_temp_dir () {
    if [[ -f "$GITLAB_TEMP_DIR" ]]; then
        rm "$GITLAB_TEMP_DIR"
    fi
    if [[ ! -d "$GITLAB_TEMP_DIR" ]]; then
        mkdir "$GITLAB_TEMP_DIR"
    fi
}

# Converts a string to lowercase.
# Usage: echo 'FOO' | __to_lowercase
__to_lowercase () {
    if [[ "$#" -gt '0' ]]; then
        printf '%s' "$*" | __to_lowercase
        return 0
    fi
    tr "[:upper:]" "[:lower:]"
}

# Makes sure that a provided value is a number between the min and max (inclusive).
# Only whole numbers are allowed.
# The min and max parameters are interchangable. If larger of the two is the max, and the other is the min.
# If the provided value is not a number, then the default is returned.
# If it's less than the min, the min is returned.
# If it's less than the max, the max is returned.
# Usage: __clamp <value> <min> <max> <default>
__clamp () {
    local val min max default result
    val="$( __ensure_number_or_default "$1" "" )"
    min="$( __ensure_number_or_default "$2" "" )"
    max="$( __ensure_number_or_default "$3" "" )"
    default="$4"
    if [[ -n "$min" && -n "$max" && "$min" > "$max" ]]; then
        local temp
        temp="$min"
        min="$max"
        max="$temp"
    fi
    if [[ -z "$val" ]]; then
        result="$default"
    elif [[ -n "$min" && "$val" -lt "$min" ]]; then
        result="$min"
    elif [[ -n "$max" && "$val" -gt "$max" ]]; then
        result="$max"
    else
        result="$val"
    fi
    echo -E -n "$result"
}

# Makes sure that a provide entry is a whole number (either positive or negative). If not, the provided default is returned.
# Usage: __ensure_number_or_default <value> <default>
__ensure_number_or_default () {
    local val default result
    val="$1"
    default="$2"
    if [[ -n "$val" && "$val" =~ ^-?[[:digit:]]+$ ]]; then
        result="$val"
    else
        result="$default"
    fi
    echo -E -n "$result"
}

# GitLab API documentation: https://docs.gitlab.com/ee/api/api_resources.html

# Usage: __get_gitlab_base_url
__get_gitlab_base_url () {
    echo -E -n 'https://gitlab.com'
}

# Usage: __get_gitlab_api_url
__get_gitlab_api_url () {
    __get_gitlab_base_url
    echo -E -n '/api/v4'
}

# Usage: __get_gitlab_url_user [<user id>]
__get_gitlab_url_user () {
    local user_id
    user_id="$1"
    __get_gitlab_api_url
    echo -E -n '/user'
    if [[ -n "$user_id" ]]; then
        echo -E -n "s/$user_id"
    fi
}

# Usage: __get_gitlab_url_merge_requests
__get_gitlab_url_mrs () {
    __get_gitlab_api_url
    echo -E -n '/merge_requests'
    # Note: This endpoint does not currently have the option to provide any sort of id for more specific information.
}

# Usage: __get_gitlab_url_projects [<project id>]
__get_gitlab_url_projects () {
    local project_id
    project_id="$1"
    __get_gitlab_api_url
    echo -E -n '/projects'
    if [[ -n "$project_id" ]]; then
        echo -E -n "/$project_id"
    fi
}

# Usage: __get_gitlab_url_project_jobs <project id> [<job id>]
__get_gitlab_url_project_jobs () {
    local project_id job_id
    project_id="$1"
    job_id="$2"
    __get_gitlab_url_projects "$project_id"
    echo -E -n '/jobs'
    if [[ -n "$job_id" ]]; then
        echo -E -n "/$job_id"
    fi
}

# Usage: __get_gitlab_url_project_jobs_log <project id> <job id>
__get_gitlab_url_project_jobs_log () {
    __get_gitlab_url_project_jobs "$1" "$2"
    echo -E -n '/trace'
}

# Usage: __get_gitlab_url_project_mrs <project id> [<merge request iid>]
__get_gitlab_url_project_mrs () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __get_gitlab_url_projects "$project_id"
    echo -E -n '/merge_requests'
    if [[ -n "$merge_request_iid" ]]; then
        echo -E -n "/$merge_request_iid"
    fi
}

# Usage: __get_gitlab_url_project_mr_approvals <project id> <merge request iid>
__get_gitlab_url_project_mr_approvals () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __get_gitlab_url_project_mrs "$project_id" "$merge_request_iid"
    echo -E -n '/approvals'
}

# Usage: __get_gitlab_url_project_mr_discussions <project id> <merge request iid>
__get_gitlab_url_project_mr_discussions () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __get_gitlab_url_project_mrs "$project_id" "$merge_request_iid"
    echo -E -n '/discussions'
}

# Usage: __get_gitlab_url_todos [<todo id>]
__get_gitlab_url_todos () {
    local todo_id
    todo_id="$1"
    __get_gitlab_api_url
    echo -E -n '/todos'
    if [[ -n "$todo_id" ]]; then
        echo -E -n "/$todo_id"
    fi
}

# Usage: __get_gitlab_url_todo_mark_as_done <todo id>
__get_gitlab_url_todo_mark_as_done () {
    local todo_id
    todo_id="$1"
    __get_gitlab_url_todos "$todo_id"
    echo -E -n "/mark_as_done"
}

# Usage: __get_gitlab_url_todo_mark_all_as_done
__get_gitlab_url_todo_mark_all_as_done () {
    __get_gitlab_url_todos
    echo -E -n "/mark_as_done"
}

if [[ "$sourced" == 'YES' ]]; then
    # Try to set up tab completion.
    if [[ -n "$( type complete 2>&1 | grep -v 'not found' )" ]]; then
        complete -W "$( __gitlab_auto_options )" gitlab
        complete -W "$( __gmr_auto_options )" gmr
        complete -W "$( __glclone_auto_options )" glclone
        complete -W "$( __gtd_auto_options )" gtd
        complete -W "$( __gljobs_auto_options )" gljobs
        complete -W "$( __glclean_auto_options )" glclean
        complete -W "$( __glmerged_auto_options )" glmerged
        complete -W "$( __glopen_auto_options )" glopen
    elif [[ -n "$( type compctl 2>&1 | grep -v 'not found' )" ]]; then
        compctl -x 'p[1]' -k "( $( __gitlab_auto_options ) )" -- gitlab
        compctl -k "( $( __gmr_auto_options ) )" gmr
        compctl -k "( $( __glclone_auto_options ) )" glclone
        compctl -k "( $( __gtd_auto_options ) )" gtd
        compctl -k "( $( __gljobs_auto_options ) )" gljobs
        compctl -k "( $( __glclean_auto_options ) )" glclean
        compctl -k "( $( __glmerged_auto_options ) )" glmerged
        compctl -k "( $( __glopen_auto_options ) )" glopen
    fi
else
    if [[ "$#" -gt '0' ]]; then
        gitlab "$@"
    else
        echo "For Usage: ./$( basename "$0" ) --help"
    fi
fi

