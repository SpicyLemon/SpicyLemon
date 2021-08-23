#!/bin/bash
# This file contains several functions for generating palettes for use in terminal color escape sequences.
# This file can be sourced to add the functions to your environment.
# This file can also be executed to run the palette_vector_generate function without adding functions to your environment.
#
# File contents:
#   palette_generators  -------> Just outputs some usage information on the stuff in here.
#   palette_vector_generate  --> Generates a random palette.
#   palette_vector_random  ----> Picks random numbers for palette generation and provides them to palette_vector_generate.
#
# "private" contents:
#   __palette_validate_color  --> Validates a color value.
#   __palette_validate_d  ------> Validates a change-in-color value.
#   __palette_color_move  ------> Changes a palette color using provided movements.
#

# The numbers generated in here are for defining terminal colors.
# For text colors, use \033[38;5;<num>m
# For background colors, use \033[48;5;<num>m
# In both of those, replace <num> with a number generated in here.
# For example:
#   palette_vector_generate 124 1 0 0
# outputs:
#   124 125 126 127 128 129
# To view them as colored text, you could do the following:
#   for c in $( palette_vector_generate 124 1 0 0 ); do printf '\033[38;5;%dm color %d \033[0m\n' "$c" "$c"; done
# To view them as background colors, you could do this:
#   for c in $( palette_vector_generate 124 1 0 0 ); do printf '\033[48;5;%dm color %d \033[0m\n' "$c" "$c"; done

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Usage: palette_generators
# This is really only here because some other stuff expects this file to have a function named after it.
palette_generators () {
    cat << EOF
Generate a palette vector: palette_vector_generate <start> <dx> <dy> <dz>
    All arguments are required.
    <start> must be 16 to 231 inclusive.
    <dx>, <dy>, and <dz> must be integers (positive or negative).

Generate a random palette vector: palette_vector_random [<start>] [<dx>] [<dy>] [<dz>]
    Any arguments that aren't provided will have random numbers generated for them.
    To set a later argument while still getting a random earlier argument, supply '' for the random one.
    E.g. the command  palette_vector_random '' 1 0 0  will pick a random starting point and use dx=1, dy=0, dz=0.
    dx, dy, and dz are chosen from -2, -1, 0, 1, and 2.

EOF
}

# Usage: palette_vector_generate <start> <dx> <dy> <dz>
palette_vector_generate () {
    local val dx dy dz exit_code palette i d
    exit_code=0
    if [[ "$#" -eq '0' ]]; then
        exit_code=1
    elif [[ "$#" -lt '4' ]]; then
        printf 'Not enough arguments.\n' >&2
        exit_code=1
    elif [[ "$#" -gt '4' ]]; then
        printf 'Too many arguments.\n' >&2
        exit_code=1
    fi
    if [[ "$exit_code" -ne '0' ]]; then
        printf 'Usage: palette_vector_generate <start> <dx> <dy> <dz>\n' >&2
        return $exit_code
    fi
    val="$( __palette_validate_color 'start' "$1" )" || exit_code=$?
    dx="$( __palette_validate_d 'dx' "$2" )" || exit_code=$?
    dy="$( __palette_validate_d 'dy' "$3" )" || exit_code=$?
    dz="$( __palette_validate_d 'dz' "$4" )" || exit_code=$?
    if [[ "$exit_code" -ne '0' ]]; then
        return $exit_code
    fi
    palette=( "$val" )
    for i in $( seq 5 ); do
        val="$( __palette_color_move "$val" "$dx" "$dy" "$dz" )"
        palette+=( "$val" )
    done
    printf '%s\n' "${palette[*]}"
}

# Usage: palette_vector_random [<start>] [<dx>] [<dy>] [<dz>]
# Arguments are positional. Provide an empty string to keep it random while setting a later one.
# E.g. palette_vector_random '' '' 1 1
#   For a random start, random dx, but dy=1 and dz=1.
# The call being made to palette_vector_random will be printed to stderr.
palette_vector_random () {
    local val dx dy dz
    if [[ "$#" -gt 4 || "$*" =~ -h ]]; then
        printf 'Usage: palette_vector_random [<start>] [<dx>] [<dy>] [<dz>]\n' >&2
        return 1
    fi
    # The rest of the provided argument validation is just passed off on palette_vector_generate.
    if [[ -n "$1" ]]; then
        val="$1"
    else
        val="$(( RANDOM%216 + 16 ))"
    fi
    if [[ -n "$2" ]]; then
        dx="$2"
    else
        dx="$(( RANDOM%5 - 2 ))"
    fi
    if [[ -n "$3" ]]; then
        dy="$3"
    else
        dy="$(( RANDOM%5 - 2 ))"
    fi
    if [[ -n "$4" ]]; then
        dz="$4"
    else
        dz="$(( RANDOM%5 - 2 ))"
    fi
    printf 'palette_vector_generate "%d" "%d" "%d" "%d"\n' "$val" "$dx" "$dy" "$dz" >&2
    palette_vector_generate "$val" "$dx" "$dy" "$dz"
}

# Usage: val="$( __palette_validate_color "<field name>" "<value>" )" || return $?
# If not valid, an error message is printed to stderr, 16 is printed to stdout, and the exit code will be 1 (false).
# Otherwise, the provided value is printed to stdout and the exit code will be 0 (true).
# A valid color is an integer between 16 and 231 inclusive.
__palette_validate_color () {
    if [[ "$2" =~ [^[:digit:]] ||  "$2" -lt '16' || "$2" -gt '231' ]]; then
        printf 'Invalid start: [%s]. Must be a number between 16 and 231 (inclusive).\n' "$1" "$2" >&2
        printf '16'
        return 1
    fi
    printf "$2"
    return 0
}

# Usage: val="$( __palette_validate_d "<field name>" "<d value>" )" || return $?
# If not valid, an error message is printed to stderr, 0 is printed to stdout and the exit code will be 1 (false).
# Otherwise, a valid d value is printed to stdout that might be different but equivalent from the provided d value and the exit code will be 2 (true).
# Really, any positive or negative integer is valid.
# Any commas, underscore, or space characters will be removed.
# Any integer is valid as long as the system can hold it as a number. E.g. On a 32 bit system, 2,147,483,648 is too large.
# Since it's being applied to a 6x6x6 cube, mod 6 will be applied to the number.
# Then, if less than -3, 6 will be added, or if greater than 3, 6 will be subtracted.
# The result will then be one of -3, -2, -1, 0, 1, 2, 3.
__palette_validate_d () {
    local d ec
    d="$2"
    ec=0
    [[ "$d" =~ [,_[:space:]] ]] && d="$( sed 's/[,_[:space:]]//g' <<< "$d" )"
    if [[ ! "$d" =~ ^-?[[:digit:]]+$ ]]; then
        printf 'Invalid %s: [%s]: Must be a whole number (positive or negative).\n' "$1" "$2" >&2
        ec=1
    elif [[ "$d" != "$(( d + 0 ))" ]]; then
        # Using string comparison != there since we're checking for overflow and don't want both sides to overflow.
        printf 'Invalid %s: [%s]: Integer out of bounds for system.\n' "$1" "$2" >&2
        ec=1
    else
        # It's valid. Reduce it if needed.
        if [[ "$d" -le '-6' || "$d" -ge '6' ]]; then
            d=$(( d % 6 ))
        fi
        if [[ "$d" -le '-4' ]]; then
            d=$(( d + 6 ))
        elif [[ "$d" -ge '4' ]]; then
            d=$(( d - 6 ))
        fi
    fi
    if [[ "$ec" -ne '0' ]]; then
        d=0
    fi
    printf '%s' "$d"
    return $ec
}

# Usage: __palette_color_move <start> <dx> <dy> <dz>
__palette_color_move () {
    local val d
    val="$1"
    # apply dx: East/West
    if [[ "$2" -lt '0' ]]; then
        # West
        d="$2"
        while [[ "$d" -lt '0' ]]; do
            if [[ "$(( val % 6 ))" -eq '4' ]]; then
                val=$(( val + 5 ))
            else
                val=$(( val - 1 ))
            fi
            d=$(( d + 1 ))
        done
    elif [[ "$2" -gt '0' ]]; then
        # East
        d="$2"
        while [[ "$d" -gt '0' ]]; do
            if [[ "$(( val % 6 ))" -eq '3' ]]; then
                val=$(( val - 5 ))
            else
                val=$(( val + 1 ))
            fi
            d=$(( d - 1 ))
        done
    fi
    # apply dy: North/South
    if [[ "$3" -lt '0' ]]; then
        # North
        d="$3"
        while [[ "$d" -lt '0' ]]; do
            if [[ "$(( (val-16)%36/6 ))" -eq '0' ]]; then
                val=$(( val + 30 ))
            else
                val=$(( val - 6 ))
            fi
            d=$(( d + 1 ))
        done
    elif [[ "$3" -gt '0' ]]; then
        # South
        d="$3"
        while [[ "$d" -gt '0' ]]; do
            if [[ "$(( (val-16)%36/6 ))" -eq '5' ]]; then
                val=$(( val - 30 ))
            else
                val=$(( val + 6 ))
            fi
            d=$(( d - 1 ))
        done
    fi
    # apply dz: Up/Down
    if [[ "$4" -lt '0' ]]; then
        # Down
        d="$4"
        while [[ "$d" -lt '0' ]]; do
            if [[ "$val" -le '51' ]]; then
                val=$(( val + 180 ))
            else
                val=$(( val - 36 ))
            fi
            d=$(( d + 1 ))
        done
    elif [[ "$4" -gt '0' ]]; then
        # Up
        d="$4"
        while [[ "$d" -gt '0' ]]; do
            if [[ "$val" -ge '196' ]]; then
                val=$(( val - 180 ))
            else
                val=$(( val + 36 ))
            fi
            d=$(( d - 1 ))
        done
    fi
    printf '%s' "$val"
}

if [[ "$sourced" != 'YES' ]]; then
    palette_vector_generate "$@"
    exit $?
fi
unset sourced

return 0

# Here are some notes on how this works.
# The command `show_colors --256` (defined in echo_color.sh) has a section with six 6x6 color gradient boxes.
# I turned those into a 6x6x6 cube where, from the bottom up, the upper left corners are 16, 52, 88, 124, 160, 196.
# Bottom
#   16   17   18   19   20   21        88   89   90   91   92   93       160  161  162  163  164  165
#   22   23   24   25   26   27        94   95   96   97   98   99       166  167  168  169  170  171
#   28   29   30   31   32   33       100  101  102  103  104  105       172  173  174  175  176  177
#   34   35   36   37   38   39       106  107  108  109  110  111       178  179  180  181  182  183
#   40   41   42   43   44   45       112  113  114  115  116  117       184  185  186  187  188  189
#   46   47   48   49   50   51       118  119  120  121  122  123       190  191  192  193  194  195
#
#   52   53   54   55   56   57       124  125  126  127  128  129       196  197  198  199  200  201
#   58   59   60   61   62   63       130  131  132  133  134  135       202  203  204  205  206  207
#   64   65   66   67   68   69       136  137  138  139  140  141       208  209  210  211  212  213
#   70   71   72   73   74   75       142  143  144  145  146  147       214  215  216  217  218  219
#   76   77   78   79   80   81       148  149  150  151  152  153       220  221  222  223  224  225
#   82   83   84   85   86   87       154  155  156  157  158  159       226  227  228  229  230  231
#                                                                                                   Top
#
# I chose 6 numbers for a palette because of that cube size.
#
# To create a palette, you basically pick a starting value and a cartesian vector (dx, dy, dz).
# The starting value is the 1st palette number, then you follow the vector 5 times to get the other 5 numbers.
#
# Directions then come as follows:
#   West  = -dx: If edge, val+=5, else val-=1
#       Edges: val % 6 == 4
#   East  = +dx: If edge, val-=5, else val+=1
#       Edges: val % 6 == 3
#   North = -dy: If edge, val+=30, else val-=6
#       Edges: (val - 16) % 36 / 6 == 0
#   South = +dy: If edge, val-=30, else val+=6
#       Edges: (val - 16) % 36 / 6 == 5
#   Down  = -dx: If edge, val+=180, else val-=36
#       Edges: val <= 51
#   Up    = +dz: If edge, val-=180, else val+=36
#       Edges: val >= 196
#
# There's a little confusion in there because the y-axis numbers are reversed from what they'd be if in quadrant 1 on a graph.
#    Basically, North and South are backwards from what they'd be if graphed normally.
#
# The algorithm above applies each single cardinal vector move one at a time.
#   I went with that because I had already figured out how to know if a value is an edge,
#   but I haven't figured out how to know if a move has gone past an edge.
#   For example, if applying dy = 2 to value 188, you should end on 164.
#   The single step way does this:
#       val=188, dy=2. It isn't an edge. val+=6
#       val=194, dz=1. It is an edge. val-=30
#       val=164, dz=0. Done.
#   The whole step way would do this:
#       val=188, dz=2. val+=6*dz.
#       val=200.
#       Compensage for any edges passed.
#   It's that last part. I'm not entirely sure how to know when an edge has been passed for each direction, and by how much.
#   For the x axis, I think the west edge is something like val - val % 6. For the West edge, val + 5 - val % 6.
#   But the other axes aren't as straight forward since the possible values aren't continous.
#   There's got to be a formula though.
#   Without actually trying to represent the cube, I'm not even sure if the more mathematical way will be more efficient.
