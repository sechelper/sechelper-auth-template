package application

import (
	"bufio"
	"context"
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultResourceSampleInterval = 5 * time.Second

type ResourceMetrics struct {
	SampledAt        time.Time `json:"sampledAt"`
	CPUPercent       *float64  `json:"cpuPercent"`
	MemoryPercent    *float64  `json:"memoryPercent"`
	MemoryUsedBytes  *uint64   `json:"memoryUsedBytes"`
	MemoryTotalBytes *uint64   `json:"memoryTotalBytes"`
	DiskPercent      *float64  `json:"diskPercent"`
	DiskUsedBytes    *uint64   `json:"diskUsedBytes"`
	DiskTotalBytes   *uint64   `json:"diskTotalBytes"`
	Goroutines       int       `json:"goroutines"`
}

type cpuCounters struct {
	total uint64
	idle  uint64
}

type SystemResourcesCollector struct {
	workDir        string
	interval       time.Duration
	cancel         context.CancelFunc
	done           chan struct{}
	closeOnce      sync.Once
	mu             sync.RWMutex
	snapshot       ResourceMetrics
	hasSnapshot    bool
	previousCPU    *cpuCounters
	readCPU        func() (cpuCounters, error)
	readMemory     func() (uint64, uint64, error)
	readDisk       func(string) (uint64, uint64, error)
	readGoroutines func() int
	now            func() time.Time
}

func NewSystemResourcesCollector(workDir string, interval time.Duration) *SystemResourcesCollector {
	if strings.TrimSpace(workDir) == "" {
		workDir = "."
	}
	if interval <= 0 {
		interval = defaultResourceSampleInterval
	}
	ctx, cancel := context.WithCancel(context.Background())
	collector := &SystemResourcesCollector{
		workDir:        workDir,
		interval:       interval,
		cancel:         cancel,
		done:           make(chan struct{}),
		readCPU:        func() (cpuCounters, error) { return readLinuxCPU("/proc/stat") },
		readMemory:     func() (uint64, uint64, error) { return readLinuxMemory("/proc/meminfo") },
		readDisk:       readFilesystemUsage,
		readGoroutines: runtime.NumGoroutine,
		now:            time.Now,
	}
	go collector.run(ctx)
	return collector
}

func (c *SystemResourcesCollector) Snapshot() (ResourceMetrics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot, c.hasSnapshot
}

func (c *SystemResourcesCollector) Close() {
	c.closeOnce.Do(c.cancel)
	<-c.done
}

func (c *SystemResourcesCollector) run(ctx context.Context) {
	defer close(c.done)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	c.collect()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *SystemResourcesCollector) collect() {
	snapshot := ResourceMetrics{SampledAt: c.now().UTC(), Goroutines: c.readGoroutines()}
	if current, err := c.readCPU(); err == nil {
		if c.previousCPU != nil {
			if percent, ok := cpuUsagePercent(*c.previousCPU, current); ok {
				snapshot.CPUPercent = floatPointer(percent)
			}
		}
		c.previousCPU = &current
	}
	if used, total, err := c.readMemory(); err == nil && total > 0 && used <= total {
		snapshot.MemoryUsedBytes = uintPointer(used)
		snapshot.MemoryTotalBytes = uintPointer(total)
		snapshot.MemoryPercent = floatPointer(float64(used) * 100 / float64(total))
	}
	if used, total, err := c.readDisk(c.workDir); err == nil && total > 0 && used <= total {
		snapshot.DiskUsedBytes = uintPointer(used)
		snapshot.DiskTotalBytes = uintPointer(total)
		snapshot.DiskPercent = floatPointer(float64(used) * 100 / float64(total))
	}
	c.mu.Lock()
	c.snapshot = snapshot
	c.hasSnapshot = true
	c.mu.Unlock()
}

func readLinuxCPU(path string) (cpuCounters, error) {
	file, err := os.Open(path)
	if err != nil {
		return cpuCounters{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return cpuCounters{}, err
		}
		return cpuCounters{}, errors.New("CPU counters are unavailable")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuCounters{}, errors.New("invalid aggregate CPU counters")
	}
	var counters cpuCounters
	values := make([]uint64, 0, len(fields)-1)
	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuCounters{}, errors.New("invalid aggregate CPU counter value")
		}
		values = append(values, value)
		// Linux guest and guest_nice are already included in user and nice.
		if len(values) <= 8 {
			counters.total += value
		}
	}
	counters.idle = values[3]
	if len(values) > 4 {
		counters.idle += values[4]
	}
	return counters, nil
}

func cpuUsagePercent(previous, current cpuCounters) (float64, bool) {
	if current.total <= previous.total || current.idle < previous.idle {
		return 0, false
	}
	totalDelta := current.total - previous.total
	idleDelta := current.idle - previous.idle
	if totalDelta == 0 || idleDelta > totalDelta {
		return 0, false
	}
	percent := float64(totalDelta-idleDelta) * 100 / float64(totalDelta)
	if percent < 0 || percent > 100 {
		return 0, false
	}
	return percent, true
}

func readLinuxMemory(path string) (used, total uint64, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	var totalKB, availableKB uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			continue
		}
		switch strings.TrimSuffix(fields[0], ":") {
		case "MemTotal":
			totalKB = value
		case "MemAvailable":
			availableKB = value
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}
	if totalKB == 0 || availableKB > totalKB {
		return 0, 0, errors.New("host memory counters are unavailable")
	}
	total = totalKB * 1024
	used = (totalKB - availableKB) * 1024
	return used, total, nil
}

func readFilesystemUsage(path string) (used, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}
	if stat.Bsize <= 0 {
		return 0, 0, errors.New("filesystem block size is unavailable")
	}
	blockSize := uint64(stat.Bsize)
	total = stat.Blocks * blockSize
	available := stat.Bavail * blockSize
	if total == 0 || available > total {
		return 0, 0, errors.New("filesystem usage counters are unavailable")
	}
	return total - available, total, nil
}

func floatPointer(value float64) *float64 { return &value }
func uintPointer(value uint64) *uint64    { return &value }
