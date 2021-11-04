package capturer

import (
	"os"
)

// CapturedOutput contains a breakdown of the various captured outputs.
type CapturedOutput struct {
	Stdout   string
	Stderr   string
	Combined string
}

// CaptureOutput capture all the things written to stdout and stderr during some code execution.
func CaptureOutput(runner func()) (CapturedOutput, error) {
	// Create a buffered pipe for the combined stdout stderr.
	combinedPipe, err := StartNewBufferedPipe("combined")
	if err != nil {
		return CapturedOutput{}, err
	}
	defer combinedPipe.Close()

	// Create a buffered pipe for stdout and replicate it to the current stdout and the combined buffered pipe.
	stdoutPipe, err := StartNewBufferedPipe("stdout", os.Stdout, combinedPipe)
	if err != nil {
		return CapturedOutput{}, err
	}
	defer stdoutPipe.Close()

	// Create a buffered pipe for stderr and replicate it to the current stderr and the combined buffered pipe.
	stderrPipe, err := StartNewBufferedPipe("stderr", os.Stderr, combinedPipe)
	if err != nil {
		return CapturedOutput{}, err
	}
	defer stderrPipe.Close()

	// Swap out the existing stdout and stderr with our buffered pipes that capture and replicate.
	origStdOut := os.Stdout
	origStdErr := os.Stderr
	defer func() {
		os.Stdout = origStdOut
		os.Stderr = origStdErr
	}()
	os.Stdout = stdoutPipe.Writer
	os.Stderr = stderrPipe.Writer

	// Run the stuff we want to capture.
	runner()

	// Collect all the output.
	// Since writes go to the stdout buffered pipe, then stdout, then the combined buffered pipe (and same with stderr),
	// this is ordered to collect the combined buffered pipe last.
	// That way, there isn't a writer open that is replicating to a closed combined pipe.
	stdoutBz := stdoutPipe.Collect()
	stderrBz := stderrPipe.Collect()
	combinedBz := combinedPipe.Collect()

	return CapturedOutput{
		Stdout:   string(stdoutBz),
		Stderr:   string(stderrBz),
		Combined: string(combinedBz),
	}, nil
}
