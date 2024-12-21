#!/bin/bash
# Usage: ./clean-points
# It takes the clipboard and finds all point strings: (<num>,<num>) in it.
# It outputs all of them as well as a window that contains all of them + 1 in each direction.

points=( $( pbpaste | grep -Eo '\([[:digit:]]+,[[:digit:]]+\)' | tr ' ' '\n' | sort -u ) )

xmin="$( printf '%s\n' "${points[@]}" | sed -E 's/^\(([[:digit:]]+),([[:digit:]]+)\)$/\1/' | sort -n | head -n 1 )"
xmax="$( printf '%s\n' "${points[@]}" | sed -E 's/^\(([[:digit:]]+),([[:digit:]]+)\)$/\1/' | sort -n | tail -n 1 )"
ymin="$( printf '%s\n' "${points[@]}" | sed -E 's/^\(([[:digit:]]+),([[:digit:]]+)\)$/\2/' | sort -n | head -n 1 )"
ymax="$( printf '%s\n' "${points[@]}" | sed -E 's/^\(([[:digit:]]+),([[:digit:]]+)\)$/\2/' | sort -n | tail -n 1 )"
xmin=$(( xmin - 1 ))
xmax=$(( xmax + 1 ))
ymin=$(( ymin - 1 ))
ymax=$(( ymax + 1 ))

printf 'There are %d points in a window that is %d x %d.\n' "${#points[@]}" "$(( xmax - xmin + 1 ))" "$(( ymax - ymin + 1 ))"
printf ' --lines %s %s\n' "'$( cat <<< "${points[@]}" )'" "'$( printf '(%d,%d)-(%d,%d)' "$(( xmin + 1 ))" "$(( ymin + 1 ))" "$(( xmax + 1 ))" "$(( ymax + 1 ))" )'"
printf ' --lines %s %s\n' "'$( cat <<< "${points[@]}" )'" "'$( printf '(%d,%d)-(%d,%d)' "$(( xmin + 1 ))" "$(( ymin + 1 ))" "$(( xmax + 1 ))" "$(( ymax + 1 ))" )'" | pbcopy
printf '     (copied to clipboard)\n'
