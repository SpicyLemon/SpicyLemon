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
use POSIX qw(ceil);
#use SDBM_File;
#use Storable qw(dclone);
#use Time::HiRes qw(time usleep);  #this forces the time() function to return nanosecond information
#use XML::Simple;

#use lib qw(.);

my $DEBUG = $ENV{DEBUG};
my $PLAYER = 'PLAYER';
my $BOSS = 'BOSS';

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
}

sub mainProgram1 {
    my $boss = getBoss();
    my @winCombos = ();
    my @loseCombos = ();
    foreach my $dam (4..13) {
        foreach my $arm (0..10) {
            my %player = (
                HP => 100,
                Dam => $dam,
                Arm => $arm,
            );
            if (findWinner(\%player, $boss) eq $PLAYER) {
                push(@winCombos, [$dam, $arm]);
            }
            else {
                push(@loseCombos, [$dam, $arm]);
            }
        }
    }
    if ($DEBUG) {
        print "Player loses with these stats ".(scalar @loseCombos).":\n";
        foreach my $combo (@loseCombos) {
            printf('  Dam: %2d, Arm: %2d'."\n", $combo->[0], $combo->[1]);
        }
        print "Player wins with these stats ".(scalar @winCombos).":\n";
        foreach my $combo (@winCombos) {
            printf('  Dam: %2d, Arm: %2d'."\n", $combo->[0], $combo->[1]);
        }
    }
    my %gearByStats = ();
    foreach my $g (@{getGearCombos()}) {
        my $key = sprintf("%d-%d", $g->{Dam}, $g->{Arm});
        if (! exists $gearByStats{$key}) {
            $gearByStats{$key} = [];
        }
        push(@{$gearByStats{$key}}, $g);
    }
    print Dumper(\%gearByStats)."\n" if ($DEBUG);
    my $minCost = 500;
    my $minGear = undef;
    foreach my $win (@winCombos) {
        my $key = sprintf("%d-%d", $win->[0], $win->[1]);
        foreach my $g (@{$gearByStats{$key}}) {
            if ($g->{Cost} < $minCost) {
                $minCost = $g->{Cost};
                $minGear = $g;
            }
        }
    }
    print "Min Win Cost: ".Data::Dumper->new([$minGear])->Indent(0)->Terse(1)->Dump()."\n";
    print "Answer: ".$minCost."\n";
}

sub mainProgram2 {
    my $boss = getBoss();
    my @winCombos = ();
    my @loseCombos = ();
    foreach my $dam (4..13) {
        foreach my $arm (0..10) {
            my %player = (
                HP => 100,
                Dam => $dam,
                Arm => $arm,
            );
            if (findWinner(\%player, $boss) eq $PLAYER) {
                push(@winCombos, [$dam, $arm]);
            }
            else {
                push(@loseCombos, [$dam, $arm]);
            }
        }
    }
    if ($DEBUG) {
        print "Player loses with these stats ".(scalar @loseCombos).":\n";
        foreach my $combo (@loseCombos) {
            printf('  Dam: %2d, Arm: %2d'."\n", $combo->[0], $combo->[1]);
        }
        print "Player wins with these stats ".(scalar @winCombos).":\n";
        foreach my $combo (@winCombos) {
            printf('  Dam: %2d, Arm: %2d'."\n", $combo->[0], $combo->[1]);
        }
    }
    my %gearByStats = ();
    foreach my $g (@{getGearCombos()}) {
        my $key = sprintf("%d-%d", $g->{Dam}, $g->{Arm});
        if (! exists $gearByStats{$key}) {
            $gearByStats{$key} = [];
        }
        push(@{$gearByStats{$key}}, $g);
    }
    print Dumper(\%gearByStats)."\n" if ($DEBUG);
    my $maxCost = 0;
    my $maxGear = undef;
    foreach my $lose (@loseCombos) {
        my $key = sprintf("%d-%d", $lose->[0], $lose->[1]);
        foreach my $g (@{$gearByStats{$key}}) {
            if ($g->{Cost} > $maxCost) {
                $maxCost = $g->{Cost};
                $maxGear = $g;
            }
        }
    }
    print "Max Lose Cost: ".Data::Dumper->new([$maxGear])->Indent(0)->Terse(1)->Dump()."\n";
    print "Answer: ".$maxCost."\n";
}

sub findWinner {
    my $p = shift;
    my $b = shift;
    my $pz = $p->{Dam} > $b->{Arm} ? ceil($b->{HP}/($p->{Dam} - $b->{Arm})) : 100000;
    my $bz = $b->{Dam} > $p->{Arm} ? ceil($p->{HP}/($b->{Dam} - $p->{Arm})) : 100000;
    if ($pz <= $bz) {
        printf('(%2d) Player: HP: %2d, Dam: %2d, Arm: %2d wins vs Boss: HP: %2d, Dam: %2d, Arm: %2d (%2d)'."\n",
            $pz, $p->{HP}, $p->{Dam}, $p->{Arm}, $b->{HP}, $b->{Dam}, $b->{Arm}, $bz) if ($DEBUG);
        return $PLAYER;
    }
    else {
        printf('(%2d) Player: HP: %2d, Dam: %2d, Arm: %2d loses vs Boss: HP: %2d, Dam: %2d, Arm: %2d (%2d)'."\n",
            $pz, $p->{HP}, $p->{Dam}, $p->{Arm}, $b->{HP}, $b->{Dam}, $b->{Arm}, $bz) if ($DEBUG);
        return $BOSS;
    }
}

sub costsToString {
    my $costs = shift;
    my @lines = ();
    foreach my $val (sort {$a <=> $b} keys %$costs) {
        my @vals = ();
        foreach my $c (@{$costs->{$val}}) {
            push(@vals, sprintf("%3d", $c));
        }
        push(@lines, sprintf('%2d: %s', $val, join(' ', @vals)));
    }
    return join("\n", @lines);
}

sub getGearCombos {
    my @rv = ();
    my $weapons = gearToList(getWeapons());
    my $armor = gearToList(getArmor());
    my $rings = gearToList(getRings());
    foreach my $w (@$weapons) {
        foreach my $arm ((undef, @$armor)) {
            push(@rv, newGearCombo($w, $arm));
            foreach my $r1 (@$rings) {
                push(@rv, newGearCombo($w, $arm, $r1));
                foreach my $r2 (@$rings) {
                    if ($r1->{Cost} != $r2->{Cost}) {
                        push(@rv, newGearCombo($w, $arm, $r1, $r2));
                    }
                }
            }
        }
    }
    return \@rv;
}

sub newGearCombo {
    my $weapon = shift;
    my $armor = shift;
    my $ring1 = shift;
    my $ring2 = shift;
    my %rv = (
        Weapon => $weapon,
        Armor => $armor,
        Ring1 => $ring1,
        Ring2 => $ring2,
        Arm => 0,
        Dam => 0,
        Cost => 0,
    );
    foreach my $g (($weapon, $armor, $ring1, $ring2)) {
        if (defined $g) {
            $rv{Dam} += $g->{Dam};
            $rv{Arm} += $g->{Arm};
            $rv{Cost} += $g->{Cost};
        }
    }
    return \%rv;
}

sub getWeapons {
    return {
        Dagger     => { Cost =>  8, Dam => 4, Arm => 0 },
        Shortsword => { Cost => 10, Dam => 5, Arm => 0 },
        Warhammer  => { Cost => 25, Dam => 6, Arm => 0 },
        Longsword  => { Cost => 40, Dam => 7, Arm => 0 },
        Greataxe   => { Cost => 74, Dam => 8, Arm => 0 },
    };
}

sub getArmor {
    return {
        Leather    => { Cost =>  13, Dam => 0, Arm => 1 },
        Chainmail  => { Cost =>  31, Dam => 0, Arm => 2 },
        Splintmail => { Cost =>  53, Dam => 0, Arm => 3 },
        Bandedmail => { Cost =>  75, Dam => 0, Arm => 4 },
        Platemail  => { Cost => 102, Dam => 0, Arm => 5 },
    };
}

sub getRings {
    return {
        Dam1 => { Cost =>  25, Dam => 1, Arm => 0 },
        Dam2 => { Cost =>  50, Dam => 2, Arm => 0 },
        Dam3 => { Cost => 100, Dam => 3, Arm => 0 },
        Arm1 => { Cost =>  20, Dam => 0, Arm => 1 },
        Arm2 => { Cost =>  40, Dam => 0, Arm => 2 },
        Arm3 => { Cost =>  80, Dam => 0, Arm => 3 },
    };
}

sub gearToList {
    my $gear = shift;
    my @rv = ();
    foreach my $name (keys %$gear) {
        push(@rv, { Name => $name, %{$gear->{$name}} });
    }
    return \@rv;
}

sub getBoss {
    return {
        HP => 103,
        Dam => 9,
        Arm => 2,
    };
}

sub getStr {
   return q^Hit Points: 103
Damage: 9
Armor: 2^;
}
