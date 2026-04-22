package lsm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type indexEntry struct {
	key    string
	offset int64
}

type SSTable struct {
	file   *os.File
	index  []indexEntry // sorted sparse index
	minKey string
	maxKey string
}

func flushToSSTable(memtable *Memtable, dir string) (*SSTable, error) {
	if len(memtable.entries) == 0 {
		return nil, fmt.Errorf("memtable is empty")
	}

	filename := fmt.Sprintf("sstable-%d.db", time.Now().UnixNano())
	path := filepath.Join(dir, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	const sparseRate = 10

	index := make([]indexEntry, 0)
	var offset int64 = 0

	minKey := memtable.entries[0].Key
	maxKey := memtable.entries[len(memtable.entries)-1].Key

	// 2. write entries sequentially
	for i := range memtable.entries {
		entry := &memtable.entries[i]

		line := fmt.Sprintf("%s %s %t\n", entry.Key, entry.Value, entry.Deleted)

		n, err := file.WriteString(line)
		if err != nil {
			file.Close()
			return nil, err
		}

		// 3. sparse index every N entries
		if i%sparseRate == 0 {
			index = append(index, indexEntry{
				key:    entry.Key,
				offset: offset,
			})
		}

		offset += int64(n)
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return nil, err
	}

	return &SSTable{
		file:   file,
		index:  index,
		minKey: minKey,
		maxKey: maxKey,
	}, nil
}

func (s *SSTable) Get(key string) (string, error) {
	if key < s.minKey || key > s.maxKey {
		return "", fmt.Errorf("not in SSTable range")
	}

	offset := s.findOffset(key)

	_, err := s.file.Seek(offset, 0)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		k, v, deleted := parseLine(line)

		if k > key {
			return "", fmt.Errorf("not found")
		}

		if k == key {
			if deleted {
				return "", fmt.Errorf("key not found")
			}
			return v, nil
		}
	}

	return "", fmt.Errorf("not found")
}

func (s *SSTable) findOffset(key string) int64 {
	var result int64

	for _, entry := range s.index {
		if entry.key <= key {
			result = entry.offset
		} else {
			break
		}
	}

	return result
}

func splitLines(data string) []string {
	return strings.Split(data, "\n")
}

func parseLine(line string) (key, value string, deleted bool) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return "", "", false
	}

	key = parts[0]
	value = parts[1]
	deleted = parts[2] == "true"
	return
}
