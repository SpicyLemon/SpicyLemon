#!/bin/bash
# This file contains lots of handy functions for dealing with git.
# File contents:
#   in_git_folder  ---> Helper function for testing if you're currently in a git folder.
#   gcb  -------------> Git Change Branch - Select a branch and switch to it.
#   gcba  ------------> Git Change Branch (All) - Gets a list of all branches (local and remote) and lets you pick one to checkout.
#   gdb  -------------> Git Delete Branches - Select branches that you want to delete, and then deletes them.
#   bn  --------------> Branch Name - Outputs your current branch name.
#   gpm  -------------> Git Pull Merge (Master) - Pull master and merge it into your branch.
#   gsu  -------------> Git Set Upstream - Sets the upstream appropriately for the repo and branch you're in.
#   clean_git_repo  --> Takes several actions to help you clean up a git repo.
#   gfb  -------------> Pulls master and creates a fresh branch from it.
#
# Depends on:
#   echo_do - Function defined in generic.sh - source generic.sh
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
Please run this command to enable the functionality contained in within.
$( echo -e "\033[1;37msource $( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )\033[0m" )
EOF
    exit 1
fi

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
            selection="$( git branch | fzf +m | sed -E 's/^[* ]+//' )"
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
        selected_entry="$( echo -E "$all_branches" | sort -t '~' -k 3 -k 2 | column -s '~' -t | fzf +m )"
        if [[ -n "$selected_entry" ]]; then
            remote_and_branch_rx='^[* ] +([^ ]+) +(.+)$'
            just_branch_rx='^[* ] +(.+)$'
            if [[ "$selected_entry" =~ $remote_and_branch_rx ]]; then
                remote="${BASH_REMATCH[1]}"
                branch="${BASH_REMATCH[2]}"
                echo_do "git checkout --track \"$remote/$branch\""
            elif [[ "$selected_entry" =~ $just_branch_rx ]]; then
                branch="${BASH_REMATCH[1]}"
                echo_do "git checkout \"$branch\""
            else
                echo -E "Unknown selection: '$selected_entry'"
            fi
        fi
    else
        echo "gcba => git change branch all. But you aren't in a git directory."
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
                    echo_do "git branch -D \"$branch\""
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
        echo_do "git checkout master" \
        && echo_do "git pull" \
        && echo_do "git checkout -" \
        && echo_do "git merge master" \
        && echo_do "git status"
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
        cmd="git branch --set-upstream-to=origin/$branch $branch"
        echo_do "$cmd"
        echo_do "git pull"
    fi
}

# Clean up a git repo
# Usage: clean_git_repo
clean_git_repo () {
    if in_git_folder; then
        echo_do "git checkout master"
        echo_do "gdb"
        echo_do "git clean -fdx -e .idea"
        echo_do "git branch -r | grep -v 'HEAD' | xargs -L 1 git branch -rD"
        echo_do "git fetch"
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
            echo_do "git checkout master"
        fi
        echo_do "git pull"
        echo_do "git checkout -b $branch"
    else
        echo "gfb => git fresh branch. But you aren't in a git repo."
        return 1
    fi
}
