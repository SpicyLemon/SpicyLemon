#!/bin/bash
# This file contains the glmerged function which can be used to look up merged merge requests.
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
        option="$( __to_lowercase "$1" )"
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
                           | jq -r --arg search "$search" 'def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                                        .[] | select( ( .path_with_namespace | contains($search) ) or ( .name_with_namespace | contains($search) ) )
                                            |         ( .name_with_namespace | clean )
                                              + "~" + ( .id | tostring )
                                              + "~" + ( .default_branch // "master" | clean )
                                              + "~" + ( .name | clean ) ' )"
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
                          | jq -r 'def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                                    .[] |         ( .name_with_namespace | clean )
                                          + "~" + ( .id | tostring )
                                          + "~" + ( .default_branch // "master" | clean )
                                          + "~" + ( .name | clean ) ' \
                          | __fzf_wrapper --tac --cycle --with-nth=1 --delimiter="~" +m --query="$search" --to-columns )"
    fi
    if [[ -z "$selected_repo" ]]; then
        return 0
    fi

    repo_id="$( echo -E "$selected_repo" | __gitlab_get_col '~' '2' )"
    branch="$( echo -E "$selected_repo" | __gitlab_get_col '~' '3' )"
    repo_name="$( echo -E "$selected_repo" | __gitlab_get_col '~' '4' )"

    [[ -n "$keep_quiet" ]] || echo -e -n "Getting merged MRs for $( __yellow "$repo_name" ) ... "
    mrs_url="$( __get_gitlab_url_project_mrs "$repo_id" )?state=merged&target_branch=$branch&"
    mrs="$( __get_pages_of_url "$mrs_url" "$page_max" "$per_page" )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."

    GITLAB_MERGED_MRS="$( echo -E "$mrs" | jq -c --arg res_count "$res_count" ' sort_by(.merged_at) | reverse | .[0:( $res_count | tonumber )] ' )"
    GITLAB_MERGED_MRS_REPO="$repo_name"

    if [[ -z "$keep_quiet" ]]; then
        ( echo -E '┌────▪ ~┌───▪ Merged~┌───▪ Author~┌───▪ Title  (newest at top)~┌───▪ Url' \
            && echo -E "$GITLAB_MERGED_MRS" \
                | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
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
                | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          def cleandate: sub("T"; " ") | sub("\\.\\d\\d\\dZ"; "");
                          [ foreach .[] as $entry (0; .+1; . as $idx | $entry | .index = $idx ) ] | .[]
                            |         ( .index | tostring )
                              + "~" + ( .merged_at | cleandate )
                              + "~" + ( .author.name | cleanname )
                              + "~" + ( .title | .[0:80] | clean )
                              + "~" + .web_url ' ) \
            | __fzf_wrapper --tac --header-lines=1 --cycle --with-nth=1,2,3,4 --delimiter="~" -m --to-columns )"
        if [[ -n "$selected_lines" ]]; then
            echo -E "$selected_lines" | while read selected_line; do
                web_url="$( echo -E "$selected_line" | __gitlab_get_col '~' '5' )"
                if [[ -n "$web_url" ]]; then
                    open "$web_url"
                fi
            done
        fi
    fi
}
