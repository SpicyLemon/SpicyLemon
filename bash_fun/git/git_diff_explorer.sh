#!/bin/bash
# This file contains the git_diff_explorer function that uses fzf to explore git diffs.
# This file can be sourced to add the git_diff_explorer function to your environment (and __gde_* helpers).
# This file can also be executed to run the git_diff_explorer function without adding it to your environment.
#
# If you source this file then move it, you'll need to source it again for it to continue to work.
#
# File contents:
#   functions:
#       git_diff_explorer  ------> Uses fzf to display and explore git diffs.
#   "private" functions:
#       __gde_preview  ----------> Creates the diff output for the fzf preview window.
#       __gde_get_root_dir  -----> Figures out the root directory of the compact summary entries.
#       __gde_parse_filenames  --> Parses the file path(s) from the summary line.
#   exported variables:
#       GIT_DIFF_EXPLORER_CMD  --> The absolute path and name of this file.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# The __gde_preview function is used in the preview window of fzf for the json_explorer.
# But fzf can't find local (non-exported) environment functions, and exporting functions isn't always an option (see shellshock).
# In order to get around that, in here, we'll just set and export the GIT_DIFF_EXPLORER_CMD env var that points to this file.
# When invoking it for preview, we'll provide a --gde-preview flag that only gets looked at when this file is invoked as a script.
# Not all systems allow for readlink -f, so I'm using dirname/pwd and basename instead.
export GIT_DIFF_EXPLORER_CMD="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )/$( basename "${BASH_SOURCE:-$0}" )"

git_diff_explorer () {
    if [[ -z "$GIT_DIFF_EXPLORER_CMD" ]]; then
        printf 'git_diff_explorer: Missing required environment variable GIT_DIFF_EXPLORER_CMD.\n' >&2
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
    local usage args commit delimiter output_type arg pargs summary_cmd summary selected root_dir line
    usage="$( cat << EOF
git_diff_explorer - Displays a git diff compact summary in fzf and shows individual file diffs in the preview window.
Selected files are then printed with their paths relative to your current directory.

Usage: git_diff_explorer [<git diff args>] [--commit <hash>] [--output-type <output type>] [--print0] [--printd <delimiter>]

    <git diff args> are the arguments to provide to git diff (see: git diff help).

    --commit <hash> lets you provide a commit hash to get the diff of just that commit.
        It is just a shortcut to providing <hash>~ <hash> as <git diff args>.

    --output-type <output type> sets the output type you want for the entries that were selected.
        This option only has meaning on lines describing a moved file.
        <output type> can be one of:
            old: output a single line with just the old file.
            new: output a single line with just the new file (or the old file if it's not a moved entry).
            combined: output the combined oldfile => newfile entry (or just the old file if it's not a moved entry).
            both: output the old file, and if the new file is different, output that too on a second line.
        Default is combined.
        If provided multiple times, the last one provided is used.

    --print0
        Print an ASCII NUL character (character code 0) after each selection instead of a newline.
        Overrides a previously provided --printd option. I.e. the --print0 or --printd option that is provided last, is used.

    --printd <delimiter>
        Print the provided delimiter after each selection.
        Default is a newline.
        If provided multiple times, the last one provided is used.
        Overrides a previously provided --print0 option. I.e. the --print0 or --printd option that is provided last, is used.

Selection window command: git diff --compact-summary --color=always <git diff args>
Preview window command: git diff --color=always <git diff args> -- <file(s)>

Note: The -- separator cannot be provided to git_diff_explorer (e.g. as part of <git diff args>).

EOF
)"
    args=()
    delimiter='\n'
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
            --commit=*)
                commit="$( printf '%s' "$1" | sed 's/^--commit=//' )"
                if [[ -z "$commit" ]]; then
                    printf 'Missing value after [%s] flag.\n' "$1" >&2
                    return 1
                fi
                args+=( "$commit~" "$commit" )
                ;;
            --output-type)
                if [[ -z "$2" ]]; then
                    printf 'Missing argument after [%s] flag.\n' "$1" >&2
                    return 1
                fi
                output_type="$2"
                shift
                ;;
            --output-type=*)
                output_type="$( printf '%s' "$1" | sed 's/^--output-type=//' )"
                ;;
            --print0)
                delimiter='\0'
                ;;
            --printd)
                if [[ -z "$2" ]]; then
                    printf 'Missing argument after [%s] flag.\n' "$1" >&2
                    return 1
                fi
                delimiter="$2"
                shift
                ;;
            --printd=*)
                delimiter="$( printf '%s' "$1" | sed 's/^--printd=//' )"
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
    if [[ -z "$output_type" ]]; then
        output_type='combined'
    fi
    # For providing the args to the preview, they'll need to be escaped and combined into a single string, with each wrapped in quotes.
    # This is better than pargs="$( printf '%q ' "${args[@]}" )" at least in the case where there aren't any args.
    # If there aren't any args, that sets pargs to a single-quoted empty string, e.g. pargs="''".
    # Then when provided to the fzf preview command, there'd an extry empty string argument being provided, which git diff then complains about.
    pargs="$( for arg in "${args[@]}"; do printf '%s ' "'$( sed 's/'"'"'/\\'"'"'/g' <<< "$arg" )'"; done )"
    # Get the compact summary and reverse the whole thing.
    # By default, fzf reverses the provided lines. So if we just got the summary and piped it to fzf, fzf would show it upside-down.
    # We reverse it right off the bat so that the summary line (e.g. "21 files changed, 757 insertions(+), 31 deletions(-)")
    # is first and can then be used as the "header" line (which is really stickied to the bottom).
    # Then send that on to fzf with the following:
    #   --ansi so that the color output from git is displayed right.
    #   --header-lines 1 stickies the first line to the bottom.
    #       Since we reversed the compact summary before giving it to fzf, the first line should be the summary line,
    #       e.g. "21 files changed, 757 insertions(+), 31 deletions(-)".
    #   --cycle so you can get to the top or bottom easily.
    #   --multi so you can select multiple files for final output.
    #   --scroll-off 2 makes fzf scroll the list with the 2 rows above and below always visible (except at the top and bottom of the list).
    #       Because of --cycle, it's easy to be looking at the preview, not paying attention to the list, and not realize you cycled.
    #       This helps with that a bit by making it easier to identify when you're close to the top or bottom before cycling.
    #       E.g. if the highlighted line is at the bottom of the selection window, you have the last line highlighted.
    #   --tac reverses the selectable lines (back to upside-down ordering (stay with me)).
    #   --layout reverse-list then effectively undoes the --tac (back to normal order) but starts fzf with the first line selected.
    #       Without both of --tac and --layout reverse-list, the list is in the same order, but the initially highlighted line is the last one.
    #   --preview-window defines the preview window layout:
    #       top: Put it at the top (with the compact summary at the bottom).
    #       75%: Have it take up 75% of the screen.
    #       border-bottom: Put a dividing border at the bottom to separate it from the compact summary.
    #                      I felt a full border around it just took up extra space and didn't really help anything.
    #       wrap: Wrap long lines so that you can see their diffs fully.
    #       ~2: Keep the first two preview lines visible a the top. These should be the diff command (including filename(s)).
    #   --preview defines the command to run to create the contents of the preview window.
    #       This command is run in a separate enviroment, so it doesn't have access to unexported functions.
    #       And due to shellshock, exporting functions is rarely an option anymore. So we call this file directly with the --gde-preview flag.
    #       We also provide all the args that were provided to this function, but have to escape them specially so that they
    #       translate properly back into their respective arguments.
    #       Lastly, we provide a -- to indicate we're done with args followed by the compact summary line currently highlighted.
    summary_cmd=( git --no-pager diff --compact-summary --color=always "${args[@]}" )
    if [[ -n "$DEBUG" ]]; then
        {
            printf '% 12s:%s\n' 'args' "$( [[ "${#args[@]}" -gt '0' ]] && printf ' %q' "${args[@]}" )"
            printf '% 12s: [%s]\n' 'delimiter' "$delimiter"
            printf '% 12s: [%s]\n' 'output_type' "$output_type"
            printf '% 12s:%s\n' 'summary_cmd' "$( printf ' %q' "${summary_cmd[@]}" )"
        } >&2
    fi
    summary="$( "${summary_cmd[@]}" )" || return $?
    selected="$(
            tac <<< "$summary" \
            | fzf --ansi --header-lines 1 --cycle --multi --scroll-off 2 --tac --layout reverse-list \
                  --preview-window='top,75%,border-bottom,wrap,~2' \
                  --preview="$GIT_DIFF_EXPLORER_CMD --gde-preview $pargs -- {}"
    )" || return $?
    if [[ -n "$selected" ]]; then
        root_dir="$( __gde_get_root_dir "${args[@]}" )"
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                printf '%s%b' "$( __gde_parse_filenames "$root_dir" "$output_type" "$line" )" "$delimiter"
            fi
        done <<< "$selected"
    fi
    return 0
}

# __gde_preview - outputs the diff of a specific file that it gets from a line from a --compact-summary.
# Usage: __gde_preview <git diff args> -- <compact summary line>
__gde_preview () {
    local args line files file diff_cmd output rc
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
        esac
        shift
    done
    if [[ -z "$line" ]]; then
        printf 'No line provided.\n' >&2
        return 1
    fi
    files=()
    while IFS= read -r file; do files+=( "$file" ); done <<< "$( __gde_parse_filenames "$( __gde_get_root_dir "${args[@]}" )" both "$line" )"
    # Put together the full diff command and output it, but output the file(s) on a second line.
    diff_cmd=( git --no-pager diff --color=always "${args[@]}" -- )
    printf '%s\\\n %s\n' "$( printf '%q ' "${diff_cmd[@]}" )" "$( printf ' %q' "${files[@]}" )"
    diff_cmd+=( "${files[@]}" )
    if [[ -n "$DEBUG" ]]; then
        {
            printf '% 12s:%s\n' 'args' "$( [[ "${#args[@]}" -gt '0' ]] && printf ' %q' "${args[@]}" )"
            printf '% 12s:%s\n' 'files' "$( printf ' %q' "${files[@]}" )"
            printf '% 12s:%s\n' 'diff_cmd' "$( printf ' %q' "${diff_cmd[@]}" )"
        } >&2
    fi
    # Run the diff command and output the result.
    output="$( "${diff_cmd[@]}" )"
    rc=$?
    if [[ -n "$output" ]]; then
        printf '%s\n' "$output"
    else
        printf 'No differences to display.\n'
    fi
    return $rc
}

# __gde_get_root_dir figures out the path to the root directory of the compact summary entries.
# Usage: __gde_get_root_dir <args>
__gde_get_root_dir () {
    local root_dir
    # If --relative is provided, the compact summary is relative to either a provided value or .
    # If --no-relative is provided (e.g. after --relative) then it goes back to normal.
    # Normally, the compact summary lists files relative to the root of the repository.
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --)
                set -- --
                ;;
            --relative=*)
                root_dir="$( printf '%s' "$1" | sed 's/^--relative=//' )"
                ;;
            --relative)
                if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                    root_dir="$2"
                    shift
                else
                    root_dir=.
                fi
                ;;
            --no-relative)
                root_dir=''
                ;;
        esac
        shift
    done
    if [[ -z "$root_dir" ]]; then
        root_dir="$( git rev-parse --show-toplevel )"
    fi
    if [[ "$root_dir" != '.' && "$( cd "$root_dir"; pwd -P )" == "$( pwd -P )" ]]; then
        root_dir=.
    fi
    printf '%s' "$root_dir"
    return 0
}

# __gde_parse_filenames parses the file path(s) from the summary line.
# Usage: __gde_parse_filenames <root dir> <out type> <summary line>
# <root dir> is the path (absolute or realtive to the current directory) to the root dir of the summary entry.
# <out type> is one of: old, new, combined, both
#   old: output a single line with just the old file.
#   new: output a single line with just the new file (or the old file if it's not a moved entry).
#   combined: output the combined oldfile => newfile entry (or just the old file if it's not a moved entry).
#   both: output the old file, and if the new file is different, output that too on a second line.
__gde_parse_filenames () {
    local root_dir out_type line file_entry rp rpec combined oldfile newfile
    root_dir="$1"
    shift
    out_type="$1"
    shift
    line="$*"
    # A compact summary line has this format:
    #   <file> <info> | <number> <plusses and minuses>
    # If the <file> was one that moved, it will either be '<old> => <new>' or it will have a part that is {<old> => <new>}.
    # The <info> part is optional, and is some stuff in parenthases if provided, e.g. "(gone)". There will always be a space before and after it.
    # The <plusses and minuses> probably still has color escape codes in/around it too.
    # So to get the <file> from the line:
    #   1: Strip out leading whitespace optionaly followed by .../
    #   2: Strip out the optional ' <info>', and any other whitespace before the | until the end of the line.
    # Since color might still be involved too, just assume everything after | <space> <numbers> isn't important here.
    file_entry="$( sed -E 's/^[[:space:]]+(\.\.\.\/)?//; s/([[:space:]]+\([^)]+\))?[[:space:]]+\|[[:space:]]+[[:digit:]]+.*$//;' <<< "$line" )"
    # The compact summary usually outputs the path to the files from the root of the repo (but not always).
    # But git diff needs paths either relative to the current directory or absoulte (starting with /).
    # Moved entries are either 'oldfile => newfile' or have a part that is '{oldfile => newfile}'.
    # If it's not a moved file, all of combined, oldfile, and newfile are the same.
    # If it is a moved file, we need to do some fancy path manipulation and splitting in order to get the paths needed by git diff.
    if [[ ! "$file_entry" =~ ' => ' ]]; then
        # Not a moved file.
        rp=''
        rpec=0
        if command -v realpath > /dev/null 2>&1; then
            rp="$( realpath --relative-to=. "$root_dir/$file_entry" 2> /dev/null )"
            rpec=$?
        fi
        if [[ "$rpec" -eq '0' && -n "$rp" ]]; then
            combined="$rp"
        elif [[ "$root_dir" == '.' ]]; then
            combined="$file_entry"
        else
            combined="$root_dir/$file_entry"
        fi
        oldfile="$combined"
        newfile="$combined"
    else
        # It's a moved file.
        # First, create the combined path. Either make it relative or as short as easily possible.
        rp=''
        rpec=0
        if command -v realpath > /dev/null 2>&1; then
            # Since we have realpath, create a preliminary version of the combined line by appending
            # the root dir and making sure the split parts are wrapped in {}.
            # Temporarily repurpose combined, oldfile, and newfile to hold the preliminarily combined version
            # and the parts begore (oldfile) and after (newfile) the beginning of the split.
            if [[ ! "$file_entry" =~ { ]]; then
                combined="$root_dir/{$file_entry}"
            else
                combined="$root_dir/$file_entry"
            fi
            # We can only use realpath on the part before the split. So do that then tack the rest back on.
            oldfile="$( sed 's/{.*$//' <<< "$combined" )"
            newfile="$( sed 's/^[^{]*{/{/' <<< "$combined" )"
            rp="$( realpath --relative-to=. "$oldfile" 2> /dev/null )/$newfile"
            rpec=$?
        fi
        if [[ "$rpec" -eq '0' && -n "$rp" ]]; then
            combined="$rp"
        elif [[ "$root_dir" == '.' ]]; then
            # reaplath isn't available, but we're in the root dir, just use the raw file entry.
            combined="$file_entry"
        elif [[ ! "$file_entry" =~ { ]]; then
            # realpath isn't available, and the entire entry is just 'oldfile => newfile'.
            # Put the root dir on and wrap the file entry in {}.
            combined="$root_dir/{$file_entry}"
        else
            # realpath isn't available, and the entry already has {} so we just need to tack the root dir onto it.
            combined="$root_dir/$file_entry"
        fi
        # Then split the combined path into the new and old paths.
        oldfile="$( sed -E 's/{(.*) => (.*)}/\1/g; s/(.*) => (.*)/\1/g' <<< "$combined" )"
        newfile="$( sed -E 's/{(.*) => (.*)}/\2/g; s/(.*) => (.*)/\2/g' <<< "$combined" )"
    fi
    if [[ -n "$DEBUG" ]]; then
        {
            printf '% 12s: [%s]\n' 'root_dir' "$root_dir"
            printf '% 12s: [%s]\n' 'out_type' "$out_type"
            printf '% 12s: [%s]\n' 'line' "$line"
            printf '% 12s: [%s]\n' 'file_entry' "$file_entry"
            printf '% 12s: [%s]\n' 'combined' "$combined"
            printf '% 12s: [%s]\n' 'oldfile' "$oldfile"
            printf '% 12s: [%s]\n' 'newfile' "$newfile"
        } >&2
    fi
    # Do the output
    case "$out_type" in
        old) printf '%s\n' "$oldfile" ;;
        new) printf '%s\n' "$newfile" ;;
        combined) printf '%s\n' "$combined" ;;
        both)
            printf '%s\n' "$oldfile"
            if [[ "$oldfile" != "$newfile" ]]; then
                printf '%s\n' "$newfile"
            fi
            ;;
        *)
            printf 'Unknown output type: [%s]\n' "$out_type" >&2
            return 1
            ;;
    esac
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    if [[ "$1" == '--gde-preview' ]]; then
        shift
        __gde_preview "$@"
        exit $?
    fi
    git_diff_explorer "$@"
    exit $?
fi
unset sourced

return 0
