#!/bin/bash
# This file contains the gmrsearch function that allows you to do generic searches for MRs.
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

__gmrsearch_options_display_behavior_01 () {
    echo -E -n '[-h|--help] [-d|--deep] [-b|--bypass-ignore] [-s|--select] [--json] [--use-last-results] [-v|--verbose]'
}
__gmrsearch_options_display_result_control_01 () {
    echo -E -n '[--order-by <field>|--order-by-created|--order-by-updated] [--sort <direction>|--asc|--desc]'
}
__gmrsearch_options_display_search_01 () {
    echo -E -n '[--state <mr state>|--opened|--closed|--locked|--merged]'
}
__gmrsearch_options_display_search_02 () {
    echo -E -n '[--scope <scope>|--created-by-me|--assigned-to-me|--scope-all]'
}
__gmrsearch_options_display_search_03 () {
    echo -E -n '[--created-after <start datetime>] [--created-before <end datetime>]'
}
__gmrsearch_options_display_search_04 () {
    echo -E -n '[--created-between <start datetime> <end datetime>] [--created-on <date>]'
}
__gmrsearch_options_display_search_05 () {
    echo -E -n '[--updated-after <start datetime>] [--updated-before <end datetime>]'
}
__gmrsearch_options_display_search_06 () {
    echo -E -n '[--updated-between <start datetime> <end datetime>] [--updated-on <date>]'
}
__gmrsearch_options_display_search_07 () {
    echo -E -n '[--search <text>] [--source-branch <branch>] [--target-branch <branch>] [--wip <yes/no>]'
}
__gmrsearch_options_display_search_08 () {
    echo -E -n '[--author <author>] [--author-id <user id>] [--author-username <username>]'
}
__gmrsearch_options_display_search_09 () {
    echo -E -n '[--assignee-id <user id>] [--approver-ids <user id list>] [--approved-by-ids <user id list>]'
}
__gmrsearch_options_display_search_10 () {
    echo -E -n '[--labels <labels>] [--milestone <milestone>] [--my-reaction-emoji <emoji>]'
}
__gmrsearch_options_display_expert_01 () {
    echo -E -n '[--with-labels-details [<yes/no>]|--without-labels-details] [--view <view>|--view-simple|--view-normal]'
}
__gmrsearch_auto_options () {
    echo -E -n "$( echo -E -n "$( __gmrsearch_options_display_behavior_01 ) " \
                              "$( __gmrsearch_options_display_result_control_01 ) " \
                              "$( __gmrsearch_options_display_search_01 ) " \
                              "$( __gmrsearch_options_display_search_02 ) " \
                              "$( __gmrsearch_options_display_search_03 ) " \
                              "$( __gmrsearch_options_display_search_04 ) " \
                              "$( __gmrsearch_options_display_search_05 ) " \
                              "$( __gmrsearch_options_display_search_06 ) " \
                              "$( __gmrsearch_options_display_search_07 ) " \
                              "$( __gmrsearch_options_display_search_08 ) " \
                              "$( __gmrsearch_options_display_search_09 )"  \
                              "$( __gmrsearch_options_display_search_10 )"  \
                              "$( __gmrsearch_options_display_expert_01 ) " \
                            | __convert_display_options_to_auto_options )"
}
gmrsearch () {
    __ensure_gitlab_token || return 1
    local usage
    usage="$( cat << EOF
gmrsearch: Gitlab Merge Request Search

Search for Merge Requests based on certain criteria.

Usage: gmrsearch $( __gmrsearch_options_display_behavior_01 )
                 $( __gmrsearch_options_display_result_control_01 )
                 $( __gmrsearch_options_display_search_01 )
                 $( __gmrsearch_options_display_search_02 )
                 $( __gmrsearch_options_display_search_03 )
                 $( __gmrsearch_options_display_search_04 )
                 $( __gmrsearch_options_display_search_05 )
                 $( __gmrsearch_options_display_search_06 )
                 $( __gmrsearch_options_display_search_07 )
                 $( __gmrsearch_options_display_search_08 )
                 $( __gmrsearch_options_display_search_09 )
                 $( __gmrsearch_options_display_search_10 )
                 $( __gmrsearch_options_display_expert_01 )

Behavior:
  -h or --help              Show this usage information.
  -d or --deep              Cause the search to happen on each project instead of using the simple search.
                            This will take longer, but might uncover some MRs that do not show up with the simple search.
  -b or --bypass-ignore     This option only makes sense along with the -d or --deep option.
                            It will cause the deep search to bypass the ignore list and do the search on all available projects.
  -s or --select            Results will be presented for you to select. Selected results will be opened in your browser.
  --json                    Output the results as json instead of the normal formatted table.
  --use-last-results        Instead of actually doing a search, the previous search results are used.
                            This can be handy when combined with -s or --json.
  -v or --verbose           Output some extra messages to possibly help with debugging.

Result Control:
  --order-by <field>        Define the field that should be used to sort the results.
                            Valid values:  created_at  updated_at
                            Default value is  created_at
    --order-by-created        Shortcut for  --order-by created_at
    --order-by-updated        Shortcut for  --order-by updated_at
  --sort <direction>        Define the direction of the ordering
                            Valid values:  asc  desc
                            Default value is  desc
    --asc                     Shortcut for  --sort asc
    --desc                    Shortcut for  --sort desc

Search Criteria:
  --state <mr state>        Limit search to the provided state.
                            Valid states:  opened  closed  locked  merged
    --opened                  Shortcut for  --state opened
    --closed                  Shortcut for  --state closed
    --locked                  Shortcut for  --state locked
    --merged                  Shortcut for  --state merged
  --scope <scope>           Limit search to the provided scope.
                            Valid scopes:  created_by_me  assigned_to_me  all
    --created-by-me           Shortcut for  --scope created_by_me
    --assigned-to-me          Shortcut for  --scope assigned_to_me
    --scope-all               Shortcut for  --scope all
  --created-after <start datetime>
                            Limit search to MRs created on or after a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 00
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 00:00
                                YYYY-MM-DD          time is assumed to be 00:00:00
  --created-before <end datetime>
                            Limit search to MRs created on or before a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 59
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 59:59
                                YYYY-MM-DD          time is assumed to be 23:59:59
    --created-between <start datetime> <end datetime>
                              Shortcut for  --created-after <start datetime> --created-before <end datetime>
                              The datetimes must be in one of these formats:  YYYY-MM-DD hh:mm:ss  YYYY-MM-DD hh:mm  YYYY-MM-DD hh  YYYY-MM-DD
                              Both datetimes must use the same format.
                              The two datetimes can be in any order and they will be applied appropriately.
  --updated-after <start datetime>
                            Limit search to MRs updated on or after a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 00
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 00:00
                                YYYY-MM-DD          time is assumed to be 00:00:00
  --updated-before <end datetime>
                            Limit search to MRs updated on or before a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 59
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 59:59
                                YYYY-MM-DD          time is assumed to be 23:59:59
    --updated-between <start datetime> <end datetime>
                              Shortcut for  --updated-after <start datetime> --updated-before <end datetime>
                              The datetimes must be in one of these formats:  YYYY-MM-DD hh:mm:ss  YYYY-MM-DD hh:mm  YYYY-MM-DD hh  YYYY-MM-DD
                              Both datetimes must use the same format.
                              The two datetimes can be in any order and they will be applied appropriately.
  --search <text>           Limit search to MRs with the provided text in either the title or description.
  --source-branch <branch>  Limit search to MRs with the provided branch as the source branch.
  --target-branch <branch>  Limit search to MRs with the provided branch as the target branch.
  --wip <yes/no>            Limit search based on WIP status.
                            Valid values:  yes  true  no  false
                            The values  yes  and  true  are equivalent.
                            The values  no  and  false  are equivalent.
  --author <author>         Limit search to a specific author.
                            If the parameter is all numbers, it is assumed to be a user id.
                            In that case, this is a shortcut for  --author-id <user id>
                            Otherwise, this is a shortcut for  --author-username <username>
  --author-id <user id>     Limit search to a specific author based on a user id.
  --author-username <username>
                            Limit search to a specific author based on a username.
  --assignee-id <user id>   Limit search to a specific approver based on a user id.
  --approver-ids <user id list>
                            Limit search to a MRs that have all of the given user ids listed as approvers.
                            The list should be comma delimited without spaces.
                            Max 5 user ids.
                            The parameter can also be either "None" or "Any".
                            A value of "None" will return MRs that do not have any approvers.
                            A value of "Any" will return MRs that have one or more approvers.
  --approved-by-ids <user id list>
                            Limit search to a MRs that have been approved by all of the given user ids.
                            The list should be comma separated without spaces.
                            Max 5 user ids.
                            A value of "None" will return MRs with no approvals.
                            A value of "Any" will return MRs with at least one approval (by anyone).
  --labels <labels>         Limit search to MRs matching the provided labels.
                            Labels should be comma separated without spaces.
                            A value of "None" will return MRs with no labels.
                            A value of "Any" will return MRs that have one or more labels.
  --milestone <milestone>   Limit search to specific milestones.
                            A value of "None" will return MRs with no milestones.
                            A value of "Any" will return MRs with any milestones.
  --my-reaction-emoji <emoji>
                            Limit search to those that you have given a certain reaction emoji to.
                            A value of "None" will return MRs that you have not reacted to.
                            A value of "Any" will return MRs that you have given any reaction emoji to.

Expert features:
  --with-labels-details <yes/no>
                            Add label details to the results.
                            This option only matters in conjunction with the --json flag.
                            The standard output from this utility will not contain the extra information.
                            Valid values:  yes  true  no  false
                            The values  yes  and  true  are equivalent.
                            The values  no  and  false  are equivalent.
                            If a <yes/no> value is not provided, then it will default to  true
  --without-labels-details  Shortcut for  --with-label_details false
  --view <view>             Set the type of view to return.
                            This option only matters when viewing the raw API results.
                            The standard output from this utility will not contain the extra information.
                            Valid values:  simple  normal
                            Default behavior is  --view normal
  --view-simple             Shortcut for  --view simple
  --view-normal             Shortcut for  --view normal

EOF
    )"
    local option_raw option option_arg option_arg_2 option_arg_raw option_arg_2_raw option_arg_temp
    local go_deep bypass_ignore do_selector do_json use_last_results verbose
    local arg_order_by arg_sort
    local arg_state arg_scope arg_created_after arg_created_before arg_updated_after arg_updated_before
    local arg_search arg_source_branch arg_target_branch arg_wip
    local arg_author_id arg_author_username arg_assignee_id arg_approver_ids arg_approved_by_ids
    local arg_labels arg_milestone arg_my_reaction_emoji
    local arg_with_labels_details arg_view
    if [[ "$#" -eq '0' ]]; then
        echo "$usage"
        return 0
    fi
    while [[ "$#" -gt '0' ]]; do
        option_raw="$1"
        shift
        option="$( printf %s "$option_raw" | __to_lowercase | sed 's/_/-/;' )"
        option_arg=
        option_arg_2=
        option_arg_raw=
        option_arg_2_raw=
        option_arg_temp=
        case "$option" in
        # Behavior options
        -h|--help)
            echo "$usage"
            return 0
            ;;
        -d|--deep)
            go_deep='YES'
            ;;
        -b|--bypass-ignore)
            bypass_ignore='YES'
            ;;
        -s|--select)
            do_selector='YES'
            ;;
        --json)
            do_json='YES'
            ;;
        --use-last-results)
            use_last_results='YES'
            ;;
        -v|--verbose)
            verbose='YES'
            ;;
        # Result control options
        --order-by)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase | sed 's/-/_/g;' )"
            if [[ "$option_arg" == 'created_at' || "$option_arg" == 'updated_at' ]]; then
                arg_order_by="$option_arg"
            elif [[ "$option_arg" == 'created' || "$option_arg" == 'updated' ]]; then
                arg_order_by="${option_arg}_at"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  created_at  updated_at."
                return 1
            fi
            ;;
        --order-by-created|--order-by-created-at)
            arg_order_by="created_at"
            ;;
        --order-by-updated|--order-by-updated_at)
            arg_order_by="updated_at"
            ;;
        --sort)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'asc' || "$option_arg" == 'desc' ]]; then
                arg_sort="$option_arg"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  asc  desc."
                return 1
            fi
            ;;
        --asc|--sort-asc)
            arg_sort="asc"
            ;;
        --desc|--sort-desc)
            arg_sort="desc"
            ;;
        # Search options
        --state)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'opened' || "$option_arg" == 'closed' || "$option_arg" == 'locked' || "$option_arg" == 'merged' ]]; then
                arg_state="$option_arg"
            elif [[ "$option_arg" == 'open' ]]; then
                arg_state='opened'
            elif [[ "$option_arg" == 'close' ]]; then
                arg_state='closed'
            elif [[ "$option_arg" == 'lock' ]]; then
                arg_state='locked'
            elif [[ "$option_arg" == 'merge' ]]; then
                arg_state='merged'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  opened  closed  locked  merged."
                return 1
            fi
            ;;
        --opened|--state-opened)
            arg_state='opened'
            ;;
        --closed|--state-closed)
            arg_state='closed'
            ;;
        --locked|--state-locked)
            arg_state='locked'
            ;;
        --merged|--state-merged)
            arg_state='merged'
            ;;
        --scope)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase | sed 's/-/_/g;' )"
            if [[ "$option_arg" == 'created_by_me' || "$option_arg" == 'assigned_to_me' || "$option_arg" == 'all' ]]; then
                arg_scope="$option_arg"
            elif [[ "$option_arg" == 'created' ]]; then
                arg_scope='created_by_me'
            elif [[ "$option_arg" == 'assigned' ]]; then
                arg_scope='assigned_to_me'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  created_by_me  assigned_to_me  all."
                return 1
            fi
            ;;
        --created-by-me)
            arg_scope='created_by_me'
            ;;
        --assigned-to-me)
            arg_scope='assigned_to_me'
            ;;
        --scope-all)
            arg_scope='all'
            ;;
        --created-after)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00:00"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg 00:00:00"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --created-before)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_before="$option_arg"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_before="$option_arg:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_created_before="$option_arg:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_created_before="$option_arg 23:59:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --created-between)
            if [[ -z "$1" || -z "$2" || "$1" =~ ^- || "$2" =~ ^- ]]; then
                >&2 echo -E "The $option option requires two parameters."
                return 1
            fi
            option_arg_raw="$1"
            option_arg_2_raw="$2"
            shift
            shift
            if [[ -n "$1" && ! "$1" =~ ^- ]]; then
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    >&2 echo -E "Both parameters provided to $option must have the same format."
                    return 1
                else
                    option_arg_raw="$option_arg_raw $option_arg_2_raw"
                    option_arg_2_raw="$1 $2"
                    shift
                    shift
                fi
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            option_arg_2="$( printf %s "$option_arg_2_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$( printf %s "$option_arg" | tr '[:digit:]' 'd' )" != "$( printf %s "$option_arg_2" | tr '[:digit:]' 'd' )" ]]; then
                >&2 echo -E "Both parameters provided to $option must have the same format."
                return 1
            fi
            # We now know they both have the same format.
            if [[ "$option_arg" > "$option_arg_2" ]]; then
                option_arg_temp="$option_arg"
                option_arg="$option_arg_2"
                option_arg_2="$option_arg_temp"
            fi
            # We now know they both have the same format, and option_arg is before option_arg_2
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg"
                arg_created_before="$option_arg_2"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00"
                arg_created_before="$option_arg_2:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00:00"
                arg_created_before="$option_arg_2:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg 00:00:00"
                arg_created_before="$option_arg_2 23:59:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw] [$option_arg_2_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --created-on)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg 00:00:00"
                arg_created_before="$option_arg 23:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00:00"
                arg_created_before="$option_arg:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_created_after="$option_arg:00"
                arg_created_before="$option_arg:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD."
                return 1
            fi
            ;;
        --updated-after)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00:00"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg 00:00:00"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --updated-before)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_before="$option_arg"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_before="$option_arg:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_updated_before="$option_arg:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_updated_before="$option_arg 23:59:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --updated-between)
            if [[ -z "$1" || -z "$2" || "$1" =~ ^- || "$2" =~ ^- ]]; then
                >&2 echo -E "The $option option requires two parameters."
                return 1
            fi
            option_arg_raw="$1"
            option_arg_2_raw="$2"
            shift
            shift
            if [[ -n "$1" && ! "$1" =~ ^- ]]; then
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    >&2 echo -E "Both parameters provided to $option must have the same format."
                    return 1
                else
                    option_arg_raw="$option_arg_raw $option_arg_2_raw"
                    option_arg_2_raw="$1 $2"
                    shift
                    shift
                fi
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            option_arg_2="$( printf %s "$option_arg_2_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$( printf %s "$option_arg" | tr '[:digit:]' 'd' )" != "$( printf %s "$option_arg_2" | tr '[:digit:]' 'd' )" ]]; then
                >&2 echo -E "Both parameters provided to $option must have the same format."
                return 1
            fi
            # We now know they both have the same format.
            if [[ "$option_arg" > "$option_arg_2" ]]; then
                option_arg_temp="$option_arg"
                option_arg="$option_arg_2"
                option_arg_2="$option_arg_temp"
            fi
            # We now know they both have the same format, and option_arg is before option_arg_2
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg"
                arg_updated_before="$option_arg_2"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00"
                arg_updated_before="$option_arg_2:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00:00"
                arg_updated_before="$option_arg_2:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg 00:00:00"
                arg_updated_before="$option_arg_2 23:59:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw] [$option_arg_2_raw]. Use format:  YYYY-MM-DD hh:mm:ss."
                return 1
            fi
            ;;
        --updated-on)
            __ensure_option "$1" "$option" || return 1
            if [[ -z "$2" || "$2" =~ ^- ]]; then
                option_arg_raw="$1"
                shift
            else
                option_arg_raw="$1 $2"
                shift
                shift
            fi
            option_arg="$( printf %s "$option_arg_raw" | sed -E 's/[T ]+/ /; s/[^[:digit:] :]/-/g;' )"
            if [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg 00:00:00"
                arg_updated_before="$option_arg 23:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00:00"
                arg_updated_before="$option_arg:59:59"
            elif [[ "$option_arg" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
                arg_updated_after="$option_arg:00"
                arg_updated_before="$option_arg:59"
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Use format:  YYYY-MM-DD."
                return 1
            fi
            ;;
        --search)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            while [[ -n "$1" && ! "$1" =~ ^- ]]; do
                option_arg_raw="$option_arg_raw $1"
                shift
            done
            arg_search="$( printf %s "$option_arg_raw" | __url_encode )"
            ;;
        --source-branch)
            __ensure_option "$1" "$option" || return 1
            arg_source_branch="$( printf %s "$1" | __url_encode )"
            shift
            ;;
        --target-branch)
            __ensure_option "$1" "$option" || return 1
            arg_target_branch="$( printf %s "$1" | __url_encode )"
            shift
            ;;
        --wip)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'yes' || "$option_arg" == 'true' ]]; then
                arg_wip='yes'
            elif [[ "$option_arg" == 'no' || "$option_arg" == 'false' ]]; then
                arg_wip='no'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  yes  true  no  false."
                return 1
            fi
            ;;
        --author)
            __ensure_option "$1" "$option" || return 1
            if [[ "$1" =~ ^[[:digit:]]+$ ]]; then
                arg_author_id="$1"
                arg_author_username=
            else
                arg_author_id=
                arg_author_username="$( printf %s "$1" | __url_encode )"
            fi
            shift
            ;;
        --author-id)
            __ensure_option "$1" "$option" || return 1
            if [[ "$1" =~ ^[[:digit:]]+$ ]]; then
                arg_author_id="$1"
                arg_author_username=
            else
                >&2 echo -E "Invalid $option: [$1]. Value must be a number."
                return 1
            fi
            shift
            ;;
        --author-username)
            __ensure_option "$1" "$option" || return 1
            arg_author_id=
            arg_author_username="$( printf %s "$1" | __url_encode )"
            shift
            ;;
        --assignee-id)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" =~ ^[[:digit:]]+$ ]]; then
                arg_assignee_id="$option_arg"
            elif [[ "$option_arg" == 'none' ]]; then
                arg_assignee_id='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_assignee_id='Any'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Value must be a number or else  None  or  Any."
                return 1
            fi
            ;;
        --approver-ids)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" =~ ^[[:digit:]]+(,[[:digit:]]+){0,4}$ ]]; then
                arg_approver_ids="$option_arg"
            elif [[ "$option_arg" == 'none' ]]; then
                arg_approver_ids='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_approver_ids='Any'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Value must be a comma separated list of up to five numbers or else  None  or  Any."
                return 1
            fi
            ;;
        --approved-by-ids)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" =~ ^[[:digit:]]+(,[[:digit:]]+){0,4}$ ]]; then
                arg_approved_by_ids="$option_arg"
            elif [[ "$option_arg" == 'none' ]]; then
                arg_approved_by_ids='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_approved_by_ids='Any'
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Value must be a comma separated list of up to five numbers or else  None  or  Any."
                return 1
            fi
            ;;
        --labels)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'none' ]]; then
                arg_labels='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_labels='Any'
            else
                arg_labels="$( printf %s "$option_arg_raw" | __url_encode )"
            fi
            ;;
        --milestone)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'none' ]]; then
                arg_milestone='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_milestone='Any'
            else
                arg_milestone="$( printf %s "$option_arg_raw" | __url_encode )"
            fi
            ;;
        --my-reaction-emoji)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'none' ]]; then
                arg_my_reaction_emoji='None'
            elif [[ "$option_arg" == 'any' ]]; then
                arg_my_reaction_emoji='Any'
            else
                arg_my_reaction_emoji="$( printf %s "$option_arg_raw" | __url_encode )"
            fi
            ;;
        # Expert options
        --with-labels-details)
            if [[ -z "$1" || "$1" =~ ^- ]]; then
                arg_with_labels_details='true'
            else
                option_arg_raw="$1"
                shift
                option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
                if [[ "$option_arg" == 'yes' || "$option_arg" == 'true' ]]; then
                    arg_with_labels_details='true'
                elif [[ "$option_arg" == 'no' || "$option_arg" == 'false' ]]; then
                    arg_with_labels_details='false'
                else
                    >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  yes  true  no  false."
                    return 1
                fi
            fi
            ;;
        --without-labels-details)
            arg_with_labels_details='false'
            ;;
        --view)
            __ensure_option "$1" "$option" || return 1
            option_arg_raw="$1"
            shift
            option_arg="$( printf %s "$option_arg_raw" | __to_lowercase )"
            if [[ "$option_arg" == 'simple' ]]; then
                arg_view="simple"
            elif [[ "$option_arg" == 'normal' ]]; then
                arg_view=
            else
                >&2 echo -E "Invalid $option: [$option_arg_raw]. Valid options:  simple  normal."
                return 1
            fi
            ;;
        --view-simple)
            arg_view='simple'
            ;;
        --view-normal)
            arg_view=
            ;;
        # Default
        *)
            >&2 echo -E "Invalid option: [$option_raw]."
            return 1
            ;;
        esac
    done
    # All provided options/arguments have been parsed.
    # Set everything up as api parameters and create the desired query string.
    local parameters datetime_format arg_approver_id arg_approved_by_id query_string
    parameters=()
    datetime_format='+%FT%T%z'
    if [[ -n "$arg_order_by" ]]; then
        parameters+=( "order_by=$arg_order_by" )
    fi
    if [[ -n "$arg_sort" ]]; then
        parameters+=( "sort=$arg_sort" )
    fi
    if [[ -n "$arg_state" ]]; then
        parameters+=( "state=$arg_state" )
    fi
    if [[ -n "$arg_scope" ]]; then
        parameters+=( "scope=$arg_scope" )
    else
        parameters+=( "scope=all" )
    fi
    if [[ -n "$arg_created_after" ]]; then
        parameters+=( "created_after=$( date -j -f '%F %T' "$arg_created_after" "$datetime_format" )" )
    fi
    if [[ -n "$arg_created_before" ]]; then
        parameters+=( "created_before=$( date -j -f '%F %T' "$arg_created_before" "$datetime_format" )" )
    fi
    if [[ -n "$arg_updated_after" ]]; then
        parameters+=( "updated_after=$( date -j -f '%F %T' "$arg_updated_after" "$datetime_format" )" )
    fi
    if [[ -n "$arg_updated_before" ]]; then
        parameters+=( "updated_before=$( date -j -f '%F %T' "$arg_updated_before" "$datetime_format" )" )
    fi
    if [[ -n "$arg_search" ]]; then
        parameters+=( "search=$arg_search" )
    fi
    if [[ -n "$arg_source_branch" ]]; then
        parameters+=( "source_branch=$arg_source_branch" )
    fi
    if [[ -n "$arg_target_branch" ]]; then
        parameters+=( "target_branch=$arg_target_branch" )
    fi
    if [[ -n "$arg_wip" ]]; then
        parameters+=( "wip=$arg_wip" )
    fi
    if [[ -n "$arg_author_id" ]]; then
        parameters+=( "author_id=$arg_author_id" )
    elif [[ -n "$arg_author_username" ]]; then
        parameters+=( "author_username=$arg_author_username" )
    fi
    if [[ -n "$arg_assignee_id" ]]; then
        parameters+=( "assignee_id=$arg_assignee_id" )
    fi
    if [[ -n "$arg_approver_ids" ]]; then
        for arg_approver_id in $( printf %s "$arg_approver_ids" | tr ',' ' ' ); do
            parameters+=( "approver_ids\[\]=$arg_approver_id" )
        done
    fi
    if [[ -n "$arg_approved_by_ids" ]]; then
        for arg_approved_by_id in $( printf %s "$arg_approved_by_ids" | tr ',' ' ' ); do
            parameters+=( "approved_by_ids\[\]=$arg_approved_by_id" )
        done
    fi
    if [[ -n "$arg_labels" ]]; then
        parameters+=( "labels=$arg_labels" )
    fi
    if [[ -n "$arg_milestone" ]]; then
        parameters+=( "milestone=$arg_milestone" )
    fi
    if [[ -n "$arg_my_reaction_emoji" ]]; then
        parameters+=( "my_reaction_emoji=$arg_my_reaction_emoji" )
    fi
    if [[ -n "$arg_with_labels_details" ]]; then
        parameters+=( "with_labels_details=$arg_with_labels_details" )
    fi
    if [[ -n "$arg_view" ]]; then
        parameters+=( "view=$arg_view" )
    fi
    query_string="$( __gl_join '&' "${parameters[@]}" )"
    local json_results result_count
    if [[ -z "$use_last_results" ]]; then
        if [[ -n "$go_deep" ]]; then
            __get_gitlab_mrs_deep '' "$bypass_ignore" "$query_string" "$verbose"
            GITLAB_MRS_SEARCH_RESULTS="$GITLAB_MRS_DEEP_RESULTS"
        else
            GITLAB_MRS_SEARCH_RESULTS="$( __get_pages_of_url "$( __get_gitlab_url_mrs )?${query_string}&" '' '' "$verbose" )"
        fi
    fi
    json_results="$GITLAB_MRS_SEARCH_RESULTS"
    result_count="$( jq ' length ' <<< "$json_results" )"
    # Add the project name to each result.
    local project_id project_name
    for project_id in $( jq -r ' [ .[] | .project_id ] | unique | .[] ' <<< "$json_results" ); do
        project_name="$( __get_project_name "$project_id" )"
        json_results="$( jq -c --arg project_id "$project_id" --arg project_name "$project_name" \
                        ' [ .[] | if (.project_id == ($project_id|tonumber)) then (.project_name = $project_name) else . end ] ' <<< "$json_results" )"
    done
    if [[ -n "$do_json" ]]; then
        jq --sort-keys '.' <<< "$json_results"
    else
        echo -E -n "Found $result_count result"
        [[ "$result_count" -ne '1' ]] && echo -E -n 's'
        echo -E "."
        if [[ "$result_count" -ge '1' ]]; then
            ( echo -E '┌──▪ State~┌──▪ Date~┌──▪ Repo~┌──▪ Author~┌──▪ Title~┌──▪ Url' \
                && jq -r \
                        ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                          def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                          .[] | .col_head = "├▪ "
                              |         .col_head + .state
                                + "~" + .col_head + ( .merged_at // .closed_at // .created_at | .[0:10] )
                                + "~" + .col_head + ( .project_name | clean )
                                + "~" + .col_head + ( .author.name | cleanname | .[0:20] )
                                + "~" + .col_head + ( .title | .[0:30] | clean )
                                + "~" + .col_head + .web_url ' \
                        <<< "$json_results" ) \
                | sed '$s/├/└/g' \
                | column -s '~' -t
        fi
    fi
    if [[ -n $do_selector && "$result_count" -ge '1' ]]; then
        local selected_lines selected_line web_url
        selected_lines="$( ( echo -E "State~ Date~ Repo~ Author~ Title" \
            && jq -r \
                    ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                      def cleanname: sub(" - [sS][oO][fF][iI].*$"; "") | clean;
                      .[] |         .state
                            + "~" + ( .merged_at // .closed_at // .created_at | .[0:10] )
                            + "~" + ( .project_name | clean )
                            + "~" + ( .author.name | cleanname | .[0:20] )
                            + "~" + ( .title | .[0:80] | clean )
                            + "~" + .web_url ' \
                    <<< "$json_results" ) \
            | fzf_wrapper --tac --header-lines=1 --cycle --with-nth=1,2,3,4,5 --delimiter="~" -m --to-columns )"
        echo -E "$selected_lines" | while read selected_line; do
            web_url="$( echo -E "$selected_line" | __gitlab_get_col '~' '6' )"
            if [[ -n $web_url ]]; then
                open "$web_url"
            fi
        done
    fi
}
