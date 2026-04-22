package lsm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type LSMEngine struct {
	memtable       *Memtable
	wal            *os.File
	sstables       []*SSTable
	memtableSize   int
	flushThreshold int
	maxSSTables    int
	dataDir        string
	mu             sync.RWMutex
}

func NewLSMEngine(dir string) (*LSMEngine, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	walPath := filepath.Join(dir, "wal.log")
	wal, err := os.OpenFile(walPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	memtable := NewMemtable()

	if err := replayWAL(wal, memtable); err != nil {
		return nil, err
	}

	engine := &LSMEngine{
		memtable:       memtable,
		wal:            wal,
		sstables:       make([]*SSTable, 0),
		flushThreshold: 4096,
		maxSSTables:    10,
		dataDir:        dir,
	}

	if err := engine.loadSSTables(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *LSMEngine) loadSSTables() error {
	entries, err := os.ReadDir(e.dataDir)
	if err != nil {
		return err
	}

	var files []string

	for _, entry := range entries {
		name := entry.Name()

		if filepath.Ext(name) == ".db" && len(name) >= 8 && name[:8] == "sstable-" {
			files = append(files, filepath.Join(e.dataDir, name))
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, path := range files {

		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		sstable := &SSTable{
			file:  file,
			index: []indexEntry{},
		}

		// 3. rebuild sparse index by scanning file
		if err := rebuildIndex(sstable); err != nil {
			file.Close()
			return err
		}

		e.sstables = append(e.sstables, sstable)
	}

	return nil
}

func rebuildIndex(s *SSTable) error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}

	const sparseRate = 10
	var offset int64
	var i int

	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := splitEntry(line)
		if len(parts) < 3 {
			continue
		}

		key := parts[0]

		if s.minKey == "" || key < s.minKey {
			s.minKey = key
		}
		if key > s.maxKey {
			s.maxKey = key
		}

		if i%sparseRate == 0 {
			s.index = append(s.index, indexEntry{key: key, offset: offset})
		}

		offset += int64(len(scanner.Bytes())) + 1 // +1 for newline
		i++
	}

	return scanner.Err()
}

func splitEntry(line string) []string {
	return strings.Split(line, " ")
}

func (l *LSMEngine) Set(key, value string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := appendWAL(l.wal, "SET", key, value); err != nil {
		return err
	}

	l.memtable.Set(key, value)

	l.memtableSize += len(key) + len(value)

	return l.maybeFlush()
}

func (l *LSMEngine) Get(key string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if val, ok := l.memtable.Get(key); ok {
		if val == "" {
			return "", fmt.Errorf("key not found") // tombstone
		}
		return val, nil
	}

	for i := range l.sstables {
		val, err := l.sstables[i].Get(key)
		if err == nil {
			if val == "" {
				return "", fmt.Errorf("key not found")
			}
			return val, nil
		}
	}

	return "", fmt.Errorf("key not found")
}

func (l *LSMEngine) Delete(key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := appendWAL(l.wal, "DELETE", key, ""); err != nil {
		return err
	}

	l.memtable.Delete(key)
	l.memtableSize += len(key)

	return l.maybeFlush()
}

func (l *LSMEngine) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.wal.Close()
	for _, sst := range l.sstables {
		sst.file.Close()
	}
	return nil
}

func (l *LSMEngine) maybeFlush() error {
	if l.memtableSize < l.flushThreshold {
		return nil
	}

	sst, err := flushToSSTable(l.memtable, l.dataDir)
	if err != nil {
		return err
	}

	l.sstables = append([]*SSTable{sst}, l.sstables...)

	l.memtable = NewMemtable()
	l.memtableSize = 0

	if err := l.wal.Truncate(0); err != nil {
		return err
	}
	if _, err := l.wal.Seek(0, 0); err != nil {
		return err
	}

	return nil
}
