# date-math

Do calculations with datetimes and durations.

[LICENSE](LICENSE)



## Installation

To build the `date-math` executable, you can use either:
* `make install` - places it your standard go installation directory.
* `make build` - places it in the `build/` directory.
* `go build .` - places it in your current directory.



## Usage

```plaintext
date-math: Do calculations with datetimes and durations.

Usage: date-math (<formula>|formats) [flags]

A <formula> has the format <value> <op> (<value>|<formula>)

A <value> can either be a <date>, <epoch>, <dur>, or <num>.
  <time> A datetime string. Multiple formats are supported.
         To see all possible formats, execute: date-math formats
         Datetimes that do not have a time zone are assumed to be local which is
         controllable by setting the TZ environment variable.
  <epoch> A possibly signed number with optional fractional seconds.
          An <epoch> is treated as a <time> for the purposes of the calculations.
  <dur> A possibly signed sequence of decimal numbers, each with optional fraction
        and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
        The "d" and "w" time units are non-standard and represent days and weeks.
        It's assumed that 1w = 7d and 1d = 24h = 1440m = 86400s, even though that
        isn't always the case, e.g. time changes and leap seconds.
  <num> A possibly signed whole number.

A whole number might be either an <epoch> or <num>. By default, a whole number
greater than 1,000,000 or less than -1,000,000 is treated as an <epoch>. A whole
number between -1,000,000 and -1,000,000 (inclusive) is treated as a <num>.
To force a whole number to be an <epoch>, prepend it with 'e', e.g. 'e1000000'.
To force a whole number to be a <num>, prepend it with 'n', e.g. 'n1000001'.

The <op> can be + - x or /. Only the following operations are defined:
  <time> - <time> => <dur>   e.g. 2020-01-09 4:30:00 - 2020-01-09 3:29:28 => 1h2s
                               or 2020-01-09 3:29:28 - 2020-01-09 4:30:00 => -1h2s
  <time> + <dur>  => <time>  e.g. 2020-01-09 4:30:00 + 1h2s => 2020-01-09 5:30:02
  <dur>  + <time> => <time>  e.g. 1h2s + 2020-01-09 4:30:00 => 2020-01-09 5:30:02
  <time> - <dur>  => <time>  e.g. 2020-01-09 4:30:00 - 1h2s => 2020-01-09 3:29:28
  <dur>  + <dur>  => <dur>   e.g. 1h2s + 3m5s => 1h3m7s (communicative)
  <dur>  - <dur>  => <dur>   e.g. 1h2s - 3m5s => 56m57s  or  3m5s - 1h2s => -56m57s
  <dur>  / <dur>  => <num>   e.g. 2h / 40m => 3
  <dur>  x <num>  => <dur>   e.g. 40m x 3 => 2h
  <num>  x <dur>  => <dur>   e.g. 5 x 40m => 3h20m
  <dur>  / <num>  => <dur>   e.g. 2h / 3 => 40m
  <num>  + <num>  => <num>   e.g. 5 + 3 => 8 (communicative)
  <num>  - <num>  => <num>   e.g. 5 - 3 => 2  or  3 - 5 => -2
  <num>  x <num>  => <num>   e.g. 5 x 3 => 15 (communicative)
  <num>  / <num>  => <num>   e.g. 6 / 3 => 2  or  5 / 3 => 1

Notes:
1. Those examples might have slightly different output, but same values.
2. Division is done using integers which will truncate the result.
   A <dur> is handled as an integer amount of nanoseconds. So <dur> / <num>
   will be truncated to the nearest nanosecond.
3. Multiplication is done using x instead of * because shells will expand *,
   and I didn't want to have to always remember to escape it.


The <formula> is calculated from left to right and can have multiple operations.
E.g. 2020-01-09 4:30:00 + 1h2s - 2020-01-02 11:30:18
   = 2020-01-09 5:30:02 - 2020-01-02 11:30:18
   = 6d17h59m44s

If "formats" is provided the list of named datetime format strings is printed.
These are the valid names to provide with the --output flag.

There are a few flags that can also be provided:
  --output-name|-o <name>
        Use the format with the provided <name> to convert a final <time> value
        into the result. Does nothing if the final result isn't a <time>.
        See: date-math formats
  --output-format|-f <format>
        Use the provided <format> to convert a final <time> value into the
        result. Does nothing if the final result isn't a <time>.
        See: https://pkg.go.dev/time#pkg-constants
  --input-name|-i <name>
        Use the format with the provided <name> to parse any provided <time>
        values. When this option is used, none of the other formats will be
        considered for parsing. If the final result is a <time> it will also
        have this format, unless either --output-name or --output-format are used.
        See: date-math formats
  --input-format|-g <format>
        Use the provided input to parse any provided <time> values. When this
        option is used, none of the other formats will be available. If the final
        result is a <time> it will also have this format, unless either
        --output-name or --output-format are used.
  --formats
        Same as providing just "formats"; outputs info on all named formats.
  --pipe|-p
        Read formula args from stdin and run the calculation for each line.
        Each line is inserted in place of the --pipe or -p flag among any other
        formula args that are provided. This allows for piping in values, ops,
        partial formulas, or full formulas. This flag can be omitted if there
        are no other formula args to provide.
  --verbose|-v
        Print debugging information to stderr.
        Can also be enabled by setting the VERBOSE env var.
  --help|-h
        Output this message.
```



## Formats

There are several named formats, some of which are automatically available to parse a datetime.

You can get this list by executing `date-math formats`

```plaintext
Formats (22): * = possible input format
   1: *         ANSIC = "Mon Jan _2 15:04:05 2006"
   2:        DateOnly = "2006-01-02"
   3: *      DateTime = "2006-01-02 15:04:05"
   4: *  DateTimeZone = "2006-01-02 15:04:05.999999999 -0700"
   5: * DateTimeZone2 = "2006-01-02 15:04:05.999999999Z0700"
   6:         Default = "2006-01-02 15:04:05.999999999 -0700 MST"
   7:         Kitchen = "3:04PM"
   8:          Layout = "01/02 03:04:05PM '06 -0700"
   9: *       RFC1123 = "Mon, 02 Jan 2006 15:04:05 MST"
  10: *      RFC1123Z = "Mon, 02 Jan 2006 15:04:05 -0700"
  11:         RFC3339 = "2006-01-02T15:04:05Z07:00"
  12: *   RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
  13:          RFC822 = "02 Jan 06 15:04 MST"
  14:         RFC822Z = "02 Jan 06 15:04 -0700"
  15: *        RFC850 = "Monday, 02-Jan-06 15:04:05 MST"
  16: *      RubyDate = "Mon Jan 02 15:04:05 -0700 2006"
  17:           Stamp = "Jan _2 15:04:05"
  18:      StampMicro = "Jan _2 15:04:05.000000"
  19:      StampMilli = "Jan _2 15:04:05.000"
  20:       StampNano = "Jan _2 15:04:05.000000000"
  21:        TimeOnly = "15:04:05"
  22: *      UnixDate = "Mon Jan _2 15:04:05 MST 2006"
```

If the final result is a datetime, the first match in this list dictates what format to use:

1. If `--output-name`, `-o`, `--output-format`, or `-f` are provided, that is the format that is used.
2. If `--input-name`, `-i`, `--input-format`, or `-g` are provided, that is the format that is used.
3. If exactly one format was used to parse the datetimes, and that format has all of a full date, full time, and time zone, that is the format that is used.
4. The default format is used: `"2006-01-02 15:04:05.999999999 -0700 MST"`

For formatting details, see: https://pkg.go.dev/time#pkg-constants



## Examples

The actual time zones you might end up with depend on your own system.

### time - time

```console
$ date-math 2020-01-09 4:30:00 - 2020-01-09 3:29:28
1h0m32s
```

```console
$ date-math 2020-01-09 3:29:28 - 2020-01-09 4:30:00
-1h0m32s
```

### time + dur

```console
$ date-math 2020-01-09 4:30:00 + 1h2s
2020-01-09 05:30:02 -0700 MST
```

### dur + time

```console
$ date-math 1h2s + 2020-01-09 4:30:00
2020-01-09 05:30:02 -0700 MST
```

### time - dur

```console
$ date-math 2020-01-09 4:30:00 - 1h2s
2020-01-09 03:29:58 -0700 MST
```

### dur + dur

```console
$ date-math 1h2s + 3m5s
1h3m7s
```

```console
$ date-math 3m5s + 1h2s
1h3m7s
```

### dur - dur

```console
$ date-math 1h2s - 3m5s
56m57s
```

```console
$ date-math 3m5s - 1h2s
-56m57s
```

### dur / dur

```console
$ date-math 2h / 40m
3
```

### dur x num

```console
$ date-math 40m x 3
2h
```

### num x dur

```console
$ date-math 5 x 40m
3h20m
```

### dur / num

```console
$ date-math 2h / 3
40m
```

### num + num

```console
$ date-math 5 + 3
8
```

```console
$ date-math 3 + 5
8
```

### num - num

```console
$ date-math 5 - 3
2
```

```console
$ date-math  3 - 5
-2
```

### num x num

```console
$ date-math 5 x 3
15
```

```console
$ date-math 3 x 5
15
```

### num / num

```console
$ date-math 6 / 3
2
```

```console
$ date-math 5 / 3
1
```
