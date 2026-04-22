package lsm

import (
	"sort"
)

type MemtableEntry struct {
	Key     string
	Value   string
	Deleted bool
}

type Memtable struct {
	entries []MemtableEntry
}

func NewMemtable() *Memtable {
	return &Memtable{
		entries: []MemtableEntry{},
	}
}

// internal helper: binary search
func (m *Memtable) findIndex(key string) (int, bool) {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].Key >= key
	})

	if i < len(m.entries) && m.entries[i].Key == key {
		return i, true
	}

	return i, false
}

func (m *Memtable) Set(key, value string) {
	i, found := m.findIndex(key)

	if found {
		// update existing entry
		m.entries[i].Value = value
		m.entries[i].Deleted = false // revive if previously deleted
		return
	}

	// insert new entry at correct positon
	entry := MemtableEntry{
		Key:     key,
		Value:   value,
		Deleted: false,
	}

	m.entries = append(m.entries, MemtableEntry{})
	copy(m.entries[i+1:], m.entries[i:]) // shift right
	m.entries[i] = entry
}

func (m *Memtable) Get(key string) (string, bool) {
	i, found := m.findIndex(key)

	if !found {
		return "", false
	}

	if m.entries[i].Deleted {
		return "", false
	}

	return m.entries[i].Value, true
}

func (m *Memtable) Delete(key string) {
	i, found := m.findIndex(key)

	if found {
		m.entries[i].Deleted = true
		m.entries[i].Value = ""
		return
	}

	// Insert tombstone if key does not exist
	entry := MemtableEntry{
		Key:     key,
		Value:   "",
		Deleted: true,
	}

	m.entries = append(m.entries, MemtableEntry{})
	copy(m.entries[i+1:], m.entries[i:])
	m.entries[i] = entry
}
