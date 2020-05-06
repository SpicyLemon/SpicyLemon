#!/bin/bash
# This file contains various functions for printing things to your terminal with colors.
# This file is meant to be sourced to add the functions to your environment.
#
# File contents:
#   echo_color  ------> Outputs a message using a specific color code.
#   show_colors  -----> Outputs color names and examples.
#   echo_red  --------> Shortcut for echo_color red --.
#   echo_green  ------> Shortcut for echo_color green --.
#   echo_yellow  -----> Shortcut for echo_color yellow --.
#   echo_blue  -------> Shortcut for echo_color blue --.
#   echo_cyan  -------> Shortcut for echo_color cyan --.
#   echo_bold  -------> Shortcut for echo_color bold --.
#   echo_underline  --> Shortcut for echo_color underline --.
#   echo_debug  ------> Shortcut for echo_color debug --.
#   echo_info  -------> Shortcut for echo_color info --.
#   echo_warn  -------> Shortcut for echo_color warn --.
#   echo_error  ------> Shortcut for echo_color error --.
#   echo_success  ----> Shortcut for echo_color success --.
#   echo_good  -------> Shortcut for echo_color good --.
#   echo_bad  --------> Shortcut for echo_color bad --.
#

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" != 'YES' ]]; then
    >&2 cat << EOF
This script is meant to be sourced instead of executed.
Please run this command to enable the functionality contained in within: $( printf '\033[1;37msource %s\033[0m' "$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )" )
EOF
    exit 1
fi
unset sourced

echo_color () {
    local usage
    usage="$( cat << EOF
echo_color - Makes it easier to output things in colors.

Usage: echo_color <paramters> -- <message>
    Any number of parameters can be provided in any order.
    The parameters must be followed by a --.
    Everything that follows the -- is considered part of the message to output.

Valid Parameters: <name> <color code> -n -N --explain --examples
    <name> can be one of the following:
            Text (foreground) colors:
                black      red            blue        green
                dark-gray  light-red      light-blue  light-green
                light-gray magenta        cyan        yellow
                white      light-magenta  light-cyan  light-yellow
            Background colors:
                bg-black      bg-red            bg-blue        bg-green
                bg-dark-gray  bg-light-red      bg-light-blue  bg-light-green
                bg-light-gray bg-magenta        bg-cyan        bg-yellow
                bg-white      bg-light-magenta  bg-light-cyan  bg-light-yellow
            Effects:
                bold  dim  underline  strikethrough  reversed
            Special Formats:
                debug  info  warn  error  success  good  bad
    <color code> can be one or more numerical color code numbers separated by semicolons or spaces.
            Values are delimited with semicolons and placed between "<esc>[" and "m" for output.
            Examples: "31" "38 5 200" "93;41" "2 38;5;141 48;5;230"
            This page is a good resource: https://misc.flogisoft.com/bash/tip_colors_and_formatting
            Spaces are converted to semicolons for the actual codes used.
    -n signifies that you do not want a trailing newline added to the output.
    -N signifies that you DO want a trailing newline added to the output. This is the default behavior.
            If both -n and -N are provided, whichever is latest in the paramters is used.
    --explain will cause the begining and ending escape codes to be output via stderr.
    --examples will cause the other parameters to be ignored, and instead output a set of examples.
            See the  show_colors  function.

Examples:
    > echo_color underline -- "This is underlined."
    \033[4mThis is underlined\033[24m

    > echo_color bold yellow bg-light-red -- "Would anyone like a hotdog?"
    \033[1;33;101mWould anyone like a hotdog?\033[21;22;39;49m

    > echo_color light-green -- This is a \$( echo_color reversed -- complex ) message.
    \033[92mThis is a \033[7mcomplex\033[27m message.\033[39m

EOF
)"
    local code_on_parts code_params without_newline explain debug show_examples message
    local code_on code_off_parts code_off format full_output code_param
    code_on_parts=()
    code_params=()
    while [[ "$#" -gt '0' && "$1" != '--' ]]; do
        case "$( printf %s "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help|help)
            printf '%b\n' "$usage"
            return 0
            ;;
        black)                          code_on_parts+=( '30' ); code_params+=( "$1" );;
        red)                            code_on_parts+=( '31' ); code_params+=( "$1" );;
        green)                          code_on_parts+=( '32' ); code_params+=( "$1" );;
        yellow)                         code_on_parts+=( '33' ); code_params+=( "$1" );;
        blue)                           code_on_parts+=( '34' ); code_params+=( "$1" );;
        magenta|pink)                   code_on_parts+=( '35' ); code_params+=( "$1" );;
        cyan|teal)                      code_on_parts+=( '36' ); code_params+=( "$1" );;
        light-gray|light-grey)          code_on_parts+=( '37' ); code_params+=( "$1" );;
        dark-gray|dark-grey)            code_on_parts+=( '90' ); code_params+=( "$1" );;
        light-red|bright-red)           code_on_parts+=( '91' ); code_params+=( "$1" );;
        light-green|bright-green)       code_on_parts+=( '92' ); code_params+=( "$1" );;
        light-yellow|bright-yellow)     code_on_parts+=( '93' ); code_params+=( "$1" );;
        light-blue|bright-blue)         code_on_parts+=( '94' ); code_params+=( "$1" );;
        light-magenta|bright-magenta)   code_on_parts+=( '95' ); code_params+=( "$1" );;
        light-pink|bright-pink)         code_on_parts+=( '95' ); code_params+=( "$1" );;
        light-cyan|bright-cyan)         code_on_parts+=( '96' ); code_params+=( "$1" );;
        light-teal|bright-teal)         code_on_parts+=( '96' ); code_params+=( "$1" );;
        white)                          code_on_parts+=( '97' ); code_params+=( "$1" );;
        bg-black)                           code_on_parts+=( '40' ); code_params+=( "$1" );;
        bg-red)                             code_on_parts+=( '41' ); code_params+=( "$1" );;
        bg-green)                           code_on_parts+=( '42' ); code_params+=( "$1" );;
        bg-yellow)                          code_on_parts+=( '43' ); code_params+=( "$1" );;
        bg-blue)                            code_on_parts+=( '44' ); code_params+=( "$1" );;
        bg-magenta|bg-pink)                 code_on_parts+=( '45' ); code_params+=( "$1" );;
        bg-cyan|bg-teal)                    code_on_parts+=( '46' ); code_params+=( "$1" );;
        bg-light-gray|bg-light-grey)        code_on_parts+=( '47' ); code_params+=( "$1" );;
        bg-dark-gray|bg-dark-grey)          code_on_parts+=( '100' ); code_params+=( "$1" );;
        bg-light-red|bg-bright-red)         code_on_parts+=( '101' ); code_params+=( "$1" );;
        bg-light-green|bg-bright-green)     code_on_parts+=( '102' ); code_params+=( "$1" );;
        bg-light-yellow|bg-bright-yellow)   code_on_parts+=( '103' ); code_params+=( "$1" );;
        bg-light-blue|bg-bright-blue)       code_on_parts+=( '104' ); code_params+=( "$1" );;
        bg-light-magenta|bg-bright-magenta) code_on_parts+=( '105' ); code_params+=( "$1" );;
        bg-light-pink|bg-bright-pink)       code_on_parts+=( '105' ); code_params+=( "$1" );;
        bg-light-cyan|bg-bright-cyan)       code_on_parts+=( '106' ); code_params+=( "$1" );;
        bg-light-teal|bg-bright-teal)       code_on_parts+=( '106' ); code_params+=( "$1" );;
        bg-white)                           code_on_parts+=( '107' ); code_params+=( "$1" );;
        bold)               code_on_parts+=( '1' ); code_params+=( "$1" );;
        dim)                code_on_parts+=( '2' ); code_params+=( "$1" );;
        underline)          code_on_parts+=( '4' ); code_params+=( "$1" );;
        reverse|reversed)   code_on_parts+=( '7' ); code_params+=( "$1" );;
        strikethrough)      code_on_parts+=( '9' ); code_params+=( "$1" );;
        debug)          code_on_parts+=( '96;100' );   code_params+=( "$1" );;
        info)           code_on_parts+=( '97;100' );   code_params+=( "$1" );;
        warn|warning)   code_on_parts+=( '93;100' );   code_params+=( "$1" );;
        error)          code_on_parts+=( '1;91;100' ); code_params+=( "$1" );;
        success)        code_on_parts+=( '92;100' );   code_params+=( "$1" );;
        good)           code_on_parts+=( '97;42' );    code_params+=( "$1" );;
        bad)            code_on_parts+=( '1;97;41' );  code_params+=( "$1" );;
        -n) without_newline='YES';;
        -N) without_newline=;;
        --explain) explain='YES';;
        -d|--debug) debug='--debug';;
        --examples) show_examples='--examples';;
        *)
            if [[ "$1" =~ ^[[:digit:]]+(([[:space:]]+|\;)[[:digit:]]+)*$ ]]; then
                code_on_parts+=( "$( printf %s "$1" | sed -E 's/[[:space:]]+/;/g' )" )
                code_params+=( "$1" )
            else
                printf 'echo_color: Invalid parameter: [%s].\n' "$1" >&2
                return 1
            fi
            ;;
        esac
        shift
    done

    if [[ -n "$show_examples" ]]; then
        [[ "$debug" ]] && printf 'Showing examples.\n' >&2
        show_colors $debug
        return 0
    fi

    if [[ "$1" != '--' ]]; then
        printf 'No -- separator found.\n' >&2
        return 1
    fi
    shift
    message="$@"
    # Allowing for an empty message because:
    #   a) It won't really hurt anything.
    #   b) It makes calling this function with user-defined input easier and nicer.
    #   c) It'll make it easier to expand functionality to allow messages to be piped in.

    if [[ -n "$without_newline" ]]; then
        format='%b'
    else
        format='%b\n'
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

    if [[ -n "$debug" || -n "$explain" ]]; then
        {
            if [[ -n "$debug" ]]; then
                printf '   Code params:'
                printf ' [%s]' "${code_params[@]}"
                printf '\n'
            fi
            printf '  Opening code: [%s] -> [\\033[%sm].\n' "$code_on" "$code_on"
            [[ -n "$debug" ]] && printf '       Message: [%s].\n' "$message"
            printf '  Closing code: [%s] -> [\\033[%sm].\n' "$code_off" "$code_off"
            [[ -n "$debug" ]] && printf 'Adding newline: [%s].\n' "$( [[ -n "$without_newline" ]] && printf 'NO' || printf 'yes' )"
            [[ -n "$debug" ]] && printf '        Format: [%s].\n' "$format"
            [[ -n "$debug" ]] && printf 'Without interpretation: [%s].\n' "$full_output"
            [[ -n "$debug" ]] && printf '   With interpretation: [%b].\n' "$full_output"
        } >&2
    fi
    printf "$format" "$full_output"
}

# Displays examples of some color codes
# Usage: show_colors [-v|--verbose] [-d|--debug] [-c|--combos]
show_colors () {
    usage="$( cat << EOF
show_colors - Outputs a bunch of color examples.

Usage: show_colors [-c|--combos] [--256] [-v|--verbose] [-d|--debug]
    -c or --combos will add a few sections that show different color code combinations.
    --256 will add sections showing all the extended colors.
    -v or --verbose will add the uninterpreted output string as output to stderr
                    prior to the interpreted output being sent to stdout.
    -d or --debug adds to --verbose by passing the --debug flag to calls made to echo_color.
                  Be ready for a wall of text being sent to stderr.

EOF
)"
    local debug verbose show_combos show_256 text_colors background_colors effects special_formats fg_codes bg_codes effect_codes output
    while [[ "$#" -gt '0' ]]; do
        case "$( printf '%s' "$1" | tr '[:upper:]' '[:lower:]' )" in
        -h|--help)
            printf '%s\n' "$usage"
            return 0
            ;;
        -v|--verbose)   verbose='--verbose' ;;
        -d|--debug)     debug='--debug' ;;
        -c|--combos)    show_combos='--combos' ;;
        --256)          show_256='--256' ;;
        *)
            printf 'show_colors: Unknown parameter: [%s].\n' "$1" >&2
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
        debug  info  warn  error  success  good  bad
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
            printf '%s:\n' "$title"
            i=0
            for color in "$@"; do
                i=$(( i + 1 ))
                printf ' '
                if [[ "$i" -eq '1' ]]; then
                    printf '   '
                fi
                printf "%${name_width}s:[" "$color"
                echo_color -n "$color" $debug -- " Example "
                printf ']'
                if [[ "$i" -eq "$per_line" ]]; then
                    i=0
                    printf '\n'
                fi
            done
            [[ "$i" -ne '0' ]] && printf '\n'
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
            printf '%s:\n' "$title"
            for fg_code in "${fg_codes[@]}"; do
                printf '    '
                for bg_code in "${bg_codes[@]}"; do
                    code="${added_code}${fg_code};${bg_code}"
                    printf ' %-9b' "\033[${code}m${pad}${code}${pad}\033[0m"
                done
                printf '\n'
            done
        }

        output_256_section () {
            local title base row column code
            title="$1"
            base="$2"
            make_color_piece () {
                printf '\033[%d;5;%dm %3d \033[0m' "$base" "$1" "$1"
            }
            printf '%s: %d;5;\n' "$title" "$base"
            printf '  '
            for code in $( seq 0 15 ); do
                printf '%s' "$( make_color_piece "$code" )"
            done
            printf '\n'
            for row in $( seq 0 11 ); do
                printf '  '
                for column in $( seq 0 5 ); do
                    code=$(( row * 6 + column + 16))
                    printf '\033[%d;5;%dm %3d \033[0m' "$base" "$code" "$code"
                done
                printf '     '
                for column in $( seq 0 5 ); do
                    code=$(( row * 6 + column + 88))
                    printf '\033[%d;5;%dm %3d \033[0m' "$base" "$code" "$code"
                done
                printf '     '
                for column in $( seq 0 5 ); do
                    code=$(( row * 6 + column + 160))
                    printf '\033[%d;5;%dm %3d \033[0m' "$base" "$code" "$code"
                done
                printf '\n'
            done
            printf '  '
            for code in $( seq 232 256 ); do
                printf '%s' "$( make_color_piece "$code" )"
            done
            printf '\n'
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
        if [[ -n "$show_256" ]]; then
            output_256_section 'Text colors' '38'
            output_256_section 'Background colors' '48'
        fi
    )"
    if [[ -n "$debug" || -n "$verbose" ]]; then
        {
            printf 'Without Interpretation:\n'
            printf '%s\n' "$output" | escape_escapes
            printf 'With Interpretation:\n'
        } >&2
    fi
    printf '%b\n' "$output"
}

# Usage: echo_red <message>
echo_red () {
    echo_color 'red' -- "$@"
}

# Usage: echo_green <message>
echo_green () {
    echo_color 'green' -- "$@"
}

# Usage: echo_yellow <message>
echo_yellow () {
    echo_color 'yellow' -- "$@"
}

# Usage: echo_blue <message>
echo_blue () {
    echo_color 'blue' -- "$@"
}

# Usage: echo_cyan <message>
echo_cyan () {
    echo_color 'cyan' -- "$@"
}

# Usage: echo_bold <message>
echo_bold () {
    echo_color 'bold' -- "$@"
}

# Usage: echo_underline <message>
echo_underline () {
    echo_color 'underline' -- "$@"
}

# Usage: echo_debug <message>
echo_debug () {
    echo_color 'debug' -- "$@"
}

# Usage: echo_info <message>
echo_info () {
    echo_color 'info' -- "$@"
}

# Usage: echo_warn <message>
echo_warn () {
    echo_color 'warn' -- "$@"
}

# Usage: echo_error <message>
echo_error () {
    echo_color 'error' -- "$@"
}

# Usage: echo_success <message>
echo_success () {
    echo_color 'success' -- "$@"
}

# Usage: echo_good <message>
echo_good () {
    echo_color 'good' -- "$@"
}

# Usage: echo_bad <message>
echo_bad () {
    echo_color 'bad' -- "$@"
}

return 0
