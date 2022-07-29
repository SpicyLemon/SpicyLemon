#!/bin/bash
# This file contains the git_diff_explorer function that uses fzf to explore git diffs.
# This file can be sourced to add the git_diff_explorer and git_diff_explorer_preview functions to your environment.
# This file can also be executed to run the git_diff_explorer function without adding it to your environment.
#
# File contents:
#   git_diff_explorer  ----------> Uses fzf to display and explore git diffs.
#   git_diff_explorer_preview  --> Creates the diff output for the fzf preview window.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

git_diff_explorer () {
    if [[ -z "$GIT_DIFF_EXPLORER_CMD" ]]; then
        printf 'Unable to locate git_diff_explorer file.\n' >&2
        return 1
    fi
    local do_not_run req_cmd
    for req_cmd in 'git' 'fzf' 'tac' "$GIT_DIFF_EXPLORER_CMD"; do
        if ! command -v "$req_cmd" > /dev/null 2>&1; then
            do_not_run='yes'
            printf 'git_diff_explorer: Missing required command: %s\n' "$req_cmd" >&2
            "$req_cmd"
        fi
    done
    if [[ -n "$do_not_run" ]]; then
        return 1
    fi
    local usage
    usage="$( cat << EOF
git_diff_explorer - Displays a git diff summary and shows individual files in the preview window.
Selected files are then printed.

Usage: git_diff_explorer <git diff options> [--commit <hash>]
    For details on <git diff options>, see: git help diff
    The --commit <hash> option lets you provide a commit hash to get the diff of that commit.
    It is just a shortcut to git diff <hash>~ <hash>.

For the selection window, --compact-summary is added to the provided <git diff options>.
For the preview window, the highlighted file is provided after the <git diff options>.

EOF
)"
    local args arg pargs summary_cmd summary selected line
    args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help|help)
                printf '%s\n' "$usage"
                return 0
                ;;
            --commit)
                if [[ -z "$2" ]]; then
                    printf 'Missing argument after [%s] flag.\n' "$1" >&2
                    return 1
                fi
                args+=( "$2~" "$2" )
                shift
                ;;
            --)
                printf 'The -- separator/argument cannot be provided as an argument to git_diff_explorer.\n' >&2
                return 1
                ;;
            *)
                args+=( "$1" )
                ;;
        esac
        shift
    done
    # For providing the args to the preview, they'll need to be escaped and combined into a single string, with each wrapped in quotes.
    # This is better than pargs="$( printf '%q ' "${args[@]}" )" at least in the case where there aren't any args.
    # If there aren't any args, that sets pargs to a single-quoted empty string, e.g. pargs="''".
    # Then when provided to the fzf preview command, there'd an extry empty string argument being provided, which git diff then complains about.
    pargs="$( for arg in "${args[@]}"; do printf '%s ' "'$( sed 's/'"'"'/\\'"'"'/g' <<< "$arg" )'"; done )"
    # Get the compact summary and reverse the whole thing.
    # Reversing it does two things:
    #   1) Puts the summary line at the top, making it usable as a header line in fzf.
    #   2) Undoes the reversing that fzf usually does. I.e. the summary shows up in fzf in the same order it would in your terminal.
    # Then send it on to fzf.
    #   --ansi so that the color output from git is displayed right. --header-lines 1 is the summary line (of the summary).
    #   --cycle so you can hit down to go the top right away. --multi so you can select multiple files for final output.
    #   For the preview, call this file with the --gde-preview flag and the rest of the args that were provided here; include
    #   the current line as a final arg. FZF's preview stuff can't use unexported functions. And thanks to shellshock, exporting
    #   functions is not really an option. That's why the git_diff_explorer_preview function isn't used directly for the preview.
    #   The preview window will take up the top 75% of the screen, there will be a border below it and the first 2 lines are the header.
    #   The first 2 lines should be the command being run to get the diff, with the 2nd line containing the file(s).
    summary_cmd=( git --no-pager diff --color=always --compact-summary "${args[@]}" )
    printf 'Summary command> ' && printf '%q ' "${summary_cmd[@]}" && printf '\n'
    summary="$( "${summary_cmd[@]}" )" || return $?
    selected="$(
            tac <<< "$summary" \
            | fzf --ansi --header-lines 1 --cycle --multi \
                  --preview="$GIT_DIFF_EXPLORER_CMD --gde-preview $pargs -- {}" \
                  --preview-window='top,75%,border-bottom,~2'
    )" || return $?
    if [[ -n "$selected" ]]; then
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                git_diff_explorer_preview "${args[@]}" -- "$line"
            fi
        done <<< "$selected"
    fi
    return 0
}

# git_diff_explorer_preview - outputs the diff of a specific file that it gets from a line from a --compact-summary.
# Usage: git_diff_explorer_preview <git diff args> -- <compact summary line>
git_diff_explorer_preview () {
    local args root_dir line file_entry file1_full file2_full file1 file2 diff_cmd output ec
    args=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --)
                shift
                line="$*"
                set -- --
                ;;
            *)
                args+=( "$1" )
                case "$1" in
                    --relative=*)
                        root_dir="$( printf '%s' "$1" | sed 's/^--relative=//' )"
                        ;;
                    --relative)
                        root_dir=.
                        ;;
                    --no-relative)
                        root_dir=''
                        ;;
                esac
                ;;
        esac
        shift
    done
    if [[ -z "$line" ]]; then
        printf 'No line provided.\n' >&2
        return 1
    fi
    if [[ -z "$root_dir" ]]; then
        root_dir="$( git rev-parse --show-toplevel )"
    fi
    # A compact summary line has this format:
    #   <file> <info> | <number> <plusses and minuses>
    # If the <file> was one that moved, it will either be '<old> => <new>' or it will have a part that is {<old> => <new>}.
    # The <info> part is optional, and is some stuff in parenthases if provided, e.g. "(gone)". There will always be a space before and after it.
    # So to get the <file> from the line:
    #   1: Strip out leading whitespace optionaly followed by .../
    #   2: Strip out the optional ' <info>', and any other whitespace before the | until the end of the line.
    # Since color might still be involved too, just assume everything after | <space> <numbers>
    file_entry="$( sed -E 's/^[[:space:]]+(\.\.\.\/)?//; s/([[:space:]]+\([^)]+\))?[[:space:]]+\|[[:space:]]+[[:digit:]]+.*$//;' <<< "$line" )"
    # If the file entry is for a moved file, we need to get both to provide to the git diff command.
    # Luckily, if we provide the same file twice, it only gets used once. So we can just always provide two.
    # Now it gets tricky. The compact summary usually outputs the path to the files from the root of the repo.
    # However, git diff expects the paths to be relative or absolute (starting with /).
    # That's why, above, we do special handling of the --relative and --no-relative flags to identify the relative path,
    # Or if not provided, we ask git to get us the repo's root directory.
    file1_full="$root_dir/$( sed -E 's/{(.*) => (.*)}/\1/g; s/(.*) => (.*)/\1/g' <<< "$file_entry" )"
    file2_full="$root_dir/$( sed -E 's/{(.*) => (.*)}/\2/g; s/(.*) => (.*)/\2/g' <<< "$file_entry" )"
    # Then, to make them easier for a human to look at, we'll try to make them relative to the current location.
    if command -v realpath > /dev/null 2>&1; then
        file1="$( realpath --relative-to=. "$file1_full" )"
        file2="$( realpath --relative-to=. "$file2_full" )"
    else
        file1="$file1_full"
        file2="$file2_full"
    fi
    # Put together the full diff command and output it, but output the file(s) on a second line.
    diff_cmd=( git --no-pager diff --color=always "${args[@]}" -- )
    [[ "$IN_PREVIEW" == 'YES' ]] || printf 'Diff command> '
    printf '%q ' "${diff_cmd[@]}"
    if [[ "$file1" == "$file2" ]]; then
        if [[ "$IN_PREVIEW" == 'YES' ]]; then
            printf '\\\n  %q\n' "$file1"
        else
            printf '%q\n' "$file1"
        fi
        diff_cmd+=( "$file1" )
    else
        if [[ "$IN_PREVIEW" == 'YES' ]]; then
            printf '\\\n  %q %q\n' "$file1" "$file2"
        else
            printf '%q %q\n' "$file1" "$file2"
        fi
        diff_cmd+=( "$file1" "$file2" )
    fi
    if [[ -n "$DEBUG" ]]; then
        printf '      line: [%s]\n' "$line"
        printf '  root_dir: [%s]\n' "$root_dir"
        printf 'file_entry: [%s]\n' "$file_entry"
        printf 'file1_full: [%s]\n' "$file1_full"
        printf 'file2_full: [%s]\n' "$file2_full"
        printf '     file1: [%s]\n' "$file1"
        printf '     file2: [%s]\n' "$file2"
        printf '      args: [%s]\n' "${args[*]}"
        printf '  diff_cmd:%s\n' "$( printf ' %q' "${diff_cmd[@]}" )"
    fi
    output="$( "${diff_cmd[@]}" )"
    ec=$?
    if [[ -n "$output" ]]; then
        printf '%s\n' "$output"
    else
        printf 'No differences to display.\n'
    fi
    if [[ "$IN_PREVIEW" != 'YES' ]]; then
        # In very light testing, I couldn't add to history when invoking this as a script.
        # However, just to be on the safe side, the IN_PREVIEW variable is used to make sure
        # the fzf preview command isn't adding to the history each time it's invoked.
        # We absolutely don't want that. What we do want is to add the specific diff commands to history
        # that are run because their entries were selected by the explorer.
        # In bash, history -s does this. In zsh, it is print -s. In zsh, executing history -s returns an error
        # about not having a -s option. So we hide all that possible output and just try both.
        history -s "${diff_cmd[@]}" > /dev/null 2>&1 || print -s "${diff_cmd[@]}" > /dev/null 2>&1
    fi
    return $ec
}

# The git_diff_explorer_preview command is used in the preview window of fzf for the json_explorer.
# But fzf can't find local (non-exported) environment functions, and exporting functions isn't always an option.
# In order to get around that, in here, we'll just set a GIT_DIFF_EXPLORER_CMD env var that points to this file.
# When invoking it for preview, we'll provide a --gde-preview flag that only gets looked at when this file is invoked as a script.
export GIT_DIFF_EXPLORER_CMD="$( readlink -f "${BASH_SOURCE:-$0}" )"

if [[ "$sourced" != 'YES' ]]; then
    if [[ "$1" == '--gde-preview' ]]; then
        shift
        IN_PREVIEW='YES'
        git_diff_explorer_preview "$@"
        exit $?
    fi
    git_diff_explorer "$@"
    exit $?
fi
unset sourced

return 0
