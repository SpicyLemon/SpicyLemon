#!/usr/bin/env bash
# This file will update a provenance configuration so that it's set up to start using state-sync.
# Usage: state-sync-setup.sh [<provenanced command>] [<args for the command>]
# If <provenanced command> is not provided, provenanced is used.
# <args for the commmand> might be --testnet or --home ~/.provenanced.
# These will be included every single time the provenanced command is used.
# If your environment has exported PIO_ env vars, those will be used too.

prov_default='provenanced'
prov_cmd=''
prov_args=()
while [[ "$#" > 0 ]]; do
    case "$1" in
        -h|--help)
            cat << EOF
Usage: state-sync-setup.sh [<provenanced command>] [<persistent provenanced args>]

The <provenanced command> is the Provenanced Blockchain executable to use. The default is provenanced.
    If provided, it must be the first argument, and it cannot start with a dash.
The <persistent provenanced args> are any arguments to always provide with the <provenanced command>.
    Example <persistent provenanced args>: --home ~/.provenanced --testnet

Any exported PIO_ variables defined in your environment will also be used.

EOF
            exit 0
            ;;
        -t|--testnet)
            # Handled specially since we need to know for some urls and stuff.
            PIO_TESTNET='true'
            export PIO_TESTNET
            prov_args+=( "$1" )
            ;;
        *)
            if [[ -z "$prov_cmd" ]]; then
                if [[ "$1" =~ ^- ]]; then
                    prov_cmd="$prov_default"
                    prov_args+=( "$1" )
                else
                    prov_cmd="$1"
                fi
            else
                prov_args+=( "$1" )
            fi
            ;;
    esac
    shift
done

if [[ -z "$prov_cmd" ]]; then
    prov_cmd="$prov_default"
fi

# Make sure a few needed commands are available before we continue.
for c in jq curl "$prov_cmd"; do
    if ! command -v "$c" > /dev/null 2>&1; then
        printf 'Missing required command: %s\n' "$c" >&2
        "$c"
        exit $?
    fi
done

# get_ip gets the ip address for the domain in the provided url.
# Example Usage: get_ip http://example.com:1234/banana
get_ip () {
    local domain addr
    # Strip out any protocal and port, path, query, or fragment (should leave only the domain).
    domain="$( sed 's/^.*:\/\///; s/[:/?#].*$//' <<< "$1" )"
    # loose is-ip test is good enough in here. Any false positives here would not happen by accident.
    if [[ "$domain" =~ ^[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*$ ]]; then
        addr="$domain"
    else
        addr="$( host "$domain" | head -n1 | sed 's/^.* has address //' )" || exit $?
    fi
    if [[ ! "$addr" =~ ^[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*$ ]]; then
        printf 'Could not get ip address for domain: %s\n' "$domain" >&2
        printf '%s\n' "$addr" >&2
        return 1
    fi
    printf '%s' "$addr"
    return 0
}

prov=( "$prov_cmd" "${prov_args[@]}" )

printf 'Using Provenanced Blockchain command: %s\n' "${prov[*]}"

# Need to know the home so we can check for and create files.
if [[ -z "$PIO_HOME" ]]; then
    PIO_HOME="$( "${prov[@]}" config home )"
    export PIO_HOME
fi

printf 'Using home directory: %s\n' "$PIO_HOME"

if [[ -z "$PIO_TESTNET" || "$( tr '[:upper:]' '[:lower:]' <<< "$PIO_TESTNET" )" != 'true' ]]; then
    printf 'Setting up for mainnet.\n'
else
    printf 'Setting up for testnet.\n'
    is_testnet='yes'
fi


# Identify the config pieces that need to be downloaded or created.
conf_dir="$PIO_HOME/config"
node_key_file="$conf_dir/node_key.json"
priv_val_key_file="$conf_dir/priv_validator_key.json"
genesis_file="$conf_dir/genesis.json"
tm_config_file="$conf_dir/config.toml"
app_config_file="$conf_dir/app.toml"
client_config_file="$conf_dir/client.toml"
if [[ ! -e "$node_key_file" ]]; then
    need_node_key=true
fi
if [[ ! -e "$priv_val_key_file" ]]; then
    need_priv_val_key=true
fi
if [[ ! -e "$genesis_file" ]]; then
    need_genesis=true
fi
if [[ ! -e "$tm_config_file" ]]; then
    need_tm_config=true
fi
if [[ ! -e "$app_config_file" ]]; then
    need_app_config=true
fi
if [[ ! -e "$client_config_file" ]]; then
    need_client_config=true
fi

# Get the monikier. If the config doesn't exist yet, this will get the default that's based on this system.
moniker="$( "${prov[@]}" config get moniker | grep '^moniker=' | sed 's/^[^"]*"//; s/"[^"]*$//' )" || exit $?

if [[ -n "$need_node_key" || "$need_priv_val_key" ]]; then
    # The init command is used here to generate the node key and/or private validator key files.
    # That command only creates them if they don't already exist.
    # It will only run, though, if there isn't a genesis file.
    # So if there is one, we need to move it out of the way first, then move it back.
    # If there isn't one already, we will just overwrite it later with the downloaded one.
    # The init command also creates any config files that don't yet exist, but it shouldn't change them otherwise.
    # If they're created new, they're just defaults and need to be corrected, which is why we checked for their existence earlier.
    if [[ -e "$genesis_file" ]]; then
        genesis_file_bak="$genesis_file.bak"
        mv "$genesis_file" "$genesis_file_bak" || exit $?
    fi
    printf 'Creating node key and/or private validator key: %s init %s\n' "${prov[*]}" "'$moniker'"
    "${prov[@]}" init "$moniker" || exit $?
    if [[ -n "$genesis_file_bak" ]]; then
        mv "$genesis_file_bak" "$genesis_file" || exit $?
    fi
fi

if [[ -n "$need_genesis" ]]; then
    genesis_url='https://raw.githubusercontent.com/provenance-io/mainnet/main/pio-mainnet-1/genesis.json'
    if [[ -n "$is_testnet" ]]; then
        genesis_url='https://raw.githubusercontent.com/provenance-io/testnet/main/pio-testnet-1/genesis.json'
    fi
    printf 'Downloading Genesis file: curl %s > %s\n' "'$genesis_url'" "'$genesis_file'"
    curl "$genesis_url" > "$genesis_file" || exit $?
    if ! jq '.' "$genesis_file" > /dev/null 2>&1; then
        printf 'Downloaded genesis file is not json.\n' >&2
        exit 1
    fi
fi

if [[ -n "$need_tm_config" ]]; then
    tm_config_url='https://raw.githubusercontent.com/provenance-io/mainnet/main/pio-mainnet-1/node-config.toml'
    if [[ -n "$is_testnet" ]]; then
        tm_config_url='https://raw.githubusercontent.com/provenance-io/testnet/main/pio-testnet-1/node-config.toml'
    fi
    printf 'Downloading config.toml file: curl %s > %s\n' "'$tm_config_url'" "'$tm_config_file'"
    curl "$tm_config_url" > "$tm_config_file" || exit $?
    printf 'Setting moniker to the default: %s config set moniker %s\n' "${prov[*]}" "'$moniker'"
    "${prov[@]}" config set moniker "$moniker" || exit $?
fi

if [[ -n "$need_app_config" ]]; then
    app_config_url='https://raw.githubusercontent.com/provenance-io/mainnet/main/pio-mainnet-1/node-app.toml'
    if [[ -n "$is_testnet" ]]; then
        app_config_url='https://raw.githubusercontent.com/provenance-io/testnet/main/pio-testnet-1/node-app.toml'
    fi
    printf 'Downloading app.toml file: curl %s > %s\n' "'$app_config_url'" "'$app_config_file'"
    curl "$app_config_url" > "$app_config_file" || exit $?
fi

if [[ -n "$need_client_config" ]]; then
    # There's nothing to download for this, so just set the needed defaults.
    if [[ -z "$is_testnet" ]]; then
        chain_id='pio-mainnet-1'
        n0="$(( $RANDOM % 3 ))"
        node="tcp://rpc-$n0.provenance.io:26657"
        # Also define the rpc_servers value to use the other two nodes.
        rpc1="$( get_ip "rpc-$(( ( n0 + 1 ) % 3 )).provenance.io" ):26657" || exit $?
        rpc2="$( get_ip "rpc-$(( ( n0 + 2 ) % 3 )).provenance.io" ):26657" || exit $?
        printf -v rpc_servers '["%s","%s"]' "$rpc1" "$rpc2"
    else
        chain_id='pio-testnet-1'
        node='https://rpc.test.provenance.io:443'
    fi
    printf 'Setting client config values: %s config set chain-id %s node %s\n' "${prov[*]}" "'$chain_id'" "'$node'"
    "${prov[@]}" config set chain-id "$chain_id" node "$node" || exit $?
fi

# If we didn't define it above, set the rpc_servers.
if [[ -z "$rpc_servers" ]]; then
    printf 'Identifying RPC address.\n'
    if [[ -z "$is_testnet" ]]; then
        rpc_addr="$( get_ip "$( "${prov[@]}" config get node | grep '^node=' | sed 's/^[^"]*"//; s/"[^"]*$//' )" )" || exit $?
    else
        # (Temporary workaround due to how the tesntet hosts are currently configured)
        rpc_addr='34.66.209.228'
    fi
    rpc_addr="$rpc_addr:26657"
    printf -v rpc_servers '["%s","%s"]' "$rpc_addr" "$rpc_addr"
fi

# Get the current block and a previous block in order to set up the statesync config.
printf 'Getting latest block: %s query block\n' "${prov[*]}"
latest_block="$( "${prov[@]}" query block )" || exit $?
printf 'Getting block height: jq -r %s\n' "'.block.header.height'"
latest_height="$( jq -r '.block.header.height' <<< "$latest_block" )" || exit $?
printf 'Latest height: %d\n' "$latest_height"
trust_height="$(( latest_height - 1000 ))"
printf 'Getting trust block: %s query block --height %d\n' "${prov[*]}" "$trust_height"
trust_block="$( "${prov[@]}" query block "$trust_height" )" || exit $?
printf 'Getting trust hash: jq -r %s\n' "'.block_id.hash'"
trust_hash="$( jq -r '.block_id.hash' <<< "$trust_block" )" || exit $?

printf 'Updating config to set these statesync values:\n'
printf '  statesync.enable      : true\n'
printf '  statesync.rpc_servers : %s\n' "$rpc_servers"
printf '  statesync.trust_height: %d\n' "$trust_height"
printf '  statesync.trust_hash  : %s\n' "$trust_hash"
"${prov[@]}" config set \
    statesync.enable true \
    statesync.rpc_servers "$rpc_servers" \
    statesync.trust_height "$trust_height" \
    statesync.trust_hash "$trust_hash" || exit $?

data_dir="$PIO_HOME/data"
priv_val_state_file='priv_validator_state.json'
# If the data directory already exists, and has anything other than priv_validator_state.json,
# move it to a backup and create a fresh, empty version.
if ls "$data_dir" 2> /dev/null | grep -vFq "$priv_val_state_file"; then
    data_backup_dir="$data_dir-$( date +%s )"
    printf 'Moving existing data dir to a backup: mv %s %s\n' "'$data_dir'" "'$data_backup_dir'"
    mv "$data_dir" "$data_backup_dir" || exit $?
fi
if [[ ! -e "$data_dir" ]]; then
    printf 'Creating data dir: mkdir %s\n' "'$data_dir'"
    mkdir "$data_dir" || exit $?
fi
# Make sure the priv_validator_state.json exists.
if [[ ! -e "$data_dir/$priv_val_state_file" ]]; then
    if [[ -n "$data_backup_dir" && -e "$data_backup_dir/$priv_val_state_file" ]]; then
        printf 'Copying %s from backup: cp %s %s\n' "$priv_val_state_file" "'$data_backup_dir/$priv_val_state_file'" "'$data_dir/$priv_val_state_file'"
        cp "$data_backup_dir/$priv_val_state_file" "$data_dir/$priv_val_state_file" || exit $?
    else
        printf 'Creating new %s file.\n' "$priv_val_state_file"
        printf '{"height":"0","round":0,"step":0}' | jq '.' > "$data_dir/$priv_val_state_file" || exit $?
    fi
fi

# All done. Output some next steps.
cat << EOF

Your node is now configured to use statesync.
When starting your node, include the --x-crisis-skip-assert-invariants flag.
For example:
    $ ${prov[@]} start --x-crisis-skip-assert-invariants

EOF

if [[ -n "$need_tm_config" ]]; then
    cat << EOF
The moniker was set using the default value '$moniker'.
You probably want to change that to something more meaningful.
For example:
    $ ${prov[@]} config set moniker my-cool-node

EOF
fi

exit 0
