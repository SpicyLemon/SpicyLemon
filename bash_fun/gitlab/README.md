# SpicyLemon / bash_fun / gitlab

This directory contains files that define functions for interacting with GitLab's API from your command line.

<hr>

## Table of Contents

* [Usage](#user-content-usage)
  * [Setup/Installation](#user-content-setupinstallation)
  * [Invocation](#user-content-invocation)
  * [Customization](#user-content-customization)
* [Function Details](#user-content-function-details)
  * [gitlab](#user-content-gitlab)
  * [gitlab clone](#user-content-gitlab-clone)
  * [gitlab merge-requests](#user-content-gitlab-merge-requests)
  * [gitlab mr-search](#user-content-gitlab-mr-search)
  * [gitlab merged-mrs](#user-content-gitlab-merged-mrs)
  * [gitlab ignore-list](#user-content-gitlab-ignore-list)
  * [gitlab open](#user-content-gitlab-open)
  * [gitlab todo](#user-content-gitlab-todo)
  * [gitlab jobs](#user-content-gitlab-jobs)
  * [gitlab clean](#user-content-gitlab-clean)
* [Directory Contents](#user-content-directory-contents)
* [Disclaimer](#user-content-disclaimer)

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

### Invocation

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

## Function Details

### gitlab

The `gitlab` function provides an entry point to all the other functions.
It also provides an easy way to find out what functionality is available.

```console
$ gitlab --help
gitlab - This is the gateway to all GitLab CLI functions.

Usage:
    gitlab (help|merge-requests|mr-search|merged-mrs|ignore-list|clone|open|todo|jobs|clean) [command options]

    gitlab help
        Display this message.
        All commands also have -h or --help.

    gitlab clone [-b <dir>|--base-dir <dir>] [-f|--force] [-r|--refresh] [-h|--help] [-p <project name>|--project <project name>] [-s|--select-project]
        Easily clone repos from GitLab.
        Same as the glclone function.

    gitlab merge-requests [-r|--refresh] [-d|--deep] [-b|--bypass-ignore] [-i|--include-approved] [-m|--mine]
                          [-u|--update] [-q|--quiet] [-s|--select] [-o|--open-all] [-h|--help]
        Get information about merge requests.
        Same as the gmr function.

    gitlab mr-search <options>
        Do a search for merge requests with given criteria.
        Same as the gmrsearch function.

    gitlab merged-mrs <project> [-n <count>|--count <count>|--all] [-s|--select] [-q|--quiet]
        Lists MRs that have been merged.
        Same as the glmerged function.

    gitlab ignore-list [add|remove|update|clear|prune|status|list [<state(s)>]] [-h|--help]
        Manage a project ignore list that gmr -d will pay attention to.
        Same as the gmrignore function.

    gitlab open [-r [<repo>]|--repo [<repo>]|--select-repo] [-b [<branch]|--branch [<branch>]|--select-branch]
                [-d [<branch>]|--diff [<branch>]|--select-diff-branch] [-m|--mrs] [-p|--pipelines] [-q|--quiet] [-x|--do-not-open]
        Open various webpages of a GitLab repo.
        Same as the glopen function.

    gitlab todo [-s|--select] [-o|--open] [-m|--mark-as-done] [--mark-all-as-done] [-q|--quiet] [-h|--help]
        Get and manage your GitLab todo list.
        Same as the gtd function.

    gitlab jobs [-r <repo>|--repo <repo>] [-b <branch>|--branch <branch>|-a|--all-branches] [-q|--quiet] [-s|--select] [-o|--open]
                [-p <page count>|--page-count <page count>|-d|--deep] [-x|--no-refresh] [-t <type>|--type <type>|--all-types] [-h|--help]
        Get information about jobs in GitLab.
        Same as the gljobs function.

    gitlab clean [-v|--verbose] [-l|--list] [-h|--help]
        Cleans up environment variables storing GitLab information.
        Same as the glclean function.
```

### gitlab clone

The `gitlab clone` command calls the `glclone` function.

This function makes it easy to clone projects from gitlab.
If you already know the name of the project that you want, you can supply that as a parameter to the command.
It can also use fzf to display your available projects, allowing you to select any that you wish to clone.
Projects are cloned locally to either the directory defined by provided arguments, the `GITLAB_REPO_DIR` environment variable, or your current location.

```console
$ gitlab clone --help
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

### gitlab merge-requests

The `gitlab merge-requests` command calls the `gmr` function.

This function gets a list of merger requests that you are an approver on, but have not yet approved.
However, there's a bug in GitLab that causes some projects to not be included.
If you find that happening, you can use the `-d` or `--deep` option.
In that case, it's probably good to also use `gitlab ignore-list` to reduce the list of projects in the search.

```console
$ gitlab merge-requests --help
gmr: GitLab Merge Requests

Gets information about merge requests you are involved in.

Usage: gmr [-r|--refresh] [-d|--deep] [-b|--bypass-ignore] [-i|--include-approved] [-m|--mine]
           [-u|--update] [-q|--quiet] [-s|--select] [-o|--open-all] [-h|--help]

  With no options, if there is no previous results, new results are looked up.
  With no options, if there ARE previous results, those old results are displayed.
  The -r or --refresh option causes gmr to reach out to GitLab to get a current list of your MRs (the easy, but incomplete way).
  The -d or --deep option causes gmr to go through each project you can see to check for merge requests that request your approval.
        This will take longer, but might uncover some MRs that do not show up with the simple (-r) lookup.
        If supplied with the -r option, the -r option is ignored.
  The -b or --bypass-ignore option only makes sense along with the -d or --deep option.
        It will cause gmr to bypass the ignore list that you have setup using gmrignore.
  The -i or --include-approved flag tells gmr to also display mrs that you have already approved.
  The -m or --mine option lists MRs that you created.
  The -u or --update option causes gmr to go through the known lists of MRs and update them with respect to comments and approvals.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option makes gmr prompt you to select entries that will be opened in your browser.
        You can select multiple entries using the tab key. All selected entries will be opened in your browser.
  The -o or --open-all option causes all MRs to be opened in your browser.

Basically, the first time you run the  gmr  command, you will get a list of MRs (eventually).
After that (in the same terminal) running  gmr  again will display the previous results.
In order to update the list again, do a  gmr --refresh
```

### gitlab mr-search

The `gitlab mr-search` command calls the `gmrsearch` function.

This function provides easy access to the full set of merge request search options available.
Similar to `gitlab merge-requests`, though, sometimes it misses specific projects.
So `-d` or `--deep` is an option here too, and it similarly uses the ignore list managed through `gitlab ignore-list`.

```console
$ gitlab mr-search --help
gmrsearch: GitLab Merge Request Search

Search for Merge Requests based on certain criteria.

Usage: gmrsearch [-h|--help] [-d|--deep] [-b|--bypass-ignore] [-s|--select] [--json] [--use-last-results] [-v|--verbose]
                 [--order-by <field>|--order-by-created|--order-by-updated] [--sort <direction>|--asc|--desc]
                 [--state <mr state>|--opened|--closed|--locked|--merged]
                 [--scope <scope>|--created-by-me|--assigned-to-me|--scope-all]
                 [--created-after <start datetime>] [--created-before <end datetime>]
                 [--created-between <start datetime> <end datetime>] [--created-on <date>]
                 [--updated-after <start datetime>] [--updated-before <end datetime>]
                 [--updated-between <start datetime> <end datetime>] [--updated-on <date>]
                 [--search <text>] [--source-branch <branch>] [--target-branch <branch>] [--wip <yes/no>]
                 [--author <author>] [--author-id <user id>] [--author-username <username>]
                 [--assignee-id <user id>] [--approver-ids <user id list>] [--approved-by-ids <user id list>]
                 [--labels <labels>] [--milestone <milestone>] [--my-reaction-emoji <emoji>]
                 [--with-labels-details [<yes/no>]|--without-labels-details] [--view <view>|--view-simple|--view-normal]

Behavior:
  -h or --help              Show this usage information.
  -d or --deep              Cause the search to happen on each project instead of using the simple search.
                            This will take longer, but might uncover some MRs that do not show up with the simple search.
  -b or --bypass-ignore     This option only makes sense along with the -d or --deep option.
                            It will cause the deep search to bypass the ignore list and do the search on all available projects.
  -s or --select            Results will be presented for you to select. Selected results will be opened in your browser.
  --json                    Output the results as json instead of the normal formatted table.
  --use-last-results        Instead of actually doing a search, the previous search results are used.
                            This can be handy when combined with -s or --json.
  -v or --verbose           Output some extra messages to possibly help with debugging.

Result Control:
  --order-by <field>        Define the field that should be used to sort the results.
                            Valid values:  created_at  updated_at
                            Default value is  created_at
    --order-by-created        Shortcut for  --order-by created_at
    --order-by-updated        Shortcut for  --order-by updated_at
  --sort <direction>        Define the direction of the ordering
                            Valid values:  asc  desc
                            Default value is  desc
    --asc                     Shortcut for  --sort asc
    --desc                    Shortcut for  --sort desc

Search Criteria:
  --state <mr state>        Limit search to the provided state.
                            Valid states:  opened  closed  locked  merged
    --opened                  Shortcut for  --state opened
    --closed                  Shortcut for  --state closed
    --locked                  Shortcut for  --state locked
    --merged                  Shortcut for  --state merged
  --scope <scope>           Limit search to the provided scope.
                            Valid scopes:  created_by_me  assigned_to_me  all
    --created-by-me           Shortcut for  --scope created_by_me
    --assigned-to-me          Shortcut for  --scope assigned_to_me
    --scope-all               Shortcut for  --scope all
  --created-after <start datetime>
                            Limit search to MRs created on or after a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 00
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 00:00
                                YYYY-MM-DD          time is assumed to be 00:00:00
  --created-before <end datetime>
                            Limit search to MRs created on or before a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 59
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 59:59
                                YYYY-MM-DD          time is assumed to be 23:59:59
    --created-between <start datetime> <end datetime>
                              Shortcut for  --created-after <start datetime> --created-before <end datetime>
                              The datetimes must be in one of these formats:  YYYY-MM-DD hh:mm:ss  YYYY-MM-DD hh:mm  YYYY-MM-DD hh  YYYY-MM-DD
                              Both datetimes must use the same format.
                              The two datetimes can be in any order and they will be applied appropriately.
  --updated-after <start datetime>
                            Limit search to MRs updated on or after a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 00
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 00:00
                                YYYY-MM-DD          time is assumed to be 00:00:00
  --updated-before <end datetime>
                            Limit search to MRs updated on or before a date/time.
                            The datetime must be in one of these formats:
                                YYYY-MM-DD hh:mm:ss
                                YYYY-MM-DD hh:mm    seconds are assumed to be 59
                                YYYY-MM-DD hh       minutes and seconds are assumed to be 59:59
                                YYYY-MM-DD          time is assumed to be 23:59:59
    --updated-between <start datetime> <end datetime>
                              Shortcut for  --updated-after <start datetime> --updated-before <end datetime>
                              The datetimes must be in one of these formats:  YYYY-MM-DD hh:mm:ss  YYYY-MM-DD hh:mm  YYYY-MM-DD hh  YYYY-MM-DD
                              Both datetimes must use the same format.
                              The two datetimes can be in any order and they will be applied appropriately.
  --search <text>           Limit search to MRs with the provided text in either the title or description.
  --source-branch <branch>  Limit search to MRs with the provided branch as the source branch.
  --target-branch <branch>  Limit search to MRs with the provided branch as the target branch.
  --wip <yes/no>            Limit search based on WIP status.
                            Valid values:  yes  true  no  false
                            The values  yes  and  true  are equivalent.
                            The values  no  and  false  are equivalent.
  --author <author>         Limit search to a specific author.
                            If the parameter is all numbers, it is assumed to be a user id.
                            In that case, this is a shortcut for  --author-id <user id>
                            Otherwise, this is a shortcut for  --author-username <username>
  --author-id <user id>     Limit search to a specific author based on a user id.
  --author-username <username>
                            Limit search to a specific author based on a username.
  --assignee-id <user id>   Limit search to a specific approver based on a user id.
  --approver-ids <user id list>
                            Limit search to a MRs that have all of the given user ids listed as approvers.
                            The list should be comma delimited without spaces.
                            Max 5 user ids.
                            The parameter can also be either "None" or "Any".
                            A value of "None" will return MRs that do not have any approvers.
                            A value of "Any" will return MRs that have one or more approvers.
  --approved-by-ids <user id list>
                            Limit search to a MRs that have been approved by all of the given user ids.
                            The list should be comma separated without spaces.
                            Max 5 user ids.
                            A value of "None" will return MRs with no approvals.
                            A value of "Any" will return MRs with at least one approval (by anyone).
  --labels <labels>         Limit search to MRs matching the provided labels.
                            Labels should be comma separated without spaces.
                            A value of "None" will return MRs with no labels.
                            A value of "Any" will return MRs that have one or more labels.
  --milestone <milestone>   Limit search to specific milestones.
                            A value of "None" will return MRs with no milestones.
                            A value of "Any" will return MRs with any milestones.
  --my-reaction-emoji <emoji>
                            Limit search to those that you have given a certain reaction emoji to.
                            A value of "None" will return MRs that you have not reacted to.
                            A value of "Any" will return MRs that you have given any reaction emoji to.

Expert features:
  --with-labels-details <yes/no>
                            Add label details to the results.
                            This option only matters in conjunction with the --json flag.
                            The standard output from this utility will not contain the extra information.
                            Valid values:  yes  true  no  false
                            The values  yes  and  true  are equivalent.
                            The values  no  and  false  are equivalent.
                            If a <yes/no> value is not provided, then it will default to  true
  --without-labels-details  Shortcut for  --with-label_details false
  --view <view>             Set the type of view to return.
                            This option only matters when viewing the raw API results.
                            The standard output from this utility will not contain the extra information.
                            Valid values:  simple  normal
                            Default behavior is  --view normal
  --view-simple             Shortcut for  --view simple
  --view-normal             Shortcut for  --view normal
```

### gitlab merged-mrs

The `gitlab merged-mrs` command calls the `glmerged` function.

This function finds merged MRs for a project, sorted by merged date, with the most recent at the top.

```console
$ gitlab merged-mrs --help
glmerged: Looks up merged MRs for a GitLab repo.

glmerged <project> [-n <count>|--count <count>|--all] [-s|--select] [-q|--quiet]

  If the provided project uniquely identifies a GitLab repository you have access to, it will be used.
  If there are multiple matches, you will be prompted to select the one you want.

  The -n or --count option indicates that you want to get a specific number of most recent merge requests.
        Default is 20.
        Cannot be combined with --all.
  The --all option indicates you want all the merge requests.
        Cannot be combined with -n or --count.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option prompts you to select entries that will then be opened in your browser.
        Select multiple using the tab key.
```

### gitlab ignore-list

The `gitlab ignore-list` command calls the `gmrignore` function.

This function provides a way to ignore certain projects when doing "deep" searches for MRs.

```console
$ gitlab ignore-list --help
gmrignore: GitLab Merge Request Ignore (Projects)

Manages an ignore list for projects scanned by gmr.

gmrignore [add|remove|update|clear|prune|status|list [<state(s)>]] [-h|--help]

  Exactly one of these commands must be provided:
    add    - Display a list of projects that are not currently ignored and let you select ones to add.
             Selected entries will be added to the ignore list (so they won't be searched).
    remove - Display a list of projects that are currently ignored, and let you select ones to remove from that list.
             Selected entries will be removed from the ignore list (so that they'll be searched again).
    update - Display a list of all projects, let you select ones to become the new ignore list.
             Selected entries become the new ignore list.
    clear  - Clear the list of ignored projects.
    prune  - Get rid of the unknown projects.
                As projects are moved, deleted, or renamed in gitlab, sometimes their id changes.
                When that happens to a project that's ignored, the new project will not be ignored anymore and the
                old project id will still be in the ignore list.
                This command will clear out those old ids.
                WARNING:
                    Sometimes communication errors occur with GitLab and not all projects end up being retrieved.
                    When that happens, a large number of unknown entries will appear.
                    Once communication is fully restored, though, they'll show up again.
                    In these cases, you don't want to do a prune or else a lot of ignored projects will show up again.
    status - Output some consolidated information about what's being ignored.
    list   - Output repositories according to their state.
                This command can optionally take in one or more states.
                Output will then be limited to the provided states.
                    Valid <states> are:  ignored  shown  unknown  all
                If none are provided, all will be used.
```

### gitlab open

The `gitlab open` command calls the `glopen` function.

This function is useful for opening up various pages in your browser.

```console
$ gitlab open --help
glopen: GitLab Open

Opens up the webpage of a repo.

glopen [-r [<repo>]|--repo [<repo>]|--select-repo] [-b [<branch]|--branch [<branch>]|--select-branch]
       [-d [<branch>]|--diff [<branch>]|--select-diff-branch] [-m|--mrs] [-p|--pipelines] [-q|--quiet] [-x|--do-not-open]

  The -r <repo> or --repo <repo> option is a way to provide the desired repo.
    If the desired repo cannot be found, you will be prompted to select one.
    If one of these are supplied without a desired repo,
      then these options are just like --select-repo.
    If this option is not supplied, and you are in a git repo, that repo will be used.
    If this option is not supplied, and you are not in a git repo, you will be prompted to select one.
    Additionally, if the GitLab project cannot be located by either the repo you are in, or the
      provided repo, you will be prompted to select one.
  The --select-repo option causes glopen to prompt you to select a repo.
  The -b <branch> or --branch <branch> option indicates that you want to open a specific branch.
    If a branch name is supplied with this option, that branch name will be used.
    If a branch name is not supplied, and -b or --branch are supplied,
      and you are in a git repo, your current branch will be used.
    If a branch is not supplied, and you are not in a git repo, you will be prompted to select a branch.
    Note: If you are in repo x, but provide or select repo y (using -r --repo or --select-repo), and
      -b or --branch is provided without a branch name, the branch that you are currently in (in repo x)
      will be used in conjuction with the base url for repo y.
      This makes it easier to open your current branch in multiple repos.
  The --select-branch option cause glopen to prompt you to select a specific branch.
  The -d <branch> or --diff <branch> option indicates that you want to open a diff page (instead of a specific branch page).
    This option defines the "from" branch for the diff.  The "to" branch is defined by the -b or --branch option behavior.
    If -d or --diff is provided without a branch, the "from" branch will default to master.
        The "to" branch is whatever is defined by a -b, --branch, or --select-branch option.
        If none of those are supplied, the "to" branch will be determined as if the -b option were provided.
  The --select-diff-branch option causes glopen to prompt you to select a specific "from" branch.
    It also indicates that you want to open the diff page (instead of the main page).
    The "to" branch will be determined the same way it would be if the -d or --diff option is supplied.
  The -m or --mrs option will open the mrs page for the repo.
  The -p or --pipelines option will open the pipelines page for the repo.
  The -q or --quiet option suppresses normal terminal output.
  The -x or --do-not-open option will prevent the pages from being opened and only output the info.
    Technically, you can provide both -q and -x, but then nothing will really happen.
```

### gitlab todo

The `gitlab todo` command calls the `gtd` function.

This function allows you to show and interact with your ToDo list.

```console
$ gitlab todo --help
gtd: GitLab ToDo List

Gets your GitLab TODO list.
You must create an api token from your profile in GitLab first. See: https://gitlab.com/profile/personal_access_tokens
Then, you must set the token value as the GITLAB_PRIVATE_TOKEN environment variable in your terminal (e.g. in .bash_profile)

Usage: gtd [-s|--select] [-o|--open] [-m|--mark-as-done] [--mark-all-as-done] [-q|--quiet] [-h|--help]

  With no options, your todo list is looked up and displayed.
  The -s or --select option will prompt you to choose entries.
        Selected entries are opened in your browser.
        This does not mark the entry as done, but some actions on the page might.
        You can select multiple entries using the tab key.
  The -o or --open option will cause your main todo page to be opened in your browser.
  The -m or --mark-as-done option will prompt you to choose entries (like with the -s option) and they will be marked as done.
        You can select multiple entries using the tab key.
        All selected entries will be marked as done.
  The -q or --quiet option will prevent unneeded output to stdout.
  The --mark-all-as-done option will mark all entries as done for you.
```

### gitlab jobs

The `gitlab jobs` command calls the `gljobs` function.

This function allows you to get information on jobs for a repo and branch.

```console
$ gitlab jobs --help
gljobs: GitLab Jobs

Get info about jobs in GitLab.

Usage: gljobs [-r <repo>|--repo <repo>] [-b <branch>|--branch <branch>|-a|--all-branches] [-q|--quiet] [-s|--select] [-o|--open]
              [-p <page count>|--page-count <page count>|-d|--deep] [-x|--no-refresh] [-t <type>|--type <type>|--all-types] [-h|--help]

  By default, if you are in a git repo, that will be used as the repo, and your current branch will be used as the branch.
  Also, by default, only the first page (100 jobs) of most recent jobs are retrieved for the repo.

  The -r <repo> or --repo <repo> option allows you to provide the repo instead of using the default.
  The -b <branch> or --branch <branch> option allows you to provide the branch instead of using the default.
        Cannot be used with -a or --all-branches.
  The -a or --all-branches option will display all branches.
        Cannot be used with a -b or --branch option.
  The -q or --quiet option suppresses normal terminal output. If used with -s, the selection page will still be displayed.
  The -s or --select option prompts you to select entries that will then be opened in your browser.
        Select multiple using the tab key.
  The -o or --open option will cause the first result to automatically be opened in your browser.
  The -p <page count> or --page-count <page count> option generally defines how far back in time to look.
        By default, only the first page of results (100) is retrieved across all branches for your repo.
        This option gives you a way to retrieve more jobs before filtering for your branch (or not filtering if you used -a).
        Cannot be used with -d or --deep.
  The -d or --deep option will retrieve all jobs for the repo.
        Cannot be used with the -p or --page-count option.
  The -x or --no-refresh option prevents a new lookup from happening and just displays the last results retrieved.
        Can only be combined with the -s and/or -q flags.
  The -t or --type option allows you to filter on build type.
        The list of jobs will be filtered to only include the supplied type.
        Common types are "build" "client" and "sdlc"
        If the provided type starts with a ~ then filtering will be to remove the supplied type.
        By default, there is a filter type of "~sdlc".
  The --all-types option disables the type filter.
        This is the same as -t "".
```

### gitlab clean

The `gitlab clean` command calls the `glclean` function.

This function cleans up the environment variables and temp files that are created and populated by the other functions.
This can be useful if you need to refresh some cached information such as your projects list.

```console
$ gitlab clean --help
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

## Directory Contents

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

## Disclaimer

All of this was developed and tested on a Mac in Bash.
Some light testing was also done in ZSH.

