# SpicyLemon
Dumping ground for fun random stuff that I've put together.

## Contents

* `bash_fun` - Scripts and functions for doing stuff in bash (and also in other shells in most cases).
* `go_fun` - Packages and stuff relating to Go.
* `js_fun` - Stuff relating to javascript.
* `perl_fun` - Stuff that I've done in Perl, and felt like sharing.
* `ticker` - A bash script for looking up stock prices.

## Favorites

### [bootRun.sh](bash_fun/bootRun.sh)

This script makes it easier to run Gradle based Spring Boot programs with command-line arguments.

Hopefully, in most cases, it can just be copied into a project next to the gradle wrapper file.
In the other cases, hopefully the only customization needed is with a couple variables that are set at the top of the script.

Examples:
```bash
./bootRun.sh
```
Just calls
```bash
./gradlew bootRun
```

Arguments can also be supplied just like we're all used to.
```bash
./bootRun.sh arg1 --flag2 argument3 argument4 'arguments "with" fancy stuff.'
```
Will end up running
```bash
./gradlew bootRun -Pargs=arg1,--flag2,argument3,argument4,"arguments \"with\" fancy stuff."
```

### [GitLab CLI Stuff](bash_fun/gitlab)

This directory houses a several functions for interacting with GitLab through the command line.

To use it, download the directory to somehwere handy for you and in your environment startup script (e.g. `.bash_profile`) add a line to source the gitlab-setup.sh file.
This stuff also depends on the functions in [curl_link_header.sh](bash_fun/curl_link_header.sh) and [fzf_wrapper.sh](bash_fun/fzf_wrapper.sh).
Either place those in the same folder as the `gitlab-setup.sh` file, or the same folder as the `gitlab` directory.
ALternatively, you could also place them somewhere else and source them prior to sourceing the `gitlab-setup.sh` file.

See `gitlab --help` for more info.

### [calculation-template.html](js_fun/calculation-template.html)

This is an HTML page with a bunch of template Javascript on it for doing calculation-heavy javascript stuff.

There are two script sections to it. The top one becomes a web worker that is called by the bottom one.
This allows you to put all the input, ouput, parsing and calculation pieces in one page while also making the page remain responsive while the calculation is running.

The file has a comment at the top with details on usage.

### [curl_link_header.sh](bash_fun/curl_link_header.sh)

This is a script/function that uses curl and follows entries in the [link response header](https://tools.ietf.org/html/rfc5988#section-5) in order to collect extra information.
The file can either be sourced (e.g. in your `.bash_profile`) to add the function to your environment, or else it can be executed like most other script files.

This is handy for getting paginated API results where the API responses include link headers.

You basically provide the parameters that you'd normally provide to curl, with some extra options available, but also some restrictions.
Curl gets the contents of the provided url, then looks in the response header for a link header, gets the desired entry, and uses curl to get that entry.
This continues until either the link header no longer contains a desired entry, there is no link header, or a maximum number of calls is made.

You can see this in action in the `__gl_get_all_results` function defined in [bash_fun/gitlab/gl-core.sh](bash_fun/gitlab/gl-core.sh)

### [fzf_wrapper.sh](bash_fun/fzf_wrapper.sh)

The primary purpose of this file is to define the `fzf_wrapper` function.
This function adds a `--to-columns` option to [fzf](https://github.com/junegunn/fzf).
When `--to-columns` is supplied, the string defined by `-d` or `--delimiter` becomes the string given to the `column` command in order to display columnar entries in fzf.
But the selected entries remain the same as they were when provided.

For example:
```bash
echo -e "a1111~a2~a3~a4\nb1~b222~b3~b4\nc1~c2~c3~c44\n" | __fzf_wrapper --with-nth='1,2,4' --delimiter='~' --to-columns
```
will show this:
```
c1     c2    c44
b1     b222  b4
a1111  a2    a4
```
But the selected entry will still be what was originally supplied, e.g. `a1111~a2~a3~a4`.

This is done by wrapping the provided delimiter with a zero-width space to the left of it, and a zero-width non-joiner to the right of it.
Then the input is sent to the column command using the provided delimiter.
That is then sent to fzf using a zero-width non-joiner for the `--delimiter` (the rest of the options are unchanged).
Lastly, the result(s) from fzf are changed back to their original state by replacing the zero-width space,
followed by spaces, then the zero-width non-joiner, with the original delimiter.

Without the `--to-columns` option, there is no change to the functionality of fzf or any of the provided options.
As such, it should be safe to `alias fzf=fzf_wrapper` and not notice any difference except the availability of the `--to-columns` option.

### [bashcache.sh](bash_fun/bashcache.sh)

This script/function is a handy way for other CLI stuff to cache things in temp files. The directory to use can be either set with an environment variable or provided as an argument. The max age can be set similarly. Like a lot of my bash scripts, the file can either be sourced to load the `bashcache` function into your environment, or it can be executed.

For full usage info, see `./bashcache.sh --help`

```
Usage: bashcache <command> <cache name> [<options>]
```

The `<command>`s are `write`, `read`, `check`, `file`, `list`, and `delete`.

The `<cache name>` is whatever you want to use to identify the data being cached. It's actually used as part of the filename, so some restrictions might apply.

The idea is that, whenever you feel like it, you can `write` the cache data. It's up to you to get the data to cache though. It can be piped in, sent in with a heredoc, sent in with a herestring, or even provided as part of the command arguments. If anything already exists there, it is overwritten.

Then, when you want the data, you would `read` it. If you just want to see if it needs updating, you can `check` it. If you need to write the file yourself, (e.g. caching a response header with `curl`), you can use the `file` command to get the full path to the cache file. All three of those (`read`, `check`, and `file`) exit with specific codes to indicate the state of the data. It's up to you to decide if/when to update the data, but this will make it easier to get the data and know when to update it.

Exit codes:
* `0` (zero): Okay. The data exists and is fresh.
* `1` (one): Bad arguments. The arguments provided to `bashcache` are invalid.
* `10` (ten): The requested data exists, but is stale.
* `11` (eleven): The requested data is not available.

Basically, if the return code is 10 or more, you might want to update the data.

Example:
```
my_name="$( bashcache read my_name )"
bce=$?
if [[ "$bce" -ge '10' ]]; then
    printf('The my_name cache should be updated.\n')
fi

```

The `delete` command is used to remove a cached entry.

The `list` command will output all the `<cache name>`s currently available. You can also provide the `--details` flag to the `list` command for extra information.

There are only a few options:
* `--verbose` or `-v`: Cause bashcache to output extra information to stderr.
* `--dir` or `-d` `<directory>`: This allows you to define the directory where the cached data will live. The default is `/tmp/bashcache`.
* `--max-age` or `--age` or `-a` `<max age>`: Define the age limit for stale data. This uses the same format as the `-atime` option of `find`, except without the leading `+`. E.g. `10m`, `24h`, `6d5h4m3s`. The default is `24h`

There are environment variables for the `<directory>` and `<max age>` too: `BASHCACHE_DIR` and `BASHCACHE_MAX_AGE`.

Order of precidence: CLI argument > environment variable > default.

Whether the data is fresh or stale is checked on every invocation. For example, if some data is 1 hour old, and you request it with a `--max-age` of `1m`, it will indicate that it's stale. Immediately after that, if you request it again with a `--max-age` of `24h`, it'll indicate that it's fresh.

An example if it being used can be found in the [get_hash_price.sh](bash_fun/figure/get_hash_price.sh) script. This script is used to display the results of an api call in my command prompt. Rather than making that request (and waiting for it) every time I hit enter, I used `bashcache` to manage the data. Additinally, because it's being held in files, it's shared by all my terminals, so I'm not spamming requests out.
