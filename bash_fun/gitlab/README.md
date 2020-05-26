# SpicyLemon / bash_fun / gitlab
This directory contains files that define functions for interacting with GitLab's API from your command line.

## Contents

* `gitlab.sh` - Contains the main `gitlab` function that can access the rest of these functions.
* `gl-core.sh` - Contains all generic/core functions that actually do the work.
* `glclean.sh` - Contains the `glclean` function, used to clean up environment variables used by these functions.
* `glclone.sh` - Contains the `glclone` function, used to clone GitLab repos.
* `gljobs.sh` - Contains the `gljobs` function, used to look up ci/cd job information.
* `glmerged.sh` - Contains the `glmerged` function, used to look up merged MRs.
* `glopen.sh` - Contains the `glopen` function, used to open certain GitLab pages in your browser.
* `gmr.sh` - Contains the `gmr` function, used to find MRs for you.
* `gmrignore.sh` - Contains the `gmrignore` function, used to manage a project ignore list.
* `gmrsearch.sh` - Contains the `gmrsearch` function, used to interact with the GitLab MR search api.
* `gtd.sh` - Contains the `gtd` function, used to interact with your GitLab ToDo list.

## Usage

### Setup/Installation

#### Create your GitLab API token

In order to interact with GitLab through their api, you will need an access token.

1.  Log into GitLab.
1.  Go to your personal settings page then to the "Access Tokens" page (e.g `https://gitlab.com/profile/personal_access_tokens`)
1.  Create a token with the `api` scope.
1.  In your terminal environment, set the `GITLAB_PRIVATE_TOKEN` environment variable to the value of that token.
    For example, you could put `GITLAB_PRIVATE_TOKEN=123abcABC456-98ZzYy7` in your `.bash_profile` file (or similar)
    so that it's set every time you open a terminal (use your own actual token of course).

#### Add these functions to your environment

1.  Copy the `gitlab/` directory and its contents to a safe place on your system.
    I personally, have a `~/.functions/` folder for such files and directories.
    So I've got a `~/.functions/gitlab/` folder with all these files.
1.  Copy the [fzf_wrapper.sh](../fzf_wrapper.sh) file to either the same directory as the `gitlab/` directory, or into the `gitlab/` directory itself.
1.  Copy the [curl_link_header.sh](../curl_link_header.sh) file to either the same directory as the `gitlab/` directory, or into the `gitlab/` directory itself.
1.  In your environment setup file (e.g. `.bash_profile`), add a line to source the `gitlab-setup.sh` file.
    For example, in mine, I have this line:
    ```bash
    source "$HOME/.functions/gitlab/gitlab-setup.sh"
    ```
    In order to add these functions to an already open environment, execute the same command.

If you need to troubleshoot the setup, you can add a `-v` flag when sourcing the setup file: `source gitlab-setup.sh -v`.

#### Program/Function Requirements

These GitLab functions depend on some external programs/functions.
Availability of the programs/functions is checked when `gitlab-setup.sh` is sourced.

These functions are looked for, and if not found, the file containing them is the looked for and sourced if possible:
* `fzf_wrapper` - Adds column support to `fzf`. See https://github.com/SpicyLemon/SpicyLemon/blob/master/bash_fun/fzf_wrapper.sh
* `curl_link_header` - Adds link header processing to `curl`. See https://github.com/SpicyLemon/SpicyLemon/blob/master/bash_fun/curl_link_header.sh

These programs are required, and don't usually come pre-installed:
* `jq` - Json processor. See https://github.com/stedolan/jq
* `fzf` - Fuzzy finder. See https://github.com/junegunn/fzf
* `git` - The stupid content tracker. https://git-scm.com/book/en/v2/Getting-Started-Installing-Git

These programs are required, and are almost always available already:
* `awk` - Pattern scanning and processing.
* `sed` - Stream editor.
* `grep` - Pattern search.
* `curl` - Url transfer utility.

### Invocation:

The main function added to your environment is `gitlab`.
You can use it to access any other pieces of functionality in here.

To find out more:
```bash
gitlab --help
```

The `gitlab` function is just a wrapper for the other main functions.
For example, `gitlab open <options>` is the same as `glopen <options>`.

All main functions have `-h` and `--help` available too.

For example:
```bash
gitlab clone --help
```

### Customization

The following environment variables can be defined:
* `GITLAB_REPO_DIR` -
  The directory where your GitLab repositories are to be stored.
  This should be absolute, (starting with a `/`), but it should not end with a `/`.
  If not defined, functions that look for it will require it to be provided as input.
* `GITLAB_CONFIG_DIR` -
  The directory where you'd like to store some configuration information used in these functions.
  This should be absolute, (starting with a `/`), but it should not end with a `/`.
  If not defined, then, if `HOME` is defined, `$HOME/.config/gitlab` will be used.
  If `HOME` is not defined, then, if `GITLAB_REPO_DIR` is defined, `$GITLAB_REPO_DIR/.gitlab_config` will be used.
  If `GITLAB_REPO_DIR` is not defined either, then any functions that use configuration information will be unavailable.
  If a config dir can be determined, but it doesn't exist yet, it will be created automatically when needed.
* `GITLAB_TEMP_DIR` -
  The temporary directory you'd like to use for some random file storage.
  This should be absolute, (starting with a `/`), but it should not end with a `/`.
  If not defined, `/tmp/gitlab` will be used.
  If the directory does not exist, it will be created automatically when needed.
* `GITLAB_PROJECTS_MAX_AGE` -
  The max age that the projects list can be before it's considered out-of-date.
  Format is `<number>[smhdw]` where `s` -> seconds, `m` -> minutes, `h` -> hours, `d` -> days, `w` -> weeks.
  See `man find` in the `-atime` section for more info.
  Do not include a leading `+` or `-`.
  If not defined, the default is `23h`.

## Main Functions

TODO

## Disclaimer

All of this was developed and tested on a Mac in Bash.
Some light testing was also done in ZSH.

