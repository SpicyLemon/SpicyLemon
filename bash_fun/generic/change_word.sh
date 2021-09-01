#!/bin/bash
# This file contains the change_word function that helps change a word (or phrase) from one thing to another in one or more files.
# This file can be sourced to add the change_word function to your environment.
# This file can also be executed to run the change_word function without adding it to your environment.
#
# File contents:
#   change_word  --> Changes a word (or phrase) from one thing to another in one or more files.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: change_word <old word> <new word> <files>
change_word () {
    local old_word new_word files old_color new_color before_color after_color file
    old_word="$1"
    new_word="$2"
    shift
    shift
    files=( "$@" )
    old_color="${OLD_COLOR:-31}"    # Default to red text
    new_color="${NEW_COLOR:-32}"    # Default to green text
    if [[ -z "$old_color" || "$old_color" =~ [^[:digit:]] || "$old_color" -lt '30' || "$old_color" -gt '37' ]]; then
        printf 'Illegal OLD_COLOR: [%s]. It must be a number between 30 and 37 (inclusive).\n' >&2
    fi
    if [[ -z "$new_color" || "$new_color" =~ [^[:digit:]] || "$new_color" -lt '30' || "$new_color" -gt '37' ]]; then
        printf 'Illegal NEW_COLOR: [%s]. It must be a number between 30 and 37 (inclusive).\n' >&2
    fi
    # The headings will be bold with grey text, and a backround matching the provided colors.
    # The background color escape codes for each color are 10 more than their text color escape codes.
    # E.g. 41 for a red background and 42 for a green background.
    printf -v before_color '1;37;%d' "$(( old_color + 10 ))"
    printf -v after_color '1;37;%d' "$(( new_color + 10 ))"
    # Now, make the text colors bold
    old_color="1;$old_color"
    new_color="1;$new_color"
    if [[ -z "$old_word" || -z "$new_word" || "${#files[@]}" -eq '0' ]]; then
        printf 'Usage: change_word <old word> <new word> <files>\n'
        return 0
    fi
    if [[ "$old_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        printf 'Only letters, numbers, dashes, underscores and spaces are allowed in the provided words.\n' >&2
        return 1
    fi
    if [[ "$new_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        printf 'Only letters, numbers, dashes, underscores and spaces are allowed in the provided words.\n' >&2
        return 1
    fi
    printf 'Changing \033[%sm%s\033[0m to \033[%sm%s\033[0m\n' "$old_color" "$old_word" "$new_color" "$new_word"
    printf 'In: %s\n' "${files[*]}"
    printf '\n'
    printf '\033[%sm Before: \033[0m\n' "$before_color"
    grep -E "\<($old_word|$new_word)\>" "${files[@]}" \
        | GREP_COLOR="$old_color" grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR="$new_color" grep --color=always "\<$new_word\>\|$"
    printf '\n'
    for file in "${files[@]}"; do
        sed -i '' "s/[[:<:]]$old_word[[:>:]]/$new_word/g;" "$file"
    done
    printf '\033[%sm After: \033[0m\n' "$after_color"
    grep -E "\<($old_word|$new_word)\>" "${files[@]}" \
        | GREP_COLOR="$old_color" grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR="$new_color" grep --color=always "\<$new_word\>\|$"
    printf '\n'
}

if [[ "$sourced" != 'YES' ]]; then
    change_word "$@"
    exit $?
fi
unset sourced

return 0
