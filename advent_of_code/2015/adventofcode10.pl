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

#use easyopen qw(openr openread openw openwrite opena openappend slurpFile opend slurpDir);
#use fake_list_data qw(d l L c w date firstname lastname fullname business address city state zip zip4 location amount purpose word rib ri r Time dateAndTime);
#use list_helpers qw(inOneListButNotTheOther inOneStrButNotTheOther strToList strToHash strToListOfHashes isIn);
#use logger qw(printlog dielog dieMessage logfile printlogWithBorder dielogWithBorder addBorderTo);
#use printers; #qw(listOfHashRefs csv hash indent table listForSql breakMultiLine prettyList);
      #$output = printers::table($data, $col_order, $delimeter);
use Prompt;   #imports the prompt($string, $default, $required) function
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
   my @tests = (
      { input => '1', expected => '11' },
      { input => '11', expected => '21' },
      { input => '21', expected => '1211' },
      { input => '1211', expected => '111221' },
      { input => '111221', expected => '312211' },
   );
   my $all_good = 1;
   foreach my $t (@tests) {
      my $actual = transform($t->{input});
      if ($actual ne $t->{expected}) {
         print 'Test failed: '.$t->{input}.': expected: '.$t->{expected}.', actual: '.$actual."\n";
         $all_good = 0;
      }
   }
   if ($all_good) {
      print "All tests passed.\n";
   }
}


sub mainProgram1 {
   my $input = '1113122113';
   foreach my $i (1..40) {
      $input = transform($input);
   }
   print 'Answer: '.length($input)."\n";
}


sub mainProgram2 {
   my $input = '1113122113';
   foreach my $i (1..50) {
      $input = transform($input);
   }
   print 'Answer: '.length($input)."\n";
}


sub transform {
   my $str = shift;
   my $retval = '';
   while ($str =~ /((\d)\2*)/g) {
      my $chunk = $1;
      my $l = length($chunk);
      my $v = substr($chunk, 0, 1);
      $retval .= $l.$v;
   }   
   return $retval;
}


sub getStr {
   return q^^;
}