#!/bin/bash
# This file can be copied from one year to the next without needing alterations.
#

print_usage () {
    cat << EOF
github-urls.sh - Generate URLs to files and/or dirs in GitHub.

Usage: github-urls.sh [--cur] [--count|-n <num>] [<dir>|<file> [...]]

  If no args are provided, it is the same as providing just --cur.

  --cur prints links to all parts of the most recent day.

  --count <num> or -n <num> prints links to the <num> most recent solutions.
    E.g. if the most recent solutions are day-03a and day-03b, and -n 1 is provided,
        this will print the link for just day-03b.

  <dir> is a directory for a day, e.g. 'day-01a'.
      If the file '<dir>/<dir>.go' exists, the link will go to that file.
      Otherwise, it will go to the whole directory.
  <file> is a specific file, e.g. 'day-01a/day-01a.go'.
  Multiple <dir>s and <file>s can be provided to get the links for all of them.

EOF

}

entries=()

add_cur () {
    local last
    last="$( ls | grep '^day' | sort -r | head -n "1" )"
    [[ -n "$VERBOSE" ]] && printf 'Most recent: "%s"\n' "$last"
    if [[ -z "$last" ]]; then
        printf 'No current day directories found.\n'
        exit 1
    fi
    # Remove the last char from the most recent dir and get all dirs that start with what's left.
    entries+=( $( ls | grep "^${last::${#last}-1}" | sort ) )
    [[ -n "$VERBOSE" ]] && printf 'Entries after adding current (%d): %s\n' "${#entries[@]}" "${entries[*]}"
}

add_count () {
    entries+=( $( ls | grep '^day' | sort | tail -n "$1" ) )
    [[ -n "$VERBOSE" ]] && printf 'Entries after adding %d (%d): %s\n' "$1" "${#entries[@]}" "${entries[*]}"
}

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        --help|-h)
            print_usage
            exit 0
            ;;
        --verbose|-v)
            VERBOSE="$1"
            ;;
        --cur)
            add_cur
            ;;
        --count|-n)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after [%s].\n' "$1"
                exit 1
            fi
            if [[ "$2" =~ [^[:digit:]] ]]; then
                printf 'Invalid count "%s": can only contain digits.\n' "$2"
                exit 1
            fi
            add_count "$2"
            shift
            ;;
        --branch)
            if [[ -z "$2" ]]; then
                printf 'No argument provided after [%s].\n' "$1"
                exit 1
            fi
            branch="$2"
            shift
            [[ -n "$VERBOSE" ]] && printf 'Branch provided: [%s].\n' "$branch"
            ;;
        *)
            entries+=( "$1" )
            ;;
    esac
    shift
done

if [[ "${#entries[@]}" -eq '0' ]]; then
    add_cur
fi
if [[ "${#entries[@]}" -eq '0' ]]; then
    printf 'Nothing found.\n'
    exit 0
fi

repo_root="$( git rev-parse --show-toplevel )"
[[ -n "$VERBOSE" ]] && printf 'Repository root dir: [%s].\n' "$repo_root"
if [[ -z "$branch" ]]; then
    branch="$( git branch --show-current )"
    [[ -n "$VERBOSE" ]] && printf 'Using current branch: [%s].\n' "$branch"
fi
url_base='https://github.com/SpicyLemon/SpicyLemon'
[[ -n "$VERBOSE" ]] && printf 'URL base: [%s].\n' "$url_base"
[[ -n "$VERBOSE" ]] && printf 'Entries (%d): %s\n' "${#entries[@]}" "${entries[*]}"

i=0
c="${#entries[@]}"
for entry in "${entries[@]}"; do
    i=$(( i + 1 ))
    [[ -n "$VERBOSE" ]] && printf '  [%d/%d]: entry = "%s"\n' "$i" "$c" "$entry"
    if [[ -d "$entry" && -f "$entry/$entry.go" ]]; then
        entry="$entry/$entry.go"
        [[ -n "$VERBOSE" ]] && printf '  [%d/%d]: File found in dir. Now, entry = "%s"\n' "$i" "$c" "$entry"
    fi
    rel="$( realpath --relative-to="$repo_root" "$entry" )"
    [[ -n "$VERBOSE" ]] && printf '  [%d/%d]: relative path = "%s"\n' "$i" "$c" "$rel"
    t='blob'
    if [[ -d "$entry" ]]; then
        t='tree'
    fi
    [[ -n "$VERBOSE" ]] && printf '  [%d/%d]: link type = "%s"\n' "$i" "$c" "$t"
    printf '%s/%s/%s/%s\n' "$url_base" "$t" "$branch" "$rel"
done
