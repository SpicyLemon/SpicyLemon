#!/bin/bash

run_capture_demo () {
    local mydir curdir
    mydir="$( dirname "$1" )"
    curdir="$( PWD )"
    cd "$mydir"
    go install spicylemon/libs/capturer
    go run demos/capturer.go
    cd "$curdir"
}

run_capture_demo "$0"
