#!/bin/bash
# This file contains various functions for printing things to your terminal with colors.
# This file can be sourced to add the functions to your environment.
# This file can also be executed to run the echo_color function without adding it to your environment.
#
# File contents:
#   echo_color  ----------------------> Outputs a message using a specific color code.
#   show_colors  ---------------------> Outputs color names and examples.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: echo_color [<color name>|<color code>|-n|-N|--debug] -- <message>
echo_color () {
    local usage
    usage="$( cat << EOF
echo_color - Makes it easier to output things in colors.

Usage: echo_color [<color name>|<color code>|-n|-N|-d|--debug|--examples] -- <message>
    <color name> can be one of:
                    black  red  green  yellow  blue  magenta  cyan  light-gray
                    dark-gray  light-red  light-green  light-yellow  light-blue  light-magenta  light-cyan  white
                    bg-black  bg-red  bg-green  bg-yellow  bg-blue  bg-magenta  bg-cyan  bg-light-gray
                    bg-dark-gray  bg-light-red  bg-light-green  bg-light-yellow  bg-light-blue  bg-light-magenta  bg-light-cyan  bg-white
                    bold  dim  underline  strikethrough  reversed
                    success  info  warn  error  good  bad
    <color code> can be one or more numerical color code numbers separated by semicolons or spaces.
                    These values will be formatted correctly and placed between \"<esc>[\" and \"m\" for output.
                    I have found this page to be a good resource: https://misc.flogisoft.com/bash/tip_colors_and_formatting
    -n signifies that you do not want a trailing newline added to the output.
    -N signifies that you DO want a trailing newline added to the output. This is the default behavior.
    --debug will cause extra messages to be sent to stderr that can help with debugging if needed.
    --examples will cause the other parameters (except --debug) to be ignored, and instead output a set of examples.
EOF
)"
    local code_on_parts without_newline debug show_examples message
    local code_on code_off_parts code_off newline_flag full_output
    code_on_parts=()
    while [[ "$#" -gt '0' && "$1" != '--' ]]; do
        case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help|help)
            echo "$usage"
            return 0
            ;;
        black)                  code_on_parts+=( '30' );;
        red)                    code_on_parts+=( '31' );;
        green)                  code_on_parts+=( '32' );;
        yellow)                 code_on_parts+=( '33' );;
        blue)                   code_on_parts+=( '34' );;
        magenta|pink)           code_on_parts+=( '35' );;
        cyan|teal)              code_on_parts+=( '36' );;
        light-gray|light-grey)  code_on_parts+=( '37' );;
        dark-gray|dark-grey)                                    code_on_parts+=( '90' );;
        light-red|bright-red)                                   code_on_parts+=( '91' );;
        light-green|bright-green)                               code_on_parts+=( '92' );;
        light-yellow|bright-yellow)                             code_on_parts+=( '93' );;
        light-blue|bright-blue)                                 code_on_parts+=( '94' );;
        light-magenta|bright-magenta|light-pink|bright-pink)    code_on_parts+=( '95' );;
        light-cyan|bright-cyan|light-teal|bright-teal)          code_on_parts+=( '96' );;
        white)                                                  code_on_parts+=( '97' );;
        bg-black)                       code_on_parts+=( '40' );;
        bg-red)                         code_on_parts+=( '41' );;
        bg-green)                       code_on_parts+=( '42' );;
        bg-yellow)                      code_on_parts+=( '43' );;
        bg-blue)                        code_on_parts+=( '44' );;
        bg-magenta|bg-pink)             code_on_parts+=( '45' );;
        bg-cyan|bg-teal)                code_on_parts+=( '46' );;
        bg-light-gray|bg-light-grey)    code_on_parts+=( '47' );;
        bg-dark-gray|bg-dark-grey)                                          code_on_parts+=( '100' );;
        bg-light-red|bg-bright-red)                                         code_on_parts+=( '101' );;
        bg-light-green|bg-bright-green)                                     code_on_parts+=( '102' );;
        bg-light-yellow|bg-bright-yellow)                                   code_on_parts+=( '103' );;
        bg-light-blue|bg-bright-blue)                                       code_on_parts+=( '104' );;
        bg-light-magenta|bg-bright-magenta|bg-light-pink|bg-bright-pink)    code_on_parts+=( '105' );;
        bg-light-cyan|bg-bright-cyan|bg-light-teal|bg-bright-teal)          code_on_parts+=( '106' );;
        bg-white)                                                           code_on_parts+=( '107' );;
        bold)               code_on_parts+=( '1' );;
        dim)                code_on_parts+=( '2' );;
        underline)          code_on_parts+=( '4' );;
        reverse|reversed)   code_on_parts+=( '7' );;
        strikethrough)      code_on_parts+=( '9' );;
        success)            code_on_parts+=( '97;42' );;
        info)               code_on_parts+=( '97;100' );;
        warn|warning)       code_on_parts+=( '93;100' );;
        error)              code_on_parts+=( '1;91;100' );;
        good)               code_on_parts+=( '92;100' );;
        bad)                code_on_parts+=( '1;97;41' );;
        -n) without_newline='YES' ;;
        -N) without_newline= ;;
        -d|--debug) debug='--debug' ;;
        --examples) show_examples='--examples' ;;
        *)
            if [[ "$1" =~ ^[[:digit:]]+(([[:space:]]+|\;)[[:digit:]]+)*$ ]]; then
                code_on_parts+=( "$( printf %s $1 | sed -E 's/[[:space:]]+/;/g' )" )
            else
                echo -e "echo_color: Invalid color name or code: [$1]." >&2
                return 1
            fi
            ;;
        esac
        shift
    done

    if [[ -n "$show_examples" ]]; then
        [[ "$debug" ]] && echo -E "Showing examples." >&2
        show_colors $debug
        return 0
    fi

    if [[ "$1" != '--' ]]; then
        echo -E "No '--' separator found." >&2
        return 1
    fi
    shift
    message="$@"
    # Allowing for an empty message because:
    #   a) It won't really hurt anything.
    #   b) It makes calling this function with user-defined input easier and nicer.
    #   c) It'll make it easier to expand functionality to allow messages to be piped in.

    if [[ -n "$without_newline" ]]; then
        newline_flag='-n'
    fi

    if [[ "${#code_on_parts[@]}" -gt '0' ]]; then
        code_on="$( join_str ';' "${code_on_parts[@]}" )"
        # This regex is a combination of all of the cases handled below.
        # Basically, if there's a part of the code that we don't know how to turn off specifically, we'll just reset everything.
        # Otherwise, if we know how to turn off ALL pieces of the desired code, we'll just turn off what we're turning on.
        # This allows for commands like this to behave as expected:
        #   echo_color red -- "This is a $( echo_color bold -- "test" ) of the echo_color function."
        if [[ ! "$code_on" =~ ^([12479]|(3|4|9|10)[01234567]|[34]8\;5\;[[:digit:]]{1,3})(\;([12479]|(3|4|9|10)[01234567]|[34]8\;5\;[[:digit:]]{1,3}))*$ ]]; then
            code_off='0'
        else
            code_off_parts=()
            # This part of the regex: (^|\;) is checking for either the start of the string, or a semicolon.
            # This part of the regex: (\;|$) is checking for either a semicolon or the end of the string.
            # Check for bold or dim: 1 or 2
            if [[ "$code_on" =~ (^|\;)1(\;|$) ]]; then
                code_off_parts+=( '21' '22' )
            elif [[ "$code_on" =~ (^|\;)2(\;|$) ]]; then
                code_off_parts+=( '22' )
            fi
            # Check for underline: 4
            if [[ "$code_on" =~ (^|\;)4(\;|$) ]]; then
                code_off_parts+=( '24' )
            fi
            # Check for reversed: 7
            if [[ "$code_on" =~ (^|\;)7(\;|$) ]]; then
                code_off_parts+=( '27' )
            fi
            # Check for strikethrough: 9
            if [[ "$code_on" =~ (^|\;)9(\;|$) ]]; then
                code_off_parts+=( '29' )
            fi
            # Check for text/foreground colors: 30-37, 90-97, or 38;5;ddd
            if [[ "$code_on" =~ (^|\;)[39][01234567](\;|$) || "$code_on" =~ (^|\;)38\;5\;[[:digit:]]{1,3}(\;|$) ]]; then
                code_off_parts+=( '39' )
            fi
            # Check for background colors: for 40-47, 100-107, or 48;5;ddd
            if [[ "$code_on" =~ (^|\;)(4|10)[01234567](\;|$) || "$code_on" =~ (^|\;)48\;5\;[[:digit:]]{1,3}(\;|$) ]]; then
                code_off_parts+=( '49' )
            fi
            code_off="$( join_str ';' "${code_off_parts[@]}" )"
        fi
        full_output="\033[${code_on}m${message}\033[${code_off}m"
    else
        full_output="$message"
    fi

    if [[ -n "$debug" ]]; then
        {
            printf '  Opening code: [%s].\n' "$code_on"
            printf '       Message: [%s].\n' "$message"
            printf '  Closing code: [%s].\n' "$code_off"
            printf 'Adding newline: [%s].\n' "$( [[ -n "$newline_flag" ]] && echo 'NO' || echo 'yes' )"
            printf 'Without interpretation: [%s].\n' "$full_output"
            printf '   With interpretation: [%b].\n' "$full_output"
        } >&2
    fi
    echo -e $newline_flag "$full_output"
}

# Displays examples of some color codes
# Usage: show_colors [-v|--verbose] [-d|--debug] [-c|--combos]
show_colors () {
    local debug verbose show_combos text_colors background_colors effects special_formats fg_codes bg_codes effect_codes output
    while [[ "$#" -gt '0' ]]; do
        case "$( printf '%s' "$1" | tr '[:upper:]' '[:lower:]' )" in
        -v|--verbose)   verbose='--verbose' ;;
        -d|--debug)     debug='--debug' ;;
        -c|--combos)    show_combos='--combos' ;;
        *)
            echo "Unknown parameter: [$1]." >&2
            return 1
            ;;
        esac
        shift
    done
    text_colors=(
        black      red            blue        green
        dark-gray  light-red      light-blue  light-green
        light-gray magenta        cyan        yellow
        white      light-magenta  light-cyan  light-yellow
    )
    background_colors=(
        bg-black      bg-red            bg-blue        bg-green
        bg-dark-gray  bg-light-red      bg-light-blue  bg-light-green
        bg-light-gray bg-magenta        bg-cyan        bg-yellow
        bg-white      bg-light-magenta  bg-light-cyan  bg-light-yellow
    )
    effects=(
        bold  dim  underline  strikethrough  reversed
    )
    special_formats=(
        success  info  warn  error  good  bad
    )
    fg_codes=( 30 90 37 97 31 91 35 95 34 94 36 96 32 92 33 93 )
    bg_codes=( 40 100 47 107 41 101 45 105 44 104 46 106 42 102 43 103 )
    output="$(
        # Sneaky private functions with access to variables from parent function. Teehee.
        output_simple_section () {
            local title name_width per_line i text_color
            title="$1"
            shift
            name_width="$1"
            shift
            per_line="$1"
            shift
            echo -E "$title:"
            i=0
            for color in "$@"; do
                i=$(( i + 1 ))
                echo -n ' '
                if [[ "$i" -eq '1' ]]; then
                    echo -n '   '
                fi
                printf "%${name_width}s:[" "$color"
                echo_color -n "$color" $debug -- " Example "
                printf ']'
                if [[ "$i" -eq "$per_line" ]]; then
                    i=0
                    echo ''
                fi
            done
            [[ "$i" -ne '0' ]] && echo ''
        }
        output_combo_section () {
            local title added_code pad fg_code bg_code code
            title="$1"
            added_code="$2"
            if [[ -n "$added_code" ]]; then
                added_code="${added_code};"
            else
                pad=' '
            fi
            echo -E "$title:"
            for fg_code in "${fg_codes[@]}"; do
                printf '    '
                for bg_code in "${bg_codes[@]}"; do
                    code="${added_code}${fg_code};${bg_code}"
                    printf ' %-9b' "\033[${code}m${pad}${code}${pad}\033[0m"
                done
                echo ''
            done
        }
        output_simple_section 'Text Colors'       '16' '4' "${text_colors[@]}"
        output_simple_section 'Background Colors' '16' '4' "${background_colors[@]}"
        output_simple_section 'Text Effects'      '1'  '9' "${effects[@]}"
        output_simple_section 'Special Formats'   '1'  '9' "${special_formats[@]}"
        if [[ -n "$show_combos" ]]; then
            output_combo_section 'Color Combos'            ''
            output_combo_section 'Color Combos - Bold'     '1'
            output_combo_section 'Color Combos - Dim'      '2'
            output_combo_section 'Color Combos - Reversed' '7'
        fi
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
