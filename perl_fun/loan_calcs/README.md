# SpicyLemon / perl_fun / loan_calcs
This directory contains some loan information calculation stuff written in Perl.

## Contents

* `loancalc.pl` - This is the script that facilitates the loan info calculation interaction.
* `Term.pm` - This module describes a Term object, making it a bit easier to go to and from years and months.
* `LoanInfo.pm` - This module describes the info associated with a Loan such as Principal, Term, Monthly Payment, and Rate.
* `LoanInfoBuilder.pm` - This is a module that makes it a little easier to create a LoanInfo object.

## Notes

### Warning

Financial instituions pay a lot of attention to rounding.
These calculations do not. There are just too many different ways to do it.
In here, rounding is generally done as late as possible,
and is done with `sprintf` so it might act differently on different systems.
That means that results are not guaranteed, but they should at least be close enough to make personal decisions.

### Assumptions

The calculations in here assume a few things:

1.  Payments will be made on a monthly bases.
1.  Interest is calculated monthly (as opposed to yearly, daily, or continuously).
1.  The monthly payment amounts are constant.
1.  The rate is fixed (as opposed to variable).
1.  Repayment of the loan begins when the loan is received.

## loancalc.pl

For usage, see `./loancalc.pl --help`.

Four pieces of information can be provided to this script.

* Principal - The amount of money being received for the loan.
* Monthly Payment - The amount of money that will be paid each month towards the loan.
* Rate - The yearly interest rate
* Term - The length of the loan, usually in years, but can also be supplied as months.

The purpose of the script is that you can supply any three of them, and the fourth will be calculated for you.

### Examples:

```bash
$ ./loancalc.pl -r 10% -t 3 -p 25000

           Term:     3 years
           Rate:   10.0000 %
      Principal: $ 25,000.00
--------------------------------
Monthly Payment: $    806.68
     Total Paid: $ 29,040.47
 Total Interest: $  4,040.47
Int per payment: $    112.24
```

```bash
$ ./loancalc.pl -r .0905 -t 5 -m 2000

           Term:      5 years
           Rate:     9.0500 %
Monthly Payment: $   2,000.00
--------------------------------
      Principal: $  96,234.21
     Total Paid: $ 120,000.00
 Total Interest: $  23,765.79
Int per payment: $     396.10
```

```bash
$ ./loancalc.pl -p 35000 -m 1000 -r 0.0849

           Rate:    8.4900 %
      Principal: $ 35,000.00
Monthly Payment: $  1,000.00
--------------------------------
           Term: 3 years 4.36 months
     Total Paid: $ 40,357.00
 Total Interest: $  5,357.00
Int per payment: $    132.74
```

```bash
$ ./loancalc.pl -p 10000 -m 500 -t 2

           Term:     2 years
      Principal: $ 10,000.00
Monthly Payment: $    500.00
--------------------------------
           Rate:   18.1570 %
     Total Paid: $ 12,000.00
 Total Interest: $  2,000.00
Int per payment: $     83.33
```

## Term.pm

This module describes a length of time with loan length in mind.

### Usage

Make sure that the Term.pm file is in your Perl path, then `use Term;`.

Values provided to and retrived from the Term objects do not have to be whole numbers.

### Constructors

Given a term length in years:
```Perl
my $term = Term->newYears($term_in_years);
```

Given a term length in months:
```Perl
my $term = Term->newMonths($term_in_months);
```

Generic constructor (initializes the object with a time-span of 0):
```Perl
my $term = Term->new();
```

### Setters

Set the term from a number of years:
```Perl
$term->setTermInYears($term_in_years);
```

Set the term from a number of months:
```Perl
$term->setTermInMonths($term_in_months);
```

### Getters

Get the number of years for the term (not necessarily a whole number):
```Perl
my $years = $term->inYears()
```

Get the number of months for the term (not necessarily a whole number):
```Perl
my $months = $term->inMonths()
```

Get the number of whole years in the term (without any fractional parts):
```Perl
my $whole_years = $term->wholeYears();
```

Get the fractional part of the years in the term (will be less than 1):
```Perl
my $remainder_years = $term->remainderYears();
```

Get the fractional part of the years in the term, but as months (will be less than 12):
```Perl
my $remainder_months = $term->remainderMonths();
```

Get a nice string representation:
```Perl
print $term->toString();
```
Examples: `2 years 4.21 months`, `3 years`, `1 year 11.4 months`

## LoanInfo.pm

This module describes an object that holds info associated with a loan.
It also houses the calculations for the various pieces.

### Usage

Make sure that the LoanInfo.pm and Term.pm files are in your Perl path, then:
```Perl
use LoanInfo;
use Term;
```

### Constructor

It is recommended that you use the LoanInfoBuilder object for creation of a LoanInfo object.
But you don't have to if you don't want to. I'm not your father.

```Perl
my $loan_info = LoanInfo->new(\%params);
```
There are four keys looked for in the `%params` hash: `principal`, `monthly_payment`, `term`, and `rate`.
None of them are required.
The `term` key should be a `Term` object if supplied. The rest are expected to be numbers.

When constructed with exactly three of the four parameters defined, the fourth will be calculated automatically.
Otherwise, calculations of the derivative values will be attempted.
This automatica calculation is not done at any other time.
For example, if you create an empty LoanInfo object, then use the setters to define three values,
you must manully call the calculation method for the fourth.

No checks are made to ensure that all four parameters make sense.
For example, it's possible to set the principal at `1000000`, the rate at `0.0001`,
the monthly payment at `5.00` and the term at `1.5`.

### Calculations

All of these calculation methods will update the LoanInfo object with the newly calculated value, as well as return it.
If you just want the value, but don't need anything recalculated, use the appropriate getter instead.

Calculate the principal from the rate, term, and monthly payment:
```Perl
$loan_info->calculatePrincipal();
```

Calculate the monthly payment from the principal, rate, and term:
```Perl
$loan_info->calculateMonthlyPayment();
```

Calculate the term from the principal, rate, and monthly payment:
```Perl
$loan_info->calculateTerm();
```

Calculate the rate from the principal, monthly payment, and term:
```Perl
$loan_info->calculateRate();
```

Calculate the single missing parameter:
```Perl
$loan_info->calculateMissing();
```
If none are missing, or there are two or more missing, only `updateCosts` will be run.
But if there's exactly one parameter missing, the appropriate calculation method will be run.

The return value of `calculateMissing` is the LoanInfo object.

Update the other values in the loan info, e.g. total interest:
```Perl
$loan_info->updateCosts();
```
This is automatically called whenever one of the other calculation methods is called.
Odds are, you don't need to use it.

The return value of `updateCosts` is the LoanInfo object.

### Getters and Setters

This object uses the same method for a getter and setter for each parameter.
These methods will always return the associated parameter.
And if provided with a defined argument, the paremter will be set prior to being returned.

Principal:
```Perl
my $principal = $loan_info->principal();
$loan_info->principal($new_principal);
```

Monthly Payment:
```Perl
my $monthly_payment = $loan_info->monthlyPayment();
$loan_info->monthlyPayment($new_monthly_payment);
```

Rate:
```Perl
my $rate = $loan_info->rate();
$loan_info->rate($new_rate);
```

Term (Gets and sets the actual Term object):
```Perl
my $term = $loan_info->term();
$loan_info->term($new_term);
```

Total Paid (getter only):
```Perl
my $total_paid = $loan_info->totalPaid();
```

Total Interest (getter only):
```Perl
my $total_interest = $loan_info->totalInterest();
```

Average Interest Paid Per Month (getter only):
```Perl
my $ave_interest_per_payment = $loan_info->averageInterestPerPayment();
```

Get a multi-line report with the data in this LoanInfo object:
```Perl
print $loan_info->toString();
```
See the loancalc.pl section for examples of this output.

## LoanInfoBuilder.pm

This module describes an object that will hopefully make it easier to create a LoanInfo object.

### Usage

Make sure that the LoanInfoBuilder.pm, LoanInfo.pm, and Term.pm files are in your Perl path, then:
```Perl
use LoanInfoBuilder;
use LoanInfo;
use Term;
```

### Constructor

Create a new builder.
```Perl
my $builder = LoanInfoBuilder->new();
```

Have the builder create a LoanInfo object from what it's got:
```Perl
my $loan_info = $builder->build();
```

### Setters

All of these setters return the builder.
That allows you to chain the setters if desired.

Setting the same parameter a second time will simply overwrite the previously set value.

Principal:
```Perl
$builder->withPrincipal($principal);
```

Monthly Payment:
```Perl
$builder->withMonthlyPayment($monthly_payment);
```

Rate:
```Perl
$builder->withRate($rate);
```

Term (providing the value in years):
```Perl
$builder->withTermInYears($term_in_years);
```

Term (providing the value in months):
```Perl
$builder->withTermInMonths($term_in_months);
```

Term (providing a Term object):
```Perl
$builder->withTerm(Term->newMonths($term_in_months));
```


