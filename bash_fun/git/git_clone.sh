#!/bin/bash
# This file contains the git_clone function that cds into the cloned directory and sets some other stuff up for me.
# This file can be sourced to add the git_clone function to your environment.
# This file can also be executed to run the git_clone function without adding it to your environment.
#
# File contents:
#   git_clone  --> Clones a git repo and sets some stuff up for it.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_clone () {
    local usage personal here force stay git_args last_arg provided_dir cwd showed_cd dirs_pre ec dirs_post new_dir ec2
    usage="$( cat << EOF
Usage: git_clone [--personal] [--here [--force]] [--stay] <args for git clone>
    --personal indicates that it is a personal repo.
    --here indicates you want the new directory where you are rather than in \$GIT_REPO_DIR.
    --force allows cloning into another git repo when using --here in a git repo.
    --stay prevents git_clone from changing into the newly cloned repo.
    <args for git clone> are anything that you would traditionally provide to the git clone command.
EOF
)"
    git_args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help)
                printf '%s\n' "$usage"
                return 0
                ;;
            --personal)
                personal='YES'
                ;;
            --here)
                here='YES'
                ;;
            --force)
                force='YES'
                ;;
            --stay)
                stay='YES'
                ;;
            *)
                last_arg="$1"
                git_args+=( "$1" )
                ;;
        esac
        shift
    done
    cwd="$( pwd )"
    # In the git clone syntax ends with [--] <repository> [<directory>]
    # If a <directory> is provided, we need to know so that we can react appropriately.
    # If the last argument doesn't have any slashes before the first colon, it's definitely a <repository> url.
    # Otherwise, it's either a <directory> or a local <repository>.
    # If it's a local <repository>, we can actually check if it's a git directory.
    # Basically, if it has a colon without a slash before it, or if it's a local git directory, it's a <repository>.
    #   Otherwise, it's a <directory>.
    # Since we only care when it's a <directory> we negate that logic.
    if [[ ! "$last_arg" =~ ^[^/]*: && 'true' != "$( cd "$last_arg" 2> /dev/null && git rev-parse --is-inside-git-dir 2> /dev/null )" ]]; then
        provided_dir="$last_arg"
    fi
    # If a destination directory was provided, and it's not absolute, either go to the default location or check the current location.
    if [[ ! "$provided_dir" =~ ^/ ]]; then
        if [[ -z "$here" ]]; then
            if [[ -z "$GIT_REPO_DIR" ]]; then
                printf 'environment variable not defined: GIT_REPO_DIR.\n' >&2
                return 1
            fi
            if [[ "$cwd" != "$GIT_REPO_DIR" ]]; then
                __git_echo_do cd "$GIT_REPO_DIR" || return $?
                showed_cd='YES'
            fi
        elif [[ -z "$force" ]] && in_git_repo; then
            printf 'Already in a git repo. Use --force to bypass this check.\n' >&2
            return 1
        fi
    fi
    # I like the default behavior of git clone that shows some progress info as it runs.
    # It's sent to stderr, but if you try to capture/redirect it, git clone will only output the "Cloning into ..." line.
    # So, in order to capture the output and still default to showing the progress info, I'd have to mess around a lot more
    #   with the arguments being provided, which I'm trying to avoid.
    # Doing so would probably mean that this function should also need to detect capture/redirect and alter the output similarly.
    # It's just a bag of worms I don't want to open.
    # So capturing and parsing the git clone output isn't really an option for identifying the destination directory.
    # If we didn't detect one provided as arguments, we'll just use a before/after comparison of ls results to identify it.
    # There's some obvious possible problems with this, but hopefully they are rare.
    # If this does become a problem, another option is to identify the <repository> while parsing arguments, and try
    #   to replicate the parsing logic that git clone uses to extract the destination directory name.
    [[ -z "$provided_dir" ]] && dirs_pre="$( ls -d */ 2> /dev/null )"
    __git_echo_do git clone "${git_args[@]}"
    ec=$?
    if [[ "$ec" -eq '0' ]]; then
        if [[ -z "$provided_dir" ]]; then
            # Get the new list of directories and find the new entry (hopefully not entries).
            dirs_post="$( ls -d */ 2> /dev/null )"
            new_dir="$( printf '%s\n%s\n' "$dirs_pre" "$dirs_post" | sort | uniq -u | grep -v '^[[:space:]]*$' )"
        else
            new_dir="$provided_dir"
        fi
        if [[ -z "$new_dir" ]]; then
            printf 'No new directory detected.\n' >&2
            ec=1
        elif [[ "$( wc -l <<< "$new_dir" )" -gt '1' ]]; then
            printf 'Multiple new directories detected, cannot proceed:\n%s\n' "$new_dir" >&2
            ec=1
            new_dir=''
        else
            if [[ -n "$stay" ]]; then
                cd "$new_dir" || ec=$?
            else
                __git_echo_do cd "$new_dir" || ec=$?
                showed_cd='YES'
            fi
            if [[ "$ec" -eq '0' ]]; then
                new_dir="$( pwd )"
            else
                new_dir=''
            fi
        fi
    fi
    if [[ "$ec" -eq '0' ]]; then
        __git_echo_do git_set_default_branch
        if [[ -n "$personal" ]]; then
            __git_echo_do github_config_as_personal || ec2=$?
        else
            __git_echo_do git remote get-url origin || ec2=$?
            __git_echo_do git config --local --list || ec2=$?
            printf '\n'
        fi
    fi
    if [[ -n "$stay" || "$ec" -ne '0' ]]; then
        [[ -n "$new_dir" ]] && printf 'New repo directory: %s\n\n' "$new_dir"
        if [[ "$cwd" != "$( pwd )" ]]; then
            if [[ -n "$showed_cd" ]]; then
                __git_echo_do cd "$cwd"
            else
                cd "$cwd"
            fi
        fi
    fi
    if [[ "$ec2" -ne '0' ]]; then
        return "$ec2"
    fi
    return $ec
}

if [[ "$sourced" != 'YES' ]]; then
    where_i_am="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
    require_command () {
        local cmd cmd_fn
        cmd="$1"
        if ! command -v "$cmd" > /dev/null 2>&1; then
            cmd_fn="$where_i_am/$cmd.sh"
            if [[ -f "$cmd_fn" ]]; then
                source "$cmd_fn"
                if [[ "$?" -ne '0' ]] || ! command -v "$cmd" > /dev/null 2>&1; then
                    ( printf 'This script relies on the [%s] function.\n' "$cmd"
                      printf 'The file [%s] was found and sourced, but there was a problem loading the [%s] function.\n' "$cmd_fn" "$cmd" ) >&2
                    return 1
                fi
            else
                ( printf 'This script relies on the [%s] function.\n' "$cmd"
                  printf 'The file [%s] was looked for, but not found.\n' "$cmd_fn" ) >&2
                return 1
            fi
        fi
    }
    require_command '__git_echo_do' || exit $?
    require_command 'in_git_folder' || exit $?
    require_command 'github_config_as_personal' || exit $?
    git_clone "$@"
    exit $?
fi
unset sourced

return 0
