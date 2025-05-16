# big-sum

Add any amount of numbers of any size.

[LICENSE](LICENSE)



## Installation

To build the `big-sum` executable, you can use either:
* `make install` - places it your standard go installation directory.
* `make build` - places it in the `build/` directory.
* `go build .` - places it in your current directory.



## Usage

```plaintext
big-sum: Add a bunch of numbers together with nearly infinite precision.

Usage: big-sum <number 1> [<number 2> ...] [--pipe|-] [--pretty|-p] [--verbose|-v]
  or : <stuff> | big-sum

The --pipe or - flag is implied if there are no arguments provided.
The --pretty or -p flag will add commas to the result.

Warning: In rare circumstances, floating point numbers may result in unwanted rounding.
```

