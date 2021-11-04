# SpicyLemon / go_fun / libs / capturer
This directory contains the spicylemon/libs/capturer package.

## Contents

* `bufferedpipe.go` - Defines a BufferedPipe, used to capture processed data.
* `capture.go` - Defines the CaptureOutput function that will run a provided function while capturing stdout and stderr.

## Example

See [demos/capturer.go](../../demos/capturer.go) for an example.

## Overview

The most commonly needed thing here is the `CaptureOutput` function.
The `BufferedPipe` can be used to customize your own version of `CaptureOutput`, if needed.

### CaptureOutput

The `CaptureOutput` function will capture both stdout and stderr as well as the combined stdout/stderr output.
It does not gobble output, it just captures a copy of it and allows it to continue on to its normal destination.
The ordering of the combined stdout/stderr will be the same as it is in your terminal (or environment, or whatever).
Be aware, though, that sometimes an environment doesn't always combine stdout and stderr in the expected ordering.

Example usage:
```golang
runner := func() {
    // Code for doing the things that will generate output that you want to capture.
}
output, err := capturer.CaptureOutput(runner)
if err != nil {
    panic(err) // or whatever
}
fmt.Printf("Captured stdout:\n%s\n", output.Stdout)
```

### BufferedPipe

A `BufferedPipe` contains a matched reader/writer pair of files and a buffer to copy what goes through it.
The power of a `BufferedPipe` comes through it's ability to replicate the data going through it.
The data can be replicated to anything that implements the `io.Writer` interface (e.g. `os.File` or `log.Writer()`).
Data is only captured when written to the `BufferedPipe.Writer`; writing is not captured when done directly to one of the replicated writers.

A `BufferedPipe` can be created using either `StartNewBufferedPipe` or `NewBufferedPipe`. Using `NewBufferedPipe` allows further setup but requires later calling `Start()` in order to initiate buffering.

Every `BufferedPipe` must be closed when done:
```golang
bpipe, err := StartNewBufferedPipe("random")
if err != nil {
    panic(err) // or whatever
}
defer bpipe.Close()
```

If the `BufferedPipe` is created using `NewBufferedPipe` then you must call `Start()` on it before it will capture anything.
```golang
bpipe, _ :=  NewBufferedPipe("random")
defer bpipe.Close()
bpipe.Start()
```

If the `BufferedPipe` has not yet been started, you can still add new replcation writers using `AddReplicationTo`.
```golang
bpipe, _ := NewBufferedPipe("stdout", os.Stdout)
defer bpipe.Close()
f, _ := os.OpenFile("stdout.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
bpipe.AddReplicationTo(f)
bpipe.Start()
```

If you have all of your writers when creating the `BufferedPipe`, you can use `StartNewBufferedPipe` to both create the new `BufferedPipe` and call `Start()` on it.
```golang
bpipe, _ := StartNewBufferedPipe("stdout", os.Stdout)
```
is equivalent to
```golang
bpipe, _ := NewBufferedPipe("stdout", os.Stdout)
bpipe.Start()
```

Once you're ready to stop capturing, use `Collect()` to get the captured content:
```golang
bpipe, _ := StartNewBufferedPipe("stdout", os.Stdout)
defer bpipe.Close()
origStdout := os.Stdout
defer func() {
    os.Stdout = origStdout
}
os.Stdout = bpipe.Writer
// do stuff
// ...
stdout := string(bpipe.Collect())
```

