# SpicyLemon / bash_fun / deprecated
This directory contains files for stuff that has either been replaced, or is no longer maintained.

## Contents

* `add_to_filename.sh` - Script for adding text before the extension in filenames. Not really used anymore though.
* `check_system_log_timestamp_order.sh` - Old system-specific function for checking system log timestamps.
* `chrome_cors.sh` - Haven't used it in a couple years. I'm betting Google's changed the needed options since then.
* `file-helpers.sh` - Some old file formatting stuff for lists of entries.
  Most of it can now be handled by [re_line.sh](../generic/re_line.sh).
  The rest wasn't really very useful.
* `generic.sh` - This is an old monolithic dumping ground of environment functions.
  It has been replaced by the contents of the [bash_fun/generic/](../generic) directory.
  If you were previously sourcing this file, you should source `generic-setup.sh` instead.
* `get_all_system_logs.sh` - Old system-specific function for getting all system logs.
* `get_shell_type.sh` - Never really handy except for learning. Too specific, and it's better to just check for functionality.
* `git-helpers.sh` - Some functions that help interact with git.
  These were awkwardly named, though.
  They've been renamed to more easily convey their purpose.
  The file has been replaced by the contents of the [bash_fun/git/](../git) directory.
  If you were previously sourcing this file, you should source `git-setup.sh` instead.
  Then, maybe add some aliases to the new function names.
* `github.sh` - This was the beginnings of work on some GitHub API command line interaction.
  It hasn't been touched in quite some time, and what was already there wasn't really a viable solution anyway.
* `gitlab.sh` - This is the old monolithic version of the GitLab CLI interaction functions.
  It has been replaced with a much more managable and split out set of files in the [bash_fun/gitlab/](../gitlab) directory.
  If you were previously sourcing this file, you should source `gitlab-setup.sh` instead.
* `i_can.sh` - If this were a standard thing, it'd be useful, but for protability, I wasn't using it in any of my stuff. So I'm deprecating it.
  I did set up an alias for it though: `alias ican='command -v > /dev/null 2>&1'`
* `java_8_activate.sh` - Old system-specific command. The [java_sdk_switcher.sh](../gneric/java_sdk_switcher.sh) function is better.
* `java_8_deactivate.sh` - Old system-specific command. The [java_sdk_switcher.sh](../gneric/java_sdk_switcher.sh) function is better.
* `json_diff.sh` - The multidiff command can do this, and is more versatile.
* `jqq.sh` - Not needed because of herestrings. E.g. instead of `jqq "$foo" --sort-keys '.'` use `jq --sort-keys '.' <<< "$foo"`.
* `pretty_json.sh` - Not really handier than just typing out the jq stuff.
* `strip_colors.sh` - Replaced with an alias: `alias strip_colors='sed -E "s/$( printf "\033" )\[[[:digit:]]+(;[[:digit:]]+)*m//g"'`
* `text-helpers.sh` - Some old functions for converting lists of things into nicer formats.
  These are mostly replaced by the [re_line.sh](../generic/re_line.sh) function.
  Some other entries were moved into the [bash_fun/generic/](../generic) directory.
* `ugl_json.sh` - Not really handier than just typing out the jq stuff.

