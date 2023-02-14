#!/bin/bash
# This file contains the bash_prompt_newline function that prints a newline if the last thing didn't.
# This file is meant to be sourced to add the bash_prompt_newline function to your environment.
# This function is only designed for bash and won't work in ZSH (or any shell with 1-based arrays).
#
# Usage: bash_prompt_newline <indicator> [<command> [<arg1> [<arg2> ...]]]
#   If the previous line did not end in a newline, the <indicator> is printed followed by a newline.
#   If a <command> is given, it is always executed (regardless of whether the newline was printed).
#   If no arguments are provided, the <indicator> will be a bright yellow percent sign on
#       a dark-gray background: '\033[100;93m%\033[0m'
#   In order to provide a <command>, an <indicator> argument must be provided.
#   You can provide the magic argument "default-indicator" for the <indicator> to use the default for it.
#
# To use this, either include it at the start of your PS1 or call it as your PROMPT_COMMAND.
# In your PS1:
#   Use this method if you currently define your prompt by setting PS1.
#       export PS1='$( bash_prompt_newline "\033[97m%\033[0m" )<previous PS1>'
#   The single quotes are important. If the $( bash_prompt_newline ) is in double quotes,
#   It will be called and converted to a string (probably an empty string) as the variable is
#   being set and exported. By placing it in single quotes, it isn't called until the prompt
#   is being printed, which is what's needed here.
#   You could add it to your existing PS1 using the following line:
#       export PS1='$( bash_prompt_newline "default-indicator" )'"$PS1"
# With PROMPT_COMMAND:
#   Use this method if you currently define your prompt by setting PROMPT_COMMAND.
#       export PROMPT_COMMAND="bash_prompt_newline '\033[97m%\033[0m' '<command>' ['<arg1>' ['<arg2>'...]]"
#   Note: The type of quotes use here isn't as important, but it's important to still have them.
#   I use double quotes on the outside here because, in my use, <arg1> and <arg2> are other variables
#   that I want interpreted as I set define PROMPT_COMMAND. I each thing in single quotes because they
#   sometimes have spaces in them, which I do NOT want interpreted as a break between arguments.
#
# File contents:
#   bash_prompt_newline  --> Prints a newline if the last thing didn't.
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
    unset sourced
    exit 1
fi
unset sourced

bash_prompt_newline () {
    # Preserve the last exit code.
    local exit_code=$?
    local ind cmd curpos
    if [[ "$#" -eq '0' || "$1" == 'default-indicator' ]]; then
        # 100 = dark-gray background, 97 = bright yellow text
        ind='\033[100;93m%\033[0m'
    else
        ind="$1"
    fi
    shift
    cmd=( "$@" )

    # From https://stackoverflow.com/questions/19943482/configure-shell-to-always-print-prompt-on-new-line-like-zsh
    # CSI 6n reports the cursor position as ESC[n;mR, where n is the row
    # and m is the column. Issue this control sequence and silently read
    # the resulting report until reaching the "R". By setting IFS to ";"
    # in conjunction with read's -a flag, fields are placed in an array.
    printf '\033[6n'
    IFS=';' read -s -d R -a curpos
    # Since we don't care about the row, this isn't needed.
    # It's kept here because it's interesting and I might care about the row sometime later.
    #curpos[0]="${curpos[0]:2}"  # strip leading ESC[
    # If not at column 1, print the indicator followed by a newline.
    (( curpos[1] > 1 )) && printf '%b\n' "$ind"

    # If a cmd was provided, execute it now.
    [[ "${#cmd[@]}" -gt '0' ]] && "${cmd[@]}"

    return $exit_code
}

return 0
