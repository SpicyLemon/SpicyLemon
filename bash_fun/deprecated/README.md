# SpicyLemon / bash_fun / deprecated
This directory contains files for stuff that has either been replaced, or is no longer maintained.

## Contents

* `gitlab.sh` - This is the old monolithic version of the GitLab CLI interaction functions.
  It has been replaced with a much more managable and split out set of files in the `bash_fun/gitlab` directory.
  If you were previously sourcing this file, you will want to change to source `gitlab-setup.sh` instead.
* `generic.sh` - This is an old monolithic dumping ground of environment functions.
  It has been replaced by the contents of the `bash_fun/generic` directory.
  If you were previously sourcing this file, you will want to change to source `generic-setup.sh` instead.
* `file-helpers.sh` - Some old file formatting stuff for lists of entries. This can now easily be handled by `re_line.sh`.
* `text-helpers.sh` - Some old functions for converting lists of things into nicer formats. These are all replaced by the `re_line.sh` function.

