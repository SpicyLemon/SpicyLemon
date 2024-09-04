# SpicyLemon / bash_fun / figure

This directory contains files and scripts for doing things on a bash command-line.
These scripts/functions are specific to activities associated with Figure Technology Inc.

## Contents

* `figure-setup.sh` - This file is sourced in order to add all the functionality from the other scripts in this directory.
* `get_hash_price.sh` - This file has some functions used to display the HASH (Provenance utility coin) price in my command prompt.
* `b642id.sh` - Converts base64 encoded strings into a `MetadataAddress`, and display it's various pieces.
* `id2b64.sh` - Converts hex values (meant to make up a `MetadataAddress`) into a base64 encoded string.
* `query_prov_using_next_key.sh` - Gets multiple pages of a paginated provenanced query.
* `decode_events.sh` - Deprecated in favor of `get_events`: Decodes the event strings returned from a tx query.
* `get_events.sh` - Concisely print tx events and optionally decode them.
* `state-sync-setup.sh` - Sets up a directory to house a node that uses statesync.
* `cosmovisor-setup.sh` - Sets up a cosmovisor directory.
* `test_all.sh` - Runs a standard set of test make targets.
* `to_hash.sh` - Converts amounts of nhash into hash and includes commas.
* `estimate-block-time.sh` - Estimate future block heights for a time or time for a height.
* `get-block-times.sh` - Get information about block times, and how long each took to cut.

## Details

### `figure-setup.sh`

[figure-setup.sh](figure-setup.sh) - Checks and sources all the stuff for the functions in this folder.

Use this commmand to utilize the file:
```console
$ source figure-setup.sh
```

If you run into problems, you can use the `-v` option to get more information:
```console
$ source figure-setup.sh -v
```

### `get_hash_price.sh`

[get_hash_price.sh](get_hash_price.sh) - File with a few functions for looking up the current HASH price and caching it for use in my command prompt.

The functionality in here relies on the [bashcache](../bashcache.sh) command.

The main function is `get_hash_price`.

```console
$ get_hash_price
0.058000000000000000
```

The `get_hash_price` function outputs the currently cached HASH price (in USD).
If nothing is cached yet, the cache is updated and then the HASH price is printed.
If the cached HASH price is more than 10 minutes old, a background process is initiated to update the cache.

To refresh the cache and get the new value:
```console
$ get_hash_price --refresh
0.059000000000000000
```

To get the cached value (or default), without waiting on a refresh (e.g. when it doesn't exist yet):
```console
$ get_hash_price --no-wait
0.060000000000000000
```

To get the currently cached value and force an update in the background.
```console
$ get_hash_price --refresh --no-wait
0.061000000000000000
```

The `get_hash_price_for_prompt` function uses `get_hash_price --no-wait` and reformats the output for use in a command prompt.

```console
$ PS1='$( get_hash_price_for_prompt ) $ '
 #⃣  0.0580  $ get_hash_price
0.058000000000000000
 #⃣  0.0580  $
```

![screenshot of get hash price for prompt](/bash_fun/figure/get-hash-price-for-prompt-screenshot.png)

Some customizations can be made through the following environment variables:
- `HASH_C_DIR`: The directory the data is cached in. Default is `/tmp/hash`.
- `HASH_C_MAX_AGE`: The max age the cache can be to be considered still fresh. Default is `10m`.
- `HASH_PRICE_URL`: The url to use to get the json with the HASH price. Default is `https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=symbol,regularMarketPrice&symbols=HASH1-USD`.
- `HASH_JQ_FILTER`: The filter provided to `jq` to extract the HASH price from the results of `HASH_DAILY_PRICE_URL`. Default is `.quoteResponse.result[0].regularMarketPrice`.
- `HASH_DEFAULT_VALUE`: The value to set as the HASH price if one can't be found. Default is `-69.420000000000000000`.
- `HASH_PROMPT_FORMAT`: The format to apply to the HASH price to create the output of `get_hash_price_for_prompt`. Default is `\033[48;5;238;38;5;15m #\xE2\x83\xA3  %1.4f \033[0m`.

To use DLOB for pricing:
- `export HASH_PRICE_URL='https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price'`
- `export HASH_JQ_FILTER='.latestDisplayPricePerDisplayUnit'`

The rest of the functions are to help facilitate caching.
- `hashcache` is a wrapper over `bashcache` supplying the directory and max age desired for this stuff.
- `hashcache_refresh` actually does the work of making the API call and updating the cache.
- `hashcache_check_required_commands` checks to make sure some possibly missing commands are available.

### `b642id.sh`

[b642id.sh](b642id.sh) - Function/script to convert base64 encoded strings into a `MetadataAddress`, and display its various pieces.

Usage: `b642id <base64> [<base64 2> ...]`

Example:
```console
$ b642id AANhGlOv5EOnqFVbYpszHKs=
AANhGlOv5EOnqFVbYpszHKs= => 00 (scope) 03611a53-afe4-43a7-a855-5b629b331cab
```

It's counterpart is the `id2b64` function.

### `id2b64.sh`

[id2b64.sh](id2b64.sh) - Function/script to convert `MetadataAddress` hex into it's base64 representation.

Usage: `id2b64 <hex digits>`

Example:
```console
$ id2b64 00 03611a53-afe4-43a7-a855-5b629b331cab
AANhGlOv5EOnqFVbYpszHKs=
```

It's counterpart is the `b642id` function.

### `query_prov_using_next_key`

[query_prov_using_next_key.sh](query_prov_using_next_key.sh) - Function/script for getting multiple pages of a paginated provenanced query.

```
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
                         then the default --node is ''. Otherwise the default is 'tcp://rpc-0.provenance.io:26657'.
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
```

### `decode_events`

[decode_events.sh](decode_events.sh) - Function/script decoding the base64 encoded events from a tx JSON response.

**DEPRECATED**: This `decode_events` function/script has been deprecated in favor of `get_events`.
The direct replacement of `decode_events` is `get_events --decode --long`.
The only difference is that each line of the new output will start with `.events` instead of just `events`.

Either provide a JSON file or stream in some JSON with the results of a tx query and it will decode and output the events.
The output is one line per event attribute with this format:
```
{JSON path to event attribute} ({event type}): "{event attribute key}" = "{event attribute value}"
```

Example Use from file:
```console
$ provenanced q tx --type=hash 0ABDB417D4EBDE76AA4F3F2E8CBCE71600C385E955D5F7EA980B85E44A533639 -o json > 0ABDB417.json
$ decode_events 0ABDB417.json
events[0].attributes[0] (coin_spent): "spender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[0].attributes[1] (coin_spent): "amount" = "90000000000nhash"
events[1].attributes[0] (coin_received): "receiver" = "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"
events[1].attributes[1] (coin_received): "amount" = "90000000000nhash"
events[2].attributes[0] (transfer): "recipient" = "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"
events[2].attributes[1] (transfer): "sender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[2].attributes[2] (transfer): "amount" = "90000000000nhash"
events[3].attributes[0] (message): "sender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[4].attributes[0] (tx): "fee" = "100000000000nhash"
events[5].attributes[0] (tx): "acc_seq" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys/170"
events[6].attributes[0] (tx): "signature" = "Kn46lGBBbEyT8vkltURU8b0Q0h6aMQZ4mwAN5t6VclNbJAUJ7n5rJhxT9NhhUwstYcVPQZeL2AILEeFZ88mlVQ=="
```

Example Use from stream:
```console
$ provenanced q tx --type=hash 0ABDB417D4EBDE76AA4F3F2E8CBCE71600C385E955D5F7EA980B85E44A533639 -o json | decode_events
events[0].attributes[0] (coin_spent): "spender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[0].attributes[1] (coin_spent): "amount" = "90000000000nhash"
events[1].attributes[0] (coin_received): "receiver" = "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"
events[1].attributes[1] (coin_received): "amount" = "90000000000nhash"
events[2].attributes[0] (transfer): "recipient" = "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"
events[2].attributes[1] (transfer): "sender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[2].attributes[2] (transfer): "amount" = "90000000000nhash"
events[3].attributes[0] (message): "sender" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys"
events[4].attributes[0] (tx): "fee" = "100000000000nhash"
events[5].attributes[0] (tx): "acc_seq" = "tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys/170"
events[6].attributes[0] (tx): "signature" = "Kn46lGBBbEyT8vkltURU8b0Q0h6aMQZ4mwAN5t6VclNbJAUJ7n5rJhxT9NhhUwstYcVPQZeL2AILEeFZ88mlVQ=="
```

### `get_events`

[get_events.sh](get_events.sh) - Function/script that consicely outputs tx event info from a json file.

Example Use from file:
```console
$ provenanced q tx --type hash 0ABDB417D4EBDE76AA4F3F2E8CBCE71600C385E955D5F7EA980B85E44A533639 -o json > 0ABDB417.json
$ get_events 0ABDB417.json
[0]coin_spent[0]: spender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
[0]coin_spent[1]: amount = 90000000000nhash
[1]coin_received[0]: receiver = tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
[1]coin_received[1]: amount = 90000000000nhash
[2]transfer[0]: recipient = tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
[2]transfer[1]: sender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
[2]transfer[2]: amount = 90000000000nhash
[3]message[0]: sender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
[4]tx[0]: fee = 100000000000nhash
[5]tx[0]: acc_seq = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys/170
[6]tx[0]: signature = Kn46lGBBbEyT8vkltURU8b0Q0h6aMQZ4mwAN5t6VclNbJAUJ7n5rJhxT9NhhUwstYcVPQZeL2AILEeFZ88mlVQ==
```

Example Use from stream and with long output format:
```console
$ provenanced q tx --type hash 0ABDB417D4EBDE76AA4F3F2E8CBCE71600C385E955D5F7EA980B85E44A533639 -o json | get_events --long
.events[0].attributes[0] (coin_spent): spender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
.events[0].attributes[1] (coin_spent): amount = 90000000000nhash
.events[1].attributes[0] (coin_received): receiver = tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
.events[1].attributes[1] (coin_received): amount = 90000000000nhash
.events[2].attributes[0] (transfer): recipient = tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
.events[2].attributes[1] (transfer): sender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
.events[2].attributes[2] (transfer): amount = 90000000000nhash
.events[3].attributes[0] (message): sender = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys
.events[4].attributes[0] (tx): fee = 100000000000nhash
.events[5].attributes[0] (tx): acc_seq = tp172yscg9eu72hknhue4sae5z3yyddxlfsfntcys/170
.events[6].attributes[0] (tx): signature = Kn46lGBBbEyT8vkltURU8b0Q0h6aMQZ4mwAN5t6VclNbJAUJ7n5rJhxT9NhhUwstYcVPQZeL2AILEeFZ88mlVQ==
```

```console
$ get_events --help
get_events - Concisely display tx events.

Usage: get_events <tx json file> [--path|-p <path>] [--decode|-d] [--long|-l]
   or: <stuff> | get_events [--path|-p <path>] [--decode|-d] [--long|-l]

The --path <path> option allows you to define the json path to the list of events.
    The default <path> is '.events'.
The --decode flag will cause the attribute keys and values to be base64 decoded.
The --long flag causes the full json path to each attribute to be displayed instead of a shorter form.
    Standard output format: [<event index>]<event type>[<attribute index>]: <key> = <value>
    Long output format:     <path to events>[<event index>].attributes[<attribute index>] (<event type>): <key> = <value>

```


### `state-sync-setup`

[state-sync-setup.sh](state-sync-setup.sh) - Script that sets up the current directory to house a node that uses statesync.

```console
> ./state-sync-setup.sh --help
Usage: state-sync-setup.sh [<provenanced command>] [<persistent provenanced args>]

The <provenanced command> is the Provenanced Blockchain executable to use. The default is provenanced.
    If provided, it must be the first argument, and it cannot start with a dash.
The <persistent provenanced args> are any arguments to always provide with the <provenanced command>.
    Example <persistent provenanced args>: --home ~/.provenanced --testnet

Any exported PIO_ variables defined in your environment will also be used.
```

### `cosmovisor-setup`

[cosmovisor-setup.sh](cosmovisor-setup.sh) - Script that creates a cosmovisor directory.

```console
> ./cosmovisor-setup.sh --help
Usage: ./cosmovisor-setup.sh [--home <daemon_home>] [--name <daemon_name>] [--path <path_to_daemon>]

This script will create the initial cosmovisor directory structure.

<daemon_home> is the directory that will hold the cosmovisor/ directory.
    If not provided, the DAEMON_HOME environment variable is used.
    If DAEMON_HOME is not defined, the PIO_HOME environment variable is used.
    If PIO_HOME is also not defined, an error is returned.

<daemon_name> is the name of the executable.
    If not provided, the DAEMON_NAME environment variable is used.
    If DAEMON_NAME is not defined, but a <path_to_daemon> is provided, the filename from that will be used.

<path_to_daemon> is the full path to the executable.
    If not provided, the location will be found using  command -v <daemon_name> .
    If the executable file cannot be found, an error is returned.

```

### `test_all`

[test_all.sh](test_all.sh) - Function/script to run some test-related make targets available in the provenance and cosmos-sdk repos.

```console
> test_all --help
Usage: test_all [[--skip|-s] <targets>] [[--also|-a] <targets>] [[--targets|-t] <targets>]
                [--continue|-c|--break-b] [--sound [on|off|beep|say]|--noisy|--quiet|--beep|--say]

By default, the following make targets are run:
  test test-sim-nondeterminism test-sim-import-export test-sim-after-import test-sim-multi-seed-short

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
```

### `to_hash`

[to_hash.sh](to_hash.sh) - Function/script that converts nhash amounts to hash, with commas.

Example:

```console
$ to_hash 2398473897439322
2,398,473.897439322 hash
```

### `estimate-block-time`

[estimate-block-time.sh](estimate-block-time.sh) - Script that will estimate future blocks and times.

The `figure-setup.sh` script will create the alias `estimate-block-time` to this file.

Estimate the height at a future time:
```console
$ estimate-block-time '2025-04-20 04:20:00'
Chain-Id: pio-mainnet-1
Current: Mon 2024-08-12 18:03:49 -0600 (MDT) Height: 18318320
Desired: Sun 2025-04-20 04:20:00 -0600 (MDT) Height: 22573864
Elapsed milliseconds: 21636970229 = 4255544 blocks (at 5084.419 milliseconds per block from last 10000 blocks).
```

Estimate when a block will happen:
```console
$ estimate-block-time 20000000
Chain-Id: pio-mainnet-1
Current: Mon 2024-08-12 18:05:00 -0600 (MDT) Height: 18318336
Desired: Thu 2024-11-07 02:41:36 -0700 (MST) Height: 20000000
Elapsed milliseconds: 7464995624 = 1681664 blocks (at 4439.053 milliseconds per block from last 1681664 blocks).
```

See `estimate-block-time.sh --help` for full usage information.

### `get-block-times`

[get-block-times.sh](get-block-times.sh) - Script that will output information about block times including how long each took to cut.

A cache directory is used to store block information. If any blocks are needed, but not in the cache, they will be retrieved first.

Example:
```
$ get-block-times 18310001 18310010
Querying for 12 blocks.
[1/12]: Querying for block 18310000 and storing result: ./archive/blocks/block-18310000.json
[11/12]: Querying for block 18310010 and storing result: ./archive/blocks/block-18310010.json
[12/12]: Querying for block 18310011 and storing result: ./archive/blocks/block-18310011.json
Stamp                           Height    Time     Tx  R  Votes
2024-08-12T13:35:20.783455733Z  18310001   4.945s   0  1  56/63
2024-08-12T13:35:30.334802694Z  18310002   9.551s   0  0  56/63
2024-08-12T13:35:35.273044370Z  18310003   4.939s   0  0  56/63
2024-08-12T13:35:37.749266832Z  18310004   2.476s   0  0  56/63
2024-08-12T13:35:40.135845046Z  18310005   2.386s   0  1  56/63
2024-08-12T13:35:49.667859637Z  18310006   9.532s   0  0  56/63
2024-08-12T13:35:52.097876318Z  18310007   2.430s   0  0  56/63
2024-08-12T13:35:54.470095116Z  18310008   2.373s   0  2  56/63
2024-08-12T13:36:09.652685386Z  18310009  15.182s   0  1  56/63
2024-08-12T13:36:19.260191563Z  18310010   9.608s   0  0  56/63
```

See `get-block-times --help` for full usage information.

The only thing `get-block-times` prints to stdout is the resulting table (including its header).
All other output is to stderr (e.g. the "Querying for block ..." lines).

