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

#use easyopen qw(openr openread openw openwrite opena openappend slurpFile opend slurpDir);
#use fake_list_data qw(d l L c w date firstname lastname fullname business address city state zip zip4 location amount purpose word rib ri r Time dateAndTime);
#use list_helpers qw(inOneListButNotTheOther inOneStrButNotTheOther strToList strToHash strToListOfHashes isIn);
#use logger qw(printlog dielog dieMessage logfile printlogWithBorder dielogWithBorder addBorderTo);
#use printers; #qw(listOfHashRefs csv hash indent table listForSql breakMultiLine prettyList);
      #$output = printers::table($data, $col_order, $delimeter);
#use Prompt;   #imports the prompt($string, $default, $required) function
#use stats qw(getStats countMinMax mean median modes);

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
   my $perms = getPermutations([0, 1, 2]);
   foreach my $p (@$perms) {
      print join(', ', @$p)."\n";
   }
}


sub mainProgram1 {
   my $map = getMap();
   my $routes = getPermutations([keys %$map]);
   my $min_length = 999999;
   my @min_routes = ();
   foreach my $route (@$routes) {
      my $route_length = 0;
      my $cur_city = $route->[0];
      for (my $i = 1; $i < scalar @$route; $i++) {
         $route_length += $map->{$cur_city}->{$route->[$i]};
         $cur_city = $route->[$i];
      }
      if ($route_length == $min_length) {
         push (@min_routes, $route);
      }
      elsif ($route_length < $min_length) {
         $min_length = $route_length;
         @min_routes = ( $route );
      }
   }
   print "Routes: \n";
   foreach my $route (@min_routes) {
      print join(' -> ', @$route)."\n";
   }
   print "route length: $min_length\n";
}


sub mainProgram2 {
   my $map = getMap();
   my $routes = getPermutations([keys %$map]);
   my $max_length = 0;
   my @max_routes = ();
   foreach my $route (@$routes) {
      my $route_length = 0;
      my $cur_city = $route->[0];
      for (my $i = 1; $i < scalar @$route; $i++) {
         $route_length += $map->{$cur_city}->{$route->[$i]};
         $cur_city = $route->[$i];
      }
      if ($route_length == $max_length) {
         push (@max_routes, $route);
      }
      elsif ($route_length > $max_length) {
         $max_length = $route_length;
         @max_routes = ( $route );
      }
   }
   print "Routes: \n";
   foreach my $route (@max_routes) {
      print join(' -> ', @$route)."\n";
   }
   print "route length: $max_length\n";
}


sub getPermutations {
   my $list = shift;
   my @retval = ();
   if (scalar @$list == 1) {
      push (@retval, [ $list->[0] ]);
   }
   else {
      for(my $i = 0; $i < scalar @$list; $i++) {
         my @sublist = ();
         for(my $j = 0; $j < scalar @$list; $j++) {
            if ($i != $j) {
               push (@sublist, $list->[$j]);
            }
         }
         foreach my $subpermutation (@{getPermutations(\@sublist)}) {
            push (@retval, [ $list->[$i], @$subpermutation ]);
         }
      }
   }
   return \@retval;
}

sub getMap {
   my %retval = ();
   foreach my $path (split(/\n/, getStr())) {
      if ($path =~ /(\w+) to (\w+) = (\d+)/) {
         my ($city1, $city2, $distance) = ($1, $2, $3);
         if (! exists $retval{$city1}) {
            $retval{$city1} = {};
         }
         if (! exists $retval{$city2}) {
            $retval{$city2} = {};
         }
         $retval{$city1}->{$city2} = $distance;
         $retval{$city2}->{$city1} = $distance;
      }
   }
   return \%retval;
}

sub getStr {
   return q^Faerun to Norrath = 129
Faerun to Tristram = 58
Faerun to AlphaCentauri = 13
Faerun to Arbre = 24
Faerun to Snowdin = 60
Faerun to Tambi = 71
Faerun to Straylight = 67
Norrath to Tristram = 142
Norrath to AlphaCentauri = 15
Norrath to Arbre = 135
Norrath to Snowdin = 75
Norrath to Tambi = 82
Norrath to Straylight = 54
Tristram to AlphaCentauri = 118
Tristram to Arbre = 122
Tristram to Snowdin = 103
Tristram to Tambi = 49
Tristram to Straylight = 97
AlphaCentauri to Arbre = 116
AlphaCentauri to Snowdin = 12
AlphaCentauri to Tambi = 18
AlphaCentauri to Straylight = 91
Arbre to Snowdin = 129
Arbre to Tambi = 53
Arbre to Straylight = 40
Snowdin to Tambi = 15
Snowdin to Straylight = 99
Tambi to Straylight = 70^;
}