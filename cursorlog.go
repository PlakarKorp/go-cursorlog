package gocursorlog

import (
	"io"
	"os"
	"path/filepath"
)

type CursorLog struct {
	state *cursorLogState
}

func NewCursorLog(path string) (*CursorLog, error) {
	state, err := LoadOrCreate(path)
	if err != nil {
		return nil, err
	}

	return &CursorLog{
		state: state,
	}, nil
}

func (cl *CursorLog) Close() error {
	return cl.state.Save()
}

func (cl *CursorLog) Open(path string) (io.ReadSeekCloser, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	size := fi.Size()
	start := cl.state.getCursor(abs)
	if start > size {
		start = size // file truncated; start at end
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, err
	}

	return &Tailer{
		f:       f,
		path:    abs,
		state:   cl.state,
		current: start,
	}, nil
}

func (cl *CursorLog) Reset(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	cl.state.resetCursor(abs)
	return nil
}

func (cl *CursorLog) ResetTo(path string, off int64) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	cl.state.resetCursorTo(abs, off)
	return nil
}
