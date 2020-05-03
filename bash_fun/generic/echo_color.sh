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
    local code_on_parts code_off_parts without_newline debug show_examples code_on_part reset_to_default message
    code_on_parts=()
    code_off_parts=()
    while [[ "$#" -gt '0' && "$1" != '--' ]]; do
        case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help|help)
            echo "$usage"
            return 0
            ;;
        black)                  code_on_parts+=( '30' ); code_off_parts+=( '39' ) ;;
        red)                    code_on_parts+=( '31' ); code_off_parts+=( '39' ) ;;
        green)                  code_on_parts+=( '32' ); code_off_parts+=( '39' ) ;;
        yellow)                 code_on_parts+=( '33' ); code_off_parts+=( '39' ) ;;
        blue)                   code_on_parts+=( '34' ); code_off_parts+=( '39' ) ;;
        magenta|pink)           code_on_parts+=( '35' ); code_off_parts+=( '39' ) ;;
        cyan|teal)              code_on_parts+=( '36' ); code_off_parts+=( '39' ) ;;
        light-gray|light-grey)  code_on_parts+=( '37' ); code_off_parts+=( '39' ) ;;
        dark-gray|dark-grey)                                    code_on_parts+=( '90' ); code_off_parts+=( '39' ) ;;
        light-red|bright-red)                                   code_on_parts+=( '91' ); code_off_parts+=( '39' ) ;;
        light-green|bright-green)                               code_on_parts+=( '92' ); code_off_parts+=( '39' ) ;;
        light-yellow|bright-yellow)                             code_on_parts+=( '93' ); code_off_parts+=( '39' ) ;;
        light-blue|bright-blue)                                 code_on_parts+=( '94' ); code_off_parts+=( '39' ) ;;
        light-magenta|bright-magenta|light-pink|bright-pink)    code_on_parts+=( '95' ); code_off_parts+=( '39' ) ;;
        light-cyan|bright-cyan|light-teal|bright-teal)          code_on_parts+=( '96' ); code_off_parts+=( '39' ) ;;
        white)                                                  code_on_parts+=( '97' ); code_off_parts+=( '39' ) ;;
        bg-black)                       code_on_parts+=( '40' ); code_off_parts+=( '49' ) ;;
        bg-red)                         code_on_parts+=( '41' ); code_off_parts+=( '49' ) ;;
        bg-green)                       code_on_parts+=( '42' ); code_off_parts+=( '49' ) ;;
        bg-yellow)                      code_on_parts+=( '43' ); code_off_parts+=( '49' ) ;;
        bg-blue)                        code_on_parts+=( '44' ); code_off_parts+=( '49' ) ;;
        bg-magenta|bg-pink)             code_on_parts+=( '45' ); code_off_parts+=( '49' ) ;;
        bg-cyan|bg-teal)                code_on_parts+=( '46' ); code_off_parts+=( '49' ) ;;
        bg-light-gray|bg-light-grey)    code_on_parts+=( '47' ); code_off_parts+=( '49' ) ;;
        bg-dark-gray|bg-dark-grey)                                          code_on_parts+=( '100' ); code_off_parts+=( '49' ) ;;
        bg-light-red|bg-bright-red)                                         code_on_parts+=( '101' ); code_off_parts+=( '49' ) ;;
        bg-light-green|bg-bright-green)                                     code_on_parts+=( '102' ); code_off_parts+=( '49' ) ;;
        bg-light-yellow|bg-bright-yellow)                                   code_on_parts+=( '103' ); code_off_parts+=( '49' ) ;;
        bg-light-blue|bg-bright-blue)                                       code_on_parts+=( '104' ); code_off_parts+=( '49' ) ;;
        bg-light-magenta|bg-bright-magenta|bg-light-pink|bg-bright-pink)    code_on_parts+=( '105' ); code_off_parts+=( '49' ) ;;
        bg-light-cyan|bg-bright-cyan|bg-light-teal|bg-bright-teal)          code_on_parts+=( '106' ); code_off_parts+=( '49' ) ;;
        bg-white)                                                           code_on_parts+=( '107' ); code_off_parts+=( '49' ) ;;
        bold)           code_on_parts+=( '1' ); code_off_parts+=( '21' '22' ) ;;
        dim)            code_on_parts+=( '2' ); code_off_parts+=( '22' ) ;;
        underline)      code_on_parts+=( '4' ); code_off_parts+=( '24' ) ;;
        reversed)       code_on_parts+=( '7' ); code_off_parts+=( '27' ) ;;
        strikethrough)  code_on_parts+=( '9' ); code_off_parts+=( '29' ) ;;
        success)        code_on_parts+=( '97' '42' ); code_off_parts+=( '39' '49' ) ;;
        info)           code_on_parts+=( '97' '100' ); code_off_parts+=( '39' '49' ) ;;
        warn|warning)   code_on_parts+=( '93' '100' ); code_off_parts+=( '39' '49' ) ;;
        error)          code_on_parts+=( '1' '91' '100' ); code_off_parts+=( '21' '22' '39' '49' ) ;;
        good)           code_on_parts+=( '92' '100' ); code_off_parts+=( '39' '49' ) ;;
        bad)            code_on_parts+=( '1' '97' '41' ); code_off_parts+=( '21' '22' '39' '49' ) ;;
        -n) without_newline='YES' ;;
        -N) without_newline= ;;
        -d|--debug) debug='--debug' ;;
        --examples) show_examples='--examples' ;;
        *)
            if [[ "$1" =~ ^[[:digit:]]+(([[:space:]]+|\;)[[:digit:]]+)*$ ]]; then
                code_on_part="$( printf %s $1 | sed -E 's/[[:space:]]+/;/g' )"
                code_on_parts+=( "$code_on_part" )
                # TODO: Enhance this to handle multiple settings.
                if [[ "$code_on_part" =~ ^38\;5\;[[:digit:]]+$ ]]; then
                    code_off_parts+=( '39' )
                elif [[ "$code_on_part" =~ ^48\;5\;[[:digit:]]+$ ]]; then
                    code_off_parts+=( '49' )
                else
                    reset_to_default='YES'
                fi
            else
                echo -e "echo_color: Invalid color name or code: [$1]." >&2
                return 1
            fi
            ;;
        esac
        shift
    done
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

    if [[ -n "$show_examples" ]]; then
        [[ "$debug" ]] && echo -E "Showing examples." >&2
        show_colors "$debug"
        return 0
    fi

    local code_on code_off newline_flag full_output
    code_on="$( join_str ';' "${code_on_parts[@]}" )"
    if [[ -n "$reset_to_default" ]]; then
        code_off='0'
    else
        code_off="$( join_str ';' "${code_off_parts[@]}" )"
    fi
    if [[ -n "$without_newline" ]]; then
        newline_flag='-n'
    fi
    full_output="\033[${code_on}m${message}\033[${code_off}m"
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
                    echo -n ' '
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
                printf ' '
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
