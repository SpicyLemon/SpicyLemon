#!/bin/bash
# This file contains the gtd function that can be used to interact with your gitlab todo list.
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
        option="$( __to_lowercase "$1" )"
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
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
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
            local selected_lines selected_line todo_id web_url
            selected_lines="$( ( echo -E ' ID~ Repo~ Type~ Title (oldest at top)~ Author' \
                && echo -E "$GITLAB_TODOS" \
                    | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                              def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                              .[] |         ( .id | tostring )
                                    + "~" + ( .project.name | clean )
                                    + "~" + ( .target_type | clean )
                                    + "~" + ( .body | .[0:80] | clean )
                                    + "~" + ( .author.name | cleanname )
                                    + "~" + .target_url ' ) \
                | fzf_wrapper --tac --header-lines=1 --cycle --with-nth=2,3,4,5 --delimiter="~" -m --to-columns )"
            echo -E "$selected_lines" | while read selected_line; do
                if [[ -n "$do_mark_as_done" ]]; then
                    todo_id="$( echo -E "$selected_line" | __gitlab_get_col '~' '1' )"
                    __mark_gitlab_todo_as_done "$keep_quiet" "$todo_id"
                else
                    web_url="$( echo -E "$selected_line" | __gitlab_get_col '~' '6' )"
                    if [[ -n $web_url ]]; then
                        open "$web_url"
                    fi
                fi
            done
        fi
    fi
}
