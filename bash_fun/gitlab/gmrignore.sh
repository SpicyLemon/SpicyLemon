#!/bin/bash
# This file contains the gmrignore function which allows you to manage a list of repos to ignore when deep scanning for MRs.
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

__gmrignore_options_display () {
    echo -E -n '[add|remove|update|clear|prune|status|list [<state(s)>]] [-h|--help]'
}
__gmrignore_auto_options () {
    echo -E -n "$( __gmrignore_options_display | __convert_display_options_to_auto_options )"
}
gmrignore () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
gmrignore: GitLab Merge Request Ignore (Projects)

Manages an ignore list for projects scanned by gmr.

gmrignore $( __gmrignore_options_display )

  Exactly one of these commands must be provided:
    add    - Display a list of projects that are not currently ignored and let you select ones to add.
    remove - Display a list of projects that are currently ignored, and let you select ones to remove from that list.
    update - Display a list of all projects, let you select ones to become the new ignore list.
    clear  - Clear the list of ignored projects.
    prune  - Get rid of the unknown projects.
                As projects are moved, deleted, or renamed in gitlab, sometimes their id changes.
                When that happens to a project that's ignored, the new project will not be ignored anymore and the
                old project id will still be in the ignore list.
                This command will clear out those old ids.
                WARNING:
                    Sometimes communication errors occur with GitLab and not all projects end up being retrieved.
                    When that happens, a large number of unknown entries will appear.
                    Once communication is fully restored, though, they'll show up again.
                    In these cases, you don't want to do a prune or else a lot of ignored projects will show up again.
    status - Output some consolidated information about what's being ignored.
    list   - Output repositories according to their state.
                This command can optionally take in one ore more states.
                Output will then be limited to the provided states.
                    Valid states are:  ignored  shown  unknown  all
                If none are provided, all will be used.

EOF
)"
    local option do_add do_remove do_update do_clear do_prune do_status do_list state list_ignored list_shown list_unknown list_all cmd_count
    while [[ "$#" -gt '0' ]]; do
        option="$( __to_lowercase "$1" )"
        case "$option" in
        h|help|-h|--help)
            echo "$usage"
            return 0
            ;;
        a|add|-a|--add)         do_add='YES';;
        r|remove|-r|--remove)   do_remove='YES';;
        u|update|-u|--update)   do_update='YES';;
        c|clear|-c|--clear)     do_clear='YES';;
        p|prune|-p|--prune)     do_prune='YES';;
        s|status|-s|--status)   do_status='YES';;
        l|list|-l|--list)
            do_list='YES'
            while [[ "$#" -gt '1' ]]; do
                state="$( __to_lowercase "$2" )"
                case "$state" in
                i|ignored)  list_ignored='YES';;
                s|shown)    list_shown='YES';;
                u|unknown)  list_unknown='YES';;
                a|all)      list_all='YES';;
                *)
                    >&2 echo "Unknown state [$2]."
                    return 1
                    ;;
                esac
                shift
            done
            ;;
        *)
            >&2 echo "Unknown command [$1]."
            return 1
            ;;
        esac
        shift
    done
    cmd_count="$( __count_non_empty "$do_add" "$do_remove" "$do_update" "$do_clear" "$do_prune" "$do_status" "$do_list" )"
    if [[ "$cmd_count" -eq '0' ]]; then
        >&2 echo -E "No command provided: $( __gmrignore_auto_options )"
        return 1
    elif [[ "$cmd_count" -ge '2' ]]; then
        >&2 echo -E "Only one command can be provided."
        return 1
    fi
    if [[ -n "$do_list" ]]; then
        state_count="$( __count_non_empty "$list_ignored" "$list_shown" "$list_unknown" )"
        if [[ "$state_count" -eq '0' ]]; then
            list_all='YES'
        fi
    fi
    __ensure_gl_config_dir || return 1
    local gmr_ignore_filename
    gmr_ignore_filename="$( __get_gmr_ignore_filename )"
    if [[ -n "$do_clear" ]]; then
        if [[ -f "$gmr_ignore_filename" ]]; then
            rm "$gmr_ignore_filename"
        fi
        if [[ -f "$gmr_ignore_filename" ]]; then
            >&2 "Unable to clear gitlab ignore list stored in $gmr_ignore_filename."
            return 1
        fi
        echo "gmr ignore list cleared."
        return 0
    fi
    __ensure_gitlab_projects
    local current_ignore_list ignored shown unknown all_count shown_count ignored_count unknown_count
    if [[ -f "$gmr_ignore_filename" ]] && grep -q '[^[:space:]]' "$gmr_ignore_filename" && [[ "$( jq -r ' length ' "$gmr_ignore_filename" )" -gt '0' ]]; then
        current_ignore_list="$( cat "$gmr_ignore_filename" )"
    else
        current_ignore_list='[]'
    fi
    ignored="$( echo -E "$GITLAB_PROJECTS" | jq -c --argjson ignore_list "$current_ignore_list" ' [ .[] | select( null != ( .id as $id | $ignore_list | index( $id ) ) ) ] ' )"
    shown="$( echo -E "$GITLAB_PROJECTS" | jq -c --argjson ignore_list "$current_ignore_list" ' [ .[] | select( null == ( .id as $id | $ignore_list | index( $id ) ) ) ] ' )"
    unknown="$( echo -E "$GITLAB_PROJECTS" | jq -c --argjson ignore_list "$current_ignore_list" ' $ignore_list - [ .[] | .id ] ' )"
    all_count="$( echo -E "$GITLAB_PROJECTS" | jq ' length ' )"
    shown_count="$( echo -E "$shown" | jq ' length ' )"
    ignored_count="$( echo -E "$ignored" | jq ' length ' )"
    unknown_count="$( echo -E "$unknown" | jq ' length ' )"

    if [[ -n "$do_prune" ]]; then
        local line_count_max count id just_did_nl yn new_ignore_list
        if [[ "$unknown_count" -eq '0' ]]; then
            echo -E "There are no unknown project ids to prune."
            return 0
        elif [[ "$unknown_count" -eq '1' ]]; then
            echo -E "There is 1 unknown project id to prune."
        else
            echo -E "There are $unknown_count unknown project ids to prune."
        fi
        line_count_max=10
        count=0
        for id in $( echo -E "$unknown" | jq -r ' .[] ' ); do
            echo -E -n "$id"
            count=$(( count + 1 ))
            just_did_nl=
            if [[ "$count" -lt "$unknown_count" ]]; then
                echo -n ","
                if [[ "$(( count % line_count_max ))" -eq '0' ]]; then
                    echo ''
                    just_did_nl='YES'
                else
                    echo -n ' '
                fi
            fi
        done
        if [[ -z "$just_did_nl" ]]; then
            echo ''
        fi
        echo 'Are you sure you want to prune these ids? [y|N]'
        read yn
        if [[ "$yn" =~ ^[Yy]([eE][sS])?$ ]]; then
            new_ignore_list="$( echo -E "[$current_ignore_list,$unknown]" | jq -c ' .[0] - .[1] | sort | unique ' )"
            echo -E "$new_ignore_list" > "$gmr_ignore_filename"
            echo "Ignore list pruned."
            return 0
        else
            echo "Nothing pruned."
            return 0
        fi
    fi
    if [[ -n "$do_status" ]]; then
        local report l
        report=()
        report+=( "gmr ignore list status:" )
        report+=( "    Known Projects:   $( printf '%5d' "$all_count" )" )
        report+=( "    Shown Projects:   $( printf '%5d' "$shown_count" )" )
        report+=( "    Ignored Projects: $( printf '%5d' "$ignored_count" )" )
        report+=( "    Unknown Projects: $( printf '%5d' "$unknown_count" )" )
        for l in "${report[@]}"; do
            echo -E "$l"
        done
        return 0
    fi
    if [[ -n "$do_list" ]]; then
        local report l
        report=()
        if [[ -n "$list_shown" || -n "$list_all" ]]; then
            if [[ "$shown_count" -eq '0' ]]; then
                report+=( "There are no projects that are shown." )
            elif [[ "$shown_count" -eq '1' ]]; then
                report+=( "There is one project that is shown." )
            else
                report+=( "There are $shown_count projects that are shown." )
            fi
            if [[ "$shown_count" -gt '0' ]]; then
                report+=( "$( echo -E "$shown" \
                                | jq -r 'def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                                         sort_by( .name_with_namespace )
                                         | .[]
                                         | "  +~" + ( .name_with_namespace | clean ) + "~" + .web_url + "~" + ( .id | tostring ) ' \
                                | column -s '~' -t )" )
            fi
            report+=( '' )
        fi
        if [[ -n "$list_ignored" || -n "$list_all" ]]; then
            if [[ "$ignored_count" -eq '0' ]]; then
                report+=( "There are no projects that are ignored." )
            elif [[ "$ignored_count" -eq '1' ]]; then
                report+=( "There is one project that is ignored." )
            else
                report+=( "There are $ignored_count projects that are ignored." )
            fi
            if [[ "$ignored_count" -gt '0' ]]; then
                report+=( "$( echo -E "$ignored" \
                                | jq -r 'def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                                         sort_by( .name_with_namespace )
                                         | .[]
                                         | "  -~" + ( .name_with_namespace | clean ) + "~" + .web_url + "~" + ( .id | tostring ) ' \
                                | column -s '~' -t )" )
            fi
            report+=( '' )
        fi
        if [[ -n "$list_unknown" || -n "$list_all" ]]; then
            if [[ "$unknown_count" -eq '0' ]]; then
                report+=( "There are no projects that are unknown." )
            elif [[ "$unknown_count" -eq '1' ]]; then
                report+=( "There is one project that is unknown." )
            else
                report+=( "There are $unknown_count projects that are unknown." )
            fi
            if [[ "$unknown_count" -gt '0' ]]; then
                report+=( "$( echo -E "$unknown" \
                                | jq -r 'def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                                         sort
                                         | .[]
                                         | "  ?~moved renamed or deleted~" + ( . | tostring ) ' \
                                | column -s '~' -t )" )
            fi
            report+=( '' )
        fi
        for l in "${report[@]}"; do
            echo -E "$l"
        done
        return 0
    fi
    if [[ -n "$do_add" || -n "$do_remove" || -n "$do_update" ]]; then
        local to_choose_from fzf_header selected selected_count selected_ids selected_names new_ignore_list new_ignore_count report l
        if [[ -n "$do_add" ]]; then
            if [[ "$shown_count" -eq '0' ]]; then
                >&2 echo -E "You do not currently have any shown projects."
                return 0
            fi
            to_choose_from="$shown"
            fzf_header="Select projects to ignore."
        elif [[ -n "$do_remove" ]]; then
            if [[ "$ignored_count" -eq '0' ]]; then
                >&2 echo -E "You do not currently have any ignored projects."
                return 0
            fi
            to_choose_from="$ignored"
            fzf_header="Select projects to show again."
        elif [[ -n "$do_update" ]]; then
            if [[ "$all_count" -eq '0' ]]; then
                >&2 echo -E "Your projects list is empty."
                return 0
            fi
            to_choose_from="$GITLAB_PROJECTS"
            fzf_header="Select projects to ignore."
        fi
        selected="$( echo "$to_choose_from" \
                        | jq -r ' sort_by( .name_with_namespace ) | .[] | ( .id | tostring ) + "~" + .name_with_namespace ' \
                        | __fzf_wrapper -m --with-nth='2' --to-columns -d '~' --header="$fzf_header" --cycle --tac )"
        if [[ -z "$selected" ]]; then
            echo "No selection made. The gmr ignore list is unchanged."
            return 0
        fi
        selected_count="$( echo "$selected" | wc -l | sed -E 's/[^[:digit:]]//g' )"
        selected_ids="[$( echo "$selected" | __gitlab_get_col '~' '1' | sed -E 's/[^[:digit:]]//g' | tr '\n' ',' | sed -E 's/,$//' )]"
        selected_names="$( echo "$selected" | __gitlab_get_col '~' '2' )"
        report=()
        if [[ -n "$do_add" ]]; then
            new_ignore_list="$( echo -E "[$current_ignore_list,$selected_ids]" | jq -c ' add | sort | unique ' )"
            if [[ "$selected_count" -eq '1' ]]; then
                report+=( "Added $selected_count project to the ignore list." )
            else
                report+=( "Added $selected_count projects to the ignore list." )
            fi
        elif [[ -n "$do_remove" ]]; then
            new_ignore_list="$( echo -E "[$current_ignore_list,$selected_ids]" | jq -c ' .[0] - .[1] | sort | unique ' )"
            if [[ "$selected_count" -eq '1' ]]; then
                report+=( "Removed $selected_count project from the ignore list." )
            else
                report+=( "Removed $selected_count projects from the ignore list." )
            fi
        elif [[ -n "$do_update" ]]; then
            new_ignore_list="$( echo -E "$selected_ids" | jq -c ' sort | unique ' )"
            if [[ "$selected_count" -eq '1' ]]; then
                report+=( "Updated the ignore list with $selected_count project." )
            else
                report+=( "Updated the ignore list with $selected_count projects." )
            fi
        fi
        if [[ "$selected_count" -le '20' ]]; then
            report+=( "$( echo -E "$selected_names" | sed 's/^/    /' )" )
        fi
        new_ignore_count="$( echo -E "$new_ignore_list" | jq ' length ' )"
        report+=( "There are $new_ignore_count projects being ignored." )
        echo -E "$new_ignore_list" > "$gmr_ignore_filename"
        for l in "${report[@]}"; do
            echo -E "$l"
        done
        return 0
    fi
    >&2 echo "Unkown command."
    return 1
}
