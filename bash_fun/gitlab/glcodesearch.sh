#!/bin/bash
# This file contains the glcodesearch function that allows you to do generic searches for code.
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

__glcodesearch_options_display () {
    printf '[-h|--help] [--global|--group <group id or name>|--project <project id or name>|--area <area identifier>]'
}
__glcodesearch_auto_options () {
    __glcodesearch_options_display | __gl_convert_display_options_to_auto_options
}
glcodesearch () {
    __gl_require_token || return 1
    local usage
    usage="$( cat << EOF
glcodesearch: GitLab Code Search

Search for Code in Gitlab.

Usage: glcodesearch $( __glcodesearch_options_display ) <search>

  The --global option is shorthand for '--area GLOBAL'.
  The --group <group id or name> option is shorthand for '--area GROUP <id or name>'.
  The --project <project id or name> option is shorthand for '--area PROJECT <id or name>'.
  The --area <area identifier> option defines the area to search in.
      <area identifier> must be one of the following:
          GLOBAL                The search will be global across Gitlab.
          GROUP <id or name>    The search will be contained to the provided group.
          PROJECT <id or name>  The search will be contained to the provided project.
  Only one --global, --group, --project, or --area option can be used.
  If more than one search areas are provided, the last one provided is the one that will be used.

  The search area must be defined either through one of the above options, 
    or through the GITLAB_CODE_SEARCH_DEFAULT_OPTIONS environment variable.
  If the GITLAB_CODE_SEARCH_DEFAULT_OPTIONS is defined, it will be treated like the first options provided.
    For example, if  GITLAB_CODE_SEARCH_DEFAULT_OPTIONS='--area GROUP 12345'  then the command
        glcodesearch someMethodName
      is the same as
        glcodesearch --area GROUP 12345 someMethodName
    In this same example, if you then execute
        glcodesearch --project 'my project' someMethodName
      it will be treated as
        glcodesearch --area GROUP 12345 --project 'my project' someMethodName
      Howerver, since the last area defined is the one that is used, the default '--area GROUP 12345' is effectively ignored.

  Any other parameters are treated as the search to execute.
  If needed, the search can also be split off from the options using --.
  If -- is provided, everything after it is treated as the search to execute.

EOF
    )"
    local verbose area area_spec area_id search
    if [[ -n "$GITLAB_CODE_SEARCH_DEFAULT_OPTIONS" ]]; then
        set -- $GITLAB_CODE_SEARCH_DEFAULT_OPTIONS "$@"
    fi
    while [[ "$#" -gt '0' ]]; do
        case "$( printf %s "$1" | __gl_lowercase )" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        -v|--verbose)
            verbose="$1"
            ;;
        --global)
            area='GLOBAL'
            area_spec=
            ;;
        --group)
            area='GROUP'
            area_spec=$2
            shift
            ;;
        --project)
            area='PROJECT'
            area_spec=$2
            shift
            ;;
        --area)
            area="$( printf %s "$2" | __gl_uppercase )"
            if [[ "$area" ne 'GLOBAL' ]]; then
                area_spec=$3
                shift
            fi
            shift
            ;;
        --)
            if [[ -n "$search" ]]; then
                printf 'Unknown options: [%s].\n' "$search" >&2
                return 1
            fi
            search="$@"
            set --
            ;;
        *)
            if [[ -z "$search" ]]; then
                search="$1"
            else
                search+=" $1"
            fi
            ;;
        esac
        shift
    done
    local search_url query_string
    if [[ -z "$area" ]]; then
        printf 'No search area provided.\n' >&2
        return 1
    elif [[ "$area" != 'GLOBAL' && "$area" != 'GROUP' && "$area" != 'PROJECT' ]]; then
        printf 'Invalid area: [%s].\n' "$area" >&2
        return 1
    elif [[ "$area" != 'GLOBAL' ]]; then
        if [[ -z "$area_spec" ]]; then
            printf 'The <%s id or name> is missing and required.\n' "$( printf %s "$area" | __gl_lowercase )"
            return 1
        elif [[ "$area_spec" =~ ^[[:digit:]]*$ ]]; then
            area_id="$area_spec"
        elif [[ "$area" == 'GROUP' ]]; then
            # TODO: Add group id lookup from name.
            printf 'Group name lookup not yet supported; use the group id instead.\n' "$area" >&2
            return 2
        elif [[ "$area" == 'PROJECT' ]]; then
            # TODO: Add project id lookup from name.
            printf 'Project name lookup not yet supported; use the project id instead.\n' "$area" >&2
            return 2
        fi
    fi
    if [[ -z "$search" ]]; then
        printf 'No search provided.\n' >&2
        return 1
    fi

    if [[ "$area" == 'GLOBAL' ]]; then
        search_url="$( __gl_url_api_search_global )"
    elif [[ "$area" == 'GROUP' ]]; then
        search_url="$( __gl_url_api_search_in_group )"
    elif [[ "$area" == 'PROJECT' ]]; then
        search_url="$( __gl_url_api_search_in_project )"
    fi
    query_string="scope=blobs&search=$( printf %s "$search" | __gl_encode_for_url )"

    GITLAB_CODE_SEARCH_RESULTS="$( __gl_get_all_results "${search_url}?${query_string}&" '' '' "$verbose" )"

    local results_count results
    result_count="$( jq ' length ' <<< "$GITLAB_CODE_SEARCH_RESULTS" )"

    printf "Found $result_count result"
    [[ "$result_count" -ne '1' ]] && printf 's'
    printf '.\n'

    if [[ "$result_count" -ge '1' ]]; then
        # TODO: Convert the results into a better format for output.
        results="$GITLAB_CODE_SEARCH_RESULTS"
        jq '.' <<< "$results"
    fi
}

return 0
