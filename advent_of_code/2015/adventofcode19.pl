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

use List::PriorityQueue;

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
    my $input = parseInput(getTestStr());
    print Dumper($input)."\n";
    my $reps = getPossibilities($input->{rep}, $input->{mol});
    printf("Possibilities (%d): %s\n", scalar @$reps, Data::Dumper->new([$reps])->Terse(1)->Indent(0)->Dump());
    countSteps($input->{rep}, 'e', $input->{mol});
    countSteps($input->{rep}, 'e', 'HOHOHO');
}

sub mainProgram1 {
    my $input = parseInput(getStr());
    my $reps = getPossibilities($input->{rep}, $input->{mol});
    print "There are ".(scalar @$reps)." Possibilities.\n";
}

sub mainProgram2 {
    my $input = parseInput(getStr());
    countSteps($input->{rep}, 'e', $input->{mol});
}

sub countSteps {
    my $reps = shift;
    my $from = shift;
    my $to = shift;
    my $steps = dijkstraCountStepsRev($reps, $from, $to);
    printf('It took %d steps to get from %s to %s.'."\n", $steps, $from, $to);
}

sub dijkstraCountStepsRev {
    my $reps = shift;
    my $from = shift;
    my $to = shift;
    my $fromL = length($from);
    my $unvisited = new List::PriorityQueue;
    $unvisited->insert($to, 0);
    my $qSize = 1;
    my %known = (
        "$to" => { steps => 0, visited => 0 },
    );
    my $i = 0;
    while (1) {
        $i++;
        my $next = $unvisited->pop();
        $qSize--;
        last if (!defined $next);
        next if ($known{$next}->{visited});
        $known{$next}->{visited} = 1;
        my $nextStep = $known{$next}->{steps} + 1;
        my $mols = getRevPossibilities($reps, $next, $from);
        foreach my $mol (@$mols) {
            if ($mol eq $from) {
                return $nextStep;
            }
            if (length($mol) >= $fromL) {
                my $priority = sprintf("%d%04d", length($mol), $nextStep);
                if (! exists $known{$mol}) {
                    $unvisited->insert($mol, $priority);
                    $qSize++;
                    $known{$mol} = { steps => $nextStep, visited => 0 };
                }
                elsif (!$known{$mol}->{visited} && $known{$mol}->{steps} > $nextStep) {
                    $known{$mol}->{steps} = $nextStep;
                    $unvisited->update($mol, $priority);
                }
            }
        }
        if ($i <= 10 || $i % 10000 == 0) {
            my ($min, $max, $count, $visited, $unvisited, $maxStep) = getKeyLenMinMax(\%known);
            printf('%d: %d of %d molecules checked, %d to go with lengths %d to %d, and max steps %d.'."\n",
                $i, $visited, $count, $unvisited, $min, $max, $maxStep);
        }
    }
    print "No solution found.\n";
    return -1;
}

sub getRevPossibilities {
    my $reps = shift;
    my $mol = shift;
    my $toFind = shift;
    my $toFindL = length($toFind);
    my %rv = ();
    foreach my $rep (keys %$reps) {
        foreach my $toRep (@{$reps->{$rep}}) {
            my $toRepL = length($toRep);
            if ($rep eq $toFind) {
                if ($mol eq $toRep) {
                    return [$toFind];
                }
                next;
            }
            while ($mol =~ m{$toRep}g) {
                my $m = $mol;
                substr($m, $-[0], $toRepL, $rep);
                if (length($m) >= $toFindL) {
                    $rv{$m} = 1;
                }
            }
        }
    }
    return [keys %rv];
}

sub dijkstraCountSteps {
    my $reps = shift;
    my $from = shift;
    my $to = shift;
    my $toL = length($to);
    my $unvisited = new List::PriorityQueue;
    $unvisited->insert($from, 0);
    my $qSize = 1;
    my %known = (
        "$from" => { steps => 0, visited => 0 },
    );
    my $i = 0;
    while (1) {
        $i++;
        my $next = $unvisited->pop();
        $qSize--;
        next if ($known{$next}->{visited});
        $known{$next}->{visited} = 1;
        my $nextStep = $known{$next}->{steps} + 1;
        my $mols = getPossibilities($reps, $next, $to);
        foreach my $mol (@$mols) {
            if ($mol eq $to) {
                return $nextStep;
            }
            if (length($mol) <= $toL) {
                my $priority = sprintf("%d%04d", $toL - length($mol), $nextStep);
                if (! exists $known{$mol}) {
                    $unvisited->insert($mol, $priority);
                    $qSize++;
                    $known{$mol} = { steps => $nextStep, visited => 0 };
                }
                elsif (!$known{$mol}->{visited} && $known{$mol}->{steps} > $nextStep) {
                    $known{$mol}->{steps} = $nextStep;
                    $unvisited->update($mol, $priority);
                }
            }
        }
        if ($i <= 10 || $i % 10000 == 0) {
            my ($min, $max, $count, $visited, $unvisited, $maxStep) = getKeyLenMinMax(\%known);
            printf('%d: %d of %d molecules checked, %d to go with lengths %d to %d, and max steps %d.'."\n",
                $i, $visited, $count, $unvisited, $min, $max, $maxStep);
        }
    }
}

sub getKeyLenMinMax {
    my $h = shift;
    my $min = 500;
    my $max = 0;
    my $count = 0;
    my $visited = 0;
    my $maxStep = 0;
    foreach my $k (keys %$h) {
        $count++;
        if ($h->{$k}->{visited}) {
            $visited++;
            if (length($k) < $min) {
                $min = length($k);
            }
            if (length($k) > $max) {
                $max = length($k);
            }
        }
        if ($h->{$k}->{steps} > $maxStep) {
            $maxStep = $h->{$k}->{steps};
        }
    }
    return ($min, $max, $count, $visited, $count - $visited, $maxStep);
}

sub getPossibilities {
    my $reps = shift;
    my $mol = shift;
    my $toFind = shift;
    my $toFindL = defined $toFind ? length($toFind) : 0;
    my %rv = ();
    foreach my $toRep (keys %$reps) {
        my $toRepL = length($toRep);
        while ($mol =~ m{$toRep}g) {
            foreach my $rep (@{$reps->{$toRep}}) {
                my $m = $mol;
                substr($m, $-[0], $toRepL, $rep);
                if (! defined $toFind) {
                    $rv{$m} = 1;
                }
                elsif ($m eq $toFind) {
                    return [$m];
                }
                elsif (length($m) <= $toFindL) {
                    $rv{$m} = 1;
                }
            }
        }
    }
    return [sort keys %rv];
}

sub parseInput {
    my $str = shift;
    my %rv = (
        rep => {},
        mol => undef,
    );
    my $nextIsMol = 0;
    foreach my $line (split("\n", $str)) {
        $line =~ s{^\s+}{};
        $line =~ s{\s+$}{};
        if ($nextIsMol) {
            $rv{mol} = $line;
        }
        elsif (length $line == 0) {
            $nextIsMol = 1;
        }
        else {
            if ($line =~ m{^(\w+) => (\w+)$}) {
                if (! exists $rv{rep}->{$1}) {
                    $rv{rep}->{$1} = [];
                }
                push (@{$rv{rep}->{$1}}, $2);
            } else {
                die "Failed to parse replacement: $line";
            }
        }
    }
    return \%rv;
}

sub getTestStr {
    return q^e => H
e => O
H => HO
H => OH
O => HH

HOH^;
}

sub getStr {
   return q^Al => ThF
Al => ThRnFAr
B => BCa
B => TiB
B => TiRnFAr
Ca => CaCa
Ca => PB
Ca => PRnFAr
Ca => SiRnFYFAr
Ca => SiRnMgAr
Ca => SiTh
F => CaF
F => PMg
F => SiAl
H => CRnAlAr
H => CRnFYFYFAr
H => CRnFYMgAr
H => CRnMgYFAr
H => HCa
H => NRnFYFAr
H => NRnMgAr
H => NTh
H => OB
H => ORnFAr
Mg => BF
Mg => TiMg
N => CRnFAr
N => HSi
O => CRnFYFAr
O => CRnMgAr
O => HP
O => NRnFAr
O => OTi
P => CaP
P => PTi
P => SiRnFAr
Si => CaSi
Th => ThCa
Ti => BP
Ti => TiTi
e => HF
e => NAl
e => OMg

CRnSiRnCaPTiMgYCaPTiRnFArSiThFArCaSiThSiThPBCaCaSiRnSiRnTiTiMgArPBCaPMgYPTiRnFArFArCaSiRnBPMgArPRnCaPTiRnFArCaSiThCaCaFArPBCaCaPTiTiRnFArCaSiRnSiAlYSiThRnFArArCaSiRnBFArCaCaSiRnSiThCaCaCaFYCaPTiBCaSiThCaSiThPMgArSiRnCaPBFYCaCaFArCaCaCaCaSiThCaSiRnPRnFArPBSiThPRnFArSiRnMgArCaFYFArCaSiRnSiAlArTiTiTiTiTiTiTiRnPMgArPTiTiTiBSiRnSiAlArTiTiRnPMgArCaFYBPBPTiRnSiRnMgArSiThCaFArCaSiThFArPRnFArCaSiRnTiBSiThSiRnSiAlYCaFArPRnFArSiThCaFArCaCaSiThCaCaCaSiRnPRnCaFArFYPMgArCaPBCaPBSiRnFYPBCaFArCaSiAl^;
}
