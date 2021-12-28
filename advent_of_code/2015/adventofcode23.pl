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
    my $prog = parseInput(getTestStr());
    printProg($prog);
    my ($va, $vb, $i) = runProg(0, 0, $prog);
    printf('After running, register a = %d, register b = %d exeucting %d lines.'."\n", $va, $vb, $i);
}

sub mainProgram1 {
    my ($va, $vb, $i) = runProg(0, 0, parseInput(getStr()));
    printf('After running, register a = %d, register b = %d exeucting %d lines.'."\n", $va, $vb, $i);
}

sub mainProgram2 {
    my ($va, $vb, $i) = runProg(1, 0, parseInput(getStr()));
    printf('After running, register a = %d, register b = %d exeucting %d lines.'."\n", $va, $vb, $i);
}

sub runProg {
    my $ra = shift;
    my $rb = shift;
    my $prog = shift;
    my %reg = (
        a => $ra,
        b => $rb,
    );
    my $cur = 0;
    my $i = 0;
    while ($cur < scalar @$prog) {
        my $inst = $prog->[$cur];
        $i++;
        my $oldcur = $cur;
        if ($inst->{Inst} eq 'hlf') {
            $reg{$inst->{Reg}} = int($reg{$inst->{Reg}}/2);
            $cur++;
        }
        elsif ($inst->{Inst} eq 'tpl') {
            $reg{$inst->{Reg}} *= 3;
            $cur++;
        }
        elsif ($inst->{Inst} eq 'inc') {
            $reg{$inst->{Reg}}++;
            $cur++;
        }
        elsif ($inst->{Inst} eq 'jmp') {
            $cur += $inst->{Off};
        }
        elsif ($inst->{Inst} eq 'jie') {
            if ($reg{$inst->{Reg}} % 2 == 0) {
                $cur += $inst->{Off};
            }
            else {
                $cur++
            }
        }
        elsif ($inst->{Inst} eq 'jio') {
            if ($reg{$inst->{Reg}} == 1) {
                $cur += $inst->{Off};
            }
            else {
                $cur++
            }
        }
        else {
            die "Unknown instruction: ".instStr($inst);
        }
        printf('After %d: %d %s -> a = %d, b = %d, cur = %d'."\n", $i, $oldcur, instStr($inst), $reg{a}, $reg{b}, $cur) if ($DEBUG);
    }
    return ($reg{a}, $reg{b}, $i);
}

sub parseInput {
    my $str = shift;
    my @rv = ();
    foreach my $line (split("\n", $str)) {
        $line =~ s{^\s+}{};
        $line =~ s{\s+$}{};
        if (length($line) > 0) {
            if ($line =~ m{^(hlf|tpl|inc) (a|b)$}) {
                push(@rv, { Inst => $1, Reg => $2 });
            }
            elsif ($line =~ m{^(jio|jie) (a|b), \+?(-?\d+)$}) {
                push(@rv, { Inst => $1, Reg => $2, Off => $3+0 });
            }
            elsif ($line =~ m{^(jmp) \+?(-?\d+)$}) {
                push(@rv, { Inst => $1, Off => $2+0 });
            }
            else {
                die "Could not parse line: [$line].";
            }
        }
    }
    return \@rv;
}

sub printProg {
    my $prog = shift;
    my $i = 0;
    foreach my $line (@$prog) {
        $i++;
        printf("%2d: %s\n", $i, instStr($line));
    }
}

sub instStr {
    my $inst = shift;
    if (exists $inst->{Reg}) {
        if (exists $inst->{Off}) {
            return sprintf('%s %s %+d', $inst->{Inst}, $inst->{Reg}, $inst->{Off});
        }
        return sprintf('%s %s', $inst->{Inst}, $inst->{Reg});
    }
    return sprintf('%s %+d', $inst->{Inst}, $inst->{Off});
}


sub getTestStr {
    return q^inc a
jio a, +2
tpl a
inc a^;
}

sub getStr {
   return q^jio a, +19
inc a
tpl a
inc a
tpl a
inc a
tpl a
tpl a
inc a
inc a
tpl a
tpl a
inc a
inc a
tpl a
inc a
inc a
tpl a
jmp +23
tpl a
tpl a
inc a
inc a
tpl a
inc a
inc a
tpl a
inc a
tpl a
inc a
tpl a
inc a
tpl a
inc a
inc a
tpl a
inc a
inc a
tpl a
tpl a
inc a
jio a, +8
inc b
jie a, +4
tpl a
inc a
jmp +2
hlf a
jmp -7^;
}
