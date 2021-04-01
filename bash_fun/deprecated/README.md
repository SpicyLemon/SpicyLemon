# SpicyLemon / bash_fun / deprecated
This directory contains files for stuff that has either been replaced, or is no longer maintained.

## Contents

* `file-helpers.sh` - Some old file formatting stuff for lists of entries.
  Most of it can now be handled by [re_line.sh](../generic/re_line.sh).
  The rest wasn't really very useful.
* `generic.sh` - This is an old monolithic dumping ground of environment functions.
  It has been replaced by the contents of the [bash_fun/generic/](../generic) directory.
  If you were previously sourcing this file, you should source `generic-setup.sh` instead.
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
* `jqq.sh` - Not needed because of herestrings. E.g. instead of `jqq "$foo" --sort-keys '.'` use `jq --sort-keys '.' <<< "$foo"`.
* `text-helpers.sh` - Some old functions for converting lists of things into nicer formats.
  These are mostly replaced by the [re_line.sh](../generic/re_line.sh) function.
  Some other entries were moved into the [bash_fun/generic/](../generic) directory.

