#!/bin/bash

# Usage: newday.sh [<from day>]
#   <from day> can be one of these:
#       Just a number (e.g. 8 or 15) to represent day-##a (e.g. day-08a or day-15a).
#       A number and letter to represent the day and part (e.g. 8b or 15b).
#   if ommitted, the highest numbered day's 'a' solution will be used.

# Make sure we're in the right directory (the one containing this script).
curDir="$( pwd )"
scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
if [[ "$curDir" != "$scriptDir" ]]; then
    cd "$scriptDir"
fi

# Figure out and set the needed directory names
lastDayDir="$( ls | grep '^day' | sort -r | head -n 1 )"
if [[ -n "$lastDayDir" ]]; then
    lastDayNum="$( sed 's/[^[:digit:]]//g; s/^0//;' <<< "$lastDayDir" )"
    lastDayPart="${lastDayDir: -1}"
    if [[ -n "$1" ]]; then
        fromDayNum="$( sed 's/[^[:digit:]]//g; s/^0//;' <<< "$1" )"
        fromDayPart="${1: -1}"
        if [[ "$fromDayPart" =~ [[:digit:]] ]]; then
            fromDayPart='a'
        fi
    else
        fromDayNum="$lastDayNum"
        fromDayPart='a'
    fi
    if [[ "$lastDayPart" == 'a' ]]; then
        newDayNum="$lastDayNum"
        newDayPart='b'
    else
        newDayNum=$(( lastDayNum +1 ))
        newDayPart='a'
    fi
else
    fromDayNum='0'
    fromDayPart=''
    newDayNum='1'
    newDayPart='a'
fi
fromDayDir="$( printf 'day-%02d%s' "$fromDayNum" "$fromDayPart" )"
newDayDir="$( printf 'day-%02d%s' "$newDayNum" "$newDayPart" )"

mkdir "$newDayDir"
ec=$?
if [[ "$ec" -ne '0' ]]; then
    exit $ec
fi
if [[ "$curDir" == "$scriptDir" ]]; then
    fullNewDir="$newDayDir"
else
    fullNewDir="$scriptDir/$newDayDir"
fi
printf 'New directory created: %s/\n' "$fullNewDir"
if [[ "$fromDayNum" -eq "$newDayNum" && "$fromDayPart" == 'a' && "$newDayPart" == 'b' ]]; then
    cp $fromDayDir/*.go "$newDayDir/"
    mv "$newDayDir/$fromDayDir.go" "$newDayDir/$newDayDir.go"
    cp $fromDayDir/*.input "$newDayDir/" 2> /dev/null
    printf '  Content copied from: %s/\n' "$fromDayDir"
else
    cp template.go "$newDayDir/$newDayDir.go"
    printf '  Content copied from: template\n'
fi
printf '         Program file: %s/%s\n' "$fullNewDir" "$newDayDir.go"
exit 0
