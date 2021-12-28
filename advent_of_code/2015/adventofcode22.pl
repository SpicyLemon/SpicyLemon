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
use List::PriorityQueue;
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
}

sub mainProgram1 {
    my $minCost = getMinManaCost1(50, 500, 71);
    print "Min cost for player to win: $minCost\n";
}

sub mainProgram2 {
    my $minCost = getMinManaCost2(50, 500, 71);
    print "Min cost for player to win: $minCost\n";
}

sub getMinManaCost1 {
    my $playerHP = shift;
    my $playerMana = shift;
    my $bossHP = shift;
    my $spells = getSpells();
    my $initialState = newState(0, 0, $playerHP, $playerMana, 0, 0, $bossHP, 0);
    my $initKey = stateKey($initialState);
    my %known =(
        $initKey => $initialState,
    );
    my $rv = undef;
    my $unvisited = new List::PriorityQueue;
    $unvisited->insert($initKey, 0);
    my $qSize = 1;
    my $i = 0;
    while (my $next = $unvisited->pop()) {
        $i++;
        $qSize--;
        my $state = $known{$next};
        printf('%4d: %4d checked %4d, unchecked %4d, total %4d, checking: %s'."\n", $i, defined $rv ? $rv : -1,
            (scalar keys %known) - $qSize, $qSize, scalar keys %known, stateString($state)) if ($DEBUG);
        last if (defined $rv && $rv < $state->{ManaSpent});
        next if ($state->{Visited});
        $state->{Visited} = 1;
        foreach my $spell (@$spells) {
            next if (!canCast($state, $spell));
            my $newState = doTurns1($state, $spell);
            if ($newState->{BossHP} <= 0) {
                if (! defined $rv || $newState->{ManaSpent} < $rv) {
                    $rv = $newState->{ManaSpent};
                }
                next;
            }
            next if ($newState->{PlayerHP} <= 0);
            my $key = stateKey($newState);
            if (exists $known{$key}) {
                if ($newState->{ManaSpent} < $known{$key}->{ManaSpent}) {
                    $known{$key}->{ManaSpent} = $newState->{ManaSpent};
                    if ($known{$key}->{Visited}) {
                        $known{$key}->{Visited} = 0;
                        $unvisited->insert($key, $newState->{ManaSpent});
                        $qSize++;
                    }
                    else {
                        $unvisited->update($key, $known{$key}->{ManaSpent});
                    }
                }
            }
            else {
                $known{$key} = $newState;
                $unvisited->insert($key, $newState->{ManaSpent});
                $qSize++;
            }
        }
    }
    return $rv;
}

sub doTurns1 {
    my $state = shift;
    my $spell = shift;
    my $rv = copyState($state);
    # Player Turn
    if ($rv->{PlayerRecharge}) {
        $rv->{PlayerMana} += 101;
        $rv->{PlayerRecharge}--;
    }
    if ($rv->{PlayerShield}) {
        $rv->{PlayerShield}--;
    }
    if ($rv->{BossPoison}) {
        $rv->{BossHP} -= 3;
        $rv->{BossPoison}--;
    }
    $rv->{PlayerMana} -= $spell->{Cost};
    $rv->{ManaSpent} += $spell->{Cost};
    if ($spell->{Turns}) {
        if ($spell->{Arm}) {
            $rv->{PlayerShield} = $spell->{Turns};
        }
        elsif ($spell->{Mana}) {
            $rv->{PlayerRecharge} = $spell->{Turns};
        }
        elsif ($spell->{Dam}) {
            $rv->{BossPoison} = $spell->{Turns};
        }
        else {
            die "unknown effect spell: ".Dumper($spell);
        }
    }
    else {
        $rv->{BossHP} -= $spell->{Dam};
        if ($spell->{Heal}) {
            $rv->{PlayerHP} += $spell->{Heal};
        }
    }
    # Boss Turn
    return $rv if ($rv->{BossHP} <= 0);
    if ($rv->{PlayerRecharge}) {
        $rv->{PlayerMana} += 101;
        $rv->{PlayerRecharge}--;
    }
    my $shield = 0;
    if ($rv->{PlayerShield}) {
        $shield = 7;
        $rv->{PlayerShield}--;
    }
    if ($rv->{BossPoison}) {
        $rv->{BossHP} -= 3;
        $rv->{BossPoison}--;
    }
    $rv->{PlayerHP} -= 10 - $shield;
    return $rv;
}

sub getMinManaCost2 {
    my $playerHP = shift;
    my $playerMana = shift;
    my $bossHP = shift;
    my $spells = getSpells();
    my $initialState = newState(0, 0, $playerHP, $playerMana, 0, 0, $bossHP, 0);
    my $initKey = stateKey($initialState);
    my %known =(
        $initKey => $initialState,
    );
    my $rv = undef;
    my $unvisited = new List::PriorityQueue;
    $unvisited->insert($initKey, 0);
    my $qSize = 1;
    my $i = 0;
    while (my $next = $unvisited->pop()) {
        $i++;
        $qSize--;
        my $state = $known{$next};
        printf('%4d: %4d checked %4d, unchecked %4d, total %4d, checking: %s'."\n", $i, defined $rv ? $rv : -1,
            (scalar keys %known) - $qSize, $qSize, scalar keys %known, stateString($state)) if ($DEBUG);
        last if (defined $rv && $rv < $state->{ManaSpent});
        next if ($state->{Visited});
        $state->{Visited} = 1;
        foreach my $spell (@$spells) {
            next if (!canCast($state, $spell));
            my $newState = doTurns2($state, $spell);
            if ($newState->{BossHP} <= 0) {
                if (! defined $rv || $newState->{ManaSpent} < $rv) {
                    $rv = $newState->{ManaSpent};
                }
                next;
            }
            next if ($newState->{PlayerHP} <= 0);
            my $key = stateKey($newState);
            if (exists $known{$key}) {
                if ($newState->{ManaSpent} < $known{$key}->{ManaSpent}) {
                    $known{$key}->{ManaSpent} = $newState->{ManaSpent};
                    if ($known{$key}->{Visited}) {
                        $known{$key}->{Visited} = 0;
                        $unvisited->insert($key, $newState->{ManaSpent});
                        $qSize++;
                    }
                    else {
                        $unvisited->update($key, $known{$key}->{ManaSpent});
                    }
                }
            }
            else {
                $known{$key} = $newState;
                $unvisited->insert($key, $newState->{ManaSpent});
                $qSize++;
            }
        }
    }
    return $rv;
}

sub doTurns2 {
    my $state = shift;
    my $spell = shift;
    my $rv = copyState($state);
    $rv->{PlayerHP}--;
    return $rv if ($rv->{PlayerHP} <= 0);
    # Player Turn
    if ($rv->{PlayerRecharge}) {
        $rv->{PlayerMana} += 101;
        $rv->{PlayerRecharge}--;
    }
    if ($rv->{PlayerShield}) {
        $rv->{PlayerShield}--;
    }
    if ($rv->{BossPoison}) {
        $rv->{BossHP} -= 3;
        $rv->{BossPoison}--;
    }
    $rv->{PlayerMana} -= $spell->{Cost};
    $rv->{ManaSpent} += $spell->{Cost};
    if ($spell->{Turns}) {
        if ($spell->{Arm}) {
            $rv->{PlayerShield} = $spell->{Turns};
        }
        elsif ($spell->{Mana}) {
            $rv->{PlayerRecharge} = $spell->{Turns};
        }
        elsif ($spell->{Dam}) {
            $rv->{BossPoison} = $spell->{Turns};
        }
        else {
            die "unknown effect spell: ".Dumper($spell);
        }
    }
    else {
        $rv->{BossHP} -= $spell->{Dam};
        if ($spell->{Heal}) {
            $rv->{PlayerHP} += $spell->{Heal};
        }
    }
    # Boss Turn
    return $rv if ($rv->{BossHP} <= 0);
    if ($rv->{PlayerRecharge}) {
        $rv->{PlayerMana} += 101;
        $rv->{PlayerRecharge}--;
    }
    my $shield = 0;
    if ($rv->{PlayerShield}) {
        $shield = 7;
        $rv->{PlayerShield}--;
    }
    if ($rv->{BossPoison}) {
        $rv->{BossHP} -= 3;
        $rv->{BossPoison}--;
    }
    $rv->{PlayerHP} -= 10 - $shield;
    return $rv;
}

sub canCast {
    my $state = shift;
    my $spell = shift;
    return 0 if ($state->{PlayerMana} < $spell->{Cost});
    if ($spell->{Turns}) {
        return 0 if ($spell->{Arm} && $state->{PlayerShield} > 1);
        return 0 if ($spell->{Mana} && $state->{PlayerRecharge} > 1);
        return 0 if ($spell->{Dam} && $state->{BossPoison} > 1);
    }
    return 1;
}

sub copyState {
    my $state = shift;
    my %rv = (
        Visited => 0,
    );
    foreach my $k (qw(ManaSpent PlayerHP PlayerMana PlayerShield PlayerRecharge BossHP BossPoison)) {
        $rv{$k} = $state->{$k};
    }
    return \%rv;
}

sub newState {
    return {
        Visited => 0,
        ManaSpent => $_[1],
        PlayerHP => $_[2],
        PlayerMana => $_[3],
        PlayerShield => $_[4],
        PlayerRecharge => $_[5],
        BossHP => $_[6],
        BossPoison => $_[7],
    }
}

sub stateKey {
    my $state = shift;
    return join("-", ($state->{PlayerHP}, $state->{PlayerMana}, $state->{PlayerShield}, $state->{PlayerRecharge}, $state->{BossHP}, $state->{BossPoison}));
}

sub stateString {
    my $state = shift;
    return sprintf('%16s: %d [%6d] (%2d %3d %d%d) (%d %d)', stateKey($state), $state->{Visited}, $state->{ManaSpent},
        $state->{PlayerHP}, $state->{PlayerMana}, $state->{PlayerShield}, $state->{PlayerRecharge},
        $state->{BossHP}, $state->{BossPoison});

}

sub getSpells {
    return [
        { Name => 'Magic Missile', Cost => 53,              Dam =>    4            },
        { Name => 'Drain',         Cost => 73,              Dam =>    2, Heal => 2 },
        { Name => 'Shield',        Cost => 113, Turns => 6, Arm =>    7            },
        { Name => 'Poison',        Cost => 173, Turns => 6, Dam =>    3            },
        { Name => 'Recharge',      Cost => 229, Turns => 5, Mana => 101            },
    ];
}

sub getPlayer {
    return {
        HP => 50,
        Mana => 500,
    };
}

sub getBoss {
    return {
        HP => 71,
        Damage => 10,
    }
}
