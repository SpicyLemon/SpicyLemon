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

__glcodesearch_options_display_1 () {
    printf '[-h|--help] [-v|--verbose] [--summary|--project-summary|--context] [--links|--no-links] [--previous-results|--new-search]'
}
__glcodesearch_options_display_2 () {
    printf '[--global|--group <group id or name>|--project <project id or name>|--area <area identifier>]'
}
__glcodesearch_auto_options () {
    printf '%s' "$( __glcodesearch_options_display_1 ) $( __glcodesearch_options_display_2 )" | __gl_convert_display_options_to_auto_options
}
glcodesearch () {
    __gl_require_token || return 1
    local usage
    usage="$( cat << EOF
glcodesearch: GitLab Code Search

Search for Code in GitLab.

Usage: glcodesearch $( __glcodesearch_options_display_1 )
                    $( __glcodesearch_options_display_2 ) <search>

  The --global option is shorthand for '--area GLOBAL'.
  The --group <group id or name> option is shorthand for '--area GROUP <id or name>'.
  The --project <project id or name> option is shorthand for '--area PROJECT <id or name>'.
  The --area <area identifier> option defines the area to search in.
      <area identifier> must be one of the following:
          GLOBAL                The search will be global across GitLab.
          GROUP <id or name>    The search will be contained to the provided group.
          PROJECT <id or name>  The search will be contained to the provided project.
  Only one --global, --group, --project, or --area option can be used.
  If more than one search areas are provided, the last one provided is the one that will be used.

  The search area must be defined either through one of the above options,
    or through the GITLAB_CODE_SEARCH_DEFAULT_OPTIONS environment variable.

  The --context option causes the output to contain the project names, file names, and lines from the files.
  The --summary option causes the output to contain just the project names and file names.
  The --project-summary option causes the output to only contain the project names.
  If more than one of --context, --summary, or --project-summary are provided, the last one provided is the one that will be used.
  If none of them are provided, the default behavior is --context.

  The --links option causes the output to contain links to the repository and files.
  The --no-links option causes the repository and file links to not appear in the output.
  If more than one of --links or --no-links are provided, the last one provided is the one that will be used.
  If neither are provided, the default behavior is based on the --summary or --context options.
    By default, --context uses --links and both --summary and --project-summary use --no-links.

  The -v or --verbose flag causes this to output extra information that might be helpful in troubleshooting.

  Any other parameters are treated as the search to execute.
  If needed, the search can also be split off from the options using --.
  If -- is provided, everything after it is treated as the search to execute.

  The --previous-results option will cause the search to be ignored, and will instead output the desired
    report from the previous search results.
  The --new-search option indicates that a new search is to be done.
  If more than one of --previous-results or --new-search are provided, the last one provided is the one that will be used.
  If neither are provided, the default behavior is --new-search.

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

EOF
    )"
    local verbose area area_spec area_id search group project output_level include_links use_previous
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
            if [[ "$area" != 'GLOBAL' ]]; then
                area_spec=$3
                shift
            fi
            shift
            ;;
        --context)
            output_level="CONTEXT"
            ;;
        --summary)
            output_level="SUMMARY"
            ;;
        --project-summary)
            output_level="PROJECT_SUMMARY"
            ;;
        --links)
            include_links="YES"
            ;;
        --no-links)
            include_links="NO"
            ;;
        --previous-results)
            use_previous="YES"
            ;;
        --new-search)
            use_previous="NO"
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
    local search_url query_string results_count
    if [[ -z "$output_level" ]]; then
        output_level="CONTEXT"
    fi

    if [[ -z "$include_links" ]]; then
        if [[ "$output_level" == "CONTEXT" ]]; then
            include_links="YES"
        else
            include_links="NO"
        fi
    fi

    if [[ "$use_previous" != "YES" ]]; then
        if [[ -z "$area" ]]; then
            printf 'No search area provided.\n' >&2
            return 1
        elif [[ "$area" != 'GLOBAL' && "$area" != 'GROUP' && "$area" != 'PROJECT' ]]; then
            printf 'Invalid area: [%s].\n' "$area" >&2
            return 1
        elif [[ "$area" != 'GLOBAL' ]]; then
            if [[ -z "$area_spec" ]]; then
                printf 'The <%s id or name> is missing and required.\n' "$( printf %s "$area" | __gl_lowercase )" >&2
                return 1
            elif [[ "$area_spec" =~ ^[[:digit:]]*$ ]]; then
                area_id="$area_spec"
            elif [[ "$area" == 'GROUP' ]]; then
                __gl_ensure_groups
                group="$( __gl_group_lookup '' "$area_spec" )"
                if [[ -n "$group" ]]; then
                    area_id="$( jq -r ' .id ' <<< "$group" )"
                fi
                if [[ -z "$area_id" ]]; then
                    printf 'Unknown group name: [%s].\n' "$area_spec" >&2
                    return 1
                fi
            elif [[ "$area" == 'PROJECT' ]]; then
                __gl_ensure_projects
                project="$( __gl_project_lookup '' "$area_spec" )"
                if [[ -n "$project" ]]; then
                    area_id="$( jq -r ' .id ' <<< "$project" )"
                fi
                if [[ -z "$area_id" ]]; then
                    printf 'Unknown project name: [%s].\n' "$area_spec" >&2
                    return 1
                fi
            fi
        fi
        if [[ -z "$search" ]]; then
            printf 'No search provided.\n' >&2
            return 1
        fi

        if [[ "$area" == 'GLOBAL' ]]; then
            search_url="$( __gl_url_api_search_global )"
        elif [[ "$area" == 'GROUP' ]]; then
            search_url="$( __gl_url_api_search_in_group "$area_id" )"
        elif [[ "$area" == 'PROJECT' ]]; then
            search_url="$( __gl_url_api_search_in_project "$area_id" )"
        fi
        query_string="scope=blobs&search=$( printf %s "$search" | __gl_encode_for_url )"

        GITLAB_CODE_SEARCH_RESULTS="$( __gl_get_all_results "${search_url}?${query_string}&" '' '' "$verbose" )"
    fi

    result_count="$( jq ' length ' <<< "$GITLAB_CODE_SEARCH_RESULTS" )"

    printf "Found $result_count result"
    [[ "$result_count" -ne '1' ]] && printf 's'
    [[ "$result_count" -eq '0' ]] && printf '.\n'

    if [[ "$result_count" -ge '1' ]]; then
        __gl_ensure_projects
        local project_ids projects
        local gitlab_code_search_results_1 gitlab_code_search_results_2 gitlab_code_search_results_3 gitlab_code_search_results_4
        local project_count project_index project_results project_name project_web_url project_file_count project_output
        local project_file_index project_file project_file_path project_file_web_url project_file_line_count

        # Get info on all the projects involved in the results.
        project_ids="$( jq -c ' [ .[] | .project_id ] | unique ' <<< "$GITLAB_CODE_SEARCH_RESULTS" )"
        projects="$( jq -c --argjson project_ids "$project_ids" \
                        ' [ .[] | select( .id | first( ( $project_ids[] == . ) // empty ) // false ) ]
                        | map({ (.id|tostring): .}) | add ' <<< "$GITLAB_PROJECTS" )"

        # Note: These jq commands are intentionally split out like this instead of combining them into
        #       one big long command. This way, it's easier to tweak and mess with in the future.

        # Add project info to each result and split the data field into lines.
        gitlab_code_search_results_1="$( jq -c --sort-keys --argjson projects "$projects" \
                                          '[ .[] | $projects[.project_id|tostring] as $project
                                                 | .startline as $startline
                                                 | .project_name = $project.name
                                                 | .project_sort_key = ( $project.name_with_namespace | ascii_downcase )
                                                 | .file_sort_key = ( .path | ascii_downcase )
                                                 | .project_web_url = $project.web_url
                                                 | .file_web_url = $project.web_url + "/-/blob/" + .ref + "/" + .path
                                                 | .lines = ( ( .data | rtrimstr("\n") ) / "\n"
                                                                | to_entries
                                                                | map( { line: .value, line_number: ( .key + $startline ) } ) )
                                                 | del(.data)
                                            ]' <<< "$GITLAB_CODE_SEARCH_RESULTS" )"
        # Group the results together by project, then by filename.
        gitlab_code_search_results_2="$( jq -c ' group_by( .project_id )
                                                    | [ .[] | group_by( .path ) ] ' <<< "$gitlab_code_search_results_1" )"
        # Combine each filename results into a single result with all the lines from each result with that filename.
        gitlab_code_search_results_3="$( jq -c ' [ .[] | [ .[]
                                                    | { all_lines: ( .[] | .lines
                                                        | unique_by( .line_number )
                                                        | sort_by( .line_number ) ) } + .[0]
                                                    | del(.startline, .lines) ]
                                                ] ' <<< "$gitlab_code_search_results_2" )"
        # Convert each project entry into an object containing the project info and list of file results.
        gitlab_code_search_results_4="$( jq -c ' [ .[] | { files: ( . | sort_by ( .file_sort_key ) ) }
                                                            + ( .[0] | { project_id, project_name,
                                                                         project_sort_key, project_web_url } )
                                                ] | sort_by( .project_sort_key ) ' <<< "$gitlab_code_search_results_3" )"

        # Do the output!
        project_count="$( jq -r 'length' <<< "$gitlab_code_search_results_4" )"
        printf ' in %d repositor' "$project_count"
        [[ "$result_count" -eq '1' ]] && printf 'y.\n' || printf 'ies.\n'

        for project_index in $( seq 0 $(( project_count - 1 )) ); do
            project_results="$( jq -c --arg project_index "$project_index" \
                                ' .[$project_index|tonumber] ' <<< "$gitlab_code_search_results_4" )"
            project_name="$( jq -r ' .project_name ' <<< "$project_results" )"
            project_web_url="$( jq -r ' .project_web_url ' <<< "$project_results" )"
            project_file_count="$( jq -r ' .files | length ' <<< "$project_results" )"
            # Put together the output chunk for the whole project.
            # This way, it can be displayed all-at-once instead of as each file is ready.
            # Doing the output for each file was just a little overly-jerky.
            project_output=""
            if [[ "$output_level" == "PROJECT_SUMMARY" ]]; then
                project_output+="$( printf '\033[97m%s\033[0m: %d file' "$project_name" "$project_file_count" )"
            else
                project_output+="$( printf '\033[97m%s\033[0m (%d of %d): %d file' "$project_name" "$(( project_index + 1 ))" "$project_count" "$project_file_count" )"
            fi
            project_output+="$( [[ "$project_file_count" -ne '1' ]] && printf 's' )\n"
            if [[ "$include_links" == "YES" ]]; then
                project_output+="$( printf '\033[4;96m%s\033[0m' "$project_web_url" )\n"
            fi
            if [[ "$output_level" == "PROJECT_SUMMARY" ]]; then
                printf '%b' "$project_output"
            else
                for project_file_index in $( seq 0 $(( project_file_count - 1)) ); do
                    project_file="$( jq -c --arg project_file_index "$project_file_index" \
                                     ' .files[$project_file_index|tonumber] ' <<< "$project_results" )"
                    project_file_path="$( jq -r ' .path ' <<< "$project_file" )"
                    project_file_web_url="$( jq -r ' .file_web_url ' <<< "$project_file" )"
                    project_file_line_count="$( jq -r ' .all_lines | length ' <<< "$project_file" )"
                    if [[ "$output_level" == "CONTEXT" ]]; then
                        project_output+="$( printf '    \033[97m%s\033[0m - \033[93m%s\033[0m (%d of %d):' "$project_name" "$project_file_path" "$(( project_file_index + 1 ))" "$project_file_count" )\n"
                    else
                        project_output+="$( printf '    \033[93m%s\033[0m (%d of %d):' "$project_file_path" "$(( project_file_index + 1 ))" "$project_file_count" )\n"
                    fi
                    if [[ "$include_links" == "YES" ]]; then
                        project_output+="$( printf '    \033[4;96m%s\033[0m' "$project_file_web_url" )\n"
                    fi
                    if [[ "$output_level" == "CONTEXT" ]]; then
                        project_output+="$( jq -r ' .all_lines[] | "        [" + ( "    " + (.line_number|tostring) | .[-4:] ) + "]: " + .line ' <<< "$project_file" )\n"
                    fi
                done
                printf '%b\n' "$project_output"
            fi
        done
    fi
}

return 0
