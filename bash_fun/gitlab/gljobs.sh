#!/bin/bash
# This file contains the gljobs function that can be used to find jobs for repos and branches.
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
        option="$( __to_lowercase "$1" )"
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
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
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
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                              .[] |         .display_time
                                    + "~" + .status
                                    + "~" + .short_type
                                    + "~" + ( .commit.title | .[0:80] | clean )
                                    + "~" + .web_url ' ) \
                | __fzf_wrapper --tac --header-lines=1 --cycle --with-nth=1,2,3,4 --delimiter="~" -m --to-columns )"
            if [[ -n "$selected_lines" ]]; then
                echo -E "$selected_lines" | while read selected_line; do
                    web_url="$( echo -E "$selected_line" | __gitlab_get_col '~' '5' )"
                    if [[ -n $web_url ]]; then
                        open "$web_url"
                    fi
                done
            fi
        fi
    fi
}
