#!/bin/bash
# This file contains the go_use.sh function that switches the go bin symlink to a provided version.
# This file can be sourced to add the go_use function to your environment.
# This file can also be executed to run the go_use function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# To add a new version, add/update stuff above the ADD_VERSION comments: [1] [2] [3] [4].

go_use () {
    local list verbose listing desired which_go cur_link desired_link rv ln_flags
    # To add new versions, just add a new entry to this list.
    # Required line format: <one or more spaces><version>:<one or more spaces><link path>
    # When invoking this function to make a switch, a user would supply just the <version> as an arg.
    # If the <link path> is relative, it's relative to the result of `which go` (which for me is /opt/homebrew/bin/go).
    # The full version of 1.20 is 1.20.1. The others have the patch versions in their paths.
    list="$( cat << EOF
  1.18: ../Cellar/go@1.18/1.18.10/bin/go
  1.20: /usr/local/go/bin/go
  1.21: ../Cellar/go/1.21.4/bin/go
  1.23: ../Cellar/go/1.23.3/bin/go
EOF

)"

    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help|help)
                printf 'Usage: go_use {%slist} [-v|--verbose]\n' "$( sed -E 's/^[[:space:]]+//; s/:.*$//' <<< "$list" | tr '\n' '|' )"
                return 0
                ;;
            -v|--verbose)
                verbose=1
                ;;
            -l|--list|l|list)
                listing=1
                ;;
            *)
                if [[ -n "$desired" ]]; then
                    printf 'Unknown argument: %q\n' "$1"
                    return 1
                fi
                desired="$1"
                ;;
        esac
        shift
    done

    rv=0

    if [[ -z "$desired" ]]; then
        listing=1
    else
        # If desired starts with a v, remove it now.
        if [[ "$desired" =~ ^v ]]; then
            desired="${desired:1}"
        fi

        [[ "$verbose" ]] && printf 'Identifying desired link for %q: ' "$desired"
        desired_link="$( grep -F " $desired:" <<< "$list" | sed -E 's/^[^:]*:[[:space:]]+//' )"
        [[ "$verbose" ]] && printf '%q\n' "$desired_link"

        if [[ -z "$desired_link" ]]; then
            printf 'Unknown version: %q\n' "$desired" >&2
            listing=1
            rv=1
        fi
    fi

    [[ "$verbose" ]] && printf 'which go: '
    which_go="$( which go )"
    [[ "$verbose" ]] && printf '%q\n' "$which_go"
    if [[ ! -L "$which_go" ]]; then
        printf 'Not a symlink: ' >&2
        ls -al "$which_go" >&2
        return 1
    fi

    [[ "$verbose" ]] && printf 'readlink %q: ' "$which_go"
    cur_link="$( readlink "$which_go" )"
    [[ "$verbose" ]] && printf '%q\n' "$cur_link"


    if [[ "$listing" ]]; then
        printf 'Available Versions:\n'
        awk -v cur_link="$cur_link" -v which_go="$which_go" '{
            if (index($0,cur_link) > 0) {
                cur=$0;
                sub(/^[[:space:]]+/,"",cur);
                sub(/:.*$/,"",cur);
                i=index($0,cur);
                print substr($0,1,i-1) "\033[1m" cur "\033[0m" substr($0,i+length(cur)) "  \033[1m<- " which_go "\033[0m";
                found="1";
            } else {
                print;
            };
        }
        END {
            if (found=="") {
                print "  \033[1mActual\033[0m: " cur_link "  \033[1m<- " which_go "\033[0m";
            };
        }' <<< "$list"
    elif [[ "$cur_link" == "$desired_link" ]]; then
        printf 'Already: '
        ls -al "$which_go"
    else
        printf '    Was: '
        ls -al "$which_go"
        ln_flags='-sf'
        [[ "$verbose" ]] && ln_flags="${ln_flags}v"
        [[ "$verbose" ]] && printf 'ln %s %q %q\n' "$ln_flags" "$desired_link" "$which_go"
        ln $ln_flags "$desired_link" "$which_go" || rv=$?
        printf ' Is Now: '
        ls -al "$which_go"
    fi

    printf 'Current: ' && go version
    return $rv
}

if [[ "$sourced" != 'YES' ]]; then
    go_use "$@"
    exit $?
fi
unset sourced

return 0
