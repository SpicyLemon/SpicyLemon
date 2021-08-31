#!/bin/bash
# This file contains several functions for generating palettes for use in terminal color escape sequences.
# This file can be sourced to add the functions to your environment.
# This file can also be executed to run the palette_vector_generate function without adding functions to your environment.
#
# File contents:
#   palette_generators  -------> Just outputs some usage information on the stuff in here.
#   palette_vector_generate  --> Generates a random palette.
#   palette_vector_no_wrap  ---> Generates a 6 color palette vector that doesn't wrap.
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

Generate a palette vector that doesn't wrap in the cube: palette_vector_no_wrap [<choice>]
    The <choice> is optional but must be a number from 0 to 295 (inclusive).
    If not provided, one will be chosen randomly.

Generate a random palette vector: palette_vector_random [<start>] [<dx>] [<dy>] [<dz>]
    Any arguments that aren't provided will have random numbers generated for them.
    To set a later argument while still getting a random earlier argument, supply '' for the random one.
    E.g. the command  palette_vector_random '' 1 0 0  will pick a random starting point and use dx=1, dy=0, dz=0.
    dx, dy, and dz are chosen from -2, -1, 0, 1, and 2.

An easy way to view these palettes in action is with the test_palette function.
    test_palette \$( palette_vector_random )
    If you don't have the test_palette function, source the hrr.sh file to add it.

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

# Usage: palette_vector_no_wrap [<choice>]
# The <choice> is optional and must be a number from 0 to 295 (inclusive).
# If not provided, one will randomly be chosen for you.
# This basically picks a 6-cell vector through the cube without wrapping.
palette_vector_no_wrap () {
    local choice
    if [[ -n "$1" ]]; then
        if [[ "$1" == '-h' || "$1" == '--help' ]]; then
            printf 'Usage: palette_vector_no_wrap [<choice>]\n' >&2
            return 0
        fi
        if [[ "$1" =~ [^[:digit:]] || "$1" -gt '295' ]]; then
            printf 'Invalid choice: [%s]. Must be a number between 0 and 295 (inclusive).\n' "$1" >&2
            return 1
        fi
        choice=$1
    else
        choice=$(( RANDOM%296 ))
    fi
    # Enumeration:
    #   Vectors with exactly 1 changing dimension:
    #       Any point on a side can be a start = 36 per side * 6 sides = 216
    #   Vectors with exactly 2 changing dimensions:
    #       Any point on an edge can be a start = 6 per edge * 12 edges = 72
    #   Vectors with exactly 3 changing dimensions:
    #       Any corner can be a start = 8.
    #   There are 216 + 72 + 8 = 296 different 6-cell vectors through a 6x6x6 cube (without wrapping).
    local val dx dy dz
    [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d]\n' "$choice" >&2
    if [[ "$choice" -ge '0' && "$choice" -le '215' ]]; then
        # These are the 1d vectors. There are 216 of them.
        # There are two aspects: the side, and where to start on the side.
        # The side also dictates the direction of the vector.
        local b side
        b=$(( choice / 6 ))     # 0 to 35: max choice (here) is 215. 215 / 6 = 35.
        side=$(( choice % 6 ))  # 0 to 5
        [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d] = side: [%d], b: [%d]\n' "$choice" "$side" "$b" >&2
        case "$side" in
        # For West or East, b represents y and z, and x is 0 for West, 5 for East.
        #   Let s = 16 + x = either 16 for West or 21 for East.
        #   Let y = b / 6, z = b % 6, and then b = y + 6 * z.
        #   Then
        #       val = 16 + x + 6 * y + 36 * z
        #       val = s + 6 * y + 36 * z
        #       val = s + 6 * (y + 36 * z)
        #       val = s + 6 * b
            0) # West side: start with x = 0, constant y and z.
                dx=1
                dy=0
                dz=0
                val=$(( b * 6 + 16 ))
                ;;
            1) # East side: start with x = 5, constant y and z.
                dx=-1
                dy=0
                dz=0
                val=$(( b * 6 + 21 ))
                ;;
        # For North or South, b represents x and z, and y is 0 for North, 5 for South.
        #   Let s = 16 + 6 * y = either 16 for North or 46 for South.
        #   Let x = b / 6, z = b % 6, and then b = x + 6 * z.
        #   Then
        #       val = 16 + x + 6 * y + 36 * z
        #       val = s + b / 6 + 36 * (b % 6)
            2) # North side: start with y = 0, constant x and z
                dx=0
                dy=1
                dz=0
                val=$(( b % 6 * 36 + b / 6 + 16 ))
                ;;
            3) # South side: start with y = 5, constant x and z
                dx=0
                dy=-1
                dz=0
                val=$(( b % 6 * 36 + b / 6 + 46 ))
                ;;
        # For Up or Down, b represents x and y, and z is 0 for Down, 5 for Up.
        #   Let s = 16 + 36 * z = either 16 for Down or 196 for Up.
        #   Let x = b / 6, y = b % 6, and then b = x + 6 * y
        #   Then
        #       val = 16 + x + 6 * y + 36 * z
        #       val = s + x + 6 * y
        #       val = s + b
            4) # Down side: start with z = 0, constant x and y
                dx=0
                dy=0
                dz=1
                val=$(( b + 16 ))
                ;;
            5) # Up side: start with z = 5, constant x and y
                dx=0
                dy=0
                dz=-1
                val=$(( b + 196 ))
                ;;
            *)
                printf 'Bug in code palette_vector_no_wrap-1d: choice: [%s], side: [%s], b: [%s]\n' "$choice" "$side" "$b" >&2
                return 10
                ;;
        esac
    elif [[ "$choice" -ge '216' && "$choice" -le '287' ]]; then
        # These are the 2d vectors. There are 72 of them.
        # There are 2 aspects: The edge and the location on the edge to start.
        # The edge has 3 aspects to it: The constant dimension and the sign of change in each of the other two dimensions.
        # There are 6 locations to start at on each edge. Call it p, and it'll have a value from 0 to 5.
        # There are 3 constant dimensions. Call it c, and it'll have a value of 0, 1, or 2 for z, y, or x respectively)
        # Each rate of change has two options. Call them d1 and d2, and they'll have values 0 or 1.
        # But to be tricky, I want consecutive choice numbers to behave a certain way.
        # An odd numbered choice should be the reverse of the one before it. E.g. 217 is the reverse of 216.
        # Each consecutive even number should be the next cell on the given edge.
        # Once all 6 cells have gone there and back (12 total), then keep the constant dimension the same but rotate the vector 90 degrees.
        # Go along the edge there and back again (another 12, for 24 total by now).
        # Then finally move to the next constant dimension. 3 constant dimensions * 24 each = 72.
        # So, from least significance to most, here's how that 72 breaks down:
        #   {c: 0 to 3}{db: 0 to 1}{p: 0 to 5}{da: 0 to 1}
        # To extract that, you start with the following equations:
        #   i = choice - 216
        #   da = i % 2
        #   daq = i / 2
        #   p = daq % 6
        #   pq = daq / 6
        #   db = pq % 2
        #   dbq = pq / 2
        #   c = dbq % 3
        # Those can be simplified down to these:
        #   i = choice - 216
        #   da = i % 2
        #   p = i / 2 % 6
        #   db = i / 12 % 2
        #   c = i / 24
        # Then, in order to get the down and back behavior I want for consecutive numbers, da and db must be transformed to get d1 and d2.
        #   If you think of da and db as a 2 bit number (0 to 4), it has ordering 00, 01, 10, 11 (0123).
        #   But we want opposites to be consecutive, so 00, 11, 10, 01 (0321) would be a better order.
        #   So we need to "rotate" it left 1 and then "flip" it.
        #   start:  d = da * 2 + db
        #   rotate: d = (d + 3) % 4 = (da * 2 + db) % 4
        #   flip:   d = 3 - d = 3 - (da * 2 + db) % 4
        #   Then, to skip assignment of the da and db variables, we can do this:
        #       d = 3 - (i / 12 % 2 * 2 + i % 2 + 3) % 4
        #   Now, you can pull d1 and d2 out of it:
        #       d1 = d / 2
        #       d2 = d % 2
        # The starting value then has the coordinates where two of the dimensions must be 0 or 5
        #   and the constant dimension has value p.
        #   Whether each is 0 or 5 depends on d1 and d2 using simply d1*5 and d2*5.
        #   Which dimension gets d1*5, which gets d2*5, and which gets p is dictated by the value of c.
        #   Then, val = 16 + x + 6 * y + 36 * z is used.
        #   But below, since the handling of c is hard-coded, the 5* multiplier is sometimes baked into the equation.
        #   6*5 = 30, and 36*5 = 180. That's where those two numbers come from below.
        local i p d d1 d2 c
        i=$(( choice - 216 ))                   # 0 to 71
        p=$(( i / 2 % 6 ))                      # 0 to 5
        d=$(( 3 - (i/12%2*2 + i%2 + 3) % 4 ))   # 0 to 3
        d1=$(( d / 2 ))                         # 0 to 1
        d2=$(( d % 2 ))                         # 0 to 1
        c=$(( i / 24 ))                         # 0 to 2: 71/24 = 2 (r 23)
        [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d] = i: [%2d], p: [%d], d1: [%d], d2: [%d], c: [%d]\n' "$choice" "$i" "$p" "$d1" "$d2" "$c" >&2
        case "$c" in
            0) # Constant z. Edges NW, NE, SW, SE.
                dx=$(( 1 - d1*2 ))
                dy=$(( 1 - d2*2 ))
                dz=0
                val=$(( 16 + 5*d1 + 30*d2 + 36*p ))
                ;;
            1) # Constant y. Edges ND, NU, SD, SU.
                dx=$(( 1 - d1*2 ))
                dy=0
                dz=$(( 1 - d2*2 ))
                val=$(( 16 + 5*d1 + 6*p + 180*d2 ))
                ;;
            2) # Constant x. Edges WD, WU, ED, EU.
                dx=0
                dy=$(( 1 - d1*2 ))
                dz=$(( 1 - d2*2 ))
                val=$(( 16 + p + 30*d1 + 180*d2 ))
                ;;
            *)
                printf 'Bug in code palette_vector_no_wrap-2d: choice: [%s], i: [%s], p: [%s], d1: [%s], d2: [%s], c: [%s]\n' "$choice" "$i" "$p" "$d1" "$d2" "$c" >&2
                return 10
                ;;
        esac
    elif [[ "$choice" -ge '288' && "$choice" -le '295' ]]; then
        # These are the 3d vectors. There are 8 of them.
        # There's only one aspect: the corner.
        # That dictates both the starting value and the direction vector.
        # Here again, I want odd numbers to be the reverses of the numbers before them.
        # And the math for doing that here is pretty gnarly.
        # So, since there's only 8, I'm just going to hard code things.
        # Set x y and z as either 0 or 1.
        # The change in a dimension then is 1 - that value * 2, e.g. dx = 1 - x*2
        # And then the starting coordinates should be either 0 or 5, so multiply them by 5 for that.
        # So then val = 16 + X + 6*Y + 36*Z = 16 + 5*x + 6*5*y + 36*5*z = 16 + 5*x + 30*y + 180*z.
        local corner x y z
        corner=$(( choice - 288 ))  # 0 to 7
        case "$corner" in
            0) x=0; y=0; z=0;;
            1) x=1; y=1; z=1;;
            2) x=0; y=0; z=1;;
            3) x=1; y=1; z=0;;
            4) x=0; y=1; z=0;;
            5) x=1; y=0; z=1;;
            6) x=0; y=1; z=1;;
            7) x=1; y=0; z=0;;
            *)
                printf 'Bug in code palette_vector_no_wrap-3d: choice: [%s], corner: [%s]\n' "$choice" "$corner" >&2
                return 10
                ;;
        esac
        [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d] = corner: [%d] = (%d, %d, %d)\n' "$choice" "$corner" "$x" "$y" "$z" >&2
        val=$(( 16 + 5*x + 30*y + 180*z ))
        dx=$(( 1 - x*2 ))
        dy=$(( 1 - y*2 ))
        dz=$(( 1 - z*2 ))
    else
        [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d] = Unknown\n' "$choice" >&2
        # This is a bug because previous validation combined with the if/elif chain should have caught everything by now.
        printf 'Bug in code palette_vector_no_wrap-xd: choice: [%s]\n' "$choice" "$corner" >&2
        return 10
    fi
    [[ -n "$PVNW_DEBUG" ]] && printf 'Choice: [%3d] => palette_vector_generate "%d" "%d" "%d" "%d"\n' "$choice" "$val" "$dx" "$dy" "$dz" >&2
    palette_vector_generate "$val" "$dx" "$dy" "$dz"
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
    [[ -n "$PVR_DEBUG" ]] && printf 'palette_vector_generate "%d" "%d" "%d" "%d"\n' "$val" "$dx" "$dy" "$dz" >&2
    palette_vector_generate "$val" "$dx" "$dy" "$dz"
}

# Usage: val="$( __palette_validate_color "<field name>" "<value>" )" || return $?
# If not valid, an error message is printed to stderr, 16 is printed to stdout, and the exit code will be 1 (false).
# Otherwise, the provided value is printed to stdout and the exit code will be 0 (true).
# A valid color is an integer between 16 and 231 inclusive.
__palette_validate_color () {
    if [[ "$2" =~ [^[:digit:]] || "$2" -lt '16' || "$2" -gt '231' ]]; then
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
# The color 16 resides at (0,0,0); 21 at (5,0,0); 46 at (0,5,0); 196 at (0,0,5); and 231 is at (5,5,5).
#
# To convert from a cell value to its coordinates:
#   x = (val - 16) % 6
#   y = (val - 16) % 36 / 6
#   z = (val - 16) / 36
# To convert from coordinates to a cell value:
#   val = 16 + x + 6 * y + 36 * z
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
