################################################################################
#
# LoanInfo.pm
#
# Author:      Danny Wedul
# Date:        January 14, 2020
#
# Description: This describes a LoanInfo object.
#              This will help facilitate some calculations related to loans.
#
# Usage:       use LoanInfo;
# Creation:    my $loan_info = LoanInfo->new(\%params);
#              It's probably easier to use the LoanInfoBuilder.
#              Valid param keys:
#                   principal, monthly_payment, term, rate
#              If exactly three of those are given, the fourth will be calculated.
#
################################################################################
package LoanInfo;
use strict;
use warnings;
use Carp;
use Scalar::Util 'blessed';
use Term;

# Usage: my $loan_info = LoanInfo->new(\%params);
# Valid param keys: principal, monthly_payment, term, rate
sub new {
    my $class = shift;
    my $params = shift;
    my $self = {
        principal => undef,
        monthly_payment => undef,
        term => undef,
        rate => undef,
        total_paid => undef,
        total_interest => undef,
        last_calc => undef,
    };
    for my $k (qw( principal monthly_payment term rate )) {
        if (defined $params->{$k}) {
            $self->{$k} = $params->{$k};
        }
    }
    bless $self, $class;
    if (defined $self->{term} && ! $self->_termIsTerm()) {
        croak "The 'term' parameter provided to the LoanInfo constructor is not a Term object."
    }
    $self->calculateMissing();
    return $self;
}

# Usage: $loan_info->calculateMissing();
sub calculateMissing {
    my $self = shift;
    $self->{last_calc} = undef;
    my @missing = ();
    for my $k (qw( principal monthly_payment term rate )) {
        if (! defined $self->{$k}) {
            push (@missing, $k);
        }
    }
    if (scalar @missing == 1) {
        if ($missing[0] eq 'principal') {
            $self->calculatePrincipal();
        }
        elsif ($missing[0] eq 'monthly_payment') {
            $self->calculateMonthlyPayment();
        }
        elsif ($missing[0] eq 'term') {
            $self->calculateTerm();
        }
        elsif ($missing[0] eq 'rate') {
            $self->calculateRate();
        }
    }
    else {
        $self->updateCosts();
    }
    return $self;
}

# Usage: $loan_info->calculatePrincipal();
sub calculatePrincipal {
    my $self = shift;
    my $principal = undef;
    my $last_calc = undef;
    if (defined $self->{monthly_payment} && defined $self->{rate} && $self->_termIsTerm()) {
        my $monthly_rate = $self->{rate} / 12;
        $principal = $self->{monthly_payment} / ( $monthly_rate / ( 1 - ( 1 + $monthly_rate ) ** ( -1 * $self->{term}->inMonths() ) ) );
        $last_calc = 'principal';
    }
    $self->{principal} = $principal;
    $self->{last_calc} = $last_calc;
    $self->updateCosts();
    return $self->{principal};
}

# Usage: $loan_info->calculateMonthlyPayment();
sub calculateMonthlyPayment {
    my $self = shift;
    my $monthly_payment = undef;
    my $last_calc = undef;
    if (defined $self->{principal} && defined $self->{rate} && $self->_termIsTerm()) {
        my $monthly_rate = $self->{rate} / 12;
        $monthly_payment = $self->{principal} * $monthly_rate / ( 1 - ( 1 + $monthly_rate ) ** ( -1 * $self->{term}->inMonths() ) );
        $last_calc = 'monthly_payment';
    }
    $self->{monthly_payment} = $monthly_payment;
    $self->{last_calc} = $last_calc;
    $self->updateCosts();
    return $self->{monthly_payment};
}

# Usage: $loan_info->calculateTerm();
sub calculateTerm {
    my $self = shift;
    my $term = undef;
    my $last_calc = undef;
    if (defined $self->{principal} && defined $self->{rate} && defined $self->{monthly_payment}) {
        my $monthly_rate = $self->{rate} / 12;
        $term = Term->newMonths(log( $self->{monthly_payment} / ( $self->{monthly_payment} - $self->{principal} * $monthly_rate ) ) / log( 1 + $monthly_rate ));
        $last_calc = 'term';
    }
    $self->{term} = $term;
    $self->{last_calc} = $last_calc;
    $self->updateCosts();
    return $self->{term};
}

# Usage: $loan_info->calculateRate();
sub calculateRate {
    my $self = shift;
    my $rate = undef;
    my $last_calc = undef;
    $self->updateCosts();
    if (defined $self->{principal} && $self->_termIsTerm() && defined $self->{monthly_payment}) {
        my $target = sprintf("%.3f", $self->{principal});
        my $actual = '';
        my @guesses = ();
        my $max_iterations = 20;
        my $iteration = 0;
        while ($target ne $actual && $iteration < $max_iterations) {
            $iteration += 1;
            my $rate_guess = undef;
            if (scalar @guesses == 0) {
                $rate_guess = $self->{total_paid} / $self->{principal} - 1;
            }
            elsif (scalar @guesses == 1) {
                @guesses = sort { $a->{diff_abs} <=> $b->{diff_abs} } @guesses;
                if ($guesses[0]->{diff} < 0) {
                    $rate_guess = $guesses[0]->{rate} * 0.9;
                }
                else {
                    $rate_guess = $guesses[0]->{rate} * 1.1;
                }
            }
            else {
                my $p = $self->{principal};
                my $r0 = $guesses[0]->{rate};
                my $p0 = $guesses[0]->{principal};
                my $r1 = $guesses[1]->{rate};
                my $p1 = $guesses[1]->{principal};
                $rate_guess = ( $r1*($p0-$p) + $r0*($p-$p1) ) / ($p0 - $p1);
            }
            my $monthly_rate = $rate_guess / 12;
            my $principal = sprintf("%.3f", $self->{monthly_payment} / ( $monthly_rate / ( 1 - ( 1 + $monthly_rate ) ** ( -1 * $self->{term}->inMonths() ) ) ));
            $actual = sprintf("%.3f", $principal);
            my %guess = (
                iteration => $iteration,
                rate => $rate_guess,
                principal => $principal,
                diff => $principal - $self->{principal},
            );
            $guess{diff_abs} = abs($guess{diff});
            unshift (@guesses, \%guess);
        }
        if ($target ne $actual) {
            my @message = ('Unable to calculate rate from provided parameters:',
                           '      Principal: '.$self->{principal},
                           'Monthly Payment: '.$self->{monthly_payment},
                           '           Term: '.$self->{term}->toString(),
                           '--------------------------------------------',
                           "i\trate\tprincipal\tdiff");
            push (@message, map { join("\t", ($_->{iteration},
                                              sprintf("%.6f", $_->{rate}),
                                              sprintf("%.2f", $_->{principal}),
                                              sprintf("%.2f", $_->{diff_abs}))) }
                            sort { $a->{iteration} <=> $b->{iteration} } @guesses );
            warn join("\n", @message)."\n";
            exit(2);
        }
        $rate = $guesses[0]->{rate};
        $last_calc = 'rate'
    }
    $self->{rate} = $rate;
    $self->{last_calc} = $last_calc;
    return $self->{rate};
}

# Usage: $loan_info->updateCosts();
sub updateCosts {
    my $self = shift;
    for my $k (qw( total_paid total_interest average_interest_per_payment )) {
        if (defined $self->{$k}) {
            $self->{$k} = undef;
        }
    }
    if (defined $self->{monthly_payment} && $self->_termIsTerm()) {
        $self->{total_paid} = $self->{monthly_payment} * $self->{term}->inMonths();
    }
    if (defined $self->{total_paid} && defined $self->{principal}) {
        $self->{total_interest} = $self->{total_paid} - $self->{principal};
    }
    if (defined $self->{total_interest} && $self->_termIsTerm()) {
        $self->{average_interest_per_payment} = $self->{total_interest} / $self->{term}->inMonths();
    }
    return $self;
}

# Usage: my $principal = $loan_info->principal();
#    or: $loan_info->principal($new_principal);
sub principal {
    my $self = shift;
    if (scalar @_) {
        $self->{principal} = shift;
        $self->updateCosts();
    }
    return $self->{principal};
}

# Usage: my $monthly_payment = $loan_info->monthlyPayment();
#    or: $loan_info->monthlyPayment($new_monthly_payment);
sub monthlyPayment {
    my $self = shift;
    if (scalar @_) {
        $self->{monthly_payment} = shift;
        $self->updateCosts();
    }
    return $self->{monthly_payment};
}

# Usage: my $rate = $loan_info->rate();
#    or: $loan_info->rate($new_rate);
sub rate {
    my $self = shift;
    if (scalar @_) {
        $self->{rate} = shift;
        $self->updateCosts();
    }
    return $self->{rate};
}

# Usage: my $term = $loan_info->term();
#    or: $loan_info->term($new_term);
sub term {
    my $self = shift;
    if (scalar @_) {
        $self->{term} = shift;
        if (defined $self->{term} && ! $self->_termIsTerm()) {
            croak "Value provided to set as the term must be a Term object."
        }
        $self->updateCosts();
    }
    return $self->{term};
}

# Usage: my $total_paid = $loan_info->totalPaid();
sub totalPaid {
    my $self = shift;
    return $self->{total_paid};
}

# Usage: my $total_interest = $loan_info->totalInterest();
sub totalInterest {
    my $self = shift;
    return $self->{total_Interest};
}

# Usage: my $ave_interest_per_payment = $loan_info->averageInterestPerPayment();
sub averageInterestPerPayment {
    my $self = shift;
    return $self->{average_interest_per_payment};
}

# Usage: my $output = $loan_info->toString();
# Resulting string is a multi-lined report with info about the data in here.
sub toString {
    my $self = shift;
    my $bottom_field = shift || $self->{last_calc};
    my %values = (
        term =>            $self->termToString(),
        rate =>            $self->rateToString(),
        principal =>       $self->principalToString(),
        monthly_payment => $self->monthlyPaymentToString(),
        total_paid =>      $self->totalPaidToString(),
        total_interest =>  $self->totalInterestToString(),
        ave_int_per_pay => $self->averageInterestPerPaymentToString(),
    );
    my $field_length = 0;
    for my $k (qw( rate principal monthly_payment total_paid total_interest ave_int_per_pay )) {
        my $new_length = length $values{$k};
        if ($new_length > $field_length) {
            $field_length = $new_length;
        }
    }
    for my $k (qw( term rate )) {
        $values{$k} = sprintf("%*s", $field_length, $values{$k});
    }
    if (length $values{term} < $field_length) {
        $values{term} = sprintf("%*s", $field_length, $values{term});
    }
    $values{rate} = sprintf("%*s", $field_length, $values{rate});
    for my $k (qw( principal monthly_payment total_paid total_interest ave_int_per_pay )) {
        $values{$k} = __stretchMoney($values{$k}, $field_length);
    }
    my %lines = (
        term =>            '           Term: ' . $values{term},
        rate =>            '           Rate: ' . $values{rate},
        principal =>       '      Principal: ' . $values{principal},
        monthly_payment => 'Monthly Payment: ' . $values{monthly_payment},
        divider =>         '--------------------------------',
        total_paid =>      '     Total Paid: ' . $values{total_paid},
        total_interest =>  ' Total Interest: ' . $values{total_interest},
        ave_int_per_pay => 'Int per payment: ' . $values{ave_int_per_pay},
    );
    my @retval = ();
    for my $k (qw( term rate principal monthly_payment )) {
        if (! defined $bottom_field || $k ne $bottom_field) {
            push (@retval, $lines{$k});
        }
    }
    push (@retval, $lines{divider});
    if (defined $bottom_field && exists $lines{$bottom_field}) {
        push (@retval, $lines{$bottom_field});
    }
    for my $k (qw( total_paid total_interest ave_int_per_pay )) {
        push (@retval, $lines{$k});
    }
    return join("\n", @retval)."\n";
}

sub principalToString {
    my $self = shift;
    return __formatMoney($self->{principal});
}

sub rateToString {
    my $self = shift;
    return __formatPercent($self->{rate});
}

sub termToString {
    my $self = shift;
    return $self->_termIsTerm() ? $self->{term}->toString() : '';
}

sub monthlyPaymentToString {
    my $self = shift;
    return __formatMoney($self->{monthly_payment});
}

sub totalPaidToString {
    my $self = shift;
    return __formatMoney($self->{total_paid});
}

sub totalInterestToString {
    my $self = shift;
    return __formatMoney($self->{total_interest});
}

sub averageInterestPerPaymentToString {
    my $self = shift;
    return __formatMoney($self->{average_interest_per_payment});
}

# Checks if the term parameter is a Term object.
# Returns 1 if defined and blessed as a Term object. 0 otherwise.
sub _termIsTerm {
    my $self = shift;
    return defined $self->{term} && blessed $self->{term} && $self->{term}->isa('Term') ? 1 : 0;
}

# Non-object method.
# Usage: my $money_string = __formatMoney($value);
sub __formatMoney {
    my $value = shift;
    my $retval = '';
    if (defined $value) {
        $retval = sprintf("%.2f", $value);
        # add commas
        $retval =~ s/(^[-+]?\d+?(?=(?>(?:\d{3})+)(?!\d))|\G\d{3}(?=\d))/$1,/g;
        $retval = '$ '.$retval;
    }
    return $retval;
}

# Non-object method.
# Usage: my $string = __stretchMoney($money, $length);
sub __stretchMoney {
    my $retval = shift;
    my $length = shift;
    if (defined $retval && $retval =~ m{^\$\s+\d}) {
        while (length $retval < $length) {
            $retval =~ s{^\$ }{\$  };
        }
    }
    return $retval;
}

# Non-object method.
# Usage: my $percent_string = __formatPercent($value);
sub __formatPercent {
    my $value = shift;
    my $retval = '';
    if (defined $value) {
        $retval = sprintf("%.4f", $value * 100) . ' %';
    }
    return $retval;
}

1;
