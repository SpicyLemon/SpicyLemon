################################################################################
#
# Term.pm
#
# Author:      Danny Wedul
# Date:        January 14, 2020
#
# Description: This is an object that describes the term of a loan.
#
# Usage:       use Term;
# Creation:    my $term = Term->newYears($term_in_years);
#       or:    my $term = Term->newMonths($term_in_months);
#       or:    my $term = Term->new();
#              $term->setTermInYears($term_in_years);
#       or:    my $term = Term->new();
#              $term->setTermInMonths($term_in_months);
#
# Accessors:   my $years = $term->inYears();                    # Can be either an integer or real number.
#              my $months = $term->inMonths();                  # Can be either an integer or real number.
#              my $whole_years = $term->wholeYears();           # Will only ever be an integer.
#              my $remainder_years = $term->remainderYears();   # Real number less than 1.
#              my $remainder_months = $term->remainderMonths(); # Real number less than 12.
#              my $term_output = $term->toString();             # Examples: '2 years 4.21 months' or '3 years'.
#
################################################################################
package Term;
use strict;
use warnings;

# Usage: my $term = Term->new();
sub new {
    my $class = shift;
    my $self = {
        term_in_years => 0,
        term_in_months => 0,
        whole_years => 0,
        remainder_years => 0,
        remainder_months => 0,
    };
    bless $self, $class;
    return $self;
}

# Usage: my $term = Term->newYears($term_in_years);
sub newYears {
    my $class = shift;
    my $term_in_years = shift;
    my $self = new($class);
    $self->setTermInYears($term_in_years);
    return $self;
}

# Usage: my $term = Term->newMonths($term_in_months);
sub newMonths {
    my $class = shift;
    my $term_in_months = shift;
    my $self = new($class);
    $self->setTermInMonths($term_in_months);
    return $self;
}

# Usage: $term->setTermInYears($term_in_years);
sub setTermInYears {
    my $self = shift;
    my $term_in_years = shift;
    if (defined $term_in_years && $term_in_years =~ m{\w}) {
        if ($term_in_years =~ m{^(\d+\.?\d*|\.\d+)$}) {
            $self->{term_in_years} = $1 + 0;
            $self->{term_in_months} = $self->{term_in_years} * 12;
            $self->enforceStandards();
        }
        else {
            die "Invalid term value provided to setTermInYears [$term_in_years]";
        }
    }
    else {
        warn "No term value provided to setTermInYears."
    }
    return $self;
}

# Usage: $term->setTermInMonths($term_in_months);
sub setTermInMonths {
    my $self = shift;
    my $term_in_months = shift;
    if (defined $term_in_months && $term_in_months =~ m{\w}) {
        if ($term_in_months =~ m{^(\d+\.?\d*|\.\d+)$}) {
            $self->{term_in_months} = $1 + 0;
            $self->{term_in_years} = $self->{term_in_months} / 12;
            $self->enforceStandards();
        }
        else {
            die "Invalid term value provided to setTermInMonths [$term_in_months]";
        }
    }
    else {
        warn "No term value provided to setTermInMonths."
    }
    return $self;
}

# Usage: $term->enforceStandards();
# You probably only need this if you're modifying the fields manually.
sub enforceStandards {
    my $self = shift;
    $self->{term_in_years} = sprintf("%.4f", $self->{term_in_years});
    $self->{term_in_months} = sprintf("%.3f", $self->{term_in_months});
    $self->{whole_years} = int($self->{term_in_years});
    $self->{remainder_years} = $self->{term_in_years} - $self->{whole_years};
    if ($self->{remainder_years} <= 0.0027) {
        $self->{remainder_years} = 0;
    }
    elsif ($self->{remainder_years} >= 0.9973) {
        $self->{remainder_years} = 0;
        $self->{whole_years} += 1;
    }
    $self->{remainder_years} = sprintf("%.4f", $self->{remainder_years});
    $self->{remainder_months} = sprintf("%.3f", $self->{remainder_years} * 12);
}

# Usage: my $years = $term->inYears();
# Can be either integer or real.
sub inYears {
    my $self = shift;
    return $self->{term_in_years};
}

# Usage: my $months = $term->inMonths();
# Can be either integer or real.
sub inMonths {
    my $self = shift;
    return $self->{term_in_months};
}

# Usage: my $years = $term->wholeYears();
# Integer part of the term in years.
sub wholeYears {
    my $self = shift;
    return $self->{whole_years};
}

# Usage: my $remainder_years = $term->remainderYears();
# Fractional part of the term in years.
# Real number less than 1.
sub remainderYears {
    my $self = shift;
    return $self->{remainder_years};
}

# Usage: my $remainder_months = $term->remainderMonths();
# Fractional part of the term in years, converted to months.
# Real number less than 12.
sub remainderMonths {
    my $self = shift;
    return $self->{remainder_months};
}

# Usage: my $ouptut = $term->toString();
# Examples: '3 years', '2 years 11.55 months', '1 year 8.3 months'
sub toString {
    my $self = shift;
    my $years = $self->{whole_years};
    my $months = 0;
    if ($self->{remainder_months} >= 11.995) {
        $years += 1;
    }
    elsif ($self->{remainder_months} >= 0.01) {
        $months = sprintf("%.2f", $self->{remainder_months});
        $months =~ s{0+$}{};
        $months =~ s{\.$}{};
    }
    my $year_label = $years eq '1' ? 'year' : 'years';
    my $retval = $years . ' ' . $year_label;
    if ($months ne '0') {
        my $month_label = $months eq '1' ? 'month' : 'months';
        $retval = $retval . ' ' . $months . ' ' . $month_label;
    }
    return $retval;
}

1;
