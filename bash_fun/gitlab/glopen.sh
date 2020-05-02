#!/bin/bash
# This file contains the glopen function which is a handy way to open various repo pages in gitlab.
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

__glopen_options_display_1 () {
    echo -E -n '[-r [<repo>]|--repo [<repo>]|--select-repo] [-b [<branch]|--branch [<branch>]|--select-branch]'
}
__glopen_options_display_2 () {
    echo -E -n '[-d [<branch>]|--diff [<branch>]|--select-diff-branch] [-m|--mrs] [-p|--pipelines] [-q|--quiet] [-x|--do-not-open]'
}
__glopen_auto_options () {
    echo -E -n "$( echo -E -n "$( __glopen_options_display_1 ) $( __glopen_options_display_2 )" | __gl_convert_display_options_to_auto_options )"
}
glopen () {
    __gl_require_token || return 1
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
  The -m or --mrs option will open the mrs page for the repo.
  The -p or --pipelines option will open the pipelines page for the repo.
  The -q or --quiet option suppresses normal terminal output.
  The -x or --do-not-open option will prevent the pages from being opened and only output the info.
    Technically, you can provide both -q and -x, but then nothing will really happen.

EOF
    )"
    local provided_repos provided_branches option select_repo random_repo
    local use_branch select_branch do_diff diff_branch select_diff_branch
    local open_mrs open_pipelines keep_quiet do_not_open
    provided_repos=()
    provided_branches=()
    while [[ "$#" -gt '0' ]]; do
        option="$( __gl_lowercase "$1" )"
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
        -m|--mrs)
            open_mrs="YES"
            ;;
        -p|--pipelines)
            open_pipelines="YES"
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
    __gl_ensure_projects "$keep_quiet"
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
    projects="$( __gl_project_subset "$select_repo" "$in_repo" "$( echo -E "${provided_repos[@]}" )" )"
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
            repo_branches="$( __gl_get_repo_branches "$project_ssh_url" )"
            diff_branch="$( echo -E "$repo_branches" | fzf --tac --cycle +m --header="$project_name (from)" )"
        fi
        if [[ -n "$select_branch" || (( -n "$use_branch" && "${#provided_branches[@]}" -eq '0' && -z "$in_branch" )) ]]; then
            fzf_header="$project_name"
            if [[ -n "$do_diff" ]]; then
                fzf_header="$fzf_header (to)"
            fi
            if [[ -z "$repo_branches" ]]; then
                repo_branches="$( __gl_get_repo_branches "$project_ssh_url" )"
            fi
            for branch in $( echo -E "$repo_branches" | fzf --tac --cycle -m --header="$fzf_header" ); do
                url="$( __gl_url_web_repo "$project_url" "$branch" "$diff_branch" )"
                urls_to_add+=( "$url" )
                messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$url" "$branch" "$diff_branch" "" )" )
            done
        elif [[ "${#provided_branches[@]}" -gt '0' ]]; then
            for branch in "${provided_branches[@]}"; do
                url="$( __gl_url_web_repo "$project_url" "$branch" "$diff_branch" )"
                urls_to_add+=( "$url" )
                messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$url" "$branch" "$diff_branch" "" )" )
            done
        elif [[ -n "$use_branch" && -n "$in_branch" ]]; then
            branch="$in_branch"
            url="$( __gl_url_web_repo "$project_url" "$branch" "$diff_branch" )"
            urls_to_add+=( "$url" )
            messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$url" "$branch" "$diff_branch" "" )" )
        fi
        if [[ -n "$open_mrs" ]]; then
            url="$( __gl_url_web_repo_mrs "$project_url" )"
            urls_to_add+=( "$url" )
            messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$url" "" "" "mrs" )" )
        fi
        if [[ -n "$open_pipelines" ]]; then
            url="$( __gl_url_web_repo_pipelines "$project_url" )"
            urls_to_add+=( "$url" )
            messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$url" "" "" "pipelines" )" )
        fi
        # If no urls for this project have yet been added to the list, add the main page url.
        if [[ "${#urls_to_add[@]}" -eq 0 ]]; then
            urls_to_add+=( "$project_url" )
            messages+=( "$( __gl_glopen_create_message_entry "$project_name" "$project_url" "" "" "" )" )
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

return 0
