#!/bin/bash
# This file contains the sdkman_fzf function which adds fzf selection functionality to the sdkman system.
# This file is meant to be sourced to add the sdkman_fzf function to your environment.
#
# To always use skdman_fzf (instead of fzf directly), set it up as an alias:
#   alias sdk='sdkman_fzf'
#
# Dependencies:
#   * sdkman SDK manager: https://sdkman.io/
#   * fzf fuzzy finder: https://github.com/junegunn/fzf#installation
#
# In order to use fzf to select an argument, provide an underscore as the argument.
# Examples:
#   Select the version(s) of java to install:
#       sdkman_fzf install java _
#   Select the candidate(s) to list:
#       sdkman_fzf list _
#   Select the version of ant to set as the default:
#       sdkman_fzf default ant _
#   Select the version of java to use:
#       sdkman_fzf use java _
#   Select a candidate and then versions you want the home directory for:
#       sdkman_fzf home _ _
#
# File contents:
#   sdkman_fzf -> enhance the sdkman system with fzf selection capabilities for arguments.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    cat >&2 << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

sdkman_fzf () {
    # The standard sdk command is a function, so it cannot be run using `command sdk`.
    # But if an sdk alias is defined prior to this function being loaded, the alias will be substituted for all the sdk calls in here.
    # That would be bad if the alias included a call to this function.
    # So, in order to allow for an sdk alias to be defined to reference this function, the actual sdk command will be put into a variable.
    # And that variable will be used to make the call, thus circumventing the alias subsitution upon function definition.
    local sdkmancmd sdk_exit_code
    sdkmancmd='sdk'
    # Make sure we even have the needed command.
    if ! command -v "$sdkmancmd" > /dev/null 2>&1; then
        printf 'Missing required command: %s\n' "$sdkmancmd" >&2
        "$sdkmancmd"
        return $?
    fi

    # Add some info about usage if just asking for help.
    if [[ "$#" -eq '0' || ( "$#" -eq '1' && "$1" =~ ^(help|-h|--help)$ ) ]]; then
        if [[ "$#" -eq '0' ]]; then
            "$sdkmancmd"
            sdk_exit_code=$?
        else
            "$sdkmancmd" help
            sdk_exit_code=$?
        fi
        cat << EOF
sdkman_fzf added functionality:
    An underscore can be used in place of an argument
      in order to use fzf to select the desired value.
    Not applicable to [local-path].
    Some commands allow for multi-select (e.g. install).

EOF
        return $sdk_exit_code
    fi
    # Parameters are generally <command> [candidate] [version].
    # A few commands have different options instead of the [candidate], but they can be selected too.
    # The only possibly selectable arguments are $1, $2, and $3. If none of them indicate selection, just pass it all straight on to sdkman
    if [[ "$1" != '_' && "$2" != '_' && "$3" != '_' ]]; then
        "$sdkmancmd" "$@"
        return $?
    fi

    # We're gonna need fzf at some point, so now we should check to make sure we've got it.
    if ! command -v fzf > /dev/null 2>&1; then
        printf 'Missing required command: fzf\n' >&2
        fzf
        return $?
    fi

    local abort_code sdk_cmd sdk_args sdkman_cannary flush_broadcast options fzf_m_flag selections candidate version_start version last_arg_list last_arg return_code
    abort_code=2
    sdk_args=()
    last_arg_list=()

    # Sometimes, the first sdk command in a while will ask a yes/no question (probably about upgrading).
    # Since I don't know a whole lot about that yet, I'm going to answer no to anything it might ask, but store the whole thing in a variable.
    # This will also trigger a broadcast if there is one (so that it's not contained in the other stuff that needs to be parsed).
    sdkman_cannary="$( yes n | "$sdkmancmd" )"

    # If there was a broadcast, we'll want to flush it later so it shows up in the actual sdk command being requested.
    if grep -q 'BROADCAST' <<< "$sdkman_cannary" > /dev/null 2>&1; then
        flush_broadcast='YES'
    fi

    # If there was a y/n question (or even a question mark), export a variable so that it's easier to figure out how to automatically handle later.
    if grep -iqE '(y/n|\?)' <<< "$sdkman_cannary" > /dev/null 2>&1; then
        export SDKMAN_CANNARY="$sdkman_cannary"
        printf 'Question detected in initial sdk call. Output exported in variable SDKMAN_CANNARY.\n' >&2
    fi

    # Figure out which command is being run
    if [[ "$#" -gt '0' ]]; then
        if [[ "$1" == '_' ]]; then
            sdk_cmd="$( printf 'install uninstall list use config default home env current upgrade version broadcast help offline selfupdate update flush' | tr ' ' '\n' | sort | fzf +m --cycle --tac )"
            if [[ -z "$sdk_cmd" ]]; then
                printf '%s %s command aborted.\n' "$sdkmancmd" "$*" >&2
                return $abort_code
            fi
        else
            sdk_cmd="$1"
        fi
        sdk_args+=( "$sdk_cmd" )
        shift
    fi

    # Handle the 1st argument for the desired sdk_cmd.
    if [[ "$#" -gt '0' ]]; then
        if [[ "$1" == '_' ]]; then
            fzf_m_flag='+m'
            options=''
            case "$sdk_cmd" in
                # These can only have one candidate selected
                # install   or i    <candidate> [version] [local-path]
                # uninstall or rm   <candidate> <version>
                # use       or u    <candidate> <version>
                # home      or h    <candidate> <version>
                # These can have multiple candidate selections
                # list      or ls   [candidate]
                # current   or c    [candidate]
                # upgrade   or ug   [candidate]
                # This can have muliple candidate selections if there's no version argument.
                # default   or d    <candidate> [version]
                i|install|rm|uninstall|u|use|h|home|ls|list|c|current|ug|upgrade|d|default)
                    options="$( "$sdkmancmd" list | grep -o 'sdk install .*$' | sed 's/^sdk install //' | sort )"
                    if [[ "$sdk_cmd" =~ ^(ls|list|c|current|ug|upgrade)$ || ( "$sdk_cmd" =~ ^(d|default)$ && "$#" -eq '1' ) ]]; then
                        fzf_m_flag="-m"
                    fi
                    ;;
                # env       or e    [init|install|clear]
                e|env)
                    options="$( printf 'init|install|clear' | tr '|' '\n' )"
                    ;;
                # offline           [enable|disable]
                offline)
                    options="$( printf 'enable|disable' | tr '|' '\n' )"
                    ;;
                # selfupdate        [force]
                selfupdate)
                    options="$( printf 'force|do not force' | tr '|' '\n' )"
                    ;;
                # flush             [archives|tmp|broadcast|version]
                flush)
                    options="$( printf 'archives|tmp|broadcast|version' | tr '|' '\n' )"
                    ;;
                # These shouldn't have any other arguments. Just add the arg to the arg list and let sdk handle it.
                # config
                # version   or v
                # broadcast or b
                # help
                # update
                config|v|version|b|broadcast|help|update)
                    sdk_args+=( "$1" )
                    ;;
                *)
                    printf 'Cannot select argument 1 for unknown %s command [%s]\n' "$sdkmancmd" "$sdk_cmd" >&2
                    return $abort_code
                    ;;
            esac
            if [[ -n "$options" ]]; then
                selections=( $( fzf $fzf_m_flag --cycle --tac <<< "$options" ) )
                if [[ "${#selections[@]}" -eq '0' ]]; then
                    printf '%s %s %s command aborted.\n' "$sdkmancmd" "$sdk_cmd" "$*" >&2
                    return $abort_code
                elif [[ "${#selections[@]}" -eq '1' ]]; then
                    # some shells are 0 based, some are 1 based. One of these will give the value, the other will be an empty string.
                    # It might not actually be the candidate, but whatever, that variable only gets used if it really is a candidate.
                    candidate="${selections[0]}${selections[1]}"
                    if [[ "$candidate" != 'do not force' ]]; then
                        sdk_args+=( "$candidate" )
                    fi
                else
                    last_arg_list=( "${selections[@]}" )
                fi
            else
                printf 'There is nothing to select.\n%s %s %s command aborted.\n' "$sdkmancmd" "$sdk_cmd" "$*" >&2
                return $abort_code
            fi
        else
            candidate="$1"
            sdk_args+=( "$1" )
        fi
        shift
    fi

    # Handle the 2nd argument for the desired sdk_cmd.
    if [[ "$#" -gt '0' ]]; then
        if [[ "$1" == '_' ]]; then
            fzf_m_flag='+m'
            options=''
            case "$sdk_cmd" in
                # These can only have one version selected
                # default   or d    <candidate> [version]
                # use       or u    <candidate> <version>
                # These can have multiple versions selected
                # install   or i    <candidate> [version] [local-path]
                # uninstall or rm   <candidate> <version>
                # home      or h    <candidate> <version>
                d|default|i|install|rm|uninstall|u|use|h|home)
                    if [[ "$candidate" == 'java' ]]; then
                        # The output of sdk list java is different from the other candidates.
                        options="$( "$sdkmancmd" list "$candidate" )"
                        version_start=$(( $( head -n 4 <<< "$options" | tail -n 1 | grep -oE '^[^|]+' | wc -c ) + 5 ))
                        options=$( tail -n +6 <<< "$options" \
                                    | tail -r | tail -n +6 | tail -r \
                                    | awk '{ split($0, a, "|");
                                             if (a[1] !~ /^[[:space:]]*$/) { vendor = a[1]; } else { a[1] = vendor; };
                                             if (a[2] !~ /^[[:space:]]*$/) { a[2] = ">"; } else { a[2] = " "; };
                                             if (a[5] ~ /installed/) { a[5] = "*"; } else if (a[5] !~ /^[[:space:]]*$/) { a[5] = "+"; } else { a[5] = " "; };
                                             sub(/[[:space:]]*$/,"",a[6]);
                                             print a[1]" "a[2]" "a[5]" "a[6]; }' \
                                    | sort -V -t '\n' -k "1.$version_start" )
                        # Unless we're installing, we only care about the ones already installed.
                        if [[ ! "$sdk_cmd" =~ ^(i|install)$ ]]; then
                            options="$( grep -E "^.{$(( version_start - 3 ))}[+*]" <<< "$options" )"
                        fi
                    else
                        options="$( "$sdkmancmd" list "$candidate" | tail -n +4 | tail -r | tail -n +6 | grep -v '^[[:space:]]*$' | grep -oE '[> ] [+* ] [^[:space:]]+' | sort -V -t '\n' -k 1.5 )"
                        # Unless we're installing, we only care about the ones already installed.
                        if [[ ! "$sdk_cmd" =~ ^(i|install)$ ]]; then
                            options="$( grep '^..[+*]' <<< "$options" )"
                        fi
                    fi
                    if [[ -z "$options" ]]; then
                        printf 'There are no versions to select.\n' >&2
                    fi
                    if [[ "$sdk_cmd" =~ ^(i|install|rm|uninstall|h|home)$ ]]; then
                        fzf_m_flag='-m'
                    fi
                    ;;
                # These shouldn't have any other arguments. Just add the arg to the arg list and let sdk handle it.
                # list      or ls   [candidate]
                # current   or c    [candidate]
                # upgrade   or ug   [candidate]
                # env       or e    [init|install|clear]
                # offline           [enable|disable]
                # selfupdate        [force]
                # flush             [archives|tmp|broadcast|version]
                # config
                # version   or v
                # broadcast or b
                # help
                # update
                ls|list|c|current|ug|upgrade|e|env|offline|selfupdate|flush|config|v|version|b|broadcast|help|update)
                    sdk_args+=( "$1" )
                    ;;
                *)
                    printf 'Cannot select argument 2 for unknown %s command [%s]\n' "$sdkmancmd" "$sdk_cmd" >&2
                    return $abort_code
                    ;;
            esac
            if [[ -n "$options" ]]; then
                selections=( $( fzf $fzf_m_flag --cycle --tac <<< "$options" | grep -oE '[^[:space:]]+$' ) )
                if [[ "${#selections[@]}" -eq '0' ]]; then
                    printf '%s %s %s %s command aborted.\n' "$sdkmancmd" "$sdk_cmd" "$candidate" "$*" >&2
                    return $abort_code
                elif [[ "${#selections[@]}" -eq '1' ]]; then
                    # some shells are 0 based, some are 1 based. One of these will give the value, the other will be an empty string.
                    version="${selections[0]}${selections[1]}"
                    sdk_args+=( "$version" )
                else
                    last_arg_list=( "${selections[@]}" )
                fi
            else
                printf '%s %s %s %s command aborted.\n' "$sdkmancmd" "$sdk_cmd" "$candidate" "$*" >&2
                return $abort_code
            fi
        else
            version="$1"
            sdk_args+=( "$1" )
        fi
        shift
    fi

    # If there was a broadcast in the initial output, flush the broadcast now so that it'll show up again to the user.
    if [[ -n "$flush_broadcast" ]]; then
        "$sdkmancmd" flush broadcast > /dev/null 2>&1
    fi
    # Run whatever needs to be run!
    if [[ "${#last_arg_list[@]}" -gt '0' ]]; then
        return_code=0
        for last_arg in "${last_arg_list[@]}"; do
            printf '\033[1m%s %s %s %s\033[0m\n' "$sdkmancmd" "${sdk_args[*]}" "$last_arg" "$*"
            "$sdkmancmd" "${sdk_args[@]}" "$last_arg" "$@"
            sdk_exit_code=$?
            if [[ "$sdk_exit_code" -ne '0' ]]; then
                return_code=$sdk_exit_code
            fi
        done
    else
        printf '\033[1m%s %s %s\033[0m\n' "$sdkmancmd" "${sdk_args[*]}" "$*"
        "$sdkmancmd" "${sdk_args[@]}" "$@"
        return_code=$?
    fi
    return $return_code
}

return 0
