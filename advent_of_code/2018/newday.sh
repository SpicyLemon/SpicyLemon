#!/bin/bash

# Make sure we're in the right directory (the one containing this script).
scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$scriptDir"

last="$( ls adventOfCode* | sort | tail -n 1 )"
pd="$( sed -E 's/^adventOfCode0?//; s/p.*$//;' <<< "$last" )"
op="$( sed -E 's/^.*p//; s/\..*$//;' <<< "$last" )"
d=$(( pd + op - 1 ))
p=$(( op % 2 + 1 ))
nd=$(( pd + 1 ))
ap=''
if [[ "$p" -eq '2' ]]; then
    ap='#part2'
fi
printf -v prev '%02dp%d' "$pd" "$op"
printf -v cur '%02dp%d' "$d" "$p"
printf -v next '%02dp%d' "$nd" "$op"
printf -v l '%d%s' "$d" "$ap"
printf -v nf 'adventOfCode%s.html' "$cur"
sed 's/#####/'"$next"'/g; s/####/'"$cur"'/g; s/###/'"$prev"'/g; s/~~~~/'"$l"'/g;' template.html > "$nf"
printf '%s\n' "$nf"
