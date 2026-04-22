# store

A lightweight key-value store written in Go, built as a companion project to reading [Designing Data-Intensive Applications](https://dataintensive.net/) by Martin Kleppmann.

The goal is to implement progressively more sophisticated storage backends as each chapter is covered — starting simple and adding complexity only when the theory demands it.

---

## Architecture

```
HTTP Request
    ↓
API          → parses requests, calls Store
Store        → engine-agnostic, delegates to Engine interface
Engine       → interface every backend must satisfy
    ↓
memoryEngine / fileEngine / LSMEngine
```

The `Engine` interface is the core abstraction. Every storage backend implements the same three methods:

```go
Set(key, value string) error
Get(key string) (string, error)
Delete(key string) error
Close() error
```

Swapping the engine requires changing one line in `main.go`. Nothing above it changes.

---

## API

| Method   | Endpoint        | Description          |
|----------|-----------------|----------------------|
| `POST`   | `/set/{key}`    | Store a value        |
| `GET`    | `/get/{key}`    | Retrieve a value     |
| `DELETE` | `/delete/{key}` | Delete a key         |

**Set example:**
```bash
curl -X POST http://localhost:8080/set/name \
  -H "Content-Type: application/json" \
  -d '{"value": "daniel"}'
```

**Get example:**
```bash
curl http://localhost:8080/get/name
# {"value":"daniel"}
```

**Delete example:**
```bash
curl -X DELETE http://localhost:8080/delete/name
```

---

## Storage Engines

### ✅ memoryEngine
The simplest possible engine. Stores key-value pairs in a Go map.

- Reads and writes are O(1)
- Data is lost on restart — no persistence
- No crash recovery

**When to use:** Development, testing, or when persistence doesn't matter.

---

### ✅ fileEngine (Chapter 3 — Hash Indexes)
An append-only log on disk with an in-memory hash index.

- Every write appends a line to a log file (`SET key value` / `DELETE key`)
- An in-memory map tracks `key → byte offset` in the file
- On startup, replays the log to rebuild the index
- Survives crashes — log is the source of truth

**Limitations:**
- The entire key index must fit in RAM
- No support for range queries (keys are unordered)

**When to use:** Write-heavy workloads with a bounded key set.

---

### ✅ LSMEngine — *current engine* (Chapter 3 — SSTables & LSM Trees)
A Log-Structured Merge-Tree implementation. The foundation of databases like Cassandra, RocksDB, and LevelDB.

**Write path:**
1. Write appended to WAL (Write-Ahead Log) on disk for crash safety
2. Entry inserted into memtable — a sorted in-memory structure
3. When memtable exceeds threshold, flushed to an immutable SSTable file on disk

**Read path:**
1. Check memtable first (most recent writes)
2. Search SSTables newest → oldest, using a sparse index to skip to the right offset
3. Min/max key range check skips SSTables that can't contain the key

**Crash recovery:**
- WAL replays into a fresh memtable on restart
- Existing SSTable files are reloaded and their sparse indexes rebuilt

**Key concepts implemented:**
- Sorted String Tables (SSTables)
- Sparse in-memory index per SSTable
- Write-Ahead Log (WAL)
- Tombstones for deletes
- Memtable flushing

**Still to implement:**
- Compaction — merging SSTables and discarding stale/deleted values
- Bloom filters — skip SSTables that definitely don't contain a key

---

## Project Structure

```
store/
├── main.go                  → wires engine, store, API and starts server
├── store.go                 → Store struct, delegates to Engine interface
├── api.go                   → HTTP handlers
└── engines/
    ├── engine.go            → Engine interface definition
    ├── memory.go            → memoryEngine (in-memory map)
    ├── file_engine.go       → fileEngine (append-only log + hash index)
    └── lsm/
        ├── lsm.go           → LSMEngine — main engine, wires all components
        ├── memtable.go      → sorted slice with binary search
        ├── sstable.go       → SSTable flush and read logic
        └── wal.go           → Write-Ahead Log append and replay
```

---

## Running

```bash
go run .
```

Server starts on `:8080`. Data is written to `data/lsm/` by default.

---

## Roadmap

| Chapter | Concept | Status |
|---------|---------|--------|
| Chapter 3 | memoryEngine | ✅ Done |
| Chapter 3 | fileEngine — append-only log + hash index | ✅ Done |
| Chapter 3 | LSMEngine — memtable, WAL, SSTables | ✅ Done |
| Chapter 3 | LSM compaction | 🔲 Next |
| Chapter 3 | B-Tree engine | 🔲 Upcoming |
| Chapter 4 | Binary encoding for SSTable entries | 🔲 Upcoming |
| Chapter 5 | Replication | 🔲 Stretch goal |
| Chapter 6 | Partitioning | 🔲 Stretch goal |