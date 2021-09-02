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

The `get_hash_price` function outputs the currently cached value (or a default if nothing's cached yet).
Then, if it thinks a refresh is needed (hasn't been refreshed in 10 minutes), it'll kick off that refresh in the background.

The `get_hash_price_for_prompt` function uses `get_hash_price` and reformats the output for use in a command prompt.

```console
$ PS1='$( get_hash_price_for_prompt ) $ '
 #⃣  0.0580  $ get_hash_price
0.058000000000000000
 #⃣  0.0580  $
```

The text color is bold white on a dark grey background.

The rest of the functions are to help facilitate caching.
- `dlobcache` is a wrapper over `bashcache` supplying the directory and age desired for in here.
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
