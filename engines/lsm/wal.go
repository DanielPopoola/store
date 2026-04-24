package lsm

import (
	"bufio"
	"os"
)

// appennWAL writes an operation to the WAL file
func appendWAL(file *os.File, key, value string, deleted bool) error {
	line, err := encodeLine(key, value, deleted)
	if err != nil {
		return err
	}

	if _, err := file.WriteString(line); err != nil {
		return err
	}

	return file.Sync()
}

// replayWAL reads the WAL file and rebuilds the memtable
func replayWAL(file *os.File, memtable *Memtable) error {
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		key, value, deleted, err := decodeLine(line)
		if err != nil {
			continue
		}

		if deleted {
			memtable.Delete(key)
		} else {
			memtable.Set(key, value)
		}
	}

	return scanner.Err()
}
