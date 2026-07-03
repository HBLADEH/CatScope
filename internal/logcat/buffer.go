package logcat

import "sync"

type RingBuffer struct {
	mu        sync.RWMutex
	max       int
	nextID    int64
	discarded int64
	entries   []LogEntry
}

func NewRingBuffer(max int) *RingBuffer {
	if max <= 0 {
		max = 100000
	}
	return &RingBuffer{
		max:    max,
		nextID: 1,
	}
}

func (b *RingBuffer) Add(entry LogEntry) LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry.ID = b.nextID
	b.nextID++
	b.entries = append(b.entries, entry)

	if len(b.entries) > b.max {
		overflow := len(b.entries) - b.max
		b.entries = append([]LogEntry(nil), b.entries[overflow:]...)
		b.discarded += int64(overflow)
	}

	return entry
}

func (b *RingBuffer) AppendMultiline(line string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return false
	}

	last := &b.entries[len(b.entries)-1]
	last.Multiline = append(last.Multiline, line)
	return true
}

func (b *RingBuffer) AppendContinuation(entry LogEntry) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return false
	}

	last := &b.entries[len(b.entries)-1]
	if !IsContinuation(*last, entry) {
		return false
	}

	last.Multiline = append(last.Multiline, entry.Raw)
	return true
}

func (b *RingBuffer) GetAfter(afterID int64, limit int) LogBatch {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit <= 0 || limit > 5000 {
		limit = 1000
	}

	entries := make([]LogEntry, 0, limit)
	for _, entry := range b.entries {
		if entry.ID > afterID {
			entries = append(entries, entry)
			if len(entries) >= limit {
				break
			}
		}
	}

	return LogBatch{
		Entries:        entries,
		Count:          len(b.entries),
		DiscardedCount: b.discarded,
		LastID:         b.lastIDLocked(),
	}
}

func (b *RingBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = nil
	b.discarded = 0
	b.nextID = 1
}

func (b *RingBuffer) Stats() (count int, discarded int64, lastID int64) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.entries), b.discarded, b.lastIDLocked()
}

func (b *RingBuffer) lastIDLocked() int64 {
	if len(b.entries) == 0 {
		return b.nextID - 1
	}
	return b.entries[len(b.entries)-1].ID
}
