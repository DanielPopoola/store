package engines

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type FileEngine struct {
	file    *os.File
	offsets map[string]int64
	mu      sync.RWMutex
}

func NewFileEngine(path string) (*FileEngine, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	engine := &FileEngine{
		file:    file,
		offsets: make(map[string]int64),
	}

	if err := engine.replay(); err != nil {
		return nil, err
	}

	return engine, nil
}

// rebuild index from file
func (f *FileEngine) replay() error {
	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	var offset int64
	reader := bufio.NewReader(f.file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		line = strings.TrimSuffix(line, "\n")
		parts := strings.SplitN(line, " ", 3)

		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "SET":
			if len(parts) == 3 {
				f.offsets[parts[1]] = offset
			}
		case "DELETE":
			delete(f.offsets, parts[1])
		}

		offset += int64(len(line)) + 1
	}

	return nil
}

func (f *FileEngine) Set(key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// move to end before writing
	offset, err := f.file.Seek(0, 2)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("SET %s %s\n", key, value)

	if _, err := f.file.WriteString(line); err != nil {
		return err
	}

	// force write to disk
	if err := f.file.Sync(); err != nil {
		return err
	}

	f.offsets[key] = offset
	return nil
}

func (f *FileEngine) Get(key string) (string, error) {
	f.mu.RLock()
	offset, ok := f.offsets[key]
	f.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("key not found")
	}

	// jump to exact position
	_, err := f.file.Seek(offset, 0)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(f.file)

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("corrupted record")
	}

	return parts[2], nil
}

func (f *FileEngine) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.offsets[key]; !ok {
		return fmt.Errorf("key not found")
	}

	line := fmt.Sprintf("DELETE %s\n", key)

	if _, err := f.file.WriteString(line); err != nil {
		return err
	}

	if err := f.file.Sync(); err != nil {
		return err
	}

	delete(f.offsets, key)
	return nil
}
