#!/bin/bash
# This file primarily contains the query_prov_using_next_key function for getting multiple paginated results of provenance queries.
# This file should be sourced to make the functions available in your environment.
#
# Primary Functions of Interest:
#   query_prov_using_next_key  -- Gets multiple pages of a paginated provenance query.
#
# Other Functions:
#   prov_node  -- Get the --node value for either mainnet or testnet based on the USE_PROD env var.
#
# Exported Environment Variables:
#   NODE_TESTNET  -- A --node value for testnet.
#   NODE_MAINNET  -- A --node value for mainnet.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

NODE_TESTNET='https://rpc.test.provenance.io:443'
NODE_MAINNET='https://rpc.provenance.io:443'

if [[ "$sourced" == 'YES' ]]; then
    export NODE_TESTNET
    export NODE_MAINNET
fi

# Usage: prov_node
# If the USE_PROD environment variable isn't set, or is set to one of 'f', 'false', 'n', 'no' (ignoring case),
# then this prints $NODE_TESTNET. Otherwise, this prints NODE_MAINNET.
prov_node() {
    if [[ -z "$USE_PROD" || "$USE_PROD" =~ ^([fF]([aA][lL][sS][eE])?|[nN]([oO])?)$ ]]; then
        printf '%s' "$NODE_TESTNET"
    else
        printf '%s' "$NODE_MAINNET"
    fi
    return 0
}

query_prov_using_next_key() {
    # Different help handling from the usual.
    # Only give help for this function if no args are given or only 1 arg is given and it's -h, --help, or help.
    # This way, it's easier to get help about the query being run instead of this wrapper.
    if [[ "$#" -eq '0' || ( "$#" -eq '1' && ( "$1" == '-h' || "$1" == '--help' || "$1" == 'help' ) ) ]]; then
        # First, show the provenanced query help
        provenanced q --help
        # Then add in info about this wrapper.
        cat << EOF
query_prov_using_next_key: Runs a paginated query on provenance multiple times using the pagination next_key.

Usage: query_prov_using_next_key {qp args} {query args}

  {qp args} are arguments specific to query_prov_using_next_key:
      --qp-start {index}     Indicates the index number to start the counting at (for the filenames).
                             Default is 1.
      --qp-max-reqs {count}  Indicates the maximum number of pages to retrieve.
                             Default is 1000 - {start index}
      --qp-no-node           Flag indicating you do not want the --node query arg provided automatically.
                             Note: If the {query args} contain a --node then this -qp-no-node flag is meaningless.
      --qp-fn-base {string}  The beginning of the result filenames.
                             Default is based on the {query args}. See Below for more info.
      --qp-fn-ext {string}   The ending of th eresult filenames.
                             Default is '.json'.
                             If -o yaml or --output yaml is provided, the default is '.yaml' instead.

  {query args} are arguments for the   provenanced q   command.
      For the most part, they are directly provided to the command each time.
      There are a few special cases, though:
        --o or -output   If not provided in {query args}, --output json is added.
        --node           If not provided in {query args}, and --qp-no-node is not provided, then
                         a default is provided.
                         If the the USE_PROD environment variable is not set, or is set to 'n', 'no', 'f', or 'false',
                         then the default --node is '$NODE_TESTNET'. Otherwise the default is '$NODE_MAINNET'.
        --page-key       If provided in {query args} it is only used for the first query. All subsquent queries
                         will use the next key from the previous query.

  The {qp args} and {query args} can be intertwined. E.g. you can provide --qp-start {index} as the last arguments.

Output is stored to files in your current directory. Each filename has the format '{base}{index}{ext}'.
    The {base} is either the provided --qp-fn-base value or else a default based on {query args}.
    The {index} is a 4 digit number (i.e. the first page will be '0001').
    The {ext} is either the provided --qp-fn-ext or else a default based on the --output.

The default {base} comes from the {query args}. The intent is to make it reflect the query being run.
All {query arg} entries up to (but not including) the first entry that starts with a dash is taken to be the query being run.
The entries of the query being run are joined using dashes and a dash is added to the end of it to make the default {base}.
If the first or second query arg starts with a dash, then a --qp-fn-base must be provided.

Example:
  query_prov_using_next_key --qp-start 21 --qp-max-reqs 5 md scopes all --page-key 'AhjMOuIYQEm8GQYRpYkgcg=='
      Runs the   provenanced q md scopes all   query up to 5 times starting with the provided page key.
      The first result will be stored in md-scopes-all-0021.json and if there are enough results for 5 pages,
      the last result will be stored in md-scopes-all-0025.json.

EOF
        return 0
    fi
    local start_i max_reqs have_node fn_base fn_ext query qpargs qargs output next_key q last_i i fn cmd cmd_exit
    query=()
    qpargs=()
    qargs=()
    while [[ "$#" -gt '0' ]]; do
        case "$1" in
            --qp-start)
                if [[ -z "$2" || "$2" =~ [^[:digit:]] ]]; then
                    printf 'Invalid %s: [%s].\n' "$1" "$2" >&2
                    return 1
                fi
                start_i="$2"
                shift
                ;;
            --qp-max-reqs)
                if [[ -z "$2" || "$2" =~ [^[:digit:]] ]]; then
                    printf 'Invalid %s: [%s].\n' "$1" "$2" >&2
                    return 1
                fi
                max_reqs="$2"
                shift
                ;;
            --qp-no-node)
                have_node='YES'
                ;;
            --qp-fn-base)
                if [[ "$#" -lt '2' ]]; then
                    printf 'No value provided for %s.\n' "$1" >&2
                    return 1
                fi
                fn_base="$2"
                shift
                ;;
            --qp-fn-ext)
                if [[ "$#" -lt '2' ]]; then
                    printf 'No value provided for %s.\n' "$1" >&2
                    return 1
                fi
                fn_ext="$2"
                shift
                ;;
            --page-key|--next-key)
                if [[ -z "$2" ]]; then
                    printf 'No value provided for %s.\n' "$1" >&2
                    return 1
                fi
                next_key="$2"
                shift
                ;;
            --node)
                have_node='YES'
                qargs+=( "$1" )
                ;;
            -o|--output)
                output="$2"
                qargs+=( "$1" )
                ;;
            *)
                if [[ "${#qargs[@]}" -gt '0' || "$1" =~ ^- ]]; then
                    qargs+=( "$1" )
                else
                    query+=( "$1" )
                fi
                ;;
        esac
        shift
    done
    if [[ -z "$fn_base" ]]; then
        # The provenanced q command isn't runnable itself so zero query args is invalid.
        # And each sub-command either requires another arg or also isn't runnable (but has more sub-commands).
        # This limit of at least 2 then comes from: one arg for the sub-command and one for another sub-command or an arg.
        if [[ "${#query[@]}" -le '2' ]]; then
            printf 'Unable to define base filename. Either reorder the {query args} to put flags/options last or provide a --qp-fn-base.\n' >&2
            return 1
        fi
        for q in "${query[@]}"; do
            fn_base="${fn_base}${q}-"
        done
    fi
    if [[ -z "$have_node" ]]; then
        qpargs+=( --node "$( prov_node )" )
    fi
    if [[ -z "$output" ]]; then
        output='json'
        qpargs+=( --output json )
    fi
    if [[ -z "$fn_ext" ]]; then
        printf -v fn_ext '.%s' "$( tr '[:upper:]' '[:lower:]' <<< "$output" )"
    fi
    if [[ -z "$start_i" ]]; then
        start_i='1'
    fi
    if [[ -n "$max_reqs" ]]; then
        last_i="$(( start_i - 1 + max_reqs ))"
    elif [[ "$start_i" -le '9999' ]]; then
        last_i='9999'
    else
        last_i="$start_i"
    fi
    printf 'Getting pages %d to %d of %s\n' "$start_i" "$last_i" "${q[*]}"
    for i in $( seq "$start_i" "$last_i" ); do
        printf -v fn '%s%04d%s' "$fn_base" "$i" "$fn_ext"
        cmd=( provenanced q "${query[@]}" "${qpargs[@]}" "${qargs[@]}" )
        if [[ -n "$next_key" ]]; then
            cmd+=( --page-key "$next_key" )
        fi
        printf '%4d/%4d: %s ' "$i" "$last_i" "${cmd[*]}"
        "${cmd[@]}" > "$fn"
        cmd_exit="$?"
        printf '> %s\n' "$fn"
        if [[ "$cmd_exit" -ne '0' ]]; then
            printf 'Stopping due to error.\n'
            return "$cmd_exit"
        fi
        next_key="$( jq -r '.pagination.next_key' "$fn" )"
        if [[ -z "$next_key" || "$next_key" == "null" ]]; then
            printf 'Stopping because the last result does not have a next_key\n'
            break
        fi
    done
    if [[ -n "$next_key" && "$next_key" != 'null' ]]; then
        printf 'Last result has next_key: %s\n' "$next_key"
    fi
    return 0
}

if [[ "$sourced" != 'YES' ]]; then
    query_prov_using_next_key "$@"
    exit $?
fi
unset sourced

return 0
