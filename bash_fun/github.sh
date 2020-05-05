#!/bin/bash
# Work in progress - Not really usable.
# This script creates some functions that are handing for interacting with GitHub.
# File contents:
#   ghclone  --> GitHub Clone - Easily find and clone a repo right from your terminal.
#   ghclean  --> GitHub Clean - Cleans up all the environment variables used by these functions.
#
# In order to use any of these functions, you will first have to create a GitHub private token.
#   1) Log into GitHub.
#   2) Click on your profile in the upper right and select "Settings".
#   3) Click on "Developer settings" on the left side.
#   4) Click on "Personal access tokens" on the left side.
#   5) Click the "Generate new token" button.
#   6) Give the token a name, and the scopes you want to have available with it.
#       I recommend all the read scopes and none of the write scopes.
#   7) Click the "Generate token" button at the bottom.
#   4) Set the GITHUB_PRIVATE_TOKEN environment variable to the value of that token.
#       For example, you could put   GITHUB_PRIVATE_TOKEN=1234567890abcdef1234567890abcdef12345678  in your .bash_profile file
#       so that it's set every time you open a terminal (use your own actual token of course).
#
# To make these functions usable in your terminal, use the source command on this file.
#   For example, you could put  source github.sh  in your .bash_profile file.
#
# Then you call these functions just like you would a normal program.
#   Examples:
#      ghclone
#         Looks up all your projects and displays them so that you can select one or more to clone locally.
#      ghclean
#         Removes all the persistant environment variables created and used by the functions in this file.
#
# NOTE: The functions in here rely on the following programs (that you might not have installed yet):
#   * fzf - Command-line fuzzy finder - https://github.com/junegunn/fzf
#   * jq - Command-line JSON processor - https://github.com/stedolan/jq
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

ghclone () {
    __ensure_github_token || return 1
    local usage destination option use_the_force refresh orig_pwd cloned_repo_count repo_lines selected_repo repo_url cmd cmd_output new_repo_dir
    usage="$(
        cat << EOF
ghclone: GitHub Clone

This will look up all the projects you have access to in GitHub, and provide a way for you to select one or more to clone.
You must create an api token from your profile in GitHub first. See: $( __get_github_base_url )/settings/tokens
Then, you must set the token value as the GITHUB_PRIVATE_TOKEN environment variable in your terminal (e.g. in .bash_profile).

Additionally, if you set the GITHUB_BASE_DIR environment variable to you root git directory,
new repos will automatically go into that directory regardless of where you are when running the command.
If that variable is not set, and no -b or --base-dir parameter is provided, the current directory is used.

Usage: glclone [-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh]

  The -b <dir> or --base-dir <dir> option will designate the directory to create your repo in.
        Providing this option overrides the default setting from the GITHUB_BASE_DIR.
  The -f or --force option will allow cloning into directories already under a git repo.
  The -r or --refresh option will cause your projects to be reloaded.

EOF
    )"
    destination="$GITHUB_BASE_DIR"
    while [[ $# -gt 0 ]]; do
        option=$( echo -E "$1" | tr "[:upper:]" "[:lower:]" )
        case "$option" in
        -h|--help)
            echo -e "$usage"
            return 0
            ;;
        -b|--base-dir)
            __gh_ensure_option "$2" "$option" || ( >&2 echo -e "$usage" && return 1 ) || return 1
            destination="$2"
            shift
            ;;
        -f|--force)
            use_the_force="YES"
            ;;
        -r|--refresh)
            refresh="YES"
            ;;
        *)
            >&2 echo -E "Unknown option provided: '$option'"
            >&2 echo -e "$usage"
            return 2
            ;;
        esac
        shift
    done
    if [[ -n "$destination" ]]; then
        if [[ ! -d "$destination" ]]; then
            >&2 echo -E "Destination directory [$destination] does not exist."
            >&2 echo -e "$usage"
            return 1
        fi
        orig_pwd="$( pwd )"
        cd "$destination"
    fi
    if [[ -z $use_the_force && $(git rev-parse --is-inside-work-tree 2>/dev/null) ]]; then
        if [[ -n "$destination" ]]; then
            >&2 echo -E "$destination is already inside a git repo. If you'd still like to clone a repo into this directory, use the command ghclone -f"
        else
            >&2 echo -E "You are already inside a git repo. If you'd still like to clone a repo into this directory, use the command ghclone -f"
        fi
        >&2 echo -e "$usage"
        if [[ -n "$orig_pwd" ]]; then
            cd "$orig_pwd"
        fi
        return 1
    fi

    if [[ -n "$refresh" ]]; then
        GITHUB_USER=''
        GITHUB_REPOS=''
    fi

    __ensure_github_user_info
    __ensure_github_repos

    cloned_repo_count=0
    repo_lines=$( echo -E "$GITHUB_REPOS" | jq -r ' sort_by(.full_name) | .[] | .full_name ' | fzf --cycle -m )
    if [[ -n "$repo_lines" ]]; then
        echo -E ""
        for selected_repo in "$repo_lines"; do
            if [[ -n "$GITHUB_USE_HTTPS" && "$GITHUB_USE_HTTPS" == "1" ]]; then
                repo_url="https://github.com/${selected_repo}.git"
            else
                repo_url="git@github.com:${selected_repo}.git"
            fi
            cmd="git clone --progress $repo_url"
            echo -e "\033[1;37m$cmd\033[0m"
            exec 3>&1
            cmd_output="$( $cmd 2>&1 | tee >(cat - >&3) )"
            exec 3>&-
            echo -E ""
            cloned_repo_count=$(( cloned_repo_count + 1 ))
        done
    fi
    if [[ "$cloned_repo_count" -eq "0" ]]; then
        if [[ -n "$orig_pwd" ]]; then
            cd "$orig_pwd"
        fi
    elif [[ "$cloned_repo_count" -eq "1" ]]; then
        new_repo_dir="$( echo -E "$cmd_output" | grep '^Cloning into' | sed -E "s/^Cloning into '//; s/'\.+[[:space:]]*$//;" )"
        cd "$new_repo_dir"
    fi
}

ghclean () {
    local vars_to_clean vars_str usage verbose v
    vars_to_clean=("GITHUB_USER" "GITHUB_USER_ID" "GITHUB_USERNAME" "GITHUB_REPOS")
    vars_str="$( echo -E "${vars_to_clean[*]}" | sed -E 's/ /~/g; s/([^~]+~[^~]+~[^~]+~[^~]+)~/\1\\n/g;' )"
    vars_str="$( echo -e "$vars_str" | column -s '~' -t | sed 's/^/    /' )"
    usage="$(
        cat << EOF
glclean - GitHub Clean

Cleans up all the persistant variables used by the functions in this file.
Use this when you want a fresh start with respects to the data these GitHub functions use.

This will NOT affect your GITHUB_PRIVATE_TOKEN variable.

The following variables will be removed:
$( echo -e "$vars_str" )

Usage: glclean [-v]

  The -v option will output the values of each variable before being deleted.

EOF
    )"
    if [[ -n "$1" && "$1" == "-v" ]]; then
        verbose="YES"
        shift;
    fi
    if [[ $# -gt 0 ]]; then
        echo -e "$usage"
        return 0
    fi
    for v in ${vars_to_clean[*]}; do
        if [[ -n "$verbose" ]]; then
            if [[ -n "$( ps -o command= $$ | grep -E "zsh$" )" ]]; then
                echo -E "$v=${(P)v}"
            else
                echo -E "$v=${!v}"
            fi
        fi
        unset $v
    done
    echo -E "GitHub-associated variables cleaned."
}

# Makes sure that a GitHub token is set.
# This must be set outside of this file, and is kind of a secret thing.
# Usage: __ensure_github_token
__ensure_github_token () {
    if [[ -z "$GITHUB_PRIVATE_TOKEN" ]]; then
        cat << EOF
No GITHUB_PRIVATE_TOKEN has been set.
To create one, go to $( __get_github_base_url )/settings/tokens and create one with the scopes desired.
Then you can set it using
GITHUB_PRIVATE_TOKEN=whatever-your-token-is
It is probably best to put that line somewhere so that it will get executed whenever you start your terminal (e.g. .bash_profile).
EOF
        return 1
    fi
}

# Makes sure that an option was provided with a flag.
# Usage: __gh_ensure_option "$2" "option name" || echo "bad option."
__gh_ensure_option () {
    local value option
    value="$1"
    option="$2"
    if [[ -z "$value" || ${value:0:1} == "-" ]]; then
        >&2 echo -E "A parameter must be supplied with the $option option."
        return 1
    fi
    return 0
}

# Makes sure that your GitHub user info has been loaded.
# If not, it is looked up.
# Usage: __ensure_github_user_info <keep quiet>
__ensure_github_user_info () {
    local keep_quiet
    keep_quiet="$1"
    if [[ -z "$GITHUB_USER" || -z "$GITHUB_USER_ID" || -z "$GITHUB_USERNAME" ]]; then
        __get_github_user_info "$keep_quiet"
    fi
}

# Looks up your GitHub user info. Results are stored in $GITHUB_USER, $GITHUB_USER_ID, and $GITHUB_USERNAME.
# Usage: __get_github_user_info <keep quiet>
__get_github_user_info () {
    local keep_quiet
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting your GitHub user info... "
    GITHUB_USER="$( curl -s --header "Authorization: token $GITHUB_PRIVATE_TOKEN" "$( __get_github_url_user )" )"
    GITHUB_USER_ID="$( echo -E "$GITHUB_USER" | jq -r '.id' )"
    GITHUB_USERNAME="$( echo -E "$GITHUB_USER" | jq -r '.login' )"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Makes sure that the $GITHUB_REPOS varialbe has a value.
# If not, the list is looked up.
# Usage: __ensure_github_repos <keep quiet>
__ensure_github_repos () {
    local keep_quiet
    keep_quiet="$1"
    if [[ -z "$GITHUB_REPOS" ]]; then
        __get_github_repos "$keep_quiet"
    fi
}

# Look up info on all available projects. Results are stored in $GITHUB_REPOS.
# Usage: __get_github_repos <keep quiet>
__get_github_repos () {
    local keep_quiet repos_url page repos
    keep_quiet="$1"
    [[ -n "$keep_quiet" ]] || echo -E -n "Getting all your GitHub repos... "
    repos_url="$( __get_github_url_repos )"
    repos="$( __gh_get_pages_of_url "$repos_url" )"
    GITHUB_REPOS="$repos"
    [[ -n "$keep_quiet" ]] || echo -E "Done."
}

# Gets all the pages for a given endpoint.
# The url is required.
# page count max is optional. Default is 9999. It is forced to be between 1 and 9999 (inclusive).
# Usage: __gh_get_pages_of_url <url> [<page count max>]
__gh_get_pages_of_url () {
    local url page all_done results respose header body next_url
    url="$1"
    results="[]"
    page=1
    while [[ -z "$all_done" ]]; do
        response="$( curl --silent --include --header "Authorization: token $GITHUB_PRIVATE_TOKEN" "$url" )"
        header="$( echo "$response" 2> /dev/null | sed '/^[[:space:]]*$/q' )"       #Gets everything up to (and including) the first empty line.
        body="$( echo "$response" 2> /dev/null | sed -n '/^[[:space:]]*$/,$p' )"    #Gets everything after (and including) the first empty line.
        if [[ -n "$( echo -E "$body" | jq -r ' if type=="array" then "okay" else "" end ' )" ]]; then
            results="$( echo -E "[$results,$body]" | jq -c ' add ' )"
            next_url="$( echo -E "$header" | grep '^Link:' | tr ',' "\n" | grep 'next' | sed -E 's/^.*<//; s/>.*$//' )"
            if [[ -n "$next_url" ]]; then
                url="$next_url"
            else
                all_done="YUP"
            fi
        else
            >&2 echo -E "$full_url -> $body"
            all_done="YUP"
        fi
        page=$(( page + 1 ))
        if [[ "$page" -gt "25" ]]; then
            all_done="YUP"
        fi
    done
    echo -E "$results"
}

# Makes sure that a provided value is a number between the min and max (inclusive).
# Only whole numbers are allowed.
# The min and max parameters are interchangable. If larger of the two is the max, and the other is the min.
# If the provided value is not a number, then the default is returned.
# If it's less than the min, the min is returned.
# If it's less than the max, the max is returned.
# Usage: __gh_clamp <value> <min> <max> <default>
__gh_clamp () {
    local val min max default result
    val="$( __gh_ensure_number_or_default "$1" "" )"
    min="$( __gh_ensure_number_or_default "$2" "" )"
    max="$( __gh_ensure_number_or_default "$3" "" )"
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
# Usage: __gh_ensure_number_or_default <value> <default>
__gh_ensure_number_or_default () {
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

# GitHub API documentation: https://developer.github.com/v3/

# Usage: __get_github_base_url
__get_github_base_url () {
    echo -E -n 'https://github.com'
}

# Usage: __get_github_api_url
__get_github_api_url () {
    echo -E -n 'https://api.github.com'
}

# Usage: __get_github_url_user [<username>]
__get_github_url_user () {
    local username
    username="$1"
    __get_github_api_url
    echo -E -n '/user'
    if [[ -n "$username" ]]; then
        echo -E -n "s/$username"
    fi
}

# Usage: __get_github_url_user_repos [<username>]
__get_github_url_user_repos () {
    local username
    username="$1"
    __get_github_url_user "$username"
    echo -E -n '/repos'
}

# Usage: __get_github_url_repos
__get_github_url_repos () {
    __get_github_api_url
    echo -E -n '/repositories'
}

__call_github_url () {
    local url cli_header cmd
    url="$1"
    extra_cli="$2"
    cmd="curl --silent $extra_cli --header \"Authorization: token $GITHUB_PRIVATE_TOKEN\" $url"
    echo -e "\033[1;37m$cmd\033[0m"
    exec 3>&1
    CALL_GITHUB_URL_LAST_RESPONSE="$( $cmd 2>&1 | tee >(cat - >&3) )"
    exec 3>&-
}

return 0
