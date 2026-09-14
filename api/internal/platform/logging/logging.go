package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sechelper-auth-template/api/internal/platform/config"
)

// New constructs the process logger from validated application configuration.
// Relative file paths are resolved beside the executable, never against the process working directory.
func New(cfg config.LogConfig, service, environment, version string) (*zap.Logger, func(), error) {
	level := zapcore.Level(0)
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, nil, fmt.Errorf("invalid log.level: %w", err)
	}
	if cfg.Encoding != "json" {
		return nil, nil, fmt.Errorf("unsupported log.encoding: %s", cfg.Encoding)
	}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: cfg.MessageKey, LevelKey: cfg.LevelKey, TimeKey: cfg.TimeKey,
		NameKey: "logger", CallerKey: cfg.CallerKey, StacktraceKey: cfg.StacktraceKey,
		LineEnding: zapcore.DefaultLineEnding, EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime: zapcore.ISO8601TimeEncoder, EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})
	var writer io.WriteCloser
	if cfg.Mode == "stdout" {
		writer = nopWriteCloser{Writer: os.Stdout}
	} else {
		path, err := executableRelativePath(cfg.Output)
		if err != nil {
			return nil, nil, err
		}
		if info, statErr := os.Stat(path); (statErr == nil && info.IsDir()) || filepath.Ext(path) == "" {
			path = filepath.Join(path, service+".jsonl")
		}
		rotator, err := newDailyRotator(path, cfg)
		if err != nil {
			return nil, nil, err
		}
		writer = rotator
	}
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), level)
	if cfg.Sampling {
		core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	}
	options := []zap.Option{zap.Fields(zap.String("service", service), zap.String("environment", environment), zap.String("version", version))}
	if cfg.Development {
		options = append(options, zap.Development())
	}
	if !cfg.DisableCaller {
		options = append(options, zap.AddCaller())
	}
	if !cfg.DisableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	logger := zap.New(core, options...)
	return logger, func() { _ = logger.Sync(); _ = writer.Close() }, nil
}

func executableRelativePath(output string) (string, error) {
	if filepath.IsAbs(output) {
		return output, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(executable), output), nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

type dailyRotator struct {
	mu        sync.Mutex
	path, day string
	file      *os.File
	bytes     int64
	cfg       config.LogConfig
}

func newDailyRotator(path string, cfg config.LogConfig) (*dailyRotator, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	r := &dailyRotator{path: path, cfg: cfg}
	if err := r.open(time.Now()); err != nil {
		return nil, err
	}
	r.cleanup()
	return r, nil
}

func (r *dailyRotator) open(now time.Time) error {
	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	r.file, r.day, r.bytes = file, now.Format("2006-01-02"), info.Size()
	return nil
}

func (r *dailyRotator) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if r.day != now.Format("2006-01-02") || r.bytes+int64(len(p)) > int64(r.cfg.MaxSizeMB)*1024*1024 {
		if err := r.rotate(now); err != nil {
			return 0, err
		}
	}
	n, err := r.file.Write(p)
	r.bytes += int64(n)
	return n, err
}

func (r *dailyRotator) rotate(now time.Time) error {
	if err := r.file.Close(); err != nil {
		return err
	}
	base := strings.TrimSuffix(r.path, filepath.Ext(r.path))
	rotated := fmt.Sprintf("%s-%s-%d%s", base, time.Now().Format("20060102-150405"), time.Now().UnixNano(), filepath.Ext(r.path))
	if err := os.Rename(r.path, rotated); err != nil && !os.IsNotExist(err) {
		return err
	}
	if r.cfg.Compress {
		_ = compressFile(rotated)
	}
	r.cleanup()
	return r.open(now)
}

func (r *dailyRotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	return r.file.Close()
}

func (r *dailyRotator) cleanup() {
	base := strings.TrimSuffix(r.path, filepath.Ext(r.path))
	entries, _ := filepath.Glob(base + "-*" + filepath.Ext(r.path) + "*")
	sort.Slice(entries, func(i, j int) bool { return entries[i] > entries[j] })
	cutoff := time.Now().AddDate(0, 0, -min(r.cfg.RetentionDays, r.cfg.MaxAgeDays))
	for i, name := range entries {
		info, err := os.Stat(name)
		if err != nil {
			continue
		}
		if i >= r.cfg.MaxBackups || info.ModTime().Before(cutoff) {
			_ = os.Remove(name)
		}
	}
}

func compressFile(path string) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(path+".gz", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer output.Close()
	gz := gzip.NewWriter(output)
	if _, err := io.Copy(gz, input); err != nil {
		_ = gz.Close()
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return os.Remove(path)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
