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

