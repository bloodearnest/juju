// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package sshstorage

import (
	"fmt"
	"io"
)

type lineWrapWriter struct {
	out    io.Writer
	remain int
	max    int
}

// NewLineWrapWriter returns an io.Writer that encloses the given
// io.Writer, wrapping lines at the the specified line length.
//
// Note: there is no special consideration for input that
// already contains newlines; this will simply add newlines
// after every "lineLength" bytes.
//
// This is currently only appropriate for wrapping base64-encoded
// data, which is why it lives here.
func newLineWrapWriter(out io.Writer, lineLength int) (io.Writer, error) {
	if lineLength <= 0 {
		return nil, fmt.Errorf("line length %d <= 0", lineLength)
	}
	return &lineWrapWriter{
		out:    out,
		remain: lineLength,
		max:    lineLength,
	}, nil
}

func (w *lineWrapWriter) Write(buf []byte) (int, error) {
	total := 0
	for len(buf) >= w.remain {
		n, err := w.out.Write(buf[0:w.remain])
		w.remain -= n
		total += n
		if err != nil {
			return total, err
		}
		if _, err := w.out.Write([]byte("\n")); err != nil {
			return total, err
		}
		w.remain = w.max
		buf = buf[n:]
	}
	n, err := w.out.Write(buf)
	w.remain -= n
	return total + n, err
}
