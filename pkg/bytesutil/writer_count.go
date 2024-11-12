package bytesutil

import (
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

var ErrWrite = errors.New("failed to write")

// WriterCounter is counter for io.Writer.
type WriterCounter struct {
	count uint64
	io.Writer
}

// NewWriterCounter function create new WriterCounter.
func NewWriterCounter(w io.Writer) *WriterCounter {
	return &WriterCounter{
		count:  0,
		Writer: w,
	}
}

func (counter *WriterCounter) Write(buf []byte) (int, error) {
	n, err := counter.Writer.Write(buf)

	// Write() should always return a non-negative `n`.
	// But since `n` is a signed integer, some custom
	// implementation of an io.Writer may return negative
	// values.
	//
	// Excluding such invalid values from counting,
	// thus `if n >= 0`:
	if n >= 0 {
		atomic.AddUint64(&counter.count, uint64(n))
	}

	if err != nil {
		return n, fmt.Errorf("%w:%w", ErrWrite, err)
	}

	return n, nil
}

// Count function return counted bytes.
func (counter *WriterCounter) Count() uint64 {
	return atomic.LoadUint64(&counter.count)
}
