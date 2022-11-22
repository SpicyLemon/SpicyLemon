# SpicyLemon / bash_fun / generic

This directory contains a whole bunch of functions/scripts for doing various things in bash.
In most cases, the scripts/functions should work in other shells too, but that hasn't been fully tested.

<hr>

## Table of Contents

* [Usage](#user-content-usage)
  * [Setup/Installation](#user-content-setupinstallation)
  * [Program/Function Requirements](#user-content-programfunction-requirements)
* [Directory Contents](#user-content-directory-contents)
  * [generic-setup.sh](#user-content-generic-setupsh)
  * [add.sh](#user-content-addsh)
  * [b2h.sh](#user-content-b2hsh)
  * [b642h.sh](#user-content-b642hsh)
  * [beepbeep.sh](#user-content-beepbeepsh)
  * [change_word.sh](#user-content-change_wordsh)
  * [cpm.sh](#user-content-cpmsh)
  * [echo_color.sh](#user-content-echo_colorsh)
  * [echo_do.sh](#user-content-echo_dosh)
  * [fp.sh](#user-content-fpsh)
  * [getlines.sh](#user-content-getlinessh)
  * [go_mod_fix.sh](#user-content-go_mod_fixsh)
  * [h2b64.sh](#user-content-h2b64sh)
  * [hrr.sh](#user-content-hrrsh)
  * [java_sdk_switcher.sh](#user-content-java_sdk_switchersh)
  * [join_str.sh](#user-content-join_strsh)
  * [list.sh](#user-content-listsh)
  * [max.sh](#user-content-maxsh)
  * [min.sh](#user-content-minsh)
  * [modulo.sh](#user-content-modulosh)
  * [multi_line_replace.sh](#user-content-multi_line_replacesh)
  * [multidiff.sh](#user-content-multidiffsh)
  * [multiply.sh](#user-content-multiplysh)
  * [palette_generators.sh](#user-content-palette_generatorssh)
  * [print_args.sh](#user-content-print_argssh)
  * [pvarn.sh](#user-content-print_pvarnsh)
  * [ps_grep.sh](#user-content-ps_grepsh)
  * [re_line.sh](#user-content-re_linesh)
  * [sdkman_fzf.sh](#user-content-sdkman_fzfsh)
  * [set_title.sh](#user-content-set_titlesh)
  * [show_last_exit_code.sh](#user-content-show_last_exit_codesh)
  * [show_palette.sh](#user-content-show_palettesh)
  * [string_repeat.sh](#user-content-string_repeatsh)
  * [tee_strip_colors.sh](#user-content-tee_strip_colorssh)
  * [timealert.sh](#user-content-timealertsh)
  * [to_date.sh](#user-content-to_datesh)
  * [to_epoch.sh](#user-content-to_epochsh)
  * [tryhard.sh](#user-content-tryhardsh)

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

    'cat'       'printf'  'head'  'tail'  'grep'
    'sed'       'awk'     'tr'    'tee'   'sort'
    'column'    'ps'      'seq'   'date'  'dirname'
    'basename'  'pwd'

## Directory Contents

### `generic-setup.sh`

[generic-setup.sh](generic-setup.sh) - The file to source to source the rest of these files, importing the functions into your environment.

Use this commmand to utilize the file:
```console
$ source generic-setup.sh
```

If you run into problems, you can use the `-v` option to get more information:
```console
$ source generic-setup.sh -v
```

This will also create several alias:
- `epoch`: Usage `epoch`. Outputs the epoch in milliseconds. But since date doesn't have that precision, it'll alwasy end in 000.
- `pvar`: Usage `pvar "$foo"`. Outputs a string with brackets around it and a newline after.
- `pevar`: Usage `pevar "$foo"`. Outputs a string in its shell-quoted (escaped) format with a newline after it.
- `strim`: Usage `<stuff> | strim`. Strips leading and trailing whitespace but keeps a newline at the very end.
- `strimr`: Usage `<stuff> | strimr`. Strips trailing (right) whitespace but keeps a newline at the very end.
- `striml`: Usage `<stuff> | striml`. Strips leading whitespace.
- `ican`: Usage `ican <command> && <do something with command>`. Checks if the `<command>` is available. Returns 0 if you can, 1 if you cannot.
- `strip_colors`: Usage `<stuff> | strip_colors`. Strips the color codes out of stuff.
- `strip_final_newline`: Usage `<stuff> | strip_final_newline`. Strips the newline off the very end (if there is one).
- `tee_pbcopy`: Usage `<stuff> | tee_pbcopy`. Outputs to stdout and also puts it in the clipboard.
- `tee_strip_colors_pbcopy`: Usage `<stuff> | tee_strip_colors_pbcopy`. Outputs to stdout and also puts the color stripped version into the clipboard.
- `escape_escapes`: Usage `<stuff> | escape_escapes`. Escapes the escape character 033 so that it's easier to see (and color codes can be seen).
- `fnl`: Usage `<stuff> | fnl`. Adds a final newline if there wasn't one already.
- `clearx': Usage `clearx`. Clears therminal screen and scrollback.

### `add.sh`

[add.sh](add.sh) - A function/script for adding up numbers.

Examples:
```console
$ add 1 2 3 4 5
15
$ seq 6 20 | add
195
$ seq 6 20 | add - 1 2 3 4 5
210
```

### `b2h.sh`

[b2h.sh](b2h.sh) - A function/script for converting byte values to human readable ones.

```console
$ b2h --help
Converts byte values to human readable values.

Usage: b2h [flags] <value1> [<value2> ...]
   or: <stuff that outputs values> | b2h [flags] [<values>]

    Values:
        Values must be positive numbers.
        Any fractional portions (a period followed by digits) will be ignored.
        Commas are okay.
        One or more spaces will separate numbers (even if provided as the same argument).
        Values have an upper limit based on your system.
            e.g. a 32-bit system has a max of 2,147,483,647 -> 1.99 GiB or 2.14 GB
                 a 64-bit system has a max of 9,223,372,036,854,775,807 -> 7.99 EiB or 9.22 EB
        Values will be processed in the order they are provided.

    Flags:
        --help -h
            Display help (and ignore everything else).
        --base-ten --base-10 --ten --10 -t
            Calculate using 1000 as the divisor instead of 1024.
        --base-two --base-2 --two --binary -2 -b
            Calculate using 1024 as the divisor.
            This is the default behavior.
            If both --base-ten and --base-two (or their aliases) are provided, whichever is last will be used.
        --verbose -v
            Include the original bytes value in the output.
        --stdin -s -
            Get values from stdin.
            If values are also provided as arguments, the ordering depends on where in the arguments this flag is first given.
            I.e. If this flag is before any provided values, the piped in values will be processed first.
                 If this flag is after all provided values, the piped in values will be processed last.
                 If this flag is between provided values, the values before it will be processed,
                    then the piped in values, followed by the values provided after.
            E.g. All of these commands will process the numbers 1, 2, 3, 4, and 5 in the same order.
                printf '1 2' | b2h - 3 4 5
                printf '4 5' | b2h 1 2 3 --stdin
                printf '2 3 4' | b2h 1 --pipe 5
                printf '2 3' | b2h 1 - 4 -s 5
                    (the second instance of the flag is just ignored)

```

### `b642h.sh`

[b642h.sh](b642h.sh) - A function/script for converting base64 encoded strings to hexadecimal.

```console
$ b642h --help
Converts base64 values to hex.

Usage: b642h <val1> [<val2>...]
   or: <stuff> | b642h

```

Example:
```console
$ b642h VxtfaZoHQ+CWlFa0Kgb4TQ== KGBYwiamQwqmVh5xD+IOOQ==
571b5f699a0743e0969456b42a06f84d
286058c226a6430aa6561e710fe20e39
```

### `beepbeep.sh`

[beepbeep.sh](beepbeep.sh) - Prints two bell characters .3 seconds apart.

If invoked as a function, the previous exit code is preserved.

Usage: `beepbeep`

### `change_word.sh`

[change_word.sh](change_word.sh) - A function/script for rudamentarily changing one string to another in a set of files.

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

### `cpm.sh`

[cpm.sh](cpm.sh) - A function/script for copying stuff to multiple places.

```console
$ cpm --help
cpm - Copies things to multiple places (cp multiple).

Usage: cpm [<flags for cp>] [--] source1 [source2 ... --] target1 [target2 ...]

The <flags for cp> must come first are are anything that start with a -.
If the first source file starts with a -, then put a -- before the source files.
If there are multiple sources, put a -- between the sources and targets.

You can also identify sources using --source <source> and targets using --target <target>.
Similarly, the --file <entry>, --dir <entry>, --entry <entry>, and --name <entry> flags all do the same thing:
  add the <entry> to either the sources or targets depending on it's position.

```

### `echo_color.sh`

[echo_color.sh](echo_color.sh) - A function/script for outputting command-line stuff with colors.

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

[echo_do.sh](echo_do.sh) - A function/script for outputting a command prior to executing it.

The provided command is printed in bold white, then executed.

Example Usage:
This would go through each .sh file in the current directory, and diff that file with the same-named file in `/some/other/dir`.
The diff command would show in bold white followed by the results of the diff command.
```console
$ for f in $( ls *.sh ); do echo_do diff "$f" "/some/other/dir/$f"; done
```

### `fp.sh`

[fp.sh](fp.sh) - Function/script for getting the full path to a file.

If the filename in question does not start with a slash, this will prepend your current directory onto it.
Then the path will be simplified (if it exists).
Finally, the full path to the file is printed.

If no files are provided, and fzf is available, you will be prompted to select the file(s) from the current directory.

Files can also be piped in by supplying a dash as an argument: `ls | fp -`.

If the path exists to be simplified, `pwd` is used to do so. If you want `-L` or `-P` provided to `pwd`, you can provide them to `fp` and they'll be passed on.

If there is only one path to print, and the `pbcopy` command is available, then the path will be put into your clipboard as well as printed to stdout.

Example: Get the full path to the README in the deprecated folder.
```console
$ pwd
/Users/spicylemon/git/SpicyLemon/bash_fun/generic
$ fp ../deprecated/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/deprecated/README.md - copied to clipboard.
```

Example: Get the full path to all the files in this folder that start with `"to_"`.
```console
$ pwd
/Users/spicylemon/git/SpicyLemon/bash_fun/generic
$ fp to_*
/Users/spicylemon/git/SpicyLemon/bash_fun/generic/to_date.sh
/Users/spicylemon/git/SpicyLemon/bash_fun/generic/to_epoch.sh
```

Example: Use find to find files, then get their full paths:
```console
$ pwd
/Users/spicylemon/git/SpicyLemon
$ find . -name README.md | fp -
/Users/spicylemon/git/SpicyLemon/README.md
/Users/spicylemon/git/SpicyLemon/js_fun/README.md
/Users/spicylemon/git/SpicyLemon/ticker/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/gitlab/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/figure/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/generic/README.md
/Users/spicylemon/git/SpicyLemon/bash_fun/deprecated/README.md
/Users/spicylemon/git/SpicyLemon/perl_fun/README.md
/Users/spicylemon/git/SpicyLemon/perl_fun/loan_calcs/README.md
```

### `getlines.sh`

[getlines.sh](getlines.sh) - Function/script for getting specific lines from a file by line number.

It can take in specific line numbers, line number ranges, and any combination of those.
A filename can also be provided.
If no filename is provided, it'll attempt to get input that is piped in.
Lines are printed in numerical order (as opposed to the order that they're provided as arguments).

Example Usage:
```console
$ getlines 86-90 11-15 fp.sh
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

### `go_mod_fix.sh`

[go_mod_fix.sh](go_mod_fix.sh) - Function/script for running go mod tidy; go mod vendor; go mod fix on all go.mod files under a directory.

Usage: `go_mod_fix [<root_dir>]`
If no `<root_dir>` is provided, `.` is used.

Example:
```console
> go_mod_fix
./go.mod: go mod tidy ... go mod vendor ... go mod verify ... all modules verified
./submodule/go.mod: go mod tidy ... go mod vendor ... go mod verify ... all modules verified
```

### `h2b64.sh`

[h2b64.sh](h2b64.sh) - Function/script for converting hexadecimal to base64.

```console
$ h2b64 --help
Converts hex values to base64.

Usage: h2b64 <val1> [<val2>...]
   or: <stuff> | h2b64
```

Example:
```console
$ h2b64 571b5f699a0743e0969456b42a06f84d 286058c226a6430aa6561e710fe20e39
VxtfaZoHQ+CWlFa0Kgb4TQ==
KGBYwiamQwqmVh5xD+IOOQ==
```

### `hrr.sh`

[hrr.sh](hrr.sh) - Function/script that outputs a colorful horizontal rule in your terminal.

![various hr command results](/bash_fun/generic/screenshots/hrr-example.png)

A message can be provided to include in the output too.
A random palette is chosen each time it's called.

If you source the `hrr.sh` file, the following functions become available:
- `hr`: Displays a single line without any padding added to the provided message (if provided).
- `hr1`: Displays a single line, adding a single space to each side of the provided message.
- `hr3`: Displays 3 lines containing a message on the middle one (if provided).
- `hrr`: Same as `hr3`, just here for historical (hysterical?) reasons.
- `hhr`: Same as `hr3`, just here for historical reasons.
- `hr5`: Displays 5 lines containing a message on the middle one (if provided).
- `hr7`: Displays 7 lines containing a message on the middle one (if provided).
- `hr9`: Displays 9 lines containing a message on the middle one (if provided).
- `hr11`: Displays 11 lines containing a message on the middle one (if provided).
- `pick_a_palette`: If a palette has not already been set, it will pick a random one.
- `what_palette_was_that`: Outputs the color codes of the palette most previously used.
- `show_all_palettes`: Loops through all palette options and uses hr to display a message with them.
- `test_palette`: Outputs each of the various hr heights using the provided palette and optional message.

The width is determined using `tput` to get your terminal width, or if that's not available, it defaults to `80`.

Example Usage:
```console
$ hr This is a test of hr
##################################################This is a test of hr##################################################
```

```console
$ hr1 This is a test of hr1
################################################ This is a test of hr1 #################################################
```

```console
$ hr3 This is a test of hr3
########################################################################################################################
###############################################  This is a test of hr3  ################################################
########################################################################################################################
```

```console
$ hr5 This is a test of hr5
########################################################################################################################
########################################################################################################################
##############################################   This is a test of hr5   ###############################################
########################################################################################################################
########################################################################################################################
```

```console
$ hr7 This is a test of hr7
########################################################################################################################
########################################################################################################################
################################################                       #################################################
##############################################   This is a test of hr7   ###############################################
################################################                       #################################################
########################################################################################################################
########################################################################################################################
```

```console
$ hr9 This is a test of hr9
########################################################################################################################
########################################################################################################################
########################################################################################################################
###############################################                         ################################################
#############################################    This is a test of hr9    ##############################################
###############################################                         ################################################
########################################################################################################################
########################################################################################################################
########################################################################################################################
```

```console
$ hr11 This is a test of hr11
########################################################################################################################
########################################################################################################################
########################################################################################################################
########################################################        ########################################################
##############################################                            ##############################################
############################################     This is a test of hr11     ############################################
##############################################                            ##############################################
########################################################        ########################################################
########################################################################################################################
########################################################################################################################
########################################################################################################################
```

### `java_sdk_switcher.sh`

[java_sdk_switcher.sh](java_sdk_switcher.sh) - Function for using `sdkman` to switch your java versions in your environment.

This is really only still around for historical reasons since it's just a wrapper for `sdkman_fzf use java _`.

It interacts with sdkman and uses fzf to let you select the java version you want to switch to.

### `join_str.sh`

[join_str.sh](join_str.sh) - Function/script for joining strings using a delimiter.

The first argument provided is treated as the delimiter.
All other arguments are then joined together into one line, with the delimiter between each entry.
A trailing newline character is *not* added.

Example Usage:
```console
$ join_str ', ' $( seq 1 10 )
1, 2, 3, 4, 5, 6, 7, 8, 9, 10
```

### `list.sh`

[list.sh](list.sh) - Function/script for listing directory contents.

Really, I just wanted an easy way to either list all directories or all files, and tell it whether or not to include hidden files, or even only look for hidden files.

```console
$ list --help
list - lists files and/or directories.

Usage: list [-f|--files|-F|--no-files] [-d|--dirs|-D|--no-dirs] [-h|--hidden|-H|--no-hidden|-I|--hidden-only]
            [-t|--dot|-T|--no-dot] [-0|--print0|-n|--newline] [-b|--base|-a|--absolute|-B|--no-base]
            [[--] <directory> [<directory2>...]]

    -f or --files        Include files in the output.
    -F or --no-files     Do not include files in the output.
        If multiple of -f -F --files --no-files are given, the last one is used.
        Default behavior depends on the presence of other flags.
            If -d -d --dirs or --no-dirs is provided, the default is -F or --no-files
            Otherwise, the default is -f or --files.

    -d or --dirs         Include directories in the output.
    -D or --no-dirs      Do not include directories in the output.
        If multiple of -d -D --dirs --no-dirs are given, the last one is used.
        Default behavior depends on the presense of other flags.
            If -f -F --files or --no-files is provided, the default is -D or --no-dirs.
            Otherwise, the default is -d or --dirs.

    -h or --hidden       Include hidden files and/or directories in the output.
    -H or --no-hidden    Do not include hidden files and/or directories in the ouptut.
    -I or --hidden-only  Only include hidden files and/or directoriesi in the output.
        If multiple of -h -H -I --hidden --no-hidden --hidden-only are given, the last one is used.
        If none of them are given, default behavior is -H or --no-hidden.

    -t or --dot          Include . as a directory for output.
    -T or --no-dot       Do not include . as a directory for output.
        If multiple of -t -T --dot or --no-dot are given, the last one is used.
        If none of them are given, default behavior is -T or --no-dot.

    -0 or --print0       Terminate each entry with a null character (handy with xargs -0).
    -n or --newline      Terminate each entry with a newline character.
        If multiple of -0 -n --print0 or --newline are given, the last one is used.
        If none of them are given, default behavior is -n or --newline.

    -b or --base        Include the base directory for each entry (as provided with the <directory> args).
    -a or --absolute    List the absolute path to each entry.
    -B or --no-base     Do not include the base directory for each entry.
        If multiple of -b -a -B --base --absolute or --no-base are given, the last one is used.
        If none of them are given, default behavior depends on the number of directories provided as arguments.
            If zero or one are provided, -B or --no-base is the default.
            If two or more are provided, -b or --base is the default.

    [--] <directory> [<directory2>...]
        Any number of directories can be provided as a base directory to list the contents of.
        Any arguments that do not start with a - are taken to be directories.
        Additionaly, any arguments provided after -- are taken to be directories.
        So if your directory of interest starts with a dash, you must provide it after a -- argument.
        If no directories are provided, the current directory (.) is used.

Default behavior: All of these behave the same:
    list
    list --files --dirs --no-hidden --no-dot --newline --no-base
    list -fdHTnB

Examples:
    Get just the (non-hidden) files in the current directory:
        list -f
        list -D
    Get just the (non-hidden) directories in the current directory:
        list -d
        list -F
    Get just the hidden directories in the home directory:
        list -Id ~
    Get ls long-format information on the hidden files in the /users/Spicylemon directory:
        list -If0 /users/SpicyLemon | xargs -0 ls -l
    Get ls long-format information on the entire contents of the foo/ and
    bar/ directories (in the current directory), sorted by date, newest at the bottom,
    and including the directories themselves:
        list -th0 foo bar | xargs -0 ls -ldtr

Exit codes:
    0   Normal execution with output.
    1   Normal execution but there was nothing to output.
    2   Invalid argument provided.
    3   Invalid directory provided.

```

### `max.sh`

[max.sh](max.sh) - Function/script for getting the max number.

```console
$ max 17 15 3 10 8 7 4 9 11 6 13 18 19 20 1 12 2 14 5 16
20
$ printf '17 15 3 10 8 7 4 9 11 6 13 18 19 20 1 12 2 14 5 16' | max
20
```

### `min.sh`

[min.sh](min.sh) - Function/script for getting the minimum number.

```console
$ min 17 15 3 10 8 7 4 9 11 6 13 18 19 20 1 12 2 14 5 16
1
$ printf '17 15 3 10 8 7 4 9 11 6 13 18 19 20 1 12 2 14 5 16' | min
1
```

### `modulo.sh`

[modulo.sh](modulo.sh) - Function/script for integer division with a remainder.

```console
$ modulo 8 3
8 / 3 = 2 r 2
```

### `multi_line_replace.sh`

[multi_line_replace.sh](multi_line_replace.sh) - Function/script for replacing a matched area of a single line with a multi-line replacement.

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

### `multidiff.sh`

[multidiff.sh](multidiff.sh) - Function/script for using diff to get comparisons of 2 or more files.

For example, say you have three files: `a.txt`, `b.txt`, and `c.txt`. You want to compare each of them to the others.

The command:
```console
$ multidiff a.txt b.txt c.txt
```
will do that.

It will basically do this:
```console
$ diff a.txt b.txt
$ diff a.txt c.txt
$ diff b.txt c.txt
```

Each file is assigned its own color for the command to make it easier to identify which file is which.

Json files can also be pre-processed using either `json_info` or `jq`. The diffs will then be done on the results of that pre-processing.

```console
$ multidiff --help
Gets differences between sets of files.

Usage: multidiff [[<diff args>] [--pre-process <pre-processor>] --] <file1> <file2> [<file3>...]

    <file1> <file2> [<file3>...] are the files to diff. Up to 12 can be supplied.
        Diffs are done between each possible pair of files.
        For example, with 3 files, there are 3 pairs: 1-2, 1-3, 2-3.
        With 4 files, you would end up with 6 pairs: 1-2, 1-3, 1-4, 2-3, 2-4, 3-4.

    If any arguments other than files are provided, the files must all follow a -- argument.

    <diff args> are any arguments that you want provided to each diff command.
    --pre-process <pre-processor> defines any pre-processing that should be done to each file before the diff.
        <pre-processor> values:
            none       This is the default. Do not do any pre-processing of the files.
            jq         Apply the command  jq --sort-keys '.' <file>  to each file and get the differences of the results.
            json_info  Apply the command  json_info -r -f <file>     to each file and get the differences of the results.
```

Text file example:
```console
$ printf 'common line\nfile a\ncommon line 2\n' > file-a.txt
$ printf 'common line\nfile b\ncommon line 2\n' > file-b.txt
$ printf 'common line\nfile c\ncommon line 2\nextra line' > file-c.txt
$ multidiff file-*.txt
diff file-a.txt file-b.txt
2c2
< file a
---
> file b

diff file-a.txt file-c.txt
2c2
< file a
---
> file c
3a4
> extra line
\ No newline at end of file

diff file-b.txt file-c.txt
2c2
< file b
---
> file c
3a4
> extra line
\ No newline at end of file
```

Json example:
```console
$ printf '{"filename":"file-a.json","elements":2}' > file-a.json
$ printf '{"filename":"file-b.json","elements":3,"extra":{}}' > file-b.json
$ printf '{"filename":"file-c.json","elements":2,"anarray":["str1","str2"]}' > file-c.json
$ multidiff --pre json_info -- file-*.json
diff file-a.json file-b.json
1,3c1,4
< . = object: 2 keys: ["filename","elements"]
< .filename = string: "file-a.json"
< .elements = number: 2
---
> . = object: 3 keys: ["filename","elements","extra"]
> .filename = string: "file-b.json"
> .elements = number: 3
> .extra = object: 0 keys: []

diff file-a.json file-c.json
1,2c1,2
< . = object: 2 keys: ["filename","elements"]
< .filename = string: "file-a.json"
---
> . = object: 3 keys: ["filename","elements","anarray"]
> .filename = string: "file-c.json"
3a4,6
> .anarray = array: 2 entries: string
> .anarray[0] = string: "str1"
> .anarray[1] = string: "str2"

diff file-b.json file-c.json
1,4c1,6
< . = object: 3 keys: ["filename","elements","extra"]
< .filename = string: "file-b.json"
< .elements = number: 3
< .extra = object: 0 keys: []
---
> . = object: 3 keys: ["filename","elements","anarray"]
> .filename = string: "file-c.json"
> .elements = number: 2
> .anarray = array: 2 entries: string
> .anarray[0] = string: "str1"
> .anarray[1] = string: "str2"
```

### `multiply.sh`

[multiply.sh](multiply.sh) - Function/script for multiplying numbers together.

```console
$ multiply 2 3 5 7 11
2310
$ printf '2 3 5 7 11' | multiply
2310
```

### `palette_generators.sh`

[palette_generators.sh](palette_generators.sh) - Some functions for generating sets of color codes to use in the terminal.

There are several functions in this file that revolve around the 256 color codes that can be used in escape codes to color text in your terminal, e.g. the "`100`" in `\033[38;5;100m` for text color or `\033[48;5;100m` for background color.

The color codes can generally be broken down into three groups:
- `0` to `15`: Some standard colors.
- `16` to `231`: A set of gradients giving more color options.
- `232` to `255`: A gradient from black to white.

The `echo_color.sh` file contains a `show_colors` function that can be used to display these.
```console
$ source echo_color.sh
$ show_colors --256
```
The `--256` flag adds this sections:
![show colors 256 screenshot](/bash_fun/generic/screenshots/show-colors-256-screenshot.png)

These functions focus on the 216 colors from `16` to `231`. They can be thought of as a 6x6x6 cube of colors where (0,0,0) is `16`, (5,0,0) is `21`, (0,5,0) is `46`, (0,0,5) is `51`, and (5,5,5) is `231`.

The file contains the following functions:
- `palette_generators`: Just outputs some information about the generators available.
- `palette_vector_generate`: Generates a six number set of color codes that are a straight line through the cube, wrapping if needed.
- `palette_vector_no_wrap`: Generates a six number set of color codes that are a straight line through the cube that doesn't include any wrapping.
- `palette_vector_random`: Picks random numbers and provides them to `palette_vector_generate`.

If you use the functions from `hrr.sh`, and have also sourced this file, then `palette_vector_no_wrap` is what's used by `pick_a_palette`. That file's `test_palette` function is also handy for viewing these.

```console
$ test_palette $( palette_vector_random )
```

The `show_palette` function from `show_palette.sh` is also handy for viewing these color vectors.
```console
$ show_palette $( palette_vector_no_wrap )
```

If you care to go crazy digging through my notes, they can be found here: [notes-on-palette-generation.txt](notes-on-palette-generation.txt)

### `print_args.sh`

[print_args.sh](print_args.sh) - Function/script for printing out the arguments passed into it.

This is a really simple thing that is usefuly when you want to see what arguments look like as they're passed into a function.

Example Usage:
```console
$ print_args turn down for='what'
Arguments received:
 1: [turn]
 2: [down]
 3: [for=what]
```

### `pvarn.sh`

[pvarn.sh](pvarn.sh) - Function for printing a variable from it's name.

Example Usage:
```console
$ foo='bar'
$ pvarn foo
foo: [bar]
```

### `ps_grep.sh`

[ps_grep.sh](ps_grep.sh) - Function/script for grepping through `ps aux` output.

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

[re_line.sh](re_line.sh) - Function/script for reformatting delimited and/or line-separated entries.

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

### `sdkman_fzf.sh`

[sdkman_fzf.sh](sdkman_fzf.sh) - Function wrapper for [sdkman](https://sdkman.io/) that adds fzf selection ability to most options.

You use `sdkman_fzf` the same way you would `sdkman` except if there's an argument you want to use fzf to select, give it as an underscore.

Examples:
- Select the version of java to install:
  ```console
  $ sdkman_fzf install java _
  ```
- Select the candidate(s) to list:
  ```console
  $ sdkman_fzf list _
  ```
- Select the version of ant to set as the default:
  ```console
  $ sdkman_fzf default ant _
  ```
- Select the version of java to use:
  ```console
  $ sdkman_fzf use java _
  ```
- Select a candidate and then version you want the home directory for:
  ```console
  $ sdkman_fzf home _ _
  ```

If none of the provided areguments are underscores, it's the same as the the `sdk` command provided by `sdkman`.

I even have it aliased:
```console
$ alias sdk='sdkman_fzf'
```

### `set_title.sh`

[set_title.sh](set_title.sh) - Function for setting the title of an iTerm window.

Allows for providing the title to set, or uses some defaults if nothing is provided.

If in a git repo, it'll default to the directory containing the root of the repo. Otherwise it'll just set it as the current directory.

### `show_last_exit_code.sh`

[show_last_exit_code.sh](show_last_exit_code.sh) - Function for outputting an indicator of the previous command's exit code.

This is a function I use in my command prompt to indicate the exit code of the previous command.
Basically, the first part of my PS1 value is `$( show_last_exit_code )`.
Depending on your shell, you might need to turn on some extra command prompt processing for that to work, though.

![show last exit code example](/bash_fun/generic/screenshots/show-last-exit-code-example.png)

Example Usage:
```console
$ ( exit 1; ); show_last_exit_code
 💀   1
```
and
```console
$ ( exit 0; ); show_last_exit_code
 ⭐️   0
```

The `show_last_exit_code` also exits with the same code as the previous command.

### `show_palette.sh`

[show_palette.sh](show_palette.sh) - Function/script for displaying various colors and combinations in the terminal.

```console
$ show_palette --help
Displays color palettes in your terminal.

Usage: show_palette [[-f] <fg col1> [<fg col2> ...]] [-b <bg col1> [<bg col2> ...]] [-a|--all] [-t <sample text>]
   or: show_palette [<pair1> [<pair2> ...]] [-a|--all] [-t <sample text>]

    The colors must be numbers between 0 and 255 inclusive.
        0 to 15 are some standard colors.
        16 to 231 are gradiented colors.
        231 to 255 are a black to white gradient.

    Desired colors can be provided in one of two ways:
        1: Single color entries.
            Any numbers provided first or after a -f flag are foreground colors.
            Any numbers provided after a -b flag are background colors.
            The -f and -b flags can be provided as many times as needed.
            If no foreground colors are provided, 7 is used.
            If no background colors are provided, 0 is used.
           E.g. show_palette 16 54 92 124 162 200 -b 27
           E.g. show_palette -b 200 -f 16 -b 162 -f 54 -b 124 -f 92
        2: Pairs of fg,bg entries.
            Each pair should be provided in a single argument.
            The foreground color should be first, then a comma, then the background color.
           E.g. show_palette '16,27' '54,27' '92,27' '124,27' '162,27' '200,27'
           E.g. show_palette '16,200' '54,162' '92,124'

    By default a number of lines are printed with the fg,bg color pair first (in the terminal default),
        followed by some sample text in that color combination.
        The 1st foreground color is paired with the 1st background color and printed.
        Then the 2nd foreground color is paired with the 2nd background color and printed.
        And so on.
        If an unequal number of foregrounds and backgrounds are provided,
        the smaller list cycles until the larger list has been completely shown.
        The sample text can be changed using the -t option.

    If the -a or --all flag is provided, a grid of all combinations of foreground and background colors is printed.
        Column and row headers are printed in the terminal default.
        The columns are the background colors.
        The rows are the foreground colors.
        The default sample text in this mode is the fg,bg pair.
        This can be changed to static supplied text using the -t option.
```

![show palette example](/bash_fun/generic/screenshots/show-palette-example.png)

### `string_repeat.sh`

[string_repeat.sh](string_repeat.sh) - Function/script for repeating a string a given number of times.

Example Usage:
```console
$ string_repeat Banana 3
BananaBananaBanana
```

The output does not contain an ending newline.

### `tee_strip_colors.sh`

[tee_strip_colors.sh](tee_strip_colors.sh) - Function/script for outputting a stream, and also stripping the color control sequences before appending it to a file.

The output that goes to stdout will still contain colors, but the color information will be removed as the file is created.

Supply the `-a` flag to append to the file instead of starting a new one.

```console
$ tee_strip_colors --help
Usage: tee_strip_colors [-a] <filename>
```

Example Usage:
```console
$ grep --color=always foo bar.txt | tee_strip_colors 'foo-lines-in-bar.txt'
```

### `timealert.sh`

[timealert.sh](timealert.sh) - Function/script that uses time to time a commmand and beeps twice when done.

Usage is the same as the `time` builtin.

```console
$ timealert git pull
```

The above is the same as the following command:

```console
$ time git pull; beepbeep
```


### `to_date.sh`

[to_date.sh](to_date.sh) - Function/script for converting milliseconds since the epoch into a date.

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

[to_epoch.sh](to_epoch.sh) - Function/script for converting a date into milliseconds since the epoch.

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

### `tryhard.sh`

[tryhard.sh](tryhard.sh) - Function/script for repeating a command until it succeeds.

Usage is similar to the `sudo` command.

```console
$ tryhard git pull
```

The provided command will be executed until it exits with code 0.
There will be a half-second of sleep between executions.
The time, a counter, and the command are printed before each execution.
Once it succeeds, it will beep twice.

