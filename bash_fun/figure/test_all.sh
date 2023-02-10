#!/bin/bash
# This file contains the test_all function that runs the tests and sims in provenance and cosmos-sdk repos.
# This file can be sourced to add the test_all function to your environment.
# This file can also be executed to run the test_all function without adding it to your environment.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

test_all () {
    local can_hr11 can_show_last_exit_code can_say
    command -v hr11 > /dev/null 2>&1 && can_hr11='YES'
    command -v show_last_exit_code > /dev/null 2>&1 && can_show_last_exit_code='YES'
    command -v say > /dev/null 2>&1 && can_say='YES'

    local fail_fast sound targets skips added_targets
    fail_fast='YES'
    sound='on'
    targets=( test test-sim-nondeterminism test-sim-import-export test-sim-after-import test-sim-multi-seed-short )
    skips=()
    added_targets=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            -h|--help|help)
                cat << EOF
Usage: test_all [[--skip|-s] <targets>] [[--also|-a] <targets>] [[--targets|-t] <targets>]
                [--continue|-c|--break-b] [--sound [on|off|beep|say]|--noisy|--quiet|--beep|--say]

By default, the following make targets are run:
  ${targets[@]}

Testing stops at the first failure.
To continue on failures, provide the --continue or -c flag.
To break on failure (default), provide the --break or -b flag.
If multiple --continue, -c, --break, or -b flags are provided, the last one is used.

This list can be overwritten using the --targets or -t option.
To overwrite the list with multiple other targets, provide them as args after a single --targets or -t flag.
If multiple --target or -t flags are provided, the last set is used.

To skip targets, use the --skip or -s option.
Skipped targets are noted in the output as being skipped.
If multiple --skip or -s options are provided, they are combined.

To add targets, use the --also or -a option.
Added targets are run in the order provided after the main set of targets.
If multiple --also or -a options are provided, they are combined.

By default, when a test fails, noise is made. Noise is also made once everything completes.
This can be controlled using the --sound option.
    --sound on    - (default) Use normal sound behavior.
    --sound off   - Do not make any sound.
    --sound beep  - Use bell characters for sound even if the say command is available.
    --sound say   - Use the say command to make noise.
    --noisy       - Alias for --sound on
    --quiet       - Alias for --sound off
    --beep        - Alias for --sound beep
    --say         - Alias for --sound say
If multiple --sound, --quiet, --beep, or --say options are given, the last one is used.
Proving --sound without specififying an option is the same as providing --sound on.


EOF
                return 0
                ;;
            --continue|-c)
                fail_fast=''
                ;;
            --break|-b)
                fail_fast='YES'
                ;;
            --targets|-t|--tests|--test)
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    printf 'At least one target is required after the %s flag.\n' "$1" >&2
                    return 1
                fi
                targets=()
                while [[ -n "$2" && ! "$2" =~ ^- ]]; do
                    targets+=( "$2" )
                    shift
                done
                ;;
            --skip|-s|--skips)
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    printf 'At least one target is required after the %s flag.\n' "$1" >&2
                    return 1
                fi
                while [[ -n "$2" && ! "$2" =~ ^- ]]; do
                    skips+=( "$2" )
                    shift
                done
                ;;
            --also|-a|--alsos|--add|adds)
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    printf 'At least one target is required after the %s flag.\n' "$1" >&2
                    return 1
                fi
                while [[ -n "$2" && ! "$2" =~ ^- ]]; do
                    added_targets+=( "$2" )
                    shift
                done
                ;;
            --sound)
                if [[ -z "$2" || "$2" =~ ^- ]]; then
                    sound='on'
                else
                    if [[ ! "$2" =~ ^(on|off|beep|say)$ ]]; then
                        printf 'Unknown argument after the %s flag: [%s].\n' "$1" "$2" >&2
                        return 1
                    fi
                    sound="$2"
                    shift
                fi
                ;;
            --noisy) sound='on' ;;
            --quiet) sound='off' ;;
            --beep) sound='beep' ;;
            --say) sound='say' ;;
            *)
                printf 'Unknown argument: [%s].\n' "$1" >&2
                return 1
                ;;
        esac
        shift
    done

    case "$sound" in
        off) sound='';;
        beep) can_say='';;
        say)
            if [[ -z "$can_say" ]]; then
                printf 'The say command is not available.\n' >&2
                return 1
            fi
            ;;
    esac

    if [[ "${#added_targets[@]}" -gt '0' ]]; then
        targets+=( "${added_targets[@]}" )
    fi

    local ec t st not_first tec count
    ec=0
    for t in "${targets[@]}"; do
        # Output a header
        if [[ -n "$not_first" ]]; then
            printf '\n\n\n\n'
        else
            not_first='YES'
        fi
        if [[ -n "$can_hr11" ]]; then
            hr11 "$t"
        else
            printf '\033[1m#\n# %s\n########################################\033[0m\n' "$t"
        fi

        # Skip the target if we're supposed to.
        if [[ "${#skips[@]}" -gt '0' ]]; then
            for st in "${skips[@]}"; do
                if [[ "$t" == "$st" ]]; then
                    printf '\nSkipped\n\n'
                    continue 2
                fi
            done
        fi

        # Make the target
        make "$t"
        tec="$?"

        # Output the target's exit code.
        printf '\nFinished: make %s\n' "$t"
        printf 'Exit Code:' "$t"
        if [[ -n "$can_show_last_exit_code" ]]; then
            ( exit "$tec"; )
            show_last_exit_code
        elif [[ "$tec" -eq '0' ]]; then
            # 97 = bright white, 42 = green background
            printf '\033[97;42m %3d \033[0m' "$tec"
        else
            # 97 = bright white, 41 = red background
            printf '\033[97;41m %3d \033[0m' "$tec"
        fi
        printf '\n'

        # If it failed, update the final error code and make some noise.
        if [[ "$tec" -ne '0' ]]; then
            ec="$tec"
            if [[ -n "$sound" ]]; then
                if [[ -n "$can_say" ]]; then
                    say "make $t failed with exit code $tec"
                else
                    # Beep 4 times very fast.
                    count=4
                    while :; do
                        printf '\a'
                        count="$(( count - 1 ))"
                        [[ "$count" -le '0' ]] && break
                        sleep .1
                    done
                fi
            fi
            [[ -n "$fail_fast" ]] && break
        fi
    done

    # Make some noise about being done.
    if [[ -n "$sound" ]]; then
        if [[ -n "$can_say" ]]; then
            if [[ "$ec" -eq '0' ]]; then
                say 'All tests completed successfully.'
            else
                say 'Done running tests. There were failures.'
            fi
        else
            if [[ "$ec" -eq '0' ]]; then
                # Two beeps, nicely spaced.
                printf '\a'
                sleep .3
                printf '\a'
            else
                # Four beeps, nicely spaced.
                count=4
                while :; do
                    printf '\a'
                    count="$(( count - 1 ))"
                    [[ "$count" -le '0' ]] && break
                    sleep .3
                done
            fi
        fi
    fi

    return "$ec"
}

if [[ "$sourced" != 'YES' ]]; then
    test_all "$@"
    exit $?
fi
unset sourced

return 0
