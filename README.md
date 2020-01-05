# SpicyLemon
Dumping ground for fun random stuff that I've put together.

## Contents

* `bash_fun` - Scripts and functions for doing stuff in bash.
* `js_fun` - Stuff relating to javascript.

## Favorites

### [bootRun.sh](bash_fun/bootRun.sh)

This script makes it easier to run Spring Boot programs, through gradle, with command-line arguments.

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

### [gitlab.sh](bash_fun/gitlab.sh)

This file houses a whole suite of functions for interacting with GitLab through the command line.

Recommended usage is to source it to add the functions to your environment, but it can also be run as a script.
The downside to running it as a script is that it doesn't cache as much that way.

See `gitlab --help` or `./gitlab.sh --help` for more info.

### [calculation-template.html](js_fun/calculation-template.html)

This is an HTML page with a bunch of template Javascript on it for doing calculation-heavy javascript stuff.

There are two script sections to it. The top one becomes a web worker that is called by the bottom one.
This allows you to put all the input, ouput, parsing and calculation pieces in one page while also making the page remain responsive while the calculation is running.

The file has a comment at the top with details on usage.

### [echo_do](https://github.com/SpicyLemon/SpicyLemon/blob/200e222352378578c602039226cdead87b3ba78c/bash_fun/generic.sh#L66)

The `echo_do` bash function will echo a command then execute it.

Features:
1.  The command being executed is printed in bright white before it is executed.
    It might be slightly different than the command you provided, but you should still be able to copy/paste it to run it again.
1.  Command output still goes to your console as it happens.
1.  The exit code of your command is maintained.
    For example, if `generic command` returns with an exit code of 3, then `echo_do generic command` will also have an exit code of 3.
1.  Simple commands can be provided normally but with `echo_do` as the first part of the command.
    For example: `echo_do git pull`.
1.  More complex commands can be provided as a string.
    For example: `echo_do '( echo "stdout"; echo "stderr" >&2; return 3 )'`
1.  Various aspects of the command and results end up stored in the following environment variables:
    * `ECHO_DO_CMD_PARTS` - An array containing each element of the command executed.
    * `ECHO_DO_CMD_STR` - A string containing the command as it appeared in output before being executed.
    * `ECHO_DO_STDOUT` - A string with just the stdout output of the command.
    * `ECHO_DO_STDERR` - A string with just the stderr output of the command.
    * `ECHO_DO_STDALL` - A string with both stdout and stderr output of the command (should match what ends up on your screen).
    * `ECHO_DO_EXIT_CODE` - The exit code that your command produced.
      This is the same as `$?` except that it won't change until your next `echo_do`.
1.  If no command is provided, `echo_do` will have an exit code of 124, and none of the environment variables will be set.

Current drawbacks:
1.  Temporary files are used to store stdout, stderr, and combined stdout/stderr while the command is running.
    The contents of the files are then pulled into the appropriatel environment variables and deleted.
    The `mktemp` command is used to create these files, and only the current user should have read or write access.
    But still, for a bit, they exist as files.
1.  Setting variables can be tricky with `echo_do`.
    Some shells will try to be helpful and alter your command as it comes into `echo_do` as parameters.
    For example, the command `echo_do foo='bar baz'` will end up being seen as "foo=bar baz".
    Then, when trying to execute it, it'll get confused when it gets to "baz".
    To prevent this, you can send the command in a string, e.g. `echo_do "foo='bar baz'"`.
1.  Putting `echo_do` in a piped command probably won't work as expected.
    If echo_do is part of the first command, the printed commmand will be part of the output, which probably isn't what you want.
    If echo_do is in a piped part of the command, the environment variables might not get properly set (due to different shell behaviors).
    Also, if echo_do is in a piped part of the command, the provided command is what will be receiving the incoming stream.

### [fzf_wrapper.sh](bash_fun/fzf_wrapper.sh)

The primary purpose of this file is to define the `__fzf_wrapper` function.
This function adds a `--to-columns` option to [fzf](https://github.com/junegunn/fzf).
When `--to-columns` is supplied, the string defined by `-d` or `--delimiter` becomes the string given to the `column` command.

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
As such, it should be safe to `alias fzf=__fzf_wrapper` and not notice any difference except the availability of the `--to-columns` option.

