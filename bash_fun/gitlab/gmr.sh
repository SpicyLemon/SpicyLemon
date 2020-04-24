#!/bin/bash
# This file contains the gmr function that finds MRs that you're an approver on.
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

__gmr_options_display_1 () {
    echo -E -n '[-r|--refresh] [-d|--deep] [-b|--bypass-ignore] [-i|--include-approved] [-m|--mine]'
}
__gmr_options_display_2 () {
    echo -E -n '[-u|--update] [-q|--quiet] [-s|--select] [-o|--open-all] [-h|--help]'
}
__gmr_auto_options () {
    echo -E -n "$( echo -E "$( __gmr_options_display_1 ) $( __gmr_options_display_2 )" | __gl_convert_display_options_to_auto_options )"
}
gmr () {
    __gl_require_token || return 1
    local usage
    usage="$( cat << EOF
gmr: GitLab Merge Requests

Gets information about merge requests you are involved in.

Usage: gmr $( __gmr_options_display_1 )
           $( __gmr_options_display_2 )

  With no options, if there is no previous results, new results are looked up.
  With no options, if there ARE previous results, those old results are displayed.
  The -r or --refresh option causes gmr to reach out to GitLab to get a current list of your MRs (the easy, but incomplete way).
  The -d or --deep option causes gmr to go through each project you can see to check for merge requests that request your approval.
        This will take longer, but might uncover some MRs that do not show up with the simple (-r) lookup.
        If supplied with the -r option, the -r option is ignored.
  The -b or --bypass-ignore option only makes sense along with the -d or --deep option.
        It will cause gmr to bypass the ignore list that you have setup using gmrignore.
  The -i or --include-approved flag tells gmr to also display mrs that you have already approved.
  The -m or --mine option lists MRs that you created.
  The -u or --update option causes gmr to go through the known lists of MRs and update them with respect to comments and approvals.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option makes gmr prompt you to select entries that will be opened in your browser.
        You can select multiple entries using the tab key. All selected entries will be opened in your browser.
  The -o or --open-all option causes all MRs to be opened in your browser.

Basically, the first time you run the  gmr  command, you will get a list of MRs (eventually).
After that (in the same terminal) running  gmr  again will display the previous results.
In order to update the list again, do a  gmr --refresh

EOF
    )"
    local option do_refresh do_update do_deep bypass_ignore show_approved do_mine do_selector keep_quiet open_all \
          refresh_type filter_type discussion_type mrs todo_count
    while [[ "$#" -gt 0 ]]; do
        option="$( __gl_lowercase "$1" )"
        case "$option" in
        -h|--help|help)
            echo -e "$usage"
            return 0
            ;;
        -r|--refresh)           do_refresh="YES" ;;
        -d|--deep)              do_deep="YES" ;;
        -b|--bypass-ignore)     bypass_ignore='YES' ;;
        -i|--include-approved)  show_approved='YES' ;;
        -m|--mine)              do_mine="YES" ;;
        -u|--update)            do_update="YES" ;;
        -q|--quiet)             keep_quiet="YES" ;;
        -s|--select)            do_selector="YES" ;;
        -o|--open-all)          open_all="YES" ;;
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
        return 0
    fi
    if [[ -n "$do_refresh" && -n "$do_update" ]]; then
        >&2 echo -E "The --refresh option overrides the --update option; --update is being ignored."
        do_update=
    fi
    if [[ -n "$do_mine" && -n "$show_approved" ]]; then
        >&2 echo -E "The --include-approved option has no meaning with --mine; --include-approved is being ignored."
        show_approved=
    fi
    if [[ -n "$do_update" && -n "$do_mine" ]]; then
        >&2 echo -E "The --update option has no meaning with --mine; treating it as --refresh instead."
        do_refresh='YES'
        do_update=
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

    __gl_ensure_user_info "$keep_quiet"
    __gl_ensure_projects "$keep_quiet"

    if [[ -n "$refresh_type" ]]; then
        case "$refresh_type" in
        "DEEP") __gl_get_mrs_to_approve_deep "$keep_quiet" "$bypass_ignore" ;;
        "MINE") __gl_get_mrs_i_created "$keep_quiet" ;;
        *)      __gl_get_mrs_to_approve_simple "$keep_quiet" ;;
        esac
    fi

    if [[ -n $filter_type ]]; then
        __gl_mrs_filter_by_approver "$keep_quiet" "$filter_type"
    fi

    if [[ -n "$discussion_type" ]]; then
        __gl_add_discussion_info_to_mrs "$keep_quiet" "$discussion_type"
    fi

    if [[ -n "$do_mine" ]]; then
        mrs="$GITLAB_MRS_BY_ME"
    elif [[ -n "$show_approved" ]]; then
        mrs="$GITLAB_MRS_TODO"
    else
        mrs="$( echo -E "$GITLAB_MRS_TODO" | jq -c ' [ .[] | select(.i_approved == false) ] ' )"
    fi
    todo_count="$( echo -E "$mrs" | jq ' length ' )"
    if [[ $todo_count -eq 0 ]]; then
        if [[ -z "$keep_quiet" ]]; then
            echo -E -n "You have no"
            [[ -n "$show_approved" ]] && echo -E -n " open"
            echo -E -n " MRs"
            if [[ -z "$do_mine" && -z "$show_approved" ]]; then
                echo -E " to review!!"
            else
                echo -E "."
            fi
        fi
    else
        if [[ -z "$keep_quiet" ]]; then
            echo -E -n "You have $todo_count"
            [[ -n "$show_approved" ]] && echo -E -n " open"
            echo -E -n " MRs"
            [[ -z "$do_mine" && -z "$show_approved" ]] && echo -E -n " to review"
            [[ -z "$do_mine" ]] && echo -E -n " (oldest on top)"
            echo -E "."
            ( echo -E '┌───▪ Repo~┌───▪ Author~┌───▪ Discussions~┌───▪ Title~┌───▪ Url' \
                && echo -E "$mrs" \
                    | jq -r --arg box_checked '☑' --arg box_empty '☐' --arg root '√' \
                        ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          .[] | .col_head = "├" + ( if .i_approved == true then $root else "─" end )
                                                + ( if .approved == true then $box_checked else $box_empty end) + " "
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
            selected_lines="$( ( echo -E "  ~ Repo~ Author~ Discussions~ Title$( [[ -z "$do_mine" ]] && echo -E " (oldest on top)" )" \
                && echo -E "$mrs" \
                    | jq -r --arg box_checked '☑' --arg box_empty '☐' --arg root '√' \
                        ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          .[] |         ( if .i_approved == true then $root else " " end )
                                      + ( if .approved == true then $box_checked else $box_empty end )
                                + "~" + ( .project_name | clean )
                                + "~" + ( .author.name | cleanname )
                                + "~" + .discussion_stats
                                + "~" + ( .title | .[0:80] | clean )
                                + "~" + .web_url ' ) \
                | fzf_wrapper --tac --header-lines=1 --cycle --with-nth=1,2,3,4,5 --delimiter="~" -m --to-columns )"
            echo -E "$selected_lines" | while read selected_line; do
                web_url="$( echo -E "$selected_line" | __gl_column_value '~' '6' )"
                if [[ -n $web_url ]]; then
                    open "$web_url"
                fi
            done
        fi
    fi
}
