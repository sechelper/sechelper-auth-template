package application

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadLinuxCPUParsesAggregateCounters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stat")
	if err := os.WriteFile(path, []byte("cpu 100 20 30 800 50 0 0 0 0 0\ncpu0 40 10 10 400 20 0 0 0 0 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readLinuxCPU(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.total != 1000 || got.idle != 850 {
		t.Fatalf("counters = %+v, want total=1000 idle=850", got)
	}
}

func TestCPUUsagePercentUsesDeltaAndRejectsInvalidSamples(t *testing.T) {
	got, ok := cpuUsagePercent(cpuCounters{total: 1000, idle: 850}, cpuCounters{total: 2000, idle: 1600})
	if !ok || got != 25 {
		t.Fatalf("cpuUsagePercent = (%v, %v), want (25, true)", got, ok)
	}
	if _, ok := cpuUsagePercent(cpuCounters{total: 2000, idle: 1600}, cpuCounters{total: 1000, idle: 850}); ok {
		t.Fatal("expected a reset counter sample to be rejected")
	}
}

func TestReadLinuxMemoryUsesMemAvailable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meminfo")
	contents := "MemTotal:       102400 kB\nMemFree:         10000 kB\nMemAvailable:    25600 kB\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	used, total, err := readLinuxMemory(path)
	if err != nil {
		t.Fatal(err)
	}
	if used != 76800*1024 || total != 102400*1024 {
		t.Fatalf("memory = (%d, %d), want (%d, %d)", used, total, 76800*1024, 102400*1024)
	}
}

func TestReadFilesystemUsageSamplesWorkingDirectory(t *testing.T) {
	used, total, err := readFilesystemUsage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 || used > total {
		t.Fatalf("filesystem usage = (%d, %d), want 0 <= used <= total and total > 0", used, total)
	}
}

func TestResourceCollectorKeepsPartialSnapshotWhenMetricIsUnavailable(t *testing.T) {
	collector := &SystemResourcesCollector{
		workDir:        t.TempDir(),
		interval:       time.Second,
		now:            func() time.Time { return time.Date(2026, time.September, 24, 10, 10, 1, 0, time.UTC) },
		readCPU:        func() (cpuCounters, error) { return cpuCounters{total: 100, idle: 50}, nil },
		readMemory:     func() (uint64, uint64, error) { return 0, 0, os.ErrNotExist },
		readDisk:       func(string) (uint64, uint64, error) { return 25, 100, nil },
		readGoroutines: func() int { return 7 },
	}
	collector.collect()
	got, ok := collector.Snapshot()
	if !ok {
		t.Fatal("expected a resource snapshot")
	}
	if got.SampledAt.Format(time.RFC3339) != "2026-09-24T10:10:01Z" {
		t.Fatalf("sampledAt = %s", got.SampledAt.Format(time.RFC3339))
	}
	if got.CPUPercent != nil || got.MemoryPercent != nil || got.MemoryUsedBytes != nil || got.MemoryTotalBytes != nil {
		t.Fatalf("first CPU sample and unavailable memory should be null: %+v", got)
	}
	if got.DiskPercent == nil || *got.DiskPercent != 25 || got.Goroutines != 7 {
		t.Fatalf("available resource values missing: %+v", got)
	}
}
