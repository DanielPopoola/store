package lsm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// appennWAL writes an operation to the WAL file
func appendWAL(file *os.File, op, key, value string) error {
	var line string

	switch op {
	case "SET":
		line = fmt.Sprintf("SET %s %s\n", key, value)
	case "DELETE":
		line = fmt.Sprintf("DELETE %s\n", key)
	default:
		return fmt.Errorf("unkown operation: %s", op)
	}

	if _, err := file.WriteString(line); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}

// replayWAL reads the WAL file and rebuilds the memtable
func replayWAL(file *os.File, memtable *Memtable) error {
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	reader := bufio.NewReader(file)

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

		op := parts[0]
		key := parts[1]

		switch op {
		case "SET":
			if len(parts) < 3 {
				continue // malformed SET
			}
			value := parts[2]
			memtable.Set(key, value)

		case "DELETE":
			memtable.Delete(key)

		default:
			continue
		}
	}

	return nil
}
