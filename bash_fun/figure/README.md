# SpicyLemon / bash_fun / figure

This directory contains files and scripts for doing things on a bash command-line.
These scripts/functions are specific to activities associated with Figure Technology Inc.

## Contents

* `figure-setup.sh` - This file is sourced in order to add all the functionality from the other scripts in this directory.
* `get_hash_price.sh` - This file has some functions used to display the HASH (Provenance utility coin) price in my command prompt.
* `b642id.sh` - Converts base64 encoded strings into a `MetadataAddress`, and display it's various pieces.
* `id2b64.sh` - Converts hex values (meant to make up a `MetadataAddress`) into a base64 encoded string.
* `query_prov_using_next_key.sh` - Gets multiple pages of a paginated provenanced query.
* `decode_events.sh` - Decodes the event strings returned from a tx query.

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
- `DLOB_C_DIR`: The directory the data is cached in. Default is `/tmp/dlob`.
- `DLOB_C_MAX_AGE`: The max age the cache can be to be considered still fresh. Default is `10m`.
- `DLOB_DAILY_PRICE_URL`: The url to use to get the json with the HASH price. Default is `https://www.dlob.io/aggregator/external/api/v1/order-books/pb18vd8fpwxzck93qlwghaj6arh4p7c5n894vnu5g/daily-price`.
- `DLOB_JQ_FILTER`: The filter provided to `jq` to extract the HASH price from the results of `DLOB_DAILY_PRICE_URL`. Default is `.latestDisplayPricePerDisplayUnit`.
- `DLOB_DEFAULT_VALUE`: The value to set as the HASH price if one can't be found. Default is `-69.420000000000000000`.
- `DLOB_PROMPT_FORMAT`: The format to apply to the HASH price to create the output of `get_hash_price_for_prompt`. Default is `\033[48;5;238;38;5;15m #\xE2\x83\xA3  %1.4f \033[0m`.

The rest of the functions are to help facilitate caching.
- `dlobcache` is a wrapper over `bashcache` supplying the directory and max age desired for this stuff.
- `dlobcache_refresh` actually does the work of making the API call and updating the cache.
- `dlobcache_check_required_commands` checks to make sure some possibly missing commands are available.

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
