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
    local old_word new_word files file
    old_word="$1"
    new_word="$2"
    shift
    shift
    files="$@"
    if [[ -z "$old_word" || -z "$new_word" || "${#files[@]}" -eq '0' ]]; then
        echo "Usage: change_word <old word> <new word> <files>"
        return 0
    fi
    if [[ "$old_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        echo "Only letters, numbers, dashes, underscores and spaces are allowed in the provided words." >&2
        return 1
    fi
    if [[ "$new_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        echo "Only letters, numbers, dashes, underscores and spaces are allowed in the provided words." >&2
        return 1
    fi
    echo -e "Changing \033[0m \033[1;31m$old_word\033[0m to \033[1;32m$new_word\033[0m"
    echo -e "In: ${files[@]}"
    echo ''
    echo -e "\033[1;37;41m Before: \033[0m"
    grep -E "\<($old_word|$new_word)\>" "${files[@]}" \
        | GREP_COLOR='1;31' grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR='1;32' grep --color=always "\<$new_word\>\|$"
    echo ''
    for file in ${files[@]}; do
        sed -i '' "s/[[:<:]]$old_word[[:>:]]/$new_word/g;" "$file"
    done
    echo -e "\033[1;37;42m After: \033[0m"
    grep -E "\<($old_word|$new_word)\>" "${files[@]}" \
        | GREP_COLOR='1;31' grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR='1;32' grep --color=always "\<$new_word\>\|$"
    echo ''
}

if [[ "$sourced" != 'YES' ]]; then
    change_word "$@"
    exit $?
fi
unset sourced

return 0
