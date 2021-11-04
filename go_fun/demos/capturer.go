package main

import (
	"fmt"
	"io"
	"os"
	"spicylemon/libs/capturer"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println(" --- Before capture start --- ")
	output, err := capturer.CaptureOutput(doCrazyOutput)
	if err != nil {
		panic(err)
	}
	fmt.Println(" --- After capture end --- ")
	fmt.Printf("\n\n")
	fmt.Printf("Captured stdout:\n%s\n\n", indent(output.Stdout))
	fmt.Printf("Captured stderr:\n%s\n\n", indent(output.Stderr))
	fmt.Printf("Captured combined:\n%s\n\n", indent(output.Combined))
}

func indent(str string) string {
	return prefixLines(str, "  ")
}

func prefixLines(lines, pre string) string {
	return pre + strings.ReplaceAll(lines, "\n", "\n"+pre)
}

func doCrazyOutput() {
	stdoutFunc("This should be stdout.")()
	stderrFunc("This should be stderr.")()
	var wg sync.WaitGroup
	wg.Add(2)
	go delayStdout(&wg, 1, "Not much wait on this one.")
	go delayStderr(&wg, 1500, "This stderr statement took a long time to come through.")
	for i := 500; i >= 50; i -= 50 {
		stdoutFunc("Delaying stdout message %d ms and stderr message %d ms", i, i+25)()
		wg.Add(2)
		go delayStdout(&wg, i, "stdout delayed %d ms", i)
		go delayStderr(&wg, i+25, "stderr delayed %d ms", i+25)
	}
	stdoutFunc("Waiting for routines to finish.")()
	wg.Wait()
	stdoutFunc("All routines done.")()
	stderrFunc("Some final stderr output.")()
	fmt.Println("Just a normal fmt.Println")
}

func delayStdout(wg *sync.WaitGroup, ms int, format string, a ...interface{}) {
	delayDo(wg, ms, stdoutFunc(format, a...))
}

func delayStderr(wg *sync.WaitGroup, ms int, format string, a ...interface{}) {
	delayDo(wg, ms, stderrFunc(format, a...))
}

func delayDo(wg *sync.WaitGroup, ms int, runner func()) {
	defer wg.Done()
	time.Sleep(time.Millisecond * time.Duration(ms))
	runner()
}

func stdoutFunc(format string, a ...interface{}) func() {
	return fprintfFunc(os.Stdout, format, a...)
}

func stderrFunc(format string, a ...interface{}) func() {
	return fprintfFunc(os.Stderr, format, a...)
}

func fprintfFunc(w io.Writer, format string, a ...interface{}) func() {
	return func() {
		fmt.Fprintf(w, format+"\n", a...)
	}
}
