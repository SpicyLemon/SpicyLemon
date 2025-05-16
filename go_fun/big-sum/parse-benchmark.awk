# This script will parse the cleaned output of the benchmark tests, and reformat it in
# a way that makes it hopefully easier to analyze.
# Example usage: awk -f parse-benchmark.awk <benchmark results file>
# Extra info can be included if you provide the -v verbose=true flag when invoking this script.
#
# See parse-benchmark.sh for details about converting the output of a benchmark run into what this script uses.
# It expects lines like this:
# Positive_Small_Floats 2 Sum2              465498          2521 ns/op
# that have the format "<test group> <arg count> <sum func> <iterations> <time per op>".
#
# To run the benchmarks, use this command.
# $ go test -bench=.
# or
# $ make benchmark

# processSumSet will look at the Sum2, Sum3, Sum4, LastCount, and possibly LastGroup variables
# to create a new data row. If Sum4 is not the best, the dato is added to the Notes variable.
# If doing verbose output, the data row is also printed.
function processSumSet() {
    bestAmt=Sum2;
    secondAmt=0;
    sums[2]=Sum2;
    sums[3]=Sum3;
    sums[4]=Sum4;
    for (i = 3; i <= 4; i++) {
        amt=sums[i];
        if (amt <= bestAmt) {
            secondAmt=bestAmt;
            bestAmt=amt;
        } else if (secondAmt == 0 || amt < secondAmt) {
            secondAmt=amt;
        }
    }

    best="????";
    noteworthy="YES";
    if (bestAmt == secondAmt) {
        best="\033[7;95mTIE \033[0m"; # Reversed and bright magenta for a Tie.
    } else if (bestAmt == Sum2) {
        best="\033[7;93mSum2\033[0m"; # Reversed and bright yellow for Sum2.
    } else if (bestAmt == Sum3) {
        best="\033[7;96mSum3\033[0m"; # Reversed and bright cyan for Sum3.
    } else if (bestAmt == Sum4) {
        best="Sum4"; # Normal for Sum4.
        noteworthy="";
    }

    for (i = 2; i <= 4; i++) {
        if (sums[i] == bestAmt) {
            sums[i]=sprintf("\033[1m%9s\033[0m", sums[i]); # Highlight the largest.
        } else if (sums[i] != secondAmt) {
            sums[i]=sprintf("\033[2m%9s\033[0m", sums[i]); # Gray out the smallest.
        }
        # Leave the seconod amount as normal text.
    }

    diff=secondAmt-bestAmt;
    diffP=diff*100/bestAmt;
    row=sprintf("  %5s  %9s  %9s  %9s  %s  by %9s = %4.1f%%", LastCount, sums[2], sums[3], sums[4], best, diff, diffP);
    if (verbose!="") {
        print row;
    }
    if (noteworthy!="") {
        if (Notes=="") {
            Notes=sprintf("%21s    %5s  %9s  %9s  %9s  %s  (numbers are ns/op)\n", "Number Type   ", "Count", "Sum2", "Sum3", "Sum4", "Best");
        }
        Notes=Notes sprintf("%21s: ", LastGroup) row "\n";
    }
};

# Main portion of awk processing script.
# Expected line format:
# $1: The test group, e.g. "Positive_Small_Floats".
# $2: The number of args used, e.g. "10".
# $3: The name of the sum funciton used, e.g. "Sum3".
# $4: The number of iterations it was able to do, e.g. "260043".
# $5: nanoseconds per operation, e.g. "4564".
# $6: The string "ns/op" (unused).
{
    # If we're in a new set, process the previous one and reset for the next set.
    if ($2!=LastCount && LastCount!="") {
        processSumSet();
        Sum2="";
        Sum3="";
        Sum4="";
    }

    # If starting a new test group and doing verbose output, print a header for the data rows.
    if ($1!=LastGroup && verbose!="") {
        printf "%s:\n  %5s  %9s  %9s  %9s  %s\n", $1, "Count", "Sum2", "Sum3", "Sum4", "Best";
    }

    # Set the appropriate sum variable.
    if ($3=="Sum2") {
        Sum2=$5;
    } else if ($3=="Sum3") {
        Sum3=$5;
    } else if ($3=="Sum4") {
        Sum4=$5;
    } else {
        printf "ERROR: Unknown sum function name: %q\n", $3;
    }

    # Make note of this count and group so we can compare it to the next line.
    LastCount=$2;
    LastGroup=$1;
}

END {
    # Since we process a set on the line after the set, we haven't processed the very last set yet.
    processSumSet();
    # And output the notes.
    if (verbose!="") {
        print "";
        print "------------------------------------------------------------------------------------------";
    }
    if (Notes!="") {
        print Notes;
    } else {
        print "Sum4 was the best on all test.";
    }
}
