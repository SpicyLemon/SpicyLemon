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
    my @keyKeys = qw( principal monthly_payment term rate );
    for my $k (@keyKeys) {
        if (defined $params->{$k}) {
            $self->{$k} = $params->{$k};
        }
    }
    bless $self, $class;
    my @missing = ();
    for my $k (@keyKeys) {
        if (! defined $self->{$k}) {
            push (@missing, $k);
        }
    }
    if (scalar @missing == 0) {
        $self->updateCosts();
    }
    elsif (scalar @missing == 1) {
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
    return $self;
}

# Usage: $loan_info->calculatePrincipal();
sub calculatePrincipal {
    my $self = shift;
    my $monthly_payment = $self->{monthly_payment};
    my $monthly_rate = $self->{rate} / 12;
    my $term_in_months = $self->{term}->inMonths();
    $self->{principal} = $monthly_payment / ( $monthly_rate / ( 1 - ( 1 + $monthly_rate ) ** ( -1 * $term_in_months ) ) );
    $self->{last_calc} = 'principal';
    $self->updateCosts();
    return $self->{principal};
}

# Usage: $loan_info->calculateMonthlyPayment();
sub calculateMonthlyPayment {
    my $self = shift;
    my $principal = $self->{principal};
    my $monthly_rate = $self->{rate} / 12;
    my $term_in_months = $self->{term}->inMonths();
    $self->{monthly_payment} = $principal * $monthly_rate / ( 1 - ( 1 + $monthly_rate ) ** ( -1 * $term_in_months ) );
    $self->{last_calc} = 'monthly_payment';
    $self->updateCosts();
    return $self->{monthly_payment};
}

# Usage: $loan_info->calculateTerm();
sub calculateTerm {
    my $self = shift;
    my $principal = $self->{principal};
    my $monthly_rate = $self->{rate} / 12;
    my $monthly_payment = $self->{monthly_payment};
    $self->{term} = Term->newMonths(log( $monthly_payment / ( $monthly_payment - $principal * $monthly_rate ) ) / log( 1 + $monthly_rate ));
    $self->{last_calc} = 'term';
    $self->updateCosts();
    return $self->{term}->inYears();
}

# Usage: $loan_info->calculateRate();
sub calculateRate {
    my $self = shift;
    $self->updateCosts();
    my $target = sprintf("%.3f", $self->{principal});
    my $actual = '';
    my @guesses = ();
    my $max_iterations = 20;
    my $sign = undef;
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
    $self->{rate} = $guesses[0]->{rate};
    return undef;
}

# Usage: $loan_info->updateCosts();
sub updateCosts {
    my $self = shift;
    $self->{total_paid} = $self->{monthly_payment} * $self->{term}->inMonths();
    $self->{total_interest} = $self->{total_paid} - $self->{principal};
    $self->{average_interest_per_payment} = $self->{total_interest} / $self->{term}->inMonths();
    return $self;
}

# Usage: my $principal = $loan_info->principal();
#    or: $loan_info->principal($new_principal);
sub principal {
    my $self = shift;
    my $input = shift;
    if (defined $input) {
        $self->{principal} = $input;
    }
    return $self->{principal};
}

# Usage: my $monthly_payment = $loan_info->monthlyPayment();
#    or: $loan_info->monthlyPayment($new_monthly_payment);
sub monthlyPayment {
    my $self = shift;
    my $input = shift;
    if (defined $input) {
        $self->{monthly_payment} = $input;
    }
    return $self->{monthly_payment};
}

# Usage: my $rate = $loan_info->rate();
#    or: $loan_info->rate($new_rate);
sub rate {
    my $self = shift;
    my $input = shift;
    if (defined $input) {
        $self->{rate} = $input;
    }
    return $self->{rate};
}

# Usage: my $term = $loan_info->term();
#    or: $loan_info->term($new_term);
sub term {
    my $self = shift;
    my $input = shift;
    if (defined $input) {
        $self->{term} = $input;
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

# Usage: my $output = $loan_info->toString();
# Resulting string is a multi-lined report with info about the data in here.
sub toString {
    my $self = shift;
    my $bottom_field = shift || $self->{last_calc};
    my %lines = (
        principal =>       '      Principal: ' . sprintf("%.2f", $self->{principal}),
        rate =>            '           Rate: ' . sprintf("%.6f", $self->{rate}),
        term =>            '           Term: ' . $self->{term}->toString(),
        monthly_payment => 'Monthly Payment: ' . sprintf("%.2f", $self->{monthly_payment}),
        divider =>         '--------------------------------',
        total_paid =>      '     Total Paid: ' . sprintf("%.2f", $self->{total_paid}),
        total_interest =>  ' Total Interest: ' . sprintf("%.2f", $self->{total_interest}),
        ave_int_per_pay => 'Int per payment: ' . sprintf("%.2f", $self->{average_interest_per_payment}),
    );
    my @retval = ();
    for my $k (qw( principal rate term monthly_payment )) {
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

1;
