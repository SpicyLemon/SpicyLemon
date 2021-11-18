#!/usr/bin/perl -T
################################################################################
#
# sorted-find.pl
#
# Author:      Danny Wedul
# Date:        November 17, 2021
#
# Description: This script will take in a <matches file> and <main file> and output
#              lines from the <main file> that match lines in the <matches file>.
#              Each line from the <matches file> is used only once and in order.
#              An <main file> line is a match if the first unused line of
#              <matches file> is a substring of the <main file> line.
#
# Usage:       ./sorted-find.pl --help
# To debug:    W=1 ./sorted-find.pl [args]
#
################################################################################
use strict;
use warnings;
use Carp qw(cluck croak);
use Data::Dumper;   $Data::Dumper::Sortkeys = 1;
use File::Basename;
use POSIX qw(strftime);

use lib qw(.);

# To get debug warnings, set the W environment variable to something true (e.g. W=1).
my $W = exists $ENV{W} ? $ENV{W} : 0;
debug('Debug warnings are on.') if ($W);

my $exit_code = mainProgram(\@ARGV);
debug('Exiting with code ['.$exit_code.']') if ($W);
exit($exit_code);


##############################################################
# Sub          debug
# Usage        debug($message) if ($W);
#    or        debug($message, $show_stack_trace) if ($W);
#
# Parameters   $message = the message to output as a warning.
#                   If it doesn't end in a newline, one will be added.
#                   If undefined, '' is used.
#              $show_stack_trace = a flag.
#                   Set to 1 (true) to cause a stack trace to be output after the message.
#                   Set to 0 (false) to prevent a stack trace even when the message is undef or an empty string.
#                   Defaults to 0 (false).
#
# Description  Outputs a message as a debug warning with a timestamp.
#
# Returns      1 (true) always.
##############################################################
sub debug {
    my $message = shift;
    my $show_stack_trace = shift;

    if (! defined $message) {
        $message = '';
    }

    my $stamp = strftime('%Y-%m-%d %H:%M:%S', localtime);
    my $label = 'debug';
    my $line_format = '%19s %5s: ';
    my $header_1 = sprintf($line_format, $stamp, $label);   # For the first line
    my $header_2 = sprintf($line_format, '', $label);       # For subsequent lines.

    my @lines = split(/^/m, $message);
    my $to_warn = $header_1.( scalar @lines > 0 ? shift(@lines) : '' );
    chomp($to_warn);    # Get rid of the newline if there is one so that we
    $to_warn .= "\n";   # can add one back to make sure there's just one there.
    for my $line (@lines) {
        chomp($line);
        $to_warn .= $header_2.$line."\n";
    }

    if ($show_stack_trace) {
        cluck($to_warn);
    }
    else {
        warn $to_warn;
    }

    return 1;
}


##############################################################
# Sub          quoteAndEscape
# Usage        my $str = quoteAndEscape($val);
#
# Parameters   $val = the value to quote and escape.
#
# Description  Wraps the value in double quotes and escapes stuff as needed.
#              If $val is undefined, or an empty string, "" is returned.
#
# Returns      A string
##############################################################
sub quoteAndEscape {
    my $val = shift;
    if (! defined $val || $val eq '') {
        return '""';
    }
    my $retval = Data::Dumper->new([$val])->Indent(0)->Terse(1)->Useqq(1)->Dump;
    # If $val is numeric, dumper won't wrap it in quotes, so we have to do that on our own.
    if ($retval !~ m{^"}) {
        $retval = '"'.$retval.'"';
    }
    return $retval;
}


##############################################################
# Sub          mainProgram
# Usage        mainProgram(\@ARGV);
#
# Parameters   \@ARGV = a reference to the list of arguments to use for this run.
#
# Description  The main program for this script.
#              It will parse the arguments.
#              If there are errors parsing the arguments, they are printed from here.
#              If help is requested, it is printed from here.
#              If there aren't any argument parsing errors and help isn't requested, then mainSub(\%params) is called,
#                   Where %params is the hash created by parseArgs.
#
# Returns      The exit code that should be used to terminate the script.
##############################################################
sub mainProgram {
    my $args = shift || [];
    debug('mainProgram starting.') if ($W);

    my ($params, $errors, $show_help) = parseArgs($args);
    debug('parseArgs results:'."\n".Data::Dumper->new([$params, $errors, $show_help], [qw(*params *errors show_help)])->Dump) if ($W);

    my $early_exit = undef;

    if (@$errors) {
        print join("\n", @$errors)."\n";
        if ($show_help) {
            print "\n".usage();
        }
        $early_exit = 1;
    }
    elsif ($show_help) {
        print usage();
        $early_exit = 0;
    }

    if (defined $early_exit) {
        debug('mainProgram returning early.') if ($W);
        return $early_exit;
    }

    my $exit_code = mainSub($params);

    debug('mainProgram done.') if ($W);
    return $exit_code;
}


##############################################################
# Sub          usage
# Usage        print usage($simple);
#
# Parameters   $simple = flag. Set to true to get just the simple usage line(s).
#              Omit or set to false to get the full usage including details.
#
# Description  Gets a multi-line string explaining usage of this script.
#
# Returns      a multi-line string.
##############################################################
sub usage {
    my $simple = shift;
    my @description = (
        'This script will take in a <matches file> and <main file> and output',
        'lines from the <main file> that match lines in the <matches file>.',
        'Each line from the <matches file> is used only once and in order.',
        'An <main file> line is a match if the first unused line of',
        '<matches file> is a substring of the <main file> line.',
    );
    my @simple_usage = (
        'Usage: '.basename($0).' <matches file> <main file>',
    );
    my @details = (
        '  <matches file> is a filename of a file containing all the match strings (one match per line).',
        '  <main file> is the filename of the file with the lines to find.',
        '',
        '  A dash can be provided for either file to use STDIN.',
    );
    my @lines = ();
    if ($simple) {
        push(@lines, @simple_usage);
    }
    else {
        push(@lines, @description);
        push(@lines, '');
        push(@lines, @simple_usage);
        push(@lines, '');
        push(@lines, @details);
        push(@lines, '');
    }
    return join("\n", @lines)."\n";
}


##############################################################
# Sub          parseArgs
# Usage        my ($params, $errors, $show_help) = parseArgs($args);
#
# Parameters   \@args = a reference to a list of arguments to parse.
#
# Description  Parses the provided arguments into a hash, checking for errors and/or help.
#
# Returns      A list with three items.
#              0: A hash with keys/values for the various arguments provided.
#              1: A list of errors encountered. This list will be empty if there are no errors.
#              2: Whether or not to show usage (or help).
##############################################################
sub parseArgs {
    my $args_in = shift;
    my @args = @$args_in;
    debug('Args provided: '.Data::Dumper->new([\@args])->Indent(0)->Terse(1)->Useqq(1)->Dump) if ($W);

    my %params = ();
    my @errors = ();
    my $show_help = undef;

    if (scalar @args == 0) {
        my $simple_usage = usage(1);
        chomp($simple_usage);
        push(@errors, $simple_usage);
    }

    # Split any args that start with a single dash and have multiple characters, into multiple flag args (one for each char).
    my @new_args = ();
    for my $arg (@args) {
        if ($arg =~ m{^-([a-zA-Z0-9]{2,})$}) {
            # Split it into a list of characters, and add a dash to the front of each.
            my @flag_args = map { "-$_" } split(//, $1);
            push(@new_args, @flag_args);
        }
        else {
            push(@new_args, $arg)
        }
    }
    @args = @new_args;
    debug('Args to handle '.Data::Dumper->new([\@args])->Indent(0)->Terse(1)->Useqq(1)->Dump) if ($W);

    while (scalar @args > 0) {
        my $arg = shift(@args);
        debug('Handling argument '.quoteAndEscape($arg).'.') if ($W);
        my $arg_lc = lc($arg);  # saves a little bit of processing over using the i flag on regexes.
        if ($arg_lc =~ m{^(?:-h|--help|help)$}) {
            debug('  help argument found.') if ($W);
            $show_help = 1;
        }
        elsif ($arg =~ m{^(.+)$}) {
            my $arg_given = $1;
            if (! defined $params{match_file}) {
                debug("  Param {match_key}: ".quoteAndEscape($arg_given).'.') if ($W);
                $params{match_file} = $arg_given;
            }
            elsif (! defined $params{main_file}) {
                debug("  Param {main_file}: ".quoteAndEscape($arg_given).'.') if ($W);
                $params{main_file} = $arg_given;
            }
            else {
                my $err = 'Unknown argument: '.quoteAndEscape($arg).'.';
                debug('  '.$err) if ($W);
                push(@errors, $err);
            }
        }
        else {
            my $err = 'Unknown argument: '.quoteAndEscape($arg).'.';
            debug('  '.$err) if ($W);
            push (@errors, $err);
        }
    }

    if (! $show_help) {
        if (! defined $params{match_file}) {
            my $err = 'No <matches file> provided.';
            debug('  '.$err) if ($W);
            push(@errors, $err);
        }
        elsif (! defined $params{main_file}) {
            my $err = 'No <main file> provided.';
            debug('  '.$err) if ($W);
            push(@errors, $err);
        }
        elsif ($params{match_file} eq '-' && $params{main_file} eq '-') {
            my $err = 'STDIN can only be used for one file.';
            debug('  '.$err) if ($W);
            push(@errors, $err);
        }
    }

    return (\%params, \@errors, $show_help);
}


##############################################################
# Sub          mainSub
# Usage        my $exit_code = mainSub(\%params);
#
# Parameters   \%params = a reference to a hash defining the parameters of this run.
#                   See the return value of the parseArgs sub for more info.
#
# Description  Does all the work that we want this script to do.
#
# Returns      The exit code that should be used to terminate the script.
##############################################################
sub mainSub {
    my $params = shift;
    debug('mainSub: Starting with params:'."\n".Data::Dumper->new([$params])->Terse(1)->Useqq(1)->Dump) if ($W);

    # Define the file handles as variables.
    my ($FH_MAIN, $FH_MATCH);
    my ($close_fh_main, $close_fh_match);
    # Create a sub to to close them.
    my $closeFiles = sub {
        if ($close_fh_main) {
            close($FH_MATCH) or warn $!;
            $close_fh_main = undef;
        }
        if ($close_fh_match) {
            close($FH_MAIN) or warn $!;
            $close_fh_match = undef;
        }
    };
    # Create a sub to die but close the files first.
    my $closeDie = sub {
        my $msg = shift;
        &$closeFiles();
        croak $msg;
    };
    # Create a sub for opening a file.
    my $openReadFile = sub {
        my $fn = shift;
        if ($fn eq '-') {
            return *STDIN, undef;
        }
        open(my $FH, '<', $fn) or &$closeDie($!);
        return $FH, 1;
    };

    ($FH_MATCH, $close_fh_match) = &$openReadFile($params->{match_file});
    my $match = <$FH_MATCH>;
    if (defined $match) {
        chomp($match);

        ($FH_MAIN, $close_fh_main) = &$openReadFile($params->{main_file});
        while (my $mainline = <$FH_MAIN>) {
            if (index($mainline, $match) != -1) {
                print $mainline;
                $match = <$FH_MATCH>;
                if (defined $match) {
                    chomp($match);
                }
                else {
                    last
                }
            }
        }
    }

    # Close up all the files.
    &$closeFiles();

    debug('mainSub: Done.') if ($W);
    return 0;
}

