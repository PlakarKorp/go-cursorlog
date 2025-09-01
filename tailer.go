package gocursorlog

import (
	"io"
	"os"
)

// Tailer implements io.ReadCloser. It starts reading at the saved cursor and
// advances the in-memory cursor as bytes are read. The cursor is committed on Close().
type Tailer struct {
	f       *os.File
	path    string
	state   *cursorLogState
	current int64
	closed  bool
}

var _ io.ReadCloser = (*Tailer)(nil)

func (t *Tailer) Read(p []byte) (int, error) {
	if t.closed {
		return 0, io.EOF
	}
	n, err := t.f.Read(p)
	t.current += int64(n)
	return n, err
}

func (t *Tailer) Seek(offset int64, whence int) (int64, error) {
	if t.closed {
		return 0, io.EOF
	}
	newOff, err := t.f.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	t.current = newOff
	return newOff, nil
}

func (t *Tailer) Close() error {
	if t.closed {
		return nil
	}
	t.closed = true
	t.state.setCursor(t.path, t.current)
	return t.f.Close()
}
