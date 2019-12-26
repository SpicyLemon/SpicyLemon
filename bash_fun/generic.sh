#!/bin/bash
# This file contains generic functions for helping do random things that I often need.
# File contents:
#   echo_do  -------------------------> Outputs a command in bright white, then executes it.
#   echo_do_ln  ----------------------> Outputs a command in bright white, then executes it and adds an extra newline to the end.
#   get_shell_type  ------------------> Gets the type of shell you're in, either "zsh" "bash" or else the process running your shell
#   kill_sophos  ---------------------> Kills sophos processes and such.
#   chrome_cors  ---------------------> Opens up a url in Chrome with CORS safety disabled.
#   provenir_tester  -----------------> Opens up the provenir test page (in a CORS disabled Chrome browser).
#   hr  ------------------------------> Creates a horizontal rule in the terminal.
#   hrr  -----------------------------> Creates a 3-line horizontal rule in the terminal.
#   hhr  -----------------------------> Creates a 3-line horizontal rule in the terminal.
#   pick_a_palette  ------------------> Sets the PALETTE environment variable if not already set.
#   to_epoch  ------------------------> Converts a date in YYYY-mm-dd HH:MM:SS format (using local time zone) to an epoch as milliseconds.
#   to_date  -------------------------> Converts an epoch as milliseconds into a date.
#   echo_white  ----------------------> Outputs a message in white.
#   echo_red  ------------------------> Outputs a message in red.
#   echo_green  ----------------------> Outputs a message in green.
#   echo_yellow  ---------------------> Outputs a message in yellow.
#   echo_blue  -----------------------> Outputs a message in blue.
#   echo_pink  -----------------------> Outputs a message in pink.
#   echo_teal  -----------------------> Outputs a message in teal.
#   echo_underline  ------------------> Outputs an underlined message.
#   echo_strikethrough  --------------> Outputs a message with strikethrough.
#   echo_bad  ------------------------> Outputs a message with bright red background and bright white text.
#   echo_color  ----------------------> Outputs a message using a specific color code.
#   strip_colors  --------------------> Strips the color stuff from a stream.
#   to_stdout_and_strip_colors_log  --> Outputs to stdout and logs to a file with color stuff stripped out.
#   to_stdout_and_strip_colors_log  --> Outputs to stderr and logs to a file with color stuff stripped out.
#   colorize  ------------------------> Easy way to set the color code for a string.
#   show_colors  ---------------------> Outputs a chunk of color info.
#   jqq  -----------------------------> Shortcut for jq to output a variable.
#   tee_pbcopy  ----------------------> Outputs to stdout as well as copy it to the clipboard.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within.
$( echo -e "\033[1;37msource $( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )\033[0m" )
EOF
    exit 1
fi

# Output a command, then execute it.
# Usage: echo_do <command> [<arg1> [<arg2> ...]]
#   or   echo_do <<< "command string"
# Examples:
#   echo_do say -vVictoria -r200 "Buu Whoa"
#   echo_do <<< "say -vVictoria -r200 \"YEAH BUDDY\""
# You can technically pipe the commands into echo_do too,
#   but then you lose the ability to get at the result environment variables.
# The string used for command display will be stored in ECHO_DO_CMD_STR.
# stdout results of the command will be stored in ECHO_DO_STDOUT.
# stderr results of the command will be stored in ECHO_DO_STDERR.
# The exit code of the command will be stored in ECHO_DO_EXIT_CODE.
#   and also returned by the function.
# If no command is provided, this will return with exit code 124
#   and none of the above variables will be set.
# If the command is provided as a single string and cannot be parsed,
#   this will return with exit code 125
echo_do () {
    unset ECHO_DO_CMD_STR ECHO_DO_STDOUT ECHO_DO_STDERR ECHO_DO_EXIT_CODE
    local cmd_pieces pieces_for_output cmd_piece tmp_stderr tmp_stdout tmp_exit_code
    if [[ -t 0 ]]; then
        cmd_pieces=( "$@" )
    else
        cmd_pieces=( "$( cat - )" )
    fi
    if [[ "${#cmd_pieces[@]}" -eq '0' || "${cmd_pieces[@]}" =~ ^[[:space:]]*$ ]]; then
        return 124
    fi
    if [[ "${#cmd_pieces[@]}" -eq '1' && "${cmd_pieces[@]}" =~ [[:space:]] ]]; then
        local cmd i c skip q
        cmd="${cmd_pieces[@]}"
        if [[ ( "$cmd" =~ ^'"' && "$cmd" =~ '"'$ ) || ( "$cmd" =~ ^"'" && "$cmd" =~ "'"$ ) ]]; then
            cmd="$( echo -E "$cmd" | sed -E 's/^.//; s/.$//;' )"
        fi
        cmd_pieces=()
        for i in $( seq 0 "$(( $( echo -E "${#cmd}" ) - 1 ))" ); do
            c="${cmd:$i:1}"
            if [[ -n "$skip" ]]; then
                c=
                skip=
            elif [[ -n "$q" ]]; then
                if [[ "$c" == '\' && "$q" == '"' && "${cmd:i+1:1}" == '"' ]]; then
                    c='"'
                    skip='Y'
                elif [[ "$c" == "$q" ]]; then
                    q=
                    c=
                fi
            elif [[ "$c" == '"' || "$c" == "'" ]]; then
                q="$c"
                c=
            elif [[ "$c" =~ ^[[:space:]]$ ]]; then
                if [[ -n "$cmd_piece" ]]; then
                    cmd_pieces+=( "$cmd_piece" )
                    cmd_piece=
                fi
                c=
            fi
            cmd_piece="$cmd_piece$c"
        done
        if [[ -n "$q" ]]; then
            return 125
        fi
        if [[ -n "$cmd_piece" ]]; then
            cmd_pieces+=( "$cmd_piece" )
        fi
    fi
    pieces_for_output=()
    for cmd_piece in "${cmd_pieces[@]}"; do
        if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
            pieces_for_output+=( "\"$( echo -E "$cmd_piece" | sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' )\"" )
        else
            pieces_for_output+=( "$cmd_piece" )
        fi
    done
    ECHO_DO_CMD_STR="${pieces_for_output[@]}"
    echo -en "\033[1;37m"
    echo -En "$ECHO_DO_CMD_STR"
    echo -e "\033[0m"
    # echo -en "\033[1;35m"
    # for c in "${cmd_pieces[@]}"; do echo -E ">$c<"; done
    # echo -e "\033[0m"
    tmp_stderr="$( mktemp -t echo_do_stderr )"
    tmp_stdout="$( mktemp -t echo_do_stdout )"
    tmp_exit_code="$( mktemp -t echo_do_exit_code )"
    ( ( "${cmd_pieces[@]}"; echo "$?" > "$tmp_exit_code" ) | tee "$tmp_stdout" ) 3>&1 1>&2 2>&3 | tee "$tmp_stderr"
    ECHO_DO_STDERR="$( cat "$tmp_stderr" )"
    ECHO_DO_STDOUT="$( cat "$tmp_stdout" )"
    ECHO_DO_EXIT_CODE="$( cat "$tmp_exit_code" )"
    rm "$tmp_stderr"
    rm "$tmp_stdout"
    rm "$tmp_exit_code"
    return "$ECHO_DO_EXIT_CODE"
}

# Same as echo_do but with an extra line at the end
echo_do_ln () {
    local retval
    if [[ -t 0 ]]; then
        echo_do "$@"
        retval=$?
    else
        echo_do <<< "$( cat - )"
        retval=$?
    fi
    echo -E ''
    return "$retval"
}

get_shell_type () {
    local shell_command
    shell_command=$( ps -o command= $$ )
    if [[ -n $( echo "$shell_command" | grep -E "zsh$" ) ]]; then
        echo "zsh"
    elif [[ -n $( echo "$shell_command" | grep -E "bash$" ) ]]; then
        echo "bash"
    else
        echo $shell_command
    fi
}

# Kills all the sophos stuff
# Usage: kill_sophos
kill_sophos () {
    ps aux | grep sophos | awk '{print $2}' | xargs sudo kill -9
}

# Open up a chrome page with CORS safety disabled.
# Usage: chrome_cors <url>
chrome_cors () {
    open -n -a "Google Chrome" "$1" --args --user-data-dir="/tmp/chrome_dev_test" --disable-web-security --new-window
}

# Opens up my provenir test page.
# Usage: provenir_tester
provenir_tester () {
    chrome_cors '/Users/dwedul/git/shelter/provenir_testing/sofi_provenir_test_case_generator_v2.html'
}

# Creates a horizontal rule with a message in it.
# Usage: hr <message>
hr () {
    local message char termwidth available sixths leftover block empty section padding left_wing right_wing unset_palette
    message="$*"
    if [[ -n "$message" ]]; then
        message=" $message "
    fi
    char='#'
    termwidth=$( tput cols )
    available=$(( $termwidth - ${#message} - 2 ))
    sixths=$(( $available / 12 ))
    leftover=$(( $(( $available - $sixths * 12 )) / 2 ))
    block="$( printf '%0.1s' "$char"{1..500} )"
    empty="$( echo -E "$block" | sed "s/$char/ /g" )"
    section="${block:0:$sixths}"
    padding="${empty:0:$leftover}"
    left_wing=""
    right_wing=""
    pick_a_palette && unset_palette="Yup"
    for i in ${PALETTE[*]}; do
        new_piece="$( echo -E "\033[38;5;${i}m${section}\033[0m" )"
        left_wing="$left_wing$new_piece"
        right_wing="$new_piece$right_wing"
    done
    echo -e "$padding$left_wing\033[38;5;15m$message\033[0m$right_wing$padding"
    [[ -n "$unset_palette" ]] && unset PALETTE
}

# Creates a 3-line horizontal rule with a message in it.
# Usage: hrr <message>
hrr () {
    local message blank unset_palette
    message="$*"
    pick_a_palette && unset_palette="Yup"
    blank="$( hr )"
    echo -e "$blank"
    hr "$message"
    echo -e "$blank"
    [[ -n "$unset_palette" ]] && unset PALETTE
}

# Creates a 3-line horizontal rule with a message in it.
# Usage: hhr <message>
hhr () {
    hrr $@
}

# Sets the PALETTE environment veriable if it's not already set.
# Usage: pick_a_palette && echo "PALETTE set to ${PALETTE[*]}"
pick_a_palette () {
    if [[ -z "${PALETTE+x}" ]]; then
        local choice
        choice=$[RANDOM%8]
        case "$choice" in
            0) PALETTE=(232 236 240 244 248 252);;   #white
            1) PALETTE=(16 17 18 19 20 21);;         #blue
            2) PALETTE=(16 22 28 34 40 46);;         #green
            3) PALETTE=(16 64 106 148 184 226);;     #yellow
            4) PALETTE=(16 94 130 166 202 208);;     #orange
            5) PALETTE=(16 52 88 124 160 196);;      #red
            6) PALETTE=(16 54 92 129 165 206);;      #purple
            7) PALETTE=(16 $[RANDOM%256] $[RANDOM%256] $[RANDOM%256] $[RANDOM%256] $[RANDOM%256]);;
        esac
        return 0
    fi
    return 1
}

# Convert a date and time into an epoch as milliseconds.
# Assumes the provided date/time is in your system's local time zone.
# Usage: to_epoch YYYY-mm-dd HH:MM:SS
#  or    to_epoch now
to_epoch () {
    local d t dt
    d="$1"
    t="$2"
    if [[ -z "$d" || "$d" == "-h" || "$d" == "--help" ]]; then
        echo "Usage: to_epoch yyyy-mm-dd hh:mm:ss"
        return 0
    fi
    if [[ "$d" == "now" ]]; then
        date '+%s000'
        return 0
    fi
    if [[ -z "$t" ]]; then
        dt="$d"
    else
        dt="$d $t"
    fi
    # If it starts with a date and time, get rid of anythings. This way the input is compatible with the output of to_date
    dt="$( echo -E "$dt" | sed -E 's/^([[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2} [[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}).*$/\1/' )"
    if [[ "$dt" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}[[:space:]][[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        date -j -f '%F %T' "$dt" '+%s000'
    else
        >&2 echo -E "'$dt' -> Invalid format. Use yyyy-mm-dd hh:mm:ss"
    fi
}

# Convert an epoch as milliseconds into a date and time.
# Usage: to_date <epoch>
to_date () {
    local e
    e="$1"
    if [[ -z "$e" || "$e" == '-h' || "$e" == '--help' ]]; then
        >&2 echo 'Usage: to_date <epoch in milliseconds>';
        return 0
    fi
    date -r "$(( $e / 1000 ))" '+%F %T %z (%Z)'
}

# Usage: echo_white <string>
echo_white () {
    echo_color '1;37' "$@"
}

# Usage: echo_red <string>
echo_red () {
    echo_color '1;31' "$@"
}

# Usage: echo_green <string>
echo_green () {
    echo_color '1;32' "$@"
}

# Usage: echo_yellow <string>
echo_yellow () {
    echo_color '1;33' "$@"
}

# Usage: echo_blue <string>
echo_blue () {
    echo_color '1;34' "$@"
}

# Usage: echo_pink <string>
echo_pink () {
    echo_color '1;35' "$@"
}

# Usage: echo_teal <string>
echo_teal () {
    echo_color '1;36' "$@"
}

echo_underline () {
    echo_color '4' "$@"
}

echo_strikethrough () {
    echo_color '9' "$@"
}

echo_reversed () {
    echo_color '7' "$@"
}

# Usage: echo_bad <string>
echo_bad () {
    echo_color '1;38;5;231;48;5;196' "$@"
}


# Usage: echo_color <color code> <message>
echo_color () {
    local c m n r
    if [[ -n "$1" && "$1" =~ ^[[:digit:]]+(\;[[:digit:]]+)*$ ]]; then
        c="$1"
        shift
    else
        c='0'
    fi
    if [[ -n "$1" && "$1" == '-n' ]]; then
        n="$1"
        shift
    fi
    case "$c" in
        4|7|9) r=$(( c + 20 ));;
        *) r=0;;
    esac
    m="$@"
    echo -e $n "\033[${c}m${m}\033[${r}m"
}

# Usage: <stuff> | strip_colors
strip_colors () {
    sed -E "s/$( echo -e "\033" )\[(;|[[:digit:]])+m//g"
}

# Usage: <stuff> | to_stdout_and_strip_colors_log "logfile"
to_stdout_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stdout_and_strip_colors_log <filename>"
    fi
    cat - > >( tee >( strip_colors >> "$1" ) )
}

# Usage: <stuff> | to_stderr_and_strip_colors_log "logfile"
to_stderr_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stderr_and_strip_colors_log <filename>"
    fi
    cat - > >( >&2 tee >( strip_colors >> "$1" ) )
}

colorize () {
    local code str
    code="$1"
    shift
    str="$*"
    echo -e "\033[${code}m$str\033[0m"
}

show_colors () {
    output=''
    for c in $(seq 0 79); do
        if [[ "$(( c % 10 ))" -eq 0 ]]; then
            output="$( echo -E "$output" | sed -E 's/~$/\\n/;' )"
        fi
        output="$output$( __get_show_color_str "1;$c" "5" )~"
    done
    echo -e "$output" | column -s '~' -t
}

# Just makes it easier to use jq on a variable.
# This is basically just a shortcut for  echo <json> | jq <options> <query>
# If the query is omitted '.' is used.
# Usage: jqq <json> [<query>] [<options>]
jqq () {
    local json query
    json="$1"
    shift
    if [[ "$json" == '-h' || "$json" == '--help' ]]; then
        cat << EOF
jqq - Quick jq command for dealing with json in variables.

Usage: jqq <json> [<query>] [<options>]

    The first argument is taken to be the json.
    The query is optional. The default is '.'.
    If the query is provided, all other arguments are passed in as options to jq.
    If the second argument starts with a - (dash) then it is treated as an option and the default query is used.

    Examples:
        jqq "\$foo"
        jqq "\$foo" -c
        jqq "\$foo" '.[]'
        jqq "\$foo" '.[3].name' -r

EOF
        return 0
    fi
    if [[ "$1" =~ ^- ]]; then
        query='.'
    else
        query="$1"
        shift
    fi
    echo "$json" | jq "$@" "$query"
}

# Usage: <do stuff> | tee_pbcopy
tee_pbcopy () {
    tee >( awk '{if(p) print(l);l=$0;p=1;} END{printf("%s",l);}' | pbcopy )
}

__get_show_color_str () {
    local code width format
    code="$1"
    width="$2"
    if [[ -n "$width" ]]; then
        format="%-${width}s"
    else
        format="%s"
    fi
    colorize "$code" "### $( printf $format $code ) ###"
}
