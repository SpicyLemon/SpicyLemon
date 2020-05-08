#!/bin/bash
# This file contains the multi_line_replace function that replaces parts of a file with multi-line text.
# This file can be sourced to add the multi_line_replace function to your environment.
# This file can also be executed to run the multi_line_replace function without adding it to your environment.
#
# File contents:
#   multi_line_replace  --> Replaces part of a file with multi-line replacement text.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Similar to sed 's/str_to_replace/replacement_text/' filename
# Except, each line that has the str_to_replace is replicated for each line in the multi-line replacement text.
# Usage: multi_line_replace <filename> <str_to_replace> <multi-line replacement text>
multi_line_replace () {
    if [[ "$#" -ne '3' ]]; then
        echo 'Usage: multi_line_replace <filename> <str_to_replace> <multi-line replacement text>' >&2
        return 1
    fi
    local filename to_replace replace_with loop_counter loop_max line_to_replace replacement_lines
    filename="$1"
    to_replace="$2"
    replace_with="$3"
    if [[ "$filename" != '-' && ! -f "$filename" ]]; then
        printf 'File not found: [%s].\n' "$filename" >&2
        return 2
    fi
    cat "$filename" | while IFS= read -r line; do
        if [[ "$line" =~ $to_replace ]]; then
            while read repl_line; do
                sed "s/$to_replace/$repl_line/" <<< "$line"
            done <<< "$replace_with"
        else
            printf '%s\n' "$line"
        fi
    done
}

if [[ "$sourced" != 'YES' ]]; then
    multi_line_replace "$@"
    exit $?
fi
unset sourced

return 0
