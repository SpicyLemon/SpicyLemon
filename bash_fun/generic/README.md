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

* `escape_escapes.sh` - Function/script for escaping the ASCII escape character (octal `\033`, hex `\x1B`).
* `fp.sh` - Function/script for getting the full path to a file. Really, it just prepends the current directory to a string and cleans it up.
* `get_all_system_logs.sh` - Function/script that pulls all system logs into a single file and sorts the entries by the stamp.
* `get_shell_type.sh` - Function for telling which shell is being used. Currently only recognizes bash and zsh.
* `getlines.sh` - Function/script for getting specific lines from a file by line number.
* `hrr.sh` - Function/script that outputs a colorful horizontal rule in your terminal.
* `i_can.sh` - Function for testing whether or not a command is available in the environment.
* `java_8_activate.sh` - Function for setting the `JAVA_HOME` variable to the Java 8 JDK.
* `java_8_deactivate.sh` - Function for clearing the `JAVA_HOME` variable, going back to the system default.
* `join_str.sh` - Function/script for joining strings using a delimiter.
* `jqq.sh` - Function/script for running json contained in a variable through the `jq` program.
* `multi_line_replace.sh` - Function/script for replacing a matched area of a single line with a multi-line replacement.
* `pretty_json.sh` - Function/script for using jq to make json pretty.
* `print_args.sh` - Function/script for printing out the arguments passed into it.
* `ps_grep.sh` - Function/script for grepping through `ps aux` output.
* `re_line.sh` - Function/script for reformatting delimited and/or line-separated entries.
* `show_last_exit_code.sh` - Function for outputting an indicator of the previous command's exit code.
* `string_repeat.sh` - Function/script for repeating a string a given number of times.
* `strip_colors.sh` - Function/script for removing color-code control sequences from a stream.
* `strip_final_newline.sh` - Function/script for removing the final newline from a stream.
* `tee_pbcopy.sh` - Function/script for outputting a stream, and also putting it into the clipboard.
* `tee_strip_colors.sh` - Function/script for outputting a stream, and also stripping the color control sequences before appending it to a file.
* `tee_strip_colors_pbcopy.sh` - Function/script for outputting a stream, and also stripping the color control sequences before putting it into the clipboard.
* `to_date.sh` - Function/script for converting milliseconds since the epoch into a date.
* `to_epoch.sh` - Function/script for converting a date into milliseconds since the epoch.
* `ugly_json.sh` - Function/script for using jq to make json ugly (compact).

