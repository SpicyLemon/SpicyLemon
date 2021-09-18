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
    local usage personal here force stay git_args cwd dirs_pre ec dirs_post new_dir ec2
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
                git_args+=( "$1" )
                ;;
        esac
        shift
    done
    cwd="$( pwd )"
    if [[ -z "$here" ]]; then
        if [[ -z "$GIT_REPO_DIR" ]]; then
            printf 'environment variable not defined: GIT_REPO_DIR.\n' >&2
            return 1
        fi
        __git_echo_do cd "$GIT_REPO_DIR" || return $?
    elif [[ -z "$force" ]] && in_git_repo; then
        printf 'Already in a git repo. Use --force to bypass this check.\n' >&2
        return 1
    fi
    dirs_pre="$( ls -d */ )"
    __git_echo_do git clone "${git_args[@]}"
    ec=$?
    if [[ "$ec" -eq '0' ]]; then
        dirs_post="$( ls -d */ )"
        new_dir="$( printf '%s\n%s\n' "$dirs_pre" "$dirs_post" | sort | uniq -u )"
        if [[ -z "$new_dir" ]]; then
            printf 'No new directory detected.\n' >&2
            ec=1
        else
            cd "$new_dir" || ec=$?
            if [[ "$ec" -eq '0' ]]; then
                new_dir="$( pwd )"
            else
                new_dir=''
            fi
        fi
    fi
    if [[ "$ec" -eq '0' && -n "$personal" ]]; then
        __git_echo_do github_config_as_personal || ec2=$?
    fi
    if [[ -n "$stay" || "$ec" -ne '0' ]]; then
        [[ -n "$new_dir" ]] && printf 'New repo directory: %s\n' "$new_dir"
        __git_echo_do cd "$cwd"
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
