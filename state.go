package gocursorlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type cursorLogState struct {
	path  string `json:"-"`
	dirty bool   `json:"-"`

	cursorsMutex sync.Mutex
	Cursors      map[string]int64 `json:"cursors"`
}

func LoadOrCreate(path string) (*cursorLogState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Ensure parent dir exists for future saves.
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return nil, fmt.Errorf("create state dir: %w", err)
			}
			return &cursorLogState{
				Cursors: make(map[string]int64),
				dirty:   false,
				path:    path,
			}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	} else if len(b) == 0 {
		return &cursorLogState{
			Cursors: make(map[string]int64),
			dirty:   false,
			path:    path,
		}, nil
	}

	var st cursorLogState
	st.path = path
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	if st.Cursors == nil {
		st.Cursors = make(map[string]int64)
	}
	st.dirty = false
	return &st, nil
}

func (s *cursorLogState) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(struct {
		Cursors map[string]int64 `json:"cursors"`
	}{Cursors: s.Cursors}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	} else if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}

	s.dirty = false
	return nil
}

func (s *cursorLogState) setCursor(absPath string, off int64) {
	s.cursorsMutex.Lock()
	defer s.cursorsMutex.Unlock()
	if cur, ok := s.Cursors[absPath]; !ok || cur != off {
		s.Cursors[absPath] = off
		s.dirty = true
	}
}

func (s *cursorLogState) getCursor(absPath string) int64 {
	s.cursorsMutex.Lock()
	defer s.cursorsMutex.Unlock()
	return s.Cursors[absPath]
}

func (s *cursorLogState) resetCursor(absPath string) {
	s.cursorsMutex.Lock()
	defer s.cursorsMutex.Unlock()
	if _, ok := s.Cursors[absPath]; ok {
		delete(s.Cursors, absPath)
		s.dirty = true
	}
}

func (s *cursorLogState) resetCursorTo(absPath string, off int64) {
	s.cursorsMutex.Lock()
	defer s.cursorsMutex.Unlock()
	if off < 0 {
		off = 0
	}
	if cur, ok := s.Cursors[absPath]; !ok || cur != off {
		s.Cursors[absPath] = off
		s.dirty = true
	}
}
