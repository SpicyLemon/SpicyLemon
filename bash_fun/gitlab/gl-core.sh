#!/bin/bash
# This file contains many functions needed for the various gitlab cli functions.
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

# GitLab API documentation: https://docs.gitlab.com/ee/api/api_resources.html

# Naming conventions:
# * All functions in here should start with __gl_
# * Functions names should be all lowercase.
# * Functions should use snake_case.
# * Functions starting with "__gl_get_" actually do HTTP GETs against the api.
# * Functions can only have "_get_" in their name as part of "__gl_get_".
# * Functions starting with "__gl_post_" actually do HTTP POSTs against the api.
# * Any function interacting with the gitlab api directly should start with either "__gl_get_" or "__gl_post_".
#       Exception: If a HTTP method besides GET or POST is used, the function should start with "__gl_{HTTP method}_" (lower-cased).
# * Functions starting with "__gl_url_" create urls.
# * Functions starting with "__gl_url_api_" create a url specifically for use with the api.
# * Functions starting with "__gl_url_web_" create a url specifically for web browser use.
# * In a function name, "filename" or "dirname" references the name/path of the file or directory,
#       and "file" or "dir" references the actual file or directory.
#       Examples: The __gl_ensure_temp_dir function makes sure that the actual temp directory exists.
#                 The __gl_temp_dirname function returns the path and name of the temp directory.

#
# Generic Helpers
#

# Usage: __gl_bold_white <text>
__gl_bold_white () {
    echo -e -n "\033[1;37m$1\033[0m"
}

# Usage: __gl_bold_yellow <text>
__gl_bold_yellow () {
    echo -e -n "\033[1;33m$1\033[0m"
}

# Converts a string to lowercase.
# Usage: echo 'FOO' | __gl_lowercase
__gl_lowercase () {
    if [[ "$#" -gt '0' ]]; then
        printf '%s' "$*" | __gl_lowercase
        return 0
    fi
    tr "[:upper:]" "[:lower:]"
}

# Usage: __gl_encode_for_url "value to encode"
#  or    <do stuff> | __gl_encode_for_url
__gl_encode_for_url () {
    if [[ "$#" -gt '0' ]]; then
        printf '%s' "$*" | __gl_encode_for_url
        return 0
    fi
    jq -sRr @uri
}

# Joins all provided parameters using the provided delimiter.
# Usage: __gl_join <delimiter> [<arg1> [<arg2>... ]]
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
# Usage: __gl_require_option "$2" "option name" || echo "bad option."
__gl_require_option () {
    local value option
    value="$1"
    option="$2"
    if [[ -z "$value" || "$value" =~ ^- ]]; then
        >&2 echo -E "A parameter must be supplied with the $option option."
        return 1
    fi
    return 0
}

# Usage: __gl_count_non_empty <val1> [<val2> [<val3> ...]]
__gl_count_non_empty () {
    local count
    count=0
    while [[ "$#" -gt '0' ]]; do
        if [[ -n "$1" ]]; then
            count=$(( count + 1 ))
        fi
        shift
    done
    echo -E -n "$count"
}

# Makes sure that a provided value is a number between the min and max (inclusive).
# Only whole numbers are allowed.
# The min and max parameters are interchangable. If larger of the two is the max, and the other is the min.
# If the provided value is not a number, then the default is returned.
# If it's less than the min, the min is returned.
# If it's less than the max, the max is returned.
# Usage: __gl_clamp <value> <min> <max> <default>
__gl_clamp () {
    local val min max default result
    val="$( __gl_number_or_default "$1" "" )"
    min="$( __gl_number_or_default "$2" "" )"
    max="$( __gl_number_or_default "$3" "" )"
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
# Usage: __gl_number_or_default <value> <default>
__gl_number_or_default () {
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

# Usage: <do stuff> | __gl_column_value <delimiter> <column index>
__gl_column_value () {
    awk -v d="$1" -v col="$2" '{split($0, a, d); print a[col]}'
}

# Usage: <do stuff> | __gl_convert_display_options_to_auto_options
__gl_convert_display_options_to_auto_options () {
    if [[ -n "$1" ]]; then
        echo -E "$1" | __gl_convert_display_options_to_auto_options
        return 0
    fi
    sed -E 's/<[^>]+>//g; s/\[|\]|\(|\)//g; s/\|/ /g; s/[[:space:]][[:space:]]+/ /g; s/^[[:space:]]//; s/[[:space:]]$//;'
}

#
# Environment Data Management
#

# Makes sure that a GitLab token is set.
# This must be set outside of this file, and is kind of a secret thing.
# Usage: __gl_require_token
__gl_require_token () {
    if [[ -z "$GITLAB_PRIVATE_TOKEN" ]]; then
        >&2 cat << EOF
No GITLAB_PRIVATE_TOKEN has been set.
To create one, go to $( __gl_url_base )/profile/personal_access_tokens and create one with the "api" scope.
Then you can set it using
GITLAB_PRIVATE_TOKEN=whatever-your-token-is
It is probably best to put that line somewhere so that it will get executed whenever you start your terminal (e.g. .bash_profile)

EOF
        return 1
    fi
}

# Makes sure that your GitLab user info has been loaded.
# If not, it is looked up.
# Usage: __gl_ensure_user_info <keep quiet>
__gl_ensure_user_info () {
    local keep_quiet
    keep_quiet="$1"
    if [[ -z "$GITLAB_USER_ID" || -z "$GITLAB_USERNAME" ]]; then
        __gl_get_user_info "$keep_quiet"
    fi
}

__gl_ensure_config_dir () {
    local config_dir
    config_dir="$( __gl_config_dirname )"
    if [[ -z "$config_dir" ]]; then
        >&2 echo "No configuration directory defined."
        return 1
    elif [[ -d "$config_dir" ]]; then
        return 0
    fi
    mkdir -p "$config_dir"
}

# Makes sure that the $GITLAB_PROJECTS variable has a value.
# A temp file is used to store the project info too.
# If the file doesn't exist, or is older than a day, or is empty,
#   the projects info will be refreshed and stored in the file.
# Otherwise, it's contents will be loaded into the $GITLAB_PROJECTS variable.
# Usage: __gl_ensure_projects <keep quiet> <verbose>
__gl_ensure_projects () {
    local keep_quiet verbose projects_file
    keep_quiet="$1"
    verbose="$2"
    __gl_ensure_temp_dir
    projects_file="$( __gl_temp_projects_filename )"
    if [[ ! -f "$projects_file" \
            || $( find "$projects_file" -mtime "+$( __gl_max_age_projects )" ) ]] \
            || ! $( grep -q '[^[:space:]]' "$projects_file" ); then
        __gl_get_projects "$keep_quiet" "$verbose"
        echo -E "$GITLAB_PROJECTS" > "$projects_file"
    else
        GITLAB_PROJECTS="$( cat "$projects_file" )"
    fi
}

__gl_max_age_projects () {
    __gl_cache_max_age "$GITLAB_PROJECTS_MAX_AGE" 'GITLAB_PROJECTS_MAX_AGE'
}

__gl_projects_clear_cache () {
    local projects_file
    projects_file="$( __gl_temp_projects_filename )"
    if [[ -f "$projects_file" ]]; then
        rm "$projects_file"
    fi
}

__gl_projects_refresh_cache () {
    __gl_projects_clear_cache
    __gl_ensure_projects
}

# Makes sure that the $GITLAB_GROUPS variable has a value.
# A temp file is used to store the groups info too.
# If the file doesn't exist, or is older than a day, or is empty,
#   the groups info will be refreshed and stored in the file.
# Otherwise, it's contents will be loaded into the $GITLAB_GROUPS variable.
# Usage: __gl_ensure_groups <keep quiet> <verbose>
__gl_ensure_groups () {
    local keep_quiet verbose groups_file
    keep_quiet="$1"
    verbose="$2"
    __gl_ensure_temp_dir
    groups_file="$( __gl_temp_groups_filename )"
    if [[ ! -f "$groups_file" \
            || $( find "$groups_file" -mtime "+$( __gl_max_age_groups )" ) ]] \
            || ! $( grep -q '[^[:space:]]' "$groups_file" ); then
        __gl_get_groups "$keep_quiet" "$verbose"
        echo -E "$GITLAB_GROUPS" > "$groups_file"
    else
        GITLAB_GROUPS="$( cat "$groups_file" )"
    fi
}

__gl_max_age_groups () {
    __gl_cache_max_age "$GITLAB_GROUPS_MAX_AGE" 'GITLAB_GROUPS_MAX_AGE'
}

__gl_groups_clear_cache () {
    local groups_file
    groups_file="$( __gl_temp_groups_filename )"
    if [[ -f "$groups_file" ]]; then
        rm "$groups_file"
    fi
}

__gl_groups_refresh_cache () {
    __gl_groups_clear_cache
    __gl_ensure_groups
}

__gl_cache_default_max_age () {
    __gl_cache_max_age "$GITLAB_CACHE_DEFAULT_MAX_AGE" 'GITLAB_CACHE_DEFAULT_MAX_AGE'
}

# Usage: if __gl_is_valid_atime "<string>"; then
__gl_is_valid_atime () {
    [[ "$1" =~ ^([[:digit:]]+[smhdw])+$ ]]
}

# Usage: max_age="$( __gl_cache_max_age <value> <variable name> )"
__gl_cache_max_age () {
    local value name retval invalid_value
    value="$1"
    name="$2"
    if [[ -n "$value" ]]; then
        if __gl_is_valid_atime "$value"; then
            retval="$value"
        else
            invalid_value='YES'
        fi
    elif [[ "$name" == 'GITLAB_CACHE_DEFAULT_MAX_AGE' ]]; then
        retval='23h'
    else
        retval="$( __gl_cache_default_max_age )"
    fi
    if [[ -n "$invalid_value" ]]; then
        printf 'Invalid %s value [%s]. Using default of %s.\n' "$name" "$value" "$retval" >&2
    fi
    printf '%s' "$retval"
    return 0
}

# Makes sure that the gitlab temp directory exists.
# Usage: __gl_ensure_temp_dir
__gl_ensure_temp_dir () {
    local tmp_dir
    tmp_dir="$( __gl_temp_dirname )"
    if [[ -f "$tmp_dir" ]]; then
        rm "$tmp_dir"
    fi
    if [[ ! -d "$tmp_dir" ]]; then
        mkdir "$tmp_dir"
    fi
}

__gl_config_dirname () {
    if [[ -n "$GITLAB_CONFIG_DIR" && "$GITLAB_CONFIG_DIR" =~ ^/ ]]; then
        echo -E -n "$GITLAB_CONFIG_DIR"
        return 0
    elif [[ -n "$HOME" && "$HOME" =~ ^/ ]]; then
        echo -E -n "$HOME/.config/gitlab"
        return 0
    elif [[ -n "$GITLAB_REPO_DIR" && "$GITLAB_REPO_DIR" =~ ^/ ]]; then
        echo -E -n "$GITLAB_REPO_DIR/.gitlab_config"
        return 0
    elif [[ -n "$GITLAB_BASE_DIR" && "$GITLAB_BASE_DIR" =~ ^/ ]]; then
        # The GITLAB_BASE_DIR environment variable is deprecated in favor of GITLAB_REPO_DIR.
        echo -E -n "$GITLAB_BASE_DIR/.gitlab_config"
        return 0
    fi
    return 1
}

__gl_config_gmrignore_filename () {
    echo -E -n "$( __gl_config_dirname )/gmr_ignore.json"
}

__gl_temp_dirname () {
    if [[ -n "$GITLAB_TEMP_DIR" && "$GITLAB_TEMP_DIR" =~ ^/ ]]; then
        echo -E -n "$GITLAB_TEMP_DIR"
    else
        echo -E -n '/tmp/gitlab'
    fi
}

# Gets the full path and name of the file to store projects info.
# Usage: __gl_temp_projects_filename
__gl_temp_projects_filename () {
    echo -E -n "$( __gl_temp_dirname )/projects.json"
}

# Gets the full path and name of the file to store groups info.
# Usage: __gl_temp_groups_filename
__gl_temp_groups_filename () {
    echo -E -n "$( __gl_temp_dirname )/groups.json"
}

#
# Data Manipulation
#

# This can be used to get a subset of $GITLAB_PROJECTS based on some searches, or usage of fzf.
# If <force select> has value:
#   * fzf will be used to prompt the user to select projects
# If <force select> does not have value:
#   * The <provided repos> are looked for, matching by name in the $GITLAB_PROJECTS data.
#   * If no <provided repos> are provided, the <current repo> is looked for.
#   * If any <provided repos> that aren't found exactly, fzf will prompt the user to select projects.
#       The projects that weren't found are used as an initial query for fzf.
# Usage: __gl_project_subset <force select> <current repo> <provided repos>
__gl_project_subset () {
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
            project="$( echo -E "$GITLAB_PROJECTS" | jq -c --arg search "$search" \
                        ' [ .[] | select( ( .name | ascii_downcase ) == ( $search | ascii_downcase )
                                     or ( .path | ascii_downcase ) == ( $search | ascii_downcase ) ) ]
                          | if ( length == 1 ) then .[0] else empty end ' )"
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
            | jq -r ' def clean: gsub("[\\n\\t]"; " ") | gsub("\\p{C}"; "") | gsub("~"; "-");
                      sort_by(.name_with_namespace) | .[]
                        |         ( .name_with_namespace | clean )
                          + "~" + ( .id | tostring ) ' \
            | fzf_wrapper --tac --cycle --with-nth=1 --delimiter="~" -m -i --query="$fzf_search" --to-columns \
            | __gl_column_value '~' '2' )"
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
# Usage: __gl_project_by_name <repo>
__gl_project_by_name () {
    __gl_project_subset '' '' "$*"
}

# Look up a project name from its id.
# Usage: __gl_project_name <project id>
__gl_project_name () {
    local project_id project_name
    project_id="$1"
    project_name=$( echo -E "$GITLAB_PROJECTS" | jq -r " .[] | select(.id==$project_id) | .name " )
    echo -E -n "$project_name"
}

# Adds the .project_name parameter to the entries in $GITLAB_MRS_BY_ME.
# Usage: __gl_mrs_i_created_add_project_names
__gl_mrs_i_created_add_project_names () {
    local mr_project_ids mr_project_id project_name
    mr_project_ids="$( echo -E "$GITLAB_MRS_BY_ME" | jq ' [ .[] | .project_id ] | unique | .[] ' )"
    for mr_project_id in $( echo -E "$mr_project_ids" | sed -l '' ); do
        project_name="$( __gl_project_name "$mr_project_id" )"
        GITLAB_MRS_BY_ME="$( echo -E "$GITLAB_MRS_BY_ME" | jq -c " [ .[] | if (.project_id == $mr_project_id) then (.project_name = \"$project_name\") else . end ] " )"
    done
}

# Usage: <project name> <url> <branch> <diff branch> <page name>
__gl_glopen_create_message_entry () {
    local project_name url branch diff_branch page_name cols
    project_name="$1"
    url="$2"
    branch="$3"
    diff_branch="$4"
    page_name="$5"
    cols=()
    if [[ -n "$branch" && -n "$diff_branch" && "$branch" != "$diff_branch" ]]; then
        cols+=( "$diff_branch to $branch" "in" )
    elif [[ -n "$branch" ]]; then
        cols+=( "$branch" "in" )
    elif [[ -n "$page_name" ]]; then
        cols+=( "$page_name page" "of" )
    else
        cols+=( "main page" "of" )
    fi
    cols+=( "$project_name:" )
    cols+=( "$url" )
    __gl_join "~" "${cols[@]}"
}

# Filter either $GITLAB_MRS_TODO or $GITLAB_MRS for only MRs where you are a suggested approver.
# The results are placed in $GIBLAB_MRS_TODO.
# This basically weeds out any MRs that either you don't need to care about, or you've already approved of.
# Usage: __gl_mrs_filter_by_approver <keep quiet> <filter type>
# If filter type is "SHORT" then $GITLAB_MRS_TODO is filtered. Otherwise $GITLAB_MRS is filtered.
__gl_mrs_filter_by_approver () {
    local keep_quiet filter_type mrs_to_filter mr_count mr_ids mr_index mr_todo_count my_mrs \
          mr_id mr mr_iid mr_project_id mr_project_name mr_approvals mr_state need_to_approve already_approved mr_approved to_add
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
        mr_project_name="$( __gl_project_name "$mr_project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rFiltering MRs: ($mr_todo_count) $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_approvals="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __gl_url_api_projects_mrs_approvals "$mr_project_id" "$mr_iid" )" )"
        mr_state="$( echo -E "$mr_approvals" | jq -r ' .state ' )"
        need_to_approve="$( echo -E "$mr_approvals" | jq -r --arg GITLAB_USER_ID "$GITLAB_USER_ID" ' .suggested_approvers[] | select(.id==($GITLAB_USER_ID|tonumber)) | "YES" ' )"
        already_approved="$( echo -E "$mr_approvals" | jq -r --arg GITLAB_USER_ID "$GITLAB_USER_ID" ' .approved_by[] | select(.user.id==($GITLAB_USER_ID|tonumber)) | "YES" ' )"
        if [[ -n "$need_to_approve$already_approved" && "$mr_state" == "opened" ]]; then
            mr_approved="$( echo -E "$mr_approvals" | jq -r ' if .approved then "YES" else "NO" end ' )"
            to_add="$( jq -c --arg mr_project_name "$mr_project_name" --arg mr_approved "$mr_approved" --arg already_approved "$already_approved" --null-input \
                    '{ project_name: $mr_project_name, approved: ($mr_approved == "YES"), i_approved: ($already_approved == "YES") }' )"
            mr="$( echo -E "$mr" | jq -c --argjson to_add "$to_add" ' . + $to_add ' )"
            my_mrs="$( echo -E "[$my_mrs,[$mr]]" | jq -c ' add ' )"
            mr_todo_count=$(( mr_todo_count + 1 ))
        fi
        mr_index=$(( mr_index + 1 ))
    done
    GITLAB_MRS_TODO="$( echo -E "$my_mrs" | jq -c ' sort_by(.created_at) ' )"
    [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r"
}

# Filters the entries in $GITLAB_JOBS to get just the ones applicable to the provided branch name.
# Usage: __gl_jobs_filter_by_branch <branch name>
__gl_jobs_filter_by_branch () {
    local branch_name
    branch_name="$1"
    if [[ -n "$branch_name" ]]; then
        GITLAB_JOBS="$( echo -E "$GITLAB_JOBS" | jq -c " [ .[] | select(.ref==\"$branch_name\") ] " )"
    fi
}

# Filters the entries in $GITLAB_JOBS based on a provided filter on the short_type parameter (added after getting all the jobs).
# The filter can be the short type to keep (e.g. "build") or the short_type to ignore (e.g. "~sdlc").
# Usage: __gl_jobs_filter_by_type <filter type>
__gl_jobs_filter_by_type () {
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

#
# GitLab Interaction
#

# Looks up your GitLab user info. Results are stored in $GITLAB_USER_ID and $GITLAB_USERNAME.
# Usage: __gl_get_user_info <keep quiet>
__gl_get_user_info () {
    local keep_quiet
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting your GitLab user id... "
    GITLAB_USER_INFO="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __gl_url_api_user )" )"
    GITLAB_USER_ID="$( echo -E "$GITLAB_USER_INFO" | jq '.id' )"
    GITLAB_USERNAME="$( echo -E "$GITLAB_USER_INFO" | jq '.username' | sed -E 's/^"|"$//g' )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Look up info on all available projects. Results are stored in $GITLAB_PROJECTS.
# Usage: __gl_get_projects <keep quiet> <verbose>
__gl_get_projects () {
    local keep_quiet verbose projects_url projects
    keep_quiet="$1"
    verbose="$2"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your GitLab projects... "
    projects_url="$( __gl_url_api_projects )?order_by=id&simple=true&membership=true&archived=false&"
    projects="$( __gl_get_all_results "$projects_url" '' '' "$verbose" 'use_keyset' )"
    GITLAB_PROJECTS="$projects"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Look up info on all available groups. Results are stored in $GITLAB_GROUPS.
# Usage: __gl_get_groups <keep quiet> <verbose>
__gl_get_groups () {
    local keep_quiet verbose groups_url groups
    keep_quiet="$1"
    verbose="$2"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your GitLab groups... "
    groups_url="$( __gl_url_api_groups )?order_by=id&"
    groups="$( __gl_get_all_results "$groups_url" '' '' "$verbose" )"
    GITLAB_GROUPS="$groups"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Gets all the open MRS that you've created. Results are stored in $GITLAB_MRS_BY_ME.
# Usage: __gl_get_mrs_i_created <keep quiet>
__gl_get_mrs_i_created () {
    local keep_quiet mrs_url mrs
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all open MRs you created... "
    mrs_url="$( __gl_url_api_mrs )?scope=created_by_me&state=opened&"
    mrs="$( __gl_get_all_results "$mrs_url" )"
    GITLAB_MRS_BY_ME="$( echo -E "$mrs" | jq -c ' sort_by(.source_branch, .project_id) ' )"
    __gl_mrs_i_created_add_project_names
    [[ -n "$keep_quiet" ]] || echo -E "Done."
    __gl_add_approved_status_to_mrs_i_created
}

# Do a superficial search for MRs. Results are put in $GITLAB_MRS.
# This is a quicker search than __gl_get_mrs_to_approve_deep, but often leaves MRs out of the list because of a bug in GitLab.
# Usage: __gl_get_mrs_to_approve_simple <keep quiet>
__gl_get_mrs_to_approve_simple () {
    local keep_quiet mrs_url mrs
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your open MRs... "
    mrs_url="$( __gl_url_api_mrs )?scope=all&state=opened&approver_ids\[\]=$GITLAB_USER_ID&"
    mrs="$( __gl_get_all_results "$mrs_url" )"
    GITLAB_MRS="$mrs"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Do a deep search for MRs by running the search on each project individually.
# This will often take a while, but will often find more MRs than __gl_get_mrs_to_approve_simple because of bugs in GitLab.
# If a custom query string is provided, then results will be placed in $GITLAB_MRS_DEEP_RESULTS.
# Otherwise, the search will be for state=opened, and results will be put in $GITLAB_MRS.
# Usage: __gl_get_mrs_to_approve_deep <keep quiet> <bypass ignore> <query_string> <verbose>
__gl_get_mrs_to_approve_deep () {
    local keep_quiet bypass_ignore query_string verbose is_custom_search ignore_list project_ids project_count
    local mrs mr_count project_index project_id project_name mrs_url project_mrs project_mr_count
    keep_quiet="$1"
    bypass_ignore="$2"
    query_string="$3"
    verbose="$4"
    if [[ -n "$query_string" ]]; then
        is_custom_search='YES'
    else
        query_string='state=opened'
    fi
    ignore_list='[]'
    if [[ -z "$bypass_ignore" ]]; then
        local ignore_fn
        ignore_fn="$( __gl_config_gmrignore_filename )"
        if [[ -r "$ignore_fn" ]] && grep -q '[^[:space:]]' "$ignore_fn"; then
            ignore_list="$( cat "$ignore_fn" )"
        fi
    fi
    project_ids="$( echo -E "$GITLAB_PROJECTS" | jq --argjson ignore_list "$ignore_list" ' [ .[] | .id ] - $ignore_list | .[] ' )"
    project_count="$( echo -E "$project_ids" | wc -l | sed -E 's/[^[:digit:]]//g' )"
    if [[ -z "$keep_quiet" ]]; then
        local all_project_count
        all_project_count="$( echo -E "$GITLAB_PROJECTS" | jq ' length ' )"
        if [[ -n "$is_custom_search" ]]; then
            echo -E -n "Doing deep search for MRs"
        else
            echo -E -n "Getting all your open MRs"
        fi
        echo -E -n " from $project_count projects"
        if [[ "$project_count" -ne "$all_project_count" ]]; then
            echo -E -n " (of $all_project_count)"
        fi
        echo -E '.'
    fi
    mrs="[]"
    mr_count=0
    project_index=1
    for project_id in $project_ids; do
        project_name="$( __gl_project_name "$project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r($mr_count) $project_index/$project_count - $project_id: $project_name "
        mrs_url="$( __gl_url_api_projects_mrs "$project_id" )?${query_string}&"
        project_mrs="$( __gl_get_all_results "$mrs_url" '' '' "$verbose" )"
        project_mr_count="$( echo -E "$project_mrs" | jq ' length ' )"
        if [[ "$project_mr_count" -gt "0" ]]; then
            mrs="$( echo -E "[$mrs,$project_mrs]" | jq -c ' add ')"
            mr_count=$(( mr_count + project_mr_count ))
        fi
        project_index=$(( project_index + 1 ))
    done
    if [[ -n "$is_custom_search" ]]; then
        GITLAB_MRS_DEEP_RESULTS="$mrs"
    else
        GITLAB_MRS="$mrs"
    fi
    [[ -n "$keep_quiet" ]] || echo -e "\033[1K\rDone."
}

# Adds an .approved boolean parameter to each entry in $GITLAB_MRS_BY_ME.
# Usage: __gl_add_approved_status_to_mrs_i_created <keep quiet>
__gl_add_approved_status_to_mrs_i_created () {
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
        mr_project_name="$( __gl_project_name "$mr_project_id" )"
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rGetting approval Status of MRs: $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_approvals="$( curl -s --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __gl_url_api_projects_mrs_approvals "$mr_project_id" "$mr_iid" )" )"
        mr_approved="$( echo -E "$mr_approvals" | jq ' .approved ' )"
        mr="$( echo -E "$mr" | jq -c " .approved = $mr_approved " )"
        my_mrs="$( echo -E "[$my_mrs,[$mr]]" | jq -c ' add ' )"
        mr_index=$(( mr_index + 1 ))
    done
    GITLAB_MRS_BY_ME="$( echo -E "$my_mrs" | jq -c ' sort_by(.source_branch, .project_id) ' )"
    [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\r"
}

# Adds discussion information to either $GITLAB_MRS_BY_ME or $GITLAB_MRS_TODO.
# Usage: __gl_add_discussion_info_to_mrs <keep quiet> <mr list type>
# If mr list type is "MINE" then $GITLAB_MRS_BY_ME is processed, otherwise $GITLAB_MRS_TODO is.
__gl_add_discussion_info_to_mrs () {
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
        mr_project_name=$( __gl_project_name "$mr_project_id" )
        [[ -n "$keep_quiet" ]] || echo -e -n "\033[1K\rGetting MR Discussion Info: $mr_index/$mr_count - $mr_project_name:$mr_iid "
        mr_discussions_url="$( __gl_url_api_projects_mrs_discussions "$mr_project_id" "$mr_iid" )?"
        mr_discussions="$( __gl_get_all_results "$mr_discussions_url" )"
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
# Usage: __gl_get_todos <keep quiet>
__gl_get_todos () {
    local keep_quiet todos todos_url
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting your GitLab ToDo List... "
    todos_url="$( __gl_url_api_todos )?"
    todos="$( __gl_get_all_results "$todos_url" )"
    GITLAB_TODOS="$( echo -E "$todos" | jq -c ' sort_by(.created_at) | reverse ' )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Gets all the jobs for the given project.
# Usage: __gl_get_project_jobs <keep quiet> <project name> <page count max>
__gl_get_project_jobs () {
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
    jobs_url="$( __gl_url_api_projects_jobs "$project_id" )?"
    gl_jobs="$( __gl_get_all_results "$jobs_url" "$page_count_max" )"
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

# Gets all the pages for a given endpoint.
# The url is required, and must end in either a ? or a &.
# page count max is optional. Default is 9999. It is forced to be between 1 and 9999 (inclusive).
# per page is optional. Default is 100. It is forced to be between 1 and 100 (inclusive).
# Usage: __gl_get_all_results <url> [<page count max>] [<per page>] [<verbose>] [<use keyset pagination>]
__gl_get_all_results () {
    local url page_count_max per_page verbose very_verbose use_keyset initial_url tmp_stderr error_info
    url="$1"
    page_count_max="$( __gl_clamp "$2" "1" "9999" "9999" )"
    per_page="$( __gl_clamp "$3" "20" "100" "100" )"
    verbose="$4"
    use_keyset="$5"
    if [[ "$url" =~ [^?\&]$ ]]; then
        printf '__gl_get_all_results [%s] must end in either a ? or a & so that extra parameters can be added.\n' "$url" >&2
        return 1
    fi
    if [[ "$verbose" == '-vv' ]]; then
        very_verbose="--verbose"
    fi
    if [[ -n "$use_keyset" ]]; then
        initial_url="${url}pagination=keyset&per_page=${per_page}"
    else
        initial_url="${url}per_page=${per_page}"
    fi
    tmp_stderr="$( mktemp -t gl_get_all_results_stderr )"
    if [[ -n "$verbose" ]]; then
        {
            printf 'Provided Url: [%s]\n' "$url"
            printf ' Initial Url: [%s]\n' "$initial_url"
            printf '   Max Pages: [%d]\n' "$page_count_max"
            printf '    Per Page: [%d]\n' "$per_page"
        } >&2
    fi
    {
        printf '['
        curl_link_header -s $very_verbose \
            --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" \
            --delimiter ',' --max-calls "$page_count_max" \
            "$initial_url"
        printf ']'
    } 2>> "$tmp_stderr" | jq -c ' add ' 2>> "$tmp_stderr"
    error_info="$( cat "$tmp_stderr" )"
    rm "$tmp_stderr"
    if [[ -n "$error_info" ]]; then
        {
            printf '__gl_get_all_results [%s] encountered one or more errors:\n' "$url"
            printf '%s\n' "$error_info" | sed 's/^/  /;'
        } >&2
        return 1
    fi
    return 0
}

# Gets all the branches of a repo.
# Usage: __gl_get_repo_branches <repo ssh url>
__gl_get_repo_branches () {
    git ls-remote "$1" 'refs/heads/*' | sed -E 's#^.*refs/heads/(.+)$#\1#;' | sort --ignore-case
}

# Marks a give todo item as done.
# Usage: __gl_post_mark_todo_as_done <keep quiet> <todo id>
__gl_post_mark_todo_as_done () {
    local keep_quiet todo_id marked_todo
    keep_quiet="$1"
    todo_id="$2"
    [[ -n "$keep_quiet" ]] || echo -E -n "Marking off ToDo... "
    marked_todo="$( curl -s --request POST --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __gl_url_api_todos_mark_as_done "$todo_id" )" )"
    if [[ -z "$keep_quiet" ]]; then
        local project_name todo_title
        project_name="$( echo -E "$marked_todo" | jq ' .project.name ' | sed -E 's/^"|"$//g' )"
        todo_title="$( echo -E "$marked_todo" | jq ' .body|.[0:80] ' )"
        echo -E "$project_name: $todo_title is marked as done."
    fi
}

# Marks all todo items as done.
# Usage: __gl_post_mark_all_todos_as_done <keep quiet>
__gl_post_mark_all_todos_as_done () {
    local keep_quiet all_done
    keep_quiet="$1"
    all_done="$( curl -s --request POST --header "PRIVATE-TOKEN: $GITLAB_PRIVATE_TOKEN" "$( __gl_url_api_todos_mark_all_as_done )" )"
    [[ -n "$keep_quiet" ]] || echo -E "All TODO items marked as done."
}

#
# GitLab Url Creators.
#

# Usage: __gl_url_base
__gl_url_base () {
    echo -E -n 'https://gitlab.com'
}

# Usage: __gl_url_api_v4
__gl_url_api_v4 () {
    __gl_url_base
    echo -E -n '/api/v4'
}

# Usage: __gl_url_api_user [<user id>]
__gl_url_api_user () {
    local user_id
    user_id="$1"
    __gl_url_api_v4
    echo -E -n '/user'
    if [[ -n "$user_id" ]]; then
        echo -E -n "s/$user_id"
    fi
}

# Usage: __get_gitlab_url_merge_requests
__gl_url_api_mrs () {
    __gl_url_api_v4
    echo -E -n '/merge_requests'
    # Note: This endpoint does not currently have the option to provide any sort of id for more specific information.
}

# Usage: __gl_url_api_projects [<project id>]
__gl_url_api_projects () {
    local project_id
    project_id="$1"
    __gl_url_api_v4
    echo -E -n '/projects'
    if [[ -n "$project_id" ]]; then
        echo -E -n "/$project_id"
    fi
}

# Usage: __gl_url_api_projects_jobs <project id> [<job id>]
__gl_url_api_projects_jobs () {
    local project_id job_id
    project_id="$1"
    job_id="$2"
    __gl_url_api_projects "$project_id"
    echo -E -n '/jobs'
    if [[ -n "$job_id" ]]; then
        echo -E -n "/$job_id"
    fi
}

# Usage: __gl_url_api_projects_jobs_log <project id> <job id>
__gl_url_api_projects_jobs_log () {
    __gl_url_api_projects_jobs "$1" "$2"
    echo -E -n '/trace'
}

# Usage: __gl_url_api_projects_mrs <project id> [<merge request iid>]
__gl_url_api_projects_mrs () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __gl_url_api_projects "$project_id"
    echo -E -n '/merge_requests'
    if [[ -n "$merge_request_iid" ]]; then
        echo -E -n "/$merge_request_iid"
    fi
}

# Usage: __gl_url_api_projects_mrs_approvals <project id> <merge request iid>
__gl_url_api_projects_mrs_approvals () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __gl_url_api_projects_mrs "$project_id" "$merge_request_iid"
    echo -E -n '/approvals'
}

# Usage: __gl_url_api_projects_mrs_discussions <project id> <merge request iid>
__gl_url_api_projects_mrs_discussions () {
    local project_id merge_request_iid
    project_id="$1"
    merge_request_iid="$2"
    __gl_url_api_projects_mrs "$project_id" "$merge_request_iid"
    echo -E -n '/discussions'
}

# Usage: __gl_url_api_todos [<todo id>]
__gl_url_api_todos () {
    local todo_id
    todo_id="$1"
    __gl_url_api_v4
    echo -E -n '/todos'
    if [[ -n "$todo_id" ]]; then
        echo -E -n "/$todo_id"
    fi
}

# Usage: __gl_url_api_todos_mark_as_done <todo id>
__gl_url_api_todos_mark_as_done () {
    local todo_id
    todo_id="$1"
    __gl_url_api_todos "$todo_id"
    echo -E -n "/mark_as_done"
}

# Usage: __gl_url_api_todos_mark_all_as_done
__gl_url_api_todos_mark_all_as_done () {
    __gl_url_api_todos
    echo -E -n "/mark_as_done"
}

# Usage: __gl_url_api_groups [<group id>]
__gl_url_api_groups () {
    local group_id
    group_id="$1"
    __gl_url_api_v4
    printf '/groups'
    if [[ -n "$group_id" ]]; then
        printf '/%s' "$group_id"
    fi
}

# Creates the desired url for glopen to use.
# Usage: __gl_url_web_repo <base url> <branch> <diff_branch>
__gl_url_web_repo () {
    local base_url branch diff_branch
    base_url="$1"
    branch="$2"
    diff_branch="$3"
    echo -E -n "$base_url"
    if [[ -n "$branch" ]]; then
        if [[ -n "$diff_branch" && "$branch" != "$diff_branch" ]]; then
            echo -E -n "/compare/$diff_branch...$branch"
        else
            echo -E -n "/-/tree/$branch"
        fi
    fi
}

# Creates the url for the mrs page of a repo.
# Usage: __gl_url_web_repo_mrs <base url>
__gl_url_web_repo_mrs () {
    local base_url
    base_url="$1"
    echo -E -n "$base_url/-/merge_requests"
}

# Creates the url for the pipelines page of a repo.
# Usage: __gl_url_web_repo_pipelines <base url>
__gl_url_web_repo_pipelines () {
    local base_url
    base_url="$1"
    echo -E -n "$base_url/pipelines"
}

return 0
