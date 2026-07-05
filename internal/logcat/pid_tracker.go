package logcat

import "sync"

type PackagePIDState struct {
	PackageName string `json:"packageName"`
	PIDs        []int  `json:"pids"`
	KnownPIDs   []int  `json:"knownPids"`
	LastPID     int    `json:"lastPid,omitempty"`
}

type PIDMapper struct {
	mu          sync.RWMutex
	packageName string
	current     map[int]bool
	known       map[int]string
	lastPID     int
}

func NewPIDMapper() *PIDMapper {
	return &PIDMapper{}
}

func (m *PIDMapper) SetPackage(packageName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.packageName == packageName {
		return
	}
	m.packageName = packageName
	m.current = map[int]bool{}
	m.known = map[int]string{}
	m.lastPID = 0
}

func (m *PIDMapper) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.packageName = ""
	m.current = nil
	m.known = nil
	m.lastPID = 0
}

func (m *PIDMapper) Restore(packageName string, knownPIDs []int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.packageName = packageName
	m.current = map[int]bool{}
	m.known = map[int]string{}
	m.lastPID = 0
	for _, pid := range knownPIDs {
		if pid <= 0 {
			continue
		}
		m.known[pid] = packageName
		m.lastPID = pid
	}
}

func (m *PIDMapper) Update(pids []int) PackagePIDState {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil {
		m.current = map[int]bool{}
	}
	if m.known == nil {
		m.known = map[int]string{}
	}
	m.current = map[int]bool{}
	for _, pid := range pids {
		if pid <= 0 {
			continue
		}
		m.current[pid] = true
		if m.packageName != "" {
			m.known[pid] = m.packageName
			m.lastPID = pid
		}
	}

	return m.stateLocked()
}

func (m *PIDMapper) Apply(entry *LogEntry) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if entry == nil || entry.PID <= 0 || m.packageName == "" {
		return
	}
	if packageName, ok := m.known[entry.PID]; ok {
		entry.PackageName = packageName
	}
}

func (m *PIDMapper) State() PackagePIDState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stateLocked()
}

func (m *PIDMapper) stateLocked() PackagePIDState {
	return PackagePIDState{
		PackageName: m.packageName,
		PIDs:        sortedPIDKeys(m.current),
		KnownPIDs:   sortedPackagePIDKeys(m.known, m.packageName),
		LastPID:     m.lastPID,
	}
}

func sortedPIDKeys(values map[int]bool) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, 0, len(values))
	for pid := range values {
		result = append(result, pid)
	}
	sortInts(result)
	return result
}

func sortedPackagePIDKeys(values map[int]string, packageName string) []int {
	if len(values) == 0 || packageName == "" {
		return nil
	}
	result := make([]int, 0, len(values))
	for pid, name := range values {
		if name == packageName {
			result = append(result, pid)
		}
	}
	sortInts(result)
	return result
}

func sortInts(values []int) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
