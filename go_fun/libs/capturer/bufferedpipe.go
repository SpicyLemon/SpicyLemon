package capturer

import (
	"bytes"
	"io"
	"os"
)

var _ io.Reader = BufferedPipe{}
var _ io.Writer = BufferedPipe{}

// BufferedPipe contains a connected read/write pair of files (a pipe),
// and a buffer of what goes through it that is populated in the background.
type BufferedPipe struct {
	// Name is a string to help humans identify this BufferedPipe.
	Name string
	// Reader is the reader end of the pipe.
	Reader *os.File
	// Writer is the writer end of the pipe.
	Writer *os.File
	// BufferReader is the reader used by this BufferedPipe while buffering.
	// If this BufferedPipe is not replicating to anything, it will be the same as the Reader.
	// Otherwise, it will be a reader encapsulating all desired replication.
	BufferReader io.Reader
	// Error is the last error encountered by this BufferedPipe.
	Error error

	// buffer is the channel used to communicate buffer contents.
	buffer chan []byte
	// stated is true if this BufferedPipe has been started.
	started bool
}

// NewBufferedPipe creates a new BufferedPipe with the given name.
// Files must be closed once you are done with them (e.g. with .Close()).
// Once ready, buffering must be started using .Start(). See also StartNewBufferedPipe.
func NewBufferedPipe(name string, replicateTo ...io.Writer) (BufferedPipe, error) {
	p := BufferedPipe{Name: name}
	p.Reader, p.Writer, p.Error = os.Pipe()
	if p.Error != nil {
		return p, p.Error
	}
	p.BufferReader = p.Reader
	p.AddReplicationTo(replicateTo...)
	return p, nil
}

// StartNewBufferedPipe creates a new BufferedPipe and starts it.
//
// This is functionally equivalent to:
//    p, _ := NewBufferedPipe(name, replicateTo...)
//    p.Start()
func StartNewBufferedPipe(name string, replicateTo ...io.Writer) (BufferedPipe, error) {
	p, err := NewBufferedPipe(name, replicateTo...)
	if err != nil {
		return p, err
	}
	p.Start()
	return p, nil
}

// AddReplicationTo adds replication of this buffered pipe to the provided writers.
//
// Panics if this BufferedPipe is already started.
func (p *BufferedPipe) AddReplicationTo(writers ...io.Writer) {
	p.panicIfStarted("cannot add further replication")
	for _, writer := range writers {
		p.BufferReader = io.TeeReader(p.BufferReader, writer)
	}
}

// Start initiates buffering in a background process.
//
// Panics if this BufferedPipe is already started.
func (p *BufferedPipe) Start() {
	p.panicIfStarted("cannot restart")
	p.buffer = make(chan []byte)
	go func() {
		var b bytes.Buffer
		if _, p.Error = io.Copy(&b, p.BufferReader); p.Error != nil {
			b.WriteString("buffer error: " + p.Error.Error())
		}
		p.buffer <- b.Bytes()
	}()
	p.started = true
}

// IsStarted returns true if this BufferedPipe has already been started.
func (p *BufferedPipe) IsStarted() bool {
	return p.started
}

// IsBuffering returns true if this BufferedPipe has started buffering and has not yet been collected.
func (p *BufferedPipe) IsBuffering() bool {
	return p.buffer != nil
}

// Collect closes this pipe's writer then blocks, returning with the final buffer contents once available.
// If Collect() has previously been called on this BufferedPipe, an empty byte slice is returned.
//
// Panics if this BufferedPipe has not been started.
func (p *BufferedPipe) Collect() []byte {
	if !p.started {
		panic("buffered pipe " + p.Name + " has not been started: cannot collect")
	}
	_ = p.Writer.Close()
	if p.buffer == nil {
		return []byte{}
	}
	rv := <-p.buffer
	p.buffer = nil
	return rv
}

// Read implements the io.Reader interface on this BufferedPipe.
func (p BufferedPipe) Read(bz []byte) (n int, err error) {
	return p.Reader.Read(bz)
}

// Write implements the io.Writer interface on this BufferedPipe.
func (p BufferedPipe) Write(bz []byte) (n int, err error) {
	return p.Writer.Write(bz)
}

// Close makes sure the files in this BufferedPipe are closed.
func (p *BufferedPipe) Close() {
	_ = p.Reader.Close()
	_ = p.Writer.Close()
}

// panicIfStarted panics if this BufferedPipe has been started.
func (p *BufferedPipe) panicIfStarted(msg string) {
	if p.started {
		panic("buffered pipe " + p.Name + " already started: " + msg)
	}
}
