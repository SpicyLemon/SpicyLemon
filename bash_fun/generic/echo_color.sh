#!/bin/bash
# This file contains various functions for printing things to your terminal with colors.
# This file can be sourced to add the functions to your environment.
# This file can also be executed to run the echo_color function without adding it to your environment.
#
# File contents:
#   echo_color  ----------------------> Outputs a message using a specific color code.
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
#   show_colors  ---------------------> Outputs a chunk of color info.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

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

echo_bad () {
    echo_color '1;38;5;231;48;5;196' "$@"
}

# Displays examples of some color codes
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

if [[ "$sourced" != 'YES' ]]; then
    echo_color "$@"
    exit $?
fi
unset sourced

return 0
