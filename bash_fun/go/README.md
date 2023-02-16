# SpicyLemon / bash_fun / go

This directory contains functions/scripts for doing various golang related things.

<hr>

## Table of Contents

* [Usage](#user-content-usage)
  * [Setup/Installation](#user-content-setupinstallation)
* [Directory Contents](#user-content-directory-contents)
  * [go-setup.sh](#user-content-go-setupsh)
  * [go_find_funcs_without_comments.sh](#user-content-go_find_funcs_without_commentssh)
  * [go_get_func.sh](#user-content-go_get_funcsh)
  * [go_list_funcs.sh](#user-content-go_list_funcssh)
  * [go_mod_fix.sh](#user-content-go_mod_fixsh)
  * [go_use_1.18.sh](#user-content-go_use_1_18sh)
  * [go_use_1.19.sh](#user-content-go_use_1_19sh)
  * [go_use_1.20.sh](#user-content-go_use_1_20sh)

## Usage

### Setup/Installation

1.  Copy this directory to a safe location on your computer.
    I personally, have a `~/.functions/` folder for such files and directories.
    So I've got a `~/.functions/go/` folder with all these files.
1.  In your environment setup file (e.g. `.bash_profile`), add a line to source the `go-setup.sh` file.
    For example, in mine, I have this line:
    ```bash
    source "$HOME/.functions/go/go-setup.sh"
    ```
    In order to add these functions to an already open environment, execute the same command.

If you need to troubleshoot the setup, you can add a `-v` flag when sourcing the setup file: `source go-setup.sh -v`.

## Directory Contents

### `go-setup.sh`

[go-setup.sh](go-setup.sh) - The file to source to source the rest of these files, importing the functions into your environment.

Use this commmand to utilize the file:
```console
$ source go-setup.sh
```

If you run into problems, you can use the `-v` option to get more information:
```console
$ source go-setup.sh -v
```



### `go_find_funcs_without_comments.sh`

[go_find_funcs_without_comments.sh](go_find_funcs_without_comments.sh) - Function/script for finding public functions that don't have comments.

Usage: `go_find_funcs_without_comments <file> [<file 2> ...]`

When a file is found containing a function without a leading comment, the filename is printed along with all functions without comments. This will only find public functions (starting with an upper-case letter).

The filenames can also be piped in. For example `find ... | go_find_funcs_without_comments <func name>`.



### `go_get_func.sh`

[go_get_func.sh](go_get_func.sh) - Function/script for getting a function from a go file.

Usage: `go_get_func <func name> <file> [<file 2> ...]`

All provided files will be searched for a function with the given name.
When one is found, the filename, leading function comment, and entire function will be printed.

The filenames can also be piped in. For example `find ... | go_get_func <func name>`.



### `go_list_funcs.sh`

[go_list_funcs.sh](go_list_funcs.sh) - Function/script for listing the functions in go files.

Usage: `go_list_funcs <files>`

```console
$ go_list_funcs --help
Usage: go_list_funcs <files>

Any number of files can be provided.

Coloring can be controlled with the following env vars:
    GLF_NO_COLOR   - Set to anything (other than an empty string) to disable output coloring.
    GLF_FILE_COLOR - The color to use for the filename. The default is 36 (cyan).
    GLF_FUNC_COLOR - The color to use for the text "func". The default is 90 (dark gray).
    GLF_RCVR_COLOR - The color to use for the function receiver. The default is 95 (light-magenta).
    GLF_NAME_COLOR - The color to use for the function name. The default is 1 (bold).
    GLF_COLORS     - Four comma separated color values for (in order):
                        the filename, "func", the receiver, the function name.
                     Specific color env vars (e.g. GLF_NAME_COLOR) take
                     precidence over an entry in GLF_COLORS.
                     The default is '36,90,95,1'
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



### `go_use_1.18.sh`

[go_use_1.18.sh](go_use_1.18.sh) - Function/script for switching the go binary to version 1.18.3.

This is specific to my system (a mac) that used brew to install go. It may not be very protable.

Usage: `go_use_1.18`

Example:
```console
> go_use_1.18
    Was: lrwxr-xr-x  1 danielwedul  admin  20 Feb 16 15:57 /opt/homebrew/bin/go -> /usr/local/go/bin/go
 Is Now: lrwxr-xr-x  1 danielwedul  admin  32 Feb 16 15:57 /opt/homebrew/bin/go -> ../Cellar/go@1.18/1.18.10/bin/go
```



### `go_use_1.19.sh`

[go_use_1.19.sh](go_use_1.19.sh) - Function/script for switching the go binary to version 1.19.

This is specific to my system (a mac) that used brew to install go. It may not be very protable.

Usage: `go_use_1.19`

Example:
```console
> go_use_1.19
    Was: lrwxr-xr-x  1 danielwedul  admin  32 Feb 16 15:57 /opt/homebrew/bin/go -> ../Cellar/go@1.18/1.18.10/bin/go
 Is Now: lrwxr-xr-x  1 danielwedul  admin  26 Feb 16 15:57 /opt/homebrew/bin/go -> ../Cellar/go/1.19.6/bin/go
```



### `go_use_1.20.sh`

[go_use_1.20.sh](go_use_1.20.sh) - Function/script for switching the go binary to version 1.20.

This is specific to my system (a mac) that used the go installer to install go. It may not be very protable.

Usage: `go_use_1.20`

Example:
```console
> go_use_1.20
    Was: lrwxr-xr-x  1 danielwedul  admin  26 Feb 16 15:57 /opt/homebrew/bin/go -> ../Cellar/go/1.19.6/bin/go
 Is Now: lrwxr-xr-x  1 danielwedul  admin  20 Feb 16 15:58 /opt/homebrew/bin/go -> /usr/local/go/bin/go
```

