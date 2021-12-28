#!/usr/bin/env perl
use strict;
use warnings;
#use Benchmark ':hireswallclock';
#use Carp;
use Data::Dumper;  $Data::Dumper::Sortkeys = 1;
#use DateTime;
#use Date::Parse; #imports str2time that takes in a string, and outputs an epoch.
#use Digest::MD5 qw(md5_hex);
#use Fcntl;
#use File::Copy;   #imports copy and move
#use HTML::Entities;
#use List::Util 'shuffle';
#use Math::BigInt;
#use Math::Round qw(nearest);
#use SDBM_File;
#use Storable qw(dclone);
#use Time::HiRes qw(time usleep);  #this forces the time() function to return nanosecond information
#use XML::Simple;

#use lib qw(.);

my $DEBUG = $ENV{DEBUG};

my @KNOWN = (
    [ 20151125,  18749137,  17289845,  30943339,  10071777,  33511524 ],
    [ 31916031,  21629792,  16929656,   7726640,  15514188,   4041754 ],
    [ 16080970,   8057251,   1601130,   7981243,  11661866,  16474243 ],
    [ 24592653,  32451966,  21345942,   9380097,  10600672,  31527494 ],
    [    77061,  17552253,  28094349,   6899651,   9250759,  31663883 ],
    [ 33071741,   6796745,  25397450,  24659492,   1534922,  27995004 ],
);

my $FIRST = 20151125;
my $MULT_BY = 252533;
my $MOD_BY = 33554393;

mainProgram();
print "Done\n";
exit(0);

sub mainProgram {
   print "Test:\n";
   mainProgramTest();
   print "part 1:\n";
   mainProgram1();
}

sub mainProgramTest {
    print 'Indexes:'."\n";
    printGrid(makeGrid(6, 6));
    my @pairs = ();
    foreach my $arg (@ARGV) {
        if (scalar @pairs == 0 || defined $pairs[$#pairs]->[1]) {
            push(@pairs, [$arg, undef]);
        }
        else {
            $pairs[$#pairs][1] = $arg;
        }
    }
    foreach my $pair (@pairs) {
        my $row = $pair->[0];
        my $col = defined $pair->[1] ? $pair->[1] : 1;
        printf('The index of row %d and col %d is %d.'."\n", $row, $col, getIndex($row, $col));
    }
    my $known = remakeKnown();
    print 'Known:'."\n";
    printGrid(\@KNOWN);
    print 'Remake of known:'."\n";
    printGrid($known);
}

sub mainProgram1 {
    my $row = 2947;
    my $col = 3029;
    my $ind = getIndex($row, $col);
    my $val = calcCell($row, $col);
    printf('Row %d, Col %d has index %d and value %d'."\n", $row, $col, $ind, $val);
}

sub printGrid {
    my $grid = shift;
    my $width = 0;
    foreach my $row (@$grid) {
        foreach my $cell (@$row) {
            my $cw = length(sprintf('%d', $cell));
            if ($cw > $width) {
                $width = $cw;
            }
        }
    }
    my $cf = '%'.$width.'d';
    my @lines = ();
    foreach my $row (@$grid) {
        my @cells = ();
        foreach my $cell (@$row) {
            push(@cells, sprintf($cf, $cell));
        }
        push (@lines, join('  ', @cells));
    }
    print join("\n", @lines)."\n";
}

sub remakeKnown {
    my $rows = 6;
    my $cols = 6;
    my @rv = ();
    foreach my $y (1..$rows) {
        my @row = ();
        foreach my $x (1..$cols) {
            push(@row, calcCell($y, $x));
        }
        push (@rv, \@row);
    }
    return \@rv;
}

sub calcCell {
    my $row = shift;
    my $col = shift;
    my $index = getIndex($row, $col);
    my $val = $FIRST;
    foreach (2..$index) {
        $val = ($val * $MULT_BY) % $MOD_BY;
    }
    return $val;
}

sub makeGrid {
    my $rows = shift;
    my $cols = shift;
    my @rv = ();
    foreach my $y (1..$rows) {
        my @row = ();
        foreach my $x (1..$cols) {
            push(@row, getIndex($y, $x));
        }
        push (@rv, \@row);
    }
    return \@rv;
}

sub getIndex {
    my $row = shift;
    my $col = shift;
    return int(($col*($col+1) + ($row-1)*($col+$col+$row-2))/2);
}
