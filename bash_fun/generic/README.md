# SpicyLemon / bash_fun / generic

This directory contains a whole bunch of functions/scripts for doing various things in bash.
In most cases, the scripts/functions should work in other shells too, but that hasn't been fully tested.

<hr>

## Table of Contents

TODO

## Usage

### Setup/Installation

1.  Copy this directory to a safe location on your computer.
    I personally, have a `~/.functions/` folder for such files and directories.
    So I've got a `~/.functions/generic/` folder with all these files.
1.  In your environment setup file (e.g. `.bash_profile`), add a line to source the `generic-setup.sh` file.
    For example, in mine, I have this line:
    ```bash
    source "$HOME/.functions/generic/generic-setup.sh"
    ```
    In order to add these functions to an already open environment, execute the same command.

If you need to troubleshoot the setup, you can add a `-v` flag when sourcing the setup file: `source generic-setup.sh -v`.

### Program/Function Requirements

These programs are used by some functions, and don't usually come pre-installed:
* `jq` - Json processor. See https://github.com/stedolan/jq
* `fzf` - Fuzzy finder. See https://github.com/junegunn/fzf

These programs are also used, but are almost always already available:

    `awk`,   `basename`,  `cat`,   `date`,  `dirname`,  `echo`,     `fzf`,
    `grep`,  `head`,      `jq`,    `open`,  `pbcopy`,   `pbpaste`,  `printf`,
    `ps`,    `sed`,       `tail`,  `tee`,   `tput`,     `tr`,       `/usr/libexec/java_home`

## Directory Contents

### `generic-setup.sh`

The file to source to source the rest of these files, importing the functions into your environment.

Use this commmand to utilize the file:
```console
$ source generic-setup.sh
```

If you run into problems, you can use the `-v` option to get more information:
```console
$ source generic-setup.sh -v
```

### `add_to_filename.sh`

A function/script for adding text to one or more filename strings, just before the extensions.

Simple example - Add "-v2" to the name of a document.
```console
$ add_to_filename "-v2" "my-important-document.doc"
my-important-document-v2.doc
```

Multi-file example - Add "-old" to the name of every .sh file in this directory that starts with a "g":
```console
$ add_to_filename "-old" g*.sh
generic-setup-old.sh
get_all_system_logs-old.sh
get_shell_type-old.sh
getlines-old.sh
```

Multi-file example coming in a stream:
```console
$ ls g*.sh | add_to_filename "-new" -
generic-setup-new.sh
get_all_system_logs-new.sh
get_shell_type-new.sh
getlines-new.sh
```

This file can also be executed as a script for the same functionality.

### `change_word.sh`

A function/script for rudamentarily changing one string to another in a set of files.

Usage: `change_word <old word> <new word> <files>`

The initial use case for this was to change the name of a function and get a record of the changes made.
As such, only letters, numbers, dashes, underscores and spaces are allowed in the "words".
Word barriers are applied to the provided "words" too.
So this won't make partial changes to words.

When executed:
1.  A colorized list of "before" entries are listed.
1.  The changes are made.
1.  A colorized list of "after" entries are listed.

This is probably best done in a git repo containing no uncommitted changes.
That way, it's easy to undo if there were some unintended changes.

### `check_system_log_timestamp_order.sh`

A function/script for finding improperly ordered entries in system logs.

I had a weird thing happen where some system log entries weren't in chronological order.
So I ended up putting this together in order to find all such entries in a file.

Simple example:
```console
$ check_system_log_timestamp_order /var/log/system.log
```

Doing it for all system logs:
```console
$ { cat /var/log/system.log; for l in $( ls /var/log/system.log.* ); do zcat < "$l"; done; } | check_system_log_timestamp_order -
```

### `chrome_cors.sh`

A function/script for opening a page in a Chrome window with CORS security disabled.
This is handy when you're trying to set up some ajax scripts locally for a service running elsewhere that includes CORS headers in the response.

```console
$ chrome_cors my-own-page.html
```

### `echo_color.sh`

A function/script for outputting command-line stuff with colors.

If the file is executed, it just runs the `echo_color` function with the provided parameters.

If the file is sourced, though (either specifically or via `generic-setup.sh`), several other helper functions become available too:

* The `show_colors` function will output all sorts of usage and color info.
  This function can also be accessed using `echo_color --examples`.
* Shortcuts for specific colors: `echo_red`, `echo_green`, `echo_yellow`, `echo_blue`, `echo_cyan`
* Shortcuts for specific effects: `echo_bold`, `echo_underline`
* Shortcuts for specific formats: `echo_debug`, `echo_info`, `echo_warn`, `echo_error`, `echo_success`, `echo_good`, `echo_bad`

```console
$ echo_color --help
echo_color - Makes it easier to output things in colors.

Usage: echo_color <paramters> -- <message>
    Any number of parameters can be provided in any order.
    The parameters must be followed by a --.
    Everything that follows the -- is considered part of the message to output.

Valid Parameters: <name> <color code> -n -N --explain --examples
    <name> can be one of the following:
            Text (foreground) colors:
                black      red            blue        green
                dark-gray  light-red      light-blue  light-green
                light-gray magenta        cyan        yellow
                white      light-magenta  light-cyan  light-yellow
            Background colors:
                bg-black      bg-red            bg-blue        bg-green
                bg-dark-gray  bg-light-red      bg-light-blue  bg-light-green
                bg-light-gray bg-magenta        bg-cyan        bg-yellow
                bg-white      bg-light-magenta  bg-light-cyan  bg-light-yellow
            Effects:
                bold  dim  underline  strikethrough  reversed
            Special Formats:
                debug  info  warn  error  success  good  bad
    <color code> can be one or more numerical color code numbers separated by semicolons or spaces.
            Values are delimited with semicolons and placed between "<esc>[" and "m" for output.
            Examples: "31" "38 5 200" "93;41" "2 38;5;141 48;5;230"
            This page is a good resource: https://misc.flogisoft.com/bash/tip_colors_and_formatting
            Spaces are converted to semicolons for the actual codes used.
    -n signifies that you do not want a trailing newline added to the output.
    -N signifies that you DO want a trailing newline added to the output. This is the default behavior.
            If both -n and -N are provided, whichever is latest in the paramters is used.
    --explain will cause the begining and ending escape codes to be output via stderr.

    --examples will cause the any previous parameters to be ignored, and instead output a set of examples.
            All parameters that follow this option are supplied to the  show_colors  function.
            See  show_colors --help  or  echo_color --examples --help  for more information.

Examples:
    > echo_color underline -- "This is underlined."
    This is underlined

    > echo_color bold yellow bg-light-red -- "Would anyone like a hotdog?"
    Would anyone like a hotdog?

    > echo_color light-green -- This is a $( echo_color reversed -- complex ) message.
    This is a complex message.
```

### `echo_do.sh`

A function/script for outputting a command prior to executing it.

The provided command is printed in bold white, then executed.

Example Usage:
This would go through each .sh file in the current directory, and diff that file with the same-named file in `/some/other/dir`.
The diff command would show in bold white followed by the results of the diff command.
```console
$ for f in $( ls *.sh ); do echo_do diff "$f" "/some/other/dir/$f"; done
```

### `escape_escapes.sh`

Function/script for escaping the ASCII escape character (octal `\033`, hex `\x1B`).

Example Usage:
```console
$ echo_color red -- This is red | escape_escapes
\033[31mThis is red\033[39m
```

### `fp.sh`

Function/script for getting the full path to a file.

This basically prepends your current working directory to the provided strings then corrects for any instances of `../`.
If only a single entry is provided, and `pbcopy` is available, the result will be loaded into your clipboard.
If no entries are provided, you will be prompted to select some from the current directory.

Example Usage:
```console
$ fp README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/generic/README.md - copied to clipboard.
```

### `get_all_system_logs.sh`

Function/script that pulls all system logs into a single file and sorts the entries by the stamp.

It gets the `/var/log/system.log` file and decompresses any `/var/log/system.log.*` files.
Then it sorts the entries by date and outputs them to stdout.
There's usually quite a lot of output.

Example Usage:
```console
$ get_all_system_logs
```

### `get_shell_type.sh`

Function for telling which shell is being used. Currently only recognizes bash and zsh.

Originally, I was using this to split code that needed to be different for bash and zsh.
But then I realized that it's better to do checks for specific functionality, so I don't really use this anymore.
I didn't feel like getting rid of it, though.

Example Usage:
```console
$ get_shell_type
bash
```

### `getlines.sh`

Function/script for getting specific lines from a file by line number.

It can take in specific line numbers, line number ranges, and any combination of those.
A filename can also be provided.
If no filename is provided, it'll attempt to get input that is piped in.
Lines are printed in numerical order (as opposed to the order that they're provided as arguments).

Example Usage:
```console
$ getlines 60-64 10-15 fp.sh
# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    fp "$@"
    exit $?
fi
unset sourced
```

### `hrr.sh`

Function/script that outputs a colorful horizontal rule in your terminal.

A message can be provided to include in the output too.
You won't see it here (in this README), but it's colorized too.
A random palette is chosen each time it's called.

If you source the `hrr.sh` file, the `hrr` and `hhr` functions both become availalbe (so you don't have to remember which one it is).
The `hr` function also becomes available.
The `hrr` (and `hhr`) functions output 3 lines, and the `hr` function only outputs one.
The width is determined by your environment.

Example Usage:
```console
$ hrr This is a test of hrr
 ################################################################################################################################################
  ############################################################ This is a test of hrr ############################################################
 ################################################################################################################################################
```
or
```console
$  hr This is a test of hr
  ############################################################ This is a test of hr ############################################################
```

### `i_can.sh`

Function for testing whether or not a command is available in the environment.

It takes in the primary command that you want to test and exits with an exit code of 0 if the command is available, or 1 if not.

This file also provides the `can_i` method that uses the `i_can` method to output whether or not a command is available.

Example Usage:
```console
$ if i_can pbcopy; then cat README.txt | pbcopy && printf 'README.txt copied to clipboard.\n'; else printf 'Nothing happened.\n'; fi
README.txt copied to clipboard.
$ can_i pbpaste
Yes. You can [pbpaste].
$ can_i dance
No. You cannot [dance].
```

### `java_8_activate.sh`

Function for setting the `JAVA_HOME` variable to the Java 8 JDK.

I'm not sure how helpful this will be to others, but it's handy for me.
Rather than having to remember what value the `JAVA_HOME` variable needs to have in order to use Java 8, I just have to remember the `java_8_activate` command.

Example Usage:
```console
$ java_8_activate
JAVA_HOME set to "/Library/Java/JavaVirtualMachines/jdk1.8.0_77.jdk/Contents/Home".
```

### `java_8_deactivate.sh`

Function for clearing the `JAVA_HOME` variable, going back to the system default.

All this does is unsets the `JAVA_HOME` variable.
But having it named similar to `java_8_activate` makes it a little easier to remember.

Example Usage:
```console
$ java_8_deactivate
JAVA_HOME unset.
```

### `join_str.sh`

Function/script for joining strings using a delimiter.

The first argument provided is treated as the delimiter.
All other arguments are then joined together into one line, with the delimiter between each entry.
A trailing newline character is *not* added.

Example Usage:
```console
$ join_str ', ' $( seq 1 10 )
1, 2, 3, 4, 5, 6, 7, 8, 9, 10
```

### `jqq.sh`

Function/script for running json contained in a variable through the `jq` program.

I got tired of typing `echo "$variable" | jq '.'`.
So I created this, which saves a couple characters of typing, but is a bit clunky.
I really don't use it much anymore.
Instead, I usually use a herestring: `jq '.' <<< "$variable"`.

Example Usage:
```console
$ jqq '{"foo":"FOO","bar":"BAR"}' '.foo' -r
FOO
```

### `multi_line_replace.sh`

Function/script for replacing a matched area of a single line with a multi-line replacement.

This is basically a sort of templating script.
It takes in three arguments:
1.  The filename of the template.
1.  The string in the template to replace.
1.  The replacement text.

The output is similar to the command `sed 's/string to replace/replacement text/g' filename` with one difference.
The line containing the `string to replace` is replicated for each line in the `replacement text`.
Then each line in the replacement text replaces the the string to replace.

For example, if your template looks like this:
```console
$ cat multi-line-sample-template.txt
This is the first line.
This line contains the replacement [__REPLACE_ME__].
This is the middle line.
This line contains another copy of the replacement [__REPLACE_ME__].
This is the last line.
```
Then you could do this:
```console
$ multi_line_replace multi-line-sample-template.txt '__REPLACE_ME__' "$( printf 'entry %d\n' $( seq 5 ) )"
This is the first line.
This line contains the replacement [entry 1].
This line contains the replacement [entry 2].
This line contains the replacement [entry 3].
This line contains the replacement [entry 4].
This line contains the replacement [entry 5].
This is the middle line.
This line contains another copy of the replacement [entry 1].
This line contains another copy of the replacement [entry 2].
This line contains another copy of the replacement [entry 3].
This line contains another copy of the replacement [entry 4].
This line contains another copy of the replacement [entry 5].
This is the last line.
```

### `pretty_json.sh`

Function/script for using jq to make json pretty.

This uses jq to make json pretty.
It's similar to just doing `jq --sort-keys '.' <filename>` except it's got some nice bonus features.

It can get the input from a file, a pipe, the clipboard, or even as a raw json string as an argument.
It can also output to a file, stdout, or the clipboard.
STDOUT output will always contain color though.
That's why the ability to output to a file was added.
File output, and clipboard output do not contain color info.
If the input is a file, and the output is a file, the output filename can be automatically generated from the input filename.

See `pretty_json --help` for more info.

Example Usage:
```console
$ pretty_json -- '{"a":"A","b":"B"}'
{
  "a": "A",
  "b": "B"
}
```

The counterpart to this function is `ugly_json` (listed below).

### `print_args.sh`

Function/script for printing out the arguments passed into it.

This is a really simple thing that is usefuly when you want to see what arguments look like as they're passed into a function.

Example Usage:
```console
$ print_args turn down for='what'
Arguments received:
 1: [turn]
 2: [down]
 3: [for=what]
```

### `ps_grep.sh`

Function/script for grepping through `ps aux` output.

I often found myself running the command `ps aux | grep <something>`.
This function makes that a few characters shorter, but also provides some niceties.
The first nice thing is that the result won't include the command you're using to do the search.
The second nice thing is that the output is colored by default, and the `ps` header is also listed.

The arguments provided to `ps_grep` are the same as you would pass to the `grep` command.

Example Usage:
```console
$ ps_grep README
USER               PID  %CPU %MEM      VSZ    RSS   TT  STAT STARTED      TIME COMMAND
spicylemon        4645   0.0  0.0  4334968   4004 s008  S+   10:52PM   0:04.43 vi README.md
```

### `re_line.sh`

Function/script for reformatting delimited and/or line-separated entries.

This is a fun function.
It takes in collections of entries, and reformats them.

For example, if you have a file with one entry per line, and you want to have five entries per line, each separated by a comma space, you could use this command:
```console
$ re_line -f one-entry-per-line.txt -n 5 -d ', ' > five-entries-per-line.txt
```

Then you decied, you want them on lines with a maximum width of 80 characters, and you want each entry contained in square brackets.
```console
$ re_line -f one-entry-per-line.txt --max-width 80 -d ', ' -l '(' -r ')' > max-width-80.txt
```

Later, you're looking at `five-entries-per-line.txt` and decide you'd rather have 12 entries per line, with each wrapped in single quotes.
```console
$ re_line -f five-entries-per-line.txt -n 12 -b ',[[:space:]]*' -d ', ' -w "'" > twelve-entries-per-line.txt
```

```console
$ re_line --help
re_line - Reformats delimited and/or line-separated entries.

Usage: re_line [-f <filename>|--file <filename>|-c|--from-clipboard|-|-p|--from-pipe|-- <input>]
               [-n <count>|--count <count>|--min-width <width>|--max-width <width>]
               [-d <string>|--delimiter <string>] [-b <string>|--break <string>]
               [-w <string>|--wrap <string>] [-l <string>|--left <string>] [-r <string>|--right <string>]

    -f or --filename defines the file to get the input from.
    -c or --clipboard dictates that the input should be pulled from the clipboard.
    - or -p or --from-pipe indicates that the input is being piped in.
        This can also be expressed with -f - or --filename -.
    -- indicates that all remaining parameters are to be considered input.
    Exactly one of these input options must be provided.

    -n or --count defines the number of entries per line the output should have.
        Cannot be combined with --min-width or --max-width.
        A count of 0 indicates that the output should not have any line-breaks.
    --min-width defines the minimum line width (in characters) the output should have.
        Once an item is added to a line that exceeds this amount, a newline is then started.
        Cannot be combined with -n, --count or --max-width.
    --max-width defines the maximum line width (in characters) the output shoudl have.
        If adding the next item to the line would exceed this amount, a newline is started
        and that next item is the first item on it.
        Note: A line can still exceed this width in cases where a single item exceeds this width.
        Cannot be combined with -n, --count or --min-width.
    If none of -n, --count, --min-width, or --max-width are provided, the default is -n 10.
    If more than one of -n, --count, --min-width or --max-width are provided, the one provided last is used.

    -d or --delimiter defines the delimiter to use for the output.
        The default (if not supplied) is a comma followed by a space.
    -b or --break defines the delimiter to use on each line of input.
        The default (if not supplied) is a comma and any following spaces.
        These are not considered to be part of any item, and will not be in the output.
        To turn off the splitting of each line, use -b ''.
        The string is used as the LHS of a sed s/// statement.
    -w or --wrap defines a string that will be added to both the beginning and end of each item.
    -l or --left defines a string that will be added to the left of each item.
        This is added after applying any -w or --wrap string.
    -r or --right defines a string that will be added to the right of each item.
        This is added after applying any -w or --wrap string.
```

### `show_last_exit_code.sh`

Function for outputting an indicator of the previous command's exit code.

This is a function I use in my command prompt to indicate the exit code of the previous command.
Basically, the first part of my PS1 value is `$( show_last_exit_code )`.
Depending on your shell, you might need to turn on some extra command prompt processing for that to work, though.

Example Usage:
```console
$ ( exit 1; ); show_last_exit_code
 ☠    1
```
and
```console
$ ( exit 0; ); show_last_exit_code
 ⭐️   0
```

You can't see it in this README, but when the exit code is zero, the background is green; otherwise it's red.
The `show_last_exit_code` also exits with the same code as the previous command.

### `string_repeat.sh`

Function/script for repeating a string a given number of times.

Example Usage:
```console
$ string_repeat Banana 3
BananaBananaBanana
```

The output does not contain an ending newline.

### `strip_colors.sh`

Function/script for removing color-code control sequences from a stream.

This is useful for stuff that outputs with terminal color codes, but you don't want them.

Example Usage:
```console
$ echo_color blue -- 'testing' | strip_colors | pbcopy
```

### `strip_final_newline.sh`

Function/script for removing the final newline from a stream.

This is most useful (for me) when sending stuff to pbcopy.
If the provided text does not have a newline at the end of the last line, nothing is changed.

Example Usage:
```console
$ cat foo.txt | strip_final_newline | pbcopy
```

### `tee_pbcopy.sh`

Function/script for outputting a stream, and also putting it into the clipboard.

Using `tee`, the input stream is sent to both stdout as well as the clipboard (using `pbcopy`).

Example Usage:
```console
$ cat foo.txt | tee_pbcopy
```

### `tee_strip_colors.sh`

Function/script for outputting a stream, and also stripping the color control sequences before appending it to a file.

The output that goes to stdout will still contain colors, but the color information will be removed as the file is created.

Example Usage:
```console
$ grep --color=always foo bar.txt | tee_strip_colors 'foo-lines-in-bar.txt'
```

### `tee_strip_colors_pbcopy.sh`

Function/script for outputting a stream, and also stripping the color control sequences before putting it into the clipboard.

The output that goes to stdout will stil contain colors, but the color information will be removed before being placed in the clipboard.

Example Usage:
```console
$ grep --color=always foo bar.txt | tee_strip_colors_pbcopy
```

### `to_date.sh`

Function/script for converting milliseconds since the epoch into a date.

Fractional seconds are allowed too, but negative numbers are not allowed.

Example Usage:
```console
$ to_date 1591340109000
2020-06-05 00:55:09 -0600 (MDT) Friday
```
or
```console
$ to_date 1591340109000.1
2020-06-05 00:55:09.0001 -0600 (MDT) Friday
```
Also available:
```console
$ to_date now
2020-06-05 00:56:31 -0600 (MDT) Friday
```

### `to_epoch.sh`

Function/script for converting a date into milliseconds since the epoch.

A date is required, and should be in `yyyy-MM-dd` format.

A time is optional and can be either `HH:mm`, `HH:mm:ss`, or `HH:mm:ss.ddd` format.
If supplied, it should immediately follow the date (with a space between them).
If not supplied, midnight is used, i.e. `00:00:00.000`
Fractional milliseconds can also be supplied simply by providing more digits after the decimal.

A timezone offset is also optional and should be in the format of `+HHmm` or `-HHmm`.
If a time is supplied, the timezone offset should immediately follow the time (with a space between them).
If the time is not supplied, the timezone offset should immediately follow the date (with a space between them).
If not supplied, your local system timezone is used, i.e. the results of `date '+%z'`.

Example Usage:
```console
$ to_epoch 2020-05-01 21:15:04.987 +0000
1588367704987
```
Also available:
```console
$ to_epoch now
1591341279000
```

### `ugly_json.sh`

Function/script for using jq to make json ugly (compact).

This is similar to just doing `jq --sort-keys -c '.' <filename>` except it's got some nice bonus features.

It can get the input from a file, a pipe, the clipboard, or even as a raw json string as an argument.
It can also output to a file, stdout, or the clipboard.
STDOUT output will always contain color though.
That's why the ability to output to a file was added.
File output, and clipboard output do not contain color info.
If the input is a file, and the output is a file, the output filename can be automatically generated from the input filename.

See `ugly_json --help` for more info.

Example Usage:
```console
$ ugly_json -- '{ "a" : "A" , "b" : "B" }'
{"a":"A","b":"B"}
```

The counterpart to this function is `pretty_json` (listed above).

