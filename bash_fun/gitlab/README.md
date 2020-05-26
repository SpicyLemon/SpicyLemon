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

### gitlab

The `gitlab` function provides an entry point to all the other functions.
It also provides an easy way to find out what functionality is available.

```bash
> gitlab --help
```
```
gitlab - This is a gateway to all GitLab functions.

Usage:
    gitlab (help|merge-requests|mr-search|merged-mrs|ignore-list|clone|open|todo|jobs|clean) [command options]

    gitlab help
        Display this message.
        All commands also have -h or --help.

    gitlab merge-requests [-r|--refresh] [-d|--deep] [-b|--bypass-ignore] [-i|--include-approved] [-m|--mine]
                          [-u|--update] [-q|--quiet] [-s|--select] [-o|--open-all] [-h|--help]
        Get information about merge requests.
        Same as the [1;37mgmr[0m function.

    gitlab mr-search <options>
        Do a search for merge requests with given criteria.
        Same as the [1;37mgmrsearch[0m function.

    gitlab merged-mrs <project> [-n <count>|--count <count>|--all] [-s|--select] [-q|--quiet]
        Lists MRs that have been merged.
        Same as the [1;37mglmerged[0m function.

    gitlab ignore-list [add|remove|update|clear|prune|status|list [<state(s)>]] [-h|--help]
        Manage a project ignore list that gmr -d will pay attention to.
        Same as the [1;37mgmrignore[0m function.

    gitlab clone [-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh] [-h|--help] [-p <project name>|--project <project name>] [-s|--select-project]
        Easily clone repos from GitLab.
        Same as the [1;37mglclone[0m function.

    gitlab open [-r [<repo>]|--repo [<repo>]|--select-repo] [-b [<branch]|--branch [<branch>]|--select-branch]
                [-d [<branch>]|--diff [<branch>]|--select-diff-branch] [-m|--mrs] [-p|--pipelines] [-q|--quiet] [-x|--do-not-open]
        Open various webpages of a GitLab repo.
        Same as the [1;37mglopen[0m function.

    gitlab todo [-s|--select] [-o|--open] [-m|--mark-as-done] [--mark-all-as-done] [-q|--quiet] [-h|--help]
        Get and manage your GitLab todo list.
        Same as the [1;37mgtd[0m function.

    gitlab jobs [-r <repo>|--repo <repo>] [-b <branch>|--branch <branch>|-a|--all-branches] [-q|--quiet] [-s|--select] [-o|--open]
                [-p <page count>|--page-count <page count>|-d|--deep] [-x|--no-refresh] [-t <type>|--type <type>|--all-types] [-h|--help]
        Get information about jobs in GitLab.
        Same as the [1;37mgljobs[0m function.

    gitlab clean [-v|--verbose] [-l|--list] [-h|--help]
        Cleans up environment variables storing GitLab information.
        Same as the [1;37mglclean[0m function.
```

### gitlab clone

The `gitlab clone` command calls the `glclone` function.

This function makes it easy to clone one or more projects from gitlab.
When called, all available projects are displayed using fzf where you can select one or more entries.
All selected entries are then cloned locally to either the directory defined by providing arguments, or else the `GITLAB_REPO_DIR` environment variable.

```bash
> gitlab clone --help
glclone: GitLab Clone

This will look up all the projects you have access to in GitLab, and provide a way for you to select one or more to clone.

If you set the GITLAB_REPO_DIR environment variable to you root git directory,
new repos will automatically go into that directory regardless of where you are when running the command.
If that variable is not set, and no -b or --base-dir parameter is provided, the current directory is used.

Usage: glclone [-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh] [-h|--help] [-p <project name>|--project <project name>] [-s|--select-project]

  The -b <dir> or --base-dir <dir> option will designate the directory to create your repo in.
        Providing this option overrides the default setting from the GITLAB_REPO_DIR.
  The -f or --force option will allow cloning into directories already under a git repo.
  The -r or --refresh option will cause your projects to be reloaded.
  The -p or --project option will allow you to supply the project name you are interested in.
    If the provided project name cannot be found, it will be used as an initial query,
    and you will be prompted to select the project.
    Multiple projects can be provided in the following ways:
        -p project1 -p project2
        -p 'project3 project4'
        -p project5 project6
    Additionallly, the -p or --project option can be omitted, and leftover parameters
    will be treated as the projects you are interested in.
        For example:
            glclone project7 project8
        Is the same as
            glclone -p project7 -p project8
    If no project name is provided after this option, it will be treated the same as -s or --select-projects
  The -s or --select-projects option forces glclone to prompt you to select projects.
      This is only needed if you are supplying projects to clone (with -p or --project),
      but also want to select others.
```

### gitlab clean

The `gitlab clean` command calls the `glclean` function.

This function cleans up the environment variables and temp files that are created and populated by the other functions.
This can be useful if you need to refresh some cached information such as your projects list.

```bash
> gitlab clean --help
glclean: GitLab Clean

Cleans up all the persistant variables used by the functions in this file.
Use this when you want a fresh start with respects to the data these GitLab functions use.

This will NOT affect your GITLAB_PRIVATE_TOKEN variable.

The following variables will be removed:
    GITLAB_USER_INFO         GITLAB_USER_ID     GITLAB_USERNAME         GITLAB_PROJECTS
    GITLAB_MRS               GITLAB_MRS_TODO    GITLAB_MRS_BY_ME        GITLAB_TODOS
    GITLAB_JOBS              GITLAB_MERGED_MRS  GITLAB_MERGED_MRS_REPO  GITLAB_MRS_SEARCH_RESULTS
    GITLAB_MRS_DEEP_RESULTS

Usage: glclean [-v|--verbose] [-l|--list] [-h|--help]

  The -v or --verbose option will output the values of each variable before being deleted.
  The -l or --list option will just show the variable names without deleting them.
    Combined with the -v command, the contents of the variables will also be displayed.
```

## Disclaimer

All of this was developed and tested on a Mac in Bash.
Some light testing was also done in ZSH.

