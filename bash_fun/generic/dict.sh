#!/bin/bash
# This file contains the dict function that greps the dict file.
# This file can be sourced to add the dict function to your environment.
# This file can also be executed to run the dict function without adding it to your environment.
#
# File contents:
#   dict  -------> Grep the dict file.
#   DICT_FILE  --> Environment variable containing the path to the dict file.
#   dict_file  --> Print the $DICT_FILE.
#   dict_dir  ---> Print the directory containing the $DICT_FILE.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# If DICT_FILE isn't set yet, set it to /usr/share/dict/words.
DICT_FILE=${DICT_FILE:-/usr/share/dict/words}

# Usage: dict <grep options>
dict () {
    grep "$@" "$DICT_FILE"
}

# Usage: dict_file
dict_file () {
    printf '%s\n' "$DICT_FILE"
}

# Usage: dict_dir
dict_dir () {
    dirname "$DICT_FILE"
}

if [[ "$sourced" != 'YES' ]]; then
    dict "$@"
    exit $?
fi
unset sourced

return 0
