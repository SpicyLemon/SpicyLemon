# SpicyLemon / bash_fun / figure

This directory contains files and scripts for doing things on a bash command-line.
These scripts/functions are specific to activities associated with Figure Technology Inc.

## Contents

* `figure-setup.sh` - This file is sourced in order to add all the functionality from the other scripts in this directory.
* `get_hash_price.sh` - This file has some functions used to display the HASH (Provenance utility coin) price in my command prompt.
* `b642id.sh` - Converts base64 encoded strings into a `MetadataAddress`, and display it's various pieces.
* `id2b64.sh` - Converts hex values (meant to make up a `MetadataAddress`) into a base64 encoded string.

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
