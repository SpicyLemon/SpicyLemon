#!/usr/bin/perl
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
    my $input = parseInput(getTestStr());
    print Data::Dumper->Dump([$input], [qw(input)]);
    my @factors = ('capacity', 'durability', 'flavor', 'texture');
    my $score = getScore({'Butterscotch' => 44, 'Cinnamon' => 56}, $input, \@factors);
    print "Score: $score\n";
    my ($bscore, $bmix) = findMaxCombo($input, \@factors);
    print "My Score: $bscore\n";
    print Data::Dumper->Dump([$bmix], [qw(mix)]);
}

sub mainProgram1 {
    my $input = parseInput(getStr());
    my @factors = ('capacity', 'durability', 'flavor', 'texture');
    my ($bscore, $bmix) = findMaxCombo($input, \@factors);
    print "My Score: $bscore\n";
    print Data::Dumper->Dump([$bmix], [qw(mix)]);
}

sub mainProgram2 {
    my $input = parseInput(getStr());
    my @factors = ('capacity', 'durability', 'flavor', 'texture');
    my ($bscore, $bmix) = findMaxComboWith500Calories($input, \@factors);
    print "My Score: $bscore\n";
    print Data::Dumper->Dump([$bmix], [qw(mix)]);
}

sub findMaxComboWith500Calories {
    my $costs = shift;
    my $factors = shift;
    my @ingredients = keys %$costs;
    my $combos = getCombos(scalar @ingredients);
    my $rv = 0;
    my $best = undef;
    foreach my $combo (@$combos) {
        my %mix = ();
        foreach my $i (0..$#ingredients) {
            $mix{$ingredients[$i]} = $combo->[$i];
        }
        if (getCalories(\%mix, $costs) == 500) {
            my $score = getScore(\%mix, $costs, $factors);
            if ($score > $rv) {
                $rv = $score;
                $best = \%mix;
            }
        }
    }
    return ($rv, $best);
}

sub findMaxCombo {
    my $costs = shift;
    my $factors = shift;
    my @ingredients = keys %$costs;
    my $combos = getCombos(scalar @ingredients);
    my $rv = 0;
    my $best = undef;
    foreach my $combo (@$combos) {
        my %mix = ();
        foreach my $i (0..$#ingredients) {
            $mix{$ingredients[$i]} = $combo->[$i];
        }
        my $score = getScore(\%mix, $costs, $factors);
        if ($score > $rv) {
            $rv = $score;
            $best = \%mix;
        }
    }
    return ($rv, $best);
}

sub getCombos {
    my $len = shift;
    if ($len != 2 && $len != 4) {
        die "arg passed to getCombos must be either 2 or for, got [$len]";
    }
    my $max = 100;
    my @rv = ();
    if ($len == 2) {
        foreach my $x (0..$max) {
            push (@rv, [$x, $max-$x]);
        }
    }
    else {
        foreach my $x (0..$max) {
            foreach my $y (0..$max) {
                if ($x+$y > $max) {
                    last;
                }
                foreach my $z (0..$max) {
                    if ($x + $y + $z > $max) {
                        last;
                    }
                    push (@rv, [$x, $y, $z, $max - $x - $y - $z]);
                }
            }
        }
    }
    print "There are ".(scalar @rv)." ways that $len numbers from 0 to $max can add up to $max.\n";
    return \@rv;
}

sub getScore {
    my $mix = shift;
    my $costs = shift;
    my $factors = shift;
    my $rv = 1;
    foreach my $factor (@$factors) {
        my $fscore = 0;
        foreach my $ing (keys %$mix) {
            $fscore += $costs->{$ing}->{$factor} * $mix->{$ing};
        }
        if ($fscore > 0) {
            $rv *= $fscore;
        }
        else {
            $rv = 0;
            last;
        }
    }
    return $rv;
}

sub getCalories {
    my $mix = shift;
    my $costs = shift;
    my $rv = 0;
    foreach my $ing (keys %$mix) {
        $rv += $costs->{$ing}->{calories} * $mix->{$ing};
    }
    return $rv;
}

sub parseInput {
    my $input = shift;
    my %rv = ();
    foreach my $line (split("\n", $input)) {
        $line =~ s{^\s+}{};
        $line =~ s{\s+$}{};
        if (length($line) > 0) {
            if ($line =~ m{^(\w+): (\w+) (-?\d+), (\w+) (-?\d+), (\w+) (-?\d+), (\w+) (-?\d+), (\w+) (-?\d+)}) {
                $rv{$1} = {
                    $2 => $3,
                    $4 => $5,
                    $6 => $7,
                    $8 => $9,
                    $10 => $11,
                }
            }
        }
    }
    return \%rv
}


sub getTestStr {
    return q^Butterscotch: capacity -1, durability -2, flavor 6, texture 3, calories 8
Cinnamon: capacity 2, durability 3, flavor -2, texture -1, calories 3^;
}

sub getStr {
   return q^Sprinkles: capacity 5, durability -1, flavor 0, texture 0, calories 5
PeanutButter: capacity -1, durability 3, flavor 0, texture 0, calories 1
Frosting: capacity 0, durability -1, flavor 4, texture 0, calories 6
Sugar: capacity -1, durability 0, flavor 0, texture 2, calories 8^;
}
