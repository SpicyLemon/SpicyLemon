#!/usr/bin/perl -T
################################################################################
#
# loancalc.pl
#
# Author:      Danny Wedul
# Date:        January 14, 2020
#
# Description: This script will take in provided loan parameters and calculate the missing part.
#
# Usage:       ./loancalc.pl --help
#
################################################################################
use strict;
use warnings;

use lib qw(.);
use LoanInfo;
use LoanInfoBuilder;
use Term;


mainProgram();
exit(10);  #Should never get to this line.

##############################################################
# Sub          usage
# Usage        print usage();
#
# Parameters   none
#
# Description  Gets a multi-line string explaining usage of this script.
#
# Returns      a multi-line string.
##############################################################
sub usage {
    return join("\n", (
        'Usage: loancalc.pl [(-p|--principal) <amount>] [(-m|--monthly-payment) <amount>] [(-r|--rate) <rate>] [(-t|--term|--term-years|--term-months) <term>]',
        '',
        '  Three of the four parameters must be provided. The fourth will be calculated.',
        '    -p or --principal defines the principal amount of the loan.',
        '    -m or --monthly-payment defines the monthly payment to be made.',
        '    -r or --rate defines the rate. It is expected to be in fractional form (e.g. 0.0649).',
        '    -t --term --term-years or --term-months defines the term.',
        '        All of -t --term and --term-years defines the term in years.',
        '        --term-months defines the term in months.',
        '',
        '  The <amount> parameters represent a monetary amount in US format.',
        '    The decimal divider should be a period (.), and commas are ignored.',
        '    A leading dollar sign ($) and any whitespace is also ignored.',
        '    Fractional dollars (cents) are optional.',
        '  The <rate> parameter represents a yearly interest rate in decimal format.',
        '    If the last character is a percent (%), then the input will be treated as a percent and divided by 100.',
        '    Otherwise, it will be treated as a raw interest rate value.',
        '  The <term> parameter represents a term in either months or years depending on which option was provided.',
        '    It can be a whole number, but it can also have a fractional part if desired.',
        '',
    ))."\n";
}


##############################################################
# Sub          mainProgram
# Usage        mainProgram();
#
# Parameters   none
#
# Description  This runs the main program for this script.
#
# Returns      Nothing, exit will be called before it can return.
##############################################################
sub mainProgram {
    my $loan_info = parseArgs(\@ARGV);

    print "\n"
         .$loan_info->toString()
         ."\n";

    exit(0);
}

##############################################################
# Sub          parseArgs
# Usage        my $loan_info = parseArgs(\@args);
#
# Parameters   \@args = a reference to a list of arguments to parse.
#                       Usually this will just be \@ARGV.
#
# Description  Parses the provided arguments and creates a loan_info object if possible.
#
# Returns      A LoanInfo object if possible.
#              Otherwise an error message will be printed exit will be called before returning.
##############################################################
sub parseArgs {
    my $args_in = shift;
    my @args = @$args_in;

    if (scalar @args == 0) {
        print usage();
        exit(0);
    }

    my $retval = LoanInfoBuilder->new();
    my %provided = ();
    my @problems = ();

    while (scalar @args > 0) {
        my $arg_in = shift(@args);
        my $arg = defined $arg_in ? lc($arg_in) : $arg_in;
        if (! defined $arg_in) {
            push (@problems, "Undefined parameter provided.");
        }
        elsif ($arg eq '-h' || $arg eq '--help') {
            print usage();
            exit(0);
        }
        elsif ($arg eq '-p' || $arg eq '--principal' || $arg eq '--principle') {
            my $amount_in = shift(@args);
            if (defined $amount_in && $amount_in =~ m{\S}) {
                my $amount = lc($amount_in);
                $amount =~ s{[\s,]}{}g;
                $amount =~ s{^\$}{}g;
                if ($amount =~ m{^(\d+\.?\d*|\.\d+)$}) {
                    $retval->withPrincipal($1);
                    $provided{principal} = 1;
                }
                elsif ($amount !~ m{^calc(?:ulate)?$}) {
                    push (@problems, "Invalid principal amount [$amount_in] provided after the $arg_in option.");
                }
            }
            else {
                push (@problems, "No principal amount provided after the $arg_in option.");
            }
        }
        elsif ($arg eq '-m' || $arg eq '--monthly-payment' || $arg eq '--payment') {
            my $amount_in = shift(@args);
            if (defined $amount_in && $amount_in =~ m{\S}) {
                my $amount = lc($amount_in);
                $amount =~ s{[\s,]}{}g;
                $amount =~ s{^\$}{}g;
                if ($amount =~ m{^(\d+\.?\d*|\.\d+)$}) {
                    $retval->withMonthlyPayment($1);
                    $provided{monthly_payment} = 1;
                }
                elsif ($amount !~ m{^calc(?:ulate)?$}) {
                    push (@problems, "Invalid monthly payment amount [$amount_in] provided after the $arg_in option.");
                }
            }
            else {
                push (@problems, "No monthly payment amount provided after the $arg_in option.");
            }
        }
        elsif ($arg eq '-r' || $arg eq '--rate') {
            my $rate_in = shift(@args);
            if (defined $rate_in && $rate_in =~ m{\S}) {
                my $rate = lc($rate_in);
                $rate =~ s{\s}{}g;
                if ($rate =~ m{^(\d+\.?\d*|\.\d+)(%)?$}) {
                    if (defined $2 && $2 eq '%') {
                        $retval->withRate($1 / 100);
                    }
                    else {
                        $retval->withRate($1);
                    }
                    $provided{rate} = 1;
                }
                elsif ($rate !~ m{^calc(?:ulate)?$}) {
                    push (@problems, "Invalid rate [$rate_in] provided after the $arg_in option.");
                }
            }
            else {
                push (@problems, "No rate provided after the $arg_in option.");
            }
        }
        elsif ($arg eq '-t' || $arg eq '--term' || $arg eq '--term-years' || $arg eq '--term-months') {
            my $term_in = shift(@args);
            if (defined $term_in && $term_in =~ m{\S}) {
                my $term = $term_in;
                $term =~ s{\s}{}g;
                if ($term =~ m{^(\d+\.?\d*|\.\d+)$}) {
                    if ($arg eq '--term-months') {
                        $retval->withTermInMonths($1);
                    }
                    else {
                        $retval->withTermInYears($1);
                    }
                    $provided{term} = 1;
                }
                elsif ($term !~ m{^calc(?:ulate)?$}) {
                    push (@problems, "Invalid term [$term_in] provided after the $arg_in option.");
                }
            }
            else {
                push (@problems, "No term provided after the $arg_in option.");
            }
        }
        else {
            push (@problems, "Unknown parameter provided: [$arg_in].");
        }
    }

    my $provided_count = 0;
    for my $k (keys %provided) {
        $provided_count += $provided{$k} ? 1 : 0;
    }
    if ($provided_count < 3) {
        push (@problems, "Not enough parameters provided.");
    }

    if (scalar @problems > 0) {
        warn join("\n", @problems)."\n";
        exit(1);
    }

    return $retval->build();
}
