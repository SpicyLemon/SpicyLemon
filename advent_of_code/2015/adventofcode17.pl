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
my $DEBUG = $ENV{DEBUG};
my $ind = 0;

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
    my $nums = parseTestStr();
    print Data::Dumper->new([$nums], ['nums'])->Indent(0)->Dump()."\n";
    my $sums = findSums($nums, 25);
    print onePerLineStr($sums)."\n";
    print "There are ".(scalar @$sums)." sums.\n";
}

sub mainProgram1 {
    my $nums = parseStr();
    print Data::Dumper->new([$nums])->Indent(0)->Dump()."\n";
    my $sums = findSums($nums, 150);
    print onePerLineStr($sums)."\n";
    print "There are ".(scalar @$sums)." sums.\n";
}

sub mainProgram2 {
    my $nums = parseStr();
    print Data::Dumper->new([$nums])->Indent(0)->Dump()."\n";
    my $asums = findSums($nums, 150);
    my $sums = filterSumsOnMinLength($asums);
    print onePerLineStr($sums)."\n";
    print "There are ".(scalar @$sums)." sums.\n";
}

sub filterSumsOnMinLength {
    my $sums = shift;
    my $min = 20; # From looking at the output of part 1, this looks to be larger than all the lines.
    my @rv = ();
    foreach my $sum (@$sums) {
        if (scalar @$sum < $min) {
            $min = scalar @$sum;
            @rv = ();
        }
        if (scalar @$sum == $min) {
            push (@rv, $sum);
        }
    }
    return \@rv;
}

sub onePerLineStr {
    my $vals = shift;
    my @lines = ();
    my $i = 0;
    foreach my $line (@$vals) {
        push (@lines, sprintf('%4d: %s', $i, Data::Dumper->new([$line])->Terse(1)->Indent(0)->Dump()));
        $i++;
    }
    return join("\n", @lines);
}

sub findSums {
    my $vals = shift;
    my $target = shift;
    print ''.('  ' x $ind).sprintf("findSums(%s, %d)\n", Data::Dumper->new([$vals])->Terse(1)->Indent(0)->Dump(), $target) if $DEBUG;
    $ind++;
    if (scalar @$vals == 0) {
        $ind--;
        print ''.('  ' x $ind).sprintf("findSums(%s, %d): no vals. Returning [].\n",
            Data::Dumper->new([$vals])->Terse(1)->Indent(0)->Dump(), $target) if $DEBUG;
        return [];
    }
    my $subStart = 0;
    my $subEnd = $#{$vals};
    while ($subStart <= $subEnd && $vals->[$subStart] > $target) {
        $subStart++;
    }
    my @rv = ();
    while ($subStart <= $subEnd && $vals->[$subStart] == $target) {
        push (@rv, [$target]);
        $subStart++;
    }
    my $keyVal = $vals->[$subStart];
    $subStart++;
    if ($subStart > $subEnd) {
        $ind--;
        print ''.('  ' x $ind).sprintf("findSums(%s, %d): no sub vals. Returning %s.\n",
            Data::Dumper->new([$vals])->Terse(1)->Indent(0)->Dump(), $target,
            Data::Dumper->new([\@rv])->Terse(1)->Indent(0)->Dump()) if $DEBUG;
        return \@rv;
    }
    my @subs = @$vals[$subStart..$subEnd];
    foreach my $subSum (@{findSums(\@subs, $target - $keyVal)}) {
        push (@rv, [$keyVal, @$subSum])
    }
    foreach my $subSum (@{findSums(\@subs, $target)}) {
        push (@rv, $subSum);
    }
    $ind--;
    print ''.('  ' x $ind).sprintf("findSums(%s, %d): Returning %s.\n",
        Data::Dumper->new([$vals])->Terse(1)->Indent(0)->Dump(), $target,
        Data::Dumper->new([\@rv])->Terse(1)->Indent(0)->Dump()) if $DEBUG;
    return \@rv;
}

sub parseTestStr {
    return parseNums(getTestStr());
}

sub parseStr {
    return parseNums(getStr());
}

sub parseNums {
    my $str = shift;
    my @rv = ();
    foreach my $line (split("\n", $str)) {
        $line =~ s{^\s+}{};
        $line =~ s{\s+$}{};
        if (length($line) != 0) {
            push (@rv, $line*1);
        }
    }
    return [sort {$b <=> $a} @rv];
}

sub getTestStr {
    return q^20
15
10
5
5^;
}

sub getStr {
   return q^50
44
11
49
42
46
18
32
26
40
21
7
18
43
10
47
36
24
22
40^;
}
