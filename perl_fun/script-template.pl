#!/usr/bin/perl -T
################################################################################
#
# _script_name_.pl
#
# Author:      Danny Wedul
# Date:        April 41, 1988
#
# Description: This is just a template script file.
#              There are several TODO comments indicating what to change.
#              Once they've been handled, those comment lines should be removed.
#              Here is a starter list:
#                   TODO: Replace _script_name_.pl with the name of this file (several places).
#                   TODO: Update the Date field above.
#                   TODO: Replace this description with one describing what your new script does.
#              Here is a general list (covered by other TODO comments):
#                   TODO: Update the usage() sub to provide info about this script.
#                   TODO: Update the parseArgs() sub to properly parse arguments as described in usage().
#                   TODO: Update the mainSub() sub to do what needs to be done.
#              You shouldn't need to update the debug(), quoteAndEscape(), or mainProgram() subs.
#              To output debug messages, use a line like:  debug('message') if ($W);
#              I generally leave the 'Usage:' comment in this commment block with just the --help flag to reduce duplication.
#
# Usage:       ./_script_name_.pl --help
# To debug:    W=1 ./_script_name_.pl [args]
#
################################################################################
use strict;
use warnings;
use Carp qw(cluck);
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
    # TODO: Fill in a short description of what this does.
    my @description = (
        'This is just a template script. It should not be doing anything.',
    );
    # TODO: Fill in the simple usage for this script.
    my @simple_usage = (
        'Usage: '.basename($0).' _args_',
    );
    # TODO: Fill in details about the various arguments/flags this script uses.
    my @details = (
        '  _args_info_',
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
        # TODO: Update for arguments specific to this script, each similar to the block below.
        # TODO: Update the initial regex to look for the specific flags for this parameter.
        # TODO: Update the $pkey value.
        # TODO: Update _some_regex_ to validate (and untaint) the value.
        # TODO: Update _details_ in the error message.
        elsif ($arg_lc =~ m{^(-p|--param)(?:=(.*))?$}) {
            my $arg_given = $1;
            my $arg_val = defined $2 ? $2 : $#args >= 0 && $args[0] !~ m{^-} ? shift(@args) : '';
            my $pkey = 'p';
            debug("  Param {$pkey}: Validating value: ".quoteAndEscape($arg_val).'.') if ($W);
            if ($arg_val =~ m{(_some_regex_)}) {
                my $pval = $1;
                debug("  Param {$pkey}: Is valid as: ".quoteAndEscape($pval).'.') if ($W);
                $params{$pkey} = $pval;
            }
            else {
                my $err = "Invalid $arg_given value: ".quoteAndEscape($arg_val).'. _details_.';
                debug("  Param {$pkey}: $err") if ($W);
                push(@errors, $err);
            }
        }
        else {
            my $err = 'Unknown argument: '.quoteAndEscape($arg).'.';
            debug('  '.$err) if ($W);
            push (@errors, $err);
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

    # TODO: Write the mainSub.

    debug('mainSub: Done.') if ($W);
    return 0;
}

