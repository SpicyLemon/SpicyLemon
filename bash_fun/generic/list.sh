#!/bin/bash
# This file contains the list function that outputs files and/or directories.
# This file can be sourced to add the list function to your environment.
# This file can also be executed to run the list function without adding it to your environment.
#
# File contents:
#   list  --> List files/dirs in the current directory.
#
# This exists because I wanted an easier way to get just the files or just the directories from a directory without extra characters.
# The command ls -d */  lists all directories but has a slash at the end of each.
# The find command (e.g. find . -type d --maxdepth 1) always includes the base directory followed by a slash at the start of each entry.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

list () {
    local args no_more_flags show_files show_dirs show_hidden show_dot terminator show_base directories arg i directory
    args=()
    no_more_flags=''
    show_files=''
    show_dirs=''
    show_hidden='NO'
    show_dot='NO'
    terminator='\n'
    show_base=''
    directories=()
    # Pre-process the provided args to allow multiple short-form ones to be provided at once, e.g. list -fdh
    while [[ "$#" -gt '0' ]]; do
        if [[ "$1" =~ ^-[^-]. ]]; then
            for i in $( seq 1 "$(( ${#1} - 1 ))" ); do
                args+=( "-${1:$i:1}" )
            done
        else
            args+=( "$1" )
        fi
        shift
    done
    # Now actually handle the args.
    for arg in "${args[@]}"; do
        if [[ "$no_more_flags" == 'YES' ]]; then
            directories+=( "$arg" )
            continue
        fi
        case "$arg" in
            # Note: Cannot use -h for help here since I wanted it for -h|--hidden
            help|--help) cat << EOF
list - lists files and/or directories.

Usage: list [-f|--files|-F|--no-files] [-d|--dirs|-D|--no-dirs] [-h|--hidden|-H|--no-hidden|-I|--hidden-only]
            [-t|--dot|-T|--no-dot] [-0|--print0|-n|--newline] [-b|--base|-a|--absolute|-B|--no-base]
            [[--] <directory> [<directory2>...]]

    -f or --files        Include files in the output.
    -F or --no-files     Do not include files in the output.
        If multiple of -f -F --files --no-files are given, the last one is used.
        Default behavior depends on the presence of other flags.
            If -d -d --dirs or --no-dirs is provided, the default is -F or --no-files
            Otherwise, the default is -f or --files.

    -d or --dirs         Include directories in the output.
    -D or --no-dirs      Do not include directories in the output.
        If multiple of -d -D --dirs --no-dirs are given, the last one is used.
        Default behavior depends on the presense of other flags.
            If -f -F --files or --no-files is provided, the default is -D or --no-dirs.
            Otherwise, the default is -d or --dirs.

    -h or --hidden       Include hidden files and/or directories in the output.
    -H or --no-hidden    Do not include hidden files and/or directories in the ouptut.
    -I or --hidden-only  Only include hidden files and/or directoriesi in the output.
        If multiple of -h -H -I --hidden --no-hidden --hidden-only are given, the last one is used.
        If none of them are given, default behavior is -H or --no-hidden.

    -t or --dot          Include . as a directory for output.
    -T or --no-dot       Do not include . as a directory for output.
        If multiple of -t -T --dot or --no-dot are given, the last one is used.
        If none of them are given, default behavior is -T or --no-dot.

    -0 or --print0       Terminate each entry with a null character (handy with xargs -0).
    -n or --newline      Terminate each entry with a newline character.
        If multiple of -0 -n --print0 or --newline are given, the last one is used.
        If none of them are given, default behavior is -n or --newline.

    -b or --base        Include the base directory for each entry (as provided with the <directory> args).
    -a or --absolute    List the absolute path to each entry.
    -B or --no-base     Do not include the base directory for each entry.
        If multiple of -b -a -B --base --absolute or --no-base are given, the last one is used.
        If none of them are given, default behavior depends on the number of directories provided as arguments.
            If zero or one are provided, -B or --no-base is the default.
            If two or more are provided, -b or --base is the default.

    [--] <directory> [<directory2>...]
        Any number of directories can be provided as a base directory to list the contents of.
        Any arguments that do not start with a - are taken to be directories.
        Additionaly, any arguments provided after -- are taken to be directories.
        So if your directory of interest starts with a dash, you must provide it after a -- argument.
        If no directories are provided, the current directory (.) is used.

Default behavior: All of these behave the same:
    list
    list --files --dirs --no-hidden --no-dot --newline --no-base
    list -fdHTnB

Examples:
    Get just the (non-hidden) files in the current directory:
        list -f
        list -D
    Get just the (non-hidden) directories in the current directory:
        list -d
        list -F
    Get just the hidden directories in the home directory:
        list -Id ~
    Get ls long-format information on the hidden files in the /users/Spicylemon directory:
        list -If0 /users/SpicyLemon | xargs -0 ls -l
    Get ls long-format information on the entire contents of the foo/ and
    bar/ directories (in the current directory), sorted by date, newest at the bottom,
    and including the directories themselves:
        list -th0 foo bar | xargs -0 ls -ldtr

Exit codes:
    0   Normal execution with output.
    1   Normal execution but there was nothing to output.
    2   Invalid argument provided.
    3   Invalid directory provided.

EOF
                return 0
                ;;
            -f|--files)    show_files='YES';;
            -F|--no-files) show_files='NO';;
            -d|--dirs)    show_dirs='YES';;
            -D|--no-dirs) show_dirs='NO';;
            -h|--hidden)      show_hidden='YES';;
            -H|--no-hidden)   show_hidden='NO';;
            -I|--hidden-only) show_hidden='ONLY';;
            -t|--dot)    show_dot='YES';;
            -T|--no-dot) show_dot='NO';;
            -0|--print0)  terminator='\0';;
            -n|--newline) terminator='\n';;
            -b|--base)     show_base='YES';;
            -a|--absolute) show_base='ABSOLUTE';;
            -B|--no-base)  show_base='NO';;
            --) no_more_flags='YES';;
            -*)
                printf 'Unknown flag: [%s]\n' "$arg" >&2
                return 2
                ;;
            *)
                directories+=( "$arg" )
                ;;
        esac
    done
    if [[ "${#directories[@]}" -gt '0' ]]; then
        for directory in "${directories[@]}"; do
            if [[ ! -d "$directory" ]]; then
                printf 'directory not found: %s\n' "$directory" >&2
                return 3
            fi
        done
    else
        directories=( '.' )
    fi
    if [[ -z "$show_base" ]]; then
        if [[ "${#directories[@]}" -le '1' ]]; then
            show_base='NO'
        else
            show_base='YES'
        fi
    fi
    if [[ -z "$show_files" && -z "$show_dirs" ]]; then
        show_files='YES'
        show_dirs='YES'
    elif [[ -z "$show_files" ]]; then
        show_files='NO'
    elif [[ -z "$show_dirs" ]]; then
        show_dirs='NO'
    fi

    local exit_code find_args grep_args base_dir cwd
    exit_code=1
    if [[ "$show_files" == 'NO' && "$show_dirs" == 'NO' ]]; then
        return $exit_code
    fi
    find_args=( -maxdepth 1 )
    if [[ "$show_files" == 'YES' && "$show_dirs" != 'YES' ]]; then
        find_args+=( -type f )
    elif [[ "$show_files" != 'YES' && "$show_dirs" == 'YES' ]]; then
        find_args+=( -type d )
    fi
    if [[ "$show_hidden" == 'ONLY' ]]; then
        if [[ "$show_dot" == 'YES' ]]; then
            # Keep everything that starts with a dot.
            grep_args=( '^\.' )
        else
            # Keep everything that starts with a dot followed by something.
            grep_args=( '^\..' )
        fi
    elif [[ "$show_hidden" == 'YES' ]]; then
        if [[ "$show_dot" == 'YES' ]]; then
            # Keep everything.
            grep_args=( '.*' )
        else
            # Get rid of the dot entry.
            grep_args=( -v '^\.$' )
        fi
    elif [[ "$show_dot" == 'YES' ]]; then
        # Get rid of everything that starts with a dot followed by something (thus keeping the dot).
        grep_args=( -v '^\..' )
    else
        # Get rid of everything that starts with a dot.
        grep_args=( -v '^\.' )
    fi
    if [[ "$show_base" == 'ABSOLUTE' ]]; then
        # If the absolute dir is requested, it's not going to change during the run, so we can set it now.
        # Escape some stuff so it can be used as a sed replacement string.
        base_dir="$( sed 's/[\/&]/\\&/g' <<< "$( pwd )/" )"
    else
        base_dir=''
    fi
    cwd="$( pwd )"
    exec 3>&1
    for directory in "${directories[@]}"; do
        if [[ "$directory" != '.' ]]; then
            cd "$directory"
        fi
        if [[ "$show_base" == 'YES' ]]; then
            # Different base for each directory, so setting it now.
            # Still need to escape some stuff so it can be used as a sed replacement string.
            base_dir="$( sed 's/[\/&]/\\&/g' <<< "$directory/" )"
        fi
        # Do everything
        output="$( find . "${find_args[@]}" | sed 's/^\.\///' | grep "${grep_args[@]}" | sed "s/^/$base_dir/" | sort | tr '\n' "$terminator" | tee >( cat - >&3 ) )"
        if [[ -n "$output" ]]; then
            exit_code=0
        fi
        if [[ "$directory" != '.' ]]; then
            cd "$cwd"
        fi
    done
    exec 3>&-
    return $exit_code
}

if [[ "$sourced" != 'YES' ]]; then
    list "$@"
    exit $?
fi
unset sourced

return 0
