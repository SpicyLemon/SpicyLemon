#!/bin/bash
# This file contains generic functions for helping do random things that I often need.
# File contents:
#   echo_do  -------------------------> Outputs a command in bright white, then executes it.
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
#   join_str  ------------------------> Joins a list of parameters using a delimiter.
#   echo_white  ----------------------> Outputs a message in white.
#   echo_red  ------------------------> Outputs a message in red.
#   echo_green  ----------------------> Outputs a message in green.
#   echo_yellow  ---------------------> Outputs a message in yellow.
#   echo_blue  -----------------------> Outputs a message in blue.
#   echo_pink  -----------------------> Outputs a message in pink.
#   echo_cyan  -----------------------> Outputs a message in teal.
#   echo_underline  ------------------> Outputs an underlined message.
#   echo_strikethrough  --------------> Outputs a message with strikethrough.
#   echo_bad  ------------------------> Outputs a message with bright red background and bright white text.
#   echo_color  ----------------------> Outputs a message using a specific color code.
#   strip_colors  --------------------> Strips the color stuff from a stream.
#   escape_escapes  ------------------> Escapes any escape characters in a stream.
#   to_stdout_and_strip_colors_log  --> Outputs to stdout and logs to a file with color stuff stripped out.
#   to_stdout_and_strip_colors_log  --> Outputs to stderr and logs to a file with color stuff stripped out.
#   show_colors  ---------------------> Outputs a chunk of color info.
#   jqq  -----------------------------> Shortcut for jq to output a variable.
#   tee_pbcopy  ----------------------> Outputs to stdout as well as copy it to the clipboard.
#   ps_grep  -------------------------> Greps ps with provided input.
#   i_can  ---------------------------> Tests if a command is available.
#   can_i  ---------------------------> Outputs results of i_can.
#   print_args  ----------------------> Outputs all parameters received.
#   change_word  ---------------------> Changes a word from one thing to another in one or more files.
#   java_8_activate  -----------------> Exports JAVA_HOME to point to Java 8.
#   java_8_deactivate  ---------------> Unsets JAVA_HOME.
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
unset sourced

# Output a command, then execute it.
# Usage: echo_do <command> [<arg1> [<arg2> ...]]
#   or   echo_do "command string"
# Examples:
#   echo_do say -vVictoria -r200 "Buu Whoa"
#   echo_do "say -vVictoria -r200 \"YEAH BUDDY\""
# If no command is provided, this will return with exit code 124.
echo_do () {
    local cmd_pieces pieces_for_output cmd_piece retval
    # Check for no parameters.
    # Make sure there's still arguments left to form the command.
    if [[ "$#" -eq '0' || "$@" =~ ^[[:space:]]*$ ]]; then
        echo 'No command provided to echo_do.' >&2
        return 124
    fi
    # Do a little processing on the provided arguments.
    if [[ "$#" -eq '1' && ( "$@" =~ [[:space:]\(=] || -z "$( command -v "$@" )" ) ]]; then
        # If there's only 1 argument and
        #   it contains a space, open parenthesis, or an equals
        #   or it is not an actual command
        # then we need to run it using eval.
        # This primarily allows for setting environment variables using this function.
        cmd_pieces=( 'eval' "$@" )
        pieces_for_output=( "$@" )
    else
        # Otherwise, we can just throw everything into the command pieces as it is.
        cmd_pieces=( "$@" )
        # We then need to slightly alter the pieces in order to properly output the command.
        pieces_for_output=()
        for cmd_piece in "$@"; do
            if [[ "$cmd_piece" =~ [[:space:]\'\"] ]]; then
                # If this piece has a space, a single, or double quote, then it needs to be escaped and wrapped.
                # Escape again all already escaped double quotes, then escape all double quotes.
                # And put the whole thing in double quotes.
                pieces_for_output+=( "\"$( echo -E "$cmd_piece" | sed -E 's/\\"/\\\\"/g; s/"/\\"/g;' )\"" )
            else
                # Otherwise, no change is needed.
                pieces_for_output+=( "$cmd_piece" )
            fi
        done
    fi

    # Show the command string in bold white.
    echo -en "\033[1;37m"
    echo -En "${pieces_for_output[@]}"
    echo -e "\033[0m"
    # Execute the command.
    "${cmd_pieces[@]}"
    retval="$?"
    echo ''
    return "$?"
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
    sudo ps aux | grep -v ' grep ' | grep -i sophos | awk '{print $2}' | xargs sudo kill -9
}

# Open up a chrome page with CORS safety disabled.
# Usage: chrome_cors <url>
chrome_cors () {
    open -n -a "Google Chrome" "$1" --args --user-data-dir="/tmp/chrome_dev_test" --disable-web-security --new-window
}

# Opens up my provenir test page.
# Usage: provenir_tester
provenir_tester () {
    chrome_cors '/Users/dwedul/git/danny-wedul/provenir_testing/sofi_provenir_test_case_generator_v2.html'
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
# Usage: to_epoch yyyy-MM-dd [HH:mm[:ss[.ddd]]] [(+|-)HHmm]
#  or    to_epoch now
to_epoch () {
    local pieces the_date the_time the_time_zone s_fractions ms_fractions ms epoch_s epoch_ms
    if [[ -z "$1" || "$1" == "-h" || "$1" == "--help" ]]; then
        echo "Usage: to_epoch yyyy-MM-dd [HH:mm[:ss[.ddd]]] [(+|-)HHmm]"
        return 0
    fi
    if [[ "$1" == "now" ]]; then
        date '+%s000'
        return 0
    fi
    # Allow for the input to be in ISO 8601 format where the date and time are combined with a T.
    pieces=( $( echo -E -n "$@" | tr 'T' ' ' ) )
    # zsh is 1 indexed, bash is 0.
    if [[ -n "${pieces[0]}" ]]; then
        the_date="${pieces[0]}"
        the_time="${pieces[1]}"
        the_time_zone="${pieces[2]}"
    else
        the_date="${pieces[1]}"
        the_time="${pieces[2]}"
        the_time_zone="${pieces[3]}"
    fi
    # Since $the_time is optional, if it starts with a + or -,
    # it's actually the time zone piece.
    if [[ "$the_time" =~ ^[+-] ]]; then
        the_time_zone="$the_time"
        the_time=
    fi
    # Try to make $the_date into yyyy-MM-dd format.
    # Allow for input to be in the formats yyyy, yyyyMM, yyyy-MM, yyyyMMdd, yyyyMM-dd, yyyy-MMdd, yyyy-MM-dd,
    # or MM-dd-yyyy
    # or have different delimiters.
    the_date="$( echo -E -n "$the_date" | tr -c "[:digit:]" "-" )"
    if [[ "$the_date" =~ ^[[:digit:]]{4}(-?[[:digit:]]{2}){0,2}$ ]]; then
        the_date="$( echo -E -n "$the_date" | tr -d '-' | sed 's/$/0101/' | head -c 8 | sed -E 's/^(....)(..)(..)$/\1-\2-\3/' )"
    elif [[ "$the_date" =~ ^[[:digit:]]{2}-[[:digit:]]{2}-[[:digit:]]{4}$ ]]; then
        pieces=( $( echo -E -n "$the_date" | tr '-' ' ' ) )
        if [[ -n "${pieces[0]}" ]]; then
            the_date="${pieces[2]}-${pieces[0]}-${pieces[1]}"
        else
            the_date="${pieces[3]}-${pieces[1]}-${pieces[2]}"
        fi
    fi
    if [[ ! "$the_date" =~ ^[[:digit:]]{4}-[[:digit:]]{2}-[[:digit:]]{2}$ ]]; then
        >&2 echo "Invalid date format [$the_date]. Use yyyy-MM-dd."
        return 1
    fi
    # Try to make $the_time into HH:mm:ss format and handle any extra precision.
    # Allow for no time input,
    # or formats of HH, HHmm, HH:mm, HHmmss, HHmm:ss, HH:mmss, HH:mm:ss
    # or formats of HHmmss.d+, HHmm:ss.d+, HH:mmss.d+, HH:mm:ss.d+
    s_fractions=
    ms_fractions=
    if [[ -z "$the_time" ]]; then
        the_time='00:00:00'
    elif [[ "$the_time" =~ ^[[:digit:]]{2}(:?[[:digit:]]{2}){0,2}$ ]]; then
        the_time="$( echo -E -n "$the_time" | tr -d ':' | sed 's/$/0000/' | head -c 6 | sed -E 's/^(..)(..)(..)$/\1:\2:\3/' )"
    elif [[ "$the_time" =~ ^[[:digit:]]{2}:?[[:digit:]]{2}:?[[:digit:]]{2}\.[[:digit:]]+$ ]]; then
        pieces=( $( echo -E "$the_time" | tr '.' ' ' ) )
        if [[ -n "${pieces[0]}" ]]; then
            the_time="${pieces[0]}"
            s_fractions="${pieces[1]}"
        else
            the_time="${pieces[1]}"
            s_fractions="${pieces[2]}"
        fi
        the_time="$( echo -E -n "$the_time" | tr -d ':' | sed -E 's/^(..)(..)(..)$/\1:\2:\3/' )"
        s_fractions="$( echo -E "$s_fractions" | sed -E 's/0+$//' )"
        if [[ "${#s_fractions}" -gt '3' ]]; then
            ms_fractions=".$( echo -E -n "$s_fractions" | sed -E 's/^...//' )"
        fi
    fi
    if [[ ! "$the_time" =~ ^[[:digit:]]{2}:[[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        >&2 echo "Invalid time format [$the_time]. Use HH:mm[:ss[.ddd]]."
        return 1
    fi
    # Make sure the milliseconds have exactly three decials by padding the right with zeros if needed.
    ms="$( echo -E "${s_fractions}000" | head -c 3 )"
    # Try to make $the_time_zone into (+|-)HHmm format.
    # Allow for no time zone, (+|-)HH, (+|-)HHmm (+|-)HH:mm
    if [[ -z "$the_time_zone" ]]; then
        the_time_zone="$( date '+%z' )"
    elif [[ "$the_time_zone" =~ ^[+-][[:digit:]]{2}$ ]]; then
        the_time_zone="${the_time_zone}00"
    elif [[ "$the_time_zone" =~ ^[+-][[:digit:]]{2}:[[:digit:]]{2}$ ]]; then
        the_time_zone="$( echo -E -n "$the_time_zone" | tr -d ':' )"
    fi
    if [[ ! "$the_time_zone" =~ ^[+-][[:digit:]]{4}$ ]]; then
        >&2 echo "Invalid timezone format [$the_time_zone]. Use (+|-)HHmm."
        return 1
    fi
    # Get the epoch as seconds
    epoch_s="$( date -j -f '%F %T %z' "$the_date $the_time $the_time_zone" '+%s' )" || return $?
    # Append the milliseconds and remove any leading zeros.
    epoch_ms="$( echo -E -n "${epoch_s}${ms}" | sed -E 's/^0+//;' )"
    # But make sure there's still at least one digit.
    if [[ -z "$epoch_ms" ]]; then
        epoch_ms="0"
    fi
    echo -E "${epoch_ms}${ms_fractions}"
    return 0
}

# Convert an epoch as milliseconds into a date and time.
# Usage: to_date <epoch in milliseconds>
#  or    to_date now
to_date () {
    local input pieces epoch_ms ms_fractions ms s_fractions epoch_s
    input="$1"
    if [[ -z "$input" || "$input" == '-h' || "$input" == '--help' ]]; then
        >&2 echo 'Usage: to_date <epoch in milliseconds>';
        return 0
    fi
    if [[ "$input" == 'now' ]]; then
        date '+%F %T %z (%Z) %A'
        return 0
    fi
    # Split out the input into milliseconds and fractional milliseconds
    if [[ "$input" =~ ^[[:digit:]]+(\.[[:digit:]]+)?$ ]]; then
        pieces=( $( echo -E -n "$input" | tr '.' ' ' ) )
        if [[ -n "${pieces[0]}" ]]; then
            epoch_ms="${pieces[0]}"
            ms_fractions="${pieces[1]}"
        else
            epoch_ms="${pieces[1]}"
            ms_fractions="${pieces[2]}"
        fi
    else
        >&2 echo "Invalid input: [$input]."
        return 1
    fi
    ms="$( echo -E -n "$epoch_ms" | tail -c 3 )"
    s_fractions="$( echo -E -n "${ms}${ms_fractions}" | sed -E 's/0+$//' )"
    if [[ -n "$s_fractions" ]]; then
        s_fractions=".$s_fractions"
    fi
    epoch_s="$( echo -E -n "$epoch_ms" | sed -E 's/...$//' )"
    date -r "$epoch_s" "+%F %T${s_fractions} %z (%Z) %A"
    return 0
}

# Joins all provided parameters using the provided delimiter.
# Usage: join_str <delimiter> [<arg1> [<arg2>... ]]
join_str () {
    local d retval
    d="$1"
    shift
    retval="$1"
    shift
    while [[ "$#" -gt '0' ]]; do
        retval="${retval}${d}${1}"
        shift
    done
    printf %s "$retval"
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

# Usage: echo_cyan <string>
echo_cyan () {
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

# Usage: echo_color <color code> [-n] <message>
echo_color () {
    local code_on debug newline_flag message code_off_parts code_on_part reset_to_default code_off
    if [[ "$1" =~ ^[[:digit:]]+(\;[[:digit:]]+)*$ ]]; then
        code_on="$1"
        shift
    else
        echo -e "echo_color: Invalid color code: [$1]. Must have format <number>[;<number>[...]]." >&2
        return 1
    fi
    if [[ "$1" == '--debug' ]]; then
        debug='--debug'
        shift
    fi
    if [[ "$1" == '-n' ]]; then
        newline_flag='-n'
        shift
    fi
    message="$@"
    code_off_parts=()
    for code_on_part in $( printf %s "$code_on" | tr ';' '\n' ); do
        case "$code_on_part" in
        1)              code_off_parts+=( 21 22 );;
        2|4|7|9)        code_off_parts+=( $(( code_on_part + 20 )) );;
        3[01234567])    code_off_parts+=( 39 );;
        4[01234567])    code_off_parts+=( 49 );;
        *)
            reset_to_default='YES'
            break
            ;;
        esac
    done
    if [[ -n "$reset_to_default" ]]; then
        code_off='0'
    else
        code_off_parts=( $( echo "${code_off_parts[@]}" | tr ' ' '\n' | sort -n -u ) )
        code_off="$( join_str ';' "${code_off_parts[@]}" )"
    fi
    [[ -n "$debug" ]] && { printf '%s' "\033[${code_on}m${message}\033[${code_off}m -> "; echo -e "[\033[${code_on}m${message}\033[${code_off}m]"; } >&2
    echo -e $newline_flag "\033[${code_on}m${message}\033[${code_off}m"
}

# Usage: <stuff> | strip_colors
strip_colors () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | strip_colors
        return 0
    fi
    sed -E "s/$( echo -e "\033" )\[[[:digit:]]+(;[[:digit:]]+)*m//g"
}

escape_escapes () {
    if [[ "$#" -gt '0' ]]; then
        printf %s "$@" | escape_escapes
        return 0
    fi
    sed -E "s/$( echo -e "\033" )/\\\033/g"
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

# Displays some color codes
# Usage: show_colors
show_colors () {
    local debug verbose codes output
    if [[ "$1" == '--debug' ]]; then
        debug='--debug'
        shift
    elif [[ "$1" == '-v' || "$1" == '--verbose' ]]; then
        verbose='-v'
        shift
    fi
    codes=( '1' '2' '4' '7' '9' $( seq 30 37 ) $( seq 40 47 ) )
    output="$(
        # Sneaky private function with access to variables from parent function. Teehee.
        output_section () {
            local title base_code width padl padr format sub_code code
            title="$1"
            base_code="$2"
            width="$(( ${#base_code} + 2 ))"
            padl="$( printf "% $(( ( 6 - width ) / 2 ))s" '' )"
            padr="$( printf "% $(( ( 6 - width + 1 ) / 2 ))s" '' )"
            format="%-${width}s"
            printf '%18s' "$title: "
            for sub_code in "${codes[@]}"; do
                code="${base_code}${sub_code}"
                echo -n '['
                echo_color "$code" $debug -n "###${padl} $( printf $format $code ) ${padr}###"
                echo -n ']'
                if [[ "$sub_code" -eq '9' || "$sub_code" -eq '37' || "$sub_code" -eq '47' ]]; then
                    printf '\n'
                else
                    printf %s '  '
                fi
            done
        }
        output_section "Normal"        ''
        output_section "Reversed"      '7;'
        output_section "Bold"          '1;'
        output_section "Bold Reversed" '1;7;'
        output_section "Dim"           '2;'
        output_section "Dim Reversed"  '2;7;'
    )"
    if [[ -n "$debug" || -n "$verbose" ]]; then
        {
            echo "escaped:"
            echo -e "$output" | escape_escapes
            echo ''
            echo "unescaped:"
        } >&2
    fi
    echo -e "$output"
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

# Usage: ps_grep <grep parameters>
ps_grep () {
    ps aux | grep "$@" | grep -v grep
}

# Usage: if i_can "foo"; then echo "I can totally foo"; else echo "There's no way I can foo."; fi
i_can () {
    if [[ "$#" -eq '0' ]]; then
        return 1
    fi
    command -v "$@" > /dev/null 2>&1
}

can_i () {
    local c
    c="$@"
    if [[ -z "$c" ]]; then
        echo -E "Usage: can_i <command>"
        return 2
    fi
    if i_can "$c"; then
        echo -E "I can [$c]."
        return 0
    else
        echo -E "I am unable to [$c]."
        return 1
    fi
}

print_args () {
    if [[ "$#" -eq '0' ]]; then
        echo "No arguments provided." >&2
        return 1
    fi
    echo -e "Arguments received:"
    while [[ "$#" -gt '0' ]]; do
        printf '[%s]\n' "$1"
        shift
    done
}

# Usage: change_word <old word> <new word> <files>
change_word () {
    local old_word new_word files file
    old_word="$1"
    new_word="$2"
    shift
    shift
    files="$@"
    if [[ -z "$old_word" || -z "$new_word" || "${#files[@]}" -eq '0' ]]; then
        echo "Usage: change_word <old word> <new word> <files>"
        return 0
    fi
    if [[ "$old_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        echo "Only letters, numbers, dashes, underscores and spaces are allowed in the provided words." >&2
        return 1
    fi
    if [[ "$new_word" =~ [^[:alnum:][:space:]\-_] ]]; then
        echo "Only letters, numbers, dashes, underscores and spaces are allowed in the provided words." >&2
        return 1
    fi
    echo -e "Changing \033[0m \033[1;31m$old_word\033[0m to \033[1;32m$new_word\033[0m"
    echo -e "In: ${files[@]}"
    echo ''
    echo -e "\033[1;37;41m Before: \033[0m"
    grep -E "\<($old_word|$new_word)\>" *.sh \
        | GREP_COLOR='1;31' grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR='1;32' grep --color=always "\<$new_word\>\|$"
    echo ''
    for file in ${files[@]}; do
        sed -i '' "s/[[:<:]]$old_word[[:>:]]/$new_word/g;" "$file"
    done
    echo -e "\033[1;37;42m After: \033[0m"
    grep -E "\<($old_word|$new_word)\>" *.sh \
        | GREP_COLOR='1;31' grep --color=always "\<$old_word\>\|$" \
        | GREP_COLOR='1;32' grep --color=always "\<$new_word\>\|$"
    echo ''
}

java_8_activate () {
    export JAVA_HOME="$( /usr/libexec/java_home -v 1.8 )"
    echo -E "JAVA_HOME set to \"$JAVA_HOME\"."
}

java_8_deactivate () {
    unset JAVA_HOME
    echo -E "JAVA_HOME unset."
}
