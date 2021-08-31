#!/usr/bin/perl -w
# Quick way to test if one or more perl modules are available.
# Usage: puse.pl <module> [<module 2> ...]
use strict;


if (! $ARGV[0] || $ARGV[0] eq '--help' || $ARGV[0] eq '-h') {
   print "usage $0 <module> [<module 2> ...]\n";
   exit(0);
}

my $gc = 0; #good count
my $bc = 0; #bad count

# Whether or not to print extra info as we go or just do summary stuff.
my $eout = ($#ARGV < 5) ? 1 : 0;

foreach my $m (@ARGV) {
   my $str = "require $m;";

   eval {
      eval($str) or die $@;
   };
   if ($@) {
      if ($eout) {
         print "'$m': $@\n";
      }
      else {
         print "failed to require module '$m'\n";
      }
      $bc += 1;
   }
   else {
      my $v = '';
      {
         no strict 'refs';
         if (defined ${ $m . '::VERSION'}) {
            $v = ' Version '.${ $m . '::VERSION'};
         }
      }
      print "'$m' is installed.$v\n";
      $gc += 1;
   }
}

if (! $eout) {
   warn "$gc modules passed.\n".
        "$bc modules failed.\n";
}
