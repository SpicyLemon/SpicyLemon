#!/usr/bin/perl
use strict;
use warnings;
#use Benchmark ':hireswallclock';
#use Carp;
use Data::Dumper;  $Data::Dumper::Sortkeys = 1;
use DateTime;
#use Date::Parse; #imports str2time that takes in a string, and outputs an epoch.
use Digest::MD5 qw(md5_hex);
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
   my ($k1, $n1) = ('abcdef', '609043');
   my $v1 = $k1.$n1;
   my $m1 = md5_hex($v1);
   print "$v1 => $m1\n";
   my ($k2, $n2) = ('pqrstuv', '1048970');
   my $v2 = $k2.$n2;
   my $m2 = md5_hex($v2);
   print "$v2 => $m2\n";
}

sub mainProgram1 {
   my $k = 'iwrupvqb';
   my $n = 0;
   my $keep_going = 1;
   while ($keep_going) {
      my $v = $k.$n;
      my $m = md5_hex($v);
      if ($m =~ /^0{5}/) {
         $keep_going = 0;
         print "'$k'.'$n' => '$v' => '$m'\n";
      }
      else {
         $n += 1;
      }
   }
}


sub mainProgram2 {
   my $k = 'iwrupvqb';
   my $n = 346386;
   my $keep_going = 1;
   while ($keep_going) {
      my $v = $k.$n;
      my $m = md5_hex($v);
      if ($m =~ /^0{6}/) {
         $keep_going = 0;
         print "'$k'.'$n' => '$v' => '$m'\n";
      }
      else {
         $n += 1;
      }
   }
}

