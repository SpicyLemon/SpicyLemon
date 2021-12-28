#!/usr/bin/env perl
use strict;
use warnings;
use Algorithm::Combinatorics qw(combinations partitions);
#use Benchmark ':hireswallclock';
#use Carp;
use Data::Dumper;  $Data::Dumper::Sortkeys = 1;  $Data::Dumper::Terse = 1;  $Data::Dumper::Indent = 0;
#use DateTime;
#use Date::Parse; #imports str2time that takes in a string, and outputs an epoch.
#use Digest::MD5 qw(md5_hex);
#use Fcntl;
#use File::Copy;   #imports copy and move
#use HTML::Entities;
#use List::Util 'shuffle';
#use Math::BigInt;
#use Math::Round qw(nearest);
use POSIX qw(floor);
#use SDBM_File;
#use Storable qw(dclone);
#use Time::HiRes qw(time usleep);  #this forces the time() function to return nanosecond information
#use XML::Simple;

#use lib qw(.);

my $DEBUG = $ENV{DEBUG};

mainProgram();
print "Done\n";
exit(0);

sub mainProgram {
   print "Test:\n";
   mainProgramTest();
   print "part 1:\n";
   mainProgram1();
   print "part 2:\n";
   mainProgram2();
}

sub mainProgramTest {
    my $weights = getTestInp();
    printf('Weights: %s'."\n", Dumper($weights));
    printf('Sum: %d'."\n", sum($weights));
    my $sg = findQEOfSmallest3Group($weights);
    printf('Answer: %d = %s %s %s'."\n", $sg->{QE}, Dumper($sg->{Group1}), Dumper($sg->{Group2}), Dumper($sg->{Group3}));
}

sub mainProgram1 {
    my $weights = getInp();
    printf('Weights: %s'."\n", Dumper($weights));
    printf('Sum: %d'."\n", sum($weights));
    my $sg = findQEOfSmallest3Group($weights);
    printf('Answer: %d = %s %s %s'."\n", $sg->{QE}, Dumper($sg->{Group1}), Dumper($sg->{Group2}), Dumper($sg->{Group3}));
}

sub mainProgram2 {
    my $weights = getInp();
    printf('Weights: %s'."\n", Dumper($weights));
    printf('Sum: %d'."\n", sum($weights));
    my $sg = findQEOfSmallest4Group($weights);
    printf('Answer: %d = %s %s %s %s'."\n", $sg->{QE}, Dumper($sg->{Group1}), Dumper($sg->{Group2}), Dumper($sg->{Group3}), Dumper($sg->{Group4}));
}

sub findQEOfSmallest3Group {
    my $weights = shift;
    my $total = sum($weights);
    my $groupTotal = $total / 3;
    # First pass, just find the shortest combinations where the numbers sum to the group total.
    my @g1s = ();
    foreach my $l (2..$#{$weights}) {
        my $iter = combinations($weights, $l);
        while (my $g = $iter->next) {
            push (@g1s, $g) if (sum($g) == $groupTotal);
        }
        last if (scalar @g1s > 0);
    }
    printf('Found %d possible first groups:'."\n"."%s"."\n", scalar @g1s, arrArrString(\@g1s)) if ($DEBUG);
    # Calculate the QE for each and sort it smallest to largest.
    my @withQE = ();
    foreach my $ig1 (@g1s) {
        push(@withQE, { QE => product($ig1), Group => $ig1 });
    }
    @withQE = sort { $a->{QE} <=> $b->{QE} } @withQE;
    # Go through the sorted list and find the first one where the other numbers can go into equal groups too.
    my $rv = {};
    foreach my $g1 (@withQE) {
        my $g2pool = getWithout($weights, $g1->{Group});
        foreach my $l (($#{$g1->{Group}}+1)..floor($#{$g2pool}/2)) {
            my $iter = combinations($g2pool, $l);
            while (my $g2 = $iter->next) {
                if (sum($g2) == $groupTotal) {
                    $rv = { QE => $g1->{QE}, Group1 => $g1->{Group}, Group2 => $g2, Group3 => getWithout($g2pool, $g2) };
                    last;
                }
            }
        }
        last if (scalar keys %$rv > 0);
    }
    if ($DEBUG) {
        if (defined $rv) {
            printf('Found solution: %s.'."\n", Dumper($rv));
        }
        else {
            print "No solution found.\n";
        }
    }
    return $rv;
}

sub findQEOfSmallest4Group {
    my $weights = shift;
    my $total = sum($weights);
    my $groupTotal = $total / 4;
    # First pass, just find the shortest combinations where the numbers sum to the group total.
    my @g1s = ();
    foreach my $l2 (2..$#{$weights}) {
        my $iter = combinations($weights, $l2);
        while (my $g = $iter->next) {
            push (@g1s, $g) if (sum($g) == $groupTotal);
        }
        last if (scalar @g1s > 0);
    }
    printf('Found %d possible first groups:'."\n"."%s"."\n", scalar @g1s, arrArrString(\@g1s)) if ($DEBUG);
    # Calculate the QE for each and sort it smallest to largest.
    my @withQE = ();
    foreach my $ig1 (@g1s) {
        push(@withQE, { QE => product($ig1), Group1 => $ig1 });
    }
    @withQE = sort { $a->{QE} <=> $b->{QE} } @withQE;
    # Go through the sorted list and find the first one where the other numbers can go into equal groups too.
    my $rv = {};
    foreach my $g1 (@withQE) {
        my $done = 0;
        my $g2pool = getWithout($weights, $g1->{Group1});
        foreach my $l2 (($#{$g1->{Group1}}+1)..floor($#{$g2pool}/3)) {
            my $iter2 = combinations($g2pool, $l2);
            while (my $g2 = $iter2->next) {
                if (sum($g2) == $groupTotal) {
                    my $g3pool = getWithout($g2pool, $g2);
                    foreach my $l3 ($#{$g2}..floor($#{$g3pool}/2)) {
                        my $iter3 = combinations($g3pool, $l3);
                        while (my $g3 = $iter3->next) {
                            if (sum($g3) == $groupTotal) {
                                $rv = { QE => $g1->{QE}, Group1 => $g1->{Group1}, Group2 => $g2, Group3 => $g3, Group4 => getWithout($g3pool, $g3) };
                                $done = 1;
                                last;
                            }
                        }
                        last if ($done);
                    }
                    last if ($done);
                }
            }
            last if ($done);
        }
        last if ($done);
    }
    if ($DEBUG) {
        if (defined $rv) {
            printf('Found solution: %s.'."\n", Dumper($rv));
        }
        else {
            print "No solution found.\n";
        }
    }
    return $rv;
}

sub getWithout {
    my $all = shift;
    my $toRem = shift;
    my @rv = ();
    foreach my $v (@$all) {
        push(@rv, $v) if (!contains($toRem, $v));
    }
    return \@rv;
}

sub contains {
    my $arr = shift;
    my $val = shift;
    foreach my $v (@$arr) {
        return 1 if ($val == $v);
    }
    return 0;
}

sub arrArrString {
    my $arrArr = shift;
    my @lines = ();
    foreach my $arr (@$arrArr) {
        push(@lines, '  '.Dumper($arr));
    }
    return "[\n".join("\n", @lines)."\n]";
}

sub sum {
    my $vals = shift;
    my $rv = 0;
    foreach my $v (@$vals) {
        $rv += $v;
    }
    return $rv
}

sub product {
    my $vals = shift;
    my $rv = 1;
    foreach my $v (@$vals) {
        $rv *= $v;
    }
    return $rv
}

sub getTestInp {
    return [(1..5), (7..11)];
}

sub getInp {
   return [qw^1
3
5
11
13
17
19
23
29
31
37
41
43
47
53
59
67
71
73
79
83
89
97
101
103
107
109
113^];
}
