#!/bin/bash
# This file contains lots of handy functions for dealing with git.
# File contents:
#   in_git_folder  --------> Helper function for testing if you're currently in a git folder.
#   gcb  ------------------> Git Change Branch - Select a branch and switch to it.
#   gcba  -----------------> Git Change Branch (All) - Gets a list of all branches (local and remote) and lets you pick one to checkout.
#   gcbm  -----------------> Git Change Branch (to) Master - Checkout the master branch.
#   gdb  ------------------> Git Delete Branches - Select branches that you want to delete, and then deletes them.
#   bn  -------------------> Branch Name - Outputs your current branch name.
#   gpm  ------------------> Git Pull Merge (Master) - Pull master and merge it into your branch.
#   gsu  ------------------> Git Set Upstream - Sets the upstream appropriately for the repo and branch you're in.
#   clean_git_repo  -------> Takes several actions to help you clean up a git repo.
#   gfb  ------------------> Pulls master and creates a fresh branch from it.
#   list_extra_branches  --> List all the local extra branches in all the local repos.
#   master_pull_all  ------> Finds all your repos and does a pull on the master branches of each one.
#   git_diff_analysis  ----> Compares the current branch with master (or provided branch) in order to get some stats for you.
#
# Depends on:
#   fzf - Command-line fuzzy finder - https://github.com/junegunn/fzf - brew install fzf
#   jq - Command-line JSON processor - https://github.com/stedolan/jq - brew install jq

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

# Output a command, then execute it.
# Usage: __git_echo_do <command> [<arg1> [<arg2> ...]]
#   or   __git_echo_do "command string"
# Examples:
#   __git_echo_do say -vVictoria -r200 "Buu Whoa"
# The exit code of the command will be returned by this function.
# If no command is provided, this will return with exit code 124
__git_echo_do () {
    local cmd_pieces pieces_for_output cmd_piece retval
    cmd_pieces=()
    if [[ "$#" > '0' ]]; then
        cmd_pieces+=( "$@" )
    fi
    if [[ "${#cmd_pieces[@]}" -eq '0' || "${cmd_pieces[@]}" =~ ^[[:space:]]*$ ]]; then
        return 124
    fi
    pieces_for_output=()
    if [[ "${#cmd_pieces[@]}" -eq '1' && ( "${cmd_pieces[@]}" =~ [[:space:]\(=] || -z "$( command -v "${cmd_pieces[@]}" )" ) ]]; then
        pieces_for_output+=( "${cmd_pieces[@]}" )
        cmd_pieces=( 'eval' "${cmd_pieces[@]}" )
    else
        for cmd_piece in "${cmd_pieces[@]}"; do
            if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
                pieces_for_output+=( "\"$( echo -E "$cmd_piece" | sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' )\"" )
            else
                pieces_for_output+=( "$cmd_piece" )
            fi
        done
    fi
    echo -en "\033[1;37m"
    echo -En "${pieces_for_output[@]}"
    echo -e "\033[0m"
    "${cmd_pieces[@]}"
    retval="$?"
    echo -E ''
    return "$retval"
}

# Tests if you're in a git folder
# Usage: in_git_folder && echo "In a git folder!" || echo "Not in a git folder."
in_git_folder () {
    [[ $( git rev-parse --is-inside-work-tree 2>/dev/null ) ]] && return 0
    return 1
}

# Change your current git branch
# Usage: gcb
gcb () {
    local branch
    branch="$1"
    if in_git_folder; then
        local selection
        if [[ -n "$1" ]]; then
            selection="$1"
        else
            selection="$( git branch | fzf +m --cycle | sed -E 's/^[* ]+//' )"
        fi
        [[ -n "$selection" ]] && git checkout "$selection"
    else
        echo "gcb => git change branch. But you aren't in a git directory."
    fi
}

# Check out a git branch from remote
# Usage: gcba
gcba () {
    if in_git_folder; then
        [[ $(which setopt) ]] && setopt local_options BASH_REMATCH KSH_ARRAYS
        local all_branches local_branches remote remote_branches new_branches selected_entry remote_and_branch_rx just_branch_rx branch
        all_branches="$( git branch | sed -E 's#^([* ]) #\1 ~ ~#' )"
        local_branches="$( echo -E "$all_branches" | sed -E 's#^[^~]*~[^~]*~##' )"
        for remote in $( git remote ); do
            git fetch -q "$remote"
            remote_branches="$( git ls-remote --heads "$remote" | sed -E 's#^.*refs/heads/##' )"
            new_branches="$( echo -E "$( echo -E "$local_branches" && echo -E "$local_branches" && echo -E "$remote_branches" )" | sort | uniq -u | sed -E "s#^#  ~$remote~#" )"
            all_branches="$( echo -E "$all_branches" && echo -E "$new_branches" )"
        done
        selected_entry="$( echo -E "$all_branches" | sort -t '~' -k 3 -k 2 | column -s '~' -t | fzf +m --cycle )"
        if [[ -n "$selected_entry" ]]; then
            remote_and_branch_rx='^[* ] +([^ ]+) +(.+)$'
            just_branch_rx='^[* ] +(.+)$'
            if [[ "$selected_entry" =~ $remote_and_branch_rx ]]; then
                remote="${BASH_REMATCH[1]}"
                branch="${BASH_REMATCH[2]}"
                __git_echo_do git checkout --track "$remote/$branch"
            elif [[ "$selected_entry" =~ $just_branch_rx ]]; then
                branch="${BASH_REMATCH[1]}"
                __git_echo_do git checkout "$branch"
            else
                echo -E "Unknown selection: '$selected_entry'"
            fi
        fi
    else
        echo "gcba => git change branch all. But you aren't in a git directory."
    fi
}

# Change your current git branch to master
# Usage: gcbm
gcbm () {
    if in_git_folder; then
        git checkout master
    else
        echo "gcbm => git change branch master. But you aren't in a git directory."
    fi
}

# Delete git branches
# Usage: gdb
gdb () {
    if in_git_folder; then
        local local_branches branches
        local_branches="$( git branch | grep -v -e '^\*' -e ' master' | sed -E 's/^ +| +$//g' | sort -r )"
        if [[ -n "$local_branches" ]]; then
            branches="$( echo "$local_branches" | fzf -m --cycle --header="Select branches to delete using tab. Press enter when ready (or esc to cancel)." )"
            if [[ -n "$branches" ]]; then
                for branch in $( echo -E "$branches" | sed -l '' ); do
                    __git_echo_do git branch -D "$branch"
                done
            else
                echo "No branches selected for deletion."
            fi
        else
            echo "No branches to delete."
        fi
    else
        echo "gdb => git delete branches. But you aren't in a git directory."
    fi
}

# Branch Name - outputs the name of the current branch.
# Usage: bn
bn () {
    if in_git_folder; then
        echo "$( git branch | grep '^\*' | sed 's/^\* //' )"
        return 0
    else
        >&2 echo "Not in a git repo."
        return 1
    fi
}

# Git Pull Merge (master) - Goes to master, does a pull, goes back to your other branch and does a merge.
# Usage: gpm
gpm () {
    if in_git_folder; then
        __git_echo_do git checkout master \
        && __git_echo_do git pull \
        && __git_echo_do git checkout - \
        && __git_echo_do git merge master
        __git_echo_do git status
    else
        echo "Not in a git repo."
    fi
}

# Git Set Upstream - Sets the upstream branch for the current branch of the repo you're in.
# Usage: gsu
gsu () {
    if in_git_folder; then
        local branch cmd
        branch="$( git branch | grep '^\*' | sed 's/^\* //' )"
        __git_echo_do git branch "--set-upstream-to=origin/$branch" "$branch"
        __git_echo_do git pull
    fi
}

# Clean up a git repo
# Usage: clean_git_repo
clean_git_repo () {
    if in_git_folder; then
        __git_echo_do git checkout master
        __git_echo_do gdb
        __git_echo_do git clean -fdx -e .idea
        __git_echo_do git branch -r | grep -v 'HEAD' | xargs -L 1 git branch -rD
        __git_echo_do git fetch
    else
        echo "Not in a git repo."
    fi
}

# Create a fresh branch
# Usage: gfb <branch name>
gfb () {
    local branch
    branch="$1"
    if in_git_folder; then
        if [[ -z "$branch" ]]; then
            >&2 echo -E "Usage: gfb <branch name>"
            return 1
        fi
        if [[ "$( bn )" != 'master' ]]; then
            __git_echo_do git checkout master
        fi
        __git_echo_do git pull
        __git_echo_do git checkout -b "$branch"
    else
        echo "gfb => git fresh branch. But you aren't in a git repo."
        return 1
    fi
}

list_extra_branches () {
    local cwd repos repo branches
    cwd="$( pwd )"
    repos=( $( __git_get_all_repos ) )
    for repo in "${repos[@]}"; do
        cd "$repo"
        branches="$( git branch )"
        if [[ -n "$( echo -E "$branches" | grep -v '^[* ] master$' )" ]]; then
            echo -e -n "\033[1;37m"
            echo -E -n "$repo"
            echo -e "\033[0m"
            echo -e "$( echo -E "$branches" | sed -E 's/^[*] (.+)$/* \\033[1;32m\1\\033[0m/; s/^/  /;' )"
        fi
    done
    cd "$cwd"
}

master_pull_all () {
    local cwd repos repo_count repo_index failed_repos repo repo_failed cur_branch
    cwd="$( pwd )"
    repos=( $( __git_get_all_repos ) )
    repo_count="${#repos[@]}"
    repo_index=0
    failed_repos=()
    for repo in "${repos[@]}"; do
        repo_index=$(( repo_index + 1 ))
        repo_failed=
        cur_branch=
        echo -e "\033[1;36m$repo_index of $repo_count\033[0m - \033[1;33m$repo\033[0m"
        __git_echo_do cd "$repo" || repo_failed='YES'
        if [[ -z "$repo_failed" ]]; then
            cur_branch="$( bn )"
            if [[ "$cur_branch" != 'master' ]]; then
                __git_echo_do git checkout master || repo_failed='YES'
            fi
        fi
        if [[ -z "$repo_failed" ]]; then
            __git_echo_do git pull || repo_failed='YES'
        fi
        if [[ -z "$repo_failed" && "$cur_branch" != 'master' ]]; then
            __git_echo_do git checkout "$cur_branch" || repo_failed='YES'
        fi
        if [[ -n "$repo_failed" ]]; then
            echo -e "\033[1;38;5;231;48;5;196m An error occurred \033[0m - \033[1;31mSee above\033[0m"
            failed_repos+=( "$repo" )
        fi
    done
    if [[ "$cwd" != "$( pwd )" ]]; then
        __git_echo_do cd "$cwd"
    fi
    if [[ "${#failed_repos[@]}" -gt '0' ]]; then
        echo -e "\033[1;31m${#failed_repos[@]} repo(s) ran into problems:\033[0m"
        for repo in "${failed_repos[@]}"; do
            echo -e "    \033[35m$repo\033[0m"
        done
    else
        echo -e "\033[1;32mAll repos successfully pulled master.\033[0m"
    fi
}

git_diff_analysis () {
    local usage
    usage="$( cat << EOF
git_diff_analysis - Gets some stats on branch differences.

Usage: git_diff_analysis [<main branch> [<branch with changes>]]

    If no branches are supplied, the diff will be from master to your current branch.
    If only one branch is supplied, the diff will be from that branch to your current branch.
    If two brances are supplied, the diff will be from the first branch to the second.
EOF
)"
    local branches
    branches=()
    while [[ "$#" -gt '0' ]]; do
        case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        *)
            branches+=( "$1" )
            ;;
        esac
        shift
    done
    if [[ "${#branches[@]}" -gt '2' ]]; then
        printf 'Only two branches can be supplied.\n' >&2
        return 1
    elif [[ "${#branches[@]}" -eq '0' ]]; then
        branches+=( 'master' )
    fi
    if ! in_git_folder; then
        printf 'This command must be run from a git folder.\n' >&2
        return 1
    fi
    local diff_numstats test_entries code_entries
    local total_lines_added total_lines_removed test_lines_added test_lines_removed code_lines_added code_lines_removed
    printf '\033[97mgit diff %s --numstat\033[0m\n' "${branches[*]}"
    diff_numstats="$( git diff ${branches[@]} --numstat )"
    test_entries="$( grep 'src/test' <<< "$diff_numstats" )"
    code_entries="$( grep -v 'src/test' <<< "$diff_numstats" )"

    total_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$diff_numstats" )"
    total_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$diff_numstats" )"
    test_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$test_entries" )"
    test_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$test_entries" )"
    code_lines_added="$( awk '{sum+=$1} END { print sum }' <<< "$code_entries" )"
    code_lines_removed="$( awk '{sum+=$2} END { print sum }' <<< "$code_entries" )"

    printf '%7s  %7s  code changes\n' "+$code_lines_added" "-$code_lines_removed"
    printf '%7s  %7s  test code changes\n' "+$test_lines_added" "-$test_lines_removed"
    printf -- '------------------------------------\n'
    printf '%7s  %7s  total changes\n' "+$total_lines_added" "-$total_lines_removed"
}

__git_get_all_repos () {
    local base_dirs repos cwd base_dir repo
    base_dirs=()
    if [[ -n "$GITLAB_REPO_DIR" && -d "$GITLAB_REPO_DIR" ]]; then
        base_dirs+=( "$GITLAB_REPO_DIR" )
    elif [[ -n "$GITLAB_BASE_DIR" && -d "$GITLAB_BASE_DIR" ]]; then
        # GITLAB_BASE_DIR is deprecated, use GITLAB_REPO_DIR instead.
        base_dirs+=( "$GITLAB_BASE_DIR" )
    fi
    if [[ -n "$GITHUB_BASE_DIR" && -d "$GITHUB_BASE_DIR" ]]; then
        base_dirs+=( "$GITHUB_BASE_DIR" )
    fi
    repos=()
    if [[ "${#base_dirs[@]}" -gt '0' ]]; then
        cwd="$( pwd )"
        for base_dir in "${base_dirs[@]}"; do
            for repo in $( ls -d $base_dir/*/ ); do
                if [[ -d "$repo" ]]; then
                    cd "$repo"
                    if [[ "$?" -eq '0' ]] && in_git_folder; then
                        repos+=( "$repo" )
                    fi
                fi
            done
        done
        cd "$cwd"
    fi
    echo -E -n "${repos[@]}"
}

return 0
