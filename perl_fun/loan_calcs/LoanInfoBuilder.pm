################################################################################
#
# LoanInfoBuilder.pm
#
# Author:      Danny Wedul
# Date:        January 14, 2020
#
# Description: A builder to help in the creation of a LoanInfo object.
#
# Usage:       use LoanInfoBuilder;
#              my $builder = LoanInfoBuilder->new();
#              $builder->withPrincipal($principal);
#              $builder->withMonthlyPayment($monthly_payment);
#              $builder->withTermInYears($term_in_years);
#              $builder->withTermInMonths($term_in_months);
#              $builder->withTerm(Term->newYears($term_in_years));
#              $builder->withRate($rate);
#              my $loan_info = $builder->build();
#
# Notes:       - All of the with* methods return the builder so that they can be
#                chained if desired.
#              - All of the withTerm* methods are just different ways of setting
#                the same thing. If, for example, withTermInYears is provided,
#                and later, withTermInMonths is provided, then the second one
#                will overwrite what was set with the first one.
#
################################################################################
package LoanInfoBuilder;
use strict;
use warnings;
use Carp;
use Scalar::Util 'blessed';
use LoanInfo;
use Term;

# Usage: my $builder = LoanInfoBuilder->new();
sub new {
    my $class = shift;
    my $self = {};
    bless $self, $class;
    return $self;
}

# Usage: $builder->withPrincipal($principal);
sub withPrincipal {
    my $self = shift;
    $self->{principal} = shift;
    return $self;
}

# Usage: $builder->withMonthlyPayment($monthly_payment);
sub withMonthlyPayment {
    my $self = shift;
    $self->{monthly_payment} = shift;
    return $self;
}

# Usage: $builder->withTermInYears($term_in_years);
# Note: Replaces anything set through the other term setters.
sub withTermInYears {
    my $self = shift;
    my $term_in_years = shift;
    if (defined $term_in_years) {
        $self->withTerm(Term->newYears($term_in_years));
    }
    else {
        $self->withTerm(undef);
    }
    return $self;
}

# Usage: $builder->withTermInMonths($term_in_months);
# Note: Replaces anything set through the other term setters.
sub withTermInMonths {
    my $self = shift;
    my $term_in_months = shift;
    if (defined $term_in_months) {
        $self->withTerm(Term->newMonths($term_in_months));
    }
    else {
        $self->withTerm(undef);
    }
    return $self;
}

# Usage: $builder->withTerm(Term->newMonths($term_in_months));
# Note: Replaces anything set through the other term setters.
sub withTerm {
    my $self = shift;
    $self->{term} = shift;
    if ( defined $self->{term} && ! (blessed $self->{term} && $self->{term}->isa('Term')) ) {
        croak "Invalid value provided to LoanInfoBuilder withTerm. Value must be Term object.";
    }
    return $self;
}

# Usage: $builder->withRate($rate);
sub withRate {
    my $self = shift;
    $self->{rate} = shift;
    return $self;
}

# Usage: my $loan_info = $builder->build();
sub build {
    my $self = shift;
    my %params = ();
    for my $k (qw( term principal monthly_payment rate )) {
        if (defined $self->{$k}) {
            $params{$k} = $self->{$k};
        }
    }
    return LoanInfo->new(\%params);
}

1;
