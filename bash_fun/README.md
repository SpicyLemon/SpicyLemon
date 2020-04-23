# SpicyLemon / bash_fun
This directory contains files and scripts for doing things on a bash command-line.

## Contents

* `bootRun.sh` - A bash script that makes it easier to run the gradle bootRun task with supplied arguments. File is executable.
* `capture_cmd.sh` - A bash file with a function for capturing the output of commands by stdout, stderr, and combined. File should be sourced.
* `curl_link_header.sh` - A bash file with a function for using curl, and looking at response headers, in order to get all pages of a result. File should be sourced.
* `/deprecated` - A home for stuff that is no longer maintained, or has been replaced.
* `file-helpers.sh` - A bash file with some functions for doing things with files. File should be sourced, not executed.
* `fzf_wrapper.sh` - A bash function/script that adds some functionality to fzf.
* `generic.sh` - A bash file with a bunch of generic functions for doing bash stuff. File should be sourced, not executed.
* `git-helpers.sh` - A bash file with a bunch of functions for making git life easier. File should be sourced, not executed.
* `github.sh` - The beginnings of some functions for interacting with GitHub. File should be sourced, not executed.
* `/gitlab` - A directory to hold the files containing functionality for interacting with GitLab.
* `gitlab-setup.sh` - The entry point to the GitLab functionality. Source this in order to add the GitLab functions.
* `mysql_runner.sh` - A bash script for running stuff on a mysql database with a bunch of defaults handy for me.
* `psql_runner.sh` - A bash script for running stuff on a postgresql database with a bunch of defaults handy for me.
* `sagemaker.sh` - A function for making calls to sagemaker. Only tested on one endpoint, though, so your mileage may vary. File is executable.
* `text-helpers.sh` - A bash file to define a bunch of text manipulation helper functions. File should be sourced, not executed.

