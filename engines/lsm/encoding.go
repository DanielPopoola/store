package lsm

import "encoding/json"

func encodeLine(key, value string, deleted bool) (string, error) {
	data, err := json.Marshal(sstableEntry{Key: key, Value: value, Deleted: deleted})
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func decodeLine(line string) (key, value string, deleted bool, err error) {
	var entry sstableEntry
	err = json.Unmarshal([]byte(line), &entry)
	return entry.Key, entry.Value, entry.Deleted, err
}
